// Package builder converts press pages to HTML using a template.
package builder

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ChristianKreuzberger/press/internal/frontmatter"
	"github.com/ChristianKreuzberger/press/internal/markdown"
	"github.com/ChristianKreuzberger/press/internal/page"
	"github.com/ChristianKreuzberger/press/internal/section"
	"github.com/ChristianKreuzberger/press/internal/themes"
)

// DefaultTemplate is the HTML template used when template.html is not found.
var DefaultTemplate = themes.Default().Template

// PageRef holds the title and URL used to generate navigation links.
type PageRef struct {
	Title string
	URL   string
}

// TOCEntry represents a single entry in a section's table of contents.
type TOCEntry struct {
	Title     string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
	Weight    int
}

// TemplateData is passed to the HTML template for each page.
type TemplateData struct {
	Title           string
	Content         template.HTML
	Pages           []PageRef
	TableOfContents []TOCEntry
}

// Build converts all pages in siteDir to HTML files in outputDir.
// It reads template.html from siteDir; if absent it falls back to DefaultTemplate.
// Top-level pages (pages/*.md) are written to outputDir directly.
// Section pages (pages/<section>/*.md) are written to outputDir/<section>/.
// When includeDrafts is false, pages with draft: true in their frontmatter are skipped.
// staticDir names a directory relative to siteDir whose contents are copied verbatim
// into outputDir preserving directory structure; if it does not exist it is silently
// skipped.
// It returns the list of absolute paths of HTML files that were written.
func Build(siteDir, outputDir string, includeDrafts bool, staticDir string) ([]string, error) {
	outputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, fmt.Errorf("resolving output dir: %w", err)
	}

	pages, err := page.List(siteDir)
	if err != nil {
		return nil, fmt.Errorf("listing pages: %w", err)
	}

	sections, err := section.List(siteDir)
	if err != nil {
		return nil, fmt.Errorf("listing sections: %w", err)
	}

	tmplContent, err := readTemplate(siteDir)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("page").Parse(tmplContent)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	// rootNavRefs contains navigation entries with paths relative to the output root
	// (e.g. "about.html", "blog/index.html"). These are used as-is for top-level pages,
	// and prefixed with "../" for pages that live one level deep inside a section
	// directory.
	rootNavRefs, err := buildRootNavRefs(pages, sections, includeDrafts)
	if err != nil {
		return nil, fmt.Errorf("building nav refs: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	var built []string

	// Build top-level pages.
	for _, p := range pages {
		if !includeDrafts && p.Draft {
			continue
		}
		outPath := filepath.Join(outputDir, p.Name+".html")
		if err := buildPageFromPath(p.Name, p.Path, outPath, rootNavRefs, nil, tmpl); err != nil {
			return nil, err
		}
		built = append(built, outPath)
	}

	// Copy non-Markdown files from pages/ to outputDir.
	if err := copyStaticAssets(siteDir, outputDir); err != nil {
		return nil, err
	}

	// Copy the static directory verbatim into the output directory.
	if err := copyStaticDir(siteDir, outputDir, staticDir); err != nil {
		return nil, err
	}

	// Build section pages.
	for _, s := range sections {
		// Read the section index up front so we can check its draft status.
		indexContent, err := os.ReadFile(s.IndexPath)
		if err != nil {
			return nil, fmt.Errorf("reading section index %s: %w", s.IndexPath, err)
		}
		// Skip the entire section when its index.md is a draft and drafts are excluded.
		if !includeDrafts && frontmatter.ParseDraft(indexContent) {
			continue
		}
		sectionPages, err := section.ListPages(siteDir, s.Name)
		if err != nil {
			return nil, fmt.Errorf("listing pages in section %s: %w", s.Name, err)
		}
		sectionOutDir := filepath.Join(outputDir, s.Name)
		if err := os.MkdirAll(sectionOutDir, 0755); err != nil {
			return nil, fmt.Errorf("creating section output directory %s: %w", sectionOutDir, err)
		}
		// Build the TOC for this section's index page.
		toc := buildSectionTOC(sectionPages, indexContent, includeDrafts)
		// Section pages are one level deep, so prefix top-level nav URLs with "../".
		sectionNavRefs := prefixNavRefs(rootNavRefs, "../")
		for _, sp := range sectionPages {
			if !includeDrafts && sp.Draft {
				continue
			}
			outPath := filepath.Join(sectionOutDir, sp.Name+".html")
			var pageTOC []TOCEntry
			if sp.Name == "index" {
				pageTOC = toc
			}
			if err := buildPageFromPath(sp.Name, sp.Path, outPath, sectionNavRefs, pageTOC, tmpl); err != nil {
				return nil, err
			}
			built = append(built, outPath)
		}
	}
	return built, nil
}

// weightedRef pairs a PageRef with its navigation weight for sorting.
type weightedRef struct {
	ref    PageRef
	weight int
}

// buildRootNavRefs assembles the navigation entry list using root-relative URLs.
// Top-level pages link to "<name>.html"; sections link to "<section>/index.html".
// Entries are sorted by ascending weight; entries with weight=0 (unset) appear last
// in their original filesystem order (stable sort).
// When includeDrafts is false, draft pages and draft sections are excluded from navigation.
func buildRootNavRefs(pages []page.Page, sections []section.Section, includeDrafts bool) ([]PageRef, error) {
	weighted := make([]weightedRef, 0, len(pages)+len(sections))
	for _, p := range pages {
		if !includeDrafts && p.Draft {
			continue
		}
		content, err := os.ReadFile(p.Path)
		if err != nil {
			return nil, fmt.Errorf("reading page %s: %w", p.Path, err)
		}
		weighted = append(weighted, weightedRef{
			ref: PageRef{
				Title: resolveTitleFromContent(p.Name, content),
				URL:   p.Name + ".html",
			},
			weight: frontmatter.ParseWeight(content),
		})
	}
	for _, s := range sections {
		content, err := os.ReadFile(s.IndexPath)
		if err != nil {
			return nil, fmt.Errorf("reading section index %s: %w", s.IndexPath, err)
		}
		if !includeDrafts && frontmatter.ParseDraft(content) {
			continue
		}
		weighted = append(weighted, weightedRef{
			ref: PageRef{
				Title: resolveTitleFromContent(s.Name, content),
				URL:   s.Name + "/index.html",
			},
			weight: frontmatter.ParseWeight(content),
		})
	}
	sort.SliceStable(weighted, func(i, j int) bool {
		return weightLess(weighted[i].weight, weighted[j].weight)
	})
	refs := make([]PageRef, len(weighted))
	for i, w := range weighted {
		refs[i] = w.ref
	}
	return refs, nil
}

// prefixNavRefs returns a copy of refs with each URL prefixed by prefix.
func prefixNavRefs(refs []PageRef, prefix string) []PageRef {
	out := make([]PageRef, len(refs))
	for i, r := range refs {
		out[i] = PageRef{Title: r.Title, URL: prefix + r.URL}
	}
	return out
}

// buildSectionTOC collects TOC entries for all non-index pages in the section,
// sorted according to the toc_sort and toc_order fields in indexContent.
// When includeDrafts is false, draft pages are excluded from the TOC.
func buildSectionTOC(pages []section.Page, indexContent []byte, includeDrafts bool) []TOCEntry {
	tocSort := frontmatter.ParseStringField(indexContent, "toc_sort")
	tocOrder := frontmatter.ParseStringField(indexContent, "toc_order")
	if tocSort == "" {
		tocSort = "weight"
	}
	if tocOrder == "" {
		tocOrder = "asc"
	}

	var entries []TOCEntry
	for _, p := range pages {
		if p.Name == "index" {
			continue
		}
		if !includeDrafts && p.Draft {
			continue
		}
		content, err := os.ReadFile(p.Path)
		if err != nil {
			entries = append(entries, TOCEntry{
				Title: p.Name,
				URL:   p.Name + ".html",
			})
			continue
		}
		body := frontmatter.Strip(string(content))
		title := markdown.ExtractTitle(body)
		if title == "" {
			title = p.Name
		}
		entries = append(entries, TOCEntry{
			Title:     title,
			URL:       p.Name + ".html",
			CreatedAt: frontmatter.ParseTimeField(content, "created_at"),
			UpdatedAt: frontmatter.ParseTimeField(content, "updated_at"),
			Weight:    frontmatter.ParseWeight(content),
		})
	}

	sortTOC(entries, tocSort, tocOrder)
	return entries
}

// sortTOC sorts entries in-place by the given field and order.
// For weight sort: entries with weight=0 (unset) always appear last regardless of order.
func sortTOC(entries []TOCEntry, by, order string) {
	sort.SliceStable(entries, func(i, j int) bool {
		switch by {
		case "title":
			ai, aj := strings.ToLower(entries[i].Title), strings.ToLower(entries[j].Title)
			if order == "desc" {
				return aj < ai
			}
			return ai < aj
		case "created_at":
			if order == "desc" {
				return entries[j].CreatedAt.Before(entries[i].CreatedAt)
			}
			return entries[i].CreatedAt.Before(entries[j].CreatedAt)
		case "updated_at":
			if order == "desc" {
				return entries[j].UpdatedAt.Before(entries[i].UpdatedAt)
			}
			return entries[i].UpdatedAt.Before(entries[j].UpdatedAt)
		default: // "weight"
			wi, wj := entries[i].Weight, entries[j].Weight
			if order == "desc" {
				// weight=0 (unset) always sorts last regardless of direction.
				// wi==0 covers both the "both zero" and "only wi zero" cases.
				if wi == 0 {
					return false
				}
				if wj == 0 {
					return true
				}
				return wj < wi
			}
			return weightLess(wi, wj)
		}
	})
}

func buildPageFromPath(name, mdPath, outPath string, pageRefs []PageRef, toc []TOCEntry, tmpl *template.Template) error {
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
		Title:           title,
		Content:         template.HTML(htmlContent), //nolint:gosec // markdown is trusted content from the user's own files
		Pages:           pageRefs,
		TableOfContents: toc,
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

// copyStaticAssets copies all non-Markdown files from the pages/ directory
// to the corresponding location in outputDir, preserving the directory structure.
func copyStaticAssets(siteDir, outputDir string) error {
	pagesDir := page.PagesDir(siteDir)
	if _, err := os.Stat(pagesDir); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(pagesDir, func(src string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		rel, err := filepath.Rel(pagesDir, src)
		if err != nil {
			return err
		}
		dst := filepath.Join(outputDir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return fmt.Errorf("creating directory for asset %s: %w", rel, err)
		}
		return copyFile(src, dst)
	})
}

// copyStaticDir copies all files from the directory named staticDirName inside
// siteDir into a same-named subdirectory of outputDir, preserving the directory
// structure.  If the source directory does not exist the function returns nil
// silently.
func copyStaticDir(siteDir, outputDir, staticDirName string) error {
	srcDir := filepath.Join(siteDir, staticDirName)
	if _, err := os.Stat(srcDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("checking static directory %s: %w", srcDir, err)
	}
	dstDir := filepath.Join(outputDir, staticDirName)
	return filepath.WalkDir(srcDir, func(src string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, src)
		if err != nil {
			return err
		}
		dst := filepath.Join(dstDir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return fmt.Errorf("creating directory for static file %s: %w", rel, err)
		}
		return copyFile(src, dst)
	})
}

// copyFile copies the file at src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening asset %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating asset %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copying asset %s: %w", dst, err)
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

// resolveTitleFromContent extracts the first Markdown heading from content as
// the page title, falling back to name when no heading is found.
func resolveTitleFromContent(name string, content []byte) string {
	body := frontmatter.Strip(string(content))
	if t := markdown.ExtractTitle(body); t != "" {
		return t
	}
	return name
}

// weightLess reports whether wi sorts before wj in ascending weight order.
// Items with weight 0 (unset) always sort last.
// Do not use for descending order by swapping arguments — the zero-last
// invariant breaks. Handle descending separately.
func weightLess(wi, wj int) bool {
	if wi == 0 && wj == 0 {
		return false
	}
	if wi == 0 {
		return false
	}
	if wj == 0 {
		return true
	}
	return wi < wj
}
