package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2EOutputDirValidation(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")

	bad := [][]string{
		{"build", "--output", "."},
		{"build", "--output", ""},
		{"build", "--output", ".."},
		{"build", "--output", "pages"},
		{"build", "--output", "pages/out"},
		{"build", "--static", "dist"},
		{"build", "--output", "static"},
		{"build", "--output", "static/out"},
		{"build", "--output", "template.html"},
		{"serve", "--output", "pages"},
	}
	for _, args := range bad {
		out := runExpectError(t, siteDir, args...)
		if !strings.Contains(out, "invalid output directory") {
			t.Errorf("press %v: want \"invalid output directory\" error, got: %s", args, out)
		}
	}
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "index.html")); err == nil {
		t.Error("HTML was written into pages/")
	}

	run(t, siteDir, "build", "--output", "public")
	if _, err := os.Stat(filepath.Join(siteDir, "public", "index.html")); err != nil {
		t.Errorf("valid custom output should build: %v", err)
	}
}
