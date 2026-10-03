package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ChristianKreuzberger/press/internal/page"
)

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err != nil {
		t.Errorf("%s must still exist: %v", path, err)
	}
}

func TestBuildRemovesStaleOutput(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if err := page.Create(siteDir, "old", []byte("# Old\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(siteDir, "pages", "old.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "old.html")); err == nil {
		t.Error("dist/old.html should be removed after the page was deleted")
	}
	if _, err := os.Stat(filepath.Join(outDir, "index.html")); err != nil {
		t.Error("dist/index.html should still exist")
	}
}

func TestBuildRefusesDangerousOutputDir(t *testing.T) {
	siteDir, _ := newAssetSite(t)
	if err := os.MkdirAll(filepath.Join(siteDir, "static"), 0755); err != nil {
		t.Fatal(err)
	}
	sibling := filepath.Join(filepath.Dir(siteDir), filepath.Base(siteDir)+"-sibling")
	t.Cleanup(func() { _ = os.RemoveAll(sibling) })
	writeTestFile(t, filepath.Join(siteDir, ".git", "HEAD"), "ref")
	writeTestFile(t, filepath.Join(siteDir, "docs", "keep.txt"), "keep")
	writeTestFile(t, filepath.Join(siteDir, "unrelated", "notes.txt"), "keep")
	writeTestFile(t, filepath.Join(sibling, "mine.txt"), "keep")
	writeTestFile(t, filepath.Join(filepath.Dir(siteDir), "next-to-site.txt"), "keep")

	// Symlinks to an unrelated non-empty dir and to the site's pages/.
	linkUnrelated := filepath.Join(siteDir, "link-unrelated")
	linkPages := filepath.Join(siteDir, "link-pages")
	if err := os.Symlink(sibling, linkUnrelated); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Join(siteDir, "pages"), linkPages); err != nil {
		t.Fatal(err)
	}

	// Source overlaps are caught by validateOutputDir; non-empty dirs that
	// press did not create are caught by the ownership check.
	cases := map[string]struct{ out, want string }{
		"site dir":              {siteDir, "invalid output directory"},
		"parent dir":            {filepath.Dir(siteDir), "invalid output directory"},
		"pages dir":             {filepath.Join(siteDir, "pages"), "invalid output directory"},
		"static dir":            {filepath.Join(siteDir, "static"), "invalid output directory"},
		"inside pages":          {filepath.Join(siteDir, "pages", "out"), "invalid output directory"},
		"template.html":         {filepath.Join(siteDir, "template.html"), "invalid output directory"},
		".git":                  {filepath.Join(siteDir, ".git"), "invalid output directory"},
		"symlink to pages":      {linkPages, "invalid output directory"},
		"new dir under symlink": {filepath.Join(linkPages, "newdir"), "invalid output directory"},
		"docs":                  {filepath.Join(siteDir, "docs"), "refusing to use output directory"},
		"non-empty unrelated":   {filepath.Join(siteDir, "unrelated"), "refusing to use output directory"},
		"sibling outside site":  {sibling, "refusing to use output directory"},
		"symlink to unrelated":  {linkUnrelated, "refusing to use output directory"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Build(siteDir, c.out, false, "static")
			if err == nil {
				t.Fatalf("Build with output %s should fail", c.out)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("expected %q, got: %v", c.want, err)
			}
		})
	}

	for _, p := range []string{
		filepath.Join(siteDir, "pages", "index.md"),
		filepath.Join(siteDir, ".git", "HEAD"),
		filepath.Join(siteDir, "docs", "keep.txt"),
		filepath.Join(siteDir, "unrelated", "notes.txt"),
		filepath.Join(sibling, "mine.txt"),
		filepath.Join(filepath.Dir(siteDir), "next-to-site.txt"),
	} {
		assertExists(t, p)
	}
	// Validation runs before any directory is created.
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "out")); err == nil {
		t.Error("a refused build must not leave pages/out behind")
	}
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "newdir")); err == nil {
		t.Error("a refused build must not create dirs through a symlink")
	}
}

func TestBuildWritesMarkerAndRebuilds(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	for i := 0; i < 2; i++ {
		if _, err := Build(siteDir, outDir, false, "static"); err != nil {
			t.Fatalf("build %d: %v", i+1, err)
		}
		assertExists(t, filepath.Join(outDir, outputMarker))
	}
}

func TestBuildCleansMarkedDirWithAnyContent(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(outDir, "old", "thing.pdf"), "x")
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "old")); err == nil {
		t.Error("stale content in a marked output dir should be removed")
	}
}

// Output dirs created by older press versions have no marker; upgrading must
// not require manual cleanup.
func TestBuildAcceptsLegacyUnmarkedOutput(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	writeTestFile(t, filepath.Join(siteDir, "static", "css", "site.css"), "a{}")
	writeTestFile(t, filepath.Join(outDir, "index.html"), "<html>")
	writeTestFile(t, filepath.Join(outDir, "gone.html"), "<html>")
	writeTestFile(t, filepath.Join(outDir, "blog", "post.html"), "<html>")
	writeTestFile(t, filepath.Join(outDir, "css", "site.css"), "a{}")

	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatalf("legacy dist should still build: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "gone.html")); err == nil {
		t.Error("stale file in legacy dist should be removed")
	}
	assertExists(t, filepath.Join(outDir, outputMarker))
}

func TestBuildRefusesUnmarkedDirWithForeignFiles(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	writeTestFile(t, filepath.Join(outDir, "index.html"), "<html>")
	writeTestFile(t, filepath.Join(outDir, "keep.txt"), "mine")
	if _, err := Build(siteDir, outDir, false, "static"); err == nil ||
		!strings.Contains(err.Error(), "refusing to use output directory") {
		t.Fatalf("expected refusal, got %v", err)
	}
	assertExists(t, filepath.Join(outDir, "keep.txt"))
}

func TestBuildBadTemplateKeepsExistingOutput(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(siteDir, "template.html"), "{{ .Broken")
	if _, err := Build(siteDir, outDir, false, "static"); err == nil {
		t.Fatal("expected template parse error")
	}
	assertExists(t, filepath.Join(outDir, "index.html"))
}
