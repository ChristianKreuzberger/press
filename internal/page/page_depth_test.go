package page

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateRejectsTooDeepNames(t *testing.T) {
	dir := t.TempDir()
	err := Create(dir, "blog/2026/my-post", []byte("x"))
	if !errors.Is(err, ErrInvalidName) {
		t.Fatalf("got %v, want ErrInvalidName", err)
	}
	if _, err := os.Stat(filepath.Join(PagesDir(dir), "blog")); err == nil {
		t.Error("rejected create must not leave directories behind")
	}
}

func TestRenameRejectsTooDeepTarget(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "a", []byte(fm)); err != nil {
		t.Fatal(err)
	}
	err := Rename(dir, "a", "blog/2026/a", time.Now())
	if !errors.Is(err, ErrInvalidName) {
		t.Fatalf("got %v, want ErrInvalidName", err)
	}
	if _, err := os.Stat(filepath.Join(PagesDir(dir), "a.md")); err != nil {
		t.Errorf("source page must remain: %v", err)
	}
}

func TestSkipped(t *testing.T) {
	dir := t.TempDir()
	files := map[string]Reason{ // path -> expected reason ("" = not reported)
		"index.md":           "",
		"about.md":           "",
		"About.MD":           ReasonExtension,
		"blog/index.md":      "",
		"blog/post.md":       "",
		"blog/Post.Md":       ReasonExtension,
		"blog/2026/post.md":  ReasonTooDeep,
		"blog/2026/img.png":  "",
		"blog/image.png":     "",
		"noindex/post.md":    ReasonNoIndex,
		".hidden/post.md":    "",
		"a/b/c/d.md":         ReasonTooDeep,
		"noindex/x/deep.md":  ReasonTooDeep,
		"assets/README.md":   ReasonNoIndex,
		"assets/logo.svg":    "",
		"onlyassets/pic.png": "",
	}
	for rel := range files {
		p := filepath.Join(PagesDir(dir), filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// A symlinked .md file is built (the builder only checks IsDir), so it is
	// not reported at the top level or in a section; a symlinked directory is
	// ignored by the builder and not scanned here.
	pages := PagesDir(dir)
	if err := os.Symlink(filepath.Join(pages, "about.md"), filepath.Join(pages, "link.md")); err != nil {
		t.Skip("symlinks unsupported:", err)
	}
	if err := os.Symlink(filepath.Join(pages, "about.md"), filepath.Join(pages, "blog", "link.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(pages, "noindex"), filepath.Join(pages, "linkdir")); err != nil {
		t.Fatal(err)
	}

	got, err := Skipped(dir)
	if err != nil {
		t.Fatal(err)
	}
	reported := map[string]Reason{}
	for _, g := range got {
		reported[g.Path] = g.Reason
	}
	for rel, want := range files {
		if reported[rel] != want {
			t.Errorf("%s: got reason %q, want %q", rel, reported[rel], want)
		}
	}
	for rel := range reported {
		if _, ok := files[rel]; !ok {
			t.Errorf("unexpected skip report %q", rel)
		}
	}
	if got, err := Skipped(t.TempDir()); err != nil || len(got) != 0 {
		t.Errorf("missing pages dir: got %v, %v", got, err)
	}
}

// Create allows a/b even when section a has no index.md; the build then
// ignores it, which Skipped must report.
func TestSkippedReportsPageInDirWithoutIndex(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "a/b", []byte("x")); err != nil {
		t.Fatal(err)
	}
	got, err := Skipped(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Path != "a/b.md" || got[0].Reason != ReasonNoIndex {
		t.Errorf("got %v", got)
	}
}
