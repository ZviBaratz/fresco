# Writing a variant

A *variant* is one field — rain, tunnel, ripple, galaxy. Adding one is the most
rewarding way to contribute to fresco, and this guide walks the whole path:
how a field is structured, the handful of places you register it, and the test
contract it has to satisfy. If you can write a function that maps a screen
position to a brightness, you can write a variant.

## The mental model: two passes

Every frame, fresco renders in two passes.

**Pass 1 — the field.** For each cell it calls a *point function* that returns
two numbers in `[0, 1]`:

```go
// field.go
type splashPointFn func(col, row int, dx, dy, phase float64) (val, aux float64)
```

- `val` — the raw **brightness** at that cell (before contrast and the edge
  vignette).
- `aux` — a **hue helper**: where along the palette gradient this cell wants to
  sit (`0` = warm anchor `A0`, `1` = cool anchor `A3`). See `splashColorIdx`.
- `dx, dy` — the cell's **focal-relative, aspect-corrected** position (the field
  emanates from `FocalRow`; `dy` is already scaled by the terminal's cell aspect
  so circles look round).
- `col, row` — the cell's integer identity, for per-column effects (rain uses
  it; most fields ignore it).
- `phase` — the **only** source of animation. It advances with the frame number.

A point function **must be pure**: same arguments → same result, on every
platform. Motion comes in only through `phase`; randomness comes in only through
the integer lattice hash (`splashHash` / `latticeVal`), never `math/rand` and
never a wall clock. This purity is fresco's whole contract — it's what makes the
engine snapshot-testable.

**Pass 2 — the shading.** fresco takes each cell's brightness and splits it
between two channels according to the variant's `lumRange`:

- **glyph density** — a brighter cell picks a heavier glyph from the ramp
  (`·` → `o` → `O` → `0` → `@`);
- **color luminance** — a brighter cell gets a lighter color from the gradient.

`lumRange` is the share that rides *color luminance*. `0` means brightness is
carried entirely by glyph size (a dim cell is a *small* one — a stipple); `1`
means a constant-weight glyph with all brightness in the color (this is rain,
whose glyphs have no light end to fade into). Most fields with a smooth gradient
to draw sit high (`0.75`–`1`), so a fading region stays a dim wash rather than
breaking into a scatter of dots.

You choose two things in Pass 2, both in `Variant.ops`:

- `stars` — whether the fixed twinkling starfield shows over your field. Say
  *yes* only if your field is calm and empty enough (ripple, a dark pool, is the
  one that does); a fixed point over a *moving* field reads as a stuck pixel.
- `lumRange` — the split above.

That's the whole design surface. Everything else — the contrast curve, the edge
vignette, the LUT, the run-length color emitter — is shared and handled for you.

## The five edits

Say you're adding a variant called `aurora`.

**1. Write the field generator.** Add `aurora.go` with a point function. If your
field is a single object that should scale with the pane (like the tunnel or
galaxy), take the pane radius `maxD` and return a closure; if it's drawn in
absolute cells (like rain and ripple, where a bigger pane just shows *more*),
return the function directly.

```go
// aurora.go
package fresco

func splashAuroraAt(col, row int, dx, dy, phase float64) (val, aux float64) {
    // ... your math, using phase for motion and latticeVal for any randomness.
    // Return brightness and a hue position, both in [0, 1].
    return val, aux
}
```

Reach for the shared building blocks: `latticeVal(x, y, seed)` for hashed
randomness, `splashValNoise` / the fBm helpers for smooth noise (see `tunnel.go`
and `galaxy.go`), `smoothstep`, `clamp01`.

**2. Declare the constant** in `variant.go`, inside the `iota` block, **before**
`variantCount` (which must stay last):

```go
    Galaxy
    Aurora        // <- new

    variantCount  // stays last
```

**3. Register the name** in `variantNames` (this is the vocabulary `ParseVariant`
and `String` share):

```go
var variantNames = map[string]Variant{
    "aurora": Aurora,   // <- new
    "galaxy": Galaxy,
    // ...
}
```

**4. Add it to the rotation** in `Variants()` (the order is the rotation order):

```go
func Variants() []Variant {
    return []Variant{Rain, Tunnel, Ripple, Galaxy, Aurora}
}
```

**5. Wire up both passes:**

```go
// field.go — splashFieldAt: hand back your point function
case Aurora:
    return splashAuroraAt   // or splashAuroraAtFor(maxD) for a scaling field

// variant.go — Variant.ops: your Pass-2 policy, stated as a literal
case Aurora:
    return splashOps{stars: false, lumRange: 0.75}
```

Every variant states both `ops` fields explicitly — there's no inherited
default — so you make a deliberate choice rather than picking up whatever the
zero value happens to be.

## The test contract

fresco's shared test loops enumerate every variant and assert the invariants for
each, so once `aurora` is registered you get most of your coverage for free — but
you have to opt in, and the suite makes sure you don't forget.

- Add your variant to **`splashTestVariants()`** in `helpers_test.go` (the
  hand-maintained name→variant map the contract loops walk).
- **`TestSplashTestVariantsCoversEnum`** and the `variantCount` guard fail the
  build if a variant is missing from that map — so a forgotten variant can't
  slip through untested.

The shared loops then assert, for `aurora` like every other field:

- **Determinism** — identical inputs produce byte-identical output.
- **Bounds** — exactly `h` lines of exactly `w` visible cells.
- **Animation** — consecutive frames differ (the field actually moves).

Pin the color profile in any test that inspects colored output (use the
`withColorProfile` helper, or set `Options.Profile`) — a test binary's stdout
isn't a TTY, so the ambient profile is `Ascii` and the SGR path won't run
otherwise.

Beyond the shared contract, add a small point-function test asserting your
field's returns stay in `[0, 1]`, mirroring `TestSplashTunnelAtRange` /
`TestSplashRippleAtRange`.

## Taste

The code comments are candid about *why* each field shades the way it does —
read a couple (`Variant.ops` in `variant.go` is a good start) before you tune
yours. A variant earns its place by having a premise the others don't: rain
falls, the tunnel recedes, ripple has a birth and a death, the galaxy turns.
What's yours?

When you're ready, open a [🎆 propose a new variant](https://github.com/ZviBaratz/fresco/issues/new?template=new_variant.yml)
issue (or just send the PR). Welcome aboard.
