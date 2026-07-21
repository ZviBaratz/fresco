package fresco

import "math"

// The aurora is the roster's sky entry: tall curtains of polar light hung over a
// dark sky, drifting sideways as a body while each curtain's vertical spine snakes,
// with the hue sliding warm↔cool by altitude. It is an *absolute* field like rain
// and ripple — a bigger pane shows more sky (more curtains, more of the altitude
// gradient) rather than the same object scaled up — so it is a plain point function
// with no maxD, not a single-object closure like the tunnel or galaxy. What makes it
// read as weather rather than as texture is the same discipline the others keep: a
// coherent sideways drift the eye tracks as motion, and dark negative space between
// the curtains for that motion to move through.
//
// One thing about it is a trap rather than taste, and it is commented where it
// bites: the curtain texture *drifts*, so unlike the galaxy's screen-static
// turbulence its fBm may not carry an octave past the grid's Nyquist limit, or the
// fine detail crawls sideways as it translates (see auroraFBM and auroraOct).
const (
	// auroraDrift is the sideways travel of the whole curtain body, in cells per unit
	// phase (x = dx − auroraDrift·phase). It is the field's primary "this is moving"
	// cue — a rigid translation the eye reads as motion — so it is deliberately slow:
	// at driftPerFrame the curtains cross a few cells over a hundred frames, an
	// atmospheric drift rather than a scroll.
	auroraDrift = 5.0
	// The snake: two detuned altitude-sines added to x so each curtain's vertical
	// spine bends, and the bend travels. Two waves rather than one because a single
	// sine bends every curtain in lockstep into one rigid wobble; detuned in both
	// altitude-frequency (F) and time-speed (S) they beat into an irregular,
	// never-repeating drape. A is the horizontal sway in cells, F how many bends span
	// the pane's height, S how fast the bend crawls.
	auroraSnakeA1 = 5.0
	auroraSnakeA2 = 3.0
	auroraSnakeF1 = 0.06
	auroraSnakeF2 = 0.11
	auroraSnakeS1 = 0.8
	auroraSnakeS2 = 1.3
	// auroraFreqX/auroraFreqY are the curtain texture's base spatial frequencies, and
	// their ratio is the whole look: freqX ≫ freqY makes the noise features many times
	// taller than wide, i.e. vertical streaks. freqX sets how many curtains cross the
	// pane (before the threshold thins them); freqY is small enough that a curtain
	// barely varies along its height, so it reads as a coherent drape the snake bends
	// rather than a field of blobs.
	auroraFreqX = 0.10
	auroraFreqY = 0.02
	// auroraRise scrolls the texture slowly *up* the curtains (a phase term on the y
	// axis), so filaments climb the drape without moving the drape itself — the drape
	// travels sideways on auroraDrift. It is gentle: the sideways body-motion is the
	// signal, and this is the internal life on top of it, not a competing scroll.
	auroraRise = 0.5
	// auroraLo/auroraHi carve the curtains out of the sky: curtain =
	// smoothstep(auroraLo, auroraHi, n) lights the cells whose texture noise sits above
	// auroraLo and saturates the cores above auroraHi, leaving everything below auroraLo
	// as dark sky. This is the "weather, not a wall of glyphs" knob, and the window is
	// deliberately fitted to the *measured* fBm distribution rather than to the [0,1]
	// nominal range: this fBm (3 octaves, normalized) clusters around 0.50 and tails off
	// to ~0.79 by its 99th percentile (a thin tail reaches ~0.94), so the window
	// straddles its upper third — auroraLo just above the median so a curtain is a real
	// band of sky and not a scatter of dots, auroraHi near the 90th percentile (~0.67)
	// so the brightest filaments reach full luminance. Both figures are asserted in
	// aurora_figures_test.go. Its width sets
	// the curtain-edge softness: kept broad enough that the edge carries no more high
	// spatial frequency than the drift can move without crawling (see auroraFBM), so no
	// separate sharpening power is wanted.
	auroraLo = 0.55
	auroraHi = 0.70
	// auroraSpread is the vertical envelope: env = exp(−(dy/spread)²), broad and soft,
	// centred on the focal point. It fades the curtains toward the top and bottom of
	// the sky so the extremes stay dark (more negative space, and it never fights the
	// edge vignette the engine already applies at the border rows). The body of the
	// curtains lives through the middle band where env ≈ 1.
	auroraSpread = 22.0
	// auroraHueSpan is how many aspect-corrected cells of altitude one full warm→cool
	// gradient sweep spans: aux = 0.5 + dy/auroraHueSpan, so the focal band sits mid-
	// gradient, the high sky (dy<0) leans to the warm anchor and the low sky (dy>0) to
	// the cool one. Chosen so the sweep fills a typical pane's height rather than a
	// fixed cell count — an absolute field's honest reading of "hue shifts with
	// altitude": a short pane shows a slice of the gradient, a tall one the whole of it.
	auroraHueSpan = 45.0
	// auroraHueVar leans the hue a little per curtain off the pure altitude sweep (from
	// the same texture noise), so adjacent curtains differ and the palette is worked
	// rather than rendered as flat horizontal bands. Small — altitude is the premise,
	// this is only variety. It moves the hue only, never the brightness, so it can open
	// no hole.
	auroraHueVar = 0.12
	// auroraOct/auroraGain are the fBm's octave count and per-octave amplitude falloff.
	// Three octaves, not more, and the reason is the drift trap: the octaves' base
	// frequency is auroraFreqX and each doubles it, so the finest lands at
	// auroraFreqX·2^(auroraOct−1) = 0.40 cycles/cell — below the 0.5 Nyquist limit, so
	// every octave stays resolvable and none crawls as the whole pattern translates. A
	// fourth octave (0.80) would alias into a sideways shimmer the moment auroraDrift
	// moved it, which the galaxy's static turbulence can carry and this drifting field
	// cannot. See auroraFBM.
	auroraOct  = 3
	auroraGain = 0.62
)

var (
	// auroraSeed decorrelates the octaves' values; auroraOff decorrelates their lattice
	// positions (distinct odd constants / offsets, same idea as galTurbSeed/galTurbOff).
	auroraSeed = [auroraOct]uint32{0x27D4EB2F, 0x165667B1, 0x85EBCA77}
	auroraOff  = [auroraOct]float64{0, 0.37, 0.71}
)

// auroraValNoise is bilinear value noise on the lattice hash, smoothstep-interpolated
// so it is C¹ (no lattice creases). It is the octave auroraFBM stacks — aurora carries
// its own rather than borrowing the galaxy's, so neither variant's tuning can move the
// other's field.
func auroraValNoise(x, y float64, seed uint32) float64 {
	xi, yi := math.Floor(x), math.Floor(y)
	xf, yf := x-xi, y-yi
	su := xf * xf * (3 - 2*xf)
	sv := yf * yf * (3 - 2*yf)
	ix, iy := int32(xi), int32(yi)
	return splashLerp(
		splashLerp(latticeVal(ix, iy, seed), latticeVal(ix+1, iy, seed), su),
		splashLerp(latticeVal(ix, iy+1, seed), latticeVal(ix+1, iy+1, seed), su), sv)
}

// auroraFBM is the normalized fBm ([0,1], mean ~0.5) the curtain mask is carved from.
// It takes coordinates already scaled to the base frequency (the caller multiplies by
// auroraFreqX/auroraFreqY), so it only doubles from there.
//
// The octave count is load-bearing rather than a richness dial, and this is the trap
// the file header names. The whole pattern *translates* under auroraDrift, and a
// translating band-limited signal crawls once its frequency passes the grid's Nyquist
// limit — the sideways analogue of the tunnel's wagon-wheeling rings. The galaxy stacks
// five octaves past Nyquist and gets away with it because its turbulence is fixed in
// screen space and so never moves; this field's does move, so every octave must stay
// resolvable. auroraFreqX·2^(auroraOct−1) = 0.40 < 0.5 keeps the finest one safe, which
// is why the stack is three deep and the base frequency is low.
func auroraFBM(x, y float64) float64 {
	sum, amp, norm := 0.0, 1.0, 0.0
	fx, fy := x, y
	for o := 0; o < auroraOct; o++ {
		sum += amp * auroraValNoise(fx+auroraOff[o], fy+auroraOff[o], auroraSeed[o])
		norm += amp
		amp *= auroraGain
		fx *= 2
		fy *= 2
	}
	return sum / norm
}

// splashAuroraAt evaluates one cell of the aurora: how bright the curtain light is
// there, and the altitude-banded hue it wears.
//
// val is curtain·envelope: a texture mask that carves bright vertical filaments out of
// dark sky, faded toward the top and bottom of the pane. Brightness is the whole
// subject — there is no structure the hue could carry instead — so it rides
// ops.lumRange the way the galaxy's disk does. aux is altitude, spent directly by
// splashColorIdx as a gradient position, with a small per-curtain lean so the palette
// is worked rather than banded flat: the hue is a property of the light (where in the
// sky it hangs), never of the raw cell address.
//
// It is a plain function rather than a maxD closure because the curtains are drawn in
// absolute cells — a wider pane shows more of them, a taller one more of the altitude
// gradient — so nothing here measures against the pane's size (see splashFieldAt). No
// division is performed, so neither return can be NaN.
func splashAuroraAt(_, _ int, dx, dy, phase float64) (val, aux float64) {
	// Sideways drift of the whole body, plus the snaking warp of each curtain's spine.
	snake := auroraSnakeA1*math.Sin(dy*auroraSnakeF1+phase*auroraSnakeS1) +
		auroraSnakeA2*math.Sin(dy*auroraSnakeF2+phase*auroraSnakeS2)
	x := dx - phase*auroraDrift + snake

	// Anisotropic texture (tall, narrow), scrolling slowly upward along the curtains.
	// Scaling x and y here is what puts the whole pattern's translation into screen
	// cells: every octave shifts by the same screen distance under the drift, so the
	// curtains move as one body rather than shearing octave against octave.
	n := auroraFBM(x*auroraFreqX, dy*auroraFreqY-phase*auroraRise)

	// Carve the curtains out of the dark sky (see auroraLo/auroraHi).
	curtain := smoothstep(auroraLo, auroraHi, n)
	// The broad vertical envelope, centred on the focal point.
	env := math.Exp(-(dy * dy) / (auroraSpread * auroraSpread))
	val = clamp01(curtain * env)

	// Altitude sweep, leaned a little per curtain. Beyond the pane's mid-band the sweep
	// runs past the gradient ends and clamps there — the sky saturating to its highest
	// and lowest hue, which is the intended reading, not a flattened bug.
	aux = clamp01(0.5 + dy/auroraHueSpan + auroraHueVar*(n-0.5))
	return val, aux
}
