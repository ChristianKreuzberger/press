package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// renameSite creates a site with a page, a section and pages that link to them.
func renameSite(t *testing.T) string {
	t.Helper()
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	pages := filepath.Join(siteDir, "pages")
	if err := os.MkdirAll(filepath.Join(pages, "blog"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(pages, "about.md"), "---\ntitle: \"About\"\n---\n# About\n")
	writeFile(t, filepath.Join(pages, "blog", "index.md"), "---\ntitle: \"Blog\"\n---\n# Blog\n")
	writeFile(t, filepath.Join(pages, "blog", "first.md"), "---\ntitle: \"First\"\n---\n# First\n")
	// Similarly named pages the renames must not touch.
	writeFile(t, filepath.Join(pages, "about-more.md"), "---\ntitle: \"More\"\n---\n# More\n")
	writeFile(t, filepath.Join(pages, "blogger.md"), "---\ntitle: \"Blogger\"\n---\n# Blogger\n")
	writeFile(t, filepath.Join(pages, "links.md"), "---\ntitle: \"Links\"\n---\n"+
		"[a](/about) [b](/about.html) [c](/about#team) [d](/about?x=1) [e](/about \"T\")\n"+
		"[f](/blog) [g](/blog/) [h](/blog/first) [i](/blog/first.html#x) [j](/blog/index.html)\n"+
		"[k](/about-more) [l](/blogger) ![img](/about)\n"+
		"`[code](/about)`\n\n```\n[fenced](/about) [fenced](/blog/first)\n```\n")
	writeFile(t, filepath.Join(pages, "blog", "second.md"), "---\ntitle: \"Second\"\ndraft: true\n---\n[back](/about) [up](/blog)\n")
	return siteDir
}

func TestE2ERenamePageRewritesLinks(t *testing.T) {
	siteDir := renameSite(t)
	out := run(t, siteDir, "rename", "page", "about", "about-us")
	if !strings.Contains(out, "updated") {
		t.Errorf("output should report updated links, got: %s", out)
	}
	links := readFile(t, filepath.Join(siteDir, "pages", "links.md"))
	for _, want := range []string{"[a](/about-us)", "[b](/about-us.html)", "[c](/about-us#team)", "[d](/about-us?x=1)", `[e](/about-us "T")`,
		"[k](/about-more)", "![img](/about)", "`[code](/about)`", "[fenced](/about) [fenced](/blog/first)", "[f](/blog)"} {
		if !strings.Contains(links, want) {
			t.Errorf("links.md should contain %q, got:\n%s", want, links)
		}
	}
	if !strings.Contains(readFile(t, filepath.Join(siteDir, "pages", "blog", "second.md")), "[back](/about-us)") {
		t.Error("links in draft pages should be rewritten too")
	}
	run(t, siteDir, "check")
}

func TestE2ERenameSectionRewritesLinks(t *testing.T) {
	siteDir := renameSite(t)
	run(t, siteDir, "rename", "section", "blog", "journal")
	links := readFile(t, filepath.Join(siteDir, "pages", "links.md"))
	for _, want := range []string{"[f](/journal)", "[g](/journal/)", "[h](/journal/first)", "[i](/journal/first.html#x)", "[j](/journal/index.html)",
		"[l](/blogger)", "[fenced](/about) [fenced](/blog/first)", "[a](/about)"} {
		if !strings.Contains(links, want) {
			t.Errorf("links.md should contain %q, got:\n%s", want, links)
		}
	}
	if !strings.Contains(readFile(t, filepath.Join(siteDir, "pages", "journal", "second.md")), "[up](/journal)") {
		t.Error("links inside the renamed section should be rewritten")
	}
	run(t, siteDir, "check")
}

func TestE2ERenameMinimalFrontmatter(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	pages := filepath.Join(siteDir, "pages")
	writeFile(t, filepath.Join(pages, "plain.md"), "# Plain\n\nBody\n")
	writeFile(t, filepath.Join(pages, "titled.md"), "---\ntitle: \"Custom\"\n---\nBody\n")
	run(t, siteDir, "rename", "page", "plain", "simple")
	run(t, siteDir, "rename", "page", "titled", "renamed")
	s := readFile(t, filepath.Join(pages, "simple.md"))
	if !strings.Contains(s, `title: "Simple"`) || !strings.Contains(s, "# Plain") {
		t.Errorf("unexpected content: %s", s)
	}
	if !strings.Contains(readFile(t, filepath.Join(pages, "renamed.md")), `title: "Custom"`) {
		t.Error("custom title should be kept")
	}
}

func TestE2ERenameLinkRewriteFailureExitsNonZero(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("file permissions are not enforced for root")
	}
	siteDir := renameSite(t)
	locked := filepath.Join(siteDir, "pages", "links.md")
	if err := os.Chmod(locked, 0444); err != nil {
		t.Fatal(err)
	}
	out := runExpectError(t, siteDir, "rename", "page", "about", "about-us")
	if !strings.Contains(out, "rename succeeded") || !strings.Contains(out, "links.md") {
		t.Errorf("should name the file and say the rename succeeded, got: %s", out)
	}
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "about-us.md")); err != nil {
		t.Errorf("rename itself should have happened: %v", err)
	}
}
