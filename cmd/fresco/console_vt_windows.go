//go:build windows

package main

import "github.com/muesli/termenv"

// EnableVT turns on virtual-terminal processing for the console so the
// alternate-screen, SGR, and cursor sequences render instead of printing as
// literal escape bytes on a legacy conhost. x/term.MakeRaw enables VT *input*
// but not this output flag, so it must be set separately. The returned closure
// restores the prior console mode (a no-op if VT was already on).
func (c *osConsole) EnableVT() (func() error, error) {
	return termenv.EnableVirtualTerminalProcessing(termenv.NewOutput(c.out))
}
