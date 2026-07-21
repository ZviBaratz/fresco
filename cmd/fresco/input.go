package main

import "context"

// key is a decoded control action. The mapping from bytes to actions
// (decodeKey) is pure and unit-tested; the reader that produces the bytes
// (watchKeys) is the thin impure wiring around it.
type key int

const (
	keyNone key = iota // an unmapped byte — ignored
	keyQuit
	keyNext
	keyPause
)

// decodeKey maps one raw input byte to a control action. In raw mode the
// terminal stops translating Ctrl-C/Ctrl-D into signals and delivers them as
// bytes 0x03/0x04, so those must quit here too. A lone Esc (0x1b) is ignored
// rather than treated as quit: it is also the first byte of arrow-key and other
// escape sequences, and a byte-at-a-time reader can't tell them apart, so
// quitting on it would make every arrow key exit.
func decodeKey(b byte) key {
	switch b {
	case 'q', 'Q', 0x03, 0x04:
		return keyQuit
	case ' ':
		return keyNext
	case 'p', 'P':
		return keyPause
	default:
		return keyNone
	}
}

// watchKeys reads the console's input one byte at a time and forwards decoded
// control actions until the reader errors (e.g. stdin closed). It is impure
// wiring — the driver injects a fake in tests — so it holds no logic beyond the
// read loop and decodeKey.
//
// A byte read from a real terminal blocks and cannot be unblocked by ctx alone;
// on ctx cancellation this goroutine may stay parked in Read until the process
// exits, which is fine for a screensaver that is on its way out. The select on
// ctx.Done() covers the case where the consumer has already stopped receiving.
func watchKeys(ctx context.Context, c Console) <-chan key {
	ch := make(chan key)
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := c.In().Read(buf)
			if err != nil {
				return
			}
			if n == 0 {
				continue
			}
			k := decodeKey(buf[0])
			if k == keyNone {
				continue
			}
			select {
			case ch <- k:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}
