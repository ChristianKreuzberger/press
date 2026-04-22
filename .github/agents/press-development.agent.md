---
name: press-development
description: "Go development workflow for press CLI. When user requests code changes, features, or bug fixes: (1) Planner validates requirements against manifest and CLI structure; (2) Go Developer writes readable, testable code; (3) Architect reviews for security, maintainability, sustainability; (4) Product Agent verifies requirements are met. Agents iterate until satisfied. Use for: implementing features, fixing bugs, refactoring, adding tests."
applyTo: "**/*.go"
---

# Press Development Workflow

Orchestrate four agents for code changes. Run phases in order. Iterate if needed. Be brief.

## Phases

**0 — Plan** (`press-planning`): Validate requirement against manifest and CLI structure. Identify ambiguities. Output: Planning Summary + Open Questions.

**1 — Dev** (`press-go-developer`): Input: Planning Summary. Write code + tests. Address open questions.

**2 — Review** (`press-architect`): Flag security, maintainability, and sustainability issues. Rate each: Critical / Important / Nice-to-have.

**3 — Verify** (`press-product`): Confirm implementation satisfies original requirements. Output: Go/No-Go.

**4 — Iterate**: Concerns raised? Return to Dev. Re-run Review and Verify. Repeat until all agents satisfied.

**5 — Done**: Report final status and any remaining caveats.
