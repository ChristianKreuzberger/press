package builder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ChristianKreuzberger/press/internal/page"
)

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
	cases := map[string]string{
		"site dir":     siteDir,
		"parent dir":   filepath.Dir(siteDir),
		"pages dir":    filepath.Join(siteDir, "pages"),
		"static dir":   filepath.Join(siteDir, "static"),
		"inside pages": filepath.Join(siteDir, "pages", "out"),
	}
	for name, out := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Build(siteDir, out, false, "static"); err == nil {
				t.Fatalf("Build with output %s should fail", out)
			}
			if _, err := os.Stat(filepath.Join(siteDir, "pages", "index.md")); err != nil {
				t.Fatal("source page must survive a refused build")
			}
		})
	}
}
