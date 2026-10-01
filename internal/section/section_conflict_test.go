package section

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateCollidesWithPage(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "pages"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pages", "about-us.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "about-us", []byte("y")); !errors.Is(err, ErrNameConflict) {
		t.Errorf("got %v, want ErrNameConflict", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "pages", "about-us")); err == nil {
		t.Error("section dir created next to page file")
	}
}

func TestRenameCollidesWithPage(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "news", []byte("y")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pages", "about-us.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Rename(dir, "news", "about-us", time.Now()); !errors.Is(err, ErrNameConflict) {
		t.Errorf("got %v, want ErrNameConflict", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "pages", "news", "index.md")); err != nil {
		t.Errorf("source section lost: %v", err)
	}
}
