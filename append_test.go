package fresco_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ZviBaratz/fresco"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func appendPalette() fresco.Palette {
	return fresco.Palette{A0: "#f7768e", A1: "#bb9af7", A2: "#7aa2f7", A3: "#7dcfff", Highlight: "#c0caf5"}
}

// TestAppendRenderMatchesRender is the additive guarantee: AppendRender(nil, …)
// yields byte-identical output to Render across variants, frames, and profiles.
func TestAppendRenderMatchesRender(t *testing.T) {
	pal := appendPalette()
	for _, v := range fresco.Variants() {
		for _, prof := range []fresco.ColorProfile{fresco.TrueColor, fresco.ANSI256, fresco.NoColor} {
			for _, frame := range []int{0, 1, 7, 100} {
				opts := fresco.Options{Palette: pal, Variant: v, Profile: prof}
				want := fresco.Render(80, 24, frame, opts)
				got := string(fresco.AppendRender(nil, 80, 24, frame, opts))
				require.Equalf(t, want, got, "variant %s profile %v frame %d", v, prof, frame)
			}
		}
	}
}

// TestAppendRenderReusesBuffer exercises the whole point of the additive path: a
// caller can render frame after frame into one buffer with AppendRender(buf[:0],
// …), each frame still matches Render, and the backing array is not reallocated
// once it is large enough (the buffer is reused, not regrown).
func TestAppendRenderReusesBuffer(t *testing.T) {
	opts := fresco.Options{Palette: appendPalette(), Variant: fresco.Rain, Profile: fresco.TrueColor}
	const w, h = 60, 24

	buf := fresco.AppendRender(nil, w, h, 0, opts)
	require.NotEmpty(t, buf)
	c := cap(buf)

	for frame := 1; frame < 30; frame++ {
		buf = fresco.AppendRender(buf[:0], w, h, frame, opts)
		require.Equalf(t, fresco.Render(w, h, frame, opts), string(buf), "reused buffer must match Render at frame %d", frame)
	}
	require.Equal(t, c, cap(buf), "same-size frames must not reallocate the buffer after the first")
}

// TestAppendRenderAppendsInPlace: a non-empty dst is preserved and the frame is
// appended after it, so a caller can build a larger buffer around the field.
func TestAppendRenderAppendsInPlace(t *testing.T) {
	opts := fresco.Options{Palette: appendPalette(), Variant: fresco.Galaxy, Profile: fresco.NoColor}
	const prefix = "PREFIX>"
	out := fresco.AppendRender([]byte(prefix), 40, 10, 3, opts)
	require.True(t, strings.HasPrefix(string(out), prefix), "existing bytes must be preserved")

	frame := string(out[len(prefix):])
	require.Equal(t, fresco.Render(40, 10, 3, opts), frame, "the appended tail must equal Render's output")
}

// TestAppendRenderDegeneratePane mirrors Render's "" contract: a degenerate pane
// appends nothing, leaving dst untouched.
func TestAppendRenderDegeneratePane(t *testing.T) {
	opts := fresco.Options{Palette: appendPalette(), Variant: fresco.Ripple}
	require.Nil(t, fresco.AppendRender(nil, 0, 10, 0, opts), "nil dst + degenerate pane stays nil")
	require.Empty(t, string(fresco.AppendRender(nil, -3, 5, 0, opts)))

	pre := []byte("keep")
	require.Equal(t, "keep", string(fresco.AppendRender(pre, 0, 0, 0, opts)), "degenerate pane appends nothing")
}

// TestAppendRenderBounds: the reused-buffer path honours the exactly-h×w-cells
// contract just as Render does.
func TestAppendRenderBounds(t *testing.T) {
	opts := fresco.Options{Palette: appendPalette(), Variant: fresco.Tunnel, Profile: fresco.TrueColor}
	var buf []byte
	for _, s := range [][2]int{{50, 18}, {80, 30}, {51, 19}, {120, 40}} {
		w, h := s[0], s[1]
		buf = fresco.AppendRender(buf[:0], w, h, 5, opts)
		lines := strings.Split(string(buf), "\n")
		require.Lenf(t, lines, h, "%dx%d row count", w, h)
		for i, l := range lines {
			require.Equalf(t, w, utf8.RuneCountInString(ansi.Strip(l)), "%dx%d line %d width", w, h, i)
		}
	}
}

// BenchmarkRenderString and BenchmarkAppendRenderReuse are the before/after pair
// for #17 (and feed #21): Render allocates a fresh string every frame, while a
// reused AppendRender buffer removes the output allocations entirely — the field
// scratch buffers that remain are the separate concern #21 addresses.
func BenchmarkRenderString(b *testing.B) {
	opts := fresco.Options{Palette: appendPalette(), Variant: fresco.Tunnel, Profile: fresco.TrueColor}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fresco.Render(120, 40, i, opts)
	}
}

func BenchmarkAppendRenderReuse(b *testing.B) {
	opts := fresco.Options{Palette: appendPalette(), Variant: fresco.Tunnel, Profile: fresco.TrueColor}
	var buf []byte
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf = fresco.AppendRender(buf[:0], 120, 40, i, opts)
	}
}
