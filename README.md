# press

[![GitHub release](https://img.shields.io/github/v/release/ChristianKreuzberger/press)](https://github.com/ChristianKreuzberger/press/releases/latest)
[![CI](https://github.com/ChristianKreuzberger/press/workflows/CI/badge.svg)](https://github.com/ChristianKreuzberger/press/actions/workflows/ci.yml)
[![Go version](https://img.shields.io/github/go-mod/go-version/ChristianKreuzberger/press)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/ChristianKreuzberger/press)](https://goreportcard.com/report/github.com/ChristianKreuzberger/press)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> A fast, single-binary static-site generator built for developers and AI agents alike.

`press` turns structured Markdown content into a clean static website — no config files, no plugin ecosystem, no magic. One binary. Zero runtime dependencies.

---

## Installation

### One-liner (Linux & macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/ChristianKreuzberger/press/main/install.sh | bash
```

The script automatically detects your OS and CPU architecture, downloads the correct binary from the [latest GitHub Release](https://github.com/ChristianKreuzberger/press/releases/latest), and installs it to `/usr/local/bin` (or `~/.local/bin` when you don't have root access).

### Manual download

Pre-built binaries for Linux, macOS, and Windows (amd64 & arm64) are attached to every [release](https://github.com/ChristianKreuzberger/press/releases).

### Build from source

```bash
go install github.com/ChristianKreuzberger/press@latest
```

---

## Quick start

```bash
# 1. Scaffold a new site in the current directory
press init

# 2. Create your first page (optionally inside a section)
press create page about
press create page blog/my-first-post

# 3. Run it locally
press serve

# 4. Edit things and observe changes
```

### All commands

| Command | Description |
|---------|-------------|
| `press init [dir] [--theme name]` | Scaffold a new site (`template.html` + `pages/`); choose a built-in theme |
| `press list page` | List all pages |
| `press create page <name> [--file f.md]` | Create a new page; `name` may include sections (e.g. `blog/my-post`, `blog/2026/my-post`) |
| `press update page <name> --file f.md` | Replace a page's content |
| `press delete page <name>` | Delete a page |
| `press list section` | List all sections |
| `press create section <name> [--file f.md]` | Create a new section (folder + `index.md`) |
| `press update section <name> --file f.md` | Replace a section's index content |
| `press delete section <name>` | Delete a section and all its pages |
| `press build [-output dir]` | Build the site into `dist/` (default) |
| `press serve [-port N] [-output dir]` | Build and serve the site locally; rebuilds on file changes |
| `press tree` | Show a tree of all pages and sections |
| `press check` | Validate pages and internal links; exits with code 1 if issues are found |
| `press --version` | Print the installed version |

Run any command with `--help` for detailed usage.

---

## Why press?

Most static-site generators come with a steep learning curve: elaborate directory conventions, dozens of config knobs, and a plugin ecosystem you need to understand before you can publish a single page. `press` takes the opposite approach:

- **One binary, nothing else.** Drop it on any machine and it works.
- **Predictable layout.** `pages/` holds your content, `template.html` is your theme, `dist/` is the output. That's the whole mental model.
- **Agent friendly.** Every action is a CLI flag or subcommand. No hidden state. Easy to script, pipeline, or drive from an AI agent.
- **Your output is yours.** No telemetry, no analytics, no phoning home — ever.

Perfect for personal websites, project landing pages, technical blogs, portfolios, and documentation sites.

---

## Themes

press ships three built-in themes. Choose one when scaffolding a new site:

```bash
press init --theme dark      # default — GitHub-inspired dark mode
press init --theme light     # clean editorial style with serif headings
press init --theme terminal  # retro green-on-black, all monospace
```

The selected theme is written to `template.html`. You can edit it freely afterwards or replace it entirely with your own design.

See [`docs/themes.md`](docs/themes.md) to learn how templates work and how to create a custom theme.

---

## Project layout

```
my-site/
├── template.html        # Your HTML theme ({{.Title}} and {{.Content}} are injected)
├── pages/
│   ├── index.md         # Becomes dist/index.html
│   ├── about.md         # Becomes dist/about.html
│   └── blog/            # A section (group of related pages)
│       ├── index.md     # Becomes dist/blog/index.html (section landing page)
│       └── my-post.md   # Becomes dist/blog/my-post.html
└── dist/                # Generated output (created by `press build`)
```

### Pages and sections

- **Pages** live directly under `pages/` and become top-level HTML files (e.g. `dist/about.html`).
- **Sections** are subdirectories under `pages/`. Each section must contain an `index.md` that acts as its landing page. Other `.md` files in the directory become pages within the section.
- Navigation generated by the default template links top-level pages by name and sections by their index page.

---

## Frontmatter

Every `.md` file starts with a YAML frontmatter block created automatically by `press create page` and `press create section`:

```yaml
---
title: "My Page"
alias: ""
tags: []
weight: 0
created_at: "2026-04-22T10:00:00Z"
updated_at: "2026-04-22T10:00:00Z"
---
```

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Display name used in navigation and the page `<title>` |
| `alias` | string | Alternative URL slug (reserved for future use) |
| `tags` | list | Content tags (reserved for future use) |
| `weight` | integer | Controls the order of pages and sections in navigation. Lower values appear first. A value of `0` (default) means the item is sorted after all weighted items, in filesystem order. |
| `created_at` | RFC 3339 | Creation timestamp |
| `updated_at` | RFC 3339 | Last-updated timestamp |

### Navigation ordering with `weight`

Set `weight` on any page or section to control where it appears in the generated navigation:

```yaml
---
title: "About"
weight: 1
---
```

```yaml
---
title: "Blog"
weight: 2
---
```

Pages without a weight (or `weight: 0`) are listed after all weighted items in their natural filesystem order.

---

## Markdown extensions

press supports a small set of shortcodes on top of standard GitHub-Flavored Markdown.

### YouTube embed

Place a `!youtube[VIDEO_ID]` shortcode on its own line to embed a responsive YouTube player:

```markdown
!youtube[dQw4w9WgXcQ]
```

The `VIDEO_ID` is the 11-character identifier from the YouTube URL (e.g. `https://www.youtube.com/watch?v=dQw4w9WgXcQ`).

The shortcode renders as a privacy-enhanced iframe (`youtube-nocookie.com`) that fills 100% of the column width and maintains a 16:9 aspect ratio:

```html
<iframe style="width:100%;aspect-ratio:16/9;"
  src="https://www.youtube-nocookie.com/embed/dQw4w9WgXcQ"
  title="YouTube video player"
  allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
  allowfullscreen></iframe>
```

The shortcode is only expanded when it appears on a standalone line. It is left as-is inside fenced code blocks (` ``` `) and inline code spans (`` ` ``), so you can document it safely.

---

## AI Agent Skill

`press` ships a GitHub Copilot skill that lets AI agents install and use the tool without any extra prompting. Once installed, the agent automatically knows how to scaffold sites, manage pages and sections, build, and serve.

### Install the skill

Download the skill into your project's `.github/skills/` directory:

```bash
mkdir -p .github/skills/press
curl -fsSL https://raw.githubusercontent.com/ChristianKreuzberger/press/main/.github/skills/press/SKILL.md \
  -o .github/skills/press/SKILL.md
```

Or install it user-wide so it is available across all projects:

```bash
mkdir -p ~/.copilot/skills/press
curl -fsSL https://raw.githubusercontent.com/ChristianKreuzberger/press/main/.github/skills/press/SKILL.md \
  -o ~/.copilot/skills/press/SKILL.md
```

Once in place, GitHub Copilot will discover it automatically. You can also invoke it explicitly in chat by typing `/press`.

---

## Contributing

Contributions are welcome. Please read [`AGENTS.md`](AGENTS.md) before opening a pull request — it explains the project conventions and how AI agents can contribute safely.

```bash
go build ./...   # build
go test ./...    # run tests
go vet ./...     # static analysis
```

## License

[MIT](LICENSE)

