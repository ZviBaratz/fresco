package fresco_test

import (
	"testing"

	"github.com/ZviBaratz/fresco"
)

// TestRenderIsDeterministic pins fresco's central promise at the public API:
// with the color profile fixed, identical (w, h, frame, Options) always yields
// identical bytes. The engine's internals have their own determinism guard
// (TestRenderSplashFieldDeterministic); this is the external-package counterpart
// that a consumer relies on for snapshot testing.
func TestRenderIsDeterministic(t *testing.T) {
	palette := fresco.Palette{
		A0: "#f7768e", A1: "#bb9af7", A2: "#7aa2f7", A3: "#7dcfff", Highlight: "#c0caf5",
	}

	for _, v := range fresco.Variants() {
		for _, frame := range []int{0, 1, 7, 100} {
			opts := fresco.Options{Palette: palette, Variant: v, Profile: fresco.TrueColor}
			first := fresco.Render(80, 24, frame, opts)
			second := fresco.Render(80, 24, frame, opts)
			if first != second {
				t.Fatalf("Render is not deterministic for variant %s, frame %d", v, frame)
			}
		}
	}
}
