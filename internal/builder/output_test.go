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
		{"template file", "template.html", "static"},
		{"inside template file", "template.html/out", "static"},
		{"git dir", ".git", "static"},
		{"nested static name contains output", "assets/sub/out", "assets/sub"},
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

func TestBuildRejectsOutputThroughDanglingSymlink(t *testing.T) {
	siteDir := t.TempDir()
	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}
	// pages/gen does not exist yet, so the link dangles.
	link := filepath.Join(siteDir, "dist")
	if err := os.Symlink(filepath.Join(siteDir, "pages", "gen"), link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	// Call the validator directly: Build would also fail later (MkdirAll on a
	// dangling link), which would mask a validation gap.
	if err := validateOutputDir(siteDir, link, "static"); err == nil {
		t.Fatal("output dangling-symlinked into pages/ should be rejected")
	}
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "gen")); err == nil {
		t.Error("build wrote through the dangling symlink into pages/")
	}
}

func TestBuildRejectsSymlinkedStaticDir(t *testing.T) {
	siteDir := t.TempDir()
	if err := page.Create(siteDir, "index", []byte("# Home\n")); err != nil {
		t.Fatal(err)
	}
	realDir := filepath.Join(siteDir, "real-assets")
	if err := os.MkdirAll(realDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(realDir, filepath.Join(siteDir, "static")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := Build(siteDir, filepath.Join(realDir, "out"), false, "static"); err == nil {
		t.Fatal("output inside the target of a symlinked static dir should be rejected")
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
