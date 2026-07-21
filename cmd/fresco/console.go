package main

import (
	"io"
	"os"

	"github.com/charmbracelet/x/term"
)

// Console is the whole terminal-syscall seam. The driver (run.go) talks only to
// this interface, so every path through it is exercised in tests with a fake:
// the restore-on-exit guarantee, the resize handling, and the frame loop are all
// verified against a bytes.Buffer, and only osConsole — a branch-free
// pass-through to os.Std*, x/term, and termenv — stays untested.
type Console interface {
	// Out is where frames and control sequences are written.
	Out() io.Writer
	// In is the key-input source (raw stdin once EnterRaw succeeds).
	In() io.Reader
	// IsTTY reports whether both stdout and stdin are real terminals; when false
	// the driver degrades to a single static frame rather than animating.
	IsTTY() bool
	// Size returns the terminal's current cell dimensions.
	Size() (Size, error)
	// EnterRaw puts the input into raw mode and returns a closure that restores
	// the prior mode. Only called on a TTY, and its failure degrades to a
	// non-interactive run rather than aborting.
	EnterRaw() (restore func() error, err error)
	// EnableVT enables virtual-terminal processing for Out and returns a restore
	// closure. It is what makes the alternate-screen/SGR/cursor sequences render
	// on a legacy Windows console; on Unix it is a no-op that never errors.
	EnableVT() (restore func() error, err error)
}

// osConsole is the real Console over os.Stdout / os.Stdin. It contains no
// branching of its own — every method is a direct call — which is why it needs
// no test.
type osConsole struct {
	out *os.File
	in  *os.File
}

func newOSConsole() *osConsole { return &osConsole{out: os.Stdout, in: os.Stdin} }

func (c *osConsole) Out() io.Writer { return c.out }
func (c *osConsole) In() io.Reader  { return c.in }

// IsTTY requires both ends to be terminals: piped input or redirected output
// each mean the screensaver should degrade. x/term.IsTerminal already answers
// false for a Cygwin/MSYS pty (where GetSize would fail anyway), so that case
// falls through to the same non-TTY degrade.
func (c *osConsole) IsTTY() bool {
	return term.IsTerminal(c.out.Fd()) && term.IsTerminal(c.in.Fd())
}

func (c *osConsole) Size() (Size, error) {
	w, h, err := term.GetSize(c.out.Fd())
	if err != nil {
		return Size{}, err
	}
	return Size{W: w, H: h}, nil
}

func (c *osConsole) EnterRaw() (func() error, error) {
	st, err := term.MakeRaw(c.in.Fd())
	if err != nil {
		return nil, err
	}
	return func() error { return term.Restore(c.in.Fd(), st) }, nil
}

// EnableVT is defined per-platform: console_vt_windows.go enables VT processing
// on the console (so the alt-screen/SGR/cursor sequences render on conhost),
// while console_vt_other.go is a no-op — termenv's function signature itself
// differs by platform (*termenv.Output on Windows, io.Writer elsewhere), so the
// call cannot be written portably in one place.
