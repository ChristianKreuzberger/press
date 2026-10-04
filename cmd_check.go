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
	validPaths, draftPaths := buildValidPaths(siteDir)

	// A problem with one file or directory is reported as an issue so the rest
	// of the site is still checked.
	checkFile := func(relPath, fullPath, baseDir string) {
		pageCount++
		content, err := os.ReadFile(fullPath) //nolint:gosec // path is built from directory entries under the site pages dir
		if err != nil {
			issues = append(issues, fmt.Sprintf("%s: cannot read page: %v", relPath, err))
			return
		}
		issues = append(issues, checkPage(relPath, baseDir, content, validPaths, draftPaths)...)
	}

	entries, err := os.ReadDir(pagesDir)
	if err != nil && !os.IsNotExist(err) {
		issues = append(issues, fmt.Sprintf("pages/: cannot read directory: %v", err))
	}

	// Top-level pages are written to the output root.
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			checkFile(e.Name(), filepath.Join(pagesDir, e.Name()), "/")
		}
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
			issues = append(issues, fmt.Sprintf("%s/: cannot read section: %v", sectionName, err))
			continue
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
				issues = append(issues, fmt.Sprintf("%s/index.md: cannot read page: %v", sectionName, statErr))
				continue
			}
			// Section directory without index.md.
			issues = append(issues, fmt.Sprintf("%s/: section has no index.md", sectionName))
			continue
		}

		// Check all .md files in this section; they are written to /<section>/.
		for _, sf := range sectionFiles {
			if sf.IsDir() || !strings.HasSuffix(sf.Name(), ".md") {
				continue
			}
			checkFile(sectionName+"/"+sf.Name(), filepath.Join(sectionPath, sf.Name()), "/"+sectionName+"/")
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

// schemeRe matches a URL scheme such as "https:" or "mailto:".
var schemeRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*:`)

// htmlHrefRe matches the href of a raw HTML anchor. Group 2 or 3 is the value.
var htmlHrefRe = regexp.MustCompile(`(?i)<a\s[^>]*?href\s*=\s*("([^"]*)"|'([^']*)')`)

// checkPage validates a single page and returns a slice of issue descriptions.
// baseDir is the output directory of the page ("/" or "/<section>/"), which
// relative links are resolved against.
func checkPage(relPath, baseDir string, content []byte, validPaths, draftPaths map[string]bool) []string {
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

	// Check for broken internal links: site-absolute ("/x") and relative ("x").
	for _, dest := range pageLinks(string(content)) {
		target, ok := internalTarget(dest, baseDir)
		if !ok {
			continue
		}
		switch {
		case validPaths[target]:
		case draftPaths[target]:
			issues = append(issues, fmt.Sprintf("%s: broken link → %s (page is a draft; build skips it)", relPath, target))
		default:
			issues = append(issues, fmt.Sprintf("%s: broken link → %s (page not found)", relPath, target))
		}
	}

	return issues
}

// internalTarget turns a link destination into the cleaned site path it points
// at, resolving relative destinations against baseDir. ok is false for
// destinations that are not checked: external and scheme links, protocol-
// relative links, and pure fragments or queries. A relative link that climbs
// above the site root yields a path that is never valid.
func internalTarget(dest, baseDir string) (target string, ok bool) {
	if idx := strings.IndexAny(dest, "#?"); idx >= 0 {
		dest = dest[:idx]
	}
	if dest == "" || strings.HasPrefix(dest, "//") || schemeRe.MatchString(dest) {
		return "", false
	}
	if u, err := url.PathUnescape(dest); err == nil {
		dest = u
	}
	if !strings.HasPrefix(dest, "/") {
		dest = baseDir + dest
		// path.Clean would silently turn "/../x" into "/x".
		depth := 0
		for _, seg := range strings.Split(dest, "/") {
			switch seg {
			case "", ".":
			case "..":
				depth--
			default:
				depth++
			}
			if depth < 0 {
				return dest, true
			}
		}
	}
	// "/" alone maps to the index page.
	dest = path.Clean(strings.TrimSuffix(dest, "/"))
	if dest == "." || dest == "/" {
		dest = "/index"
	}
	return dest, true
}

// pageLinks returns the destinations of all non-image Markdown links and raw
// HTML anchors in content. Links inside fenced code blocks and inline code
// spans are ignored. Destinations are returned as written.
func pageLinks(content string) []string {
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
		if dest := linkDestination(text[m[1]:]); dest != "" {
			links = append(links, dest)
		}
	}
	for _, m := range htmlHrefRe.FindAllStringSubmatch(text, -1) {
		links = append(links, m[2]+m[3])
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

// buildValidPaths returns the set of internal link paths that resolve to a
// built page or copied file, and a second set of page paths that exist but are
// skipped by the build because they are drafts (or in a draft section). Paths
// are slash-prefixed (e.g. "/about", "/blog", "/blog/first-post",
// "/docs/f.pdf").
func buildValidPaths(siteDir string) (valid, drafts map[string]bool) {
	pagesDir := page.PagesDir(siteDir)
	valid = make(map[string]bool)
	drafts = make(map[string]bool)

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
		return valid, drafts
	}

	// isDraft is false when the file cannot be read: check reports that
	// separately, and an unreadable page is not known to be skipped.
	isDraft := func(p string) bool {
		content, err := os.ReadFile(p) //nolint:gosec // path is built from directory entries under the site pages dir
		return err == nil && frontmatter.ParseDraft(content)
	}
	// add records the URLs of one built page: "/x" and "/x.html".
	add := func(urlPath string, draft bool) {
		set := valid
		if draft {
			set = drafts
		}
		set[urlPath] = true
		set[urlPath+".html"] = true
	}

	for _, e := range entries {
		if e.IsDir() {
			name := e.Name()
			sectionPath := filepath.Join(pagesDir, name)
			// The builder writes only <section>/index.html; a server serves it as
			// /name and /name/. A draft section index drops the whole section.
			indexPath := filepath.Join(sectionPath, "index.md")
			if _, err := os.Stat(indexPath); err != nil {
				continue
			}
			sectionDraft := isDraft(indexPath)
			set := valid
			if sectionDraft {
				set = drafts
			}
			set["/"+name] = true
			set["/"+name+"/index.html"] = true
			// Sub-pages within the section.
			subEntries, err := os.ReadDir(sectionPath)
			if err == nil {
				for _, se := range subEntries {
					if !se.IsDir() && strings.HasSuffix(se.Name(), ".md") {
						pageName := strings.TrimSuffix(se.Name(), ".md")
						if pageName == "index" {
							continue
						}
						add("/"+name+"/"+pageName, sectionDraft || isDraft(filepath.Join(sectionPath, se.Name())))
					}
				}
			}
		} else if strings.HasSuffix(e.Name(), ".md") {
			pageName := strings.TrimSuffix(e.Name(), ".md")
			add("/"+pageName, isDraft(filepath.Join(pagesDir, e.Name())))
		}
	}

	return valid, drafts
}
