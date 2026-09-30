package main

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ChristianKreuzberger/press/internal/frontmatter"
	"github.com/ChristianKreuzberger/press/internal/page"
)

// linkStartRe matches the start of a Markdown inline link, up to and including
// the "(" before the destination. Group 1 is "!" for image links.
var linkStartRe = regexp.MustCompile(`(!?)\[[^\]]*\]\(`)

// inlineCodeRe matches inline code spans so links inside them can be ignored.
var inlineCodeRe = regexp.MustCompile("`[^`\n]*`")

// defaultStaticDir is the static directory `press build` uses by default;
// `press check` takes no flags, so it assumes this one.
const defaultStaticDir = "static"

func runCheck(args []string) {
	parseOrExit(newFlagSet("check"), args, 0, 0, "press check")
	siteDir := mustGetwd()
	pagesDir := page.PagesDir(siteDir)

	var issues []string
	pageCount := 0

	// Build the set of valid internal link paths.
	validPaths := buildValidPaths(siteDir)

	// Check top-level pages.
	topPages, err := page.List(siteDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing pages: %v\n", err)
		os.Exit(1)
	}
	for _, p := range topPages {
		pageCount++
		relPath := p.Name + ".md"
		content, err := os.ReadFile(p.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", relPath, err)
			os.Exit(1)
		}
		issues = append(issues, checkPage(relPath, content, validPaths)...)
	}

	// Scan pages directory for subdirectories.
	entries, err := os.ReadDir(pagesDir)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error reading pages directory: %v\n", err)
		os.Exit(1)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sectionName := e.Name()
		sectionPath := filepath.Join(pagesDir, sectionName)

		// Only treat a subdirectory as a section if it contains at least one
		// Markdown file. Directories with only static assets (e.g. pages/assets/)
		// are intentionally skipped.
		sectionFiles, err := os.ReadDir(sectionPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading section %s: %v\n", sectionName, err)
			os.Exit(1)
		}
		hasMd := false
		for _, sf := range sectionFiles {
			if !sf.IsDir() && strings.HasSuffix(sf.Name(), ".md") {
				hasMd = true
				break
			}
		}
		if !hasMd {
			continue
		}

		indexPath := filepath.Join(sectionPath, "index.md")

		if _, statErr := os.Stat(indexPath); statErr != nil {
			if !os.IsNotExist(statErr) {
				fmt.Fprintf(os.Stderr, "error checking %s: %v\n", indexPath, statErr)
				os.Exit(1)
			}
			// Section directory without index.md.
			issues = append(issues, fmt.Sprintf("%s/: section has no index.md", sectionName))
			continue
		}

		// Check all .md files in this section.
		for _, sf := range sectionFiles {
			if sf.IsDir() || !strings.HasSuffix(sf.Name(), ".md") {
				continue
			}
			pageCount++
			relPath := sectionName + "/" + sf.Name()
			fullPath := filepath.Join(sectionPath, sf.Name())
			content, err := os.ReadFile(fullPath) //nolint:gosec // path is built from directory entries under the site pages dir
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading %s: %v\n", relPath, err)
				os.Exit(1)
			}
			issues = append(issues, checkPage(relPath, content, validPaths)...)
		}
	}

	// Print summary line.
	fmt.Printf("✓ %d pages checked\n", pageCount)
	for _, issue := range issues {
		fmt.Printf("✗ %s\n", issue)
	}

	if len(issues) > 0 {
		fmt.Printf("\n%d issue(s) found\n", len(issues))
		os.Exit(1)
	}
}

// checkPage validates a single page and returns a slice of issue descriptions.
func checkPage(relPath string, content []byte, validPaths map[string]bool) []string {
	var issues []string

	// Check for missing title in frontmatter.
	title := frontmatter.ParseStringField(content, "title")
	if title == "" {
		issues = append(issues, fmt.Sprintf("%s: missing title", relPath))
	}

	// Check for empty page content (body after stripping frontmatter).
	body := strings.TrimSpace(frontmatter.Strip(string(content)))
	if body == "" {
		issues = append(issues, fmt.Sprintf("%s: empty page content", relPath))
	}

	// Check for broken internal links (site-absolute paths starting with "/").
	for _, dest := range internalLinks(string(content)) {
		// Strip fragment and query string.
		if idx := strings.IndexAny(dest, "#?"); idx >= 0 {
			dest = dest[:idx]
		}
		// Normalise trailing slash: "/" alone maps to the index page.
		dest = strings.TrimSuffix(dest, "/")
		if dest == "" {
			dest = "/index"
		}
		if u, err := url.PathUnescape(dest); err == nil {
			dest = u
		}
		dest = path.Clean(dest)
		if !validPaths[dest] {
			issues = append(issues, fmt.Sprintf("%s: broken link → %s (page not found)", relPath, dest))
		}
	}

	return issues
}

// internalLinks returns the destinations of all non-image Markdown links in
// content that are site-absolute paths (a single leading "/"). Links inside
// fenced code blocks and inline code spans are ignored, as are external,
// scheme and protocol-relative ("//host") links.
func internalLinks(content string) []string {
	var prose []string
	var fence string // the opening fence marker while inside a fenced block
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if fence != "" {
			// A closing fence is at least as long as the opener and has no info string.
			if strings.HasPrefix(trimmed, fence) && strings.Trim(trimmed, fence[:1]) == "" {
				fence = ""
			}
			continue
		}
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fence = strings.Repeat(trimmed[:1], len(trimmed)-len(strings.TrimLeft(trimmed, trimmed[:1])))
			continue
		}
		prose = append(prose, inlineCodeRe.ReplaceAllString(line, ""))
	}
	text := strings.Join(prose, "\n")

	var links []string
	for _, m := range linkStartRe.FindAllStringSubmatchIndex(text, -1) {
		if text[m[2]:m[3]] == "!" {
			continue // skip image links
		}
		dest := linkDestination(text[m[1]:])
		if strings.HasPrefix(dest, "/") && !strings.HasPrefix(dest, "//") {
			links = append(links, dest)
		}
	}
	return links
}

// linkDestination parses a link destination from s, which starts right after
// the "(" of an inline link. It accepts "<...>" destinations and bare ones
// with balanced parentheses, and stops at whitespace so an optional title
// ("...") is left out.
func linkDestination(s string) string {
	s = strings.TrimLeft(s, " \t")
	if strings.HasPrefix(s, "<") {
		if end := strings.IndexAny(s, ">\n"); end > 0 && s[end] == '>' {
			return s[1:end]
		}
		return ""
	}
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ' ', '\t', '\n':
			return s[:i]
		case '(':
			depth++
		case ')':
			if depth == 0 {
				return s[:i]
			}
			depth--
		}
	}
	return s
}

// buildValidPaths returns the set of internal link paths that resolve to an
// existing page or file. Paths are slash-prefixed (e.g. "/about", "/blog",
// "/blog/first-post", "/docs/f.pdf").
func buildValidPaths(siteDir string) map[string]bool {
	pagesDir := page.PagesDir(siteDir)
	valid := make(map[string]bool)

	// Files the builder copies verbatim to the output root: the static dir and
	// non-Markdown files under pages/.
	addFiles := func(root string, skipMarkdown bool) {
		_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || (skipMarkdown && strings.HasSuffix(d.Name(), ".md")) {
				return nil //nolint:nilerr // a missing or unreadable dir just means no files to link to
			}
			if rel, err := filepath.Rel(root, p); err == nil {
				valid["/"+filepath.ToSlash(rel)] = true
			}
			return nil
		})
	}
	addFiles(filepath.Join(siteDir, defaultStaticDir), false)
	addFiles(pagesDir, true)

	entries, err := os.ReadDir(pagesDir)
	if err != nil {
		return valid
	}

	for _, e := range entries {
		if e.IsDir() {
			name := e.Name()
			sectionPath := filepath.Join(pagesDir, name)
			// Section index is reachable as /name and /name/index.
			indexPath := filepath.Join(sectionPath, "index.md")
			if _, err := os.Stat(indexPath); err == nil {
				valid["/"+name] = true
				valid["/"+name+"/index"] = true
				valid["/"+name+"/index.html"] = true
			}
			// Sub-pages within the section.
			subEntries, err := os.ReadDir(sectionPath)
			if err == nil {
				for _, se := range subEntries {
					if !se.IsDir() && strings.HasSuffix(se.Name(), ".md") {
						pageName := strings.TrimSuffix(se.Name(), ".md")
						valid["/"+name+"/"+pageName] = true
						valid["/"+name+"/"+pageName+".html"] = true
					}
				}
			}
		} else if strings.HasSuffix(e.Name(), ".md") {
			pageName := strings.TrimSuffix(e.Name(), ".md")
			valid["/"+pageName] = true
			valid["/"+pageName+".html"] = true
		}
	}

	return valid
}
