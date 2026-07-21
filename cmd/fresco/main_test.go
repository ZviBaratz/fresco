package main

import (
	"bytes"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/ZviBaratz/fresco"
	"github.com/muesli/termenv"
)

// usage prints something that looks like help: the header, the usage line, and
// at least one flag and the keys.
func TestUsageWritesHelp(t *testing.T) {
	var b bytes.Buffer
	usage(&b)
	out := b.String()
	for _, want := range []string{"Usage:", "--variant", "--palette", "--version", "Keys", "NO_COLOR"} {
		if !strings.Contains(out, want) {
			t.Errorf("usage output missing %q", want)
		}
	}
}

// resolveVersion prefers an ldflags-injected version, falls back to the embedded
// module version, and otherwise stays "dev" — including when build info is
// absent or reports the placeholder "(devel)".
func TestResolveVersion(t *testing.T) {
	buildInfo := func(v string) *debug.BuildInfo {
		return &debug.BuildInfo{Main: debug.Module{Version: v}}
	}
	tests := []struct {
		name    string
		ldflags string
		info    *debug.BuildInfo
		ok      bool
		want    string
	}{
		{"ldflags wins", "1.2.0", buildInfo("v9.9.9"), true, "1.2.0"},
		{"falls back to module version", "dev", buildInfo("v1.2.0"), true, "v1.2.0"},
		{"ignores the (devel) placeholder", "dev", buildInfo("(devel)"), true, "dev"},
		{"no build info stays dev", "dev", nil, false, "dev"},
		{"nil info with ok stays dev", "dev", nil, true, "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveVersion(tt.ldflags, tt.info, tt.ok); got != tt.want {
				t.Errorf("resolveVersion(%q, %v, %v) = %q, want %q", tt.ldflags, tt.info, tt.ok, got, tt.want)
			}
		})
	}
}

// termenvToProfile maps every termenv depth onto fresco's enum, and never yields
// Auto (which would re-probe stdout each frame).
func TestTermenvToProfile(t *testing.T) {
	cases := map[termenv.Profile]fresco.ColorProfile{
		termenv.TrueColor: fresco.TrueColor,
		termenv.ANSI256:   fresco.ANSI256,
		termenv.ANSI:      fresco.ANSI16,
		termenv.Ascii:     fresco.NoColor,
	}
	for in, want := range cases {
		if got := termenvToProfile(in); got != want {
			t.Errorf("termenvToProfile(%v) = %v, want %v", in, got, want)
		}
		if termenvToProfile(in) == fresco.Auto {
			t.Errorf("termenvToProfile(%v) leaked Auto", in)
		}
	}
}
