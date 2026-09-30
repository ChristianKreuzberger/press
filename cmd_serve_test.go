package main

import (
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
