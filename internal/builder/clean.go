package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ChristianKreuzberger/press/internal/page"
)

var errUnsafeOutputDir = fmt.Errorf("refusing to use output directory")

// contains reports whether child is parent or lives inside it.
func contains(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// resolve makes p absolute and follows symlinks where the path exists, so a
// symlinked output dir can't hide that it points at the site.
func resolve(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	if r, err := filepath.EvalSymlinks(abs); err == nil {
		return r, nil
	}
	return abs, nil
}

// cleanOutputDir empties outputDir so files from deleted, renamed or
// newly-drafted pages don't linger and get deployed. The directory itself is
// kept because `press serve` is already serving it. Because this deletes
// recursively, it refuses any directory that holds, or sits inside, site sources.
func cleanOutputDir(siteDir, outputDir, staticDir string) error {
	out, err := resolve(outputDir)
	if err != nil {
		return fmt.Errorf("resolving output dir: %w", err)
	}
	site, err := resolve(siteDir)
	if err != nil {
		return fmt.Errorf("resolving site dir: %w", err)
	}
	static, err := validateStaticDirName(staticDir)
	if err != nil {
		return err
	}
	protected := []string{
		site,
		page.PagesDir(site),
		filepath.Join(site, "template.html"),
		filepath.Join(site, static),
	}
	for _, p := range protected {
		// Either direction is unsafe: out holds the source, or out is inside it.
		if contains(out, p) || contains(p, out) && p != site {
			return fmt.Errorf("%w %s: it overlaps site source %s", errUnsafeOutputDir, outputDir, p)
		}
	}

	entries, err := os.ReadDir(out)
	if err != nil {
		return fmt.Errorf("reading output directory: %w", err)
	}
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(out, e.Name())); err != nil {
			return fmt.Errorf("cleaning output directory: %w", err)
		}
	}
	return nil
}
