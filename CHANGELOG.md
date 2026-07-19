# Changelog

All notable changes to fresco are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
with the pre-1.0 caveats described in
[CONTRIBUTING](CONTRIBUTING.md#versioning--releases).

## [Unreleased]

### Changed

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

- **`fresco.Galaxy`** retuned for a galaxy caught mid-turn with studded, filamentary
  arms. Two moves: its rigid pattern rotation `galRotSpd` doubles (`1.0 → 2.0`) so the
  spiral's turn — the roster's weakest motion — is plainly alive within a second or two
  of viewing while staying stately rather than spinning (phase is the field's only time
  term, so the per-frame step is far below any strobe); and the arm texture is pushed
  from softly grainy to studded (`galTurbAmp 0.62 → 0.72`, `galKnotThr 0.68 → 0.63`,
  `galKnotAmp 0.70 → 0.85`), so the star-forming knots read as distinct bright beads
  strung along filamentary arms, lighting up as the density wave sweeps through the
  static turbulence. The texture stays additive-on-peaks, so it opens no holes; the
  bright core still outshines the disk, the arm mip/anisotropy and core-finite
  guarantees are unchanged, and hue still moves without touching brightness. Rendered
  bytes change by design; determinism, bounds, and the galaxy's invariants still hold.

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

[Unreleased]: https://github.com/ZviBaratz/fresco/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/ZviBaratz/fresco/compare/v0.3.0...v1.0.0
[0.3.0]: https://github.com/ZviBaratz/fresco/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/ZviBaratz/fresco/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ZviBaratz/fresco/releases/tag/v0.1.0
