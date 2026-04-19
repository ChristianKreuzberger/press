# AGENTS.md

This file describes the project for AI coding agents.

## Project

`press` is a Go CLI tool for generating static websites, including personal websites, blogs, and portfolios.

## Repository layout

```
.
├── main.go                    # CLI entry point
├── go.mod                     # Go module definition
├── .goreleaser.yaml           # Release configuration (goreleaser)
├── .github/
│   └── workflows/
│       ├── ci.yml             # CI: vet, test, build on PRs and main
│       └── release.yml        # Release: publish binaries on version tags
├── docs/
│   └── manifest.md            # Project manifest and design principles
└── AGENTS.md                  # This file
```

## Project manifest

See [`docs/manifest.md`](docs/manifest.md) for the design principles that guide this project: Simplicity First, Developer & Agent Friendly, No Tracking/Marketing/Ads, and how documentation is structured.

## Development

```sh
go build ./...       # build
go test ./...        # run tests
go vet ./...         # static analysis
```

## Releases

Push a tag of the form `vX.Y.Z` to trigger a GitHub Release with binaries for Linux, macOS, and Windows (amd64 and arm64).

```sh
git tag v0.1.0
git push origin v0.1.0
```

## Guidelines for agents

- Keep external dependencies minimal; prefer the Go standard library.
- All public functions must have tests.
- Follow standard Go formatting (`gofmt`).
- Commit messages should be short and descriptive.
- Do not commit secrets or credentials.
