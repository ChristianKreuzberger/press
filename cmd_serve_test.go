package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCollectFileStates_ReturnsFiles(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.md"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}

	states, err := collectFileStates(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 2 {
		t.Errorf("expected 2 states, got %d", len(states))
	}
}

func TestCollectFileStates_ExcludesOutputDir(t *testing.T) {
	dir := t.TempDir()
	distDir := filepath.Join(dir, "dist")
	if err := os.Mkdir(distDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "index.md"), []byte("# Home"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "index.html"), []byte("<html>"), 0644); err != nil {
		t.Fatal(err)
	}

	states, err := collectFileStates(dir, distDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := states[filepath.Join(distDir, "index.html")]; ok {
		t.Error("file inside excluded dir should not appear in states")
	}
	if _, ok := states[filepath.Join(dir, "index.md")]; !ok {
		t.Error("source file should appear in states")
	}
}

func TestCollectFileStates_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	states, err := collectFileStates(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(states) != 0 {
		t.Errorf("expected 0 states for empty dir, got %d", len(states))
	}
}

func TestHasChanged_NoChange(t *testing.T) {
	ts := time.Now()
	prev := map[string]time.Time{"index.md": ts}
	curr := map[string]time.Time{"index.md": ts}
	if hasChanged(prev, curr) {
		t.Error("expected no change when states are identical")
	}
}

func TestHasChanged_FileAdded(t *testing.T) {
	ts := time.Now()
	prev := map[string]time.Time{"index.md": ts}
	curr := map[string]time.Time{"index.md": ts, "about.md": ts}
	if !hasChanged(prev, curr) {
		t.Error("expected change when a file is added")
	}
}

func TestHasChanged_FileRemoved(t *testing.T) {
	ts := time.Now()
	prev := map[string]time.Time{"index.md": ts, "about.md": ts}
	curr := map[string]time.Time{"index.md": ts}
	if !hasChanged(prev, curr) {
		t.Error("expected change when a file is removed")
	}
}

func TestHasChanged_FileModified(t *testing.T) {
	ts := time.Now()
	prev := map[string]time.Time{"index.md": ts}
	curr := map[string]time.Time{"index.md": ts.Add(time.Second)}
	if !hasChanged(prev, curr) {
		t.Error("expected change when a file modification time differs")
	}
}

func TestHasChanged_EmptyStates(t *testing.T) {
	prev := map[string]time.Time{}
	curr := map[string]time.Time{}
	if hasChanged(prev, curr) {
		t.Error("expected no change for two empty states")
	}
}
