package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestListenAddr(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		want string
	}{
		{"default host", defaultServeHost, 8080, "127.0.0.1:8080"},
		{"empty falls back to default", "", 8080, "127.0.0.1:8080"},
		{"whitespace falls back to default", "  ", 9000, "127.0.0.1:9000"},
		{"all interfaces", "0.0.0.0", 3000, "0.0.0.0:3000"},
		{"hostname", "localhost", 8080, "localhost:8080"},
		{"ipv6 bare", "::1", 8080, "[::1]:8080"},
		{"ipv6 bracketed", "[::1]", 8080, "[::1]:8080"},
		{"trims spaces", " 127.0.0.1 ", 8080, "127.0.0.1:8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listenAddr(tt.host, tt.port)
			if got != tt.want {
				t.Errorf("listenAddr(%q, %d) = %q, want %q", tt.host, tt.port, got, tt.want)
			}
			if strings.HasPrefix(got, ":") {
				t.Errorf("listenAddr(%q, %d) = %q must not bind all interfaces", tt.host, tt.port, got)
			}
		})
	}
}

func TestIsLoopbackHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"", true},
		{"localhost", true},
		{"127.0.0.1", true},
		{"127.0.0.2", true},
		{"::1", true},
		{"[::1]", true},
		{"0.0.0.0", false},
		{"::", false},
		{"192.168.1.10", false},
		{"example.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := isLoopbackHost(tt.host); got != tt.want {
				t.Errorf("isLoopbackHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestCollectFileStates_WatchesOnlySources(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "pages", "index.md"), "# Home")
	writeTestFile(t, filepath.Join(dir, "pages", "blog", "post.md"), "# Post")
	writeTestFile(t, filepath.Join(dir, "template.html"), "<html>")
	writeTestFile(t, filepath.Join(dir, "static", "css", "site.css"), "body{}")
	// None of these may trigger a rebuild.
	writeTestFile(t, filepath.Join(dir, "dist", "index.html"), "<html>")
	writeTestFile(t, filepath.Join(dir, ".git", "index"), "x")
	writeTestFile(t, filepath.Join(dir, "node_modules", "pkg", "a.js"), "x")
	writeTestFile(t, filepath.Join(dir, "notes.txt"), "x")

	states, err := collectFileStates(dir, "static")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		filepath.Join(dir, "pages", "index.md"),
		filepath.Join(dir, "pages", "blog", "post.md"),
		filepath.Join(dir, "template.html"),
		filepath.Join(dir, "static", "css", "site.css"),
	}
	if len(states) != len(want) {
		t.Errorf("expected %d watched files, got %d: %v", len(want), len(states), states)
	}
	for _, w := range want {
		if _, ok := states[w]; !ok {
			t.Errorf("expected %s to be watched", w)
		}
	}
}

func TestCollectFileStates_MissingSourcesAreNotAnError(t *testing.T) {
	dir := t.TempDir()

	states, err := collectFileStates(dir, "static")
	if err != nil {
		t.Fatalf("missing pages/, template.html and static dir must not be an error: %v", err)
	}
	if len(states) != 0 {
		t.Errorf("expected 0 states, got %d", len(states))
	}
}

func TestNextSnapshot_UnchangedSkipsBuild(t *testing.T) {
	ts := time.Now()
	prev := map[string]time.Time{"a": ts}
	curr := map[string]time.Time{"a": ts}

	next, built, err := nextSnapshot(prev, curr, func() error {
		t.Error("build must not run when nothing changed")
		return nil
	})
	if err != nil || built {
		t.Errorf("got built=%v err=%v, want false, nil", built, err)
	}
	if hasChanged(next, curr) {
		t.Error("snapshot should stay current")
	}
}

func TestNextSnapshot_SuccessAdvancesSnapshot(t *testing.T) {
	ts := time.Now()
	prev := map[string]time.Time{"a": ts}
	curr := map[string]time.Time{"a": ts.Add(time.Second)}

	next, built, err := nextSnapshot(prev, curr, func() error { return nil })
	if err != nil || !built {
		t.Fatalf("got built=%v err=%v, want true, nil", built, err)
	}
	if hasChanged(next, curr) {
		t.Error("snapshot should advance to curr after a successful build")
	}
}

func TestNextSnapshot_FailureKeepsOldSnapshotSoItIsRetried(t *testing.T) {
	ts := time.Now()
	prev := map[string]time.Time{"a": ts}
	curr := map[string]time.Time{"a": ts.Add(time.Second)}

	next, built, err := nextSnapshot(prev, curr, func() error { return errors.New("boom") })
	if err == nil || built {
		t.Fatalf("got built=%v err=%v, want false, error", built, err)
	}
	if !hasChanged(next, curr) {
		t.Error("snapshot must not advance after a failed build, otherwise it is never retried")
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
