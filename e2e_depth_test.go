package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Pages deeper than section/page are never built, so creating them is refused.
func TestE2ECreatePageTooDeepIsRejected(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	run(t, siteDir, "create", "page", "blog/post") // one level is fine

	out := runExpectError(t, siteDir, "create", "page", "blog/2026/post")
	if !strings.Contains(out, "invalid page name") {
		t.Errorf("expected 'invalid page name' error, got: %s", out)
	}
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "blog", "2026")); err == nil {
		t.Error("pages/blog/2026 should not have been created")
	}
}

// A hand-made nested .md file must not be dropped silently by build.
func TestE2EBuildWarnsAboutSkippedNestedMarkdown(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	run(t, siteDir, "create", "section", "blog")
	nested := filepath.Join(siteDir, "pages", "blog", "2026")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "post.md"), []byte("# hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := run(t, siteDir, "build")
	if !strings.Contains(out, "skipping") || !strings.Contains(out, "blog/2026/post.md") {
		t.Errorf("expected skip warning for nested page, got:\n%s", out)
	}
}
