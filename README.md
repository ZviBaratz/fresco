# fresco

[![Go Reference](https://pkg.go.dev/badge/github.com/ZviBaratz/fresco.svg)](https://pkg.go.dev/github.com/ZviBaratz/fresco)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Generative, free-running animated fields for the terminal.** fresco paints the
whole pane with a slow-drifting, theme-coloured procedural field — a nebula, a
receding tunnel, a pool of ripples, a turning galaxy — as plain ANSI text. It is
a pure rendering engine, not a widget: you give it a size and a frame number, it
hands you a string.

```go
frame := fresco.Render(width, height, tick, fresco.Options{
    Palette: fresco.Palette{A0: "#f7768e", A1: "#bb9af7", A2: "#7aa2f7", A3: "#7dcfff", Highlight: "#c0caf5"},
    Variant: fresco.Ripple,
})
```

<p align="center"><em>Run <code>go run github.com/ZviBaratz/fresco/cmd/fresco-demo@latest</code> to see it move.</em></p>

## What it is (and isn't)

fresco renders *open-ended, free-running fields*: continuous procedural motion
that never resolves, meant for splash screens, idle/empty states, and ambient
backdrops. That's a different thing from the neighbours:

- Not a **spinner** or progress indicator (briandowns/spinner).
- Not a **figlet/banner** generator (go-figure) — it draws fields, not letters.
- Not an **image→ASCII** converter (chafa, ascii-image-converter) — nothing is
  converted; the image is generated from math.
- Not a preset **effects roster** that animates *your text* in and out. fresco's
  fields have no subject; they just drift.

Everything is deterministic and pure over its inputs — the same
`(width, height, frame, Options)` always yields the same bytes — so it is
snapshot-testable and trivial to drive from any render loop.

## Install

```sh
go get github.com/ZviBaratz/fresco
```

Requires Go 1.25+. Depends only on the Charm colour stack (`lipgloss`,
`x/ansi`) and `go-colorful`.

## The variants

| Variant | What it is |
|---|---|
| `fresco.Rain` | Matrix-style digital rain — per-column streams with bright heads and fading tails, layered at three depths. Shades by luminance alone. |
| `fresco.Tunnel` | A textured wall flying past a vanishing point — screen position maps to (depth, angle), z-fog carries distance in luminance, hue bands by depth. |
| `fresco.Ripple` | Drops falling on a dark pool — each flashes where it lands and expands into a hue-shifting ring; rings interfere where they cross. |
| `fresco.Galaxy` | An inclined spiral turning on the focal point — soft arms, a bright core, an occluding dust lane for depth. |

`fresco.Variants()` returns them all (the natural rotation order);
`fresco.ParseVariant("ripple")` resolves a name.

## Quickstart: an animation loop

```go
package main

import (
	"fmt"
	"time"

	"github.com/ZviBaratz/fresco"
)

func main() {
	pal := fresco.Palette{A0: "#f7768e", A1: "#bb9af7", A2: "#7aa2f7", A3: "#7dcfff", Highlight: "#c0caf5"}
	for frame := 0; ; frame++ {
		fmt.Printf("\x1b[H%s", fresco.Render(96, 30, frame, fresco.Options{
			Palette: pal,
			Variant: fresco.Galaxy,
		}))
		time.Sleep(time.Second / 30)
	}
}
```

See [`cmd/fresco-demo`](cmd/fresco-demo) for a version that cycles every variant
and restores your cursor on exit.

## API

`Render(w, h, frame int, opts Options) string` — the whole surface. It returns
exactly `h` lines of exactly `w` visible cells (or `""` for a degenerate pane).

```go
type Options struct {
	Palette  Palette          // the five colour anchors (required for colour)
	Variant  Variant          // Rain, Tunnel, Ripple, or Galaxy
	FocalRow int              // the row the field emanates from; negative = centre
	LumRange *float64         // optional override for the density↔luminance split
	Profile  *termenv.Profile // optional: pin the colour depth (nil = auto-detect)
}

type Palette struct {
	A0, A1, A2, A3 string // "#rrggbb" warm→cool gradient anchors
	Highlight      string // the star / rain-head near-white
}
```

**Colour depth.** By default fresco auto-detects the terminal's colour profile
(truecolor / 256 / 16 / none). Set `Options.Profile` to pin it — useful for
tests, for writing to a non-TTY, or for forcing colour when piping. With it set,
`Render` is pure over its inputs regardless of the ambient terminal.

## Origin

fresco was extracted from [Atrium](https://github.com/ZviBaratz/atrium), where it
is the animated empty-state splash. The engine is original work; the app-side
scene composition (a logo overlay, variant selection) stayed behind in Atrium.

## License

[MIT](LICENSE) © Zvi Baratz
