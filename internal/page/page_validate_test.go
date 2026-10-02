package page

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var badNames = []string{"", "blog/", "/tmp/zzz", ".", "..", "a/./b", "a//b", "a/../b", `a\b`, "a\x00b", "../x", "blog/.."}

func TestInvalidNamesRejectedEverywhere(t *testing.T) {
	for _, name := range badNames {
		dir := t.TempDir()
		if err := Create(dir, name, nil); !errors.Is(err, ErrInvalidName) {
			t.Errorf("Create(%q) = %v, want ErrInvalidName", name, err)
		}
		if err := Update(dir, name, nil); !errors.Is(err, ErrInvalidName) {
			t.Errorf("Update(%q) = %v, want ErrInvalidName", name, err)
		}
		if err := Delete(dir, name); !errors.Is(err, ErrInvalidName) {
			t.Errorf("Delete(%q) = %v, want ErrInvalidName", name, err)
		}
		if err := Rename(dir, name, "x", time.Now()); !errors.Is(err, ErrInvalidName) {
			t.Errorf("Rename(%q, x) = %v, want ErrInvalidName", name, err)
		}
		if err := Rename(dir, "x", name, time.Now()); !errors.Is(err, ErrInvalidName) {
			t.Errorf("Rename(x, %q) = %v, want ErrInvalidName", name, err)
		}
		if _, err := os.Stat(filepath.Join(dir, "pages")); err == nil {
			t.Errorf("name %q created pages dir", name)
		}
	}
}

func TestCreateCollidesWithSection(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "pages", "blog"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "blog", nil); !errors.Is(err, ErrNameConflict) {
		t.Errorf("got %v, want ErrNameConflict", err)
	}
}

func TestCreateInsidePageFileConflicts(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "about", nil); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "about/team", nil); !errors.Is(err, ErrNameConflict) {
		t.Errorf("got %v, want ErrNameConflict", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "pages", "about")); err == nil {
		t.Error("section directory created next to about.md")
	}
}

func TestRenameCollisions(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "about-us", []byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "pages", "blog"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Rename(dir, "about-us", "blog", time.Now()); !errors.Is(err, ErrNameConflict) {
		t.Errorf("got %v, want ErrNameConflict", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "pages", "about-us.md")); err != nil {
		t.Errorf("source lost: %v", err)
	}
}

func TestRenameIndexRejected(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "index", []byte("x")); err != nil {
		t.Fatal(err)
	}
	if err := Rename(dir, "index", "home", time.Now()); !errors.Is(err, ErrInvalidName) {
		t.Errorf("got %v, want ErrInvalidName", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "pages", "index.md")); err != nil {
		t.Errorf("index.md lost: %v", err)
	}
}

func TestCreateSectionPageAndSectionIndex(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog/my-post", nil); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "blog/index", nil); err != nil {
		t.Fatal(err)
	}
}

func TestRenameNestedIndexRejected(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog/index", nil); err != nil {
		t.Fatal(err)
	}
	if err := Rename(dir, "blog/index", "blog/x", time.Now()); !errors.Is(err, ErrInvalidName) {
		t.Errorf("got %v, want ErrInvalidName", err)
	}
	if err := Create(dir, "a:b", nil); !errors.Is(err, ErrInvalidName) {
		t.Errorf("colon: got %v", err)
	}
}
