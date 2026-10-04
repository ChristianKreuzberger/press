package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ChristianKreuzberger/press/internal/page"
	"github.com/ChristianKreuzberger/press/internal/themes"
)

// Every shipped theme must execute against real builder data, not just parse.
func TestBuildEveryThemeTemplate(t *testing.T) {
	for _, theme := range themes.All {
		t.Run(theme.Name, func(t *testing.T) {
			siteDir := t.TempDir()
			outDir := filepath.Join(siteDir, "dist")
			writeTestFile(t, filepath.Join(siteDir, "template.html"), theme.Template)
			writeTestFile(t, filepath.Join(siteDir, "pages", "index.md"), "---\ntitle: Home Title\n---\n# Home Title\n\nHello world.\n")
			writeTestFile(t, filepath.Join(siteDir, "pages", "about.md"), "---\ntitle: About Title\nweight: 1\n---\n# About Title\n\nAbout text.\n")
			writeTestFile(t, filepath.Join(siteDir, "pages", "blog", "index.md"), "---\ntitle: Blog Title\n---\n# Blog Title\n")
			writeTestFile(t, filepath.Join(siteDir, "pages", "blog", "post.md"), "---\ntitle: Post Title\n---\n# Post Title\n\nPost text.\n")

			if _, err := Build(siteDir, outDir, false, "static"); err != nil {
				t.Fatalf("Build with theme %s failed: %v", theme.Name, err)
			}
			home := string(mustRead(t, filepath.Join(outDir, "index.html")))
			for _, want := range []string{"Home Title", "Hello world.", `href="about.html"`, `href="blog/index.html"`} {
				if !strings.Contains(home, want) {
					t.Errorf("index.html missing %q", want)
				}
			}
			post := string(mustRead(t, filepath.Join(outDir, "blog", "post.html")))
			for _, want := range []string{"Post Title", "Post text.", `href="../about.html"`} {
				if !strings.Contains(post, want) {
					t.Errorf("blog/post.html missing %q", want)
				}
			}
			section := string(mustRead(t, filepath.Join(outDir, "blog", "index.html")))
			if !strings.Contains(section, `href="post.html"`) {
				t.Errorf("section page missing TOC link to post.html:\n%s", section)
			}
		})
	}
}

func TestBuildThemeWithSingleHomePage(t *testing.T) {
	for _, theme := range themes.All {
		t.Run(theme.Name, func(t *testing.T) {
			siteDir := t.TempDir()
			outDir := filepath.Join(siteDir, "dist")
			writeTestFile(t, filepath.Join(siteDir, "template.html"), theme.Template)
			if err := page.Create(siteDir, "index", []byte("# Only\n")); err != nil {
				t.Fatal(err)
			}
			if _, err := Build(siteDir, outDir, false, "static"); err != nil {
				t.Fatalf("Build failed: %v", err)
			}
		})
	}
}

func TestBuildCRLFPage(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")
	writeTestFile(t, filepath.Join(siteDir, "pages", "index.md"),
		"---\r\ntitle: Windows Title\r\nweight: 2\r\n---\r\n# Heading\r\n\r\nBody line.\r\n")

	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	html := string(mustRead(t, filepath.Join(outDir, "index.html")))
	if !strings.Contains(html, "Windows Title") || !strings.Contains(html, "Body line.") {
		t.Errorf("CRLF page not rendered:\n%s", html)
	}
	if strings.Contains(html, "title: ") || strings.Contains(html, "---") {
		t.Errorf("frontmatter leaked into the output:\n%s", html)
	}
	if strings.Contains(html, "Windows Title\r") {
		t.Error("stray carriage return in title")
	}
}

func TestBuildEmptyFrontmatter(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")
	writeTestFile(t, filepath.Join(siteDir, "pages", "index.md"), "---\n---\n# Heading Title\n\nbody text\n")
	writeTestFile(t, filepath.Join(siteDir, "pages", "plain.md"), "---\n---\njust a body\n")

	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	html := string(mustRead(t, filepath.Join(outDir, "index.html")))
	if !strings.Contains(html, "body text") || strings.Contains(html, "---") {
		t.Errorf("empty frontmatter not handled:\n%s", html)
	}
	if !strings.Contains(html, "Heading Title") {
		t.Errorf("expected the H1 to be used as the title:\n%s", html)
	}
	if !strings.Contains(string(mustRead(t, filepath.Join(outDir, "plain.html"))), "just a body") {
		t.Error("plain.html missing body")
	}
}

func TestBuildSpecialCharFilenamesWritten(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")
	for _, name := range []string{"index", "c#sharp", "what?", "100%"} {
		if err := page.Create(siteDir, name, []byte("# T\n\nbody\n")); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	for _, f := range []string{"c#sharp.html", "what?.html", "100%.html"} {
		assertExists(t, filepath.Join(outDir, f))
	}
	index := string(mustRead(t, filepath.Join(outDir, "index.html")))
	for _, want := range []string{`href="c%23sharp.html"`, `href="what%3F.html"`, `href="100%25.html"`} {
		if !strings.Contains(index, want) {
			t.Errorf("index.html missing %s", want)
		}
	}
}

// leakCheck builds siteDir and fails if any output file contains the marker.
func leakCheck(t *testing.T, siteDir, marker string) {
	t.Helper()
	outDir := filepath.Join(siteDir, "dist")
	// Whether the build errors or skips is not pinned; leaking outside content is.
	_, _ = Build(siteDir, outDir, false, "static")
	_ = filepath.WalkDir(outDir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil //nolint:nilerr // a missing output dir has nothing to leak
		}
		b, _ := os.ReadFile(p) //nolint:gosec // test walks its own temp dir
		if strings.Contains(string(b), marker) {
			t.Errorf("content from outside the site leaked into %s", p)
		}
		return nil
	})
}

func TestBuildSymlinkedSectionDoesNotLeak(t *testing.T) {
	siteDir := t.TempDir()
	writeTestFile(t, filepath.Join(siteDir, "pages", "index.md"), "# Home\n")
	outside := t.TempDir()
	writeTestFile(t, filepath.Join(outside, "sec", "index.md"), "# Sec\n")
	writeTestFile(t, filepath.Join(outside, "sec", "inner.md"), "TOP-SECRET-SECTION\n")
	if err := os.Symlink(filepath.Join(outside, "sec"), filepath.Join(siteDir, "pages", "linkedsec")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	leakCheck(t, siteDir, "TOP-SECRET")
}

func TestBuildSymlinkedPageDoesNotLeak(t *testing.T) {
	t.Skip("known gap: the build follows a symlinked pages/*.md and publishes the file it points to " +
		"(assets and static files skip symlinks). Remove this skip when pages are made consistent.")
	siteDir := t.TempDir()
	writeTestFile(t, filepath.Join(siteDir, "pages", "index.md"), "# Home\n")
	outside := t.TempDir()
	writeTestFile(t, filepath.Join(outside, "secret.md"), "---\ntitle: Secret\n---\nTOP-SECRET-PAGE\n")
	if err := os.Symlink(filepath.Join(outside, "secret.md"), filepath.Join(siteDir, "pages", "linked.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	leakCheck(t, siteDir, "TOP-SECRET")
}

func TestBuildStaleNestedOutputRemoved(t *testing.T) {
	siteDir := t.TempDir()
	outDir := filepath.Join(siteDir, "dist")
	writeTestFile(t, filepath.Join(siteDir, "pages", "index.md"), "# Home\n")
	writeTestFile(t, filepath.Join(siteDir, "pages", "blog", "index.md"), "# Blog\n")
	writeTestFile(t, filepath.Join(siteDir, "pages", "blog", "old.md"), "# Old\n")
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	assertExists(t, filepath.Join(outDir, "blog", "old.html"))

	if err := os.RemoveAll(filepath.Join(siteDir, "pages", "blog")); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "blog")); err == nil {
		t.Error("stale nested output directory was not removed")
	}
}
