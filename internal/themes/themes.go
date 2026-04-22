// Package themes provides built-in HTML themes for press sites.
// A theme is a self-contained HTML template that uses Go's html/template
// syntax. The template receives a TemplateData value with these fields:
//
//   - .Title          string       — page title
//   - .Content        template.HTML — rendered HTML content
//   - .Pages          []PageRef    — navigation links (each has .Title and .URL)
//   - .TableOfContents []TOCEntry  — section table of contents (each has .Title and .URL)
package themes

// Theme is a built-in HTML template for a press site.
type Theme struct {
	Name        string
	Description string
	Template    string
}

// All is the ordered list of built-in themes. The first entry is the default.
var All = []Theme{
	{
		Name:        "dark",
		Description: "Dark developer theme inspired by GitHub dark mode",
		Template:    darkTemplate,
	},
	{
		Name:        "light",
		Description: "Clean editorial theme with serif headings and a light background",
		Template:    lightTemplate,
	},
	{
		Name:        "terminal",
		Description: "Retro green-on-black terminal aesthetic with monospace fonts",
		Template:    terminalTemplate,
	},
}

// Default returns the default theme (the first entry in All).
func Default() Theme {
	return All[0]
}

// Get returns the theme with the given name and whether it was found.
func Get(name string) (Theme, bool) {
	for _, t := range All {
		if t.Name == name {
			return t, true
		}
	}
	return Theme{}, false
}

// Names returns the names of all built-in themes in order.
func Names() []string {
	names := make([]string, len(All))
	for i, t := range All {
		names[i] = t.Name
	}
	return names
}

// ── dark ──────────────────────────────────────────────────────────────────────

const darkTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

        :root {
            --bg:        #0d1117;
            --surface:   #161b22;
            --border:    #30363d;
            --accent:    #58a6ff;
            --accent2:   #3fb950;
            --text:      #c9d1d9;
            --muted:     #8b949e;
            --heading:   #f0f6fc;
            --font-mono: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
            --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
        }

        html { font-size: 16px; scroll-behavior: smooth; }

        body {
            background: var(--bg);
            color: var(--text);
            font-family: var(--font-sans);
            line-height: 1.7;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
        }

        /* ── Header / Nav ─────────────────────────────────────────── */
        header {
            background: var(--surface);
            border-bottom: 1px solid var(--border);
            position: sticky;
            top: 0;
            z-index: 100;
        }

        .header-inner {
            max-width: 900px;
            margin: 0 auto;
            padding: 0 1.5rem;
            height: 56px;
            display: flex;
            align-items: center;
            gap: 2rem;
        }

        .logo {
            font-family: var(--font-mono);
            font-size: 0.95rem;
            color: var(--accent2);
            text-decoration: none;
            white-space: nowrap;
        }

        .logo::before { content: "> "; color: var(--muted); }

        nav {
            display: flex;
            gap: 0.25rem;
            flex-wrap: wrap;
        }

        nav a {
            color: var(--muted);
            text-decoration: none;
            font-size: 0.875rem;
            padding: 0.3rem 0.65rem;
            border-radius: 6px;
            transition: color 0.15s, background 0.15s;
        }

        nav a:hover {
            color: var(--heading);
            background: rgba(88, 166, 255, 0.1);
        }

        /* ── Main content ─────────────────────────────────────────── */
        main {
            max-width: 900px;
            width: 100%;
            margin: 3rem auto;
            padding: 0 1.5rem;
            flex: 1;
        }

        /* ── Typography ───────────────────────────────────────────── */
        main h1 {
            font-size: 2rem;
            color: var(--heading);
            font-weight: 600;
            margin-bottom: 0.5rem;
            padding-bottom: 0.5rem;
            border-bottom: 1px solid var(--border);
        }

        main h2 {
            font-size: 1.35rem;
            color: var(--heading);
            font-weight: 600;
            margin: 2rem 0 0.75rem;
        }

        main h3 {
            font-size: 1.1rem;
            color: var(--accent);
            font-weight: 600;
            margin: 1.5rem 0 0.5rem;
        }

        main p { margin-bottom: 1rem; }

        main a {
            color: var(--accent);
            text-decoration: none;
            border-bottom: 1px solid transparent;
            transition: border-color 0.15s;
        }

        main a:hover { border-bottom-color: var(--accent); }

        main ul, main ol {
            padding-left: 1.5rem;
            margin-bottom: 1rem;
        }

        main li { margin-bottom: 0.3rem; }

        main code {
            font-family: var(--font-mono);
            font-size: 0.85em;
            background: var(--surface);
            border: 1px solid var(--border);
            border-radius: 4px;
            padding: 0.1em 0.4em;
            color: var(--accent2);
        }

        main pre {
            background: var(--surface);
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 1.25rem 1.5rem;
            overflow-x: auto;
            margin-bottom: 1.25rem;
            position: relative;
        }

        main pre::before {
            content: "$ ";
            color: var(--accent2);
            font-family: var(--font-mono);
        }

        main pre code {
            background: none;
            border: none;
            padding: 0;
            font-size: 0.875rem;
            color: var(--text);
        }

        main blockquote {
            border-left: 3px solid var(--accent);
            padding: 0.75rem 1.25rem;
            margin: 1.25rem 0;
            color: var(--muted);
            background: var(--surface);
            border-radius: 0 6px 6px 0;
        }

        main hr {
            border: none;
            border-top: 1px solid var(--border);
            margin: 2rem 0;
        }

        main table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 1.25rem;
            font-size: 0.9rem;
        }

        main th, main td {
            border: 1px solid var(--border);
            padding: 0.6rem 1rem;
            text-align: left;
        }

        main th {
            background: var(--surface);
            color: var(--heading);
            font-weight: 600;
        }

        main tr:nth-child(even) { background: rgba(22, 27, 34, 0.5); }

        /* ── Footer ───────────────────────────────────────────────── */
        footer {
            border-top: 1px solid var(--border);
            text-align: center;
            padding: 1.25rem;
            font-size: 0.8rem;
            color: var(--muted);
            font-family: var(--font-mono);
        }

        /* ── Responsive ───────────────────────────────────────────── */
        @media (max-width: 600px) {
            .header-inner { gap: 1rem; height: auto; padding: 0.75rem 1rem; flex-wrap: wrap; }
            main { margin: 1.5rem auto; }
            main h1 { font-size: 1.5rem; }
        }
    </style>
</head>
<body>
    <header>
        <div class="header-inner">
            <a class="logo" href="/">portfolio</a>
            <nav>{{range .Pages}}<a href="{{.URL}}">{{.Title}}</a>{{end}}</nav>
        </div>
    </header>
    <main>
        {{.Content}}
        {{if .TableOfContents}}
        <section class="toc">
            <h2>Contents</h2>
            <ul>
                {{range .TableOfContents}}<li><a href="{{.URL}}">{{.Title}}</a></li>
                {{end}}
            </ul>
        </section>
        {{end}}
    </main>
    <footer>
        built with <a href="https://github.com/ChristianKreuzberger/press" style="color:var(--accent2);border:none;">press</a>
    </footer>
</body>
</html>`

// ── light ─────────────────────────────────────────────────────────────────────

const lightTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

        :root {
            --bg:         #fafaf8;
            --surface:    #ffffff;
            --border:     #e8e8e4;
            --accent:     #2563eb;
            --text:       #1c1c1e;
            --muted:      #6b7280;
            --heading:    #111111;
            --code-bg:    #f0ede8;
            --code-color: #b91c1c;
            --font-serif: Georgia, "Times New Roman", serif;
            --font-sans:  -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
            --font-mono:  "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
        }

        html { font-size: 16px; scroll-behavior: smooth; }

        body {
            background: var(--bg);
            color: var(--text);
            font-family: var(--font-sans);
            line-height: 1.75;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
        }

        /* ── Header / Nav ─────────────────────────────────────────── */
        header {
            background: var(--surface);
            border-bottom: 1px solid var(--border);
        }

        .header-inner {
            max-width: 860px;
            margin: 0 auto;
            padding: 0 2rem;
            height: 60px;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        .logo {
            font-family: var(--font-serif);
            font-size: 1.1rem;
            font-weight: 700;
            color: var(--heading);
            text-decoration: none;
            letter-spacing: -0.02em;
        }

        nav { display: flex; gap: 0.25rem; }

        nav a {
            color: var(--muted);
            text-decoration: none;
            font-size: 0.875rem;
            padding: 0.35rem 0.75rem;
            border-radius: 4px;
            transition: color 0.15s, background 0.15s;
        }

        nav a:hover {
            color: var(--heading);
            background: var(--border);
        }

        /* ── Main content ─────────────────────────────────────────── */
        main {
            max-width: 720px;
            width: 100%;
            margin: 4rem auto;
            padding: 0 2rem;
            flex: 1;
        }

        /* ── Typography ───────────────────────────────────────────── */
        main h1 {
            font-family: var(--font-serif);
            font-size: 2.25rem;
            color: var(--heading);
            font-weight: 700;
            letter-spacing: -0.02em;
            margin-bottom: 0.75rem;
        }

        main h2 {
            font-family: var(--font-serif);
            font-size: 1.5rem;
            color: var(--heading);
            font-weight: 600;
            letter-spacing: -0.01em;
            margin: 2.5rem 0 0.75rem;
        }

        main h3 {
            font-size: 1.125rem;
            color: var(--heading);
            font-weight: 600;
            margin: 2rem 0 0.5rem;
        }

        main p { margin-bottom: 1.25rem; }

        main a {
            color: var(--accent);
            text-decoration: underline;
            text-decoration-color: transparent;
            text-underline-offset: 3px;
            transition: text-decoration-color 0.15s;
        }

        main a:hover { text-decoration-color: var(--accent); }

        main ul, main ol {
            padding-left: 1.75rem;
            margin-bottom: 1.25rem;
        }

        main li { margin-bottom: 0.4rem; }

        main code {
            font-family: var(--font-mono);
            font-size: 0.85em;
            background: var(--code-bg);
            border-radius: 3px;
            padding: 0.15em 0.4em;
            color: var(--code-color);
        }

        main pre {
            background: #f5f2ef;
            border: 1px solid var(--border);
            border-radius: 6px;
            padding: 1.5rem;
            overflow-x: auto;
            margin-bottom: 1.5rem;
        }

        main pre code {
            background: none;
            padding: 0;
            font-size: 0.875rem;
            color: var(--text);
        }

        main blockquote {
            border-left: 3px solid var(--accent);
            padding: 0.75rem 1.5rem;
            margin: 1.5rem 0;
            color: var(--muted);
            font-style: italic;
        }

        main hr {
            border: none;
            border-top: 1px solid var(--border);
            margin: 2.5rem 0;
        }

        main table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 1.5rem;
            font-size: 0.9rem;
        }

        main th, main td {
            border: 1px solid var(--border);
            padding: 0.6rem 1rem;
            text-align: left;
        }

        main th {
            background: var(--border);
            color: var(--heading);
            font-weight: 600;
        }

        /* ── Table of Contents ────────────────────────────────────── */
        .toc { margin-top: 2rem; }

        .toc h2 {
            font-family: var(--font-sans);
            font-size: 0.75rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.1em;
            color: var(--muted);
            margin-bottom: 0.75rem;
        }

        .toc ul { list-style: none; padding: 0; }

        .toc li {
            padding: 0.35rem 0;
            border-bottom: 1px solid var(--border);
        }

        .toc a {
            color: var(--text);
            text-decoration: none;
            text-decoration-color: transparent;
        }

        .toc a:hover { color: var(--accent); }

        /* ── Footer ───────────────────────────────────────────────── */
        footer {
            border-top: 1px solid var(--border);
            text-align: center;
            padding: 1.5rem;
            font-size: 0.8rem;
            color: var(--muted);
        }

        footer a { color: var(--accent); }

        /* ── Responsive ───────────────────────────────────────────── */
        @media (max-width: 600px) {
            .header-inner { padding: 0 1rem; }
            main { margin: 2rem auto; padding: 0 1rem; }
            main h1 { font-size: 1.75rem; }
        }
    </style>
</head>
<body>
    <header>
        <div class="header-inner">
            <a class="logo" href="/">My Site</a>
            <nav>{{range .Pages}}<a href="{{.URL}}">{{.Title}}</a>{{end}}</nav>
        </div>
    </header>
    <main>
        {{.Content}}
        {{if .TableOfContents}}
        <section class="toc">
            <h2>Contents</h2>
            <ul>
                {{range .TableOfContents}}<li><a href="{{.URL}}">{{.Title}}</a></li>
                {{end}}
            </ul>
        </section>
        {{end}}
    </main>
    <footer>
        built with <a href="https://github.com/ChristianKreuzberger/press">press</a>
    </footer>
</body>
</html>`

// ── terminal ──────────────────────────────────────────────────────────────────

const terminalTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

        :root {
            --bg:      #0d0d0d;
            --surface: #141414;
            --border:  #2a2a2a;
            --green:   #39d353;
            --amber:   #f0c040;
            --text:    #c8c8c8;
            --dim:     #555555;
            --font:    "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
        }

        html { font-size: 15px; scroll-behavior: smooth; }

        body {
            background: var(--bg);
            color: var(--text);
            font-family: var(--font);
            line-height: 1.6;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
        }

        /* ── Header / Nav ─────────────────────────────────────────── */
        header {
            border-bottom: 1px solid var(--border);
            padding: 1rem 0;
        }

        .header-inner {
            max-width: 820px;
            margin: 0 auto;
            padding: 0 1.5rem;
            display: flex;
            align-items: center;
            gap: 2rem;
            flex-wrap: wrap;
        }

        .logo {
            color: var(--green);
            text-decoration: none;
            font-weight: bold;
        }

        .logo::before { content: "$ "; color: var(--dim); }

        nav { display: flex; gap: 0; flex-wrap: wrap; }

        nav a {
            color: var(--dim);
            text-decoration: none;
            padding: 0.2rem 0.5rem;
            transition: color 0.1s;
        }

        nav a::before { content: "["; }
        nav a::after  { content: "]"; }

        nav a:hover { color: var(--amber); }

        /* ── Main content ─────────────────────────────────────────── */
        main {
            max-width: 820px;
            width: 100%;
            margin: 2.5rem auto;
            padding: 0 1.5rem;
            flex: 1;
        }

        /* ── Typography ───────────────────────────────────────────── */
        main h1 {
            font-size: 1.4rem;
            color: var(--amber);
            font-weight: bold;
            margin-bottom: 0.5rem;
        }

        main h1::before { content: "# "; color: var(--dim); }

        main h2 {
            font-size: 1.15rem;
            color: var(--green);
            margin: 2rem 0 0.5rem;
        }

        main h2::before { content: "## "; color: var(--dim); }

        main h3 {
            font-size: 1rem;
            color: var(--text);
            margin: 1.5rem 0 0.4rem;
        }

        main h3::before { content: "### "; color: var(--dim); }

        main p { margin-bottom: 1rem; }

        main a {
            color: var(--amber);
            text-decoration: none;
        }

        main a:hover { text-decoration: underline; }

        main ul, main ol {
            padding-left: 2rem;
            margin-bottom: 1rem;
        }

        main ul li::marker { color: var(--green); content: "- "; }

        main li { margin-bottom: 0.25rem; }

        main code {
            color: var(--green);
            background: var(--surface);
            border: 1px solid var(--border);
            padding: 0.1em 0.35em;
        }

        main pre {
            background: var(--surface);
            border: 1px solid var(--border);
            padding: 1rem 1.25rem;
            overflow-x: auto;
            margin-bottom: 1rem;
        }

        main pre::before {
            display: block;
            content: "────────────────────────────────";
            color: var(--border);
            margin-bottom: 0.5rem;
        }

        main pre code {
            background: none;
            border: none;
            padding: 0;
            color: var(--text);
        }

        main blockquote {
            border-left: 2px solid var(--green);
            padding: 0.5rem 1rem;
            margin: 1rem 0;
            color: var(--dim);
        }

        main hr {
            border: none;
            border-top: 1px solid var(--border);
            margin: 1.5rem 0;
        }

        main table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 1rem;
        }

        main th, main td {
            border: 1px solid var(--border);
            padding: 0.4rem 0.75rem;
        }

        main th { color: var(--amber); }

        /* ── Table of Contents ────────────────────────────────────── */
        .toc { margin-top: 1.5rem; }

        .toc h2 { color: var(--green); }
        .toc h2::before { content: ""; }

        .toc ul { list-style: none; padding: 0; }

        .toc li {
            padding: 0.2rem 0;
        }

        .toc li::before { content: "> "; color: var(--dim); }

        .toc a { color: var(--text); text-decoration: none; }
        .toc a:hover { color: var(--amber); }

        /* ── Footer ───────────────────────────────────────────────── */
        footer {
            border-top: 1px solid var(--border);
            padding: 1rem 1.5rem;
            font-size: 0.8rem;
            color: var(--dim);
            text-align: center;
        }

        footer a { color: var(--green); }

        /* ── Responsive ───────────────────────────────────────────── */
        @media (max-width: 600px) {
            main { margin: 1.5rem auto; }
        }
    </style>
</head>
<body>
    <header>
        <div class="header-inner">
            <a class="logo" href="/">site</a>
            <nav>{{range .Pages}}<a href="{{.URL}}">{{.Title}}</a>{{end}}</nav>
        </div>
    </header>
    <main>
        {{.Content}}
        {{if .TableOfContents}}
        <section class="toc">
            <h2>Index</h2>
            <ul>
                {{range .TableOfContents}}<li><a href="{{.URL}}">{{.Title}}</a></li>
                {{end}}
            </ul>
        </section>
        {{end}}
    </main>
    <footer>
        [built with <a href="https://github.com/ChristianKreuzberger/press">press</a>]
    </footer>
</body>
</html>`
