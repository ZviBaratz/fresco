# fresco roadmap

fresco is a young project with a strong core: a pure, deterministic
`(width, height, frame) → ANSI` engine, ~97% test coverage, `go vet` clean, and
carefully documented internals. What it does *not* yet have is the scaffolding
that turns a good repository into an inviting, well-run open-source project — CI,
community files, a demo people can see at a glance, and a frictionless path for
contributing a new variant.

This roadmap closes that gap. It is organized into two milestones so the
welcoming work ships first and stands on its own, and the deeper engineering
follows on a stable base.

> This document is the canonical plan. Each item below becomes one self-contained
> GitHub issue. If an issue and this file ever disagree, the issue wins — but
> please keep them in sync.

## How this roadmap works

- **Two milestones.** `v0.2.0 — Open the doors` makes fresco genuinely
  contributable and fun to discover. `v0.3.0 — Refine & prove` does the
  pre-1.0 API refinement, performance, and real-world validation.
- **Labels.** Issues carry an `area:*` label (`ci`, `docs`, `community`, `api`,
  `testing`, `release`) and, where relevant, `good first issue` / `help wanted`.
- **Claiming work.** Comment on an issue to claim it. Small, self-contained
  issues are deliberate: pick one, open a PR, done.
- **The bar.** New code keeps the existing standard — tests for behavior,
  `go vet` and the linter clean, and the pure-function contract (identical
  inputs → identical bytes, exactly `h`×`w` cells) preserved.

## Baseline (as of this roadmap)

| Signal | State |
|---|---|
| Test coverage | ~96.9% of statements |
| `go vet ./...` | clean |
| Benchmarks | present (`BenchmarkRenderSplash*`) |
| CI | **none** |
| Linter | **none configured** |
| Community files | **none** (no CONTRIBUTING / CoC / SECURITY / templates) |
| Demo media | **none** (README tells you to run it) |
| Release | `v0.1.0` **tag only** — no GitHub Release, no CHANGELOG |
| Downstream | [atrium](https://github.com/ZviBaratz/atrium) still ships an in-tree copy; not yet consuming the module |

---

## Milestone `v0.2.0` — Open the doors

Make fresco a first-class, welcoming open-source repository: green CI, a linter,
community health files, a demo people can *see*, and an obvious, rewarding path
to contribute a new variant.

### CI & automation

**1. GitHub Actions CI — test + vet matrix** · `area:ci`
Run `go test ./... -race -cover` and `go vet ./...` on every push and PR across a
Go (1.25, 1.26) × OS (ubuntu, macOS, windows) matrix. Upload coverage. This is
the foundation: it unblocks the README badges (#13) and lets every later PR be
gated on green.
*Done when:* a `.github/workflows/ci.yml` runs on PRs and the badge is live.

**2. golangci-lint config + lint job** · `area:ci`
Add `.golangci.yml` (suggested: `gofumpt`, `govet`, `staticcheck`, `errcheck`,
`revive`, `misspell`, `unconvert`, `unparam`) and a CI job that runs it. Fix or
explicitly `//nolint` any findings in the same PR so main stays clean.
*Done when:* `golangci-lint run` is green in CI.

**3. Dependabot** · `area:ci` · `good first issue`
Add `.github/dependabot.yml` for the `gomod` and `github-actions` ecosystems
(weekly), so dependency and workflow bumps arrive as reviewable PRs.

### Community health

**4. `CONTRIBUTING.md`** · `area:community`
How to build, test (`go test ./...`), and lint; the PR conventions this repo
already follows (conventional-commit-style titles, one logical change per PR,
tests required); and a short "how the pieces fit" map pointing newcomers at
`field.go` (core + `Render`), `variant.go` (the vocabulary + shade policy), and
the per-field files (`rain.go`, `tunnel.go`, `ripple.go`, `galaxy.go`). Links to
the variant-authoring guide (#11).

**5. `CODE_OF_CONDUCT.md`** · `area:community` · `good first issue`
Adopt the Contributor Covenant with the maintainer contact filled in.

**6. `SECURITY.md`** · `area:community` · `good first issue`
A short policy: supported versions and how to report privately. Surface is small
(a pure rendering library), but the file signals the project is run seriously.

**7. Issue & PR templates** · `area:community`
`.github/ISSUE_TEMPLATE/` with **bug report** and **feature request** forms, plus
a `PULL_REQUEST_TEMPLATE.md` reflecting the CONTRIBUTING checklist.

**8. "Propose a new variant" issue template** · `area:community`
A dedicated template that walks a would-be contributor through the pitch: the
motion premise, the shading policy (`stars` / `lumRange`), and references. This
is the front door for the contribution fresco most wants; pairs with #11.

**9. `CODEOWNERS`** · `area:community` · `good first issue`
Route reviews to the maintainer(s).

### Demo & docs

**10. Killer animated demo in the README** · `area:docs`
Record the variants *moving* with [`charmbracelet/vhs`](https://github.com/charmbracelet/vhs)
and embed the GIF at the top of the README. **Commit the `.tape` script** so the
demo is reproducible and regenerable (candidate future CI check). This is the
single biggest "stop and look" factor for a generative-art project — today the
README only tells people to run it.
*Done when:* a looping GIF of the variants sits above the fold in the README,
with its `.tape` source in the repo.

**11. Variant-authoring guide** · `area:docs`
A `docs/variants.md` walkthrough of how a variant is built and registered:
- the two-pass model (field generator → `ops`/shade policy);
- the exact registration touchpoints — add to `variantNames`, `Variants()`, and
  a `case` in `Variant.ops()`;
- the **testing contract**: the `variantCount` guard and the `splashTestVariants()`
  contract loops (determinism, exact `w`×`h` bounds, frame-to-frame animation)
  that every variant must satisfy.
Ideally accompanied by a minimal reference variant a contributor can copy.
*Done when:* someone can add a working variant using only this guide.

**12. Runnable `Example` tests** · `area:docs` · `area:testing`
Add `ExampleRender` and `ExampleParseVariant` (with `// Output:` where feasible).
They render on the pkg.go.dev landing page and are compile-checked, so the
usage examples can never drift from the API.

**13. README badges & polish** · `area:docs` · `good first issue`
Once CI is green (#1), add CI and coverage badges beside the existing
Go-Reference / License badges, and link this roadmap. *Depends on #1.*

### Release

**14. CHANGELOG + real release + versioning policy** · `area:release`
Add `CHANGELOG.md` ([Keep a Changelog](https://keepachangelog.com/)), cut a
proper GitHub Release for the milestone (retroactively note `v0.1.0`), and write
a brief semver / pre-1.0 stability statement (what `v0.x` promises, what would
trigger `v1.0`). Note in CONTRIBUTING how releases are cut.

---

## Milestone `v0.3.0` — Refine & prove

With a stable, contributable base, refine the public surface (fresco is pre-1.0,
so breaking changes are cheap and appropriate — **each gets its own reasoned
issue**), tighten performance and testing, and prove the whole extraction against
a real consumer.

### API refinement

**15. Reduce the termenv coupling in `Options`** · `area:api`
`Options.Profile *termenv.Profile` forces every consumer to import
`github.com/muesli/termenv` just to pin color depth. Evaluate exposing a small
fresco-owned `ColorProfile` enum (or an interface) so termenv becomes an
implementation detail. Weigh the ergonomic win against the mapping cost.
*Breaking.* Present options and a recommendation in the issue before coding.

**16. Options ergonomics** · `area:api`
Evaluate functional options (`fresco.WithVariant`, `fresco.WithPalette`,
`fresco.WithLumRange(0.75)`) versus today's struct-with-`*float64`
(`LumRange *float64` as "unset = use variant default"). Decide deliberately;
the pointer-as-optional pattern works but reads awkwardly. Coordinate with #15.

**17. Zero-allocation render path (additive)** · `area:api` · `area:testing`
`Render` returns a freshly allocated string every call; a 60fps loop churns the
GC. Add an **additive** `RenderTo(w io.Writer, …)` or `AppendRender(dst []byte, …)`
so callers can reuse a buffer. Benchmark the before/after (ties to #21). Purely
additive — no break.

**18. Palette input validation** · `area:api`
Define and document behavior for malformed hex in `Palette` (e.g. `"#zz"` or
`""`): error, fallback, or panic — pick one and make it explicit. Consider a
`Palette.Validate()` or a constructor. Add tests for the boundary cases.

**19. Evaluate the mixed-case module path** · `area:api` · `area:release`
The module path `github.com/ZviBaratz/fresco` uses a mixed-case owner segment;
Go convention and the module proxy prefer all-lowercase import paths. Investigate
the practical impact and whether a lowercase path is worth the one-time break.
May well land as *document-and-keep* — the issue is to make that call on purpose.

### Testing & performance

**20. Fuzz targets** · `area:testing` · `good first issue`
Add `FuzzParseVariant` (arbitrary strings never panic) and `FuzzRender`
(arbitrary `w, h, frame` and options never panic, always yield exactly `h` lines
of `w` cells, and stay deterministic). A pure engine is an ideal fuzz subject.

**21. Allocation benchmarks + hot-path perf pass** · `area:testing`
Run the existing benchmarks with `-benchmem`, profile `Render`'s allocations
(string building, LUT access), and reduce avoidable garbage in the hot path.
Record baseline numbers in the issue / a `docs/perf.md` so regressions are
visible. Pairs with #17.

**22. Race + determinism guards in CI** · `area:testing` · `area:ci`
Ensure `-race` runs in CI (folds into #1) and add an explicit
determinism/property test asserting identical bytes across repeated `Render`
calls with pinned `Options.Profile`. Cheap insurance for the project's core
promise.

### Prove the extraction

**23. Re-wire atrium to consume the module** · `area:api`
Replace atrium's in-tree `splash/` package with a dependency on
`github.com/ZviBaratz/fresco`, adapting its scene composition to the public API.
This is the real-world validation of the surface refined in #15–#18 and the true
finish line for the extraction. Cross-repo: the work lives in
[atrium](https://github.com/ZviBaratz/atrium); this issue tracks it from fresco's
side and captures any API friction it surfaces.

---

## Explicitly out of scope (for now)

Deferred deliberately to keep the roadmap focused; revisit after `v0.3.0`:

- A WASM / web gallery demo.
- A plugin system for community variants loaded at runtime (the in-repo
  contribution path in #8/#11 is the near-term answer).
- An asciinema showcase beyond the single README GIF.
- Funding / sponsorship files.
