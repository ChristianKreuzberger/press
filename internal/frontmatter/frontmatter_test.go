package frontmatter

import (
	"strings"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	now := time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)
	got := string(Generate("Home", now))

	want := "---\ntitle: \"Home\"\nalias: \"\"\ntags: []\ncreated_at: \"2026-04-22T10:00:00Z\"\nupdated_at: \"2026-04-22T10:00:00Z\"\n---\n"
	if got != want {
		t.Errorf("Generate():\ngot:  %q\nwant: %q", got, want)
	}
}

func TestGenerate_TitleWithSpecialChars(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	got := string(Generate(`My "Blog" Post`, now))

	if !strings.Contains(got, "title:") {
		t.Error("expected frontmatter to contain title field")
	}
	if !strings.HasPrefix(got, "---\n") {
		t.Error("expected frontmatter to start with ---")
	}
	if !strings.Contains(got, "---\n") {
		t.Error("expected frontmatter to contain closing ---")
	}
}

func TestGenerate_TimestampFormat(t *testing.T) {
	now := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
	got := string(Generate("test", now))

	if !strings.Contains(got, "2026-12-31T23:59:59Z") {
		t.Errorf("expected RFC3339 timestamp in output, got: %s", got)
	}
}

func TestGenerate_CreatedAndUpdatedMatch(t *testing.T) {
	now := time.Now().UTC()
	got := string(Generate("page", now))

	ts := now.Format(time.RFC3339)
	if strings.Count(got, ts) != 2 {
		t.Errorf("expected created_at and updated_at to both equal %q\ngot: %s", ts, got)
	}
}

func TestGenerate_EmptyAliasAndTags(t *testing.T) {
	now := time.Now().UTC()
	got := string(Generate("test", now))

	if !strings.Contains(got, `alias: ""`) {
		t.Errorf("expected empty alias field, got: %s", got)
	}
	if !strings.Contains(got, "tags: []") {
		t.Errorf("expected empty tags field, got: %s", got)
	}
}
