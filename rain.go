package fresco

// Matrix-style digital rain: per-column streams of glyphs falling with bright
// heads and fading tails, layered at four depths for parallax.
//
// The first splash field that *travels*. Every other one shimmers in place, and
// a bright leading edge with a decaying trail is the canonical signal the motion
// system locks onto — which is why this one needed a brightness channel the
// palette does not have (see buildRainRamp) and why almost every constant here
// is about the *step* between a head and its tail rather than about either one.
//
// It is also the field that exposed the text clearing, and eventually retired it.
// That clearing blanked an ellipse of field around the focal point: an organic field
// hid it by fading into it, but rain's streams are long, straight and vertical,
// so a blanked row reads as a band cut clean through them. Rain opted out, the
// tunnel and ripple did the same for their own geometry, and V5 retired the last
// field the margin ever flattered — so no variant takes one now and the machinery
// is gone. It never protected the text (see TestOverlayIsOpaque); it was only
// ever charm.

import "math"

const (
	// rainFall is the fall speed in aspect units per phase unit. phase advances
	// driftPerFrame (0.015) per frame at ~60fps, so 30 aspect units/phase is
	// 30 * 0.9 / cellAspect ≈ 13.5 rows/second — cmatrix's own pace. Per frame a
	// head moves 30*0.015/2 = 0.225 rows, i.e. it sits in the same row for ~4
	// frames.
	//
	// That is exactly why brightness below is a function of the *continuous*
	// distance to the head and never of a rounded row count. Quantized rain
	// would be a 4fps stutter inside a 60fps tick, and — because the contract
	// requires consecutive frames to differ, while rain lights far fewer cells
	// than the dense fields do — it would also make that test a coin flip
	// rather than a guarantee. The sub-row gradient fixes both at once.
	rainFall = 30.0

	// Per-column speed spread. Columns must not fall in lockstep or the rain
	// reads as one sliding texture instead of many streams.
	rainSpdMin = 0.55
	rainSpdMax = 1.45

	// Tail length, hashed per stream, as a fraction of *its layer's* period —
	// not an absolute length. The layers' periods differ (58/44/34/26), so one
	// global range cannot serve them: at an absolute 8–26 the tails of the
	// shortest-period layer ran to 26 against a period of 30, so every stream's
	// tail reached the head behind it and the column rendered as one unbroken
	// line. Two of the three layers there were then came out solid, which is why
	// the field read as a uniform wash of glyphs rather than as rain. Expressing
	// the length as a fraction is also what let a fourth depth be added without
	// touching this window at all.
	//
	// The ceiling must stay below 0.5: a tail longer than half its period is a
	// column with no gap in it. The gaps are the whole rhythm — uninterrupted
	// rain reads as static noise; rain with gaps reads as falling.
	rainTailFracMin = 0.12
	rainTailFracMax = 0.38

	// The head lobe, in aspect units. rainHeadFlat is a *plateau* at full
	// brightness, and it exists because rows are sampled every cellAspect (2.0)
	// units: a pure peak is caught only when a row happens to land on it, so at
	// the old 3.2-unit radius the bright part spanned 43% of a row and over half
	// of all heads rendered with no white cell at all — they blinked. A plateau
	// wider than one row guarantees every head lands on at least one cell at the
	// top of the ramp. rainHeadR is where the lobe finally reaches zero, giving
	// the soft leading edge that slides between rows.
	//
	// rainHeadR is down from the 4.5 this field first shipped, and the reason is
	// that 4.5 was quietly cancelling rainTailAmp. That constant buys a ~28-point
	// L* gap under the head — but only *below the lobe's reach*, and a 4.5-unit
	// lobe reaches 1.68 rows. The cell one row behind a head landed at L* 71.8
	// against the head's 81.9: a 10-point step, so the head rendered as a two-cell
	// blob and the darkness the tail had paid for never showed. At 2.9 the same
	// cell falls to L* 47.4 and the step is 34.5. Measured across the sweep:
	// 4.5 → 10.1, 3.2 → 28.4, 2.9 → 34.5, 2.6 → 40.6.
	//
	// It stops at 2.9 rather than going lower because the lobe must still just
	// outrun a mid-length tail one row back — below ~2.7 the tail wins there, the
	// lobe stops being a lobe, and the head reads as a lone cell disconnected from
	// its own stream.
	//
	// Two things this does *not* do. It does not touch the anti-blink guarantee:
	// the top of the ramp is reachable only from the plateau (lit == 1 exactly),
	// which is rainHeadFlat's job, and the rendered count of top-stop cells is
	// unchanged across the whole sweep — see TestRainHeadAlwaysLandsOnACell. And
	// because the lobe is symmetric in d, shrinking it also cuts the glow *ahead*
	// of the head from ~1.7 rows to under a cell, which is its own small win: a
	// falling thing should not be lit in front of itself.
	rainHeadFlat = 1.15
	rainHeadR    = 2.9

	// rainDensity is the fraction of (column, layer, stream) slots that carry a
	// stream at all; the rest are gaps. Sparse on purpose — four layers of
	// streams compound, and a screen with a glyph in every cell is a texture,
	// not weather.
	//
	// Down from 0.62, and it is the fourth depth that spends it: the layers
	// compound, so holding the per-slot rate fixed while adding a layer would have
	// put a third again as many streams on the pane. 0.54 across four layers lands
	// the lit fraction just under where three layers at 0.62 had it (27.0% against
	// 27.8% at 96×30), so the new depth arrives as depth rather than as clutter.
	//
	// It stops at 0.54 rather than going lower because of small panes, and the
	// floor was found by counting rendered heads rather than reasoned to. At 0.52
	// a 30×10 pane spent 21 frames in 60 with no white head on screen at all,
	// against 15 for the three-layer field; at 0.54 that is 3, and every pane from
	// 36×12 up matches the old field frame for frame. Below ~36×12 the count stops
	// responding to this constant entirely — a 24×8 pane sits at 29 frames in 60
	// at 0.52, 0.54 and 0.62 alike — because there the edge vignette, not the
	// stream density, is what leaves too few rows for a head to land in.
	rainDensity = 0.54

	// rainTailAmp caps the tail's brightness, and the gap it opens under the head
	// is the whole reason a head reads as one.
	//
	// It has to be this low because the palette has no brighter white to reach
	// for: pal.Fg is #c0caf5 at L* 81.9, a mere 2.2 above pal.Cyan, so the ramp's
	// top four stops are visually one colour. At 0.82 the tail's brightest cell
	// landed on stop 12 — L* 78.0 against the head's 81.9, a gap of under four
	// points. The head was the same brightness as the cell behind it, and no
	// amount of widening its lobe could make it pop. The head cannot be made
	// brighter, so the tail is made darker: at 0.55 the near layer's tail tops
	// out around L* 54, opening a ~28-point step under a head that now reads as
	// white-hot.
	//
	// It darkens the field as a whole too, which rain wants: the tail's lower
	// half now falls below the terminal background and simply disappears, so the
	// screen is mostly dark with bright streams on it, rather than a uniform haze
	// of mid-grey glyphs.
	//
	// The value has not moved, but what it buys has: the ~28-point step it opens
	// was for a long time invisible, because the head lobe's own shoulder sat in
	// the gap and filled it back in. Tightening rainHeadR is what finally handed
	// the cell behind a head to this constant, so read the two together — this one
	// sets how dark the tail is, and rainHeadR sets whether that darkness reaches
	// the head at all.
	rainTailAmp = 0.55
)

// rainLayers are the parallax depths, near to far.
//
// Depth is luminance first. Each layer's bright caps how far up the ramp its
// streams can climb, and the ramp runs dark → the stream hue → white: the near
// layer reaches the white head, the mid layer tops out around the stream hue,
// and the far ones never leave the dim end. That is atmospheric perspective,
// and it is the cue the earlier hue-per-layer attempt was standing in for —
// badly, because hue says *which* layer without saying which is nearer.
//
// speed is the second cue, and an independent one: motion parallax is monocular
// and needs no vanishing point, so nearer simply means faster. period spaces the
// far layers' streams more tightly, the way distance packs anything together.
//
// There are four of them, up from three, and the fourth is what makes the depth
// read as a continuum rather than as two sheets. Three layers left a hole in the
// middle of the brightness histogram — a bright population and a dim one with
// little between — so the eye sorted them into "near" and "far" instead of
// reading a recession. The room to add one came from tightening rainHeadR, which
// dropped the field from 27.8% lit to 23.7%; rainDensity spends the rest.
//
// The bright column is the constrained one, and three separate guards bound it.
// It must be strictly descending, and each step must clear ~10 L* on the rendered
// ramp, or the layers stop being distinguishable (TestRainLayersSeparateInBrightness)
// — these land at L* 81.9 / 65.7 / 47.4 / 35.2, so 16.2 / 18.3 / 12.2. Each layer
// must also keep its own head above its own tail by 15 L*
// (TestRainHeadOutshinesItsTail) — 28.3 / 30.5 / 18.2 / 18.2. And the near layer's
// 1.00 is pinned rather than chosen: the ramp's top stop is reachable only at
// lit == 1, so any less and no head anywhere renders white
// (TestRainKeepsItsHeadsAwayFromTheFocalPoint). The quantization is coarse — 16
// stops — so these are stop assignments, not a smooth dial; nudging a bright by
// 0.02 often moves nothing at all, and then moves a whole 6-point stop at once.
//
// The last entry must keep the shortest period: TestRainTailFadesFromTheHead
// reads rainLayers[len-1].period to find the shortest tail the field can produce.
var rainLayers = [4]struct {
	speed, bright, period float64
}{
	{speed: 1.00, bright: 1.00, period: 58.0}, // near: reaches white
	{speed: 0.68, bright: 0.72, period: 44.0}, // mid:  the stream hue
	{speed: 0.46, bright: 0.52, period: 34.0}, // deep: below the hue, still legible
	{speed: 0.30, bright: 0.36, period: 26.0}, // haze: dim only, the far wash
}

// Lattice seeds for the per-stream draws (distinct from every field seed).
const (
	seedRainOff   uint32 = 0x51A7C39B
	seedRainSpd   uint32 = 0x7B3D2E11
	seedRainTail  uint32 = 0x2C9E4F07
	seedRainLive  uint32 = 0x6D1B8A53
	seedRainGlyph uint32 = 0x3F5B7C21
)

// rainStreamKey folds a stream's identity — the column it falls in, which head
// of that column's train it is, and which layer it belongs to — into the two
// coordinates the lattice hash takes. Every per-stream draw goes through it, so
// they all key the same stream rather than each combining the parts its own way.
//
// Folding k and li together keeps the key injective, which combining k into col
// does not: col^k and col+k both collide across columns — (col 1, head 2) and
// (col 2, head 1) land on one key and so share every draw they make. The per-
// column speed and offset scatter those twins to unrelated screen positions, so
// nothing showed; the key is injective anyway, because a draw that silently
// aliases is a bug waiting for the day something does depend on it.
func rainStreamKey(col, k, li int) (int, int) {
	return col, k*len(rainLayers) + li
}

// splashRainAt evaluates the rain field at one cell.
//
// The formulation is a *stream train*: rather than tracking one head per column
// and wrapping it at the pane height, each column carries an infinite train of
// heads spaced `period` apart, drifting downward with phase. A cell asks which
// head is nearest and how far behind it sits. Two things fall out of that.
// First, no pane height is needed — the evaluator never learns h, so a taller
// pane simply shows more of the same rain instead of the same rain stretched.
// Second, a stream's identity is the head index k, which is fixed for that
// stream's whole life, so its speed and tail length can be hashed from it and
// never flicker as it falls.
func splashRainAt(col, _ int, _, dy, phase float64) (val, aux float64) {
	best, bestAux := 0.0, 0.0
	for li := range rainLayers {
		L := rainLayers[li]

		// Per-column draws, constant for the column's whole life.
		sp := rainSpdMin + (rainSpdMax-rainSpdMin)*splashCellHash(col, li, seedRainSpd)
		// A full-period offset, not a jitter: a small scatter would leave frame 0
		// showing a rank of heads marching in lockstep, and columns only desync
		// slowly afterward via their speed spread.
		off := splashCellHash(col, li, seedRainOff) * L.period

		// Which head of this column's train is nearest, and how far behind it.
		g := (dy - phase*rainFall*sp - off) / L.period
		kf := math.Round(g)
		// Round, not Floor: it makes d signed, so the head's lobe straddles the
		// two rows it lies between instead of snapping to the one below it.
		d := (kf - g) * L.period // >0 ⇒ this cell trails the head (its tail)
		// g is finite by construction (every term is, and period is a nonzero
		// constant), so this conversion is defined — unlike a float→int of an
		// Inf, which is implementation-defined and would differ across arches.
		// It grows with phase at ~0.5 units/frame, so int32's range is some
		// centuries of continuous animation away.
		k := int(kf)

		// Per-stream draws, keyed on the stream's identity so they hold for its
		// whole life rather than changing under it as it falls.
		kc, kr := rainStreamKey(col, k, li)
		if splashCellHash(kc, kr, seedRainLive) > rainDensity {
			continue // a gap in this column's train
		}
		// Scaled to this layer's period, so every layer keeps its gaps.
		tail := L.period * (rainTailFracMin +
			(rainTailFracMax-rainTailFracMin)*splashCellHash(kc, kr, seedRainTail))

		// Head lobe, then tail. Both are continuous in d — that is the whole
		// trick (see rainFall).
		d0 := math.Abs(d)
		lit := 0.0
		switch {
		case d0 <= rainHeadFlat:
			lit = 1 // the plateau: always at least one cell wide (see rainHeadFlat)
		case d0 < rainHeadR:
			lit = (rainHeadR - d0) / (rainHeadR - rainHeadFlat)
		}
		if d > 0 {
			if t := rainTailAmp * clamp01(1-d/tail); t > lit {
				lit = t
			}
		}
		lit *= L.bright
		if lit > best {
			best = lit
			bestAux = lit // unused by the luminance path; kept in [0,1] for the contract
		}
	}
	// Layers combine by max, not by sum: a far stream crossing behind a near
	// one must not brighten its head — and taking the max is also what makes the
	// near layer *occlude* the far one rather than blend with it.
	return clamp01(best), bestAux
}

// splashRainGlyphs is the vocabulary a stream's cells are drawn from: the
// film's own compromise, half-width katakana for the look with digits and a few
// operators so it reads as a machine rather than as a language. Chosen over
// pure ASCII (which read as terminal code, not rain) and pure katakana (which
// read as text) by rendering all three and looking.
//
// Two properties every glyph here has to keep.
//
// Terminal-width-1, because the contract requires each row to be exactly w
// runes and a width-2 glyph would shift every cell after it, breaking the column
// alignment rain is made of. Half-width katakana (U+FF66–FF9D) is Unicode
// East-Asian-Halfwidth, so this is guaranteed by the standard rather than by
// hope — but TestRainGlyphsAreWidthOne is what actually settles it, since the
// tables only describe what a font *should* do. A font missing the range
// entirely would draw tofu; the ones this was rendered on do not.
//
// And even visual weight. Brightness is the luminance ramp's job now, so a light
// "." among them would read as a hole in a stream rather than as a dimmer cell.
//
// []rune, not string, and that is load-bearing rather than stylistic: the pick
// below is a modulo into this set, and on a string that indexes *bytes*. Katakana
// are three bytes each, so a string here would silently emit mangled half-runes.
var splashRainGlyphs = []rune("ｱｳｴｵｶｷｸｹｻｼｽｾﾀﾂﾃﾅﾆﾇﾈﾊﾋﾌﾍﾎﾏﾐﾑﾒﾓﾗﾘﾙﾚﾜ0123456789<>=+*")

// splashRainMutSpeed is how fast a cell re-draws its glyph, in mutations per
// phase unit. Slow on purpose: mutating every frame boils, and the eye reads
// churn as noise rather than as falling.
const splashRainMutSpeed = 1.6

// splashRainGlyph picks a cell's character. It is keyed on the cell rather than
// on the stream, so a glyph belongs to a position the rain falls *through* —
// which is what makes a stream read as passing over the screen rather than as a
// rigid object sliding down it.
func splashRainGlyph(col, row int, phase float64) rune {
	epoch := int(phase * splashRainMutSpeed)
	h := splashHash(int32(col), int32(row*977+epoch), seedRainGlyph) //nolint:gosec // G115: cell coords are pane-bounded
	return splashRainGlyphs[h%uint32(len(splashRainGlyphs))]         //nolint:gosec // G115: the glyph set is a fixed literal of a few dozen runes
}
