package fresco

import (
	"strings"
	"testing"

	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
)

// TestColorProfileResolve pins the whole enum→termenv mapping, which is the only
// place termenv is named on the Options path. Auto must defer to the ambient
// profile (both directions), and every pinned value must map to its fixed
// termenv counterpart regardless of the ambient.
func TestColorProfileResolve(t *testing.T) {
	fixed := map[ColorProfile]termenv.Profile{
		TrueColor: termenv.TrueColor,
		ANSI256:   termenv.ANSI256,
		ANSI16:    termenv.ANSI,
		NoColor:   termenv.Ascii,
	}
	// A pinned value ignores the ambient: force the opposite and assert it holds.
	withColorProfile(t, termenv.Ascii)
	for cp, want := range fixed {
		require.Equalf(t, want, cp.resolve(), "%v must resolve to a fixed profile regardless of ambient", cp)
	}

	// Auto (the zero value) tracks whatever the ambient profile is.
	for _, amb := range []termenv.Profile{termenv.TrueColor, termenv.ANSI256, termenv.ANSI, termenv.Ascii} {
		withColorProfile(t, amb)
		require.Equal(t, amb, Auto.resolve(), "Auto must defer to the ambient profile")
	}

	// An out-of-range value behaves as Auto, matching an unset field.
	withColorProfile(t, termenv.ANSI256)
	require.Equal(t, termenv.ANSI256, ColorProfile(99).resolve(), "an unknown ColorProfile must defer to the ambient like Auto")
}

// TestRenderHonorsPinnedColorProfile is the end-to-end counterpart: each pinned
// depth must stamp its own SGR signature into the bytes, under an ambient set to
// the opposite, so a Render that still read the global would fail.
func TestRenderHonorsPinnedColorProfile(t *testing.T) {
	pal := splashTestPalette()
	pal.A2 = "#0b0c0d" // a private cache entry, so other tests are unaffected
	render := func(p ColorProfile) string {
		return Render(60, 20, 3, Options{Palette: pal, Variant: Tunnel, FocalRow: centeredFocalRow(20), Profile: p})
	}

	withColorProfile(t, termenv.Ascii) // ambient is colorless; the pin must win
	require.Contains(t, render(TrueColor), "\x1b[38;2;", "TrueColor must emit 24-bit SGR")
	require.Contains(t, render(ANSI256), "\x1b[38;5;", "ANSI256 must emit 256-colour SGR")

	ansi16 := render(ANSI16)
	require.Contains(t, ansi16, "\x1b[", "ANSI16 must emit colour")
	require.NotContains(t, ansi16, "38;2;", "ANSI16 must not emit 24-bit SGR")
	require.NotContains(t, ansi16, "38;5;", "ANSI16 must not emit 256-colour SGR")

	withColorProfile(t, termenv.TrueColor) // ambient is truecolor; the pin must win
	require.NotContains(t, render(NoColor), "\x1b[", "NoColor must emit no escapes at all")
	require.Equal(t, 20, strings.Count(render(NoColor), "\n")+1, "NoColor still renders the full h×w pane")
}
