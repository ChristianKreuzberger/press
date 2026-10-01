package builder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// manifestName is the file in the output dir that records every file press
// generated there. Cleanup only ever removes files listed in it, so files the
// user put into the output dir themselves are never touched.
const manifestName = ".press-manifest"

// cleanStale removes files that an earlier build generated but this one did
// not, then records the files of this build in the manifest.
//
// It does nothing when the output dir is the site dir, one of its ancestors,
// or inside the pages or static dir: in those layouts generated and source
// files can be the same paths, and deleting would destroy the site.
func cleanStale(siteDir, outputDir, staticDir string, generated []string) error {
	if !outputIsSeparate(siteDir, outputDir, staticDir) {
		return nil
	}

	current := make(map[string]bool, len(generated))
	var rels []string
	for _, p := range generated {
		rel, err := filepath.Rel(outputDir, p)
		if err != nil || !filepath.IsLocal(rel) {
			continue
		}
		rel = filepath.ToSlash(rel)
		if !validManifestEntry(rel) || current[rel] {
			continue
		}
		current[rel] = true
		rels = append(rels, rel)
	}
	sort.Strings(rels)

	for _, rel := range readManifest(outputDir) {
		if !current[rel] {
			removeGenerated(outputDir, rel)
		}
	}
	return writeManifest(outputDir, rels)
}

// outputIsSeparate reports whether outputDir is a directory that holds only
// build output, i.e. it does not contain and is not inside the site sources.
func outputIsSeparate(siteDir, outputDir, staticDir string) bool {
	site, err := filepath.Abs(siteDir)
	if err != nil {
		return false
	}
	// Resolve symlinks so differently spelled paths to the same place compare equal.
	if p, err := filepath.EvalSymlinks(site); err == nil {
		site = p
	}
	if p, err := filepath.EvalSymlinks(outputDir); err == nil {
		outputDir = p
	}
	sources := []string{site, filepath.Join(site, "pages")}
	if name, err := validateStaticDirName(staticDir); err == nil {
		sources = append(sources, filepath.Join(site, name))
	}
	for _, src := range sources {
		// outputDir equal to or an ancestor of a source dir.
		if rel, err := filepath.Rel(outputDir, src); err != nil || filepath.IsLocal(rel) {
			return false
		}
		// outputDir equal to or inside a source dir (the site dir itself is
		// the parent of the normal output dir, so skip it here).
		if src == site {
			continue
		}
		if rel, err := filepath.Rel(src, outputDir); err != nil || filepath.IsLocal(rel) {
			return false
		}
	}
	return true
}

// validManifestEntry rejects entries that could not have been produced by a
// normal build or that make the line-based format ambiguous.
func validManifestEntry(rel string) bool {
	if rel == manifestName || !filepath.IsLocal(filepath.FromSlash(rel)) || strings.Contains(rel, "\\") {
		return false
	}
	return !strings.ContainsFunc(rel, unicode.IsControl)
}

// readManifest returns the valid entries (slash-separated paths relative to
// outputDir) of the manifest. A missing or unreadable manifest yields none.
func readManifest(outputDir string) []string {
	path := filepath.Join(outputDir, manifestName)
	if info, err := os.Lstat(path); err != nil || !info.Mode().IsRegular() {
		return nil
	}
	data, err := os.ReadFile(path) //nolint:gosec // fixed filename inside the output dir
	if err != nil {
		return nil
	}
	var rels []string
	for _, line := range strings.Split(string(data), "\n") {
		if validManifestEntry(line) {
			rels = append(rels, line)
		}
	}
	return rels
}

// writeManifest atomically replaces the manifest with the given entries.
func writeManifest(outputDir string, rels []string) error {
	var b strings.Builder
	for _, rel := range rels {
		b.WriteString(rel)
		b.WriteByte('\n')
	}
	tmp, err := os.CreateTemp(outputDir, manifestName+".tmp-*")
	if err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }() // no-op after a successful rename
	if _, err := tmp.WriteString(b.String()); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("writing manifest: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil { //nolint:gosec // output must be world-readable for web servers
		return fmt.Errorf("writing manifest: %w", err)
	}
	if err := os.Rename(tmp.Name(), filepath.Join(outputDir, manifestName)); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}
	return nil
}

// removeGenerated deletes the regular file outputDir/rel and then any parent
// directories that became empty, never outputDir itself. Removal is best
// effort: anything unexpected (symlinks, directories, paths that resolve
// outside outputDir) is left alone.
func removeGenerated(outputDir, rel string) {
	root, err := filepath.EvalSymlinks(outputDir)
	if err != nil {
		return
	}
	full := filepath.Join(outputDir, filepath.FromSlash(rel))
	parent, err := filepath.EvalSymlinks(filepath.Dir(full))
	if err != nil {
		return
	}
	if r, err := filepath.Rel(root, parent); err != nil || !filepath.IsLocal(r) {
		return
	}
	target := filepath.Join(parent, filepath.Base(full))
	info, err := os.Lstat(target)
	if err != nil || !info.Mode().IsRegular() {
		return
	}
	if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
		return
	}
	for dir := parent; dir != root; dir = filepath.Dir(dir) {
		// os.Remove fails on non-empty directories, which is what we want.
		if os.Remove(dir) != nil {
			break
		}
	}
}
