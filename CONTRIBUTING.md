# Contributing

Thanks for helping with `press`. See [`docs/manifest.md`](docs/manifest.md) for the design principles.

## Development

```sh
go build ./...
go test ./...
go vet ./...
```

## Every change needs proof that it works

Every change, whether made by a human, Claude, or any other agent, must include at least one of:

1. **An e2e test** (preferred). Add it to `e2e_test.go` (or a sibling `e2e_*_test.go`). It runs the compiled `press` binary against a temp site and asserts on output and files.
2. **A usage example.** Put the exact command(s) and real output in the PR description, and in `README.md` if a command or flag changed.

Docs-only and CI-only changes can say "n/a" in the PR, with a reason.

Bug fixes should start with a failing test, then the fix.

## Pull requests

- Work on a branch; don't push to `main`.
- Fill out the [PR template](.github/PULL_REQUEST_TEMPLATE.md), including the usage example or e2e test section.
- Keep dependencies minimal; prefer the standard library.
- Add new or changed commands to `README.md`.

## Coverage

`make coverage` runs the tests and writes `coverage.out` and `coverage.html`.
The e2e tests run a compiled `press` binary, so the target builds that binary
with `-cover` (enabled by `PRESS_E2E_COVERDIR`) and merges its coverage with the
unit-test profile. Plain `go test ./...` is unaffected.
