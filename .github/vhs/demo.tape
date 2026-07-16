# fresco — a one-tour recording of the variants (rain → tunnel → ripple → galaxy).
#
# Regenerate from the repo root:
#
#     vhs .github/vhs/demo.tape
#
# Requires vhs, ttyd, and ffmpeg on PATH. Output lands at docs/demo.gif.
#
# Sizing note: at FontSize 14 / Padding 16, a 640x384 window is ~64x21 cells,
# so the field is drawn at 62x17 to sit inside it with a small margin (the field
# must never be wider than the terminal, or every row wraps). Dimensions and
# framerate are kept modest on purpose — a generative field changes every cell
# every frame, so it compresses poorly and a larger/faster GIF balloons the file.

Output docs/demo.gif

Require go

Set Shell bash
Set FontSize 14
Set Width 640
Set Height 384
Set Padding 16
Set Framerate 15
Set LoopOffset 20%

# Build first, hidden, so the recording never captures the compile step.
Hide
Type "go build -o /tmp/fresco-demo ./cmd/fresco-demo" Enter
Wait
Type "clear" Enter
Show

# ~3s per variant gives one full tour across the roster.
Type "FRESCO_DEMO_SECONDS_PER_VARIANT=3 /tmp/fresco-demo 62 17" Enter
Sleep 12200ms
Ctrl+C
Sleep 300ms
