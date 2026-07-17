---
name: new-variant
description: >-
  Use when adding, authoring, or contributing a new fresco splash *variant* — a
  new generative field for the roster (rain/tunnel/ripple/galaxy-style) — inside
  the fresco repo (github.com/ZviBaratz/fresco). Triggers: "add a variant", "new
  fresco field/animation", writing a splashPointFn, the propose-a-new-variant
  issue, or building an aurora/plasma/flow/nebula-style field.
---

# Adding a fresco variant

## Overview

A *variant* is one field — rain, tunnel, ripple, galaxy. It is nothing more than
**a pure point-function plus a two-value Pass-2 policy**; the engine (contrast,
edge vignette, gradient LUT, ANSI emit) is shared and handled for you. A variant
earns its place by having a **motion premise the others don't** — and it is only
finished once you have *watched it move, in colour, and tuned it by looking.*

This skill is the operational path and is self-sufficient. `docs/variants.md` is
the in-repo companion with the full worked example and the deeper "why" behind
every rule — read it alongside this.

**Core principle: premise first, arithmetic last.** You *design* the motion, then
you *render and look*. You never settle a taste constant (`lumRange`, sharpness,
speed) by reasoning or by analogy to another variant. fresco's single
most-repeated lesson: **tune by rendering a sweep and looking, not by
arithmetic.**

## Do it in this order

Front-loading the craft is the point. Skip to code and you get a registered field
that renders, passes tests, and reads as dead texture.

1. **State the motion premise** — one sentence: *what moves, and why the eye reads
   it as motion.*
2. **Write the point-function** (pure).
3. **Choose the ops policy** (`stars`, `lumRange`) — with a rationale, decided by
   rendering.
4. **Register** at every touchpoint.
5. **Test** — the shared contract picks you up; add a range test.
6. **Render and look** — the beauty gate, in colour, on a sweep.

## 1 · Motion premise — the craft

Write it down before any math, then hold the field to these. Each is a hard-won
fresco rule; break one and the field looks wrong in a way no arithmetic warns you
about:

- **A moving signal, not a shimmer.** A bright leading edge with a decaying trail
  is the canonical "this is moving" cue (rain, ripple's rings). A field that only
  flickers in place reads as static grain.
- **Negative space is required.** A glyph in *every* cell is texture, not weather.
  Leave dark space for the motion to move *through*. If your reference effect
  canonically *fills the screen* (a classic plasma, a full raster), this is a
  **reinterpretation, not a copy**: carve voids into it — blobs of light in dark
  space — so it reads as weather on this medium.
- **No fixed bright points over a moving field** — they read as stuck pixels. Only
  ripple keeps `stars`, because its pool is *still*.
- **Anchor to the focal point.** The field is evaluated in focal-relative,
  aspect-corrected `(dx, dy)`; the edge vignette (blank first/last row, faded
  borders) is applied for you — don't fight it.

## 2 · The point-function

```go
// <name>.go  (package fresco)
func splash<Name>At(col, row int, dx, dy, phase float64) (val, aux float64)
```

- **`val`** — brightness in `[0,1]` (pre-contrast).
- **`aux`** — hue helper in `[0,1]`: the gradient position this cell wants (`0` =
  warm `A0`, `1` = cool `A3`). It *is* the hue — make it a property of the thing
  you're drawing, not of the cell's screen address.
- **`phase`** — the *only* source of motion (it is `frame * driftPerFrame`).
- **`dx, dy`** — continuous focal-relative position; the math wants these.
  **`col, row`** — integer cell identity, for per-column effects only (rain uses
  it; most fields ignore it).

**Purity is the whole contract** — it is what makes fresco snapshot-testable: same
args → same bytes on every platform. Motion enters *only* through `phase`;
randomness *only* through the integer lattice hash (`latticeVal` /
`splashCellHash`). **No `math/rand`, no `time`/clock, no package-level mutable
state.** (`math.Sin/Cos/Pow` are fine.) Keep both returns in `[0,1]`.

**Scaling decision — make it cleanly, for one reason:**

- A field of many things drawn in **absolute cells** — a bigger pane shows *more*,
  like rain and ripple → return `splash<Name>At` directly from `splashFieldAt`.
- A **single object** that should scale to fill the pane, like the tunnel and
  galaxy → write `splash<Name>AtFor(maxD float64) splashPointFn` and return the
  closure.

Taking the `AtFor(maxD)` closure while your spacing stays absolute is a confusing
hybrid — if you only reached for `maxD` to normalize a hue sweep, find another
formulation and return the plain function.

Reuse the shared primitives: `latticeVal`, `splashCellHash`, `smoothstep`,
`clamp01`, `splashLerp`. For smooth noise / fBm, adapt the patterns in `tunnel.go`
(`splashValNoiseWrapY`, `splashTunnelFBM`) and `galaxy.go` (`galValNoise`,
`splashGalaxyTurbulence`) — there is no single shared noise call to invoke.

## 3 · The ops policy — decide it by rendering

Pass 2 splits each cell's brightness between two channels by `lumRange`:

- **glyph density** — a brighter cell picks a heavier glyph (`·`→`o`→`O`→`0`→`@`).
- **colour luminance** — a brighter cell gets a lighter colour from the gradient.

`lumRange` is the share carried by *colour luminance*. A smooth gradient wants it
**high** (`0.75`–`1`) or a fading region breaks into a scatter of dots; structure
that wants the density ramp's texture sits **below `1`**. `stars` is the fixed
twinkle over your field — say *yes* only if the field is calm and empty enough
(only ripple qualifies). Both fields are stated as an explicit literal — there is
no inherited default.

> **Do not pick `lumRange` by analogy or by arithmetic.** Render *your* field at
> `{0, 0.5, 0.75, 1}` **in colour** and compare with your eyes. The shipped values
> (ripple and galaxy at `0.75`, rain and tunnel at `1`) were each chosen from a
> rendered sweep — there is no other honest way to choose. The same goes for every
> sharpness/speed constant you introduce — and a *field-internal* constant has no
> `Options` knob to sweep live the way `lumRange` does, so lift it to a package
> `var` (or an env read) while you tune, then fold the chosen value back to a
> `const` before shipping.

## 4 · Register — every touchpoint

Miss one and a specific test names it for you (that is by design — adding a variant
is exactly when a gap is most likely and least likely to be noticed). For a variant
`veil`:

| Edit | Where | The guard that fails if you skip it |
|---|---|---|
| `Veil` const in the `iota` block, before `variantCount`, **with a doc comment** | `variant.go` | compile / `revive` `exported` lint |
| `"veil": Veil` | `variantNames`, `variant.go` | `TestSplashVariantNamesCoverAllVariants` |
| `Veil` in the returned slice | `Variants()`, `variant.go` | `TestSplashRotationCoversEveryVariant` |
| `case Veil:` → `splashOps{stars, lumRange}` | `Variant.ops()`, `variant.go` | `TestShippedVariantsOps` |
| `case Veil:` → your point-fn | `splashFieldAt`, `field.go` | **none shared — add the §5 reach guard** |
| `"veil": Veil` | `splashTestVariants()`, `helpers_test.go` | `TestSplashTestVariantsCoversEnum` |
| `Veil` row in the `want` map | `TestShippedVariantsOps`, `variant_test.go` | `require.Len(want, variantCount)` |
| new name line | `ExampleVariants` `// Output:`, `example_test.go` | the example test |
| variant table + roster prose (+ bump the "N variants" counts) | `README.md` | — (a reviewer will) |
| `### Added` entry | `CHANGELOG.md` `[Unreleased]` | — (a reviewer will) |

The exported const **needs a doc comment** matching the others — the `revive`
linter fails without it, and it is where the variant's premise is documented, so
write the good one (read the existing four for the voice).

`docs/demo.gif` is a committed recording of the *old* roster and only changes with
a `vhs` re-record (out of authoring scope) — so update the roster **prose** and
counts, but do not reword the GIF's alt text to claim it shows a variant it
doesn't.

## 5 · Tests

Once you are in `splashTestVariants()`, the shared loops cover you for free:
**determinism**, exact **`w×h` bounds** + blank borders, **frame-to-frame
animation** (`TestSplashVariantsContract`), and a coarse point-fn `[0,1]` range
(`TestSplashPointFnRange`).

**Add the reach guard** `TestSplash<Name>ReachesItsOwnField` — the one silent
failure the shared loops *miss*. A forgotten `case` in `splashFieldAt` falls
through to the fallback, so your variant renders **as rain** wearing your ops, and
determinism, bounds, and animation all still pass. Sample your field through
`splashFieldAt(v, maxD)` and assert it differs from `sample(Rain)`, mirroring
`TestSplashTunnelReachesItsOwnField`. (The §6 render-and-look gate is the other
thing that catches this — it will look like rain.)

Add a **dedicated range test** `TestSplash<Name>AtRange` in `<name>_test.go`: a
denser sweep of your own cells and phases asserting `val, aux ∈ [0,1]` and no
`NaN`, mirroring `TestSplashRippleAtRange` / `TestSplashTunnelAtRange` (ripple's
exists because its `aux` is a live `0/0`). If your field has an invariant of its
own — rain falls, ripple's drops stay inside their lattice cell — pin that too.

## 6 · Render and look — the beauty gate

**This is not optional, and structure-in-ASCII is not enough: watch it move, in
colour.** The failure this step exists to stop is judging a field by its glyph
structure (NoColor) plus its passing tests, and never once seeing the hue the
`aux` channel produces.

This skill ships the preview program at `preview/main.go`. It lives inside the
module, so `go run` it by path and the import resolves to your **local in-progress
variant** — no temp dir, no `go.mod`, no `replace` — while the dot-directory keeps
it out of `go build ./...`, the tests, and the linter. Point its `Variant` field
at yours, then:

```sh
go run ./.claude/skills/new-variant/preview        # live: watch the colour move
go run ./.claude/skills/new-variant/preview 0.75   # pin lumRange for the sweep
```

- **Inner loop:** run it, watch the *colour* animation move. Then run the sweep —
  the same command with `0`, `0.5`, `0.75`, `1` — and pick `lumRange` by eye (§3).
- **No live terminal (an agent, CI, a non-TTY)?** You still cannot skip the colour.
  Piped or redirected, the program emits **one `TrueColor` frame and exits**, so
  `go run ./.claude/skills/new-variant/preview 0.75 | …` hands you the *emitted*
  bytes to inspect: confirm SGR colour is present (`\x1b[38;2;` runs), and that the
  foreground hue **varies across the field the way your `aux` dictates** — e.g.
  sample the fg colour along the axis your `aux` maps and check it tracks the
  gradient. Reasoning the colour is right from the `aux` formula alone, without
  rendering it, is the exact shortcut this gate exists to stop.
- **Visual checklist:** Does the motion read as motion? Enough negative space (not
  a wall of glyphs)? No stuck pixels or width-1 glyph bugs? Does the hue do what
  your `aux` intended? Legible on a dark background?
- **Scale + budget:** render small (30×10) and full-window (240×60) too — a field
  that is a clean object at one size can blank out or stretch at another; run
  `go test -run=NONE -bench=RenderSplashVariants` — the 80×30 budget is ≤3 ms/frame,
  the 240×60 screensaver ≤16.7 ms.

*(If the `terminal-animations` plugin is installed, its vhs/GIF harness records a
shareable clip of this same look — an enhancement, never a requirement.)*

## Verify (the PR gates)

```sh
go build ./... && go vet ./... && gofmt -l .   # gofmt prints nothing when clean
# golangci-lint version mirrors .github/workflows/lint.yml — keep the two in sync
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run ./...
go test ./... && go test ./... -race
```

## Red flags — you are about to ship a dead field

- "I'll use `lumRange` like galaxy's `0.75`." → Render *your* sweep in colour and
  compare. Analogy is not tuning.
- "The ASCII structure looks right and the tests pass." → You have not seen the
  colour. Run the TrueColor preview.
- "I reasoned the constants out." → fresco tunes by looking. Render and look.
- "Every cell has a glyph — it's rich." → That is texture, not weather. It needs
  negative space.
- A fixed bright point / starfield over a moving field. → Stuck pixels. `stars:
  false` unless the field is genuinely still.
- Motion from anything but `phase`, or randomness from anything but the lattice
  hash. → Purity is broken; redo it.

## Common mistakes

| Symptom | Fix |
|---|---|
| A registration test fails on a name/count you never touched | You missed a §4 touchpoint — the guard is doing its job. Add the row it names. |
| `aux` renders as a flat end of the gradient | `aux` left `[0,1]` and was clamped. Rescale it into range. |
| Looks great at one size; a vague blob when small or stretched when large | Wrong scaling choice (§2): single object → `AtFor(maxD)`; absolute field → direct. |
| "Consecutive frames must differ" flakes on a sparse field | Quantized motion in too few lit cells. Make brightness a continuous function of distance-to-edge, not a rounded count (see `rain.go`). |
