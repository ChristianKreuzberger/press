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
	files := map[string]bool{ // path -> expected to be reported
		"index.md":          false,
		"about.md":          false,
		"blog/index.md":     false,
		"blog/post.md":      false,
		"blog/2026/post.md": true,
		"noindex/post.md":   true,
		".hidden/post.md":   false,
		"blog/image.png":    false,
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
	got, err := Skipped(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %v, want blog/2026/post.md and noindex/post.md", got)
	}
	for _, g := range got {
		if !files[g] {
			t.Errorf("unexpected skip report %q", g)
		}
	}
	if got, err := Skipped(t.TempDir()); err != nil || len(got) != 0 {
		t.Errorf("missing pages dir: got %v, %v", got, err)
	}
}
