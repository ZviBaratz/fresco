# Performance notes

This file records fresco's render-path allocation profile and the perf pass that
established it (roadmap #21), so a later change that regresses the hot path is
visible against a number rather than a memory.

fresco's render is pure — identical inputs produce identical bytes — so every
figure here was taken with output held byte-for-byte constant. The #21 change was
verified identical by a SHA-256 sweep over variants × colour profiles × frames ×
pane sizes × luminance ranges (16,896 frames): the digest is unchanged before and
after. Performance work here never trades a pixel.

## How to reproduce

The benchmarks live in the package and run with `-benchmem`:

```
go test -run '^$' -bench 'BenchmarkAppendRenderReuse|BenchmarkRenderString|BenchmarkRenderSplashVariants|BenchmarkRenderSplashShaded' -benchmem -count=6 .
```

- `BenchmarkAppendRenderReuse` is the animation hot path: one buffer reused across
  frames (`AppendRender(buf[:0], …)`), so it reports the *per-frame* garbage a
  60fps caller actually generates, with the output allocation already removed.
- `BenchmarkRenderString` is the same frame via `Render`, which allocates a fresh
  string each call — that output allocation is inherent to the API.
- `BenchmarkRenderSplashVariants` / `…Shaded` sweep the four variants and the
  luminance channel at 80×30 (the preview pane) and 240×60 (the full-window
  screensaver). Colour profile is forced to truecolor — a bench binary's stdout is
  not a TTY, so the emitter would otherwise be timed with nothing to emit.

Compare two revisions with [`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat),
`-count=6` per side. The numbers below were taken on an Intel Core Ultra 7 258V
(`-8`); absolute values are machine-specific, the deltas are what travel.

## The #21 pass: allocations on the warm hot path

The roadmap asked to profile `Render`'s allocations (string building, LUT access)
and cut avoidable garbage. After #17 (`AppendRender`) removed the per-frame output
allocation, a caller reusing one buffer still paid **4 allocations per frame**.
Profiling (a `-memprofile` plus `testing.AllocsPerRun` on the suspects) attributed
them, and — the point of measuring — one long-suspected source turned out already
free:

1. **Two `[]float64` field buffers** (real, ~99% of the bytes). The renderer
   evaluated the whole field into `vals`/`aux` slices (a "Pass 1"), then walked
   them to emit ("Pass 2"). At 120×40 that is two 40 KB allocations per frame.
2. **One LUT cache key** (real, small). `splashLUTFor` built its map key with
   `strings.Join([...])` on *every* call — including the cache hit that every
   steady frame takes — a throwaway string on the render path for no lookup gain.
3. **The two `[]rune` ramp conversions were *not* an allocation source.** They were
   already stack-allocated by escape analysis (`AllocsPerRun` on the non-escaping
   local: 0), so the buffered path's 4 allocs were `2×field + LUT-key + variant
   closure`, not the field buffers plus ramps as assumed.

### What changed

- **Fused the two passes.** The field is now evaluated one cell at a time inside
  the emit loop, so the `vals`/`aux` buffers are gone entirely — no pooling, no new
  API. Pass 2 already read the field strictly in order, so the fusion is
  mechanical. (Bonus: fully-blank border rows now skip the field eval, which the
  buffered pass computed and discarded.)
- **Keyed the LUT cache by a comparable struct** (`lutKey`) instead of a joined
  string, so a cache hit hashes in place and allocates nothing (verified 1 → 0).
- **Hoisted the ramps** to package-level `[]rune` vars. Not an allocation win (see
  above); it moves the per-call UTF-8 decode of the ramp glyphs to a one-time init
  and makes the no-alloc property independent of the optimizer.

### Result — `BenchmarkAppendRenderReuse` (120×40 truecolor, tunnel, reused buffer)

| | B/op | allocs/op | sec/op |
|---|---:|---:|---:|
| before | 82,080 | 4 | 885.9 µs |
| after | 112 | 1 | 851.8 µs |
| Δ | **−99.86%** | **−75%** | **−3.85%** (p=0.015) |

On the reused-buffer path the field buffers and LUT key are gone. The single
remaining allocation is the per-frame **variant closure** for `Tunnel`/`Galaxy`
(`splashTunnelAtFor`/`splashGalaxyAtFor` capture the pane's length scale). `Rain`
and `Ripple` use plain function values that capture nothing, so on a reused buffer
they now allocate **zero times per frame**.

### Allocations per frame, by variant

`BenchmarkRenderSplashVariants` renders via `Render`, so its count includes the one
inherent output-string allocation; the reused-buffer path (above) is one lower.

| variant (80×30 & 240×60) | allocs/op before → after | B/op Δ |
|---|---|---|
| rain, ripple | 4 → **1** | −45% (80×30), −48% (240×60) |
| tunnel, galaxy | 5 → **2** | −45% / −48% |

The `B/op` roughly halves because the two field buffers are no longer allocated —
e.g. rain 240×60 goes 496 KiB → 256 KiB. `RenderString` (the `Render` path) drops
168 KiB → 88 KiB, 5 → 2 allocs.

## Wall-clock: modest, and that is the expected shape

The pass is an allocation pass, not a latency pass — the emitter (bracketing runs
with baked SGR affixes) dominates frame time and is untouched, so removing the
field-buffer round-trip moves the clock only a little. A `benchstat` of the same
benchmarks (`-count=6` per side, back-to-back) shows:

| benchmark | sec/op Δ |
|---|---|
| RenderSplashVariants/80×30 rain, ripple, galaxy | −10% to −12% (p≤0.002) |
| RenderSplashVariants/240×60 rain, ripple | −8% to −10% (p≤0.004) |
| RenderSplashVariants tunnel (both sizes), all `…Shaded`, RenderString | no significant change |
| AppendRenderReuse | −3.85% (p=0.015) |
| **geomean** | **−4.4%** |

The lighter variants (rain/ripple/galaxy) gain the most because the buffer
round-trip was a larger share of their cheaper frames; tunnel's per-cell field math
and the shaded path's `Log`/`Exp` swamp it, so those hold flat. The real win a
sustained 60fps caller sees is the garbage that no longer has to be collected, not
these per-frame microseconds.

> Caution when re-measuring: single `-count=1` before/after runs taken minutes
> apart drift with CPU frequency/thermal state and can show a fictitious ~2×. Use
> `benchstat` with `-count≥6` per side, run back-to-back, and trust the `sec/op`
> deltas with a p-value, not raw wall-clock between two separate invocations.

## Known remaining allocations (candidates, not regressions)

- **The `Tunnel`/`Galaxy` per-frame closure** (1 alloc/frame). Removable by caching
  the closure by `(variant, maxD)` or passing the length scale to a plain function,
  but both add machinery for one small allocation that only two variants pay; left
  as a deliberate follow-up.
- **`Render`'s output string** (1 alloc/frame, by design). A caller that cares uses
  `AppendRender` with a reused buffer — that is what it is for.
