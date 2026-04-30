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

func TestCreateNestedSections(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog/2026/my-post", []byte("# My Post\n")); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(PagesDir(dir), "blog", "2026", "my-post.md")
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
