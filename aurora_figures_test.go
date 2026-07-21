package fresco

import (
	"math"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

// This file is the aurora's shipped-figure guard, the fifth and last application of
// the convention galaxy_figures_test.go established (SKILL.md §7). It reuses that
// file's requireFigure/galFigTol (±5%).
//
// Aurora's existing tests (aurora_test.go) cover the point-fn contract, the
// fallback-field wiring and the altitude envelope — invariants, no quantitative
// figure. What lives only in prose is the pair of numbers the field is actually
// tuned around, and this file makes both executable:
//
//   - The DRIFT TRAP, aurora's signature constraint. The whole curtain pattern
//     translates under auroraDrift, and a translating band-limited signal crawls
//     once its frequency passes the grid's Nyquist limit — the sideways analogue of
//     the tunnel's wagon-wheeling rings. So the finest octave must stay resolvable:
//     auroraFreqX·2^(auroraOct−1) = 0.40 cycles/cell, below the 0.5 Nyquist limit,
//     where a fourth octave (0.80) would alias. This is the reason the stack is three
//     deep and the base frequency is low (auroraFBM), and pinning it fails the moment
//     someone raises auroraFreqX or adds an octave.
//   - The fBm WINDOW placement. auroraLo/auroraHi are fitted to the *measured* fBm
//     distribution, not the [0,1] nominal range, so the figures that justify them are
//     properties of the compiled auroraFBM: its median, and the percentile auroraHi
//     lands on. The window must straddle the upper third — auroraLo just above the
//     median (a curtain is a band of sky, not a scatter of dots) and auroraHi near the
//     90th percentile (the brightest filaments reach full luminance).
//
// Measuring corrected two stale prose numbers (the convention working): the fBm
// clusters around 0.50, not the 0.53 the comment claimed, and its tail tops out near
// 0.79 at the 99th percentile (a thin tail reaches ~0.94), not the "near 0.82" the
// comment gave. auroraLo's comment is corrected to match; the window-placement
// reasoning it states was right all along.
//
// The fBm distribution is measured over a fixed dense lattice sweep, so it is
// deterministic — the same median and percentile every run, no sampling noise for
// the ±5% band to absorb.
//
// Figure → where it is quoted (change both together when a retune moves one):
//
//	finest-octave frequency (0.40 cyc/cell)   aurora.go (auroraOct/auroraFreqX, auroraFBM)
//	fBm median / p90; window placement         aurora.go (auroraLo/auroraHi comment)

// auroraFBMStats samples the compiled auroraFBM over a fixed dense lattice sweep and
// returns its median and 90th percentile. The sweep spans ~90×80 frequency units
// (many lattice periods) at irrational-ish steps so samples never align to the
// integer lattice — the intrinsic distribution of the shipped fBm, deterministically.
func auroraFBMStats(t *testing.T) (median, p90 float64) {
	t.Helper()
	const n = 700
	vals := make([]float64, 0, n*n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			vals = append(vals, auroraFBM(float64(i)*0.131, float64(j)*0.117))
		}
	}
	sort.Float64s(vals)
	return vals[len(vals)/2], vals[int(0.9*float64(len(vals)-1))]
}

// TestAuroraFigureDriftNyquist pins aurora's defining constraint: the finest octave
// sits at 0.40 cycles/cell, below the 0.5 Nyquist limit, so no octave crawls as the
// curtains drift. It is the trap the whole field is shaped around (three octaves, low
// base frequency); a fourth octave would land at 0.80 and alias the moment auroraDrift
// moved it.
func TestAuroraFigureDriftNyquist(t *testing.T) {
	finest := auroraFreqX * math.Pow(2, float64(auroraOct-1))
	requireFigure(t, "finest-octave frequency (cycles/cell)", 0.40, finest,
		"aurora.go auroraOct/auroraFreqX comments; auroraFBM drift-trap")
	require.Lessf(t, finest, 0.5,
		"the finest octave (%.3f cyc/cell) must stay under Nyquist (0.5) or it crawls as "+
			"the curtains drift — aurora's signature failure mode", finest)
	fourth := auroraFreqX * math.Pow(2, float64(auroraOct))
	require.Greaterf(t, fourth, 0.5,
		"a fourth octave would land at %.2f cyc/cell, past Nyquist — which is exactly why "+
			"the stack is three deep (auroraFBM)", fourth)
}

// TestAuroraFigureFBMWindow pins that the curtain window is fitted to the measured
// fBm distribution: the fBm clusters around 0.50, its 90th percentile sits near 0.67,
// and the window straddles the upper third — auroraLo just above the median (a real
// band of sky, not a scatter of dots) and auroraHi near the 90th percentile (the
// brightest filaments reach full luminance).
func TestAuroraFigureFBMWindow(t *testing.T) {
	median, p90 := auroraFBMStats(t)
	requireFigure(t, "auroraFBM median", 0.498, median, "aurora.go auroraLo/auroraHi comment")
	requireFigure(t, "auroraFBM 90th percentile", 0.673, p90, "aurora.go auroraLo/auroraHi comment")

	require.Greaterf(t, auroraLo, median,
		"auroraLo (%.2f) must sit above the fBm median (%.3f), or the curtains are more sky "+
			"than not", auroraLo, median)
	require.Lessf(t, auroraLo-median, 0.10,
		"auroraLo (%.2f) must sit *just* above the median (%.3f) — too far above and a curtain "+
			"thins to a scatter of dots", auroraLo, median)
	require.Lessf(t, math.Abs(auroraHi-p90), 0.05,
		"auroraHi (%.2f) must sit near the fBm's 90th percentile (%.3f) so the brightest "+
			"filaments — and only those — saturate", auroraHi, p90)
}
