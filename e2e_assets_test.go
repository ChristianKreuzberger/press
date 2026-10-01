package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2EAssetCopySafety(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	pagesDir := filepath.Join(siteDir, "pages")

	secret := filepath.Join(t.TempDir(), "secret")
	writeFile(t, secret, "s3cret")
	// Only the symlink check depends on symlink support; the rest still runs.
	symlinked := os.Symlink(secret, filepath.Join(pagesDir, "leak.txt")) == nil
	writeFile(t, filepath.Join(pagesDir, ".DS_Store"), "x")
	if err := os.MkdirAll(filepath.Join(pagesDir, ".cache"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(pagesDir, ".cache", "data.txt"), "x")
	writeFile(t, filepath.Join(pagesDir, "logo.svg"), "<svg/>")

	run(t, siteDir, "build")
	dist := filepath.Join(siteDir, "dist")
	skipped := []string{".DS_Store", ".cache"}
	if symlinked {
		skipped = append(skipped, "leak.txt")
	}
	for _, name := range skipped {
		if _, err := os.Lstat(filepath.Join(dist, name)); err == nil {
			t.Errorf("dist/%s must not be copied", name)
		}
	}
	if _, err := os.Stat(filepath.Join(dist, "logo.svg")); err != nil {
		t.Error("regular asset dist/logo.svg should be copied")
	}

	// An asset that would overwrite a built page fails the build.
	writeFile(t, filepath.Join(pagesDir, "about.md"), "# About\n")
	writeFile(t, filepath.Join(pagesDir, "about.html"), "<p>x</p>")
	out := runExpectError(t, siteDir, "build")
	if !strings.Contains(out, "about.html") {
		t.Errorf("error should mention about.html, got: %s", out)
	}
}
