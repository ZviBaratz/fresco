//go:build !windows

package main

// EnableVT is a no-op off Windows: Unix terminals process VT sequences natively,
// so there is nothing to enable and nothing to restore. (termenv's own
// EnableVirtualTerminalProcessing is likewise a no-op here; this avoids importing
// it — and its platform-divergent signature — on non-Windows builds.)
func (c *osConsole) EnableVT() (func() error, error) {
	return func() error { return nil }, nil
}
