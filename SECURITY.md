# Security Policy

fresco is a pure rendering library: it turns `(width, height, frame, Options)`
into a string and performs no I/O, no network access, and no file access. The
attack surface is correspondingly small. Still, if you find a vulnerability —
for example an input that causes a panic, an unbounded allocation, or a
non-terminating render — we want to know.

## Supported versions

fresco is pre-1.0. Security fixes land on the latest `v0.x` release; there is no
back-porting to earlier tags.

| Version | Supported |
|---------|-----------|
| latest `v0.x` | ✅ |
| older tags | ❌ |

## Reporting a vulnerability

Please report privately, not in a public issue:

- Preferred: open a [private security advisory](https://github.com/ZviBaratz/fresco/security/advisories/new)
  on GitHub, or
- email **z.baratz@gmail.com**.

Include a minimal reproduction (the `width`, `height`, `frame`, and `Options`
that trigger it) if you can. We aim to acknowledge within a few days and will
credit reporters who wish to be named once a fix ships.
