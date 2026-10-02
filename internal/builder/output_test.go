package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ChristianKreuzberger/press/internal/page"
)

func TestBuildRejectsUnsafeOutputDir(t *testing.T) {
	cases := []struct {
		name   string
		out    string // relative to siteDir
		static string
	}{
		{"site dir itself", ".", "static"},
		{"empty means site dir", "", "static"},
		{"parent of site dir", "..", "static"},
		{"pages dir", "pages", "static"},
		{"inside pages dir", "pages/out", "static"},
		{"static dir", "static", "static"},
		{"inside static dir", "static/out", "static"},
		{"static contains output", "dist", "dist"},
		{"unclean path to site dir", "dist/..", "static"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			siteDir := t.TempDir()
			if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(siteDir, "static"), 0755); err != nil {
				t.Fatal(err)
			}
			if _, err := Build(siteDir, filepath.Join(siteDir, tc.out), false, tc.static); err == nil {
				t.Fatalf("Build with output %q should fail", tc.out)
			}
			if _, err := os.Stat(filepath.Join(siteDir, "pages", "index.html")); err == nil {
				t.Error("HTML must not be written into the source tree")
			}
		})
	}
}

func TestBuildRejectsOutputThroughSymlink(t *testing.T) {
	siteDir := t.TempDir()
	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(siteDir, "out")
	if err := os.Symlink(filepath.Join(siteDir, "pages"), link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := Build(siteDir, link, false, "static"); err == nil {
		t.Fatal("output symlinked to pages/ should be rejected")
	}
}

func TestBuildAcceptsSafeOutputDirs(t *testing.T) {
	for _, out := range []string{"dist", "public/site", "../press-out-sibling"} {
		t.Run(out, func(t *testing.T) {
			parent := t.TempDir()
			siteDir := filepath.Join(parent, "site")
			if err := os.MkdirAll(siteDir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
				t.Fatal(err)
			}
			if _, err := Build(siteDir, filepath.Join(siteDir, out), false, "static"); err != nil {
				t.Fatalf("Build with output %q failed: %v", out, err)
			}
		})
	}
}
