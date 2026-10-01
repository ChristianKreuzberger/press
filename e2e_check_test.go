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
