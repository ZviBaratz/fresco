package main

import (
	"context"
	"testing"
)

// decodeKey maps a raw input byte to a control action. In raw mode Ctrl-C and
// Ctrl-D arrive as bytes 0x03/0x04 (the terminal no longer turns them into a
// signal), so they must quit here; q quits, space advances, p pauses; everything
// else is ignored.
func TestDecodeKey(t *testing.T) {
	cases := map[byte]key{
		'q':  keyQuit,
		'Q':  keyQuit,
		0x03: keyQuit, // Ctrl-C
		0x04: keyQuit, // Ctrl-D
		' ':  keyNext,
		'p':  keyPause,
		'P':  keyPause,
		'x':  keyNone,
		'a':  keyNone,
		0x1b: keyNone, // lone Esc is ignored so arrow-key escape sequences don't quit
	}
	for b, want := range cases {
		if got := decodeKey(b); got != want {
			t.Errorf("decodeKey(%#x) = %v, want %v", b, got, want)
		}
	}
}

// watchKeys reads the console input byte by byte, forwarding decoded actions and
// skipping unmapped bytes, until the reader is exhausted.
func TestWatchKeysDecodesStream(t *testing.T) {
	f := newFake(true, Size{10, 3})
	f.in = &blockingReader{data: []byte("x pq")} // unmapped, space=next, p=pause, q=quit

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	keys := watchKeys(ctx, f)

	for i, want := range []key{keyNext, keyPause, keyQuit} {
		if got := <-keys; got != want {
			t.Errorf("key %d = %v, want %v", i, got, want)
		}
	}
}

// blockingReader yields its bytes one at a time, then blocks forever rather than
// returning io.EOF — mimicking a terminal stdin, whose Read parks until the next
// keypress instead of ending. The test consumes exactly the mapped keys and lets
// the deferred cancel abandon the parked goroutine.
type blockingReader struct {
	data []byte
	pos  int
	done chan struct{}
}

func (r *blockingReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		if r.done == nil {
			r.done = make(chan struct{})
		}
		<-r.done // park like a real terminal with no pending input
		return 0, nil
	}
	p[0] = r.data[r.pos]
	r.pos++
	return 1, nil
}
