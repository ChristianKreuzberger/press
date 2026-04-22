// Package builder converts press pages to HTML using a template.
package builder

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/ChristianKreuzberger/press/internal/frontmatter"
	"github.com/ChristianKreuzberger/press/internal/markdown"
	"github.com/ChristianKreuzberger/press/internal/page"
	"github.com/ChristianKreuzberger/press/internal/section"
)

// DefaultTemplate is the HTML template used when template.html is not found.
const DefaultTemplate = `<!DOCTYPE html>
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
    </main>
    <footer>
        built with <a href="https://github.com/ChristianKreuzberger/press" style="color:var(--accent2);border:none;">press</a>
    </footer>
</body>
</html>`

// PageRef holds the title and URL used to generate navigation links.
type PageRef struct {
	Title string
	URL   string
}

// TemplateData is passed to the HTML template for each page.
type TemplateData struct {
	Title   string
	Content template.HTML
	Pages   []PageRef
}

// Build converts all pages in siteDir to HTML files in outputDir.
// It reads template.html from siteDir; if absent it falls back to DefaultTemplate.
// Top-level pages (pages/*.md) are written to outputDir directly.
// Section pages (pages/<section>/*.md) are written to outputDir/<section>/.
func Build(siteDir, outputDir string) error {
	pages, err := page.List(siteDir)
	if err != nil {
		return fmt.Errorf("listing pages: %w", err)
	}

	sections, err := section.List(siteDir)
	if err != nil {
		return fmt.Errorf("listing sections: %w", err)
	}

	tmplContent, err := readTemplate(siteDir)
	if err != nil {
		return err
	}

	tmpl, err := template.New("page").Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	// rootNavRefs contains navigation entries with paths relative to the output root
	// (e.g. "about.html", "blog/index.html"). These are used as-is for top-level pages,
	// and prefixed with "../" for pages that live one level deep inside a section
	// directory.
	rootNavRefs := buildRootNavRefs(pages, sections)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Build top-level pages.
	for _, p := range pages {
		if err := buildPageFromPath(p.Name, p.Path, filepath.Join(outputDir, p.Name+".html"), rootNavRefs, tmpl); err != nil {
			return err
		}
	}

	// Build section pages.
	for _, s := range sections {
		sectionPages, err := section.ListPages(siteDir, s.Name)
		if err != nil {
			return fmt.Errorf("listing pages in section %s: %w", s.Name, err)
		}
		sectionOutDir := filepath.Join(outputDir, s.Name)
		if err := os.MkdirAll(sectionOutDir, 0755); err != nil {
			return fmt.Errorf("creating section output directory %s: %w", sectionOutDir, err)
		}
		// Section pages are one level deep, so prefix top-level nav URLs with "../".
		sectionNavRefs := prefixNavRefs(rootNavRefs, "../")
		for _, sp := range sectionPages {
			outPath := filepath.Join(sectionOutDir, sp.Name+".html")
			if err := buildPageFromPath(sp.Name, sp.Path, outPath, sectionNavRefs, tmpl); err != nil {
				return err
			}
		}
	}
	return nil
}

// buildRootNavRefs assembles the navigation entry list using root-relative URLs.
// Top-level pages link to "<name>.html"; sections link to "<section>/index.html".
func buildRootNavRefs(pages []page.Page, sections []section.Section) []PageRef {
	refs := make([]PageRef, 0, len(pages)+len(sections))
	for _, p := range pages {
		refs = append(refs, PageRef{
			Title: resolveTitleFromPath(p.Name, p.Path),
			URL:   p.Name + ".html",
		})
	}
	for _, s := range sections {
		refs = append(refs, PageRef{
			Title: resolveTitleFromPath(s.Name, s.IndexPath),
			URL:   s.Name + "/index.html",
		})
	}
	return refs
}

// prefixNavRefs returns a copy of refs with each URL prefixed by prefix.
func prefixNavRefs(refs []PageRef, prefix string) []PageRef {
	out := make([]PageRef, len(refs))
	for i, r := range refs {
		out[i] = PageRef{Title: r.Title, URL: prefix + r.URL}
	}
	return out
}

func buildPageFromPath(name, mdPath, outPath string, pageRefs []PageRef, tmpl *template.Template) error {
	mdContent, err := os.ReadFile(mdPath)
	if err != nil {
		return fmt.Errorf("reading page %s: %w", name, err)
	}

	mdStr := frontmatter.Strip(string(mdContent))
	htmlContent := markdown.ToHTML(mdStr)
	title := markdown.ExtractTitle(mdStr)
	if title == "" {
		title = name
	}

	data := TemplateData{
		Title:   title,
		Content: template.HTML(htmlContent), //nolint:gosec // markdown is trusted content from the user's own files
		Pages:   pageRefs,
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", outPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing template for page %s: %w", name, err)
	}
	return nil
}

func readTemplate(siteDir string) (string, error) {
	path := filepath.Join(siteDir, "template.html")
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultTemplate, nil
		}
		return "", fmt.Errorf("reading template: %w", err)
	}
	return string(content), nil
}

func resolveTitleFromPath(name, path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return name
	}
	if t := markdown.ExtractTitle(string(content)); t != "" {
		return t
	}
	return name
}
