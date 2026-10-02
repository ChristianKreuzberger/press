package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2EBuildEscapesSpecialCharsInLinks(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	run(t, siteDir, "create", "page", "my page#1")
	run(t, siteDir, "create", "page", "what?")

	run(t, siteDir, "build")

	data, err := os.ReadFile(filepath.Join(siteDir, "dist", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := strings.ToLower(string(data))
	for _, want := range []string{`href="my%20page%231.html"`, `href="what%3f.html"`} {
		if !strings.Contains(html, want) {
			t.Errorf("index.html missing %s", want)
		}
	}
	// The file on disk keeps the raw name; the browser decodes the link to it.
	if _, err := os.Stat(filepath.Join(siteDir, "dist", "my page#1.html")); err != nil {
		t.Errorf("expected output file: %v", err)
	}
}
