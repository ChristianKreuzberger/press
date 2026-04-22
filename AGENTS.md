# AGENTS.md

This file describes the project for AI coding agents.

## Project

`press` is a Go CLI tool for generating static websites, including personal websites, blogs, and portfolios.

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

## Custom Agent Workflow

This project includes a specialized multi-agent workflow for code development, available in `.github/agents/`:

### Workflow Overview

When you request code changes (features, bug fixes, refactoring), the workflow automatically engages four specialized agents that collaborate and iterate until satisfied:

1. **Planning Agent** (`press-planning.agent.md`)
   - Validates requirements against project manifest and CLI structure
   - Researches common patterns if requirements are unclear
   - Creates list of open questions for other agents

2. **Go Developer Agent** (`press-go-developer.agent.md`)
   - Writes clean, idiomatic, testable Go code
   - Follows project principles (Simplicity First, stdlib preference)
   - Addresses open questions from Planning phase

3. **Architect Agent** (`press-architect.agent.md`)
   - Reviews changes with pessimistic mindset
   - Focuses on security, maintainability, and sustainability
   - Identifies technical debt and improvement opportunities

4. **Product Agent** (`press-product.agent.md`)
   - Verifies changes satisfy original requirements
   - Analyzes edge cases and error scenarios
   - Provides Go/No-Go quality gate


## Guidelines for agents

- Keep external dependencies minimal; prefer the Go standard library.
- All public functions must have tests.
- Follow standard Go formatting (`gofmt`).
- Commit messages should be short and descriptive.
- Do not commit secrets or credentials.
- When adding or updating commands, make sure they are also added to README.md.

## When creating PRs

- Demonstrate the impact to the user, if any (e.g., new CLI command, updated existing command, with example input and output)
- Fill out the [PULL_REQUEST_TEMPLATE](.github/PULL_REQUEST_TEMPLATE.md)
- Let me know if you skipped something, something was unclear, or you had to deviate from the plan
