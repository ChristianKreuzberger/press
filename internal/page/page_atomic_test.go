package page

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const fm = "---\ntitle: \"X\"\nupdated_at: \"2026-01-01T00:00:00Z\"\n---\nX"

func TestCreateDoesNotOverwriteExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(PagesDir(dir), "about.md")
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	err := Create(dir, "about", []byte("new"))
	if !errors.Is(err, ErrPageExists) {
		t.Fatalf("expected ErrPageExists, got %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "original" {
		t.Errorf("existing file was modified: %q", got)
	}
}

func TestCreateFailureLeavesNoEmptyDirs(t *testing.T) {
	dir := t.TempDir()
	// An over-long file name makes the write fail after the parents are created.
	name := "newsec/deeper/" + strings.Repeat("x", 300)
	if err := Create(dir, name, []byte("x")); err == nil {
		t.Fatal("expected error")
	}
	if _, err := os.Stat(filepath.Join(PagesDir(dir), "newsec")); err == nil {
		t.Error("empty parent directories left behind after failed create")
	}
}

func TestRenameDoesNotOverwriteTarget(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "a", []byte(fm)); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "b", []byte("B-original")); err != nil {
		t.Fatal(err)
	}
	err := Rename(dir, "a", "b", time.Now())
	if !errors.Is(err, ErrPageExists) {
		t.Fatalf("expected ErrPageExists, got %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(PagesDir(dir), "b.md"))
	if string(got) != "B-original" {
		t.Errorf("target was modified: %q", got)
	}
	if _, err := os.Stat(filepath.Join(PagesDir(dir), "a.md")); err != nil {
		t.Errorf("source should remain: %v", err)
	}
}

func TestRenamePreservesMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes are not meaningful on Windows")
	}
	dir := t.TempDir()
	if err := Create(dir, "a", []byte(fm)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(PagesDir(dir), "a.md"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Rename(dir, "a", "b", time.Now()); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(PagesDir(dir), "b.md"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("mode = %v, want 0600", info.Mode().Perm())
	}
}

func TestRenameCleansEmptyParents(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog/2026/post", []byte(fm)); err != nil {
		t.Fatal(err)
	}
	if err := Rename(dir, "blog/2026/post", "top", time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(PagesDir(dir), "blog")); err == nil {
		t.Error("empty parent dirs left after rename")
	}
	if _, err := os.Stat(PagesDir(dir)); err != nil {
		t.Errorf("pages dir itself must remain: %v", err)
	}
}

func TestRenameKeepsNonEmptyParents(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"blog/a", "blog/b"} {
		if err := Create(dir, n, []byte(fm)); err != nil {
			t.Fatal(err)
		}
	}
	if err := Rename(dir, "blog/a", "top", time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(PagesDir(dir), "blog", "b.md")); err != nil {
		t.Errorf("sibling lost: %v", err)
	}
}

func TestDeleteCleansEmptyParents(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog/2026/post", []byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := Delete(dir, "blog/2026/post"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(PagesDir(dir), "blog")); err == nil {
		t.Error("empty parent dirs left after delete")
	}
}
