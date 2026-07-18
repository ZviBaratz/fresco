# Pre-1.0 API review — "last look" before freezing the surface

Reviewed at `v0.3.0` (`v0.3.0-4-g2b996e0`). Scope: the exported surface's
design for 1.0-readiness — not bugs, not security (assumed covered). Evidence is
the package source + godoc, atrium's real usage (`ui/splash.go`,
`ui/splash_variants.go`), `CONTRIBUTING` (versioning), `CHANGELOG`, `docs/perf.md`,
`docs/ROADMAP.md`.

**Verdict up front: cut `v1.0.0` as-is.** No blocking changes. The two "consider"
items below are both *recommend-against* on balance; neither should hold the tag.

> **Update — surface to freeze is 5 variants.** Between this review and the release,
> `Aurora` (a fifth variant, dev-letter "j") merged to `main` (PR #45, `0a60c1f`).
> Its entire exported delta is one new `Variant` constant plus its registration in
> `variantNames` / `Variants()` / `ParseVariant`; `aurora.go` exports nothing at top
> level. That is exactly the additive extension §4 predicted — the verdict below is
> unchanged, and Aurora landing cleanly mid-review is that verdict proven live a
> second time. Everywhere this doc lists four variants, read five (`… Galaxy, Aurora`).

---

## 1. Confirmed exported surface

Derived from `go doc -all .` + grep for exported identifiers (methods included).
The complete set — nothing else is exported:

```go
// Entry points
func Render(w, h, frame int, opts Options) string
func AppendRender(dst []byte, w, h, frame int, opts Options) []byte

// Config struct
type Options struct {
    Palette  Palette
    Variant  Variant
    FocalRow int
    LumRange *float64
    Profile  ColorProfile
}

// Colour input
type Palette struct { A0, A1, A2, A3 string; Highlight string }
func (p Palette) Validate() error

// Variant enum
type Variant int
const ( Rain Variant = iota; Tunnel; Ripple; Galaxy )   // variantCount is unexported
func Variants() []Variant
func ParseVariant(s string) (Variant, bool)
func (v Variant) String() string

// Colour-depth enum
type ColorProfile int
const ( Auto ColorProfile = iota; TrueColor; ANSI256; ANSI16; NoColor )
```

That is 2 funcs, 3 types, 2 methods, 2 enum helpers, 9 consts. `variantCount`,
`splashOps`, `(ColorProfile).resolve`, and the whole field/LUT internals are
correctly unexported.

Note: the sole downstream consumer (atrium, on `fresco v0.3.0`) drives
`Render(Options{Palette, Variant, FocalRow, LumRange})` at 60fps and **never sets
`Profile`** and **never calls `AppendRender`** (it composites the returned string
with an ANSI-aware overlay). It does use `Variants`, `ParseVariant`, `Variant`,
and `Palette.Validate` (in tests).

---

## 2. Per-element verdict

| Element | Verdict | One-line rationale (evidence) |
|---|---|---|
| `Render` | **freeze** | `w,h,frame int` + `opts` is the idiomatic dims-first shape; purity + exactly-h×w-cells is documented and SHA-swept (`docs/perf.md`, 16,896-frame digest). |
| `AppendRender` | **freeze** | Textbook Go `Append*` idiom (cf. `strconv.AppendInt`, `fmt.Appendf`); the alloc-free core that `Render` wraps byte-identically (`field.go:184`). A consumer can't add it themselves — needs internal buffer access. |
| `Render`↔`AppendRender` (`string` vs `[]byte`) | **freeze** | The consistent, not inconsistent, pairing — string form is the ergonomic default, append form the zero-alloc primitive. Matches stdlib convention exactly. |
| `Options` (plain struct) | **freeze** | Settled (functional options rejected — per-frame closure churn at 60fps). Keyed-literal construction ⇒ new fields are additive. Zero value renders validly. |
| `Options.FocalRow int` | **freeze** | Clean zero-value handling: every non-negative int is a literal row; the otherwise-wasted negative half carries "centre" — no pointer/sentinel needed (`field.go:209`). |
| `Options.LumRange *float64` | **freeze** | Settled — `0` is meaningful (density carries all brightness), so no `float64` sentinel can mean "unset" (`variant.go:100`). Pointer is the correct "optional value" encoding. |
| `Options.Profile ColorProfile` | **freeze** | Settled — `Auto=0` = auto-detect (a real-pane caller's want); pinning makes output pure. termenv's zero is `TrueColor`, so its type couldn't carry "unset" (`profile.go`). |
| `Palette` struct (strings) | **freeze** | Hex strings (not a `color.Color`) keep termenv/colorful off the surface and accept theme tokens verbatim (atrium passes theme strings straight in, `splash.go:117`). Zero value degrades to fallbacks. |
| `Palette` field names `A0..A3` | **freeze** *(consider `Anchor0..3`)* | Terse but correct: they're *ordered anchor positions*, not fixed hues — a semantic name (`Pink`) would lie for a green→teal palette. Godoc carries the meaning. Only naming judgment call; see §3. |
| `Palette.Validate() error` | **freeze** | The right 1.0 split: `Render` never errors/degrades silently; `Validate` is the opt-in, deliberately-stricter typo check that names every offender (`lut.go:52`). |
| `Variant` enum + `Rain/Tunnel/Ripple/Galaxy` | **freeze** | `Rain=0` is the documented fallback ⇒ zero value is safe. New variants append at the end (before unexported `variantCount`) — additive (§ extensibility). |
| `Variants() []Variant` | **freeze** | Returns a fresh copy so callers can't mutate the pool (`variant.go:67`); atrium captures it once, not per-frame. |
| `ParseVariant(s) (Variant, bool)` | **freeze** | Comma-ok idiom; names-not-numbers is the committed contract (config stores the name). atrium layers dev-letter aliases on top (`splash_variants.go:88`). |
| `(Variant) String()` | **freeze** | Implements `Stringer`; `"unknown"` for out-of-set. Round-trips with `ParseVariant` for the shipped set. |
| `ColorProfile` enum + 5 consts | **freeze** | Names mirror termenv's vocabulary without importing it; `Auto=0` safe default; additive at the tail. |

---

## 3. Recommended pre-1.0 changes

**None required.** Two items were weighed and rejected as changes — recorded so the
decision is on the record:

**(a) Force keyed `Options` literals with an unexported guard field — REJECTED.**
Adding `_ struct{}` (or similar) would make positional `Options{pal, v, ...}`
construction a compile error, hard-guaranteeing that future field additions never
break anyone. But `go vet`'s composites check already flags cross-package unkeyed
literals; the Go ecosystem already treats struct-field-addition as non-breaking;
and the guard is an unusual wart on an otherwise clean struct. YAGNI. Keep
`Options` clean.

**(b) Rename `Palette.A0..A3` → `Anchor0..3` — REJECTED (lean).** This is the one
change that *cannot* be made additively later (a rename is breaking), so it's a
genuine last-cheap-moment call. But `A0..A3` correctly conveys "indexed gradient
anchor," reads clearly in the keyed literal atrium writes, and is fully documented;
`Anchor0..3` buys marginal self-documentation for real verbosity and no semantic
gain. If the maintainer weights self-documenting field names highly, do it *now* or
never — otherwise freeze. Recommendation: **freeze.**

---

## 4. Overall call

**(a) Cut `v1.0.0` as-is.**

The decisive fact is extensibility: **every axis fresco is likely to grow along is
additive in Go and needs no surface change now.**

- *New variant* (the roadmap's headline growth path, #8/#11): append a const before
  `variantCount`; existing ordinals are unchanged and are not the contract (names
  are). `Variants`/`ParseVariant`/`String` gain an entry. Verified non-breaking.
- *New option*: add a field to `Options`; keyed literals (the documented,
  atrium-used form) are unaffected. Verified non-breaking.
- *New colour profile*: append a const. Non-breaking.
- *New behaviour on existing types* (e.g. `Variant` gaining `MarshalText`, a default-
  palette constructor): adding methods/funcs is non-breaking. So even the "nice to
  have someday" surface can safely wait for 1.x.

The only moves 1.0 forecloses are *renaming/removing/retyping existing
identifiers* — and this review found nothing that warrants one. The zero values are
all safe (`Options{}` renders; `Auto`, `Rain`, empty `Palette` all degrade to
documented fallbacks), the value-vs-pointer choices are each justified by a real
semantic (`LumRange` pointer, everything else a value), and the error/degradation
split (`Render` never fails; `Validate` is opt-in) and purity contract are
documented, tested, and worth promising.

**The compatibility promise a 1.0 makes.** For all of `v1.x`: `Render` and
`AppendRender` stay pure over their inputs and emit exactly `h` lines of exactly `w`
visible cells (empty string / no append on a degenerate pane), never erroring or
panicking on any `Options` — including a malformed `Palette`, which continues to
degrade to its documented fallbacks rather than failing. `Render` and
`AppendRender` stay byte-identical for identical inputs. The variant *names*
`{rain, tunnel, ripple, galaxy}` and the `ColorProfile` names keep resolving and
keep their meaning; `Options{}`, `Auto`, and `Rain` keep their current
zero-value behaviour. New variants, options, profiles, and helper methods may be
added; nothing exported is renamed, removed, or retyped without a `v2.0.0`. That is
a promise this surface can keep — freeze it.
