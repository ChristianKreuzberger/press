package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ChristianKreuzberger/press/internal/page"
	"github.com/ChristianKreuzberger/press/internal/section"
)

// newAssetSite returns a site dir with an index page and the build output dir.
func newAssetSite(t *testing.T) (siteDir, outDir string) {
	t.Helper()
	siteDir = t.TempDir()
	outDir = filepath.Join(siteDir, "dist")
	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}
	return siteDir, outDir
}

func TestBuildSkipsSymlinkedAsset(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	secret := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(secret, []byte("s3cret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secret, filepath.Join(siteDir, "pages", "leak.txt")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(outDir, "leak.txt")); err == nil {
		t.Error("symlinked asset must not be copied to output")
	}
}

func TestBuildSkipsDotfileAssets(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	pagesDir := filepath.Join(siteDir, "pages")
	if err := os.WriteFile(filepath.Join(pagesDir, ".DS_Store"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(pagesDir, "img"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pagesDir, "img", ".hidden"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	for _, p := range []string{".DS_Store", filepath.Join("img", ".hidden")} {
		if _, err := os.Lstat(filepath.Join(outDir, p)); err == nil {
			t.Errorf("dotfile %s must not be copied", p)
		}
	}
}

func TestBuildSkipsUppercaseMarkdownAsset(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if err := os.WriteFile(filepath.Join(siteDir, "pages", "notes.MD"), []byte("# raw"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(outDir, "notes.MD")); err == nil {
		t.Error("notes.MD must not be copied as an asset")
	}
}

func TestBuildFailsOnAssetPageCollision(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if err := page.Create(siteDir, "about", []byte("# About\n")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(siteDir, "pages", "about.html"), []byte("<p>x</p>"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Build(siteDir, outDir, false, "static")
	if err == nil {
		t.Fatal("expected collision error")
	}
	if !strings.Contains(err.Error(), "about.html") {
		t.Errorf("error should name the colliding file, got: %v", err)
	}
}

func TestBuildFailsOnAssetSectionPageCollision(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if err := section.Create(siteDir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	blog := filepath.Join(siteDir, "pages", "blog")
	if err := os.WriteFile(filepath.Join(blog, "post.md"), []byte("# Post\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blog, "post.html"), []byte("<p>x</p>"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := Build(siteDir, outDir, false, "static"); err == nil {
		t.Fatal("expected collision error in section")
	}
}

func TestBuildFailsOnUppercaseHTMLAssetCollision(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if err := page.Create(siteDir, "about", []byte("# About\n")); err != nil {
		t.Fatal(err)
	}
	// On case-insensitive filesystems about.HTML is the same file as the built about.html.
	if err := os.WriteFile(filepath.Join(siteDir, "pages", "about.HTML"), []byte("<p>x</p>"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := Build(siteDir, outDir, false, "static"); err == nil {
		t.Fatal("expected collision error for about.HTML")
	}
}

func TestBuildSkipsHiddenDirectories(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	cache := filepath.Join(siteDir, "pages", ".cache")
	if err := os.MkdirAll(cache, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cache, "data.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(outDir, ".cache")); err == nil {
		t.Error("hidden directory .cache must not be copied")
	}
}
