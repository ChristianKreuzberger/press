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

func TestE2EBuildRemovesStaleOutput(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	pagesDir := filepath.Join(siteDir, "pages")
	dist := filepath.Join(siteDir, "dist")

	writeFile(t, filepath.Join(pagesDir, "gone.md"), "# Gone\n")
	writeFile(t, filepath.Join(pagesDir, "secret.md"), "# Secret\n")
	run(t, siteDir, "build")
	for _, name := range []string{"gone.html", "secret.html"} {
		if _, err := os.Stat(filepath.Join(dist, name)); err != nil {
			t.Fatalf("first build should create dist/%s", name)
		}
	}

	// One page is deleted, the other becomes a draft.
	if err := os.Remove(filepath.Join(pagesDir, "gone.md")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(pagesDir, "secret.md"), "---\ndraft: true\n---\n# Secret\n")
	run(t, siteDir, "build")

	for _, name := range []string{"gone.html", "secret.html"} {
		if _, err := os.Stat(filepath.Join(dist, name)); err == nil {
			t.Errorf("dist/%s is stale and must be removed", name)
		}
	}
	if _, err := os.Stat(filepath.Join(dist, "index.html")); err != nil {
		t.Error("dist/index.html should still be built")
	}

	// Pointing -output at the site itself must not wipe the sources.
	out := runExpectError(t, siteDir, "build", "-output", ".")
	if !strings.Contains(out, "refusing to use output directory") {
		t.Errorf("error should come from the output guard, got: %s", out)
	}
	if _, err := os.Stat(filepath.Join(pagesDir, "index.md")); err != nil {
		t.Error("pages/index.md must survive a refused build")
	}
}
