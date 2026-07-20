# fresco

[![CI](https://github.com/ZviBaratz/fresco/actions/workflows/ci.yml/badge.svg)](https://github.com/ZviBaratz/fresco/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/ZviBaratz/fresco/graph/badge.svg)](https://codecov.io/gh/ZviBaratz/fresco)
[![Go Reference](https://pkg.go.dev/badge/github.com/ZviBaratz/fresco.svg)](https://pkg.go.dev/github.com/ZviBaratz/fresco)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

<p align="center">
  <img src="docs/demo.gif" width="640" alt="fresco cycling through its five variants — rain, tunnel, ripple, galaxy, and aurora">
</p>

**Generative, free-running animated fields for the terminal.** fresco paints the
whole pane with a slow-drifting, theme-coloured procedural field — a nebula, a
receding tunnel, a pool of ripples, a turning galaxy, a sky of polar curtains —
as plain ANSI text. It is a pure rendering engine, not a widget: you give it a
size and a frame number, it hands you a string.

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
| `fresco.Rain` | Matrix-style digital rain — per-column streams with bright heads and fading tails, layered at four depths. Shades by luminance alone. |
| `fresco.Tunnel` | A textured wall flying past a vanishing point — screen position maps to (depth, angle), z-fog carries distance in luminance, hue bands by depth. |
| `fresco.Ripple` | Drops falling on a dark pool — each flashes where it lands and expands into a hue-shifting ring; rings interfere where they cross. |
| `fresco.Galaxy` | An inclined spiral turning on the focal point — arms studded with star-forming knots, a bright core, an occluding dust lane for depth. |
| `fresco.Aurora` | Northern-lights curtains over dark sky — tall vertical drapes that snake and drift sideways, the hue sliding warm↔cool with altitude. |

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

For a per-frame render loop, `AppendRender(dst []byte, w, h, frame int, opts Options) []byte`
appends the same bytes to a buffer you own, so you can reuse it across frames
(`buf = fresco.AppendRender(buf[:0], w, h, frame, opts)`) instead of allocating a
fresh string each tick. `Render` is a thin wrapper over it and the two are
byte-identical.

```go
type Options struct {
	Palette  Palette      // the five colour anchors (required for colour)
	Variant  Variant      // Rain, Tunnel, Ripple, Galaxy, or Aurora
	FocalRow int          // the row the field emanates from; negative = centre
	LumRange *float64     // optional override for the density↔luminance split
	Profile  ColorProfile // colour depth; the zero value Auto = auto-detect
}

type Palette struct {
	A0, A1, A2, A3 string // "#rrggbb" warm→cool gradient anchors
	Highlight      string // the star / rain-head near-white
}
```

**Colour depth.** By default (`Profile: Auto`, the zero value) fresco auto-detects
the terminal's colour profile. Pin it — `fresco.TrueColor`, `ANSI256`, `ANSI16`,
or `NoColor` — for tests, for writing to a non-TTY, or to force colour when
piping; with a profile pinned, `Render` is pure over its inputs regardless of the
ambient terminal. `ColorProfile` is fresco's own type, so pinning depth needs no
`termenv` import.

## Roadmap

See [`docs/ROADMAP.md`](docs/ROADMAP.md) for where fresco is headed — the
`v0.2.0 — Open the doors` and `v0.3.0 — Refine & prove` milestones — and for
issues labelled [`good first issue`](https://github.com/ZviBaratz/fresco/labels/good%20first%20issue).

## Origin

fresco was extracted from [Atrium](https://github.com/ZviBaratz/atrium), where it
is the animated empty-state splash. The engine is original work; the app-side
scene composition (a logo overlay, variant selection) stayed behind in Atrium.

## License

[MIT](LICENSE) © Zvi Baratz
