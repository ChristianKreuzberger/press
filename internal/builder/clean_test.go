package builder

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/ChristianKreuzberger/press/internal/page"
)

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err != nil {
		t.Errorf("%s must still exist: %v", path, err)
	}
}

func TestBuildRemovesStaleOutput(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if err := page.Create(siteDir, "old", []byte("# Old\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(siteDir, "pages", "old.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "old.html")); err == nil {
		t.Error("dist/old.html should be removed after the page was deleted")
	}
	if _, err := os.Stat(filepath.Join(outDir, "index.html")); err != nil {
		t.Error("dist/index.html should still exist")
	}
}

func TestBuildRefusesDangerousOutputDir(t *testing.T) {
	siteDir, _ := newAssetSite(t)
	if err := os.MkdirAll(filepath.Join(siteDir, "static"), 0755); err != nil {
		t.Fatal(err)
	}
	sibling := filepath.Join(filepath.Dir(siteDir), filepath.Base(siteDir)+"-sibling")
	t.Cleanup(func() { _ = os.RemoveAll(sibling) })
	writeTestFile(t, filepath.Join(siteDir, ".git", "HEAD"), "ref")
	writeTestFile(t, filepath.Join(siteDir, "docs", "keep.txt"), "keep")
	writeTestFile(t, filepath.Join(siteDir, "unrelated", "notes.txt"), "keep")
	writeTestFile(t, filepath.Join(sibling, "mine.txt"), "keep")
	writeTestFile(t, filepath.Join(filepath.Dir(siteDir), "next-to-site.txt"), "keep")

	// Symlinks to an unrelated non-empty dir and to the site's pages/.
	linkUnrelated := filepath.Join(siteDir, "link-unrelated")
	linkPages := filepath.Join(siteDir, "link-pages")
	if err := os.Symlink(sibling, linkUnrelated); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Join(siteDir, "pages"), linkPages); err != nil {
		t.Fatal(err)
	}

	// Source overlaps are caught by validateOutputDir; non-empty dirs that
	// press did not create are caught by the ownership check.
	cases := map[string]struct{ out, want string }{
		"site dir":              {siteDir, "invalid output directory"},
		"parent dir":            {filepath.Dir(siteDir), "invalid output directory"},
		"pages dir":             {filepath.Join(siteDir, "pages"), "invalid output directory"},
		"static dir":            {filepath.Join(siteDir, "static"), "invalid output directory"},
		"inside pages":          {filepath.Join(siteDir, "pages", "out"), "invalid output directory"},
		"template.html":         {filepath.Join(siteDir, "template.html"), "invalid output directory"},
		".git":                  {filepath.Join(siteDir, ".git"), "invalid output directory"},
		"symlink to pages":      {linkPages, "invalid output directory"},
		"new dir under symlink": {filepath.Join(linkPages, "newdir"), "invalid output directory"},
		"docs":                  {filepath.Join(siteDir, "docs"), "refusing to use output directory"},
		"non-empty unrelated":   {filepath.Join(siteDir, "unrelated"), "refusing to use output directory"},
		"sibling outside site":  {sibling, "refusing to use output directory"},
		"symlink to unrelated":  {linkUnrelated, "refusing to use output directory"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Build(siteDir, c.out, false, "static")
			if err == nil {
				t.Fatalf("Build with output %s should fail", c.out)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("expected %q, got: %v", c.want, err)
			}
		})
	}

	for _, p := range []string{
		filepath.Join(siteDir, "pages", "index.md"),
		filepath.Join(siteDir, ".git", "HEAD"),
		filepath.Join(siteDir, "docs", "keep.txt"),
		filepath.Join(siteDir, "unrelated", "notes.txt"),
		filepath.Join(sibling, "mine.txt"),
		filepath.Join(filepath.Dir(siteDir), "next-to-site.txt"),
	} {
		assertExists(t, p)
	}
	// Validation runs before any directory is created.
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "out")); err == nil {
		t.Error("a refused build must not leave pages/out behind")
	}
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "newdir")); err == nil {
		t.Error("a refused build must not create dirs through a symlink")
	}
}

func TestBuildWritesMarkerAndRebuilds(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	for i := 0; i < 2; i++ {
		if _, err := Build(siteDir, outDir, false, "static"); err != nil {
			t.Fatalf("build %d: %v", i+1, err)
		}
		assertExists(t, filepath.Join(outDir, outputMarker))
	}
}

func TestBuildCleansMarkedDirWithAnyContent(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(outDir, "old", "thing.pdf"), "x")
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "old")); err == nil {
		t.Error("stale content in a marked output dir should be removed")
	}
}

// Output dirs created by older press versions have no marker; upgrading must
// not require manual cleanup.
func TestBuildAcceptsLegacyUnmarkedOutput(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	writeTestFile(t, filepath.Join(siteDir, "static", "css", "site.css"), "a{}")
	writeTestFile(t, filepath.Join(outDir, "index.html"), "<html>")
	writeTestFile(t, filepath.Join(outDir, "gone.html"), "<html>")
	writeTestFile(t, filepath.Join(outDir, "blog", "post.html"), "<html>")
	writeTestFile(t, filepath.Join(outDir, "css", "site.css"), "a{}")

	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatalf("legacy dist should still build: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "gone.html")); err == nil {
		t.Error("stale file in legacy dist should be removed")
	}
	assertExists(t, filepath.Join(outDir, outputMarker))
}

func TestBuildRefusesUnmarkedDirWithForeignFiles(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	writeTestFile(t, filepath.Join(outDir, "index.html"), "<html>")
	writeTestFile(t, filepath.Join(outDir, "keep.txt"), "mine")
	if _, err := Build(siteDir, outDir, false, "static"); err == nil ||
		!strings.Contains(err.Error(), "refusing to use output directory") {
		t.Fatalf("expected refusal, got %v", err)
	}
	assertExists(t, filepath.Join(outDir, "keep.txt"))
}

func TestBuildBadTemplateKeepsExistingOutput(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(siteDir, "template.html"), "{{ .Broken")
	if _, err := Build(siteDir, outDir, false, "static"); err == nil {
		t.Fatal("expected template parse error")
	}
	assertExists(t, filepath.Join(outDir, "index.html"))
}

// stagingLeftovers lists hidden press-new/press-old siblings of out.
func stagingLeftovers(t *testing.T, out string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(out), "."+filepath.Base(out)+".press-*"))
	if err != nil {
		t.Fatal(err)
	}
	return matches
}

// A build that fails while rendering must not touch the previous output.
func TestBuildFailureKeepsPreviousOutput(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if err := page.Create(siteDir, "old", []byte("# Old\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}

	// Parses fine, fails when a page is rendered (after the old code had cleaned).
	writeTestFile(t, filepath.Join(siteDir, "template.html"), `{{ template "missing" . }}`)
	if err := os.Remove(filepath.Join(siteDir, "pages", "old.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Build(siteDir, outDir, false, "static"); err == nil {
		t.Fatal("expected a render error")
	}

	after, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("previous index.html must survive a failed build: %v", err)
	}
	if string(after) != string(before) {
		t.Error("previous index.html was modified by a failed build")
	}
	assertExists(t, filepath.Join(outDir, "old.html"))
	assertExists(t, filepath.Join(outDir, outputMarker))
	if left := stagingLeftovers(t, outDir); len(left) != 0 {
		t.Errorf("failed build left staging dirs behind: %v", left)
	}
}

// While a rebuild runs, a reader must never see an empty or truncated page.
// A brief "not found" while the new dir is swapped in is accepted (see swapOutput).
func TestBuildNeverShowsPartialOutput(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	for i := 0; i < 6; i++ {
		body := strings.Repeat("A paragraph of text.\n\n", 5000)
		if err := page.Create(siteDir, fmt.Sprintf("big%d", i), []byte("# Big\n\n"+body)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(outDir, "big0.html"))
	if err != nil {
		t.Fatal(err)
	}

	stop := make(chan struct{})
	bad := make(chan string, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			got, err := os.ReadFile(filepath.Join(outDir, "big0.html"))
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			if err != nil || string(got) != string(want) {
				select {
				case bad <- fmt.Sprintf("read err=%v len=%d want len=%d", err, len(got), len(want)):
				default:
				}
				return
			}
		}
	}()
	for i := 0; i < 8; i++ {
		if _, err := Build(siteDir, outDir, false, "static"); err != nil {
			t.Fatal(err)
		}
	}
	close(stop)
	wg.Wait()
	select {
	case msg := <-bad:
		t.Fatalf("reader saw a partial page: %s", msg)
	default:
	}
}

func TestBuildReturnsFinalPaths(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	if err := page.Create(siteDir, "about", []byte("# About\n")); err != nil {
		t.Fatal(err)
	}
	built, err := Build(siteDir, outDir, false, "static")
	if err != nil {
		t.Fatal(err)
	}
	if len(built) == 0 {
		t.Fatal("expected built paths")
	}
	for _, p := range built {
		if !strings.HasPrefix(p, outDir+string(filepath.Separator)) {
			t.Errorf("%s is not under the real output dir %s", p, outDir)
		}
		assertExists(t, p)
	}
}

func TestBuildRemovesStaleStaging(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	stale := filepath.Join(siteDir, ".dist.press-new")
	staleOld := filepath.Join(siteDir, ".dist.press-old")
	for _, d := range []string{stale, staleOld} {
		writeTestFile(t, filepath.Join(d, "junk.txt"), "x")
		writeTestFile(t, filepath.Join(d, outputMarker), "x")
	}

	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{stale, staleOld} {
		if _, err := os.Stat(d); err == nil {
			t.Errorf("leftover %s from a crashed build should be removed", filepath.Base(d))
		}
	}
	assertExists(t, filepath.Join(outDir, "index.html"))
}

// A dir that merely has our reserved name, but is not press output, is never deleted.
func TestBuildRefusesUnownedStagingPath(t *testing.T) {
	for _, name := range []string{".dist.press-new", ".dist.press-old"} {
		t.Run(name, func(t *testing.T) {
			siteDir, outDir := newAssetSite(t)
			if _, err := Build(siteDir, outDir, false, "static"); err != nil {
				t.Fatal(err)
			}
			mine := filepath.Join(siteDir, name, "mine.txt")
			writeTestFile(t, mine, "x")

			err := func() error { _, err := Build(siteDir, outDir, false, "static"); return err }()
			if err == nil || !strings.Contains(err.Error(), name) {
				t.Fatalf("expected an error naming %s, got %v", name, err)
			}
			assertExists(t, mine)
			assertExists(t, filepath.Join(outDir, "index.html"))
		})
	}
}

func TestBuildSymlinkedOutputStaysSymlink(t *testing.T) {
	siteDir, _ := newAssetSite(t)
	target := filepath.Join(t.TempDir(), "real-dist")
	link := filepath.Join(siteDir, "dist")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	for i := 0; i < 2; i++ {
		if _, err := Build(siteDir, link, false, "static"); err != nil {
			t.Fatalf("build %d: %v", i+1, err)
		}
	}
	fi, err := os.Lstat(link)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("output must stay a symlink, got %v err=%v", fi, err)
	}
	assertExists(t, filepath.Join(target, "index.html"))
}

func TestBuildRefusedDirLeavesNoStaging(t *testing.T) {
	siteDir, outDir := newAssetSite(t)
	writeTestFile(t, filepath.Join(outDir, "keep.txt"), "mine")
	if _, err := Build(siteDir, outDir, false, "static"); err == nil {
		t.Fatal("expected refusal")
	}
	if left := stagingLeftovers(t, outDir); len(left) != 0 {
		t.Errorf("refused build created staging dirs: %v", left)
	}
	assertExists(t, filepath.Join(outDir, "keep.txt"))
}

// If removing the replaced output fails halfway, what is left must be marked, or
// every later build would refuse it. The first build after upgrading replaces an
// unmarked legacy dist/, so that is the case that matters.
func TestBuildLeavesMarkedLeftoverWhenCleanupFails(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs a non-root POSIX user to make a directory undeletable")
	}
	siteDir, outDir := newAssetSite(t)
	writeTestFile(t, filepath.Join(outDir, "index.html"), "<html>")
	locked := filepath.Join(outDir, "blog")
	writeTestFile(t, filepath.Join(locked, "post.html"), "<html>")
	if err := os.Chmod(locked, 0500); err != nil {
		t.Fatal(err)
	}
	leftover := previousPath(outDir)
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(leftover, "blog"), 0755) })

	if _, err := Build(siteDir, outDir, false, "static"); err != nil {
		t.Fatalf("legacy dist should still build: %v", err)
	}
	assertExists(t, filepath.Join(outDir, "index.html"))
	if _, err := os.Stat(leftover); err != nil {
		t.Skip("cleanup succeeded, nothing left to check")
	}
	assertExists(t, filepath.Join(leftover, outputMarker))
}
