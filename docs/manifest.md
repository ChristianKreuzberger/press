# press — Manifest

`press` is a Go CLI tool for generating static websites, including personal websites, blogs, and portfolios.

---

## 1. Simplicity First

`press` does one thing: turn structured content into a static site. There is no configuration sprawl, no plugin ecosystem to manage, and no magic. Every decision made in the project favours the smallest, most obvious solution.

- Single binary, zero runtime dependencies.
- Minimal flags and commands — discoverable in seconds.
- Straightforward project layout that any developer can read at a glance.

## 2. Developer and Agent Friendly

`press` is designed to be used by humans **and** AI coding agents alike.

- All behaviour is driven by CLI flags and well-defined file conventions — no hidden state.
- Output is plain text; errors go to `stderr`, content goes to `stdout` (or disk), making it easy to script and pipeline.
- The repository ships with an [`AGENTS.md`](../AGENTS.md) that gives AI agents the context they need to contribute safely.
- Public functions are tested so that agents can refactor with confidence.
- External dependencies are kept minimal (prefer the Go standard library) to reduce the surface area agents need to understand.

## 3. No Tracking, No Marketing, No Ads

`press` will never:

- Phone home, send telemetry, or collect usage data.
- Include sponsored content, ads, or affiliate links in generated output.
- Require account creation, authentication, or an internet connection to work.

The tool is open source and the generated output belongs entirely to you.

## 4. Documentation

Documentation lives in two places by design:

| Where | What |
|-------|------|
| `press --help` (built-in) | Quick reference for flags and commands, always in sync with the binary. |
| [`docs/`](.) | Deeper explanations, the project manifest, and guides — readable without installing the tool. |

Running `press --help` or any subcommand with `--help` will always show the authoritative usage for that version of the binary. The `docs/` folder supplements that with context and rationale that does not belong in a `--help` screen.
