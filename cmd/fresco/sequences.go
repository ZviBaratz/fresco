package main

import "github.com/ZviBaratz/fresco"

// The ANSI control sequences the driver writes, and the per-frame byte
// composition built on them. These are pure data and a pure function: nothing
// here touches the terminal, the clock, or a global — the driver (run.go) owns
// the writes, so this file stays unit-testable against a bytes.Buffer.
//
// The sequences are kept as named constants rather than inlined so the teardown
// stack and its test refer to the same bytes: a restore assertion in run_test.go
// checks for exactly these strings, which is the whole point of pinning them.
const (
	// seqEnterAlt / seqLeaveAlt switch the alternate screen buffer on and off.
	// Entering it means the animation never scrolls the user's shell history;
	// leaving it in teardown restores that history untouched — the thing the
	// earlier primary-screen, cursor-homing demo could not do.
	seqEnterAlt = "\x1b[?1049h"
	seqLeaveAlt = "\x1b[?1049l"

	// seqHideCursor / seqShowCursor hide the blinking cursor while the field
	// animates and restore it on the way out.
	seqHideCursor = "\x1b[?25l"
	seqShowCursor = "\x1b[?25h"

	// seqAutowrapOff / seqAutowrapOn toggle DECAWM. A full-bleed field is exactly
	// as wide as the pane, so writing its last column would auto-wrap the cursor
	// to the next row and desync every following line; disabling autowrap for the
	// run keeps the last column honest. Restored in teardown.
	seqAutowrapOff = "\x1b[?7l"
	seqAutowrapOn  = "\x1b[?7h"

	// seqHome moves the cursor to the top-left without clearing — the cheap
	// per-frame repaint, since every cell is overwritten anyway.
	seqHome = "\x1b[H"
	// seqClearHome erases the whole pane then homes. Used for the first paint and
	// after a resize, where stale cells from a larger previous field must go.
	seqClearHome = "\x1b[2J\x1b[H"
)

// Size is a terminal's cell dimensions. It is the currency between the size
// query, the resize watcher, and the frame composer.
type Size struct{ W, H int }

// appendFrame appends one rendered frame to dst and returns the extended slice,
// reusing the caller's buffer across ticks (the alloc-free AppendRender path).
//
// full selects the leading control sequence: a clear+home for the first paint,
// after a resize, and on a variant switch, a bare home for every steady tick.
// opts is passed through
// verbatim — in particular opts.Profile must already be a pinned depth, never
// Auto, so a caller reusing one Options across frames renders deterministically
// without re-probing stdout (see resolveProfile).
func appendFrame(dst []byte, size Size, frame int, opts fresco.Options, full bool) []byte {
	if full {
		dst = append(dst, seqClearHome...)
	} else {
		dst = append(dst, seqHome...)
	}
	return fresco.AppendRender(dst, size.W, size.H, frame, opts)
}
