# fresco — a one-tour recording of the variants (rain → tunnel → ripple → galaxy → aurora).
#
# Regenerate from the repo root:
#
#     vhs .github/vhs/demo.tape
#
# Requires vhs, ttyd, and ffmpeg on PATH. Output lands at docs/demo.gif.
#
# Sizing note: at FontSize 14 / Padding 16, a 640x384 window is ~64x21 cells,
# so the field is drawn at 62x17 to sit inside it with a small margin. The CLI
# handles a full-width field safely (it disables autowrap), but a pinned --size
# keeps the recording geometry deterministic and matching the framing here.
# Dimensions and framerate are kept modest on purpose — a generative field
# changes every cell every frame, so it compresses poorly and a larger/faster
# GIF balloons the file.

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
Type "go build -o /tmp/fresco ./cmd/fresco" Enter
Wait
Type "clear" Enter
Show

# ~3s per variant gives one full tour across the five-variant roster
# (rain → tunnel → ripple → galaxy → aurora): 5 × 3s. --duration exits cleanly
# through the terminal-restore teardown, so no Ctrl+C is needed.
Type "/tmp/fresco --variant cycle --seconds-per-variant 3 --size 62x17 --duration 15s" Enter
Sleep 15500ms
