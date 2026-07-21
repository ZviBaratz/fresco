package main

import (
	"strings"
	"testing"
)

// A preset name resolves to its shipped palette.
func TestResolvePalettePreset(t *testing.T) {
	p, err := resolvePalette("tokyo-night")
	if err != nil {
		t.Fatalf("resolvePalette(tokyo-night): %v", err)
	}
	if p.A0 != "#f7768e" || p.Highlight != "#c0caf5" {
		t.Errorf("tokyo-night preset wrong: %+v", p)
	}
}

// Five comma-separated hex anchors resolve to a custom palette, A0..A3 then
// Highlight, tolerating surrounding whitespace.
func TestResolvePaletteCustom(t *testing.T) {
	p, err := resolvePalette("#111111, #222222 , #333333,#444444,#555555")
	if err != nil {
		t.Fatalf("resolvePalette(custom): %v", err)
	}
	if p.A0 != "#111111" || p.A1 != "#222222" || p.A2 != "#333333" ||
		p.A3 != "#444444" || p.Highlight != "#555555" {
		t.Errorf("custom palette mismapped: %+v", p)
	}
}

// A custom palette with a non-hex anchor is rejected through Palette.Validate.
func TestResolvePaletteCustomInvalidHex(t *testing.T) {
	_, err := resolvePalette("#111111,#222222,#333333,#444444,#zzzzzz")
	if err == nil {
		t.Fatal("want an error for a non-hex anchor, got nil")
	}
}

// A custom palette without exactly five anchors is rejected with a clear message.
func TestResolvePaletteCustomWrongCount(t *testing.T) {
	_, err := resolvePalette("#111111,#222222")
	if err == nil {
		t.Fatal("want an error for the wrong anchor count, got nil")
	}
	if !strings.Contains(err.Error(), "five") {
		t.Errorf("error should mention the five-anchor form; got %q", err)
	}
}

// An unknown name that isn't a custom spec is rejected.
func TestResolvePaletteUnknownName(t *testing.T) {
	if _, err := resolvePalette("chartreuse"); err == nil {
		t.Fatal("want an error for an unknown preset name, got nil")
	}
}

// Every shipped preset is a valid palette — guards the hand-written hex tables.
func TestPresetPalettesAllValidate(t *testing.T) {
	for _, np := range presetPalettes {
		if err := np.palette.Validate(); err != nil {
			t.Errorf("preset %q does not validate: %v", np.name, err)
		}
	}
}

// The default palette is a real preset (the first one), so a no-flag run always
// has colour to spend.
func TestDefaultPaletteResolves(t *testing.T) {
	if err := defaultPalette().Validate(); err != nil {
		t.Errorf("default palette invalid: %v", err)
	}
}
