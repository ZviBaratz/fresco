package fresco_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ZviBaratz/fresco"
	"github.com/charmbracelet/x/ansi"
)

var fuzzPalette = fresco.Palette{
	A0: "#f7768e", A1: "#bb9af7", A2: "#7aa2f7", A3: "#7dcfff", Highlight: "#c0caf5",
}

var fuzzProfiles = []fresco.ColorProfile{
	fresco.NoColor, fresco.ANSI16, fresco.ANSI256, fresco.TrueColor,
}

// FuzzParseVariant asserts ParseVariant never panics on arbitrary input and that
// any recognized name round-trips through String.
func FuzzParseVariant(f *testing.F) {
	for _, s := range []string{"rain", "tunnel", "ripple", "galaxy", "", "RAIN", "  ripple  ", "nope", "f"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, ok := fresco.ParseVariant(s)
		if ok && v.String() != s {
			t.Fatalf("ParseVariant(%q) reported ok but String() = %q", s, v.String())
		}
	})
}

// FuzzRender asserts the pure-function contract for arbitrary inputs: Render
// never panics, is deterministic, and returns exactly h lines of exactly w
// visible cells (or "" for a degenerate pane).
//
// Dimensions are bounded so the fuzzer probes rendering behavior rather than
// rediscovering that a caller can request an arbitrarily large allocation —
// bounding the pane is the caller's responsibility, by design. An out-of-range
// Variant and unusual FocalRow / LumRange values are passed through on purpose,
// to exercise the fallback and shading paths.
func FuzzRender(f *testing.F) {
	// seeds: w, h, frame, variant, focal, profIdx, lum
	f.Add(80, 24, 0, 0, -1, 3, 0.75)
	f.Add(1, 1, 5, 2, 0, 0, 1.0)
	f.Add(0, 10, 0, 1, 3, 1, 0.0)
	f.Add(12, 3, 100, 3, 1, 2, 0.5)
	f.Add(-4, -4, 7, 9, -20, 0, 0.9)

	f.Fuzz(func(t *testing.T, w, h, frame, variantRaw, focal, profIdx int, lum float64) {
		w = boundDim(w)
		h = boundDim(h)

		lumRange := lum
		prof := fuzzProfiles[nonNegMod(profIdx, len(fuzzProfiles))]
		opts := fresco.Options{
			Palette:  fuzzPalette,
			Variant:  fresco.Variant(variantRaw),
			FocalRow: focal % 300,
			LumRange: &lumRange,
			Profile:  prof,
		}

		out := fresco.Render(w, h, frame, opts)

		// Determinism: a second identical call must produce identical bytes.
		if again := fresco.Render(w, h, frame, opts); again != out {
			t.Fatalf("Render(%d,%d,%d,...) is not deterministic", w, h, frame)
		}

		// Bounds contract. A non-positive dimension always renders "".
		if w <= 0 || h <= 0 {
			if out != "" {
				t.Fatalf("degenerate pane %dx%d should render %q, got %d bytes", w, h, "", len(out))
			}
			return
		}
		// A positive-area pane may still be degenerate when its focal-to-corner
		// radius collapses (e.g. a 1x1 pane) — the documented contract allows "".
		// But that can only happen at w == 1: for w >= 2 the horizontal radius is
		// always positive, so an empty result there would be a real regression.
		if out == "" {
			if w >= 2 {
				t.Fatalf("Render(%dx%d) unexpectedly empty", w, h)
			}
			return
		}
		lines := strings.Split(out, "\n")
		if len(lines) != h {
			t.Fatalf("Render(%d,%d,...) returned %d lines, want %d", w, h, len(lines), h)
		}
		for i, line := range lines {
			if got := utf8.RuneCountInString(ansi.Strip(line)); got != w {
				t.Fatalf("line %d has %d visible cells, want %d", i, got, w)
			}
		}
	})
}

// boundDim keeps a fuzzed dimension in a sane range: modulo caps the magnitude,
// and negatives are clamped near zero so the degenerate-pane path is still
// exercised without asking for a huge allocation.
func boundDim(v int) int {
	v %= 300
	if v < -3 {
		v = -3
	}
	return v
}

// nonNegMod returns a non-negative v mod n (n > 0).
func nonNegMod(v, n int) int {
	v %= n
	if v < 0 {
		v += n
	}
	return v
}
