package fresco_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ZviBaratz/fresco"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

// validPalette is a fully canonical palette the rejection tests mutate one field
// of, so a failure is always attributable to the single field under test.
func validPalette() fresco.Palette {
	return fresco.Palette{A0: "#f7768e", A1: "#bb9af7", A2: "#7aa2f7", A3: "#7dcfff", Highlight: "#c0caf5"}
}

// TestPaletteValidateAcceptsCanonical: every canonical form passes — long
// "#rrggbb", short "#rgb", and mixed/upper case.
func TestPaletteValidateAcceptsCanonical(t *testing.T) {
	for _, p := range []fresco.Palette{
		validPalette(),
		{A0: "#fff", A1: "#000", A2: "#abc", A3: "#123", Highlight: "#def"},
		{A0: "#FF00FF", A1: "#Ff0", A2: "#7AA2F7", A3: "#7dcfff", Highlight: "#C0CAF5"},
	} {
		require.NoError(t, p.Validate(), "canonical palette %+v must validate", p)
	}
}

// TestPaletteValidateRejectsMalformed pins the boundary: each malformed form is
// rejected, and the error names the offending field and echoes its value.
func TestPaletteValidateRejectsMalformed(t *testing.T) {
	bad := []string{
		"",         // empty
		"#zz",      // non-hex + too short
		"fff",      // missing '#'
		"#ffff",    // 4 digits (neither 3 nor 6)
		"#12345",   // 5 digits
		"#12345g",  // trailing non-hex — go-colorful accepts this, Validate must not
		"#ff1034z", // trailing garbage on a valid prefix
		"purple",   // a colour name, not hex
		"#",        // just the marker
	}
	for _, v := range bad {
		p := validPalette()
		p.A2 = v
		err := p.Validate()
		require.Errorf(t, err, "%q must be rejected", v)
		require.Containsf(t, err.Error(), "A2", "error for %q must name the field", v)
		require.Containsf(t, err.Error(), v, "error for %q must echo the value", v)
	}
}

// TestPaletteValidateReportsEveryField: a single call reports all bad anchors,
// not just the first, so a caller fixes a broken theme in one pass.
func TestPaletteValidateReportsEveryField(t *testing.T) {
	p := fresco.Palette{A0: "nope", A1: "#bb9af7", A2: "", A3: "#7dcfff", Highlight: "#zz"}
	err := p.Validate()
	require.Error(t, err)
	for _, field := range []string{"A0", "A2", "Highlight"} {
		require.Contains(t, err.Error(), field, "must report every offending field")
	}
	for _, ok := range []string{"A1", "A3"} {
		require.NotContains(t, err.Error(), ok+":", "must not report a valid field")
	}
}

// TestRenderIgnoresInvalidPalette is the contract Validate is advisory to: a
// palette that fails Validate still renders exactly h rows of w cells and never
// panics. Render degrades a bad palette; it does not reject it.
func TestRenderIgnoresInvalidPalette(t *testing.T) {
	broken := fresco.Palette{A0: "#zz", A1: "", A2: "nope", A3: "#ffff", Highlight: "bad"}
	require.Error(t, broken.Validate(), "precondition: this palette is invalid")

	const w, h = 40, 12
	for _, v := range fresco.Variants() {
		out := fresco.Render(w, h, 3, fresco.Options{Palette: broken, Variant: v})
		lines := strings.Split(out, "\n")
		require.Lenf(t, lines, h, "variant %s: row count on a broken palette", v)
		for i, l := range lines {
			require.Equalf(t, w, utf8.RuneCountInString(ansi.Strip(l)), "variant %s line %d width", v, i)
		}
	}
}

// TestPaletteValidateStricterThanRender makes the documented divergence concrete:
// "#12345g" is rejected by Validate yet Render paints a full frame from it.
func TestPaletteValidateStricterThanRender(t *testing.T) {
	p := validPalette()
	p.A0 = "#12345g" // go-colorful parses this (truncating the last byte); Validate does not
	require.Error(t, p.Validate(), "Validate must reject the lenient case")

	out := fresco.Render(30, 6, 1, fresco.Options{Palette: p, Variant: fresco.Ripple})
	require.NotEmpty(t, out, "Render still paints the lenient palette")
	require.Len(t, strings.Split(out, "\n"), 6, "Render honours the h×w contract regardless")
}
