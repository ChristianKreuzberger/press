package main_test

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ChristianKreuzberger/press/internal/themes"
)

func TestE2ERenameErrors(t *testing.T) {
	site := t.TempDir()
	run(t, site, "init")
	run(t, site, "create", "page", "about")
	run(t, site, "create", "page", "contact")
	run(t, site, "create", "section", "blog")
	pages := filepath.Join(site, "pages")
	about := readFile(t, filepath.Join(pages, "about.md"))

	t.Run("missing source", func(t *testing.T) {
		out := assertExit(t, site, 1, "rename", "page", "ghost", "new")
		if !strings.Contains(out, "not found") {
			t.Errorf("expected a not-found message, got: %s", out)
		}
	})
	t.Run("target exists", func(t *testing.T) {
		assertExit(t, site, 1, "rename", "page", "about", "contact")
		if got := readFile(t, filepath.Join(pages, "about.md")); got != about {
			t.Errorf("source changed after failed rename:\n%s", got)
		}
	})
	t.Run("same name", func(t *testing.T) {
		assertExit(t, site, 1, "rename", "page", "about", "about")
		if got := readFile(t, filepath.Join(pages, "about.md")); got != about {
			t.Errorf("source changed after failed rename:\n%s", got)
		}
	})
	t.Run("invalid name", func(t *testing.T) {
		assertExit(t, site, 1, "rename", "page", "about", "../escape")
		if exists(filepath.Join(site, "escape.md")) {
			t.Error("rename escaped the pages directory")
		}
	})
	t.Run("home page", func(t *testing.T) {
		assertExit(t, site, 1, "rename", "page", "index", "home")
		if !exists(filepath.Join(pages, "index.md")) {
			t.Error("index.md must stay")
		}
	})
	t.Run("page onto section", func(t *testing.T) {
		assertExit(t, site, 1, "rename", "page", "about", "blog")
	})
	t.Run("section onto page", func(t *testing.T) {
		assertExit(t, site, 1, "rename", "section", "blog", "contact")
		if !exists(filepath.Join(pages, "blog", "index.md")) {
			t.Error("section must be intact")
		}
	})
	t.Run("section onto section", func(t *testing.T) {
		run(t, site, "create", "section", "news")
		assertExit(t, site, 1, "rename", "section", "blog", "news")
	})
	t.Run("missing section", func(t *testing.T) {
		assertExit(t, site, 1, "rename", "section", "ghost", "x")
	})
}

func TestE2EListAndUpdateGaps(t *testing.T) {
	site := t.TempDir()
	if err := os.MkdirAll(filepath.Join(site, "pages"), 0o755); err != nil {
		t.Fatal(err)
	}

	if out := run(t, site, "list", "page"); !strings.Contains(out, "no pages found") {
		t.Errorf("empty site: got %q", out)
	}

	writeFile(t, filepath.Join(site, "pages", "wip.md"), "---\ntitle: WIP\ndraft: true\n---\n# WIP\n")
	writeFile(t, filepath.Join(site, "pages", "done.md"), "---\ntitle: Done\n---\n# Done\n")
	out := run(t, site, "list", "page")
	if !strings.Contains(out, "wip [draft]") || strings.Contains(out, "done [draft]") {
		t.Errorf("draft marking wrong:\n%s", out)
	}

	before := readFile(t, filepath.Join(site, "pages", "done.md"))
	assertExit(t, site, 1, "update", "page", "done", "--file", filepath.Join(site, "missing.md"))
	if got := readFile(t, filepath.Join(site, "pages", "done.md")); got != before {
		t.Errorf("page changed after failed update:\n%s", got)
	}
	assertExit(t, site, 1, "update", "page", "done", "--file", site) // a directory

	empty := filepath.Join(site, "empty.md")
	writeFile(t, empty, "")
	assertExit(t, site, 0, "update", "page", "done", "--file", empty)
	if got := readFile(t, filepath.Join(site, "pages", "done.md")); got != "" {
		t.Errorf("expected empty page after update with empty file, got %q", got)
	}
}

func TestE2ECreateFileErrors(t *testing.T) {
	site := t.TempDir()
	run(t, site, "init")
	pages := filepath.Join(site, "pages")

	assertExit(t, site, 1, "create", "page", "x", "--file", filepath.Join(site, "missing.md"))
	if exists(filepath.Join(pages, "x.md")) {
		t.Error("failed create left a page behind")
	}
	assertExit(t, site, 1, "create", "page", "blog/x", "--file", filepath.Join(site, "missing.md"))
	if exists(filepath.Join(pages, "blog")) {
		t.Error("failed create left a directory behind")
	}
	assertExit(t, site, 1, "create", "page", "x", "--file", site) // a directory
	if exists(filepath.Join(pages, "x.md")) {
		t.Error("failed create left a page behind")
	}
	assertExit(t, site, 2, "create", "page", "x", "--file")
	assertExit(t, site, 1, "create", "section", "s", "--file", filepath.Join(site, "missing.md"))
	if exists(filepath.Join(pages, "s")) {
		t.Error("failed create left a section behind")
	}
}

func TestE2EBuildStatic(t *testing.T) {
	site := t.TempDir()
	run(t, site, "init")

	t.Run("absent default dir", func(t *testing.T) {
		run(t, site, "build")
		if !exists(filepath.Join(site, "dist", "index.html")) {
			t.Error("build without a static dir failed to write index.html")
		}
	})
	t.Run("default dir", func(t *testing.T) {
		if err := os.MkdirAll(filepath.Join(site, "static"), 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(site, "static", "a.txt"), "default")
		run(t, site, "build")
		if got := readFile(t, filepath.Join(site, "dist", "static", "a.txt")); got != "default" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("custom dir", func(t *testing.T) {
		if err := os.MkdirAll(filepath.Join(site, "assets", "img"), 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(site, "assets", "img", "b.txt"), "custom")
		run(t, site, "build", "--static", "assets")
		if got := readFile(t, filepath.Join(site, "dist", "assets", "img", "b.txt")); got != "custom" {
			t.Errorf("got %q", got)
		}
		if exists(filepath.Join(site, "dist", "static", "a.txt")) {
			t.Error("default static dir must not be copied when --static names another")
		}
	})
	t.Run("rejected values", func(t *testing.T) {
		writeFile(t, filepath.Join(site, "afile"), "x")
		for _, v := range []string{"../outside", "/abs", "afile"} {
			if code, out := runCode(t, site, "build", "--static", v); code == 0 {
				t.Errorf("--static %q should fail; output:\n%s", v, out)
			}
		}
	})
}

func TestE2EInitEveryTheme(t *testing.T) {
	for _, name := range themes.Names() {
		t.Run(name, func(t *testing.T) {
			site := t.TempDir()
			run(t, site, "init", "--theme", name)
			run(t, site, "create", "page", "about")
			run(t, site, "create", "section", "blog")
			run(t, site, "create", "page", "blog/post")
			run(t, site, "build")
			for _, f := range []string{"index.html", "about.html", "blog/index.html", "blog/post.html"} {
				html := readFile(t, filepath.Join(site, "dist", filepath.FromSlash(f)))
				if !strings.Contains(html, "<html") {
					t.Errorf("%s does not look like HTML:\n%s", f, html)
				}
			}
			if !strings.Contains(readFile(t, filepath.Join(site, "dist", "index.html")), "about.html") {
				t.Error("home page is missing a link to about")
			}
		})
	}
}

func TestE2ETreeSections(t *testing.T) {
	site := t.TempDir()
	run(t, site, "init")
	run(t, site, "create", "page", "about")
	run(t, site, "create", "section", "blog")
	run(t, site, "create", "page", "blog/first")
	writeFile(t, filepath.Join(site, "pages", "blog", "wip.md"), "---\ntitle: W\ndraft: true\n---\n# W\n")
	writeFile(t, filepath.Join(site, "pages", "secret.md"), "---\ntitle: S\ndraft: true\n---\n# S\n")

	out := run(t, site, "tree")
	for _, want := range []string{"pages/\n", "├── about\n", "├── blog/\n", "│   ├── first\n", "│   └── wip [draft]\n", "└── secret [draft]\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("tree output missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "index") {
		t.Errorf("tree should list the home page:\n%s", out)
	}
}

func TestE2ETreeEmpty(t *testing.T) {
	site := t.TempDir()
	if err := os.MkdirAll(filepath.Join(site, "pages"), 0o755); err != nil {
		t.Fatal(err)
	}
	if out := run(t, site, "tree"); !strings.Contains(out, "no pages or sections found") {
		t.Errorf("got %q", out)
	}
}

// freePort returns a loopback TCP port that was free a moment ago.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

// startServe starts `press serve` and returns the base URL once it announces
// that it is serving. The process is stopped when the test ends.
func startServe(t *testing.T, site string, args ...string) string {
	t.Helper()
	port := freePort(t)
	cmd := exec.Command(pressBinary, append([]string{"serve", "--port", fmt.Sprint(port)}, args...)...)
	cmd.Dir = site
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	ready := make(chan struct{})
	go func() {
		sc := bufio.NewScanner(stdout)
		closed := false
		for sc.Scan() {
			if !closed && strings.HasPrefix(sc.Text(), "serving at") {
				closed = true
				close(ready)
			}
		}
	}()
	select {
	case <-ready:
	case <-time.After(10 * time.Second):
		t.Fatal("press serve did not start in time")
	}
	return fmt.Sprintf("http://127.0.0.1:%d", port)
}

// get fetches url, retrying briefly until the listener accepts connections.
func get(t *testing.T, url string) int {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		resp, err := http.Get(url) //nolint:gosec,noctx // loopback URL built from a free port in a test
		if err == nil {
			_ = resp.Body.Close()
			return resp.StatusCode
		}
		if time.Now().After(deadline) {
			t.Fatalf("GET %s: %v", url, err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestE2EServeDrafts(t *testing.T) {
	if testing.Short() {
		t.Skip("starts a server process")
	}
	newSite := func(t *testing.T) string {
		site := t.TempDir()
		run(t, site, "init")
		writeFile(t, filepath.Join(site, "pages", "wip.md"), "---\ntitle: WIP\ndraft: true\n---\n# WIP\n")
		return site
	}

	t.Run("without --drafts", func(t *testing.T) {
		base := startServe(t, newSite(t))
		if code := get(t, base+"/index.html"); code != http.StatusOK {
			t.Errorf("index.html: %d", code)
		}
		if code := get(t, base+"/wip.html"); code != http.StatusNotFound {
			t.Errorf("draft served without --drafts: %d", code)
		}
	})
	t.Run("with --drafts", func(t *testing.T) {
		base := startServe(t, newSite(t), "--drafts")
		if code := get(t, base+"/wip.html"); code != http.StatusOK {
			t.Errorf("draft not served with --drafts: %d", code)
		}
	})
}
