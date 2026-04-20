package section

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListEmpty(t *testing.T) {
	dir := t.TempDir()
	sections, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 0 {
		t.Errorf("expected 0 sections, got %d", len(sections))
	}
}

func TestListNoPages(t *testing.T) {
	// pages/ directory does not exist yet
	dir := t.TempDir()
	sections, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if sections != nil {
		t.Errorf("expected nil when pages/ absent, got %v", sections)
	}
}

func TestListSkipsDirectoriesWithoutIndex(t *testing.T) {
	dir := t.TempDir()
	// Create a subdirectory without index.md — should not appear as section.
	noIndex := filepath.Join(dir, "pages", "noindex")
	if err := os.MkdirAll(noIndex, 0755); err != nil {
		t.Fatal(err)
	}

	sections, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 0 {
		t.Errorf("directory without index.md should not be listed as section, got %v", sections)
	}
}

func TestCreateAndList(t *testing.T) {
	dir := t.TempDir()

	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "docs", []byte("# Docs\n")); err != nil {
		t.Fatal(err)
	}

	sections, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	names := map[string]bool{}
	for _, s := range sections {
		names[s.Name] = true
	}
	if !names["blog"] || !names["docs"] {
		t.Errorf("unexpected section names: %v", names)
	}
}

func TestCreateDuplicate(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "blog", []byte("dup")); err == nil {
		t.Error("expected error creating duplicate section, got nil")
	}
}

func TestCreateWritesIndexMD(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	indexPath := filepath.Join(dir, "pages", "blog", "index.md")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("expected index.md to exist: %v", err)
	}
	if string(content) != "# Blog\n" {
		t.Errorf("unexpected index.md content: %q", content)
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	if err := Delete(dir, "blog"); err != nil {
		t.Fatal(err)
	}
	sections, _ := List(dir)
	if len(sections) != 0 {
		t.Errorf("expected 0 sections after delete, got %d", len(sections))
	}
}

func TestDeleteNotFound(t *testing.T) {
	dir := t.TempDir()
	if err := Delete(dir, "missing"); err == nil {
		t.Error("expected error deleting non-existent section, got nil")
	}
}

func TestUpdate(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Old\n")); err != nil {
		t.Fatal(err)
	}
	if err := Update(dir, "blog", []byte("# New\n")); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(filepath.Join(dir, "pages", "blog", "index.md"))
	if string(content) != "# New\n" {
		t.Errorf("update did not change content: %q", content)
	}
}

func TestUpdateNotFound(t *testing.T) {
	dir := t.TempDir()
	if err := Update(dir, "missing", []byte("x")); err == nil {
		t.Error("expected error updating non-existent section, got nil")
	}
}

func TestListPages(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	// Add extra pages to the section directory.
	blogDir := filepath.Join(dir, "pages", "blog")
	if err := os.WriteFile(filepath.Join(blogDir, "post-one.md"), []byte("# Post One\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "post-two.md"), []byte("# Post Two\n"), 0644); err != nil {
		t.Fatal(err)
	}

	pages, err := ListPages(dir, "blog")
	if err != nil {
		t.Fatal(err)
	}
	// Should include index, post-one, post-two.
	if len(pages) != 3 {
		t.Fatalf("expected 3 pages, got %d", len(pages))
	}
	names := map[string]bool{}
	for _, p := range pages {
		names[p.Name] = true
	}
	if !names["index"] || !names["post-one"] || !names["post-two"] {
		t.Errorf("unexpected page names: %v", names)
	}
}

func TestListPagesNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := ListPages(dir, "missing")
	if err == nil {
		t.Error("expected error listing pages of non-existent section, got nil")
	}
}
