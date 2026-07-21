package main

// sendLatest delivers the newest size on a size-1 channel, first dropping any
// previously queued-but-unconsumed size, so a burst of resizes coalesces to
// "re-render at the latest size" instead of a backlog the loop has to drain.
// Safe with a single producer goroutine, which is how both watchers use it.
func sendLatest(ch chan Size, sz Size) {
	select {
	case <-ch: // discard a stale pending size
	default:
	}
	select {
	case ch <- sz:
	default:
	}
}
