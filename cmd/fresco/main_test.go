package main

import (
	"bytes"
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
	for _, want := range []string{"Usage:", "--variant", "--palette", "Keys", "NO_COLOR"} {
		if !strings.Contains(out, want) {
			t.Errorf("usage output missing %q", want)
		}
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
