package main_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestE2EBuildEscapesSpecialCharsInLinks(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	run(t, siteDir, "create", "page", "my page#1")
	run(t, siteDir, "create", "section", "my sec#1")
	want := []string{`href="my%20page%231.html"`, `href="my%20sec%231/index.html"`}
	// "?" is not a valid file name character on Windows; the unit test covers it there.
	if runtime.GOOS != "windows" {
		run(t, siteDir, "create", "page", "what?")
		want = append(want, `href="what%3F.html"`)
	}

	run(t, siteDir, "build")

	data, err := os.ReadFile(filepath.Join(siteDir, "dist", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range want {
		if !strings.Contains(string(data), w) {
			t.Errorf("index.html missing %s", w)
		}
	}
	// The file on disk keeps the raw name; the browser decodes the link to it.
	if _, err := os.Stat(filepath.Join(siteDir, "dist", "my page#1.html")); err != nil {
		t.Errorf("expected output file: %v", err)
	}
}
