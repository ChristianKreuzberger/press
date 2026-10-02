package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2EFrontmatterCRLF covers files saved with Windows line endings: a CRLF
// draft section must stay out of dist and the nav, and CRLF frontmatter must
// not leak into the rendered page.
func TestE2EFrontmatterCRLF(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	pagesDir := filepath.Join(siteDir, "pages")

	if err := os.MkdirAll(filepath.Join(pagesDir, "priv"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(pagesDir, "priv", "index.md"),
		"---\r\ntitle: \"Private\"\r\ndraft: true\r\n---\r\n# Private\r\n\r\nSecret.\r\n")
	writeFile(t, filepath.Join(pagesDir, "loose.md"),
		"---\r\ntitle: \"Loose\"\r\ndraft: True # wip\r\n---\r\n# Loose\r\n")
	writeFile(t, filepath.Join(pagesDir, "about.md"),
		"---\r\ntitle: \"About Me\"\r\nweight: 3\r\n---\r\n# About\r\n\r\nHello.\r\n")
	writeFile(t, filepath.Join(pagesDir, "empty.md"), "---\n---\n# Empty\n")

	run(t, siteDir, "build")
	dist := filepath.Join(siteDir, "dist")

	if _, err := os.Stat(filepath.Join(dist, "priv")); err == nil {
		t.Error("CRLF draft section must not create dist/priv")
	}
	if _, err := os.Stat(filepath.Join(dist, "loose.html")); err == nil {
		t.Error("'draft: True # wip' page must not be built")
	}

	index := readFile(t, filepath.Join(dist, "index.html"))
	if strings.Contains(index, "priv") {
		t.Error("nav must not link to the CRLF draft section")
	}

	about := readFile(t, filepath.Join(dist, "about.html"))
	// The theme CSS contains "font-weight:", so match the exact frontmatter text.
	if strings.Contains(about, "title: ") || strings.Contains(about, "weight: 3") || strings.Contains(about, "<hr") {
		t.Errorf("CRLF frontmatter leaked into page content: %s", about)
	}
	empty := readFile(t, filepath.Join(dist, "empty.html"))
	if strings.Contains(empty, "<hr") {
		t.Errorf("empty frontmatter block must be stripped, got: %s", empty)
	}
}

// TestE2EPageTitles checks title resolution: frontmatter title first, then
// the first H1 outside code fences as plain text, then the file name.
func TestE2EPageTitles(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	pagesDir := filepath.Join(siteDir, "pages")

	writeFile(t, filepath.Join(pagesDir, "fm.md"),
		"---\ntitle: \"From Frontmatter\"\n---\nNo heading here.\n")
	writeFile(t, filepath.Join(pagesDir, "both.md"),
		"---\ntitle: \"FM Wins\"\n---\n# Heading Loses\n")
	writeFile(t, filepath.Join(pagesDir, "fence.md"),
		"```sh\n# not a title\n```\n\n# Real Title\n")
	writeFile(t, filepath.Join(pagesDir, "rich.md"),
		"# Post *em* &amp; <b>bold</b>\n")
	writeFile(t, filepath.Join(pagesDir, "plain.md"), "Just text.\n")

	run(t, siteDir, "build")
	dist := filepath.Join(siteDir, "dist")

	for file, want := range map[string]string{
		"fm.html":    "<title>From Frontmatter</title>",
		"both.html":  "<title>FM Wins</title>",
		"fence.html": "<title>Real Title</title>",
		"rich.html":  "<title>Post em &amp; bold</title>",
		"plain.html": "<title>plain</title>",
	} {
		if got := readFile(t, filepath.Join(dist, file)); !strings.Contains(got, want) {
			t.Errorf("%s: want %s in output", file, want)
		}
	}
	// The nav on every page uses the same resolved titles.
	index := readFile(t, filepath.Join(dist, "index.html"))
	if !strings.Contains(index, ">From Frontmatter</a>") || !strings.Contains(index, ">Real Title</a>") {
		t.Errorf("nav must use resolved titles, got: %s", index)
	}
}
