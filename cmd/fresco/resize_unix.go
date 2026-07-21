//go:build unix

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// watchResize sends the terminal's new size whenever a SIGWINCH arrives, until
// ctx is done. The signal carries no dimensions, so the size is always
// re-queried through the Console; sends coalesce via sendLatest. This is
// event-driven — no idle wakeups while the field merely animates, which matters
// for a screensaver that may sit for hours.
func watchResize(ctx context.Context, c Console) <-chan Size {
	ch := make(chan Size, 1)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)
	go func() {
		defer signal.Stop(sig)
		for {
			select {
			case <-ctx.Done():
				return
			case <-sig:
				if sz, err := c.Size(); err == nil {
					sendLatest(ch, sz)
				}
			}
		}
	}()
	return ch
}
