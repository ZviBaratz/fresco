package main

import "github.com/ZviBaratz/fresco"

// variantSchedule decides which variant a given frame shows. It is pure over the
// frame number, so the whole "what plays when" policy is unit-testable without a
// clock: the driver advances a frame counter and asks the schedule, rather than
// tracking time itself.
type variantSchedule struct {
	// cycle rotates through pool; when false the schedule is pinned to a single
	// variant and pool/framesPer are unused.
	cycle bool
	// pinned is the variant shown when cycle is false.
	pinned fresco.Variant
	// pool is the rotation, in order, when cycle is true.
	pool []fresco.Variant
	// framesPer is how many frames each variant holds before the next
	// (secondsPerVariant × fps), computed once at config time.
	framesPer int
}

// at returns the variant for the given frame.
func (s variantSchedule) at(frame int) fresco.Variant {
	if !s.cycle {
		return s.pinned
	}
	if len(s.pool) == 0 {
		return fresco.Rain // the roster fallback; config never builds an empty pool
	}
	if s.framesPer <= 0 {
		return s.pool[0] // guard: a zero hold would divide by zero below
	}
	return s.pool[(frame/s.framesPer)%len(s.pool)]
}
