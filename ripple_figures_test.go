package fresco

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// This file is the ripple's shipped-figure guard, the fourth application of the
// convention galaxy_figures_test.go established (SKILL.md §7). It reuses that
// file's requireFigure/galFigTol (±5%).
//
// Ripple already carries a thorough invariant suite (ripple_test.go): the packet
// is compact and signed, the 3×3×2 window is exact, drops stay in their cell and
// epoch, the field is |sum|, the fade is not a threshold. Those pin the packet's
// *shape class* and the sum's *rules* — mostly as floors (the trough clears 30% of
// the crest; the worst row alignment clears 75%). What they leave unpinned are the
// exact packet-shape figures the comments quote, and ripple is the field that
// proves why that gap matters: its row-pitch capture comment carried `87.3%` for
// three PRs after rippleCyc moved 1.5 → 1.8, because the figure was folded to a
// literal `cos(0.15pi)` that outlived the constant it came from. A floor at 75%
// never noticed. An executable figure would have.
//
// So this file pins the packet's shipped shape, all measured on the compiled
// rippleDropWave (a pure function of the packet constants — no twin, no rendered
// pane needed, because these figures are defined on the packet, not on Pass 2):
//
//	row-pitch capture         82.8%  — the worst row-grid alignment still sees this
//	                                    much of the crest (the anti-blink margin)
//	packet trough / crest      54%   — how hard two rings cancel, at the shipped cyc
//	packet trough location    x=0.48 — where that trough sits (not the carrier's 0.56)
//	ring-open fraction         0.38  — rippleW/rippleMaxR, when a disc becomes a ring
//	lone crest amplitude       0.65  — a single crest's peak, the headroom two need
//
// Two things are deliberately NOT asserted here.
//
//   - The rippleCyc sweep (trough/crest 15/29/41/54/61% at cyc 1.0/1.25/1.5/1.8/2)
//     is a TUNING PROBE — it varies a const to pick a value. This file pins only the
//     shipped point (cyc 1.8 → 54%), never re-runs the sweep, exactly as the
//     convention draws the line.
//   - The candidate-skip distribution (49.7% unborn, 43.7% out-of-packet, 6.6%
//     contribute, over 1.3M candidates at 240×60) is an internal branch statistic
//     the point function never exposes; measuring it would mean replaying the birth
//     draws in a twin. It stays a prose measurement, as galaxy's ridge/gas does.
//
// The emitted-byte side of ripple's design (fade-not-threshold, drops-render-the-
// same-everywhere) is already asserted on rendered output by ripple_test.go, so
// this file stays at the packet level where its figures live.
//
// Figure → where it is quoted (change both together when a retune moves one):
//
//	row-pitch capture (82.8%)     ripple.go (rippleW); TestSplashRippleCrestSurvivesTheRowPitch (75% floor)
//	trough/crest (54%), x=0.48    ripple.go (rippleCyc); TestSplashRipplePacketIsCompactAndSigned (30% floor)
//	ring-open fraction (0.38)     ripple.go (rippleHueOpen, rippleLife, rippleDropWave)
//	lone crest amplitude (0.65)   ripple.go (rippleAmp)

// rippleFigDrop addresses one drop — cell (0,0), epoch 0 — the way every ripple
// packet test does, returning its position and birth so a figure can be sampled
// along a known ray from a known crest.
func rippleFigDrop() (px, py, ts float64) {
	px, py = rippleDropPos(0, 0, 0)
	ts = rippleDropBirth(0, 0, 0)
	return px, py, ts
}

// TestRippleFigureRowPitchCapture pins the anti-blink margin: the worst alignment
// of the cellAspect row grid against a ring still lands on 82.8% of the crest's
// true peak. This is the figure that rotted to a stale 87.3% for three PRs (see the
// file header and rippleW's comment); measured here on the packet exactly as
// TestSplashRippleCrestSurvivesTheRowPitch measures it for its 75% floor, but pinned
// to the value rather than bounded below it. Age cancels out of the ratio, so it is
// the same figure at every age — asserted at one.
func TestRippleFigureRowPitchCapture(t *testing.T) {
	px, py, ts := rippleFigDrop()
	const age = 1.4
	phase := ts + age

	truePeak := 0.0
	for d := 0.0; d < rippleMaxR+rippleW; d += 0.01 {
		c, _ := rippleDropWave(0, 0, 0, px, py+d, phase)
		truePeak = math.Max(truePeak, math.Abs(c))
	}
	require.Greater(t, truePeak, 0.0)

	worst := math.Inf(1)
	for off := 0.0; off < cellAspect; off += 0.1 {
		seen := 0.0
		for d := off; d < rippleMaxR+rippleW; d += cellAspect {
			c, _ := rippleDropWave(0, 0, 0, px, py+d, phase)
			seen = math.Max(seen, math.Abs(c))
		}
		worst = math.Min(worst, seen/truePeak)
	}
	requireFigure(t, "worst row-pitch capture", 0.828, worst,
		"ripple.go rippleW comment; TestSplashRippleCrestSurvivesTheRowPitch 75% floor")
}

// TestRippleFigurePacketShape pins how hard the shipped packet cancels: at
// rippleCyc 1.8 its trough reaches 54% of its crest, and that trough sits at
// x = 0.48 of the half-width — earlier than the carrier's own trough at 0.56,
// because the (1-x²)² envelope is already falling there. TestSplashRipplePacketIsCompactAndSigned
// only floors the trough at 30% of the crest; this pins the value the choice landed on.
func TestRippleFigurePacketShape(t *testing.T) {
	px, py, ts := rippleFigDrop()
	const age = 1.4
	phase := ts + age
	rr := rippleSpeed * age

	crest, trough, xAtTrough := 0.0, 0.0, 0.0
	for d := rr - rippleW; d <= rr+rippleW; d += 0.005 {
		c, _ := rippleDropWave(0, 0, 0, px+d, py, phase)
		if c > crest {
			crest = c
		}
		if c < trough {
			trough, xAtTrough = c, (d-rr)/rippleW
		}
	}
	require.Greater(t, crest, 0.0)
	requireFigure(t, "packet trough / crest", 0.54, -trough/crest,
		"ripple.go rippleCyc comment; TestSplashRipplePacketIsCompactAndSigned 30% floor")
	requireFigure(t, "packet trough location (x/rippleW)", 0.48, xAtTrough,
		"ripple.go rippleCyc comment (min at x=0.48, not the carrier's 0.56)")
}

// TestRippleFigureRingGeometry pins two shipped scalars the field is shaped around:
// the age-fraction at which a drop stops being a filled disc and becomes an
// expanding ring (rippleHueOpen = rippleW/rippleMaxR = 0.38), and a lone crest's
// peak amplitude (0.65 = rippleAmp) — deliberately under 1 so that two rings adding
// have the headroom to clip and read as constructive interference. The lone crest is
// measured post-flash (the flash is allowed past the clamp); that it peaks near 0.65
// and stays below 1 is the headroom claim made executable.
func TestRippleFigureRingGeometry(t *testing.T) {
	requireFigure(t, "ring-open fraction (rippleHueOpen)", 0.38, rippleHueOpen,
		"ripple.go rippleHueOpen/rippleLife comments; rippleDropWave fade")

	px, py, ts := rippleFigDrop()
	crestPeak := 0.0
	for a := rippleFlashT + 0.01; a <= rippleLife; a += 0.005 {
		phase := ts + a
		for d := 0.0; d < rippleMaxR+rippleW; d += 0.02 {
			c, _ := rippleDropWave(0, 0, 0, px+d, py, phase)
			crestPeak = math.Max(crestPeak, math.Abs(c))
		}
	}
	requireFigure(t, "lone post-flash crest amplitude", 0.65, crestPeak, "ripple.go rippleAmp comment")
	require.Lessf(t, crestPeak, 1.0,
		"a lone crest (%.3f) must stay under the clamp, or two rings adding have no "+
			"headroom to read as brighter — the whole point of rippleAmp < 1", crestPeak)
}
