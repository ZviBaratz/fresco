package main

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ZviBaratz/fresco"
)

// testPalette is the tokyo-night-leaning gradient the cmd tests render with; any
// five canonical hex anchors would do.
var testPalette = fresco.Palette{
	A0: "#f7768e", A1: "#bb9af7", A2: "#7aa2f7", A3: "#7dcfff", Highlight: "#c0caf5",
}

// An incremental frame homes the cursor (no clear) and then paints exactly h
// lines of exactly w visible cells — the same contract Render keeps, with a bare
// home in front so the alternate screen repaints in place without flicker.
func TestAppendFrameIncrementalHomesCursor(t *testing.T) {
	got := string(appendFrame(nil, Size{W: 8, H: 3}, 0, fresco.Options{
		Palette: testPalette, Variant: fresco.Rain, Profile: fresco.NoColor,
	}, false))

	rest, ok := strings.CutPrefix(got, seqHome)
	if !ok {
		t.Fatalf("incremental frame must start with home %q; got %q", seqHome, got)
	}
	lines := strings.Split(rest, "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d: %q", len(lines), rest)
	}
	for i, ln := range lines {
		if n := utf8.RuneCountInString(ln); n != 8 {
			t.Errorf("line %d: want 8 cells, got %d (%q)", i, n, ln)
		}
	}
}

// A full frame clears the pane first, so a resize that shrank the terminal can't
// leave stale cells from the previous, larger field behind.
func TestAppendFrameFullClearsFirst(t *testing.T) {
	got := string(appendFrame(nil, Size{W: 8, H: 3}, 0, fresco.Options{
		Palette: testPalette, Variant: fresco.Rain, Profile: fresco.NoColor,
	}, true))

	if !strings.HasPrefix(got, seqClearHome) {
		t.Fatalf("full frame must start with clear+home %q; got %q", seqClearHome, got)
	}
}

// appendFrame appends to the caller's buffer (the alloc-free reuse path) rather
// than allocating a fresh one each tick.
func TestAppendFrameAppendsToDst(t *testing.T) {
	got := string(appendFrame([]byte("PREFIX"), Size{W: 4, H: 2}, 0, fresco.Options{
		Palette: testPalette, Variant: fresco.Rain, Profile: fresco.NoColor,
	}, false))

	if !strings.HasPrefix(got, "PREFIX"+seqHome) {
		t.Fatalf("appendFrame must append after dst; got %q", got)
	}
}
