package fresco

import (
	"math"
	"testing"

	"github.com/muesli/termenv"
	"github.com/stretchr/testify/require"
)

// This file is the galaxy's shipped-figure guard. Every "because" the codebase
// states about the galaxy is backed by a measurement, but those measurements live
// in prose (doc comments, CHANGELOG, SKILL.md), and prose rots: the #60/#61 work
// alone quoted three figures that had already gone stale (118.5 → 135.5, the ruler
// "odd → even", "~3% → ~0.7%") and rebuilt a self-checking probe that #59 had
// already git-rm'd. A shipped figure that gets cited is a claim about the compiled
// renderer, so it belongs in a test that fails when the renderer drifts under it —
// not only in a sentence that stays green while going wrong.
//
// The distinction this file draws (see SKILL.md §7, "A shipped figure is a test,
// not a comment"):
//
//   - A TUNING PROBE is throwaway. It varies a const across a sweep to pick a
//     value, re-derives internals if it must, and is git-rm'd before the commit.
//   - A FIGURES TEST is permanent. It asserts the figure the compiled renderer
//     ACTUALLY produces, measured through the existing helpers — no twin of the
//     renderer's internals, no const-variation. That is the only kind of figure
//     here.
//
// So this file does NOT assert the ridge/gas ratio (243.8 / 26.8 → 9.1×). That
// figure is defined on the raw arm value BEFORE the turbulence lift (CHANGELOG,
// #61 entry), which splashGalaxyAtFor never returns — measuring it would mean
// copying the arm sub-computation (psi, LOD, modMip, pow) into the test, a twin
// that silently drifts from the renderer. That is exactly the rot this file
// exists to prevent, so ridge/gas stays a tuning-probe finding in prose.
//
// Tolerance is ±5% (galFigTol): the field is fully deterministic (fixed seeds,
// fixed frames), so there is no run-to-run noise to absorb — the margin is
// headroom for an innocuous refactor (a float reassociation), narrow enough that
// structural drift fails. A red build means a shipped figure moved; each assertion
// names the prose that quotes it so you update the figure and the prose together.
//
// Two measurement traps, both load-bearing here:
//  1. Colour stops need TrueColor; galaxyGlyphGrid flips the profile to Ascii. The
//     stop-density figure is therefore its own test (TestGalaxyFigureCoreStopDensity),
//     so a glyph test's Ascii profile can never clobber it.
//  2. The radial ruler is even-height-sensitive (#61); every test measures ρ against
//     splashMaxD — the renderer's own length scale — inline, exactly as the galaxy
//     band tests do, never a second re-derivation of it.
//
// Figure → where it is quoted (change both together when a retune moves one):
//
//	whole-pane / arm-annulus clipping  galaxy.go (galBulgeAmp, galKnotAmp); CHANGELOG [1.1.0], #61
//	nucleus glyph contrast             galaxy.go (galBulgeAmp); galaxy_test.go (0.25 floor); CHANGELOG [1.1.0]
//	core/arm/outskirts bead density    galaxy.go (galKnotAmp); galaxy_test.go (40 floor); CHANGELOG [1.1.0], #61; SKILL.md §7
//	core/mid colour-stop density       galaxy.go (galBulgeAmp); CHANGELOG [1.1.0]
//	lumRange 0.75→0.60 arm-annulus     variant.go (lumRange); CHANGELOG [1.1.0]

// galFigTol is the ±5% band every galaxy figure is asserted within, and galFigFrames
// are the frames the motion-averaged figures (clipping, contrast, beads) run over.
const galFigTol = 0.05

func galFigFrames() []int { return []int{0, 30, 60} }

// galaxyPlaneRho is the shipped galaxy's in-plane ρ(col, row): the un-inclined
// radius as a fraction of the disk, measured against the renderer's own length
// scale (splashMaxD) exactly as splashGalaxyAtFor and the galaxy band tests are —
// so a figure moves only when the field moves, never when the ruler is recomputed
// a second way (#61).
//
// The w/h parameters are suppressed for the unparam linter (see the directive on
// the signature line): every galaxy figure is quoted at the one canonical 240×60
// pane, so w/h are constant across all callers by design; they stay parameters so
// this ruler reads the same as the (w, h)-taking band helpers it mirrors.
func galaxyPlaneRho(w, h int) func(col, row int) float64 { //nolint:unparam // figures are all at 240×60; see doc comment
	cx, cyFocal := float64(w-1)/2, float64((h-1)/2)
	rr := galExtent * splashMaxD(w, h, centeredFocalRow(h))
	cosInc := math.Cos(galInc)
	return func(col, row int) float64 {
		dx, dy := float64(col)-cx, (float64(row)-cyFocal)*cellAspect
		return math.Hypot(dx, dy/cosInc) / rr
	}
}

// requireFigure asserts got is within galFigTol of the canonical want, logging the
// live value either way, and points a red build at the prose to update alongside it.
func requireFigure(t *testing.T, name string, want, got float64, quotedIn string) {
	t.Helper()
	t.Logf("%s = %.4f (canonical %.4g)", name, got, want)
	require.InEpsilonf(t, want, got, galFigTol,
		"shipped figure %q drifted to %.4f from the canonical %.4g (±%.0f%%). If this is "+
			"an intended change, update the figure HERE and the prose that quotes it: %s",
		name, got, want, galFigTol*100, quotedIn)
}

// TestGalaxyFigureClipping pins the field-level clipping the #60 bulge drop bought:
// the fraction of cells whose val saturates to 1.0 upstream of Pass 2. It is a
// field fact (the colour profile and the glyph ramp never see it), so it samples
// splashGalaxyAtFor directly through the same (dx, dy, phase) mapping renderField
// feeds each cell — the compiled renderer's field, not a reconstruction. Whole-pane
// counts every cell (the edge vignette is a Pass-2 dimming that cannot change a
// field val); the arm annulus is the 0.35 ≤ ρ < 0.60 band the bead figures use.
func TestGalaxyFigureClipping(t *testing.T) {
	const w, h = 240, 60
	cx, cyFocal := float64(w-1)/2, float64((h-1)/2)
	at := splashFieldAt(Galaxy, splashMaxD(w, h, centeredFocalRow(h)))
	rho := galaxyPlaneRho(w, h)

	frames := galFigFrames()
	var pane, ann float64
	for _, f := range frames {
		phase := float64(f) * driftPerFrame
		clipPane, totPane := 0, 0
		clipAnn, totAnn := 0, 0
		for row := 0; row < h; row++ {
			for col := 0; col < w; col++ {
				dx, dy := float64(col)-cx, (float64(row)-cyFocal)*cellAspect
				val, _ := at(col, row, dx, dy, phase)
				clipped := val >= 1.0
				totPane++
				if clipped {
					clipPane++
				}
				if r := rho(col, row); r >= 0.35 && r < 0.60 {
					totAnn++
					if clipped {
						clipAnn++
					}
				}
			}
		}
		pane += 100 * float64(clipPane) / float64(totPane)
		ann += 100 * float64(clipAnn) / float64(totAnn)
	}
	nf := float64(len(frames))
	requireFigure(t, "whole-pane clipping %", 1.36, pane/nf, "galaxy.go galBulgeAmp comment; CHANGELOG [1.1.0]")
	requireFigure(t, "arm-annulus clipping %", 0.92, ann/nf, "galaxy.go galKnotAmp comment; CHANGELOG #61 entry")
}

// TestGalaxyFigureNucleusContrast pins the flat-mass metric #60 turns on: the mean
// |centre − mean(8-ring)| glyph-ramp contrast in the nucleus (ρ < 0.05), which ran
// 0.13 on the pre-#60 saturated core and 0.43 after. It is the density channel, so
// it reads the glyph grid, not the luminance stops (lumRange splits brightness
// across the two — a density claim on the stops measures the wrong half).
func TestGalaxyFigureNucleusContrast(t *testing.T) {
	const w, h = 240, 60
	rho := galaxyPlaneRho(w, h)
	var contrast float64
	frames := galFigFrames()
	for _, f := range frames {
		g := galaxyGlyphGrid(t, w, h, f, splashTestPalette())
		contrast += galaxyLocalGlyphContrast(g, w, h, rho, 0, 0.05)
	}
	requireFigure(t, "nucleus glyph contrast (ρ<0.05)", 0.430, contrast/float64(len(frames)),
		"galaxy.go galBulgeAmp comment; galaxy_test.go core-structure floor; CHANGELOG [1.1.0]")
}

// TestGalaxyFigureBeadDensity pins the knot distribution: beads per 1000 lit cells
// (a cell standing a full glyph step above its eight neighbours) in the core, arm
// and outskirts bands. The arm figure (135.5) is the one TestSplashGalaxyArmsCarryKnots
// floors at 40 and the one the lumRange A/B lands on at 0.60.
func TestGalaxyFigureBeadDensity(t *testing.T) {
	const w, h = 240, 60
	rho := galaxyPlaneRho(w, h)
	var core, arms, outskirts float64
	frames := galFigFrames()
	for _, f := range frames {
		g := galaxyGlyphGrid(t, w, h, f, splashTestPalette())
		core += galaxyBeadDensity(g, w, h, rho, 0, 0.15)
		arms += galaxyBeadDensity(g, w, h, rho, 0.35, 0.60)
		outskirts += galaxyBeadDensity(g, w, h, rho, 0.60, 1.10)
	}
	nf := float64(len(frames))
	requireFigure(t, "core beads/1000", 58.4, core/nf, "CHANGELOG [1.1.0]")
	requireFigure(t, "arm beads/1000", 135.5, arms/nf,
		"galaxy.go galKnotAmp comment; galaxy_test.go arms floor; variant.go lumRange comment; CHANGELOG [1.1.0], #61; SKILL.md §7")
	requireFigure(t, "outskirts beads/1000", 67.3, outskirts/nf, "CHANGELOG [1.1.0]")
}

// TestGalaxyFigureCoreStopDensity pins that the core stays the field's brightest
// region: mean colour-stop density (blanks counted as zero) is 11.46 in the bulge
// against the mid-disk's 6.13. This is a luminance measurement, so it needs
// TrueColor — and its own test, so galaxyGlyphGrid's Ascii profile (used by the
// bead/contrast figures) can never flip it out from under this one.
func TestGalaxyFigureCoreStopDensity(t *testing.T) {
	withColorProfile(t, termenv.TrueColor)
	const w, h, frame = 240, 60, 40
	rho := galaxyPlaneRho(w, h)
	stops, _ := shadeStopGrid(t, w, h, frame, splashTestPalette(), Galaxy)
	band := func(lo, hi float64) float64 {
		sum, total := 0, 0
		for row := 0; row < h; row++ {
			for col := 0; col < w; col++ {
				if !galaxyMeasurable(col, row, w, h) {
					continue
				}
				if r := rho(col, row); r < lo || r >= hi {
					continue
				}
				total++
				if s := stops[row][col]; s > 0 {
					sum += s
				}
			}
		}
		if total == 0 {
			return 0
		}
		return float64(sum) / float64(total)
	}
	requireFigure(t, "core stop density (frame 40)", 11.46, band(0, 0.15), "galaxy.go galBulgeAmp comment; CHANGELOG [1.1.0]")
	requireFigure(t, "mid-disk stop density (frame 40)", 6.13, band(0.4, 0.62), "galaxy.go galBulgeAmp comment; CHANGELOG [1.1.0]")
}

// TestGalaxyFigureLumRangeArmBeads pins the lumRange lever's effect on the arm
// annulus: lowering it 0.75 → 0.60 (#56) lifts arm bead density 93.1 → 135.5,
// because the steep dens = lit^(1-lumRange) compressor stops spending the glyph
// ramp where no cell renders. It measures the compiled renderer at each lumRange
// through the public Options.LumRange override (withLumRange) — no twin — so it
// guards the variant.go mechanism claim rather than transcribing it. The 0.60
// endpoint is the shipped state, so it must reproduce the arm bead figure above.
func TestGalaxyFigureLumRangeArmBeads(t *testing.T) {
	const w, h = 240, 60
	rho := galaxyPlaneRho(w, h)
	frames := galFigFrames()
	armBeads := func() float64 {
		var a float64
		for _, f := range frames {
			g := galaxyGlyphGrid(t, w, h, f, splashTestPalette())
			a += galaxyBeadDensity(g, w, h, rho, 0.35, 0.60)
		}
		return a / float64(len(frames))
	}

	withLumRange(t, 0.75)
	at75 := armBeads()
	withLumRange(t, 0.60)
	at60 := armBeads()

	requireFigure(t, "arm beads/1000 @ lumRange 0.75", 93.1, at75, "variant.go lumRange comment")
	requireFigure(t, "arm beads/1000 @ lumRange 0.60", 135.5, at60, "variant.go lumRange comment; CHANGELOG [1.1.0]")
	require.Greaterf(t, at60, at75,
		"lowering lumRange must LIFT arm bead density (the dens=lit^(1-lumRange) claim in "+
			"variant.go): 0.75→%.1f, 0.60→%.1f", at75, at60)
}
