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
go run ./cmd/fresco-demo    # watch it move
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
| `cmd/fresco-demo` | The runnable demo that cycles every variant. |

## Reporting bugs & proposing ideas

Use the [issue templates](https://github.com/ZviBaratz/fresco/issues/new/choose).
A new field goes through the **🎆 propose a new variant** form. For anything
open-ended, a [discussion](https://github.com/ZviBaratz/fresco/discussions) is a
good place to start.

See the [roadmap](docs/ROADMAP.md) for where the project is headed and for
issues labelled [`good first issue`](https://github.com/ZviBaratz/fresco/labels/good%20first%20issue).
