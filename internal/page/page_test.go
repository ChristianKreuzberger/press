package page

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestListEmpty(t *testing.T) {
	dir := t.TempDir()
	pages, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 0 {
		t.Errorf("expected 0 pages, got %d", len(pages))
	}
}

func TestCreateAndList(t *testing.T) {
	dir := t.TempDir()

	if err := Create(dir, "index", []byte("# Index\n")); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "about", []byte("# About\n")); err != nil {
		t.Fatal(err)
	}

	pages, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(pages))
	}

	names := map[string]bool{}
	for _, p := range pages {
		names[p.Name] = true
	}
	if !names["index"] || !names["about"] {
		t.Errorf("unexpected page names: %v", names)
	}
}

func TestCreateDuplicate(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "index", []byte("# Index\n")); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "index", []byte("dup")); err == nil {
		t.Error("expected error creating duplicate page, got nil")
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "index", []byte("# Index\n")); err != nil {
		t.Fatal(err)
	}
	if err := Delete(dir, "index"); err != nil {
		t.Fatal(err)
	}
	pages, _ := List(dir)
	if len(pages) != 0 {
		t.Errorf("expected 0 pages after delete, got %d", len(pages))
	}
}

func TestDeleteNotFound(t *testing.T) {
	dir := t.TempDir()
	if err := Delete(dir, "missing"); err == nil {
		t.Error("expected error deleting non-existent page, got nil")
	}
}

func TestCreateInSection(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog/my-post", []byte("# My Post\n")); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(PagesDir(dir), "blog", "my-post.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file at %s: %v", path, err)
	}
}

func TestCreatePathTraversal(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "../../etc/passwd", []byte("evil")); err == nil {
		t.Error("expected error for path traversal name, got nil")
	}
}

func TestUpdate(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "index", []byte("# Old\n")); err != nil {
		t.Fatal(err)
	}
	if err := Update(dir, "index", []byte("# New\n")); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(filepath.Join(PagesDir(dir), "index.md"))
	if string(content) != "# New\n" {
		t.Errorf("update did not change content: %q", content)
	}
}

func TestUpdateNotFound(t *testing.T) {
	dir := t.TempDir()
	if err := Update(dir, "missing", []byte("x")); err == nil {
		t.Error("expected error updating non-existent page, got nil")
	}
}

func TestUpdatePathTraversal(t *testing.T) {
	dir := t.TempDir()
	err := Update(dir, "../../etc/passwd", []byte("evil"))
	if !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName for path traversal, got %v", err)
	}
}

func TestDeletePathTraversal(t *testing.T) {
	dir := t.TempDir()
	err := Delete(dir, "../../etc/passwd")
	if !errors.Is(err, ErrInvalidName) {
		t.Errorf("expected ErrInvalidName for path traversal, got %v", err)
	}
}

func TestListDraftField(t *testing.T) {
	dir := t.TempDir()

	if err := Create(dir, "normal", []byte("# Normal\n")); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "draft-page", []byte("---\ndraft: true\n---\n# Draft\n")); err != nil {
		t.Fatal(err)
	}

	pages, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(pages))
	}
	byName := map[string]Page{}
	for _, p := range pages {
		byName[p.Name] = p
	}
	if byName["normal"].Draft {
		t.Error("expected normal page to have Draft=false")
	}
	if !byName["draft-page"].Draft {
		t.Error("expected draft-page to have Draft=true")
	}
}

func TestRename(t *testing.T) {
	dir := t.TempDir()
	content := []byte("---\ntitle: \"About\"\nalias: \"\"\ntags: []\nweight: 0\ncreated_at: \"2026-01-01T00:00:00Z\"\nupdated_at: \"2026-01-01T00:00:00Z\"\n---\n# About\n")
	if err := Create(dir, "about", content); err != nil {
		t.Fatal(err)
	}
	newNow := time.Date(2027, 6, 15, 12, 0, 0, 0, time.UTC)
	if err := Rename(dir, "about", "about-us", newNow); err != nil {
		t.Fatalf("Rename() error: %v", err)
	}
	// Old file should be gone.
	if _, err := os.Stat(filepath.Join(PagesDir(dir), "about.md")); err == nil {
		t.Error("expected old file to be removed")
	}
	// New file should exist with updated frontmatter.
	newContent, err := os.ReadFile(filepath.Join(PagesDir(dir), "about-us.md"))
	if err != nil {
		t.Fatalf("expected new file to exist: %v", err)
	}
	s := string(newContent)
	if !strings.Contains(s, `title: "About Us"`) {
		t.Errorf("expected updated title in frontmatter, got: %s", s)
	}
	if !strings.Contains(s, newNow.UTC().Format(time.RFC3339)) {
		t.Errorf("expected updated_at = %s in frontmatter, got: %s", newNow.UTC().Format(time.RFC3339), s)
	}
}

func TestRenameNotFound(t *testing.T) {
	dir := t.TempDir()
	err := Rename(dir, "missing", "new-name", time.Now())
	if !errors.Is(err, ErrPageNotFound) {
		t.Errorf("expected ErrPageNotFound, got %v", err)
	}
}

func TestRenameTargetExists(t *testing.T) {
	dir := t.TempDir()
	content := []byte("---\ntitle: \"About\"\nalias: \"\"\ntags: []\nweight: 0\ncreated_at: \"2026-01-01T00:00:00Z\"\nupdated_at: \"2026-01-01T00:00:00Z\"\n---\n# About\n")
	if err := Create(dir, "about", content); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "contact", content); err != nil {
		t.Fatal(err)
	}
	err := Rename(dir, "about", "contact", time.Now())
	if !errors.Is(err, ErrPageExists) {
		t.Errorf("expected ErrPageExists, got %v", err)
	}
}

func TestRenameMinimalFrontmatter(t *testing.T) {
	now := time.Date(2027, 1, 2, 3, 4, 5, 0, time.UTC)
	for name, in := range map[string]string{
		"no frontmatter":    "# About\n\nHi\n",
		"title only":        "---\ntitle: \"about\"\n---\nHi\n",
		"no title":          "---\ntags: []\n---\nHi\n",
		"no updated_at":     "---\ntitle: \"About\"\nweight: 1\n---\nHi\n",
		"empty frontmatter": "---\n---\nHi\n",
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if err := Create(dir, "about", []byte(in)); err != nil {
				t.Fatal(err)
			}
			if err := Rename(dir, "about", "about-us", now); err != nil {
				t.Fatalf("Rename() error: %v", err)
			}
			b, err := os.ReadFile(filepath.Join(dir, "pages", "about-us.md"))
			if err != nil {
				t.Fatal(err)
			}
			s := string(b)
			if !strings.Contains(s, `title: "About Us"`) || !strings.Contains(s, `updated_at: "2027-01-02T03:04:05Z"`) || !strings.Contains(s, "Hi\n") {
				t.Errorf("unexpected content: %q", s)
			}
		})
	}
}

func TestRenameKeepsCustomTitle(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "about", []byte("---\ntitle: \"Our Story\"\nupdated_at: \"x\"\n---\nHi\n")); err != nil {
		t.Fatal(err)
	}
	if err := Rename(dir, "about", "about-us", time.Now()); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "pages", "about-us.md"))
	if !strings.Contains(string(b), `title: "Our Story"`) {
		t.Errorf("custom title should be kept, got: %s", b)
	}
}

func TestRenameNestedTitleUsesBaseName(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog/a", []byte("---\ntitle: \"a\"\n---\nHi\n")); err != nil {
		t.Fatal(err)
	}
	if err := Rename(dir, "blog/a", "blog/my-post", time.Now()); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "pages", "blog", "my-post.md"))
	if !strings.Contains(string(b), `title: "My Post"`) {
		t.Errorf("title should come from the base name, got: %s", b)
	}
}

func skipIfRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("permission errors cannot be provoked as root")
	}
}

func TestRenameInvalidNames(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "about", []byte("# About\n")); err != nil {
		t.Fatal(err)
	}
	tests := []struct{ name, oldName, newName string }{
		{"invalid old name", "../evil", "ok"},
		{"invalid new name", "about", "../evil"},
		{"new name too deep", "about", "a/b/c"},
		{"index cannot be renamed", "index", "home"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Rename(dir, tt.oldName, tt.newName, time.Now())
			if !errors.Is(err, ErrInvalidName) {
				t.Errorf("expected ErrInvalidName, got %v", err)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(PagesDir(dir), "about.md")); err != nil {
		t.Errorf("source must be untouched: %v", err)
	}
}

func TestRenameNameConflict(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "about", []byte("# About\n")); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "blog/post", []byte("# Post\n")); err != nil {
		t.Fatal(err)
	}
	// "blog" is already a section directory.
	if err := Rename(dir, "about", "blog", time.Now()); !errors.Is(err, ErrNameConflict) {
		t.Errorf("section clash: expected ErrNameConflict, got %v", err)
	}
	// "about" is a page, so it cannot also be a section.
	if err := Rename(dir, "blog/post", "about/post", time.Now()); !errors.Is(err, ErrNameConflict) {
		t.Errorf("page clash: expected ErrNameConflict, got %v", err)
	}
}

func TestRenameUnreadableSource(t *testing.T) {
	skipIfRoot(t)
	dir := t.TempDir()
	if err := Create(dir, "about", []byte("# About\n")); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(PagesDir(dir), "about.md")
	if err := os.Chmod(src, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(src, 0644) })
	if err := Rename(dir, "about", "about-us", time.Now()); err == nil {
		t.Fatal("expected an error for an unreadable source")
	}
	if _, err := os.Stat(filepath.Join(PagesDir(dir), "about-us.md")); err == nil {
		t.Error("target must not be created")
	}
}

func TestRenameReadOnlyPagesDir(t *testing.T) {
	skipIfRoot(t)
	dir := t.TempDir()
	if err := Create(dir, "about", []byte("# About\n")); err != nil {
		t.Fatal(err)
	}
	pages := PagesDir(dir)
	if err := os.Chmod(pages, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(pages, 0755) })
	if err := Rename(dir, "about", "about-us", time.Now()); err == nil {
		t.Fatal("expected an error when pages/ is read-only")
	}
	if _, err := os.Stat(filepath.Join(pages, "about.md")); err != nil {
		t.Errorf("source must survive a failed rename: %v", err)
	}
}

func TestRenameRollsBackWhenSourceCannotBeRemoved(t *testing.T) {
	skipIfRoot(t)
	dir := t.TempDir()
	if err := Create(dir, "blog/post", []byte("# Post\n")); err != nil {
		t.Fatal(err)
	}
	srcDir := filepath.Join(PagesDir(dir), "blog")
	if err := os.Chmod(srcDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(srcDir, 0755) })
	if err := Rename(dir, "blog/post", "news/post", time.Now()); err == nil {
		t.Fatal("expected an error when the source cannot be removed")
	}
	if _, err := os.Stat(filepath.Join(srcDir, "post.md")); err != nil {
		t.Errorf("source must survive: %v", err)
	}
	if _, err := os.Stat(filepath.Join(PagesDir(dir), "news")); err == nil {
		t.Error("half-written target and its new parent must be rolled back")
	}
}
