package main

import (
	"bytes"
	"context"
	"io"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/ZviBaratz/fresco"
)

// fakeConsole is an in-memory Console: Out is a buffer, and EnterRaw/EnableVT log
// their enter/restore calls so a test can assert the teardown order.
type fakeConsole struct {
	out     *bytes.Buffer
	in      io.Reader
	tty     bool
	size    Size
	sizeErr error
	rawErr  error
	log     []string
}

func newFake(tty bool, sz Size) *fakeConsole {
	return &fakeConsole{out: &bytes.Buffer{}, in: strings.NewReader(""), tty: tty, size: sz}
}

func (f *fakeConsole) Out() io.Writer      { return f.out }
func (f *fakeConsole) In() io.Reader       { return f.in }
func (f *fakeConsole) IsTTY() bool         { return f.tty }
func (f *fakeConsole) Size() (Size, error) { return f.size, f.sizeErr }

func (f *fakeConsole) EnterRaw() (func() error, error) {
	if f.rawErr != nil {
		return nil, f.rawErr
	}
	f.log = append(f.log, "enter-raw")
	return func() error { f.log = append(f.log, "restore-raw"); return nil }, nil
}

func (f *fakeConsole) EnableVT() (func() error, error) {
	f.log = append(f.log, "enable-vt")
	return func() error { f.log = append(f.log, "restore-vt"); return nil }, nil
}

func mustConfig(t *testing.T, args ...string) config {
	t.Helper()
	cfg, err := resolveConfig(args, noEnv, true, fresco.TrueColor)
	if err != nil {
		t.Fatalf("resolveConfig(%v): %v", args, err)
	}
	return cfg
}

func testDeps(tick chan time.Time, resize chan Size, keys chan key) loopDeps {
	return loopDeps{
		newTicker:   func(time.Duration) (<-chan time.Time, func()) { return tick, func() {} },
		watchResize: func(context.Context, Console) <-chan Size { return resize },
		watchKeys:   func(context.Context, Console) <-chan key { return keys },
	}
}

// incrementalFrames counts steady-tick repaints: every frame writes a home, but
// the full paints (first paint and post-resize) write a clear+home, so the
// steady frames are the homes that aren't part of a clear.
func incrementalFrames(out string) int {
	return strings.Count(out, seqHome) - strings.Count(out, seqClearHome)
}

// setup opens the terminal and teardown restores it in strict LIFO order: raw
// off first, then autowrap/cursor/alt-screen, then VT processing last (the
// cursor and alt-screen sequences are VT output, so VT must outlive them).
func TestSetupTeardownLIFO(t *testing.T) {
	f := newFake(true, Size{80, 24})
	s, rawOn, err := setup(f, true)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if !rawOn {
		t.Error("setup with wantRaw should report raw on")
	}
	if got, want := f.out.String(), seqEnterAlt+seqHideCursor+seqAutowrapOff; got != want {
		t.Errorf("setup wrote %q, want %q", got, want)
	}
	s.teardown()
	wantBuf := seqEnterAlt + seqHideCursor + seqAutowrapOff + seqAutowrapOn + seqShowCursor + seqLeaveAlt
	if got := f.out.String(); got != wantBuf {
		t.Errorf("after teardown buffer = %q, want %q", got, wantBuf)
	}
	if want := []string{"enable-vt", "enter-raw", "restore-raw", "restore-vt"}; !slices.Equal(f.log, want) {
		t.Errorf("teardown order = %v, want %v", f.log, want)
	}
}

func TestSetupNonRawSkipsRaw(t *testing.T) {
	f := newFake(true, Size{80, 24})
	s, rawOn, err := setup(f, false)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if rawOn {
		t.Error("setup without wantRaw must not enter raw mode")
	}
	s.teardown()
	if want := []string{"enable-vt", "restore-vt"}; !slices.Equal(f.log, want) {
		t.Errorf("non-raw teardown = %v, want %v", f.log, want)
	}
}

// A failed MakeRaw degrades to a non-interactive run rather than aborting.
func TestSetupRawFailureDegrades(t *testing.T) {
	f := newFake(true, Size{80, 24})
	f.rawErr = io.ErrClosedPipe
	s, rawOn, err := setup(f, true)
	if err != nil {
		t.Fatalf("setup should not fail when raw mode is unavailable: %v", err)
	}
	if rawOn {
		t.Error("rawOn should be false when MakeRaw fails")
	}
	s.teardown()
}

func TestTeardownIdempotent(t *testing.T) {
	f := newFake(true, Size{80, 24})
	s, _, _ := setup(f, true)
	s.teardown()
	s.teardown() // second call must be a no-op, not a double-restore
	if len(f.log) != 4 {
		t.Errorf("restores ran %d times across two teardowns, want 4 total entries: %v", len(f.log), f.log)
	}
}

// The full loop: setup opens the terminal, ticks paint frames, and teardown
// closes it — restore emitted exactly once, in order, at the tail.
func TestRunLoopSetupTicksTeardown(t *testing.T) {
	f := newFake(true, Size{8, 3})
	tick := make(chan time.Time)
	deps := testDeps(tick, make(chan Size), make(chan key))
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- runLoop(ctx, f, mustConfig(t, "--variant", "rain"), deps) }()
	tick <- time.Time{}
	tick <- time.Time{}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("runLoop: %v", err)
	}

	out := f.out.String()
	if !strings.HasPrefix(out, seqEnterAlt+seqHideCursor+seqAutowrapOff) {
		t.Error("loop must open with alt-screen, hidden cursor, autowrap off")
	}
	if !strings.HasSuffix(out, seqAutowrapOn+seqShowCursor+seqLeaveAlt) {
		t.Error("loop must close by restoring autowrap, cursor, primary screen")
	}
	if n := incrementalFrames(out); n != 2 {
		t.Errorf("two ticks should paint two incremental frames, got %d", n)
	}
	if want := []string{"enable-vt", "enter-raw", "restore-raw", "restore-vt"}; !slices.Equal(f.log, want) {
		t.Errorf("teardown order = %v, want %v", f.log, want)
	}
}

// A resize triggers a full clear+repaint so stale cells from a larger field go.
func TestAppRunResizeRepaintsFully(t *testing.T) {
	f := newFake(true, Size{8, 3})
	tick, resize, keys := make(chan time.Time), make(chan Size), make(chan key)
	a := &app{c: f, cfg: mustConfig(t, "--variant", "rain"), deps: testDeps(tick, resize, keys)}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- a.run(ctx) }()
	resize <- Size{20, 6}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("run: %v", err)
	}
	if n := strings.Count(f.out.String(), seqClearHome); n < 2 {
		t.Errorf("want ≥2 full paints (initial + resize), got %d", n)
	}
}

// q quits the loop without needing a signal.
func TestAppRunKeyQuitReturns(t *testing.T) {
	f := newFake(true, Size{8, 3})
	tick, resize, keys := make(chan time.Time), make(chan Size), make(chan key)
	a := &app{c: f, cfg: mustConfig(t, "--variant", "rain"), deps: testDeps(tick, resize, keys)}

	done := make(chan error, 1)
	go func() { done <- a.run(context.Background()) }()
	keys <- keyQuit
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("keyQuit did not stop the loop")
	}
}

// p pauses: ticks while paused paint nothing; unpausing resumes.
func TestAppRunPauseHaltsFrames(t *testing.T) {
	f := newFake(true, Size{8, 3})
	tick, resize, keys := make(chan time.Time), make(chan Size), make(chan key)
	a := &app{c: f, cfg: mustConfig(t, "--variant", "rain"), deps: testDeps(tick, resize, keys)}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- a.run(ctx) }()
	keys <- keyPause    // pause
	tick <- time.Time{} // ignored
	tick <- time.Time{} // ignored
	keys <- keyPause    // resume
	tick <- time.Time{} // paints
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("run: %v", err)
	}
	if n := incrementalFrames(f.out.String()); n != 1 {
		t.Errorf("only the post-resume tick should paint; got %d incremental frames", n)
	}
}

// A pinned --size drives the animated loop too (not just --once): the field is
// rendered at the pinned size even when the terminal reports something else, and
// terminal resizes don't override it.
func TestAppRunHonorsPinnedSize(t *testing.T) {
	f := newFake(true, Size{40, 5}) // the terminal claims 40×5...
	tick, resize, keys := make(chan time.Time), make(chan Size), make(chan key)
	a := &app{c: f, cfg: mustConfig(t, "--variant", "rain", "--size", "12x3"), deps: testDeps(tick, resize, keys)}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- a.run(ctx) }()
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("run: %v", err)
	}

	body := strings.TrimPrefix(f.out.String(), seqClearHome)
	if i := strings.IndexByte(body, '\n'); i >= 0 {
		body = body[:i]
	}
	if n := utf8.RuneCountInString(body); n != 12 {
		t.Errorf("pinned --size 12x3 should render 12-wide rows, got %d (used the terminal size, not the pin)", n)
	}
}

// space forces a full repaint at the next variant.
func TestAppRunKeyNextRepaints(t *testing.T) {
	f := newFake(true, Size{8, 3})
	tick, resize, keys := make(chan time.Time), make(chan Size), make(chan key)
	a := &app{c: f, cfg: mustConfig(t, "--variant", "rain"), deps: testDeps(tick, resize, keys)}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- a.run(ctx) }()
	keys <- keyNext
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("run: %v", err)
	}
	if n := strings.Count(f.out.String(), seqClearHome); n < 2 {
		t.Errorf("keyNext should force a full repaint; got %d full paints", n)
	}
}

// In cycle mode, next jumps the frame counter to the next variant's window.
func TestNextVariantCycleJumpsToBoundary(t *testing.T) {
	a := &app{cfg: mustConfig(t, "--variant", "cycle", "--fps", "10", "--seconds-per-variant", "2")}
	// framesPer = 10 × 2 = 20.
	if got := a.nextVariant(5); got != 20 {
		t.Errorf("nextVariant(5) = %d, want 20 (start of window 1)", got)
	}
	if got := a.nextVariant(25); got != 40 {
		t.Errorf("nextVariant(25) = %d, want 40 (start of window 2)", got)
	}
}

// In pinned mode, next rotates the pin through the roster and wraps.
func TestNextVariantPinnedRotatesRoster(t *testing.T) {
	a := &app{cfg: mustConfig(t, "--variant", "rain")}
	a.nextVariant(0)
	if a.cfg.schedule.pinned != fresco.Tunnel {
		t.Errorf("next from rain pinned %v, want tunnel", a.cfg.schedule.pinned)
	}
	a.cfg.schedule.pinned = fresco.Aurora
	a.nextVariant(0)
	if a.cfg.schedule.pinned != fresco.Rain {
		t.Errorf("next from aurora should wrap to rain, got %v", a.cfg.schedule.pinned)
	}
}

// --once writes a single frame to the primary screen — no alternate screen, no
// cursor hiding, no teardown sequences.
func TestRenderOnceSingleFrameNoAltScreen(t *testing.T) {
	f := newFake(true, Size{10, 4})
	cfg := mustConfig(t, "--once", "--variant", "rain")
	if err := runApp(context.Background(), f, cfg); err != nil {
		t.Fatalf("runApp --once: %v", err)
	}
	out := f.out.String()
	if strings.Contains(out, seqEnterAlt) || strings.Contains(out, seqHideCursor) {
		t.Error("--once must not touch the alternate screen or hide the cursor")
	}
	if !strings.Contains(out, "\n") {
		t.Error("--once should still emit a rendered frame")
	}
}

// A non-TTY console degrades to a single frame even without --once.
func TestRunAppNonTTYDegrades(t *testing.T) {
	f := newFake(false, Size{})
	cfg, err := resolveConfig(nil, noEnv, false, fresco.TrueColor)
	if err != nil {
		t.Fatalf("resolveConfig: %v", err)
	}
	if err := runApp(context.Background(), f, cfg); err != nil {
		t.Fatalf("runApp non-TTY: %v", err)
	}
	if strings.Contains(f.out.String(), seqEnterAlt) {
		t.Error("a non-TTY run must not enter the alternate screen")
	}
	if f.out.Len() == 0 {
		t.Error("a non-TTY run should still render one frame")
	}
}

func TestPrintList(t *testing.T) {
	f := newFake(true, Size{80, 24})
	cfg := mustConfig(t, "--list")
	if err := runApp(context.Background(), f, cfg); err != nil {
		t.Fatalf("runApp --list: %v", err)
	}
	out := f.out.String()
	for _, want := range []string{"rain", "aurora", "tokyo-night"} {
		if !strings.Contains(out, want) {
			t.Errorf("--list output missing %q:\n%s", want, out)
		}
	}
}
