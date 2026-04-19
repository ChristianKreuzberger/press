package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ChristianKreuzberger/press/internal/page"
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

	if !strings.Contains(html, "<h1>Home</h1>") {
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
