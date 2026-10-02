package builder

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ChristianKreuzberger/press/internal/page"
)

var errUnsafeOutputDir = errors.New("refusing to use output directory")

var errNotPressOutput = errors.New("not press output")

// outputMarker is written into every output dir press builds. A non-empty
// directory is only ever emptied if it carries this marker (or, for dist/
// folders made by older press versions, looks like press output).
const outputMarker = ".press-output"

// contains reports whether child is parent or lives inside it.
func contains(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// resolve makes p absolute and follows symlinks. For a path that does not
// exist yet, it resolves the nearest existing ancestor and re-appends the rest,
// so `link/newdir` with link -> site is still seen as inside the site.
func resolve(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	existing, rest := abs, ""
	for {
		if r, err := filepath.EvalSymlinks(existing); err == nil {
			return filepath.Join(r, rest), nil
		}
		parent := filepath.Dir(existing)
		if parent == existing {
			return abs, nil
		}
		rest = filepath.Join(filepath.Base(existing), rest)
		existing = parent
	}
}

// checkOutputDir is the whole safety check for a recursive delete of outputDir.
// It has no side effects, so Build can run it before creating anything. It
// returns the resolved output path. Two independent guards:
//  1. guardSources: outputDir must not overlap the site sources.
//  2. checkOwnership: a non-empty dir must be recognisably press output.
//
// Guard 1 duplicates validateOutputDir from PR #83; once that lands, replace
// guardSources (and contains/resolve) with it plus a template.html check.
func checkOutputDir(siteDir, outputDir, staticDir string) (string, error) {
	out, err := resolve(outputDir)
	if err != nil {
		return "", fmt.Errorf("resolving output dir: %w", err)
	}
	if err := guardSources(siteDir, out, outputDir, staticDir); err != nil {
		return "", err
	}
	if err := checkOwnership(siteDir, out, outputDir, staticDir); err != nil {
		return "", err
	}
	return out, nil
}

func guardSources(siteDir, out, outputDir, staticDir string) error {
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
	return nil
}

// checkOwnership allows a missing dir, an empty dir, a dir with the marker, or
// an unmarked dir that looks like output of an older press version.
func checkOwnership(siteDir, out, outputDir, staticDir string) error {
	info, err := os.Stat(out)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading output directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w %s: not a directory", errUnsafeOutputDir, outputDir)
	}
	entries, err := os.ReadDir(out)
	if err != nil {
		return fmt.Errorf("reading output directory: %w", err)
	}
	if len(entries) == 0 {
		return nil
	}
	if fi, err := os.Lstat(filepath.Join(out, outputMarker)); err == nil && fi.Mode().IsRegular() {
		return nil
	}
	if looksLikeLegacyOutput(siteDir, out, staticDir) {
		return nil
	}
	return fmt.Errorf("%w %s: it is not empty and was not created by press (no %s file). "+
		"Use an empty or new directory, or delete it yourself first", errUnsafeOutputDir, outputDir, outputMarker)
}

// looksLikeLegacyOutput reports whether every file in out is something press
// would have written before markers existed: a .html file, or a file that also
// exists in pages/ or the static dir at the same relative path. Symlinks and
// anything else (a .git dir, a README, notes) make it "not ours".
func looksLikeLegacyOutput(siteDir, out, staticDir string) bool {
	roots := []string{page.PagesDir(siteDir), filepath.Join(siteDir, staticDir)}
	err := filepath.WalkDir(out, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return errNotPressOutput
		}
		if strings.HasSuffix(d.Name(), ".html") {
			return nil
		}
		rel, err := filepath.Rel(out, p)
		if err != nil {
			return err
		}
		for _, r := range roots {
			if fi, err := os.Lstat(filepath.Join(r, rel)); err == nil && fi.Mode().IsRegular() {
				return nil
			}
		}
		return errNotPressOutput
	})
	return err == nil
}

// cleanOutputDir empties out (already validated by checkOutputDir) so files from
// deleted, renamed or newly-drafted pages don't linger and get deployed, then
// writes the marker. The directory itself is kept because `press serve` is
// already serving it. It keeps going after a failure and reports all of them.
func cleanOutputDir(out string) error {
	entries, err := os.ReadDir(out)
	if err != nil {
		return fmt.Errorf("reading output directory: %w", err)
	}
	var errs []error
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(out, e.Name())); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("cleaning output directory (it may now be partly emptied): %w", errors.Join(errs...))
	}
	return os.WriteFile(filepath.Join(out, outputMarker), []byte("generated by press; safe to delete\n"), 0644) //nolint:gosec // output must be world-readable
}
