package fresco

import (
	"math"
	"testing"

	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
)

// This file is the tunnel's shipped-figure guard, the third application of the
// convention galaxy_figures_test.go established (SKILL.md §7): a shipped figure
// that gets cited is a claim about the compiled renderer, so it belongs in a test
// that fails when the renderer drifts under it, not only in a sentence. It reuses
// that file's requireFigure/galFigTol (±5%).
//
// The tunnel already carries a thorough, mutation-tested invariant suite
// (tunnel_test.go): the angular seam, the per-octave and anisotropic mip, the
// finite vanishing point, depth-reads-as-brightness. Those pin *relationships*
// (ratios, monotonicity, continuity). What they leave in prose is the tunnel's
// one quantitative geometry claim — how deep the fog's black core reaches — which
// tunnel.go states two ways and this file makes executable:
//
//   - The ANALYTIC half-lit radius. The fog is scale-free in rho = r/maxD:
//     fog = tunFogGain·rho/(rho + tunFogA/tunRefD), so it reaches 0.5 at
//     rho = 0.5·(tunFogA/tunRefD)/(tunFogGain − 0.5) = 0.1838 — 18.4% of maxD at
//     every pane size (tunnel.go, splashTunnelAtFor). This is computed from the
//     shipped fog constants, so it fails the moment tunFogA, tunRefD or tunFogGain
//     move under the quoted 18.4% — the derivation itself becomes the test.
//   - The RENDERED half-peak crossing, and its scale invariance. On the emitted
//     luminance stops the radial profile (mean stop vs rho, the fog's black core
//     counted as 0) crosses half its peak at ≈15% of maxD, the same curve across
//     96×30 / 160×44 / 240×60 / 300×80 — just inside the analytic 18.4%, the
//     contrast curve and quantization taking their cut. This is measured on the
//     bytes a terminal receives (shadeStopGrid), the only place the claim is real.
//
// Two things are deliberately NOT asserted here.
//
//   - The ring-spacing product (tunnel.go: "the product is the knob", quoting a
//     good value of 400 against a bad 72) is STALE: the shipped tunDepthK·tunFreqU
//     is 70, at the bad end, because tunFreqU was later lowered to 0.35 for the
//     dead-zone reason its own comment gives and the per-octave mip took over rim
//     legibility. But rendered ring *spatial frequency* is not a figure this
//     convention can cleanly pin (it varies as r² across the pane by design), so
//     the discrepancy is surfaced rather than tested — a figure you cannot measure
//     without a twin, or that has no single value, stays prose (cf. galaxy's
//     ridge/gas). It is flagged in this PR's description.
//   - The rendered "16%" tunnel.go quoted was optimistic; the honest stop-profile
//     crossing is ≈15% (13.9–15.4% across the four panes), so this PR corrects the
//     prose to match what the test measures. That is the convention working: the
//     figure and the sentence now move together.
//
// Figure → where it is quoted (change both together when a retune moves one):
//
//	analytic fog half-lit radius (18.4% maxD)   tunnel.go (tunFogA/tunFogGain, splashTunnelAtFor)
//	rendered half-peak crossing, scale-invariant tunnel.go (splashTunnelAtFor); TestSplashTunnelRendersDepthAsLuminance

// tunnelFigPanes are the four pane sizes the tunnel's scale invariance is asserted
// across — a >3× span of maxD, so a figure that held only at one size fails here.
func tunnelFigPanes() [][2]int { return [][2]int{{96, 30}, {160, 44}, {240, 60}, {300, 80}} }

// tunnelRadialHalfPeak renders the tunnel and returns where its radial luminance
// profile crosses half its peak, as a fraction of maxD. The profile is the mean
// emitted stop binned by rho with the fog's black core (blank cells) counted as 0,
// sampled at phase 0 so the vanishing point sits exactly on the pane centre and rho
// is measured from there. w/h vary across callers (see tunnelFigPanes), so this is
// the scale-invariance probe, not a fixed-pane one.
func tunnelRadialHalfPeak(t *testing.T, w, h int, pal Palette) (cross, peak float64) {
	t.Helper()
	maxD := splashMaxD(w, h, centeredFocalRow(h))
	cx, cyFocal := float64(w-1)/2, float64((h-1)/2)
	const nb = 60
	sum := make([]float64, nb)
	cnt := make([]int, nb)
	stops, _ := shadeStopGrid(t, w, h, 0, pal, Tunnel)
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			dx, dy := float64(col)-cx, (float64(row)-cyFocal)*cellAspect
			b := int(math.Hypot(dx, dy) / maxD * nb)
			if b >= nb {
				b = nb - 1
			}
			if stops[row][col] >= 0 {
				sum[b] += float64(stops[row][col])
			}
			cnt[b]++
		}
	}
	mean := make([]float64, nb)
	for b := 0; b < nb; b++ {
		if cnt[b] > 0 {
			mean[b] = sum[b] / float64(cnt[b])
		}
		if mean[b] > peak {
			peak = mean[b]
		}
	}
	half := 0.5 * peak
	for b := 1; b < nb; b++ {
		if cnt[b] > 0 && mean[b] >= half && mean[b-1] < half {
			frac := (half - mean[b-1]) / (mean[b] - mean[b-1])
			cross = (float64(b-1) + 0.5 + frac) / nb
			break
		}
	}
	return cross, peak
}

// TestTunnelFigureFogGeometry pins how deep the fog's black core reaches — the
// tunnel's one quantitative geometry figure — from both sides: the analytic
// half-lit radius the shipped constants imply, and the rendered half-peak crossing
// the terminal actually receives, which must be the same curve at every pane size.
func TestTunnelFigureFogGeometry(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	pal := splashTestPalette()

	// Analytic: half-lit radius from the scale-free fog term, in units of maxD.
	frac := tunFogA / tunRefD
	rhoHalf := 0.5 * frac / (tunFogGain - 0.5)
	requireFigure(t, "analytic fog half-lit radius (fraction of maxD)", 0.184, rhoHalf,
		"tunnel.go tunFogA/tunFogGain comments; splashTunnelAtFor derivation")

	// Rendered: the half-peak crossing, and its scale invariance.
	panes := tunnelFigPanes()
	crossings := make([]float64, len(panes))
	lo, hi := math.Inf(1), math.Inf(-1)
	var meanCross float64
	for i, p := range panes {
		c, peak := tunnelRadialHalfPeak(t, p[0], p[1], pal)
		crossings[i] = c
		meanCross += c
		lo, hi = math.Min(lo, c), math.Max(hi, c)
		t.Logf("%dx%d: half-peak crossing %.1f%% of maxD (peak stop %.1f)", p[0], p[1], c*100, peak)
		require.Lessf(t, c, rhoHalf,
			"the rendered crossing (%.3f) must sit inside the analytic half-lit radius "+
				"(%.3f) — the contrast curve and quantization take their cut", c, rhoHalf)
	}
	meanCross /= float64(len(panes))

	requireFigure(t, "rendered half-peak crossing (mean of 4 panes)", 0.146, meanCross,
		"tunnel.go splashTunnelAtFor comment (≈15%); TestSplashTunnelRendersDepthAsLuminance")

	require.Lessf(t, hi-lo, 0.03,
		"the rendered profile must be the same curve at every pane size (scale invariance): "+
			"half-peak crossings spread %.3f across 96×30→300×80 (%.1f%%–%.1f%%)", hi-lo, lo*100, hi*100)
}
