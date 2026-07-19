// Command preview is the render-and-look beauty gate for authoring a fresco
// variant. It renders a variant in TrueColor so you can tune lumRange by eye and
// confirm the aux channel's hue.
//
// It lives *inside* the fresco module on purpose: `go run` it by path and the
// import below resolves to your local, in-progress code — the new variant you are
// registering right now — with no go.mod, temp dir, or replace directive. The
// dot-directory keeps it invisible to `go build ./...`, `go test ./...`, and the
// linter, so it never touches the package or CI.
//
//	# live terminal: watch the colour animation move
//	go run ./.claude/skills/new-variant/preview -variant veil
//
//	# sweep lumRange by eye — run each and compare (see SKILL.md §3)
//	go run ./.claude/skills/new-variant/preview -variant veil -lum 0
//	go run ./.claude/skills/new-variant/preview -variant veil -lum 0.5
//	go run ./.claude/skills/new-variant/preview -variant veil -lum 0.75
//	go run ./.claude/skills/new-variant/preview -variant veil -lum 1
//
//	# the glyph structure, without colour: the view that shows the density ramp
//	# (· o O 0 @) — this is what lumRange actually moves, and the one view a PNG
//	# rasterizer cannot show you (see SKILL.md §6)
//	go run ./.claude/skills/new-variant/preview -variant veil -mono
//
//	# no TTY (an agent, CI, a pipe): emits ONE TrueColor frame and exits, so you
//	# can inspect the emitted SGR bytes instead of killing an endless loop
//	go run ./.claude/skills/new-variant/preview -variant veil | less -R
//
//	# a filmstrip: N consecutive frames with `--- frame N ---` headers, which is
//	# both readable as text and the shape the plugin's ansi2png.py splits on
//	go run ./.claude/skills/new-variant/preview -variant veil -frames 6
//
//	# the two sizes SKILL.md §6 asks for
//	go run ./.claude/skills/new-variant/preview -variant veil -w 30 -h 10
//	go run ./.claude/skills/new-variant/preview -variant veil -w 240 -h 60
//
// Every knob is a flag: the variant is *not* pinned in the source. An earlier
// version of this program hardcoded one, and every authoring session that reached
// for it had to edit the source first — so sessions wrote their own throwaway
// harness instead, which is how a preview program gets bypassed by the very people
// it is for.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ZviBaratz/fresco"
)

// lumUnset is the -lum sentinel for "leave the variant's shipped policy alone".
// A real lumRange is in [0,1], so a negative value cannot collide with one.
const lumUnset = -1

// defaultFrame is a mid-animation frame rather than 0, because frame 0 is
// degenerate for some fields — ripple's drops have zero radius at phase 0, so a
// still of frame 0 is an empty pool and reads as a broken variant.
const defaultFrame = 42

func main() {
	var (
		variant = flag.String("variant", "", "variant to render (required): "+variantNames())
		w       = flag.Int("w", 96, "pane width in cells")
		h       = flag.Int("h", 30, "pane height in cells")
		frame   = flag.Int("frame", defaultFrame, "first frame to render")
		frames  = flag.Int("frames", 0, "render exactly N frames and exit; 0 means live loop on a TTY, one frame otherwise")
		lum     = flag.Float64("lum", lumUnset, "override lumRange in [0,1]; negative keeps the variant's shipped policy")
		mono    = flag.Bool("mono", false, "render with NoColor: the glyph grid, which is where the density ramp reads")
	)
	flag.Parse()

	v, ok := fresco.ParseVariant(*variant)
	if !ok {
		// Naming the valid set matters more here than usual: a new variant is
		// unparseable until its variantNames entry lands (SKILL.md §4), so this
		// message doubles as "you have not registered it yet".
		fmt.Fprintf(os.Stderr, "preview: unknown variant %q; want one of %s\n", *variant, variantNames())
		os.Exit(2)
	}

	opts := fresco.Options{
		Palette:  fresco.Palette{A0: "#f7768e", A1: "#bb9af7", A2: "#7aa2f7", A3: "#7dcfff", Highlight: "#c0caf5"},
		Variant:  v,
		Profile:  fresco.TrueColor,
		FocalRow: -1, // centre on the pane
	}
	if *mono {
		opts.Profile = fresco.NoColor
	}
	if *lum >= 0 {
		opts.LumRange = lum
	}

	// A non-TTY stdout (an agent, CI, a pipe) can't watch an animation, so emit a
	// fixed number of frames and exit — their SGR bytes are what you inspect to
	// confirm the hue tracks aux. A real terminal gets the live loop.
	live := *frames == 0 && isTerminal(os.Stdout)
	if live {
		fmt.Print("\x1b[?25l")
		defer fmt.Print("\x1b[?25h")
		for f := *frame; ; f++ {
			fmt.Printf("\x1b[H%s\n", fresco.Render(*w, *h, f, opts))
			time.Sleep(time.Second / 30)
		}
	}

	n := *frames
	if n < 1 {
		n = 1
	}
	for i := 0; i < n; i++ {
		// Header only for a real filmstrip: a lone frame is cleaner without one,
		// and every consumer handles a bare frame.
		if n > 1 {
			fmt.Printf("--- frame %d ---\n", *frame+i)
		}
		fmt.Println(fresco.Render(*w, *h, *frame+i, opts))
	}
}

// variantNames lists the parseable variant names, sorted so the usage text is
// stable. It is derived from the shipped roster rather than hand-listed, so a
// newly registered variant appears here for free.
func variantNames() string {
	names := make([]string, 0, len(fresco.Variants()))
	for _, v := range fresco.Variants() {
		names = append(names, v.String())
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
