// Command fresco-demo animates the fresco fields in your terminal. It cycles
// through every variant, a few seconds each, until you press Ctrl-C.
//
//	go run github.com/ZviBaratz/fresco/cmd/fresco-demo@latest
//	go run ./cmd/fresco-demo 100 30   # explicit width height
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ZviBaratz/fresco"
)

// A calm, cool-leaning palette (tokyo-night hues). Any five "#rrggbb" strings
// work; A0..A3 are the warm→cool gradient, Highlight is the star/head white.
var palette = fresco.Palette{
	A0:        "#f7768e",
	A1:        "#bb9af7",
	A2:        "#7aa2f7",
	A3:        "#7dcfff",
	Highlight: "#c0caf5",
}

func main() {
	w, h := 96, 30
	if len(os.Args) == 3 {
		if a, err := strconv.Atoi(os.Args[1]); err == nil {
			w = a
		}
		if b, err := strconv.Atoi(os.Args[2]); err == nil {
			h = b
		}
	}

	fmt.Print("\x1b[?25l")       // hide cursor
	defer fmt.Print("\x1b[?25h") // show cursor on return

	// Restore the cursor even on Ctrl-C.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Print("\x1b[?25h\n")
		os.Exit(0)
	}()

	const secondsPerVariant = 6
	const fps = 30
	variants := fresco.Variants()

	for frame := 0; ; frame++ {
		v := variants[(frame/(secondsPerVariant*fps))%len(variants)]
		field := fresco.Render(w, h, frame, fresco.Options{
			Palette:  palette,
			Variant:  v,
			FocalRow: -1, // centre the field on the pane
		})
		// Home the cursor and paint the frame plus a caption line.
		fmt.Printf("\x1b[H%s\n  fresco · %s · %d×%d · Ctrl-C to quit\x1b[K",
			field, v.String(), w, h)
		time.Sleep(time.Second / fps)
	}
}
