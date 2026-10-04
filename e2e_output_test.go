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

func TestE2EFailedBuildKeepsPreviousOutput(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	writeFile(t, filepath.Join(siteDir, "pages", "old.md"), "# Old\n")
	run(t, siteDir, "build")
	dist := filepath.Join(siteDir, "dist")
	before, err := os.ReadFile(filepath.Join(dist, "index.html"))
	if err != nil {
		t.Fatal(err)
	}

	// The template parses but fails when a page is rendered.
	tmplPath := filepath.Join(siteDir, "template.html")
	goodTmpl, err := os.ReadFile(tmplPath)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, tmplPath, `{{ template "missing" . }}`)
	if err := os.Remove(filepath.Join(siteDir, "pages", "old.md")); err != nil {
		t.Fatal(err)
	}
	runExpectError(t, siteDir, "build")

	after, err := os.ReadFile(filepath.Join(dist, "index.html"))
	if err != nil || string(after) != string(before) {
		t.Fatalf("dist/index.html must be unchanged after a failed build (err=%v)", err)
	}
	for _, name := range []string{"old.html", ".press-output"} {
		if _, err := os.Stat(filepath.Join(dist, name)); err != nil {
			t.Errorf("dist/%s must survive a failed build: %v", name, err)
		}
	}
	entries, err := os.ReadDir(siteDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".press-new") || strings.Contains(e.Name(), ".press-old") {
			t.Errorf("failed build left %s behind", e.Name())
		}
	}

	// Fixing the template builds again and drops the stale page.
	writeFile(t, tmplPath, string(goodTmpl))
	run(t, siteDir, "build")
	if _, err := os.Stat(filepath.Join(dist, "old.html")); err == nil {
		t.Error("dist/old.html should be gone after a successful rebuild")
	}
}
