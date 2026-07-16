package fresco

// ColorProfile is fresco's own name for the colour depth Render emits at, so a
// caller can pin depth without importing termenv — which is an implementation
// detail of the emitter (the SGR bytes are baked per-profile in splashLUTFor).
// It is the counterpart to Variant: a small, closed enum the public surface owns.

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// ColorProfile selects the colour depth Render bakes its SGR bytes for.
//
// The zero value, Auto, defers to the terminal's auto-detected profile — so an
// Options with no Profile set behaves exactly as the old unset (nil)
// *termenv.Profile did, and a caller that never pins depth needs to name neither
// this type nor termenv. Pinning any other value makes Render pure over its
// inputs: the same Options yield the same bytes regardless of ambient stdout
// state, which is what a standalone consumer (and a snapshot test) needs.
//
// termenv's own Profile cannot serve this role directly: its zero value is
// TrueColor, so "unset" could not mean auto-detect. That asymmetry — not merely
// the wish to hide termenv — is why fresco owns the type (see resolve).
type ColorProfile int

const (
	// Auto auto-detects the terminal's colour profile (the pre-enum nil default).
	Auto ColorProfile = iota
	// TrueColor pins 24-bit colour.
	TrueColor
	// ANSI256 pins the 256-colour palette.
	ANSI256
	// ANSI16 pins the 16-colour palette.
	ANSI16
	// NoColor emits no colour at all; the glyphs still render as plain text.
	NoColor
)

// resolve maps a ColorProfile to the termenv.Profile the emitter bakes SGR for.
// This is the only place termenv is named on the Options path.
//
// Auto reads the ambient auto-detected profile; every other value is fixed, so a
// pinned profile never touches ambient stdout state — an improvement over the
// previous code, which called lipgloss.ColorProfile() unconditionally and then
// discarded it when a profile was pinned. Any unrecognised value resolves to
// Auto, matching how an unset Options field behaved before.
func (p ColorProfile) resolve() termenv.Profile {
	switch p {
	case TrueColor:
		return termenv.TrueColor
	case ANSI256:
		return termenv.ANSI256
	case ANSI16:
		return termenv.ANSI
	case NoColor:
		return termenv.Ascii
	default: // Auto (and any out-of-range value)
		return lipgloss.ColorProfile()
	}
}
