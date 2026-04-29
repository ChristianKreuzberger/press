# Security Policy

## Supported Versions

Only the latest release of `press` receives security updates.

| Version | Supported |
| ------- | --------- |
| latest  | ✅        |
| older   | ❌        |

## Reporting a Vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities.

Use [GitHub private vulnerability reporting](https://github.com/ChristianKreuzberger/press/security/advisories/new) to report a vulnerability confidentially.

We aim to respond within **5 business days** and will coordinate disclosure with you.

## Security Measures

This repository uses the following automated security tooling:

- **[govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)** — scans Go dependencies for known CVEs on every push and weekly
- **[CodeQL](https://codeql.github.com/)** — static application security testing (SAST) for Go on every push and weekly
- **[OSSF Scorecard](https://securityscorecards.dev/)** — supply-chain health checks on every push to `main` and weekly
- **[Dependabot](https://docs.github.com/en/code-security/dependabot)** — automatic dependency updates for Go modules and GitHub Actions (weekly)
- **[step-security/harden-runner](https://github.com/step-security/harden-runner)** — runtime security monitoring for all CI jobs
- **Pinned action SHAs** — all GitHub Actions are pinned to immutable commit SHAs to prevent supply-chain attacks
