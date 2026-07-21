package main

import (
	"testing"

	"github.com/ZviBaratz/fresco"
)

func envFrom(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

var noEnv = envFrom(nil)

// The zero-flag run on a real TTY: cycle the whole roster at 30fps, centred,
// colour auto-detected and pinned, interactive, running forever.
func TestResolveConfigDefaults(t *testing.T) {
	cfg, err := resolveConfig(nil, noEnv, true, fresco.TrueColor)
	if err != nil {
		t.Fatalf("resolveConfig defaults: %v", err)
	}
	if !cfg.schedule.cycle || len(cfg.schedule.pool) != len(fresco.Variants()) {
		t.Errorf("default schedule should cycle the full roster, got %+v", cfg.schedule)
	}
	if cfg.fps != 30 {
		t.Errorf("default fps = %d, want 30", cfg.fps)
	}
	if cfg.schedule.framesPer != 6*30 {
		t.Errorf("framesPer = %d, want %d (6s × 30fps)", cfg.schedule.framesPer, 6*30)
	}
	if cfg.focalRow != -1 {
		t.Errorf("default focalRow = %d, want -1", cfg.focalRow)
	}
	if cfg.lumRange != nil {
		t.Errorf("default lumRange should be nil (unset), got %v", *cfg.lumRange)
	}
	if cfg.profile != fresco.TrueColor {
		t.Errorf("auto profile on a truecolor TTY = %v, want TrueColor", cfg.profile)
	}
	if cfg.duration != 0 || cfg.once || cfg.list || (cfg.size != Size{}) {
		t.Errorf("unexpected non-zero defaults: %+v", cfg)
	}
	if !cfg.raw {
		t.Error("interactive run on a TTY should request raw mode")
	}
}

// A named variant pins the schedule.
func TestResolveConfigPinnedVariant(t *testing.T) {
	cfg, err := resolveConfig([]string{"--variant", "ripple"}, noEnv, true, fresco.TrueColor)
	if err != nil {
		t.Fatalf("resolveConfig: %v", err)
	}
	if cfg.schedule.cycle || cfg.schedule.pinned != fresco.Ripple {
		t.Errorf("--variant ripple should pin ripple, got %+v", cfg.schedule)
	}
}

// "all" is an accepted alias for "cycle".
func TestResolveConfigVariantAllAliasesCycle(t *testing.T) {
	cfg, err := resolveConfig([]string{"--variant", "all"}, noEnv, true, fresco.TrueColor)
	if err != nil {
		t.Fatalf("resolveConfig: %v", err)
	}
	if !cfg.schedule.cycle {
		t.Error("--variant all should cycle")
	}
}

func TestResolveConfigUnknownVariant(t *testing.T) {
	if _, err := resolveConfig([]string{"--variant", "nope"}, noEnv, true, fresco.TrueColor); err == nil {
		t.Fatal("want an error for an unknown variant name")
	}
}

// Off a TTY the CLI degrades: render one frame, no raw mode, no colour spew.
func TestResolveConfigNonTTYDegrades(t *testing.T) {
	cfg, err := resolveConfig(nil, noEnv, false, fresco.TrueColor)
	if err != nil {
		t.Fatalf("resolveConfig: %v", err)
	}
	if !cfg.once {
		t.Error("non-TTY should degrade to --once")
	}
	if cfg.raw {
		t.Error("non-TTY must not request raw mode")
	}
	if cfg.profile != fresco.NoColor {
		t.Errorf("non-TTY auto profile = %v, want NoColor", cfg.profile)
	}
}

// --once on a TTY renders one frame and therefore never enters the interactive
// key loop.
func TestResolveConfigOnceDisablesRaw(t *testing.T) {
	cfg, err := resolveConfig([]string{"--once"}, noEnv, true, fresco.TrueColor)
	if err != nil {
		t.Fatalf("resolveConfig: %v", err)
	}
	if !cfg.once || cfg.raw {
		t.Errorf("--once should set once and clear raw, got once=%v raw=%v", cfg.once, cfg.raw)
	}
}

func TestResolveConfigSize(t *testing.T) {
	cfg, err := resolveConfig([]string{"--size", "80x24"}, noEnv, true, fresco.TrueColor)
	if err != nil {
		t.Fatalf("resolveConfig: %v", err)
	}
	if (cfg.size != Size{W: 80, H: 24}) {
		t.Errorf("size = %+v, want 80x24", cfg.size)
	}
}

func TestResolveConfigBadSize(t *testing.T) {
	for _, bad := range []string{"80", "80x", "x24", "0x24", "80x0", "-1x10", "axb"} {
		if _, err := resolveConfig([]string{"--size", bad}, noEnv, true, fresco.TrueColor); err == nil {
			t.Errorf("--size %q should error", bad)
		}
	}
}

func TestResolveConfigBadFPSAndSPV(t *testing.T) {
	if _, err := resolveConfig([]string{"--fps", "0"}, noEnv, true, fresco.TrueColor); err == nil {
		t.Error("--fps 0 should error")
	}
	if _, err := resolveConfig([]string{"--seconds-per-variant", "0"}, noEnv, true, fresco.TrueColor); err == nil {
		t.Error("--seconds-per-variant 0 should error")
	}
	if _, err := resolveConfig([]string{"--fps", "100000000"}, noEnv, true, fresco.TrueColor); err == nil {
		t.Error("an absurd --fps should error rather than spin a core with a nanosecond ticker")
	}
}

func TestResolveConfigLumRange(t *testing.T) {
	cfg, err := resolveConfig([]string{"--lum-range", "0.5"}, noEnv, true, fresco.TrueColor)
	if err != nil {
		t.Fatalf("resolveConfig: %v", err)
	}
	if cfg.lumRange == nil || *cfg.lumRange != 0.5 {
		t.Errorf("lumRange = %v, want 0.5", cfg.lumRange)
	}
	if _, err := resolveConfig([]string{"--lum-range", "2"}, noEnv, true, fresco.TrueColor); err == nil {
		t.Error("--lum-range 2 (out of [0,1]) should error")
	}
}

func TestResolveConfigFramesPerScalesWithFPSAndSPV(t *testing.T) {
	cfg, err := resolveConfig([]string{"--fps", "60", "--seconds-per-variant", "2"}, noEnv, true, fresco.TrueColor)
	if err != nil {
		t.Fatalf("resolveConfig: %v", err)
	}
	if cfg.schedule.framesPer != 120 {
		t.Errorf("framesPer = %d, want 120 (2s × 60fps)", cfg.schedule.framesPer)
	}
}

func TestResolveConfigPropagatesPaletteError(t *testing.T) {
	if _, err := resolveConfig([]string{"--palette", "bogus"}, noEnv, true, fresco.TrueColor); err == nil {
		t.Error("a bad --palette should surface an error")
	}
}

// resolveProfile: explicit flags, auto detection, and the NO_COLOR/FORCE_COLOR
// conventions — the whole "force a sane profile" contract, table-driven.
func TestResolveProfile(t *testing.T) {
	tests := []struct {
		name     string
		flag     string
		isTTY    bool
		env      map[string]string
		detected fresco.ColorProfile
		want     fresco.ColorProfile
	}{
		{"auto TTY uses detected", "auto", true, nil, fresco.ANSI256, fresco.ANSI256},
		{"auto non-TTY degrades to NoColor", "auto", false, nil, fresco.TrueColor, fresco.NoColor},
		{"NO_COLOR overrides auto", "auto", true, map[string]string{"NO_COLOR": "1"}, fresco.TrueColor, fresco.NoColor},
		{"FORCE_COLOR forces colour off a TTY", "auto", false, map[string]string{"FORCE_COLOR": "1"}, fresco.NoColor, fresco.TrueColor},
		{"explicit truecolor wins over NO_COLOR", "truecolor", true, map[string]string{"NO_COLOR": "1"}, fresco.ANSI16, fresco.TrueColor},
		{"explicit nocolor", "nocolor", true, nil, fresco.TrueColor, fresco.NoColor},
		{"explicit ansi16", "ansi16", true, nil, fresco.TrueColor, fresco.ANSI16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveProfile(tt.flag, tt.isTTY, envFrom(tt.env), tt.detected)
			if err != nil {
				t.Fatalf("resolveProfile: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
			if got == fresco.Auto {
				t.Error("resolveProfile must never return Auto (it re-probes stdout every frame)")
			}
		})
	}
}

func TestResolveProfileUnknownFlag(t *testing.T) {
	if _, err := resolveProfile("chartreuse", true, noEnv, fresco.TrueColor); err == nil {
		t.Error("unknown --profile value should error")
	}
}

// Detection should never leak Auto through the auto path even if the injected
// detected value is Auto (a broken detector); pin something concrete.
func TestResolveProfileNeverReturnsAuto(t *testing.T) {
	got, err := resolveProfile("auto", true, noEnv, fresco.Auto)
	if err != nil {
		t.Fatalf("resolveProfile: %v", err)
	}
	if got == fresco.Auto {
		t.Errorf("want a concrete profile, got Auto")
	}
}
