package main

import (
	"fmt"
	"strings"

	"github.com/ZviBaratz/fresco"
)

// The CLI's palette vocabulary: a handful of named presets plus a custom-hex
// form. This lives entirely in cmd/ — the library takes a Palette and never ships
// named themes, so a preset is purely a caller-side convenience. Every anchor is
// canonical "#rrggbb" so fresco.Palette.Validate passes (TestPresetPalettesAllValidate
// guards the hand-typed tables); the A0..A3 anchors run warm→cool and Highlight is
// the star / rain-head near-white.

// namedPalette pairs a preset name with its palette. Presets are an ordered slice
// (not a map) so --list and the default are stable.
type namedPalette struct {
	name    string
	palette fresco.Palette
}

var presetPalettes = []namedPalette{
	{"tokyo-night", fresco.Palette{A0: "#f7768e", A1: "#bb9af7", A2: "#7aa2f7", A3: "#7dcfff", Highlight: "#c0caf5"}},
	{"nord", fresco.Palette{A0: "#bf616a", A1: "#d08770", A2: "#81a1c1", A3: "#88c0d0", Highlight: "#eceff4"}},
	{"gruvbox", fresco.Palette{A0: "#fb4934", A1: "#fe8019", A2: "#83a598", A3: "#8ec07c", Highlight: "#ebdbb2"}},
	{"mono", fresco.Palette{A0: "#3a3a3a", A1: "#6c6c6c", A2: "#9e9e9e", A3: "#d0d0d0", Highlight: "#ffffff"}},
}

// defaultPalette is the palette a no-flag run uses — the first preset.
func defaultPalette() fresco.Palette { return presetPalettes[0].palette }

// paletteNames lists the preset names in order, for --list and error messages.
func paletteNames() []string {
	names := make([]string, len(presetPalettes))
	for i, np := range presetPalettes {
		names[i] = np.name
	}
	return names
}

// resolvePalette turns a --palette value into a Palette. A value that names a
// preset resolves to it; a value with commas is parsed as five custom anchors
// (A0,A1,A2,A3,Highlight) and validated through fresco.Palette.Validate, so a
// typo'd hex is surfaced deliberately rather than painted as a fallback; anything
// else is an unknown name.
func resolvePalette(spec string) (fresco.Palette, error) {
	for _, np := range presetPalettes {
		if np.name == spec {
			return np.palette, nil
		}
	}
	if strings.Contains(spec, ",") {
		return parseCustomPalette(spec)
	}
	return fresco.Palette{}, fmt.Errorf("unknown palette %q: use a preset (%s) or five comma-separated hex anchors",
		spec, strings.Join(paletteNames(), ", "))
}

// parseCustomPalette parses "A0,A1,A2,A3,Highlight" hex anchors, trimming
// surrounding whitespace, then validates the result.
func parseCustomPalette(spec string) (fresco.Palette, error) {
	parts := strings.Split(spec, ",")
	if len(parts) != 5 {
		return fresco.Palette{}, fmt.Errorf("custom palette needs five comma-separated hex anchors (A0,A1,A2,A3,Highlight), got %d", len(parts))
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	p := fresco.Palette{A0: parts[0], A1: parts[1], A2: parts[2], A3: parts[3], Highlight: parts[4]}
	if err := p.Validate(); err != nil {
		return fresco.Palette{}, err
	}
	return p, nil
}
