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

	if err := Build(siteDir, outDir); err != nil {
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

	if err := Build(siteDir, outDir); err != nil {
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

	if err := Build(siteDir, outDir); err != nil {
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

	if err := Build(siteDir, outDir); err != nil {
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

	if err := Build(siteDir, outDir); err != nil {
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

	if err := Build(siteDir, outDir); err != nil {
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

	if err := Build(siteDir, outDir); err != nil {
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

	if err := Build(siteDir, outDir); err != nil {
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

	if err := Build(siteDir, outDir); err != nil {
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

	if err := Build(siteDir, outDir); err != nil {
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
