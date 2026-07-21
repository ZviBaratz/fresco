package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ZviBaratz/fresco"
)

// The impure driver: the teardown stack, terminal setup, and the frame loop that
// ties the ticker, resize watcher, and key reader together. Everything here talks
// to the terminal only through the Console seam and takes its clock/resize/key
// sources as injectable funcs (loopDeps), so the whole thing is exercised in
// run_test.go against a fake console and hand-driven channels — no real TTY.

// session is a LIFO stack of restore closures. setup pushes them, teardown pops
// them in reverse; teardown is idempotent and best-effort so it is safe to call
// from a defer regardless of how the run ended.
type session struct {
	out  io.Writer
	undo []func() error
}

func (s *session) writeSeq(seq string) error {
	_, err := io.WriteString(s.out, seq)
	return err
}

func (s *session) push(f func() error) { s.undo = append(s.undo, f) }

func (s *session) teardown() {
	for i := len(s.undo) - 1; i >= 0; i-- {
		_ = s.undo[i]() // best-effort: a broken pipe here can't be helped
	}
	s.undo = nil // idempotent: a second call restores nothing
}

// setup enables VT processing, switches to the alternate screen, hides the
// cursor, disables autowrap, and (when wantRaw) enters raw mode — pushing each
// undo so teardown reverses them in the forced order (raw off first, VT off
// last). It reports whether raw mode actually engaged; a MakeRaw failure degrades
// to a non-interactive run rather than aborting. A write failure mid-setup tears
// down whatever succeeded and returns the error.
func setup(c Console, wantRaw bool) (*session, bool, error) {
	s := &session{out: c.Out()}

	// VT processing first so the sequences below actually render on a legacy
	// Windows console; its restore is pushed first and therefore runs last. Best
	// effort: if it can't be enabled we proceed (modern terminals have it on).
	if restore, err := c.EnableVT(); err == nil {
		s.push(restore)
	}

	for _, step := range []struct{ on, off string }{
		{seqEnterAlt, seqLeaveAlt},
		{seqHideCursor, seqShowCursor},
		{seqAutowrapOff, seqAutowrapOn},
	} {
		if err := s.writeSeq(step.on); err != nil {
			s.teardown()
			return nil, false, err
		}
		off := step.off
		s.push(func() error { return s.writeSeq(off) })
	}

	rawOn := false
	if wantRaw {
		if restore, err := c.EnterRaw(); err == nil {
			s.push(restore)
			rawOn = true
		}
	}
	return s, rawOn, nil
}

// loopDeps are the driver's time/resize/key sources, injected so the loop is
// testable without a real clock, signals, or terminal.
type loopDeps struct {
	newTicker   func(time.Duration) (<-chan time.Time, func())
	watchResize func(context.Context, Console) <-chan Size
	watchKeys   func(context.Context, Console) <-chan key
}

func defaultLoopDeps() loopDeps {
	return loopDeps{newTicker: realTicker, watchResize: watchResize, watchKeys: watchKeys}
}

func realTicker(d time.Duration) (<-chan time.Time, func()) {
	t := time.NewTicker(d)
	return t.C, t.Stop
}

// app is one interactive run: the resolved config plus its injected sources.
type app struct {
	c    Console
	cfg  config
	deps loopDeps
}

// run animates until ctx is cancelled (signal or --duration), q is pressed, or a
// write fails. Every exit is a return — never os.Exit — so the caller's deferred
// teardown always runs.
func (a *app) run(ctx context.Context) error {
	tick, stopTick := a.deps.newTicker(a.tickInterval())
	defer stopTick()

	sz, err := a.initialSize()
	if err != nil {
		return err
	}

	// A pinned --size fixes the geometry, so terminal resizes are ignored (a nil
	// channel never fires in the select); otherwise the field tracks the terminal.
	var resizes <-chan Size
	if (a.cfg.size == Size{}) {
		resizes = a.deps.watchResize(ctx, a.c)
	}

	var keys <-chan key
	if a.cfg.raw {
		keys = a.deps.watchKeys(ctx, a.c)
	}
	buf := make([]byte, 0, frameBufCap(sz))
	frame := 0
	paused := false
	if err := a.paint(&buf, sz, frame, true); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case sz = <-resizes:
			if err := a.paint(&buf, sz, frame, true); err != nil {
				return err
			}
		case k := <-keys:
			switch k {
			case keyQuit:
				return nil
			case keyNext:
				frame = a.nextVariant(frame)
				if err := a.paint(&buf, sz, frame, true); err != nil {
					return err
				}
			case keyPause:
				paused = !paused
			}
		case <-tick:
			if paused {
				continue
			}
			frame++
			if err := a.paint(&buf, sz, frame, false); err != nil {
				return err
			}
		}
	}
}

// paint composes one frame into buf (reusing its storage) and writes it. A write
// error is returned so a vanished reader unwinds the loop into teardown instead
// of spinning forever.
func (a *app) paint(buf *[]byte, sz Size, frame int, full bool) error {
	*buf = appendFrame((*buf)[:0], sz, frame, frameOptions(a.cfg, frame), full)
	_, err := a.c.Out().Write(*buf)
	return err
}

// initialSize is the size the loop renders at: a pinned --size when set,
// otherwise the terminal's current size. onceSize does the same for --once, so
// both paths honour --size identically.
func (a *app) initialSize() (Size, error) {
	if (a.cfg.size != Size{}) {
		return a.cfg.size, nil
	}
	return a.c.Size()
}

func (a *app) tickInterval() time.Duration {
	if a.cfg.fps < 1 {
		return time.Second
	}
	return time.Second / time.Duration(a.cfg.fps)
}

// nextVariant handles the space key. When cycling it jumps the frame counter to
// the next variant's boundary; when a single variant is pinned it re-pins to the
// next one in the roster.
func (a *app) nextVariant(frame int) int {
	if a.cfg.schedule.cycle {
		fp := a.cfg.schedule.framesPer
		if fp < 1 {
			return frame
		}
		return ((frame / fp) + 1) * fp
	}
	pool := fresco.Variants()
	i := 0
	for j, v := range pool {
		if v == a.cfg.schedule.pinned {
			i = j
			break
		}
	}
	a.cfg.schedule.pinned = pool[(i+1)%len(pool)]
	return frame
}

// frameOptions builds the render options for a frame, resolving the scheduled
// variant. The profile is already pinned (never Auto), so this allocates nothing
// impure and yields deterministic bytes.
func frameOptions(cfg config, frame int) fresco.Options {
	return fresco.Options{
		Palette:  cfg.palette,
		Variant:  cfg.schedule.at(frame),
		FocalRow: cfg.focalRow,
		LumRange: cfg.lumRange,
		Profile:  cfg.profile,
	}
}

func frameBufCap(sz Size) int {
	cells := sz.W * sz.H
	if cells < 1 {
		cells = 1
	}
	return cells * 8 // ~8 bytes/cell covers a truecolor SGR run
}

// runApp is the top-level dispatch: list and once are one-shot and need no
// terminal setup; otherwise it wraps the interactive loop in setup/teardown and
// an optional duration deadline.
func runApp(ctx context.Context, c Console, cfg config) error {
	switch {
	case cfg.list:
		return printList(c.Out())
	case cfg.once:
		return renderOnce(c, cfg)
	}

	if cfg.duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.duration)
		defer cancel()
	}
	return runLoop(ctx, c, cfg, defaultLoopDeps())
}

// runLoop opens the terminal, runs the frame loop, and guarantees teardown on
// every exit path via the deferred restore.
func runLoop(ctx context.Context, c Console, cfg config, deps loopDeps) error {
	s, rawOn, err := setup(c, cfg.raw)
	if err != nil {
		return err
	}
	defer s.teardown()

	a := &app{c: c, cfg: cfg, deps: deps}
	a.cfg.raw = rawOn // only wire keys if raw mode actually engaged
	return a.run(ctx)
}

// renderOnce writes a single frame to the primary screen and exits — the --once
// path, and how a non-TTY run degrades. No alternate screen, no cursor hiding:
// the output is meant to be piped, redirected, or captured.
func renderOnce(c Console, cfg config) error {
	sz := onceSize(c, cfg)
	buf := fresco.AppendRender(nil, sz.W, sz.H, 0, frameOptions(cfg, 0))
	buf = append(buf, '\n')
	_, err := c.Out().Write(buf)
	return err
}

// onceSize resolves the size for a one-shot render: an explicit --size, else the
// terminal's size when it is a TTY, else a conventional 80×24 fallback so a pipe
// or CI run never blocks on or fails a size query.
func onceSize(c Console, cfg config) Size {
	if (cfg.size != Size{}) {
		return cfg.size
	}
	if c.IsTTY() {
		if sz, err := c.Size(); err == nil && sz.W > 0 && sz.H > 0 {
			return sz
		}
	}
	return Size{W: 80, H: 24}
}

// printList prints the variants and palettes, for --list.
func printList(w io.Writer) error {
	var b strings.Builder
	b.WriteString("Variants:\n")
	for _, v := range fresco.Variants() {
		fmt.Fprintf(&b, "  %s\n", v)
	}
	b.WriteString("\nPalettes:\n")
	for _, name := range paletteNames() {
		fmt.Fprintf(&b, "  %s\n", name)
	}
	_, err := io.WriteString(w, b.String())
	return err
}
