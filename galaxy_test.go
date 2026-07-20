package fresco

import (
	"math"
	"strings"
	"testing"

	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
)

// TestSplashGalaxyReachesItsOwnField is the same silent-failure guard the tunnel
// and ripple carry: splashFieldAt's switch falls through to the fallback, so a
// galaxy wired into the enum, rotation, names, ops and both test maps — but missing
// its one case in splashFieldAt — renders rain's field wearing galaxy's Pass-2
// policy, and every coverage loop is happy with that.
//
// It samples the point function, never two renders. Galaxy's ops differ from the
// fallback's on both fields (stars and lumRange), so two ops-applied renders differ
// whatever field is underneath them — the trap the tunnel's first version of this
// test fell into. Sampling rain is load-bearing: rain is what splashFieldAt's
// default arm returns, so this asks literally "am I getting the fallback's field".
// Move that arm elsewhere without moving this probe and the test passes while
// testing nothing.
func TestSplashGalaxyReachesItsOwnField(t *testing.T) {
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
	require.NotEqual(t, sample(Rain), sample(Galaxy),
		"galaxy must reach splashGalaxyAtFor — a variant with no case in splashFieldAt "+
			"silently falls through to the fallback's field")
}

// TestSplashGalaxyArmLODIsAnisotropic pins the one thing about the arm mip that is
// a trap rather than taste: it damps the vertical axis a factor vAspect harder than
// the horizontal, because a screen row covers vAspect in-plane units where a column
// covers one. vAspect is the grid's cellAspect times the inclination's 1/cos(galInc):
// a tilted disk foreshortens the vertical, packing the arms closer there. Get it
// isotropic and an arm crossing the top or bottom of the disk crawls (wagon-wheel)
// while the sides flow — the exposure the user caught in the tunnel twice, from
// motion alone and from no test.
//
// The property is asserted independently of the winding constants, so it survives
// the render round retuning the arm count or pitch. |∇ψ| has the same magnitude at
// every point of a given radius — it is sqrt(galWind²+galArms²)/r, rotation-free —
// so two points at one radius can be chosen where the phase gradient is purely
// horizontal (point A) and purely vertical (point B) with equal magnitude G. Then
// the only thing separating their LODs is the vAspect weight on the vertical term,
// so lod_A/lod_B is exactly vAspect. An isotropic mip makes it 1.
func TestSplashGalaxyArmLODIsAnisotropic(t *testing.T) {
	vAspect := cellAspect / math.Cos(galInc)
	d := math.Hypot(galWind, galArms) // |∇ψ|·r, the same at every point of radius r

	// A radius small enough that both LODs are below 1 (the regime where the ratio
	// is legible rather than clamped flat), derived from the constants so a retune
	// that pushes it out of that band fails here loudly instead of passing vacuously.
	r := 0.5 * galLODC * d / math.Pi

	// Point A: ∇ψ purely horizontal (its w-component vanishes on the (galWind,galArms)
	// ray). Point B: ∇ψ purely vertical (u-component vanishes on the (−galArms,galWind)
	// ray). Both at radius r, so both see |∇ψ| = d/r.
	ax, ay := r*galWind/d, r*galArms/d
	bx, by := -r*galArms/d, r*galWind/d

	lodA := splashGalaxyArmLOD(ax, ay, vAspect)
	lodB := splashGalaxyArmLOD(bx, by, vAspect)

	require.Greater(t, lodA, 0.0, "point A must not clamp to zero")
	require.Less(t, lodA, 1.0, "point A must be in the unclamped band, or the ratio is vacuous")
	require.Greater(t, lodB, 0.0, "point B must not clamp to zero")
	require.Less(t, lodB, 1.0, "point B must be in the unclamped band, or the ratio is vacuous")

	// The vertical gradient is damped vAspect harder. Reintroducing an isotropic step
	// (dropping the vAspect factor) makes both points read the same |∇ψ| and the ratio
	// collapses to 1 — which is what this catches.
	require.InEpsilon(t, vAspect, lodA/lodB, 1e-9,
		"the arm LOD must weight the vertical axis by vAspect (grid × inclination)")

	// And the mip points the right way: arms are resolved far out (lod 1) and fade
	// toward the crowded core (lod < 1). An inverted or unclamped mip fails one of
	// these.
	require.Equal(t, 1.0, splashGalaxyArmLOD(60, 0, vAspect), "arms must be fully resolved out in the disk")
	require.Less(t, splashGalaxyArmLOD(1, 0, vAspect), 1.0, "arms must fade toward the core")
	require.Equal(t, 0.0, splashGalaxyArmLOD(0, 0, vAspect), "the exact centre has no resolvable arms")
}

// TestSplashGalaxyCoreIsFinite guards the centre singularity. ln(r) and atan2 are
// undefined or meaningless at r == 0, and the field routes around them by returning
// the bulge alone below galCoreFrac·R — delete that guard and the exact centre
// computes cos(±Inf) == NaN, which clamp01 does not fix (NaN survives every
// comparison) and which then paints a garbage cell at the focal row.
func TestSplashGalaxyCoreIsFinite(t *testing.T) {
	at := splashFieldAt(Galaxy, 96)
	const phase = 3 * driftPerFrame
	for _, p := range []struct{ dx, dy float64 }{
		{0, 0}, {0.001, 0}, {0, 0.001}, {-0.5, 0.3}, {1, -1}, {2, 2},
	} {
		val, aux := at(0, 0, p.dx, p.dy, phase)
		require.Falsef(t, math.IsNaN(val) || math.IsInf(val, 0), "val at (%v,%v) must be finite", p.dx, p.dy)
		require.Falsef(t, math.IsNaN(aux) || math.IsInf(aux, 0), "aux at (%v,%v) must be finite", p.dx, p.dy)
		require.GreaterOrEqual(t, val, 0.0)
		require.LessOrEqual(t, val, 1.0)
		require.GreaterOrEqual(t, aux, 0.0)
		require.LessOrEqual(t, aux, 1.0)
	}
}

// galaxyMeasurable reports whether a rendered cell can be read as the galaxy's own
// brightness. The galaxy draws no starfield (see baseOps: a fixed star would punch a
// hole in the dense disk), so unlike ripple the only exclusion is the edge vignette,
// which dims the border cells for every variant and must stay out of any claim about
// this field's own falloff.
func galaxyMeasurable(col, row, w, h int) bool {
	mx := int(math.Max(1, float64(w)*edgeVignetteFrac)) + 1
	my := int(math.Max(1, float64(h)*edgeVignetteFrac)) + 1
	return col >= mx && col < w-mx && row >= my && row < h-my
}

// TestSplashGalaxyRendersABrightCoreAndDimmingArms is the Pass-2 half: brightness is
// the whole subject, and the claim is that it RENDERS as a radial gradient (a bright
// bulge grading out through dimming arms) with visible arm structure, not as a flat
// disc or a threshold. Asserted on the decoded luminance stops the terminal would
// receive, not on the field — the tunnel/rain lesson that Pass 1 being right never
// proved Pass 2 emits it.
func TestSplashGalaxyRendersABrightCoreAndDimmingArms(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	const w, h, frame = 240, 60, 40
	stops, _ := shadeStopGrid(t, w, h, frame, splashTestPalette(), Galaxy)

	cx, cyFocal := float64(w-1)/2, float64((h-1)/2)
	// The renderer's own length scale, not a re-derivation of it: the test measures
	// bands against exactly the maxD the field was built from (see splashMaxD; #61).
	rr := galExtent * splashMaxD(w, h, centeredFocalRow(h))
	cosInc := math.Cos(galInc)
	// In-plane radius/angle: undo the inclination's vertical foreshortening so the
	// bands track the galaxy's actual (elliptical-on-screen) structure, not a screen
	// circle cutting across it.
	planeRA := func(col, row int) (rho, theta float64) {
		dx, dy := float64(col)-cx, (float64(row)-cyFocal)*cellAspect
		wy := dy / cosInc
		return math.Hypot(dx, wy) / rr, math.Atan2(wy, dx)
	}

	// Brightness density over a ρ band: the mean rendered stop across every
	// measurable cell, blanks counted as zero. That is deliberately not mean-over-lit
	// — lit cells in every band are dominated by saturated arm ridges, so a mean over
	// only-lit cells is nearly flat across radius and cannot see the envelope fade.
	// Counting the dark cells makes it an energy measure that falls as the arms dim
	// and the gaps open, which is what "brightness is the subject" has to render as.
	// Also returns the lit fraction, which the bulge and the arm gaps move.
	band := func(lo, hi float64) (dens, litFrac float64) {
		sum, lit, total := 0, 0, 0
		for row := 0; row < h; row++ {
			for col := 0; col < w; col++ {
				if !galaxyMeasurable(col, row, w, h) {
					continue
				}
				rho, _ := planeRA(col, row)
				if rho < lo || rho >= hi {
					continue
				}
				total++
				if s := stops[row][col]; s > 0 {
					lit++
					sum += s
				}
			}
		}
		if total == 0 {
			return 0, 0
		}
		return float64(sum) / float64(total), float64(lit) / float64(total)
	}

	coreDens, coreLit := band(0, 0.15)
	midDens, midLit := band(0.4, 0.62)

	require.Greater(t, coreLit, 0.9, "the bulge must render nearly solid")
	require.Greater(t, midDens, 0.0, "the mid-disk arms must render, not blank out")
	require.Greater(t, coreDens, midDens,
		"brightness must grade from a bright core to a dimmer disk, not render flat")

	// The core must render as STRUCTURE, not a flat bright mass (#60). The three
	// assertions above are all satisfied by the defect they were meant to guard: a
	// saturated block is "nearly solid" (coreLit == 1.0) and brighter than the mid-disk
	// (coreDens > midDens), so none of them can tell the flat core this issue fixed from
	// a graded one — they measure how *bright* the core is, never whether it *varies*.
	// This does: the mean local glyph contrast in the nucleus — |centre − mean(8-ring)|
	// in ramp steps, on the density channel — is ~0 for a block of one or two glyphs and
	// rises as the bulge grades and the arms wind in. It is the flat-mass metric #60
	// turns on: measured over three frames it ran 0.13 on the pre-#60 saturated core
	// (galBulgeAmp 1.0) and 0.43 after (0.60). The 0.25 floor sits between the two
	// measured states, so it fails the defect with margin rather than transcribing the
	// current number — the same construction as TestSplashGalaxyArmsCarryKnots's floor.
	// Asserted on the NoColor glyph grid: shadeStopGrid decodes colour luminance, and
	// lumRange splits brightness across the two channels, so a density claim read off
	// the luminance stops would be measuring the wrong half.
	nucRho := func(col, row int) float64 { rho, _ := planeRA(col, row); return rho }
	var coreContrast float64
	structFrames := []int{0, 30, 60}
	for _, f := range structFrames {
		g := galaxyGlyphGrid(t, w, h, f, splashTestPalette())
		coreContrast += galaxyLocalGlyphContrast(g, w, h, nucRho, 0, 0.05)
	}
	coreContrast /= float64(len(structFrames))
	require.Greaterf(t, coreContrast, 0.25,
		"the core must render as graded structure, not a flat bright mass: mean nucleus "+
			"local glyph contrast %.3f (flat block ≈0.13, graded core ≈0.43)", coreContrast)

	// Arm structure: across a *thin* in-plane annulus the mean brightness must swing
	// with angle — brighter arms, darker dust lanes and inter-arm disk — rather than a
	// flat ring. It is a brightness swing, not a lit/blank one, because the disk is a
	// full glow (the inter-arm regions render rather than blanking, which is what keeps
	// the tight coil from weaving dark holes through it); the structure lives in *how
	// bright*, so measuring lit fraction would see a uniform 1.0 and miss it. The
	// annulus is thinner than one arm's radial period, so a given angle is mostly arm
	// or mostly lane; a wider band would let the coil cross an arm at every angle and
	// average the structure away.
	const bins = 24
	binSum := make([]int, bins)
	binTot := make([]int, bins)
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			if !galaxyMeasurable(col, row, w, h) {
				continue
			}
			rho, theta := planeRA(col, row)
			if rho < 0.44 || rho >= 0.54 {
				continue
			}
			b := int((theta + math.Pi) / (2 * math.Pi) * bins)
			if b >= bins {
				b = bins - 1
			}
			binTot[b]++
			if s := stops[row][col]; s > 0 {
				binSum[b] += s
			}
		}
	}
	minMean, maxMean := math.Inf(1), math.Inf(-1)
	for b := 0; b < bins; b++ {
		if binTot[b] == 0 {
			continue
		}
		m := float64(binSum[b]) / float64(binTot[b])
		minMean = math.Min(minMean, m)
		maxMean = math.Max(maxMean, m)
	}
	require.Greater(t, maxMean-minMean, 1.5,
		"arms and dust lanes must render as an azimuthal brightness swing, not a flat ring: "+
			"mean stop ran %.2f..%.2f around the annulus", minMean, maxMean)
	require.Greater(t, midLit, 0.0)
}

// galaxyGlyphGrid reads a rendered frame back as each cell's glyph ramp index —
// the density channel. shadeStopGrid cannot see it: that one decodes colour
// luminance, and lumRange is precisely the knob that moves brightness between the
// two, so a claim about density asserted on luminance stops is measuring the other
// channel. Rendered at NoColor so the bytes are exactly the glyphs, with no SGR to
// step over; a cell's glyph does not depend on the colour profile.
func galaxyGlyphGrid(t *testing.T, w, h, frame int, pal Palette) [][]int {
	t.Helper()
	withColorProfile(t, termenv.Ascii)
	out := renderSplashField(w, h, frame, pal, centeredFocalRow(h), Galaxy)
	idx := map[rune]int{}
	for i, r := range splashRampR {
		idx[r] = i
	}
	grid := make([][]int, 0, h)
	for _, line := range strings.Split(out, "\n") {
		row := make([]int, 0, w)
		for _, r := range line {
			g, ok := idx[r]
			require.Truef(t, ok, "rendered glyph %q is not on splashRamp", r)
			row = append(row, g)
		}
		require.Lenf(t, row, w, "row must be exactly %d cells", w)
		grid = append(grid, row)
	}
	return grid
}

// galaxyBeadDensity counts cells standing a full glyph step above the mean of their
// eight neighbours, per 1000 lit cells in a rho band. A bead reads as a bead by
// standing above the cells around it, so the comparison is local: a band mean cannot
// see a sparse minority of bright cells by construction, which is how the previous
// round's "mean glyph weight 9.24 against 9.22" concluded that a knot term working
// on 1.7% of cells did nothing.
func galaxyBeadDensity(grid [][]int, w, h int, rho func(col, row int) float64, lo, hi float64) float64 {
	beads, lit := 0, 0
	for row := 1; row < h-1; row++ {
		for col := 1; col < w-1; col++ {
			if !galaxyMeasurable(col, row, w, h) {
				continue
			}
			if r := rho(col, row); r < lo || r >= hi {
				continue
			}
			if grid[row][col] <= 0 {
				continue
			}
			lit++
			sum := 0
			for dr := -1; dr <= 1; dr++ {
				for dc := -1; dc <= 1; dc++ {
					if dr != 0 || dc != 0 {
						sum += grid[row+dr][col+dc]
					}
				}
			}
			if float64(grid[row][col])-float64(sum)/8 >= 1 {
				beads++
			}
		}
	}
	if lit == 0 {
		return 0
	}
	return 1000 * float64(beads) / float64(lit)
}

// galaxyLocalGlyphContrast is the mean |centre − mean(eight-neighbour ring)| in
// glyph-ramp steps over a rho band — the emitted-byte measure of whether a region
// reads as structure or as a flat mass. A block of one or two glyphs scores near zero
// however bright it is; a graded, knotted region scores high. It shares
// galaxyBeadDensity's neighbour ring but averages the signed-magnitude difference
// rather than counting threshold crossings, because the core defect #60 fixes is the
// *absence of any variation*, not a shortage of bright peaks.
func galaxyLocalGlyphContrast(grid [][]int, w, h int, rho func(col, row int) float64, lo, hi float64) float64 {
	sum, n := 0.0, 0
	for row := 1; row < h-1; row++ {
		for col := 1; col < w-1; col++ {
			if !galaxyMeasurable(col, row, w, h) {
				continue
			}
			if r := rho(col, row); r < lo || r >= hi {
				continue
			}
			if grid[row][col] <= 0 {
				continue
			}
			ringSum := 0
			for dr := -1; dr <= 1; dr++ {
				for dc := -1; dc <= 1; dc++ {
					if dr != 0 || dc != 0 {
						ringSum += grid[row+dr][col+dc]
					}
				}
			}
			sum += math.Abs(float64(grid[row][col]) - float64(ringSum)/8)
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

// TestSplashGalaxyArmsCarryKnots is the guard #56 needed and the previous re-art
// round did not have — the reason a brighter, lower-threshold knot term could ship
// while doing nothing visible, with the whole suite green.
//
// Every other galaxy assertion is a band *mean* (see
// TestSplashGalaxyRendersABrightCoreAndDimmingArms), and a mean cannot see a sparse
// minority of bright cells by construction: the round that claimed the arms were
// studded measured "mean glyph weight 9.24 against 9.22" and concluded the knots did
// nothing, when the real finding was that the instrument could not have detected
// them either way. This one counts cells that stand above their own neighbours,
// which is what a knot reads by.
//
// The floor is placed between two measured states rather than recorded from the
// current one. A disk with the knot term switched off produces ~2.7 beads per 1000
// lit cells (galKnotAmp = 0, 120x40); the pre-#56 turbulence-gated knots produced
// 5.6 here; the shipped high-frequency knots produce ~135. 40 is an order of
// magnitude above both failing states and ~3x below the passing one, so it fails if
// the knots regress to a low-frequency swell and does not merely transcribe today's
// number.
func TestSplashGalaxyArmsCarryKnots(t *testing.T) {
	const w, h = 240, 60
	cx, cyFocal := float64(w-1)/2, float64((h-1)/2)
	rr := galExtent * splashMaxD(w, h, centeredFocalRow(h))
	cosInc := math.Cos(galInc)
	rho := func(col, row int) float64 {
		dx, dy := float64(col)-cx, (float64(row)-cyFocal)*cellAspect
		return math.Hypot(dx, dy/cosInc) / rr
	}
	var arms, outskirts, core float64
	frames := []int{0, 30, 60}
	for _, f := range frames {
		g := galaxyGlyphGrid(t, w, h, f, splashTestPalette())
		arms += galaxyBeadDensity(g, w, h, rho, 0.35, 0.60)
		outskirts += galaxyBeadDensity(g, w, h, rho, 0.60, 1.10)
		core += galaxyBeadDensity(g, w, h, rho, 0, 0.15)
	}
	n := float64(len(frames))
	arms, outskirts, core = arms/n, outskirts/n, core/n
	t.Logf("beads per 1000 lit: core %.1f  arms %.1f  outskirts %.1f", core, arms, outskirts)

	require.Greaterf(t, arms, 40.0,
		"the arms must be studded with knots that stand a glyph step above their "+
			"neighbours, not smoothly mottled: %.1f beads per 1000 lit cells", arms)
	// The defect #56 records is a distribution one — the knots landed everywhere but
	// the arms. The arm annulus is where a star-forming region belongs, so it must
	// carry them at least as densely as the faint outskirts do.
	require.Greaterf(t, arms, outskirts,
		"knots must concentrate on the arms rather than the outskirts: arms %.1f, "+
			"outskirts %.1f per 1000", arms, outskirts)
}
