// Command preview is the render-and-look beauty gate for authoring a fresco
// variant. It renders a variant in TrueColor so you can tune lumRange by eye and
// confirm the aux channel's hue.
//
// It lives *inside* the fresco module on purpose: `go run` it by path and the
// import below resolves to your local, in-progress code — the new variant you are
// registering right now — with no go.mod, temp dir, or replace directive. The
// dot-directory keeps it invisible to `go build ./...`, `go test ./...`, and the
// linter, so it never touches the package or CI.
//
//	# live terminal: watch the colour animation move
//	go run ./.claude/skills/new-variant/preview
//
//	# sweep lumRange by eye — run each and compare (see SKILL.md §3)
//	go run ./.claude/skills/new-variant/preview 0
//	go run ./.claude/skills/new-variant/preview 0.5
//	go run ./.claude/skills/new-variant/preview 0.75
//	go run ./.claude/skills/new-variant/preview 1
//
//	# no TTY (an agent, CI, a pipe): emits ONE TrueColor frame and exits, so you
//	# can inspect the emitted SGR bytes instead of killing an endless loop
//	go run ./.claude/skills/new-variant/preview 0.75 | less -R
//
// Edit the Variant field to your own variant before running.
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ZviBaratz/fresco"
)

func main() {
	opts := fresco.Options{
		Palette:  fresco.Palette{A0: "#f7768e", A1: "#bb9af7", A2: "#7aa2f7", A3: "#7dcfff", Highlight: "#c0caf5"},
		Variant:  fresco.Galaxy, // <- your variant
		Profile:  fresco.TrueColor,
		FocalRow: -1, // centre on the pane
	}
	if len(os.Args) > 1 {
		if r, err := strconv.ParseFloat(os.Args[1], 64); err == nil {
			opts.LumRange = &r // sweep this from the command line
		}
	}

	// A non-TTY stdout (an agent, CI, a pipe) can't watch an animation, so emit one
	// TrueColor frame and exit — its SGR bytes are what you inspect to confirm the
	// hue tracks aux. A real terminal gets the live loop.
	if info, _ := os.Stdout.Stat(); info.Mode()&os.ModeCharDevice == 0 {
		fmt.Println(fresco.Render(96, 30, 42, opts))
		return
	}
	fmt.Print("\x1b[?25l")
	defer fmt.Print("\x1b[?25h")
	for frame := 0; ; frame++ {
		fmt.Printf("\x1b[H%s\n", fresco.Render(96, 30, frame, opts))
		time.Sleep(time.Second / 30)
	}
}
