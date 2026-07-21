package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ZviBaratz/fresco"
)

// errVersion is a sentinel like flag.ErrHelp: resolveConfig returns it when
// --version is set, so main prints the (impure, ldflags-injected) version and
// exits 0 without the pure resolver ever touching the version string.
var errVersion = errors.New("version requested")

// maxFPS bounds --fps. A terminal gains nothing past a few tens of frames a
// second, and this keeps a typo (e.g. --fps 100000000, a ~10ns ticker) from
// pinning a core rather than animating.
const maxFPS = 1000

// config is the fully resolved run: everything the driver needs, with no flag
// parsing, environment reads, or profile detection left to do. resolveConfig
// turns argv + the environment + "is this a TTY?" into one of these, purely, so
// the whole resolution policy is unit-testable without a terminal.
type config struct {
	schedule variantSchedule     // which variant plays when
	palette  fresco.Palette      // resolved colours
	fps      int                 // animation ticks per second
	focalRow int                 // pane row the field emanates from; <0 = centre
	lumRange *float64            // nil = the variant's default density↔luminance split
	profile  fresco.ColorProfile // always pinned (never Auto), so frames are deterministic
	duration time.Duration       // 0 = run until quit
	size     Size                // zero value = auto-detect from the terminal
	raw      bool                // attempt raw-mode single-key controls
	once     bool                // render one frame and exit
	list     bool                // print the variant/palette listing and exit
}

// resolveConfig parses args and folds in the environment and TTY state to a
// resolved config. It is pure over its inputs — getenv and isTTY and detected are
// injected rather than read from the process — so a test drives every branch
// (including "off a TTY, degrade") with no real terminal. detected is the ambient
// colour depth (what lipgloss.ColorProfile() found), consulted only for --profile
// auto on a TTY.
func resolveConfig(args []string, getenv func(string) string, isTTY bool, detected fresco.ColorProfile) (config, error) {
	fs := flag.NewFlagSet("fresco", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // main owns usage output; keep parse errors as returned errors

	var (
		variant  = fs.String("variant", "cycle", "variant to show: a name (rain, tunnel, ripple, galaxy, aurora) or cycle/all")
		palette  = fs.String("palette", paletteNames()[0], "palette: a preset name or five comma-separated hex anchors")
		fps      = fs.Int("fps", 30, "animation frames per second")
		spv      = fs.Int("seconds-per-variant", 6, "how long each variant holds when cycling")
		focalRow = fs.Int("focal-row", -1, "pane row the field emanates from; negative centres it")
		lumRange = fs.Float64("lum-range", math.NaN(), "override the density↔luminance split, in [0,1]")
		profile  = fs.String("profile", "auto", "colour depth: auto, truecolor, ansi256, ansi16, or nocolor")
		duration = fs.Duration("duration", 0, "run for this long then exit (0 = until quit)")
		sizeFlag = fs.String("size", "", "override auto-size, e.g. 100x30 (default: detect from the terminal)")
		once     = fs.Bool("once", false, "render a single frame to stdout and exit")
		list     = fs.Bool("list", false, "list the variants and palettes, then exit")
		showVer  = fs.Bool("version", false, "print the fresco version and exit")
	)

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	if *showVer {
		return config{}, errVersion
	}

	cfg := config{
		fps:      *fps,
		focalRow: *focalRow,
		duration: *duration,
		once:     *once,
		list:     *list,
	}

	if *fps < 1 || *fps > maxFPS {
		return config{}, fmt.Errorf("--fps must be in [1, %d], got %d", maxFPS, *fps)
	}
	if *spv < 1 {
		return config{}, fmt.Errorf("--seconds-per-variant must be at least 1, got %d", *spv)
	}
	if *duration < 0 {
		return config{}, fmt.Errorf("--duration must not be negative, got %s", *duration)
	}

	sched, err := resolveSchedule(*variant, *spv, *fps)
	if err != nil {
		return config{}, err
	}
	cfg.schedule = sched

	pal, err := resolvePalette(*palette)
	if err != nil {
		return config{}, err
	}
	cfg.palette = pal

	if !math.IsNaN(*lumRange) {
		if *lumRange < 0 || *lumRange > 1 {
			return config{}, fmt.Errorf("--lum-range must be in [0,1], got %v", *lumRange)
		}
		lr := *lumRange
		cfg.lumRange = &lr
	}

	sz, err := parseSize(*sizeFlag)
	if err != nil {
		return config{}, err
	}
	cfg.size = sz

	prof, err := resolveProfile(*profile, isTTY, getenv, detected)
	if err != nil {
		return config{}, err
	}
	cfg.profile = prof

	// Off a TTY, degrade to a single static frame: never block on a size query,
	// never spew control codes into a pipe, never wait for keys that can't arrive.
	if !isTTY {
		cfg.once = true
	}
	// Raw-mode keys only for an interactive loop: not for --once, --list, or a
	// non-TTY. The driver still ANDs this with a successful MakeRaw.
	cfg.raw = isTTY && !cfg.once && !cfg.list

	return cfg, nil
}

// resolveSchedule turns a --variant value into a schedule.
func resolveSchedule(variant string, spv, fps int) (variantSchedule, error) {
	if variant == "cycle" || variant == "all" {
		return variantSchedule{cycle: true, pool: fresco.Variants(), framesPer: spv * fps}, nil
	}
	v, ok := fresco.ParseVariant(variant)
	if !ok {
		return variantSchedule{}, fmt.Errorf("unknown variant %q: use one of %s, or cycle/all",
			variant, strings.Join(variantNames(), ", "))
	}
	return variantSchedule{pinned: v}, nil
}

// resolveProfile maps the --profile flag, the environment, and TTY state to a
// pinned colour depth. It never returns Auto: Auto re-probes stdout on every
// Render call (see profile.go), which would be impure and slow in the animation
// loop, so the driver resolves the depth once here and passes it every frame.
//
// An explicit non-auto flag is the user's direct instruction and wins over the
// environment. In auto mode NO_COLOR forces no colour; off a TTY the default is
// no colour (a pipe/redirect), which FORCE_COLOR overrides; on a TTY the
// auto-detected depth is used.
func resolveProfile(flagVal string, isTTY bool, getenv func(string) string, detected fresco.ColorProfile) (fresco.ColorProfile, error) {
	switch flagVal {
	case "truecolor":
		return fresco.TrueColor, nil
	case "ansi256":
		return fresco.ANSI256, nil
	case "ansi16":
		return fresco.ANSI16, nil
	case "nocolor":
		return fresco.NoColor, nil
	case "auto", "":
		// fall through to detection
	default:
		return 0, fmt.Errorf("unknown --profile %q (want auto, truecolor, ansi256, ansi16, or nocolor)", flagVal)
	}

	if getenv("NO_COLOR") != "" {
		return fresco.NoColor, nil
	}
	if !isTTY {
		if getenv("FORCE_COLOR") != "" {
			return fresco.TrueColor, nil // deliberate override for a pipe
		}
		return fresco.NoColor, nil
	}
	if detected == fresco.Auto {
		return fresco.TrueColor, nil // a broken detector must not leak Auto into the loop
	}
	return detected, nil
}

// parseSize parses a "WxH" override; "" means auto-detect (the zero Size). It is
// strict: both dimensions must be positive integers with no trailing garbage.
func parseSize(s string) (Size, error) {
	if s == "" {
		return Size{}, nil
	}
	wStr, hStr, ok := strings.Cut(s, "x")
	w, errW := strconv.Atoi(wStr)
	h, errH := strconv.Atoi(hStr)
	if !ok || errW != nil || errH != nil || w < 1 || h < 1 {
		return Size{}, fmt.Errorf("--size must be WxH with positive dimensions, e.g. 100x30, got %q", s)
	}
	return Size{W: w, H: h}, nil
}

// variantNames lists the pinnable variant names in roster order, for help and
// error messages.
func variantNames() []string {
	vs := fresco.Variants()
	names := make([]string, len(vs))
	for i, v := range vs {
		names[i] = v.String()
	}
	return names
}
