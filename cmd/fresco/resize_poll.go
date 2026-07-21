//go:build !unix

package main

import (
	"context"
	"time"
)

// resizePollInterval is how often the non-Unix watcher re-queries the terminal
// size. Windows (and plan9/wasm) have no SIGWINCH, so a short poll is the
// portable option; 250ms is imperceptible for a window resize and cheap.
const resizePollInterval = 250 * time.Millisecond

// watchResize polls the terminal size and sends it whenever it changes, until
// ctx is done — the portable fallback for platforms without SIGWINCH. It funnels
// through the same Console.Size the Unix watcher uses, so the only
// platform-specific thing here is how the wakeup arrives, never what a resize
// means.
func watchResize(ctx context.Context, c Console) <-chan Size {
	ch := make(chan Size, 1)
	go func() {
		t := time.NewTicker(resizePollInterval)
		defer t.Stop()
		last, _ := c.Size()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if sz, err := c.Size(); err == nil && sz != last {
					last = sz
					sendLatest(ch, sz)
				}
			}
		}
	}()
	return ch
}
