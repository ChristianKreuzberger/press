package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ChristianKreuzberger/press/internal/page"
)

func exists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}

func write(t *testing.T, p, s string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBuildRemovesStaleOutput(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if err := page.Create(siteDir, "about", []byte("# About\n")); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(siteDir, "pages", "blog", "index.md"), "# Blog\n")
	write(t, filepath.Join(siteDir, "pages", "blog", "post.md"), "# Post\n")
	write(t, filepath.Join(siteDir, "pages", "img", "a.png"), "png")
	write(t, filepath.Join(siteDir, "static", "x", "s.css"), "css")
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	// A user-owned file that press did not generate.
	write(t, filepath.Join(outDir, "CNAME"), "example.com")

	_ = os.Remove(filepath.Join(siteDir, "pages", "about.md"))
	_ = os.Remove(filepath.Join(siteDir, "pages", "blog", "post.md"))
	_ = os.Remove(filepath.Join(siteDir, "pages", "img", "a.png"))
	_ = os.RemoveAll(filepath.Join(siteDir, "static"))
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}

	for _, gone := range []string{"about.html", "blog/post.html", "img/a.png", "img", "static"} {
		if exists(filepath.Join(outDir, gone)) {
			t.Errorf("%s should have been removed", gone)
		}
	}
	for _, kept := range []string{"index.html", "blog/index.html", "CNAME"} {
		if !exists(filepath.Join(outDir, kept)) {
			t.Errorf("%s should have been kept", kept)
		}
	}
}

func TestBuildRemovesNewlyDraftedPage(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if err := page.Create(siteDir, "wip", []byte("# WIP\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(siteDir, "pages", "wip.md"), "---\ndraft: true\n---\n# WIP\n")
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	if exists(filepath.Join(outDir, "wip.html")) {
		t.Error("draft page output should be removed")
	}
}

func TestBuildFirstRunWithoutManifestRemovesNothing(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	write(t, filepath.Join(outDir, "old.html"), "old")
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(outDir, "old.html")) {
		t.Error("files not recorded by press must never be removed")
	}
}

func TestBuildIgnoresUnsafeManifestEntries(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	victim := filepath.Join(siteDir, "victim.txt")
	write(t, victim, "keep")
	write(t, filepath.Join(outDir, "dir", "f.txt"), "keep")
	write(t, filepath.Join(outDir, manifestName), "../victim.txt\n"+victim+"\n..\n.\n\ndir\n")
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	if !exists(victim) || !exists(filepath.Join(outDir, "dir", "f.txt")) {
		t.Error("unsafe manifest entries must be ignored")
	}
}

func TestBuildDoesNotFollowSymlinkedDirInManifest(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	outside := t.TempDir()
	write(t, filepath.Join(outside, "f.txt"), "keep")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(outDir, "link")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	write(t, filepath.Join(outDir, manifestName), "link/f.txt\n")
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(outside, "f.txt")) {
		t.Error("cleanup must not follow symlinked directories out of the output dir")
	}
}

func TestBuildNoCleanupWhenOutputContainsSite(t *testing.T) {
	siteDir, _ := newAssetSite(t)
	// Misconfigured: output is the site dir itself. Stale entries pointing at
	// source files must not be deleted.
	write(t, filepath.Join(siteDir, manifestName), "pages/index.md\ntemplate.html\n")
	write(t, filepath.Join(siteDir, "template.html"), "{{.Content}}")
	if _, err := Build(siteDir, siteDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(siteDir, "pages", "index.md")) || !exists(filepath.Join(siteDir, "template.html")) {
		t.Error("site sources must never be removed")
	}
}

func TestBuildNoCleanupWhenOutputInsidePages(t *testing.T) {
	siteDir, _ := newAssetSite(t)
	out := filepath.Join(siteDir, "pages", "out")
	write(t, filepath.Join(out, manifestName), "../index.md\n")
	if _, err := Build(siteDir, out, false, "static"); err != nil {
		t.Fatal(err)
	}
	if !exists(filepath.Join(siteDir, "pages", "index.md")) {
		t.Error("site sources must never be removed")
	}
}

func TestBuildKeepsOldManifestOnFailure(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(filepath.Join(outDir, manifestName))
	write(t, filepath.Join(siteDir, "template.html"), "{{ .Broken")
	if _, err := Build(siteDir, outDir, false, "static"); err == nil {
		t.Fatal("expected failure")
	}
	after, _ := os.ReadFile(filepath.Join(outDir, manifestName))
	if string(before) != string(after) || len(before) == 0 {
		t.Error("manifest must be unchanged after a failed build")
	}
}
