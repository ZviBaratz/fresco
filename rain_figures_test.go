package fresco

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
)

// This file is the rain's shipped-figure guard, the second application of the
// convention galaxy_figures_test.go established (see SKILL.md §7, "A shipped
// figure is a test, not a comment"): a shipped figure that gets cited is a claim
// about the compiled renderer, so it belongs in a test that fails when the
// renderer drifts under it — not only in a sentence that stays green while going
// wrong. It reuses that file's requireFigure/galFigTol (±5%, same rationale: the
// field is deterministic, so the margin is refactor headroom, narrow enough that
// structural drift fails).
//
// Rain is where this convention earns its keep the hardest, because rain is the
// field whose figure already rotted ON SCREEN. The layer cascade shipped pinned in
// raw field units — upstream of the smoothstep contrast curve every value crosses
// before Pass 2 — and smoothstep flattens approaching 1, so the near and mid layers
// (1.00 and 0.72) both landed near the ramp's top and rendered 3.9 L* apart while
// the comment claimed 16.2 (CHANGELOG [1.1.0], Fixed). The documented cascade was
// never what reached the screen. So rain's figures test measures the cascade on the
// side of the curve the eye sees, two independent ways:
//
//   - TestRainFigureCascade / TestRainFigureHeadOutshinesTail assert the ramp-level
//     figures through rainScreenStopFor — which applies the REAL smoothstep and the
//     REAL stop quantization to a layer's shipped bright — and read the REAL ramp
//     (rainRampLum). This is the post-curve measurement whose absence shipped the
//     bug. It is a layer-table figure by necessity: layers combine by max, so a
//     single layer's head cannot be isolated in the emitted pane.
//   - TestRainFigureMidLayerReachesScreen closes that gap from the other side. It
//     reads the RENDERED bytes (rainStopGrid) and asserts the cascade actually
//     arrives there: the L* 60–69 band the fix moved the mid heads into dominates
//     the L* 70–79 band they vacated. This is the anchor that keeps rainScreenStopFor
//     honest — if Pass 2's real quantization ever diverged from the reconstruction,
//     the first two tests would stay green and this one would not.
//
// Two figures are deliberately NOT asserted, for the same reason galaxy's ridge/gas
// stays prose — a figure you cannot reproduce on the compiled renderer without a
// twin, or without a harness the codebase does not carry, stays a sentence:
//
//   - The exact rendered histogram counts (239 → 53 in L* 70–79, 50 → 193 in 60–69;
//     CHANGELOG [1.1.0]) were a one-off demonstration of the fix at a pane and frame
//     span the prose does not record, and they do not reproduce blind (at 96×30 the
//     bands are ~5 and ~51). What reproduces robustly is the RELATION they show —
//     60–69 dominates 70–79 once the mid layer lands below the near one — so that
//     relation is the test and the counts stay prose.
//   - There is no lumRange A/B here as there is for the galaxy. lumRange is a dead
//     lever for rain: its render branch pins lumRange at 1 (variant.go), because the
//     luminance ramp already owns rain's brightness channel end to end.
//
// Every rain figure below is pane-independent (the cascade and head/tail are pure
// layer-table × ramp) except the two rendered-output tests, which use 96×30 — rain's
// own quoted pane (rainDensity comment). rainStopGrid needs a TrueColor profile to
// decode stops, and is called here at 96×30, distinct from its existing 120×40
// caller, so the unparam linter sees w/h take more than one value and stays quiet —
// no nolint needed, unlike the galaxy file's shared (w, h) helpers.
//
// Figure → where it is quoted (change both together when a retune moves one):
//
//	layer head cascade L* / separations   rain.go (rainLayers); CHANGELOG [1.1.0]
//	per-layer head-outshines-tail L*       rain.go (rainTailAmp, rainHeadR); CHANGELOG [1.1.0]; rain_test.go (15 floor)
//	lit fraction @ 96×30                    rain.go (rainDensity)
//	mid layer lands below near (rendered)   rain.go (rainLayers); CHANGELOG [1.1.0]

// rainFigFrames are the frames the rendered-output figures (lit fraction, the
// mid-layer band) are averaged over, matching the galaxy file's cadence. The
// cascade and head/tail figures take no frame — they are the layer table through
// the ramp, identical at every phase.
func rainFigFrames() []int { return []int{0, 30, 60} }

// TestRainFigureCascade pins the depth cue itself: each layer's head, measured
// where the eye sees it (after smoothstep, on the ramp), must land at the shipped
// L* and the shipped step below the layer in front of it. These are the exact
// figures whose raw-units twin shipped the bug the whole file is built around —
// the near→mid step read 3.9 on screen against the 16.2 its comment claimed
// (CHANGELOG [1.1.0]). TestRainLayersSeparateInBrightness floors each step above
// 10; this pins the values it clears.
func TestRainFigureCascade(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	pal := splashTestPalette()

	heads := make([]float64, len(rainLayers))
	for li, L := range rainLayers {
		heads[li] = rainRampLum(t, pal, rainScreenStopFor(1.0*L.bright))
	}
	wantHeads := []float64{81.9, 65.7, 47.4, 29.1}
	for li := range rainLayers {
		requireFigure(t, "layer head L*", wantHeads[li], heads[li], "rain.go rainLayers comment; CHANGELOG [1.1.0]")
	}

	wantSep := []float64{16.2, 18.3, 18.3}
	for i, want := range wantSep {
		requireFigure(t, "layer separation L*", want, heads[i]-heads[i+1], "rain.go rainLayers comment; CHANGELOG [1.1.0]")
	}
}

// TestRainFigureHeadOutshinesTail pins the step that makes a head read as a head:
// the L* gap between a layer's head and the brightest cell of its own tail
// (rainTailAmp × bright), measured through the curve. rainTailAmp buys this gap and
// rainHeadR decides whether it reaches the head (rain.go), so the figures guard
// both constants at once. TestRainHeadOutshinesItsTail floors each gap above 15;
// this pins the values.
func TestRainFigureHeadOutshinesTail(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	pal := splashTestPalette()

	want := []float64{28.2, 36.6, 30.4, 18.1}
	for li, L := range rainLayers {
		head := rainRampLum(t, pal, rainScreenStopFor(1.0*L.bright))
		tail := rainRampLum(t, pal, rainScreenStopFor(rainTailAmp*L.bright))
		requireFigure(t, "head-outshines-tail L*", want[li], head-tail,
			"rain.go rainTailAmp/rainHeadR comments; rain_test.go 15 floor; CHANGELOG [1.1.0]")
	}
}

// TestRainFigureLitFraction pins the density the four-layer field settles at:
// rainDensity 0.54 across four compounding layers lands the lit fraction at 27.0%
// on rain's quoted 96×30 pane (rain.go), just under where three layers at 0.62 had
// it. It counts rendered ink (non-blank cells) — the whole-pane density fact — and
// is orthogonal to the cascade: the mid-layer mutation that inverts the band test
// below barely moves this one, which is exactly why both are needed.
func TestRainFigureLitFraction(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	pal := splashTestPalette()
	const w, h = 96, 30

	var pct float64
	frames := rainFigFrames()
	for _, frame := range frames {
		out := ansi.Strip(renderSplashField(w, h, frame, pal, centeredFocalRow(h), Rain))
		lit, total := 0, 0
		for _, line := range strings.Split(out, "\n") {
			for _, r := range line {
				total++
				if r != ' ' {
					lit++
				}
			}
		}
		pct += 100 * float64(lit) / float64(total)
	}
	requireFigure(t, "lit fraction %% @ 96×30", 27.0, pct/float64(len(frames)), "rain.go rainDensity comment")
}

// TestRainFigureMidLayerReachesScreen is the anchor the PR #50 story demands: it
// reads the RENDERED bytes and asserts the cascade actually arrives on screen. The
// fix moved the mid layer's heads out of the L* 70–79 band and down into 60–69
// (CHANGELOG [1.1.0]); on rendered output that lower band must therefore dominate
// the one the mid heads vacated. The exact demonstration counts (53 / 193) do not
// reproduce blind and stay prose (see the file header); the relation does, and it
// is decisive — the pre-fix bright 0.72 inverts it (60–69 collapses, 70–79 fills),
// which no ramp-level reconstruction (the first two tests) could catch if Pass 2's
// real quantization ever drifted from it.
func TestRainFigureMidLayerReachesScreen(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	pal := splashTestPalette()
	const w, h = 96, 30

	var band6069, band7079 int
	frames := rainFigFrames()
	for _, frame := range frames {
		grid := rainStopGrid(t, w, h, frame, pal)
		for _, row := range grid {
			for _, stop := range row {
				if stop < 0 {
					continue
				}
				switch L := rainRampLum(t, pal, stop); {
				case L >= 70 && L < 80:
					band7079++
				case L >= 60 && L < 70:
					band6069++
				}
			}
		}
	}
	nf := len(frames)
	mean6069, mean7079 := float64(band6069)/float64(nf), float64(band7079)/float64(nf)
	t.Logf("rendered @ 96×30: L* 60–69 mean=%.1f  70–79 mean=%.1f", mean6069, mean7079)

	require.Greaterf(t, mean6069, 30.0,
		"the L* 60–69 band is nearly empty (%.1f cells): the mid layer's heads are not "+
			"reaching the screen where the fix put them (CHANGELOG [1.1.0])", mean6069)
	require.Greaterf(t, mean6069, 3*mean7079,
		"the L* 60–69 band (%.1f) must dominate the 70–79 band (%.1f) the mid heads "+
			"vacated; an inverted ratio is the pre-fix cascade back on screen — the mid "+
			"layer collapsed into the near one (CHANGELOG [1.1.0])", mean6069, mean7079)
}
