package main_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestE2EBuildCleansStaleOutput(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	run(t, siteDir, "create", "page", "gone")
	run(t, siteDir, "create", "page", "wip")
	run(t, siteDir, "build")
	dist := filepath.Join(siteDir, "dist")
	for _, f := range []string{"gone.html", "wip.html"} {
		if _, err := os.Stat(filepath.Join(dist, f)); err != nil {
			t.Fatalf("%s should exist after first build: %v", f, err)
		}
	}
	writeFile(t, filepath.Join(dist, "CNAME"), "example.com")

	if err := os.Remove(filepath.Join(siteDir, "pages", "gone.md")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(siteDir, "pages", "wip.md"), "---\ndraft: true\n---\n# WIP\n")
	run(t, siteDir, "build")

	for _, f := range []string{"gone.html", "wip.html"} {
		if _, err := os.Stat(filepath.Join(dist, f)); err == nil {
			t.Errorf("%s should have been removed", f)
		}
	}
	for _, f := range []string{"index.html", "CNAME"} {
		if _, err := os.Stat(filepath.Join(dist, f)); err != nil {
			t.Errorf("%s should remain: %v", f, err)
		}
	}
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "wip.md")); err != nil {
		t.Errorf("source page must remain: %v", err)
	}
}

func TestE2EBuildOutputMisconfiguredKeepsSources(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	run(t, siteDir, "build", "-output", ".")
	run(t, siteDir, "build", "-output", ".")
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "index.md")); err != nil {
		t.Errorf("source must remain: %v", err)
	}
}
