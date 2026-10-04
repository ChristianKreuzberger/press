package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2ECheckNoFalsePositives covers issue #61: valid links must not be
// reported as broken.
func TestE2ECheckNoFalsePositives(t *testing.T) {
	siteDir := t.TempDir()
	pagesDir := filepath.Join(siteDir, "pages")
	run(t, siteDir, "init")

	write := func(path, content string) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, path, content)
	}
	page := func(name, body string) {
		write(filepath.Join(pagesDir, name), "---\ntitle: \"T\"\n---\n"+body+"\n")
	}
	page("about.md", "# About\n\nAbout us.")
	write(filepath.Join(pagesDir, "blog", "index.md"), "---\ntitle: \"Blog\"\n---\n# Blog\n\nPosts.\n")
	write(filepath.Join(pagesDir, "blog", "post.md"), "---\ntitle: \"Post\"\n---\n# Post\n\nBody.\n")
	write(filepath.Join(pagesDir, "blog", "first_(1).md"), "---\ntitle: \"First\"\n---\n# First\n\nBody.\n")
	write(filepath.Join(siteDir, "static", "docs", "f.pdf"), "pdf")
	write(filepath.Join(pagesDir, "files", "g.pdf"), "pdf")

	good := map[string]string{
		"title":           `[e](/about "T")`,
		"html ext":        `[a](/about.html) [b](/blog/index.html) [c](/blog/post.html) [d](/about)`,
		"protocol-rel":    `[x](//example.com/nope)`,
		"scheme":          `[x](https://example.com/nope) [y](mailto:a@b.c)`,
		"static file":     `[f](/docs/f.pdf)`,
		"pages asset":     `[g](/files/g.pdf)`,
		"fenced code":     "```\n[x](/missing)\n```",
		"tilde fence":     "~~~\n[x](/missing)\n~~~",
		"inline code":     "use `[x](/missing)` syntax",
		"parens":          `[p](/blog/first_(1))`,
		"angle brackets":  `[p](</blog/first_(1)> "T")`,
		"long fence":      "````\n```\n[x](/missing)\n```go\n[y](/missing)\n````",
		"dot segments":    `[a](/blog/../about) [b](/./about)`,
		"percent-encoded": `[p](/blog/first_%281%29)`,
	}
	for name, body := range good {
		page("links.md", body)
		if out := run(t, siteDir, "check"); strings.Contains(out, "broken link") {
			t.Errorf("%s: %q should not be reported as broken, got: %s", name, body, out)
		}
	}

	// Real breakage must still be caught, including with a title present.
	for _, body := range []string{`[x](/missing "T")`, `[x](/nope.pdf)`, `[x](/blog/nope_(1))`, `[x](/missing.html)`, `[x](/blog.html)`} {
		page("links.md", body)
		if out := runExpectError(t, siteDir, "check"); !strings.Contains(out, "broken link") {
			t.Errorf("%q should be reported as broken, got: %s", body, out)
		}
	}
}

// checkSite builds a site with a few pages for the check tests below.
func checkSite(t *testing.T) (siteDir string, write func(rel, content string)) {
	t.Helper()
	siteDir = t.TempDir()
	run(t, siteDir, "init")
	write = func(rel, content string) {
		p := filepath.Join(siteDir, "pages", filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, p, content)
	}
	return siteDir, write
}

func fm(extra string) string { return "---\ntitle: \"T\"\n" + extra + "---\n" }

// TestE2ECheckDraftLinks covers issue #62: links to pages the build skips
// (drafts, draft sections) are broken.
func TestE2ECheckDraftLinks(t *testing.T) {
	siteDir, write := checkSite(t)
	write("about.md", fm("")+"# About\n")
	write("secret.md", fm("draft: true\n")+"# Secret\n")
	write("blog/index.md", fm("")+"# Blog\n")
	write("blog/wip.md", fm("draft: true\n")+"# WIP\n")
	write("hidden/index.md", fm("draft: true\n")+"# Hidden\n")
	write("hidden/page.md", fm("")+"# P\n")

	for _, dest := range []string{"/secret", "/secret.html", "/blog/wip", "/hidden", "/hidden/page"} {
		write("links.md", fm("")+"[x]("+dest+")\n")
		out := runExpectError(t, siteDir, "check")
		if !strings.Contains(out, "broken link") || !strings.Contains(out, "draft") {
			t.Errorf("%s should be reported as a draft link, got: %s", dest, out)
		}
	}

	// Links from a draft page are still checked, and links to live pages pass.
	write("links.md", fm("draft: true\n")+"[x](/missing) [y](/about)\n")
	out := runExpectError(t, siteDir, "check")
	if !strings.Contains(out, "/missing") || strings.Contains(out, "→ /about") {
		t.Errorf("draft page links should be checked, got: %s", out)
	}
}

// TestE2ECheckRelativeAndAnchorLinks covers issue #62: relative links and
// <a href> links are validated.
func TestE2ECheckRelativeAndAnchorLinks(t *testing.T) {
	siteDir, write := checkSite(t)
	write("about.md", fm("")+"# About\n")
	write("blog/index.md", fm("")+"# Blog\n")
	write("blog/post.md", fm("")+"# Post\n")

	good := map[string]string{
		"top rel":      "links.md|[a](about.html) [b](./about) [c](blog/post.html) [d](blog/)",
		"top frag":     "links.md|[a](#top) [b](about#x)",
		"section rel":  "blog/rel.md|[a](post.html) [b](./post) [c](../about.html) [d](index.html)",
		"html anchors": "links.md|<a href=\"/about\">a</a> <a href='blog/post.html'>b</a>",
		"ext html":     "links.md|<a href=\"https://x.org/nope\">a</a> <a href=\"mailto:a@b.c\">m</a>",
		"html in code": "links.md|`<a href=\"/nope\">` and\n```\n<a href=\"/nope\">x</a>\n```",
	}
	for name, spec := range good {
		file, body, _ := strings.Cut(spec, "|")
		write(file, fm("")+body+"\n")
		if out := run(t, siteDir, "check"); strings.Contains(out, "broken link") {
			t.Errorf("%s: should be valid, got: %s", name, out)
		}
		_ = os.Remove(filepath.Join(siteDir, "pages", filepath.FromSlash(file)))
	}

	bad := map[string]string{
		"top rel":      "links.md|[a](nope.html)",
		"section rel":  "blog/rel.md|[a](about.html)",
		"escape root":  "links.md|[a](../about.html)",
		"html anchor":  "links.md|<a href=\"/nope\">a</a>",
		"html rel":     "links.md|<a href='nope.html'>a</a>",
		"section only": "blog/rel.md|[a](nope)",
	}
	for name, spec := range bad {
		file, body, _ := strings.Cut(spec, "|")
		write(file, fm("")+body+"\n")
		if out := runExpectError(t, siteDir, "check"); !strings.Contains(out, "broken link") {
			t.Errorf("%s: should be broken, got: %s", name, out)
		}
		_ = os.Remove(filepath.Join(siteDir, "pages", filepath.FromSlash(file)))
	}
}

// TestE2ECheckSectionIndexNoHtml covers issue #62: the builder only writes
// <section>/index.html, so /<section>/index is not a real URL.
func TestE2ECheckSectionIndexNoHtml(t *testing.T) {
	siteDir, write := checkSite(t)
	write("blog/index.md", fm("")+"# Blog\n")
	write("links.md", fm("")+"[a](/blog/index)\n")
	if out := runExpectError(t, siteDir, "check"); !strings.Contains(out, "/blog/index") {
		t.Errorf("/blog/index should be broken, got: %s", out)
	}
	write("links.md", fm("")+"[a](/blog) [b](/blog/) [c](/blog/index.html)\n")
	if out := run(t, siteDir, "check"); strings.Contains(out, "broken link") {
		t.Errorf("got: %s", out)
	}
}

// TestE2ECheckUnreadablePageDoesNotAbort covers issue #62: a page that cannot
// be read is reported as an issue and the other pages are still checked.
func TestE2ECheckUnreadablePageDoesNotAbort(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("file permissions are not enforced for root")
	}
	siteDir, write := checkSite(t)
	write("bad.md", fm("")+"# Bad\n")
	write("blog/index.md", fm("")+"# Blog\n")
	write("blog/bad.md", fm("")+"# Bad\n")
	write("links.md", fm("")+"[x](/missing)\n")
	for _, f := range []string{"bad.md", "blog/bad.md"} {
		p := filepath.Join(siteDir, "pages", filepath.FromSlash(f))
		if err := os.Chmod(p, 0); err != nil {
			t.Fatal(err)
		}
	}
	out := runExpectError(t, siteDir, "check")
	for _, want := range []string{"pages checked", "bad.md", "blog/bad.md", "/missing", "issue(s) found"} {
		if !strings.Contains(out, want) {
			t.Errorf("output should contain %q, got: %s", want, out)
		}
	}
}

// TestE2ECheckHugeFrontmatterLine is a regression test: a very long
// frontmatter line must not abort the check.
func TestE2ECheckHugeFrontmatterLine(t *testing.T) {
	siteDir, write := checkSite(t)
	write("big.md", "---\ntitle: \"T\"\nnote: \""+strings.Repeat("x", 200000)+"\"\n---\n# Big\n")
	write("links.md", fm("")+"[x](/missing)\n")
	out := runExpectError(t, siteDir, "check")
	if !strings.Contains(out, "pages checked") || !strings.Contains(out, "/missing") {
		t.Errorf("got: %s", out)
	}
}
