package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ChristianKreuzberger/press/internal/page"
	"github.com/ChristianKreuzberger/press/internal/section"
)

func TestBuildProducesHTMLFiles(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := page.Create(siteDir, "index", []byte("# Home\n\nWelcome!\n")); err != nil {
		t.Fatal(err)
	}
	if err := page.Create(siteDir, "about", []byte("# About\n\nLearn more.\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	for _, name := range []string{"index.html", "about.html"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Errorf("expected output file %s to exist", name)
		}
	}
}

func TestBuildHTMLContent(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := page.Create(siteDir, "index", []byte("# Home\n\nWelcome to the site.\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(content)

	if !strings.Contains(html, "<h1") || !strings.Contains(html, ">Home</h1>") {
		t.Errorf("expected <h1>Home</h1> in output, got:\n%s", html)
	}
	if !strings.Contains(html, "Welcome to the site.") {
		t.Errorf("expected page body in output, got:\n%s", html)
	}
	if !strings.Contains(html, "<title>Home</title>") {
		t.Errorf("expected <title>Home</title> in output, got:\n%s", html)
	}
}

func TestBuildNavigationLinks(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}
	if err := page.Create(siteDir, "about", []byte("# About\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(outDir, "index.html"))
	html := string(content)

	if !strings.Contains(html, "about.html") {
		t.Errorf("index.html should contain a link to about.html, got:\n%s", html)
	}
	if !strings.Contains(html, "index.html") {
		t.Errorf("index.html should contain a link to itself, got:\n%s", html)
	}
}

func TestBuildCustomTemplate(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	customTmpl := `<html><head><title>{{.Title}}</title></head><body>CUSTOM {{.Content}}</body></html>`
	if err := os.WriteFile(filepath.Join(siteDir, "template.html"), []byte(customTmpl), 0644); err != nil {
		t.Fatal(err)
	}
	if err := page.Create(siteDir, "index", []byte("# Test\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(outDir, "index.html"))
	if !strings.Contains(string(content), "CUSTOM") {
		t.Errorf("expected custom template to be used, got:\n%s", content)
	}
}

func TestBuildNoPages(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build with no pages should not fail: %v", err)
	}
}

func TestBuildFallbackTitleFromFilename(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	// Page without a heading — filename is used as title
	if err := page.Create(siteDir, "contact", []byte("Send us a message.\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(outDir, "contact.html"))
	if !strings.Contains(string(content), "<title>contact</title>") {
		t.Errorf("expected filename as fallback title, got:\n%s", content)
	}
}

func TestBuildWithSection(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}
	if err := section.Create(siteDir, "blog", []byte("# Blog\n\nAll posts.\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Section index should be generated.
	if _, err := os.Stat(filepath.Join(outDir, "blog", "index.html")); err != nil {
		t.Fatal("build should produce dist/blog/index.html")
	}

	content := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))
	if !strings.Contains(content, "All posts.") {
		t.Errorf("dist/blog/index.html should contain section body, got:\n%s", content)
	}
	if !strings.Contains(content, "<title>Blog</title>") {
		t.Errorf("dist/blog/index.html should have <title>Blog</title>, got:\n%s", content)
	}
}

func TestBuildSectionNavLinksFromTopLevel(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}
	if err := section.Create(siteDir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Top-level index.html should link to the section index with a root-relative URL.
	content := string(mustRead(t, filepath.Join(outDir, "index.html")))
	if !strings.Contains(content, "blog/index.html") {
		t.Errorf("dist/index.html nav should link to blog/index.html, got:\n%s", content)
	}
}

func TestBuildSectionNavLinksFromSection(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := page.Create(siteDir, "about", []byte("# About\n")); err != nil {
		t.Fatal(err)
	}
	if err := section.Create(siteDir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Section page nav should prefix top-level page links with "../".
	content := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))
	if !strings.Contains(content, "../about.html") {
		t.Errorf("dist/blog/index.html nav should link to ../about.html, got:\n%s", content)
	}
	if !strings.Contains(content, "../blog/index.html") {
		t.Errorf("dist/blog/index.html nav should link to ../blog/index.html, got:\n%s", content)
	}
}

func TestBuildSectionWithMultiplePages(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := section.Create(siteDir, "docs", []byte("# Docs\n")); err != nil {
		t.Fatal(err)
	}
	// Add a non-index page to the section.
	docsDir := filepath.Join(siteDir, "pages", "docs")
	if err := os.WriteFile(filepath.Join(docsDir, "getting-started.md"), []byte("# Getting Started\n\nInstall and go.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "docs", "index.html")); err != nil {
		t.Fatal("build should produce dist/docs/index.html")
	}
	if _, err := os.Stat(filepath.Join(outDir, "docs", "getting-started.html")); err != nil {
		t.Fatal("build should produce dist/docs/getting-started.html")
	}

	content := string(mustRead(t, filepath.Join(outDir, "docs", "getting-started.html")))
	if !strings.Contains(content, "Install and go.") {
		t.Errorf("dist/docs/getting-started.html should contain page body, got:\n%s", content)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestBuildNavSortedByWeight(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	// Create pages with different weights; filesystem order (alpha) is: first, last, second
	// but nav order should be: weight=1 (first), weight=2 (second), weight=0/unset (last).
	if err := page.Create(siteDir, "first", []byte("---\ntitle: \"First\"\nweight: 1\n---\n# First\n")); err != nil {
		t.Fatal(err)
	}
	if err := page.Create(siteDir, "second", []byte("---\ntitle: \"Second\"\nweight: 2\n---\n# Second\n")); err != nil {
		t.Fatal(err)
	}
	if err := page.Create(siteDir, "last", []byte("---\ntitle: \"Last\"\nweight: 0\n---\n# Last\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "first.html")))

	firstPos := strings.Index(content, "first.html")
	secondPos := strings.Index(content, "second.html")
	lastPos := strings.Index(content, "last.html")

	if firstPos == -1 || secondPos == -1 || lastPos == -1 {
		t.Fatalf("expected all pages in nav, got:\n%s", content)
	}
	if firstPos > secondPos {
		t.Errorf("expected 'first' (weight=1) before 'second' (weight=2) in nav")
	}
	if secondPos > lastPos {
		t.Errorf("expected 'second' (weight=2) before 'last' (weight=0/unset) in nav")
	}
}

func TestBuildNavWeightWithSection(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	// Section with weight=1 should appear before an unweighted top-level page.
	if err := section.Create(siteDir, "blog", []byte("---\ntitle: \"Blog\"\nweight: 1\n---\n# Blog\n")); err != nil {
		t.Fatal(err)
	}
	if err := page.Create(siteDir, "about", []byte("# About\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "about.html")))

	blogPos := strings.Index(content, "blog/index.html")
	aboutPos := strings.Index(content, "about.html")

	if blogPos == -1 || aboutPos == -1 {
		t.Fatalf("expected both nav entries, got:\n%s", content)
	}
	if blogPos > aboutPos {
		t.Errorf("expected 'blog' (weight=1) before 'about' (no weight) in nav")
	}
}

func TestBuildSectionTOCByTitleAsc(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	indexContent := "---\ntitle: \"Blog\"\ntoc_sort: \"title\"\ntoc_order: \"asc\"\n---\n# Blog\n"
	if err := section.Create(siteDir, "blog", []byte(indexContent)); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(siteDir, "pages", "blog")
	if err := os.WriteFile(filepath.Join(blogDir, "zebra.md"), []byte("# Zebra Post\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "apple.md"), []byte("# Apple Post\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))

	// TOC section must be present.
	if !strings.Contains(content, "class=\"toc\"") {
		t.Fatalf("expected TOC section in section index, got:\n%s", content)
	}

	applePos := strings.Index(content, "apple.html")
	zebraPos := strings.Index(content, "zebra.html")
	if applePos == -1 || zebraPos == -1 {
		t.Fatalf("expected both TOC entries, got:\n%s", content)
	}
	if applePos > zebraPos {
		t.Errorf("expected apple.html before zebra.html in TOC (title asc)")
	}
}

func TestBuildSectionTOCByTitleDesc(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	indexContent := "---\ntitle: \"Blog\"\ntoc_sort: \"title\"\ntoc_order: \"desc\"\n---\n# Blog\n"
	if err := section.Create(siteDir, "blog", []byte(indexContent)); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(siteDir, "pages", "blog")
	if err := os.WriteFile(filepath.Join(blogDir, "alpha.md"), []byte("# Alpha\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "zeta.md"), []byte("# Zeta\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))
	alphaPos := strings.Index(content, "alpha.html")
	zetaPos := strings.Index(content, "zeta.html")

	if alphaPos == -1 || zetaPos == -1 {
		t.Fatalf("expected both TOC entries, got:\n%s", content)
	}
	if zetaPos > alphaPos {
		t.Errorf("expected zeta.html before alpha.html in TOC (title desc)")
	}
}

func TestBuildSectionTOCByCreatedAtAsc(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	indexContent := "---\ntitle: \"Blog\"\ntoc_sort: \"created_at\"\ntoc_order: \"asc\"\n---\n# Blog\n"
	if err := section.Create(siteDir, "blog", []byte(indexContent)); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(siteDir, "pages", "blog")
	older := "---\ntitle: \"Older\"\ncreated_at: \"2024-01-01T00:00:00Z\"\n---\n# Older\n"
	newer := "---\ntitle: \"Newer\"\ncreated_at: \"2025-06-01T00:00:00Z\"\n---\n# Newer\n"
	if err := os.WriteFile(filepath.Join(blogDir, "older-post.md"), []byte(older), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "newer-post.md"), []byte(newer), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))
	olderPos := strings.Index(content, "older-post.html")
	newerPos := strings.Index(content, "newer-post.html")

	if olderPos == -1 || newerPos == -1 {
		t.Fatalf("expected both TOC entries, got:\n%s", content)
	}
	if olderPos > newerPos {
		t.Errorf("expected older-post before newer-post in TOC (created_at asc)")
	}
}

func TestBuildSectionTOCByCreatedAtDesc(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	indexContent := "---\ntitle: \"Blog\"\ntoc_sort: \"created_at\"\ntoc_order: \"desc\"\n---\n# Blog\n"
	if err := section.Create(siteDir, "blog", []byte(indexContent)); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(siteDir, "pages", "blog")
	older := "---\ntitle: \"Older\"\ncreated_at: \"2024-01-01T00:00:00Z\"\n---\n# Older\n"
	newer := "---\ntitle: \"Newer\"\ncreated_at: \"2025-06-01T00:00:00Z\"\n---\n# Newer\n"
	if err := os.WriteFile(filepath.Join(blogDir, "older-post.md"), []byte(older), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "newer-post.md"), []byte(newer), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))
	olderPos := strings.Index(content, "older-post.html")
	newerPos := strings.Index(content, "newer-post.html")

	if olderPos == -1 || newerPos == -1 {
		t.Fatalf("expected both TOC entries, got:\n%s", content)
	}
	if newerPos > olderPos {
		t.Errorf("expected newer-post before older-post in TOC (created_at desc)")
	}
}

func TestBuildSectionTOCByUpdatedAtDesc(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	indexContent := "---\ntitle: \"Blog\"\ntoc_sort: \"updated_at\"\ntoc_order: \"desc\"\n---\n# Blog\n"
	if err := section.Create(siteDir, "blog", []byte(indexContent)); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(siteDir, "pages", "blog")
	early := "---\ntitle: \"Early Update\"\nupdated_at: \"2023-03-01T00:00:00Z\"\n---\n# Early Update\n"
	late := "---\ntitle: \"Late Update\"\nupdated_at: \"2026-02-01T00:00:00Z\"\n---\n# Late Update\n"
	if err := os.WriteFile(filepath.Join(blogDir, "early.md"), []byte(early), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "late.md"), []byte(late), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))
	earlyPos := strings.Index(content, "early.html")
	latePos := strings.Index(content, "late.html")

	if earlyPos == -1 || latePos == -1 {
		t.Fatalf("expected both TOC entries, got:\n%s", content)
	}
	if latePos > earlyPos {
		t.Errorf("expected late.html before early.html in TOC (updated_at desc)")
	}
}

func TestBuildSectionTOCByWeight(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	// toc_sort defaults to "weight" when unspecified; use explicit here.
	indexContent := "---\ntitle: \"Blog\"\ntoc_sort: \"weight\"\ntoc_order: \"asc\"\n---\n# Blog\n"
	if err := section.Create(siteDir, "blog", []byte(indexContent)); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(siteDir, "pages", "blog")
	first := "---\ntitle: \"First\"\nweight: 1\n---\n# First\n"
	second := "---\ntitle: \"Second\"\nweight: 2\n---\n# Second\n"
	unweighted := "---\ntitle: \"Unweighted\"\nweight: 0\n---\n# Unweighted\n"
	if err := os.WriteFile(filepath.Join(blogDir, "first.md"), []byte(first), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "second.md"), []byte(second), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "unweighted.md"), []byte(unweighted), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))
	firstPos := strings.Index(content, "first.html")
	secondPos := strings.Index(content, "second.html")
	unweightedPos := strings.Index(content, "unweighted.html")

	if firstPos == -1 || secondPos == -1 || unweightedPos == -1 {
		t.Fatalf("expected all TOC entries, got:\n%s", content)
	}
	if firstPos > secondPos {
		t.Errorf("expected first (weight=1) before second (weight=2) in TOC")
	}
	if secondPos > unweightedPos {
		t.Errorf("expected second (weight=2) before unweighted (weight=0) in TOC")
	}
}

func TestBuildSectionTOCDefaultWeight(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	// No toc_sort/toc_order in frontmatter — defaults to weight asc.
	indexContent := "---\ntitle: \"Blog\"\n---\n# Blog\n"
	if err := section.Create(siteDir, "blog", []byte(indexContent)); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(siteDir, "pages", "blog")
	if err := os.WriteFile(filepath.Join(blogDir, "post.md"), []byte("# Post\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))
	if !strings.Contains(content, "post.html") {
		t.Errorf("expected post.html in TOC, got:\n%s", content)
	}
}

func TestBuildTOCEmptyForNonSectionPages(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := page.Create(siteDir, "about", []byte("# About\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "about.html")))
	if strings.Contains(content, "class=\"toc\"") {
		t.Errorf("non-section pages should not have a TOC section, got:\n%s", content)
	}
}

func TestBuildTOCEmptyForSectionChildPages(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := section.Create(siteDir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(siteDir, "pages", "blog")
	if err := os.WriteFile(filepath.Join(blogDir, "post.md"), []byte("# Post\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// The child page (not the index) should NOT have a TOC.
	content := string(mustRead(t, filepath.Join(outDir, "blog", "post.html")))
	if strings.Contains(content, "class=\"toc\"") {
		t.Errorf("section child pages should not have a TOC, got:\n%s", content)
	}
}

func TestBuildSectionTOCIndexNotIncluded(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	indexContent := "---\ntitle: \"Blog\"\ntoc_sort: \"title\"\ntoc_order: \"asc\"\n---\n# Blog\n"
	if err := section.Create(siteDir, "blog", []byte(indexContent)); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(siteDir, "pages", "blog")
	if err := os.WriteFile(filepath.Join(blogDir, "post.md"), []byte("# Post\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))
	// The index page should not link to itself in the TOC.
	if strings.Contains(content, ">index.html<") || strings.Contains(content, "href=\"index.html\"") {
		t.Errorf("index.html should not appear in its own TOC, got:\n%s", content)
	}
}

func TestWeightLess(t *testing.T) {
	tests := []struct {
		wi, wj int
		want   bool
	}{
		{1, 2, true},
		{2, 1, false},
		{0, 5, false}, // unset sorts after 5
		{5, 0, true},  // 5 sorts before unset
		{0, 0, false}, // equal
	}
	for _, tc := range tests {
		if got := weightLess(tc.wi, tc.wj); got != tc.want {
			t.Errorf("weightLess(%d, %d) = %v, want %v", tc.wi, tc.wj, got, tc.want)
		}
	}
}

func TestBuildSectionTOCByWeightDesc(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	indexContent := "---\ntitle: \"Blog\"\ntoc_sort: \"weight\"\ntoc_order: \"desc\"\n---\n# Blog\n"
	if err := section.Create(siteDir, "blog", []byte(indexContent)); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(siteDir, "pages", "blog")
	for name, body := range map[string]string{
		"heavy.md":      "---\nweight: 2\n---\n# Heavy\n",
		"light.md":      "---\nweight: 1\n---\n# Light\n",
		"unweighted.md": "---\nweight: 0\n---\n# Unweighted\n",
	} {
		if err := os.WriteFile(filepath.Join(blogDir, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))
	heavyPos := strings.Index(content, "heavy.html")
	lightPos := strings.Index(content, "light.html")
	unweightedPos := strings.Index(content, "unweighted.html")

	if heavyPos == -1 || lightPos == -1 || unweightedPos == -1 {
		t.Fatalf("expected all TOC entries, got:\n%s", content)
	}
	if heavyPos > lightPos {
		t.Errorf("expected heavy (weight=2) before light (weight=1) in desc TOC")
	}
	if lightPos > unweightedPos {
		t.Errorf("expected light (weight=1) before unweighted (weight=0) in desc TOC")
	}
}

func TestBuildSectionTOCByUpdatedAtAsc(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	indexContent := "---\ntitle: \"Blog\"\ntoc_sort: \"updated_at\"\ntoc_order: \"asc\"\n---\n# Blog\n"
	if err := section.Create(siteDir, "blog", []byte(indexContent)); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(siteDir, "pages", "blog")
	if err := os.WriteFile(filepath.Join(blogDir, "older.md"), []byte("---\nupdated_at: \"2024-01-01T00:00:00Z\"\n---\n# Older\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "newer.md"), []byte("---\nupdated_at: \"2025-06-01T00:00:00Z\"\n---\n# Newer\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))
	olderPos := strings.Index(content, "older.html")
	newerPos := strings.Index(content, "newer.html")

	if olderPos == -1 || newerPos == -1 {
		t.Fatalf("expected both TOC entries, got:\n%s", content)
	}
	if olderPos > newerPos {
		t.Errorf("expected older.html before newer.html in TOC (updated_at asc)")
	}
}

func TestBuildSectionTOCFallbackTitle(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := section.Create(siteDir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	// Page without a markdown heading — filename is used as the TOC title.
	blogDir := filepath.Join(siteDir, "pages", "blog")
	if err := os.WriteFile(filepath.Join(blogDir, "no-heading.md"), []byte("Just some content.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))
	if !strings.Contains(content, "no-heading.html") {
		t.Errorf("expected no-heading.html in TOC, got:\n%s", content)
	}
	if !strings.Contains(content, ">no-heading<") {
		t.Errorf("expected fallback title no-heading in TOC entry text, got:\n%s", content)
	}
}

func TestBuildInvalidTemplate(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := os.WriteFile(filepath.Join(siteDir, "template.html"), []byte("{{invalid template{{"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err == nil {
		t.Error("expected error for invalid template, got nil")
	}
}

func TestBuildUnreadableTemplate(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission test: running as root")
	}
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	tmplPath := filepath.Join(siteDir, "template.html")
	if err := os.WriteFile(tmplPath, []byte("<html></html>"), 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(tmplPath, 0644) }) //nolint:errcheck
	if _, err := os.ReadFile(tmplPath); err == nil {
		t.Skip("skipping: filesystem does not enforce permission bits")
	}

	if err := Build(siteDir, outDir, false); err == nil {
		t.Error("expected error for unreadable template, got nil")
	}
}

func TestBuildUnreadablePage(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("skipping permission test: running as root")
	}
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}
	pagePath := filepath.Join(siteDir, "pages", "index.md")
	if err := os.Chmod(pagePath, 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(pagePath, 0644) }) //nolint:errcheck
	if _, err := os.ReadFile(pagePath); err == nil {
		t.Skip("skipping: filesystem does not enforce permission bits")
	}

	if err := Build(siteDir, outDir, false); err == nil {
		t.Error("expected error for unreadable page, got nil")
	}
}

func TestBuildCopiesStaticAssets(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	// Create a markdown page so the build has something to do.
	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}

	// Place a non-Markdown file directly under pages/.
	pagesDir := filepath.Join(siteDir, "pages")
	imgData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG magic bytes
	if err := os.WriteFile(filepath.Join(pagesDir, "logo.png"), imgData, 0644); err != nil {
		t.Fatal(err)
	}

	// Place a non-Markdown file in a subdirectory under pages/.
	assetsDir := filepath.Join(pagesDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	cssData := []byte("body { margin: 0; }")
	if err := os.WriteFile(filepath.Join(assetsDir, "style.css"), cssData, 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Top-level asset should be copied.
	got, err := os.ReadFile(filepath.Join(outDir, "logo.png"))
	if err != nil {
		t.Fatalf("expected dist/logo.png to exist: %v", err)
	}
	if string(got) != string(imgData) {
		t.Errorf("dist/logo.png content mismatch")
	}

	// Subdirectory asset should be copied with directory preserved.
	got, err = os.ReadFile(filepath.Join(outDir, "assets", "style.css"))
	if err != nil {
		t.Fatalf("expected dist/assets/style.css to exist: %v", err)
	}
	if string(got) != string(cssData) {
		t.Errorf("dist/assets/style.css content mismatch")
	}
}

func TestBuildCopiesStaticAssetsInSection(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := section.Create(siteDir, "portfolio", []byte("# Portfolio\n")); err != nil {
		t.Fatal(err)
	}

	// Place an image inside the section directory.
	sectionDir := filepath.Join(siteDir, "pages", "portfolio")
	imgData := []byte("fake-image-data")
	if err := os.WriteFile(filepath.Join(sectionDir, "photo.jpg"), imgData, 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(outDir, "portfolio", "photo.jpg"))
	if err != nil {
		t.Fatalf("expected dist/portfolio/photo.jpg to exist: %v", err)
	}
	if string(got) != string(imgData) {
		t.Errorf("dist/portfolio/photo.jpg content mismatch")
	}
}

func TestBuildSkipsDraftPageByDefault(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}
	if err := page.Create(siteDir, "wip", []byte("---\ndraft: true\n---\n# WIP\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "index.html")); err != nil {
		t.Error("expected index.html to exist")
	}
	if _, err := os.Stat(filepath.Join(outDir, "wip.html")); err == nil {
		t.Error("expected wip.html to be skipped (draft)")
	}
}

func TestBuildIncludesDraftPageWithFlag(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}
	if err := page.Create(siteDir, "wip", []byte("---\ndraft: true\n---\n# WIP\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, true); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "wip.html")); err != nil {
		t.Error("expected wip.html to be built when includeDrafts=true")
	}
}

func TestBuildDraftPageExcludedFromNavigation(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}
	if err := page.Create(siteDir, "wip", []byte("---\ndraft: true\n---\n# WIP\n")); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "wip.html") {
		t.Error("expected draft page wip.html to be absent from navigation")
	}
}

func TestBuildSkipsDraftSectionPageByDefault(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}
	if err := section.Create(siteDir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(siteDir, "pages", "blog")
	if err := os.WriteFile(filepath.Join(blogDir, "post.md"), []byte("# Post\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "wip.md"), []byte("---\ndraft: true\n---\n# WIP\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "blog", "post.html")); err != nil {
		t.Error("expected blog/post.html to exist")
	}
	if _, err := os.Stat(filepath.Join(outDir, "blog", "wip.html")); err == nil {
		t.Error("expected blog/wip.html to be skipped (draft)")
	}
}

func TestBuildDraftSectionPageExcludedFromTOC(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")

	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}
	if err := section.Create(siteDir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(siteDir, "pages", "blog")
	if err := os.WriteFile(filepath.Join(blogDir, "post.md"), []byte("# Post\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "wip.md"), []byte("---\ndraft: true\n---\n# WIP\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Build(siteDir, outDir, false); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outDir, "blog", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "wip.html") {
		t.Error("expected draft page wip.html to be absent from section TOC")
	}
}
