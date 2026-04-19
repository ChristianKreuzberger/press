// Package builder converts press pages to HTML using a template.
package builder

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/ChristianKreuzberger/press/internal/markdown"
	"github.com/ChristianKreuzberger/press/internal/page"
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
func Build(siteDir, outputDir string) error {
	pages, err := page.List(siteDir)
	if err != nil {
		return fmt.Errorf("listing pages: %w", err)
	}

	tmplContent, err := readTemplate(siteDir)
	if err != nil {
		return err
	}

	tmpl, err := template.New("page").Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	// Build navigation refs for all pages (title resolved from markdown).
	pageRefs := make([]PageRef, 0, len(pages))
	for _, p := range pages {
		pageRefs = append(pageRefs, PageRef{
			Title: resolveTitle(p),
			URL:   p.Name + ".html",
		})
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	for _, p := range pages {
		if err := buildPage(p, pageRefs, tmpl, outputDir); err != nil {
			return err
		}
	}
	return nil
}

func buildPage(p page.Page, pageRefs []PageRef, tmpl *template.Template, outputDir string) error {
	mdContent, err := os.ReadFile(p.Path)
	if err != nil {
		return fmt.Errorf("reading page %s: %w", p.Name, err)
	}

	htmlContent := markdown.ToHTML(string(mdContent))
	title := markdown.ExtractTitle(string(mdContent))
	if title == "" {
		title = p.Name
	}

	data := TemplateData{
		Title:   title,
		Content: template.HTML(htmlContent), //nolint:gosec // markdown is trusted content from the user's own files
		Pages:   pageRefs,
	}

	outPath := filepath.Join(outputDir, p.Name+".html")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", outPath, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing template for page %s: %w", p.Name, err)
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

func resolveTitle(p page.Page) string {
	content, err := os.ReadFile(p.Path)
	if err != nil {
		return p.Name
	}
	if t := markdown.ExtractTitle(string(content)); t != "" {
		return t
	}
	return p.Name
}
