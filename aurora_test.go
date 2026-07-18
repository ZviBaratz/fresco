package fresco

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// auroraTestPhases spread across the drift so the sweep sees the curtains at
// several points in their travel, including the very first frames.
var auroraTestPhases = []float64{0, 5 * driftPerFrame, 0.9, 2.7, 4.5, 18.3}

// TestSplashAuroraReachesItsOwnField is the same silent-failure guard the tunnel,
// ripple and galaxy carry: splashFieldAt's switch falls through to the fallback, so
// an aurora wired into the enum, rotation, names, ops and both test maps — but
// missing its one case in splashFieldAt — renders rain's field wearing aurora's
// Pass-2 policy, and every coverage loop is happy with that.
//
// It samples the point function, never two renders, and it samples rain because rain
// is what splashFieldAt's default arm returns — so this asks literally "am I getting
// the fallback's field". maxD is passed but unused (aurora is an absolute field); any
// positive value serves.
func TestSplashAuroraReachesItsOwnField(t *testing.T) {
	const phase = 5 * driftPerFrame
	sample := func(v Variant) []float64 {
		at := splashFieldAt(v, 96)
		out := make([]float64, 0, 2*21*31)
		for row := -10; row <= 10; row++ {
			for col := -15; col <= 15; col++ {
				val, aux := at(col, row, float64(col), float64(row)*cellAspect, phase)
				out = append(out, val, aux)
			}
		}
		return out
	}
	require.NotEqual(t, sample(Rain), sample(Aurora),
		"aurora must reach splashAuroraAt — a variant with no case in splashFieldAt "+
			"silently falls through to the fallback's field")
}

// TestSplashAuroraAtRange holds the point-fn contract every field shares: both
// returns inside [0,1] and neither ever a NaN. Aurora's aux runs the altitude sweep
// past both gradient ends before clamping, so this pins that the clamp actually
// lands it in range rather than leaving it to be discovered downstream.
func TestSplashAuroraAtRange(t *testing.T) {
	for _, phase := range auroraTestPhases {
		for dy := -70.0; dy <= 70; dy += 1.3 {
			for dx := -130.0; dx <= 130; dx += 1.3 {
				val, aux := splashAuroraAt(0, 0, dx, dy, phase)
				require.Falsef(t, math.IsNaN(val) || math.IsNaN(aux),
					"NaN at (%v,%v) phase %v: val=%v aux=%v", dx, dy, phase, val, aux)
				require.GreaterOrEqual(t, val, 0.0)
				require.LessOrEqual(t, val, 1.0)
				require.GreaterOrEqual(t, aux, 0.0)
				require.LessOrEqual(t, aux, 1.0)
			}
		}
	}
}

// TestSplashAuroraFadesWithAltitude pins aurora's own invariant — the curtains hang
// over dark sky and fade toward the top and bottom of it, so far above and below the
// focal band the light is gone. It is the vertical envelope (see auroraSpread), and
// it is what keeps the field from being a wall of glyphs from edge to edge; a design
// that dropped the envelope would still pass the range and reach guards.
func TestSplashAuroraFadesWithAltitude(t *testing.T) {
	for _, phase := range auroraTestPhases {
		for dx := -130.0; dx <= 130; dx += 2.5 {
			for _, dy := range []float64{-300, 300} {
				val, _ := splashAuroraAt(0, 0, dx, dy, phase)
				require.Lessf(t, val, 1e-6,
					"aurora should be dark far from the focal band: val=%v at (%v,%v) phase %v",
					val, dx, dy, phase)
			}
		}
	}
}
