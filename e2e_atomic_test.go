package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2ERenameAndCreateAreSafe(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")

	// Hand-built: create refuses deep names, but legacy deep pages must stay renameable.
	deepDir := filepath.Join(siteDir, "pages", "blog", "2026")
	if err := os.MkdirAll(deepDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deepDir, "post.md"), []byte("---\ntitle: \"X\"\nupdated_at: \"2026-01-01T00:00:00Z\"\n---\nX"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, siteDir, "create", "page", "other")
	otherPath := filepath.Join(siteDir, "pages", "other.md")
	before := readFile(t, otherPath)

	// Renaming onto an existing page fails and leaves both files untouched.
	out := runExpectError(t, siteDir, "rename", "page", "blog/2026/post", "other")
	if !strings.Contains(out, "exists") {
		t.Errorf("expected 'exists' error, got: %s", out)
	}
	if readFile(t, otherPath) != before {
		t.Error("existing target page was modified")
	}
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "blog", "2026", "post.md")); err != nil {
		t.Errorf("source page should remain: %v", err)
	}

	// Creating over an existing page fails and keeps its content.
	runExpectError(t, siteDir, "create", "page", "other")
	if readFile(t, otherPath) != before {
		t.Error("create overwrote existing page")
	}

	// A successful nested rename removes the now-empty parent directories.
	run(t, siteDir, "rename", "page", "blog/2026/post", "moved")
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "moved.md")); err != nil {
		t.Fatalf("renamed page missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "blog")); err == nil {
		t.Error("empty parent directory left behind after rename")
	}
}
