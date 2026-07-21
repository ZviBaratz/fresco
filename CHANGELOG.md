# Changelog

All notable changes to fresco are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
with the pre-1.0 caveats described in
[CONTRIBUTING](CONTRIBUTING.md#versioning--releases).

## [Unreleased]

### Added

- **`galaxy_figures_test.go` makes the galaxy's shipped figures executable.** Every
  quantitative "because" in the galaxy's comments is backed by a measurement, but
  those measurements lived only in prose, which rots — the #60/#61 work alone quoted
  three figures that had already gone stale (`118.5 → 135.5`, the ruler "odd → even",
  "~3% → ~0.7%"). The new test asserts the shipped figures the *compiled* renderer
  produces — whole-pane/arm clipping (`1.36% / 0.92%`), nucleus glyph contrast
  (`0.430`), core/arm/outskirts bead density (`58.4 / 135.5 / 67.3`), core/mid
  colour-stop density (`11.46 / 6.13`), and the lumRange `0.75 → 0.60` arm-annulus
  A/B (`93.1 → 135.5`) — each within ±5%, measured through the existing helpers and
  the public `Options.LumRange` override, with **no twin of the renderer's
  internals**. It passes on `main` and fails on a deliberate `galBulgeAmp` tweak;
  each assertion names the prose that quotes it. It does not assert the ridge/gas
  ratio (`243.8 / 26.8`), which is defined on the raw `arm` value the point function
  never returns and so cannot be measured without such a twin — that stays a
  tuning-probe finding in prose. SKILL.md §7 gains the tuning-probe (throwaway) vs
  figures-test (permanent) distinction. No rendered-byte change; an `unparam`
  nolint on two shared test helpers (now called from a second file) is the only
  edit to existing code.

- **`rain_figures_test.go` makes the rain's shipped figures executable**, the second
  application of that convention — and the one it earns hardest, since rain is the
  field whose figure already rotted *on screen*: the layer cascade shipped pinned in
  raw field units, upstream of the `smoothstep` contrast curve, and rendered `3.9 L*`
  apart against the `16.2` its comment claimed (fixed in `[1.1.0]`). The new test
  measures the cascade on the side of the curve the eye sees, two ways: it pins the
  layer head cascade (`81.9 / 65.7 / 47.4 / 29.1`, separations `16.2 / 18.3 / 18.3`)
  and the per-layer head-outshines-tail steps (`28.2 / 36.6 / 30.4 / 18.1`) through
  `rainScreenStopFor` — the real `smoothstep` and stop quantization applied to each
  layer's shipped `bright`, read off the real ramp — and it anchors those with two
  rendered-byte figures: the lit fraction at rain's quoted 96×30 pane (`27.0%`,
  guarding `rainDensity`) and the rendered proof that the mid layer's heads land in
  the L\* 60–69 band, which must dominate the 70–79 band they vacated. All within ±5%
  (reusing the galaxy file's `requireFigure`), passing on `main` and failing on the
  pre-fix `rainLayers` mid `bright 0.64 → 0.72`, which inverts the rendered bands
  exactly as `[1.1.0]` describes. It does **not** assert the one-off histogram counts
  (`239 → 53`, `50 → 193`), which do not reproduce without the demonstration's
  unrecorded pane/frame span — their robust *relation* is the test, the counts stay
  prose — and there is no lumRange A/B, since `lumRange` is a dead lever for rain
  (pinned at `1`). SKILL.md §7 gains the rain refinements (post-curve + rendered-byte
  anchor; demonstration-count-vs-relation; the `unparam` dodge). No rendered-byte
  change, and this time no edit to existing code at all — the shared `rainStopGrid`
  helper is called at a pane distinct from its other caller, so `unparam` stays quiet
  without a nolint.

## [1.1.0] - 2026-07-20

### Fixed

- **`fresco.Rain`** renders its four parallax depths as four, not three. The layer
  cascade was pinned in raw field units, upstream of the `smoothstep` contrast curve
  every value passes through before Pass 2. Smoothstep flattens approaching `1`, so
  the near and mid layers — `1.00` and `0.72`, comfortably apart in the field — both
  landed near the ramp's top and rendered **3.9 L\*** apart, against the 10-point
  separation the design calls for and the `16.2` its own comment claimed. The
  documented cascade `81.9 / 65.7 / 47.4 / 35.2` was never what reached the screen;
  that was `81.9 / 78.0 / 47.4 / 29.1`.
  `TestRainLayersSeparateInBrightness` and `TestRainHeadOutshinesItsTail` now measure
  through the curve (`rainScreenStopFor`), which makes the guard fail on the shipped
  value, and the mid layer's `bright` moves `0.72 → 0.64` to satisfy it: the rendered
  cascade is now `81.9 / 65.7 / 47.4 / 29.1`, separations `16.2 / 18.3 / 18.3`. `0.64`
  is the centre of the stop-10 plateau rather than an edge of it, so a later palette
  change cannot quietly tip it to a neighbouring stop. Measured on rendered output:
  the mid layer's heads move out of the `L* 70–79` band (239 cells → 53) and into
  `60–69` (50 → 193), opening the gap under the near layer that the parallax reads
  from. Rendered bytes change by design; determinism, bounds and rain's other
  invariants are untouched.

### Changed

- **The README demo GIF** re-recorded to tour all five variants — it previously
  showed only the pre-campaign four (rain, tunnel, ripple, galaxy) and now includes
  aurora, with every field at the re-art campaign's current look (including galaxy's
  #60 bulge). `.github/vhs/demo.tape` extends its tour by one variant accordingly.
  Docs/asset only; no code change.

- **`fresco.Galaxy`**'s bulge grades into the disk instead of saturating into a flat
  bright mass (#60). `galBulgeAmp` drops `1.0 → 0.60`. At 1.0 the pedestal sat near 1.0
  across the inner disk, so the disk it rides on — floor + arm + knot, the knot term
  doubled to `galKnotAmp 2.0` in #56 — stacked past 1 and `clamp01` flattened the sum:
  2.84% of the pane clipped (`val == 1.0`, a field-level fact upstream of Pass 2) and
  the core rendered as a solid block of two glyphs. The clamp was discarding real
  structure, not a smooth saturated field — across the clipped cells the pre-clamp
  `bulge + disk` spans `1.02..2.75` — so dropping the pedestal is what lets it show,
  rather than making room for structure never computed. At 0.60 clipping falls to 1.36%
  (the residue is sparse knot *peaks*, not a block) and the nucleus's local glyph
  contrast — `|centre − mean(8-ring)|` in ramp steps, the flat-mass metric — rises
  `0.13 → 0.43`, while the core stays the field's brightest region (core colour-stop
  density 11.5 against the mid-disk's 6.1). A soft-knee was measured and rejected:
  removing the clamp without dropping the pedestal compresses the `1.02..2.75` spread
  into a band near 1.0 the glyph ramp cannot resolve, and the nucleus goes *flatter*
  (contrast 0.008). `galBulgeAmp` was settled by rendering `{0.50, 0.55, 0.60, 0.65}`
  in colour and mono and looking: 0.65 fills the nucleus back toward the old solid look,
  0.50 shrinks it to a dim dot that no longer reads as the bright core.
  `TestSplashGalaxyRendersABrightCoreAndDimmingArms` gains a core-structure floor
  (nucleus glyph contrast `> 0.25`, placed between the measured flat `0.13` and graded
  `0.43`): its `coreLit > 0.9` and `coreDens > midDens` were both satisfied by the flat
  block they were meant to guard and so could not have caught this. Rendered bytes
  change by design; determinism, bounds and purity hold.

- **The galaxy band tests measure against the renderer's own length scale** (#61). The
  radial ruler was computed two ways — `renderField` took the vertical term as
  `max(cyFocal, h-1-cyFocal)`, `galaxy_test.go` took `cyFocal` — which differ by one on
  an even-height pane (29 vs 30 at 240×60; on odd heights the two are equal), so the
  test's `rho` ran ~0.7% large — the raw vertical term differs ~3%, diluted through the
  hypotenuse by the dominant half-width — and slid every band boundary inward. Both
  now call one `splashMaxD(w, h, focalRow)` helper (rain's and ripple's band tests
  too), so they cannot diverge again. The knot entry
  below quotes its arm-annulus bead figures on the pre-#61 ruler and the pre-#60 bulge
  (`118.5` per 1000, ridge/gas `213.8 / 19.7`, `10.9×`); on the corrected ruler and the
  shipped field they are `135.5` per 1000, and — taking ridges as `arm ≥ 0.75` and gas
  as `arm ≤ 0.30` (the raised-cosine spiral phase, before the turbulence lift) across
  the `0.35 ≤ rho < 0.60` annulus, over frames 0/30/60 at 240×60 — `243.8 / 26.8`, a
  `9.1×` ratio (the lower bulge lifts the inner-annulus gas cells, so gas beads rise a
  little more than ridge). No rendered-byte change.

- **`fresco.Tunnel`** retuned for a warmer, textured corridor. Its `lumRange` moves
  from `1` to `0.75`, so the ring texture takes the glyph-density ramp
  (`o → O → 0 → @`) and reads as a tactile, receding surface rather than a flat
  field of a single `@`; and the depth hue now relaxes to a warm base
  (`tunHueBase`) where the sampling-rate mip leaves the sweep unresolved — the near
  field and, on a real pane, most of the view — so the corridor reads as a warm
  interior receding into the cool cyan rings that survive out where the sweep
  resolves, instead of a uniform purple-blue wash. The field's brightness (the fog
  depth cue) and its mip/angular-seam guarantees are unchanged; only the Pass-2
  luminance split and the hue's unresolved fallback moved. Rendered bytes change by
  design; determinism, bounds, and the tunnel's invariants still hold.

- **`fresco.Galaxy`** caught mid-turn. Its rigid pattern rotation `galRotSpd` doubles
  (`1.0 → 2.0`) so the spiral's turn — the roster's weakest motion — is plainly alive
  within a second or two of viewing while staying stately rather than spinning (phase
  is the field's only time term, so the per-frame step is far below any strobe). The
  arm turbulence is also raised (`galTurbAmp 0.62 → 0.72`, `galKnotThr 0.68 → 0.63`,
  `galKnotAmp 0.70 → 0.85`), which adds grain and lifts the brightest turbulence peaks
  into local highlights. The texture stays additive-on-peaks, so it opens no holes; the
  bright core still outshines the disk, the arm mip/anisotropy and core-finite
  guarantees are unchanged, and hue still moves without touching brightness. Rendered
  bytes change by design; determinism, bounds, and the galaxy's invariants still hold.

  An earlier draft of this entry claimed the knots "read as distinct bright beads
  strung along filamentary arms". Measured against `galKnotAmp = 0`, that overstates
  what the change does. The knot term is not inert — it produces 12× the local maxima
  (73 against 6 across three frames, at ≥8 L\* over their neighbours) — but those
  maxima land in the bright core and the faint outskirts, not on the arms: per 1000
  lit cells the arm annulus carries `0.0` and `3.1` beads against `14.9` at the core
  and `12.3` further out, the lowest density in the field. In the glyph-density channel
  they do nothing measurable at any radius (mean glyph weight `9.24` against `9.22` in
  the arms). That draft went on to blame saturation — "the arms already sit at 8.4–9.2
  of 11 on the ramp, so an additive term there clips instead of studding" — and that
  mechanism is false too: re-measured at 240×60 over three frames, the arm annulus
  clips (`val == 1.0`) on **11 of 14,712 cells, 0.07%**, and sits at a mean `val` of
  **0.40** of 1.0, so there was headroom where the claim said there was none. Clipping
  in the field is real but lives at the bulge, not the arms (508 cells, 1.2% of the
  pane). Both the mean-glyph figure and the local-maxima distribution above are also
  single-channel readings taken before the instrument was sound. The
  actual cause is measured in the entry below. The rotation half of this entry is
  unaffected.

- **`fresco.Galaxy`**'s arms are actually studded now, and the reason the previous
  attempt could not have worked is a spatial-frequency one rather than a gain or a
  headroom one. The knots were gated on `splashGalaxyTurbulence`, an fBm carrying 47%
  of its energy at a period of 7.7 columns and 74% at 3.8 columns or wider, so the term
  brightened a *region* four to eight cells across — and a brightened region reads as a
  brighter arm, not as a knot. Measured against the term switched off, only **40%** of
  the brightest decile of such brightenings survived subtracting its own eight
  neighbours (27.6% across every cell the term brightened at all): the blob's own skirt
  lifted the background it had to stand above.

  A bigger `galKnotAmp` could only buy that back by blowing the arms out. Swept on the
  old term, arm beads per 1000 lit go `13.1 → 32.2 → 50.8 → 70.4 → 84.7` at
  `galKnotAmp 0.85 / 2 / 4 / 8 / 16`, so gain alone does eventually clear this PR's own
  guard (`TestSplashGalaxyArmsCarryKnots` floors at 40) — the flat "no amount of gain
  could have fixed it" an earlier draft of this entry claimed is false. What it costs
  is the field: arm clipping rises `0.07% → 1.90% → 6.36%` across that sweep, against
  **118.5** beads at **1.50%** clipping from the lattice below. Four times the gain for
  half the beads at worse clipping is the case for changing the term rather than
  tuning it.

  So the knots now ride their own high-frequency lattice (`galKnotFreq 0.9`,
  `galKnotPeak 0.5`, `galKnotAmp 0.85 → 2.0`), sampled in **screen cells** rather than
  in-plane ones — the in-plane axes are anisotropic by `cellAspect/cos(galInc)` = 2.17,
  so a frequency compact enough to read horizontally packs past Nyquist vertically. The
  turbulence now only *gates* them, softly (`galKnotGas 0.35`), because multiplying two
  sparse gates together starves the count faster than amplitude pays it back. Its gate
  also normalises against `galTurbCeil 0.93` — the fBm's measured maximum over 4.7M
  samples — instead of the 1.0 it never reaches; the old divisor `1-galKnotThr` assumed
  a peak of 1.0, so over that same sweep the gate topped out at **0.82** rather than 1
  and averaged **0.157** over the 15.9% of samples where it fired at all (0.025 across
  the whole sweep).

  `lumRange` drops `0.75 → 0.60` alongside, because `dens = lit^(1-lumRange)` was
  spending the glyph ramp where it could not be seen: at 0.75 no measurable cell
  anywhere rendered below glyph 4 of 11, so the disk used only the ramp's top two
  thirds. At 0.60 the floor drops to glyph 2 and the faint disk grades `·  :  ;  +`
  into dark space rather than ending abruptly. Holding the knots fixed, that is worth
  **45.8 → 76.9** beads per 1000 lit cells across the measurable pane, and **73.5 →
  118.5** inside the arm annulus. It stops at 0.60 rather than lower because 0.45
  collapses the outskirts into the scatter of `.` and `·` the luminance channel exists
  to prevent. Settled by
  rendering `{0.35, 0.45, 0.60, 0.75}` in colour and mono and looking; `galKnotFreq`
  the same way over `{0.5, 0.7, 0.9, 1.3}` (0.5 merged the beads into clumps, 1.3
  degenerated into single-cell grain).

  Measured at 240×60 over three frames, beads per 1000 lit cells — cells standing a
  full glyph step above their eight neighbours — go from `5.6` to **`118.5`** in the
  arm annulus, and the arms carry the field's *highest* density (core `17.6`,
  outskirts `67.2`) rather than its lowest. They land on the arms rather than between
  them: taking the ridges as `arm ≥ 0.75` and the inter-arm gas as `arm ≤ 0.30` within
  that annulus, `213.8` per 1000 against `19.7`, a **10.9×** ratio.
  `TestSplashGalaxyArmsCarryKnots` is the new guard for this and fails on the
  pre-change field; every previous galaxy assertion is a band mean and could not have.
  Rendered bytes change by design; determinism, bounds, purity, the arm
  mip/anisotropy and core-finite guarantees all still hold.

- **`fresco.Ripple`** retuned to make its interference pattern — the field's jewel —
  read, in three moves. Its crest amplitude `rippleAmp` drops (`0.85 → 0.65`) to open
  headroom under the clamp: at `0.85` the render curve already sent a lone crest to
  within a hair of a doubled one, so the bright nodes where rings add did not stand
  out; the lower value opens that gap and lights the constructive lattice against the
  lone rings (the pool dims a little, the trade that keeps the fixed stars from
  outshining the water). Its carrier `rippleCyc` rises (`1.5 → 1.8`), deepening the
  packet's trough (`41% → 54%` of the crest) so the cancellation nodes where two rings
  meet out of phase darken and `|sum|` reads as a pattern rather than a stack of
  circles, and giving each drop a second, fainter concentric wavefront — the train a
  real drop throws. And the ring's hue is rebased onto its *visible* life
  (`rippleHueOpen`): a drop is a filled disc for its first ~38% and only an expanding
  ring after, so the raw age-hue spent its whole warm end on the disc and left the ring
  the cool `60%` of the gradient — rebasing hands the expanding ring the whole
  warm→cool journey, so a ring's age reads across its life. Separately the starfield —
  ripple's alone among the variants — is thinned (`starThreshold 0.986 → 0.992`) from a
  competing sparkle layer into a faint still sky the rings sit in front of. The `|sum|`
  interference, the exact 3×3×2 spawn window, the compact-support packet, and purity
  are unchanged. Rendered bytes change by design; determinism, bounds, and ripple's
  invariants still hold.

- **`fresco.Rain`** retuned so its two signatures — the bright head and the depth —
  land, in two coordinated moves. Its head lobe tightens (`rainHeadR 4.5 → 2.9`),
  which is a fix as much as a retune: `rainTailAmp` buys a ~28-point L\* gap under
  the head, but a 4.5-unit lobe reached 1.68 rows and sat in that gap itself, so the
  cell one row behind a head landed at L\* 71.8 against the head's 81.9 — a 10-point
  step, and a head that rendered as a two-cell blob rather than an edge. At `2.9`
  that step is 34.5 and the darkness the tail had already paid for finally reaches
  the head; because the lobe is symmetric, it also cuts the glow *ahead* of a falling
  head from ~1.7 rows to under a cell. And a fourth parallax depth is added
  (`rainLayers` `[3] → [4]`, with `rainDensity 0.62 → 0.54` so four compounding
  layers do not fill the pane), because three left a hole in the middle of the
  brightness histogram and the eye sorted them into "near" and "far" instead of
  reading a recession; the four now land at L\* `81.9 / 65.7 / 47.4 / 35.2`, and the
  room for the extra layer came from the darker field the first move opens. The
  anti-blink guarantee is untouched — the ramp's top stop is reachable only from the
  `rainHeadFlat` plateau, and the rendered count of top-stop cells is unchanged
  across the whole `rainHeadR` sweep. `lumRange` stays at `1`: rain's render branch
  consumes only the luminance channel and discards the density one, so lowering it
  would not hand the tail a glyph ramp at all — it would merely lift the whole field
  and close the head/tail gap, which the rendered sweep confirmed. The stream-train
  purity, the fall speed, the tail-length window, and the bespoke luminance ramp are
  unchanged. Rendered bytes change by design; determinism, bounds, and rain's
  invariants still hold.

## [1.0.0] - 2026-07-18

The **1.0 release.** The public API is now stable and committed to under Semantic
Versioning: `Render`, `AppendRender`, `Options`, `Palette` (+ `Validate`), the
`Variant` set (+ `Variants`, `ParseVariant`, `String`), and the `ColorProfile`
enum. No exported identifier will be renamed, removed, or retyped before a
`2.0.0`; new variants, options, profiles, and methods may still be added, since
those are additive. The core contract is unchanged and now permanent: `Render`
and `AppendRender` are pure over their inputs and emit exactly `h` lines of
exactly `w` visible cells, never erroring or panicking on any `Options` — a
malformed `Palette` degrades to documented fallbacks, and `Palette.Validate` is
the opt-in check. The surface was validated against its real downstream consumer
(atrium bumped to `v0.3.0` with zero code changes) and given a final last-look
review before freezing; see [`docs/api-review-v1.0.md`](docs/api-review-v1.0.md).

### Added

- **`fresco.Aurora`** — a fifth splash variant: northern-lights curtains that drift
  slowly sideways over dark sky and snake as they go, the hue sliding warm→cool with
  altitude. An absolute field like `Rain` and `Ripple` (a bigger pane shows more
  sky), shaded at `lumRange` 0.75 so the filament cores keep the density ramp while
  the soft halo rides the colour. Registered in `Variants()`, `ParseVariant`, and
  the rotation.

## [0.3.0] - 2026-07-17

The **"Refine & prove"** release. It settles the pre-1.0 API — a fresco-owned
`ColorProfile` enum, up-front `Palette.Validate`, and a buffer-reusing
`AppendRender` — then proves the hot path with an allocation pass and validates
the whole surface against its real downstream consumer (atrium).

### Added

- **`AppendRender(dst []byte, w, h, frame int, opts Options) []byte`** — a
  buffer-friendly render path that appends the frame to a caller-owned slice, so
  a per-frame loop can reuse one buffer (`buf = AppendRender(buf[:0], …)`)
  instead of allocating a fresh string every tick (#17). `Render` becomes a thin,
  byte-identical wrapper over it. Purely additive: measured at 120×40 truecolor,
  reusing the buffer removes the per-frame output allocation (≈172 KB → 82 KB).
- **`Palette.Validate() error`** — an advisory, opt-in check that reports any
  anchor that is not a canonical hex colour (`"#rgb"` or `"#rrggbb"`), naming
  every offending field (#18). `Render` is unchanged: it still never errors or
  panics on a malformed palette — each bad anchor degrades to a documented
  fallback, so the exactly-h×w-cells contract always holds. `Validate` is
  deliberately stricter than the renderer's parser, so it flags typos (a missing
  `#`, a wrong length, trailing garbage) the renderer would otherwise paint.

### Changed

- **`Options.Profile` is now a fresco-owned `ColorProfile` enum** instead of
  `*termenv.Profile`, so pinning colour depth no longer requires importing
  `termenv` (#15). The zero value, `Auto`, auto-detects the terminal exactly as
  an unset (`nil`) profile did before; pin `TrueColor`, `ANSI256`, `ANSI16`, or
  `NoColor` for a fixed depth. This also settles the options-ergonomics review
  (#16): `Options` stays a plain struct — functional options were rejected as
  per-frame allocation churn for a 60fps render call — and `LumRange` stays a
  `*float64` because `0` is a meaningful value, so no sentinel can mean "unset".
  **Breaking:** `Profile: &p` becomes `Profile: p` (a `ColorProfile` value);
  callers pinning depth swap `termenv.Ascii`→`NoColor`, `termenv.ANSI`→`ANSI16`,
  and drop the `termenv` import.

### Performance

- **Hot-path allocation pass** (#21) — the render loop no longer materialises the
  field into two per-frame `[]float64` buffers (the "Pass 1"/"Pass 2" split is
  fused: each cell is evaluated inline as it is emitted), and the per-palette LUT
  cache is keyed by a comparable struct instead of a freshly joined string on every
  call. Output is byte-identical (verified by a 16,896-frame SHA-256 sweep across
  variants × profiles × frames × sizes × luminance ranges). On a reused
  `AppendRender` buffer this takes the warm hot path from 4 allocations/frame
  (≈82 KB at 120×40) to 1 (`Rain`/`Ripple`: to **zero**); via `Render`, per-frame
  `B/op` roughly halves. Wall-clock gains are modest and variant-dependent (≈8–12%
  for the lighter variants, flat for tunnel and the shaded path) — the pass targets
  garbage, not latency. Baselines and method are recorded in
  [`docs/perf.md`](docs/perf.md).

## [0.2.0] - 2026-07-16

The first **"Open the doors"** release. It makes no change to the rendering
engine or its public API — this milestone is entirely about turning a good
repository into a well-run, contributable open-source project: green CI, a
linter, community health files, contributor docs, and a demo you can see at a
glance.

### Added

- **Continuous integration** — a GitHub Actions workflow running `go build`,
  `go vet`, `gofmt`, and the test suite across a Go 1.25 / 1.26 × Ubuntu / macOS
  / Windows matrix, with a dedicated race-detector job (#1), plus a
  `golangci-lint` configuration and lint job (#2).
- **Dependabot** for the Go module and GitHub Actions ecosystems (#3).
- **Community health files** — a Code of Conduct (#5), a security policy (#6),
  bug-report / feature-request issue forms and a pull-request template (#7), a
  dedicated "propose a new variant" form (#8), and `CODEOWNERS` (#9).
- **Contributor documentation** — `CONTRIBUTING.md` (#4), a variant-authoring
  guide at [`docs/variants.md`](docs/variants.md) (#11), and the project roadmap
  at [`docs/ROADMAP.md`](docs/ROADMAP.md).
- **Examples and expanded tests** — runnable `ExampleRender` and
  `ExampleParseVariant` (#12), fuzz targets `FuzzParseVariant` and `FuzzRender`
  (#20), and an explicit determinism property test (#22).
- **A visible demo** — an animated GIF at the top of the README, with its
  reproducible [`vhs`](https://github.com/charmbracelet/vhs) `.tape` source
  committed (#10).
- **README badges** — CI status and Codecov coverage, beside a link to the
  roadmap (#13).
- **This changelog** and a pre-1.0 versioning policy (#14).

## [0.1.0] - 2026-07-16

### Added

- Initial release: the pure `(width, height, frame, Options) → ANSI` rendering
  engine with four variants (`Rain`, `Tunnel`, `Ripple`, `Galaxy`), the
  `Options` / `Palette` API, automatic terminal colour-profile detection, and
  the `cmd/fresco-demo` runnable demo.

[Unreleased]: https://github.com/ZviBaratz/fresco/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/ZviBaratz/fresco/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/ZviBaratz/fresco/compare/v0.3.0...v1.0.0
[0.3.0]: https://github.com/ZviBaratz/fresco/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/ZviBaratz/fresco/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ZviBaratz/fresco/releases/tag/v0.1.0
