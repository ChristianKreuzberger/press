package page

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
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
