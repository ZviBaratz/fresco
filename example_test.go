package fresco_test

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/ZviBaratz/fresco"
)

// examplePalette is the tokyo-night-leaning gradient used across the examples:
// A0..A3 are the warm→cool anchors, Highlight is the star/rain-head near-white.
var examplePalette = fresco.Palette{
	A0:        "#f7768e",
	A1:        "#bb9af7",
	A2:        "#7aa2f7",
	A3:        "#7dcfff",
	Highlight: "#c0caf5",
}

// ExampleRender shows the core call and its contract: Render returns exactly h
// lines of exactly w visible cells. Pinning Options.Profile makes the output
// deterministic regardless of the ambient terminal — useful in tests and when
// writing to a non-TTY.
func ExampleRender() {
	frame := fresco.Render(12, 3, 0, fresco.Options{
		Palette: examplePalette,
		Variant: fresco.Ripple,
		Profile: fresco.NoColor, // pin the colour depth so the output is stable
	})

	lines := strings.Split(frame, "\n")
	fmt.Printf("%d lines\n", len(lines))
	fmt.Printf("%d cells per line\n", utf8.RuneCountInString(lines[0]))
	// Output:
	// 3 lines
	// 12 cells per line
}

// ExampleParseVariant resolves a pinnable variant name; the bool reports whether
// the name was recognized.
func ExampleParseVariant() {
	v, ok := fresco.ParseVariant("ripple")
	fmt.Println(v, ok)

	_, ok = fresco.ParseVariant("nope")
	fmt.Println(ok)
	// Output:
	// ripple true
	// false
}

// ExampleVariants lists the shipped variants in their rotation order.
func ExampleVariants() {
	for _, v := range fresco.Variants() {
		fmt.Println(v)
	}
	// Output:
	// rain
	// tunnel
	// ripple
	// galaxy
	// aurora
}
