package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ChristianKreuzberger/press/internal/frontmatter"
	"github.com/ChristianKreuzberger/press/internal/page"
)

// internalLinkRe matches Markdown links whose destination starts with "/".
// Group 1 is "!" for image links (to be skipped), group 2 is the link text,
// and group 3 is the destination.
var internalLinkRe = regexp.MustCompile(`(!?)\[([^\]]*)\]\((/[^)]*)\)`)

func runCheck(_ []string) {
	siteDir := mustGetwd()
	pagesDir := page.PagesDir(siteDir)

	var issues []string
	pageCount := 0

	// Build the set of valid internal link paths.
	validPaths := buildValidPaths(pagesDir)

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
			content, err := os.ReadFile(fullPath)
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

	// Check for broken internal links (absolute paths starting with "/").
	for _, m := range internalLinkRe.FindAllStringSubmatch(string(content), -1) {
		if m[1] == "!" {
			continue // skip image links
		}
		dest := m[3]
		// Strip fragment.
		if idx := strings.IndexByte(dest, '#'); idx >= 0 {
			dest = dest[:idx]
		}
		// Strip query string.
		if idx := strings.IndexByte(dest, '?'); idx >= 0 {
			dest = dest[:idx]
		}
		// Normalise trailing slash: "/" alone maps to the index page.
		dest = strings.TrimSuffix(dest, "/")
		if dest == "" {
			dest = "/index"
		}
		if !validPaths[dest] {
			issues = append(issues, fmt.Sprintf("%s: broken link → %s (page not found)", relPath, dest))
		}
	}

	return issues
}

// buildValidPaths returns the set of internal link paths that resolve to an
// existing page. Paths are slash-prefixed (e.g. "/about", "/blog", "/blog/first-post").
func buildValidPaths(pagesDir string) map[string]bool {
	valid := make(map[string]bool)

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
			}
			// Sub-pages within the section.
			subEntries, err := os.ReadDir(sectionPath)
			if err == nil {
				for _, se := range subEntries {
					if !se.IsDir() && strings.HasSuffix(se.Name(), ".md") {
						pageName := strings.TrimSuffix(se.Name(), ".md")
						valid["/"+name+"/"+pageName] = true
					}
				}
			}
		} else if strings.HasSuffix(e.Name(), ".md") {
			pageName := strings.TrimSuffix(e.Name(), ".md")
			valid["/"+pageName] = true
		}
	}

	return valid
}
