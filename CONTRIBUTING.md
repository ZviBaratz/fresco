# Contributing to fresco

Thanks for your interest! fresco is a small, focused engine, and it's meant to
be a pleasant thing to hack on. The single most welcome contribution is a **new
variant** — a new field for the roster. There's a whole guide for that:
[`docs/variants.md`](docs/variants.md).

By participating you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting set up

fresco requires **Go 1.25+** and has no build tooling beyond the Go toolchain.

```sh
git clone https://github.com/ZviBaratz/fresco
cd fresco
go test ./...        # run the suite
go run ./cmd/fresco    # watch it move
```

## Before you open a PR

Run these and make sure they're clean:

```sh
go build ./...
go vet ./...
gofmt -l .           # prints nothing when formatting is correct
go test ./...        # ~97% coverage today; keep new behavior covered
```

- **One logical change per PR.** Small, reviewable diffs merge faster.
- **Conventional-commit-style titles** are appreciated and match the history,
  e.g. `feat(variant): aurora — a drifting polar curtain`,
  `fix(ripple): rings no longer clip at the pane edge`, `docs: …`.
- **Tests are required for behavior.** fresco's core promise is that it's a
  *pure* function — the same `(width, height, frame, Options)` always yields the
  same bytes. Tests lean on that: they're deterministic and largely
  snapshot/property based. New behavior should come with coverage, and must not
  break the contract (see below).

## The contract every change must keep

`Render` is pure over its inputs and returns exactly `h` lines of exactly `w`
visible cells (or `""` for a degenerate pane). The test suite enforces this for
every variant through shared loops — determinism, exact `w`×`h` bounds, and
frame-to-frame animation. If you add a variant, those loops pick it up
automatically once it's registered; see the guide.

To keep output deterministic in tests, pin the color profile (the suite's
`withColorProfile` helper does this) or set `Options.Profile` — otherwise output
depends on the ambient terminal.

## How the pieces fit

| File | Responsibility |
|------|----------------|
| `field.go` | The engine core: `Render`, the Pass-1 field walk (`splashEvalField`), the point-function type (`splashPointFn`), and the per-variant field dispatch (`splashFieldAt`). |
| `variant.go` | The variant vocabulary (`Variant`, `Variants()`, `ParseVariant`, `String`) and each variant's Pass-2 policy (`Variant.ops` → `stars`, `lumRange`). |
| `shade.go` | Pass 2: turning a raw field value into a glyph + color (`splashShade`, `shadeAt`). |
| `lut.go` | The baked gradient look-up table (per palette + color profile). |
| `rain.go`, `tunnel.go`, `ripple.go`, `galaxy.go` | The individual field generators. |
| `cmd/fresco` | The screensaver CLI: the impure shell (terminal size, clock, signals, keys, teardown) around the pure engine, split into a testable core and a thin driver. |

## Versioning & releases

fresco follows [Semantic Versioning](https://semver.org/), with the usual pre-1.0
latitude.

**What `v0.x` promises.** While the major version is `0`, the public API is still
settling:

- **Patch** releases (`0.y.Z`) are bug fixes, docs, and internal changes — no API
  changes.
- **Minor** releases (`0.Y.0`) may add features and, where justified, make a
  breaking API change. Pre-1.0, breaking changes ride minor bumps; each one gets
  its own reasoned issue and a `Changed` / `Removed` entry in the changelog.
- The **core contract holds at every version**: `Render` is pure over its inputs
  and returns exactly `h` lines of exactly `w` visible cells.

**What triggers `v1.0`.** Once the public surface — `Options`, `Palette`, the
variant set, and the render entry points — has settled and been validated against
a real downstream consumer (see the `v0.3.0` milestone in the
[roadmap](docs/ROADMAP.md)), we'll commit to it with `v1.0.0` and the standard
SemVer compatibility guarantee.

**How a release is cut.** Releases are tagged from `main`:

1. Move the accumulated `## [Unreleased]` entries in [`CHANGELOG.md`](CHANGELOG.md)
   into a new `## [x.y.z] - YYYY-MM-DD` section, and update the compare links at
   the foot of the file.
2. Land that as a PR (e.g. `docs: release vX.Y.Z`) and merge to `main`.
3. Tag and push: `git tag -a vX.Y.Z -m "vX.Y.Z" && git push origin vX.Y.Z`.
4. Pushing the tag triggers the **Release** workflow
   ([`.github/workflows/release.yml`](.github/workflows/release.yml)): GoReleaser
   builds the `cmd/fresco` binaries for linux/darwin/windows × amd64/arm64,
   attaches the archives (each bundling `LICENSE`, `README.md`, and `CHANGELOG.md`)
   and a `checksums.txt`, and publishes the GitHub Release for the tag with its
   notes taken from that version's `CHANGELOG.md` section. There is **no** manual
   `gh release create` — watch the run under **Actions** and confirm the release
   and its assets. GoReleaser errors if a release for the tag already exists, so
   never re-tag or re-push a published version; cut a new patch/minor instead. The
   published version is baked into the binary (`fresco --version`).

## The module path

The module path is `github.com/ZviBaratz/fresco`, with a mixed-case owner
segment, and that is **deliberate** (roadmap #19). Go import paths are
case-sensitive, but mixed case is fully supported everywhere it matters: the
module proxy and the on-disk cache case-encode uppercase letters (fresco is
stored under `github.com/!zvi!baratz/fresco`), so `go get github.com/ZviBaratz/fresco`
and the import statement work verbatim on every platform, including
case-insensitive filesystems. It is a well-trodden path — `Azure/azure-sdk-for-go`
is a widely used mixed-case module.

We keep it because it matches the GitHub owner's canonical casing, and moving to
a lowercase path would buy only cosmetic conformance at a real cost: the proxy and
checksum database key releases by exact path, so a rename splits the version
history (`v0.1.0`/`v0.2.0` would live under the old path) and breaks every
existing import. Lowercasing only the `go.mod` `module` directive while leaving
the repository at `ZviBaratz/fresco` would technically resolve — GitHub serves
clone URLs case-insensitively — but it splits the history just the same and adds a
permanent repo-vs-import casing mismatch, so it is not worth doing either. If you
import fresco, copy the path as written above.

## Reporting bugs & proposing ideas

Use the [issue templates](https://github.com/ZviBaratz/fresco/issues/new/choose).
A new field goes through the **🎆 propose a new variant** form. For anything
open-ended, a [discussion](https://github.com/ZviBaratz/fresco/discussions) is a
good place to start.

See the [roadmap](docs/ROADMAP.md) for where the project is headed and for
issues labelled [`good first issue`](https://github.com/ZviBaratz/fresco/labels/good%20first%20issue).
