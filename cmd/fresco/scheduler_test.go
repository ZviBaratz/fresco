package main

import (
	"testing"

	"github.com/ZviBaratz/fresco"
)

// A pinned schedule ignores the frame and always yields the one variant.
func TestVariantSchedulerPinned(t *testing.T) {
	s := variantSchedule{pinned: fresco.Ripple}
	for _, frame := range []int{0, 1, 100, 999} {
		if got := s.at(frame); got != fresco.Ripple {
			t.Errorf("at(%d) = %v, want ripple", frame, got)
		}
	}
}

// A cycling schedule holds each variant for framesPer frames, then advances, and
// wraps around the pool.
func TestVariantScheduleCycleAdvancesAndWraps(t *testing.T) {
	s := variantSchedule{
		cycle:     true,
		pool:      []fresco.Variant{fresco.Rain, fresco.Tunnel},
		framesPer: 3,
	}
	want := []fresco.Variant{
		fresco.Rain, fresco.Rain, fresco.Rain, // frames 0..2
		fresco.Tunnel, fresco.Tunnel, fresco.Tunnel, // frames 3..5
		fresco.Rain, // frame 6 wraps
	}
	for frame, w := range want {
		if got := s.at(frame); got != w {
			t.Errorf("at(%d) = %v, want %v", frame, got, w)
		}
	}
}

// A non-positive framesPer must not divide-by-zero; it degrades to the first
// variant in the pool.
func TestVariantScheduleCycleGuardsZeroFramesPer(t *testing.T) {
	s := variantSchedule{cycle: true, pool: fresco.Variants(), framesPer: 0}
	if got := s.at(42); got != fresco.Variants()[0] {
		t.Errorf("at(42) with framesPer=0 = %v, want %v", got, fresco.Variants()[0])
	}
}
