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

// Static errors for builder operations.
var (
	errEmptyStaticDirName   = fmt.Errorf("static directory name must not be empty")
	errStaticDirNotRelative = fmt.Errorf("static directory name must be relative to the site directory")
	errStaticDirNotDir      = fmt.Errorf("static directory is not a directory")
)

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
// staticDir names a directory relative to siteDir whose files are copied into
// outputDir while preserving directory structure; if it does not exist it is
// silently skipped.
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

	rootNavRefs, err := buildRootNavRefs(pages, sections, includeDrafts)
	if err != nil {
		return nil, fmt.Errorf("building nav refs: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil { //nolint:gosec // generated site output must be world-readable for web servers
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	var built []string

	topLevelPaths, err := buildTopLevelPages(pages, outputDir, rootNavRefs, tmpl, includeDrafts)
	if err != nil {
		return nil, err
	}
	built = append(built, topLevelPaths...)

	if err := copyAssets(siteDir, outputDir, staticDir); err != nil {
		return nil, err
	}

	sectionPaths, err := buildSections(sections, siteDir, outputDir, rootNavRefs, tmpl, includeDrafts)
	if err != nil {
		return nil, err
	}
	built = append(built, sectionPaths...)

	return built, nil
}

// buildTopLevelPages builds all non-draft top-level pages.
// It returns the list of absolute paths of HTML files that were written.
func buildTopLevelPages(pages []page.Page, outputDir string, rootNavRefs []PageRef, tmpl *template.Template, includeDrafts bool) ([]string, error) {
	var built []string
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
	return built, nil
}

// copyAssets copies both static assets from pages/ and the static directory.
// Static assets are non-Markdown files under pages/.
// The static directory is copied verbatim if it exists.
func copyAssets(siteDir, outputDir, staticDir string) error {
	if err := copyStaticAssets(siteDir, outputDir); err != nil {
		return err
	}
	if err := copyStaticDir(siteDir, outputDir, staticDir); err != nil {
		return err
	}
	return nil
}

// buildSections builds all section pages.
// It returns the list of absolute paths of HTML files that were written.
func buildSections(sections []section.Section, siteDir, outputDir string, rootNavRefs []PageRef, tmpl *template.Template, includeDrafts bool) ([]string, error) {
	var built []string
	sectionNavRefs := prefixNavRefs(rootNavRefs, "../")

	for _, s := range sections {
		sectionPaths, err := buildSection(s, siteDir, outputDir, sectionNavRefs, tmpl, includeDrafts)
		if err != nil {
			return nil, err
		}
		built = append(built, sectionPaths...)
	}
	return built, nil
}

// buildSection builds a single section.
// It returns the list of absolute paths of HTML files that were written.
func buildSection(s section.Section, siteDir, outputDir string, sectionNavRefs []PageRef, tmpl *template.Template, includeDrafts bool) ([]string, error) {
	indexContent, err := os.ReadFile(s.IndexPath)
	if err != nil {
		return nil, fmt.Errorf("reading section index %s: %w", s.IndexPath, err)
	}
	if !includeDrafts && frontmatter.ParseDraft(indexContent) {
		return nil, nil
	}

	sectionPages, err := section.ListPages(siteDir, s.Name)
	if err != nil {
		return nil, fmt.Errorf("listing pages in section %s: %w", s.Name, err)
	}

	sectionOutDir := filepath.Join(outputDir, s.Name)
	if err := os.MkdirAll(sectionOutDir, 0755); err != nil { //nolint:gosec // generated site output must be world-readable for web servers
		return nil, fmt.Errorf("creating section output directory %s: %w", sectionOutDir, err)
	}

	toc := buildSectionTOC(sectionPages, indexContent, includeDrafts)

	var built []string
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

func buildPageFromPath(name, mdPath, outPath string, pageRefs []PageRef, toc []TOCEntry, tmpl *template.Template) (err error) {
	mdContent, err := os.ReadFile(mdPath) //nolint:gosec // mdPath comes from walking the site pages dir
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

	f, err := os.Create(outPath) //nolint:gosec // outPath is inside the build output dir
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", outPath, err)
	}
	// A failed close on a written file can mean lost data, so surface it.
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("closing output file %s: %w", outPath, cerr)
		}
	}()

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
			// Hidden directories (.git, .cache, ...) are never published.
			if strings.HasPrefix(d.Name(), ".") && src != pagesDir {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip symlinks and other non-regular files so a link cannot pull
		// files from outside the site into the output.
		if !d.Type().IsRegular() {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		// Markdown in any case is source, not an asset.
		if strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			return nil
		}
		rel, err := filepath.Rel(pagesDir, src)
		if err != nil {
			return err
		}
		if err := checkPageCollision(src, rel); err != nil {
			return err
		}
		dst := filepath.Join(outputDir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil { //nolint:gosec // generated site output must be world-readable for web servers
			return fmt.Errorf("creating directory for asset %s: %w", rel, err)
		}
		return copyFile(src, dst)
	})
}

// checkPageCollision fails when an asset would be written to the same output
// path as a built page, i.e. about.html next to about.md at the top level or
// directly inside a section. Deeper directories are not built, so they can't collide.
func checkPageCollision(src, rel string) error {
	if !strings.HasSuffix(src, ".html") || strings.Contains(filepath.Dir(rel), string(filepath.Separator)) {
		return nil
	}
	mdPath := strings.TrimSuffix(src, ".html") + ".md"
	if _, err := os.Lstat(mdPath); err == nil {
		return fmt.Errorf("asset %s would overwrite the page built from %s", rel, filepath.Base(mdPath))
	}
	return nil
}

// validateStaticDirName checks that staticDirName is safe to use as a path
// component relative to siteDir. It rejects empty values, absolute paths, and
// names that escape the site directory via "..".
func validateStaticDirName(staticDirName string) (string, error) {
	if staticDirName == "" {
		return "", errEmptyStaticDirName
	}
	if filepath.IsAbs(staticDirName) {
		return "", fmt.Errorf("%w: %s", errStaticDirNotRelative, staticDirName)
	}
	clean := filepath.Clean(staticDirName)
	if clean == "." {
		return "", errEmptyStaticDirName
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %s", errStaticDirNotRelative, staticDirName)
	}
	return clean, nil
}

// copyStaticDir copies all files from the directory named staticDirName inside
// siteDir into a same-named subdirectory of outputDir, preserving the directory
// structure. Symlinks are skipped. If the source directory does not exist the
// function returns nil silently.
func copyStaticDir(siteDir, outputDir, staticDirName string) error {
	cleanName, err := validateStaticDirName(staticDirName)
	if err != nil {
		return err
	}
	srcDir := filepath.Join(siteDir, cleanName)
	info, err := os.Stat(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("checking static directory %s: %w", srcDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: %s", errStaticDirNotDir, srcDir)
	}
	dstDir := filepath.Join(outputDir, cleanName)
	return filepath.WalkDir(srcDir, func(src string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		rel, err := filepath.Rel(srcDir, src)
		if err != nil {
			return err
		}
		dst := filepath.Join(dstDir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil { //nolint:gosec // generated site output must be world-readable for web servers
			return fmt.Errorf("creating directory for static file %s: %w", rel, err)
		}
		return copyFile(src, dst)
	})
}

// copyFile copies the file at src to dst.
func copyFile(src, dst string) (err error) {
	in, err := os.Open(src) //nolint:gosec // src comes from walking the site directory
	if err != nil {
		return fmt.Errorf("opening asset %s: %w", src, err)
	}
	defer func() { _ = in.Close() }() // read-only file, close error is not actionable

	out, err := os.Create(dst) //nolint:gosec // dst is inside the build output dir
	if err != nil {
		return fmt.Errorf("creating asset %s: %w", dst, err)
	}
	// A failed close on a written file can mean lost data, so surface it.
	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("closing asset %s: %w", dst, cerr)
		}
	}()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copying asset %s: %w", dst, err)
	}
	return nil
}

func readTemplate(siteDir string) (string, error) {
	path := filepath.Join(siteDir, "template.html")
	content, err := os.ReadFile(path) //nolint:gosec // fixed filename inside the user-supplied site dir
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
