// Package builder converts press pages to HTML using a template.
package builder

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

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
</head>
<body>
    <nav>{{range .Pages}}<a href="{{.URL}}">{{.Title}}</a> {{end}}</nav>
    <main>
        {{.Content}}
    </main>
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

	// rootNavRefs contains navigation entries with root-relative URLs (e.g. "about.html",
	// "blog/index.html"). These are used as-is for top-level pages, and prefixed with
	// "../" for pages that live one level deep inside a section directory.
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

	htmlContent := markdown.ToHTML(string(mdContent))
	title := markdown.ExtractTitle(string(mdContent))
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
