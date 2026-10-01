package section

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateDoesNotTouchExistingDir(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(sectionsBaseDir(dir), "blog")
	if err := os.MkdirAll(existing, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(existing, "index.md"), []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	err := Create(dir, "blog", []byte("new"))
	if !errors.Is(err, ErrSectionExists) {
		t.Fatalf("expected ErrSectionExists, got %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(existing, "index.md"))
	if string(got) != "original" {
		t.Errorf("existing index.md modified: %q", got)
	}
}

func TestRenameDoesNotOverwriteExistingEmptyDir(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "a", []byte("---\ntitle: \"A\"\n---\nA")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(sectionsBaseDir(dir), "b"), 0750); err != nil {
		t.Fatal(err)
	}
	err := Rename(dir, "a", "b", time.Now())
	if !errors.Is(err, ErrSectionExists) {
		t.Fatalf("expected ErrSectionExists, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(sectionsBaseDir(dir), "a", "index.md")); err != nil {
		t.Errorf("source should remain: %v", err)
	}
}
