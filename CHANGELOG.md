# Changelog

All notable changes to fresco are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
with the pre-1.0 caveats described in
[CONTRIBUTING](CONTRIBUTING.md#versioning--releases).

## [Unreleased]

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

[Unreleased]: https://github.com/ZviBaratz/fresco/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/ZviBaratz/fresco/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/ZviBaratz/fresco/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ZviBaratz/fresco/releases/tag/v0.1.0
