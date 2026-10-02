// Package page provides operations for managing press pages.
package page

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ChristianKreuzberger/press/internal/frontmatter"
)

const pagesDir = "pages"

// ErrPageExists is returned when a page with the given name already exists.
var ErrPageExists = errors.New("page already exists")

// ErrInvalidName is returned when the page name is malformed or would escape the pages directory.
var ErrInvalidName = errors.New("invalid page name")

// ErrNameConflict is returned when the name clashes with a section or with a
// page that already occupies part of the path.
var ErrNameConflict = errors.New("name conflicts with an existing page or section")

// ErrPageNotFound is returned when a page with the given name does not exist.
var ErrPageNotFound = errors.New("page not found")

// Page represents a single content page backed by a Markdown file.
type Page struct {
	Name  string // file name without the .md extension
	Path  string // absolute path to the .md file
	Draft bool   // true when the page has draft: true in its frontmatter
}

// PagesDir returns the path to the pages directory within siteDir.
func PagesDir(siteDir string) string {
	return filepath.Join(siteDir, pagesDir)
}

// List returns all pages found in the pages directory of siteDir.
// It returns nil (not an error) when the pages directory does not exist yet.
func List(siteDir string) ([]Page, error) {
	dir := PagesDir(siteDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var pages []Page
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			name := strings.TrimSuffix(e.Name(), ".md")
			path := filepath.Join(dir, e.Name())
			draft, err := frontmatter.ParseDraftFromFile(path)
			if err != nil {
				return nil, err
			}
			pages = append(pages, Page{
				Name:  name,
				Path:  path,
				Draft: draft,
			})
		}
	}
	return pages, nil
}

// validateName checks every "/"-separated segment: it must be non-empty, not
// "." or "..", and free of backslashes (a separator on Windows) and NUL.
// This also rejects empty names, trailing slashes and absolute paths.
func validateName(name string) error {
	for _, seg := range strings.Split(name, "/") {
		if seg == "" || seg == "." || seg == ".." || strings.ContainsAny(seg, "\\:\x00") {
			return fmt.Errorf("%w: %q", ErrInvalidName, name)
		}
	}
	return nil
}

// maxNameSegments is the deepest layout the builder publishes: page or
// section/page. Deeper files would be silently ignored, so they are refused.
const maxNameSegments = 2

// validateNewName is validateName plus the depth limit, for names that are
// about to be written. Existing deep files can still be deleted or updated.
func validateNewName(name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	if strings.Count(name, "/")+1 > maxNameSegments {
		return fmt.Errorf("%w: %q (use page or section/page; deeper nesting is not built)", ErrInvalidName, name)
	}
	return nil
}

// Skipped returns the pages/-relative paths of .md files that the build
// ignores: files nested deeper than section/page, and files in a section
// directory that has no index.md. Hidden directories are not scanned.
func Skipped(siteDir string) ([]string, error) {
	base := PagesDir(siteDir)
	var skipped []string
	err := filepath.WalkDir(base, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) && p == base {
				return nil
			}
			return err
		}
		if d.IsDir() {
			if p != base && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() || !strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			return nil
		}
		rel, err := filepath.Rel(base, p)
		if err != nil {
			return err
		}
		segs := strings.Count(filepath.ToSlash(rel), "/") + 1
		if segs > maxNameSegments {
			skipped = append(skipped, filepath.ToSlash(rel))
			return nil
		}
		if segs == maxNameSegments {
			if _, err := os.Stat(filepath.Join(filepath.Dir(p), "index.md")); err != nil {
				skipped = append(skipped, filepath.ToSlash(rel))
			}
		}
		return nil
	})
	return skipped, err
}

// pagePath validates name and returns the absolute path of its .md file.
func pagePath(siteDir, name string) (string, error) {
	if err := validateName(name); err != nil {
		return "", err
	}
	return filepath.Join(PagesDir(siteDir), filepath.FromSlash(name)+".md"), nil
}

// checkNoConflict reports ErrNameConflict when creating the page name would be
// ambiguous: a section directory already uses the name, or a parent segment is
// already a page (e.g. about.md next to about/team.md).
func checkNoConflict(siteDir, name string) error {
	dir := PagesDir(siteDir)
	segs := strings.Split(name, "/")
	for i := 1; i < len(segs); i++ {
		parent := filepath.Join(dir, filepath.FromSlash(strings.Join(segs[:i], "/"))+".md")
		if _, err := os.Lstat(parent); err == nil {
			return fmt.Errorf("%w: %q (page %q exists)", ErrNameConflict, name, strings.Join(segs[:i], "/"))
		}
	}
	if _, err := os.Lstat(filepath.Join(dir, filepath.FromSlash(name))); err == nil {
		return fmt.Errorf("%w: %q (a section with that name exists)", ErrNameConflict, name)
	}
	return nil
}

// Create creates a new page with the given name and content.
// name is either "page" or "section/page" (e.g. "blog/my-post"); deeper
// names are rejected because the builder does not publish them.
// It returns an error if a page with that name already exists.
func Create(siteDir, name string, content []byte) error {
	dir := PagesDir(siteDir)
	if err := validateNewName(name); err != nil {
		return err
	}
	path, err := pagePath(siteDir, name)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("%w: %q", ErrPageExists, name)
	}
	if err := checkNoConflict(siteDir, name); err != nil {
		return err
	}
	err = writeNew(dir, path, content, 0644)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("%w: %q", ErrPageExists, name)
	}
	return err
}

// writeNew creates path with O_EXCL so it never overwrites an existing file,
// creating missing parent directories first. On failure it removes the file
// and the directories it created, so no partial state is left behind.
// Cleanup is non-recursive: it never deletes anything this call did not create.
// perm is subject to the process umask.
func writeNew(baseDir, path string, content []byte, perm os.FileMode) error {
	var created []string // directories this call created, outermost first
	var missing []string
	// baseDir itself may not exist yet, so it is included in the walk.
	for d := filepath.Dir(path); d != filepath.Dir(filepath.Clean(baseDir)); d = filepath.Dir(d) {
		if _, err := os.Stat(d); err == nil {
			break
		}
		missing = append(missing, d)
	}
	cleanup := func() {
		for i := len(created) - 1; i >= 0; i-- {
			if os.Remove(created[i]) != nil {
				return
			}
		}
	}
	for i := len(missing) - 1; i >= 0; i-- {
		err := os.Mkdir(missing[i], 0750)
		if err == nil {
			created = append(created, missing[i])
		} else if !errors.Is(err, os.ErrExist) {
			cleanup()
			return err
		}
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm) //nolint:gosec // callers validate that path stays inside the pages dir
	if err != nil {
		cleanup()
		return err
	}
	if _, err = f.Write(content); err == nil {
		err = f.Close()
	} else {
		_ = f.Close()
	}
	if err != nil {
		_ = os.Remove(path)
		cleanup()
		return err
	}
	return nil
}

// removeEmptyParents removes empty directories from path's parent up to, but
// not including, baseDir. It stops at the first non-empty directory.
func removeEmptyParents(baseDir, path string) {
	stop := filepath.Clean(baseDir)
	for d := filepath.Dir(path); d != stop && strings.HasPrefix(d, stop); d = filepath.Dir(d) {
		if err := os.Remove(d); err != nil {
			return
		}
	}
}

// Delete removes the page with the given name.
// name may contain forward slashes (e.g. "blog/my-post").
func Delete(siteDir, name string) error {
	dir := PagesDir(siteDir)
	path, err := pagePath(siteDir, name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %q", ErrPageNotFound, name)
		}
		return err
	}
	removeEmptyParents(dir, path)
	return nil
}

// Update replaces the content of an existing page.
func Update(siteDir, name string, content []byte) error {
	path, err := pagePath(siteDir, name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%w: %q", ErrPageNotFound, name)
	}
	return os.WriteFile(path, content, 0644)
}

// Rename renames the page from oldName to newName.
// It updates the title in the frontmatter to the humanised form of newName,
// and sets updated_at to now.
func Rename(siteDir, oldName, newName string, now time.Time) error {
	dir := PagesDir(siteDir)
	oldPath, err := pagePath(siteDir, oldName)
	if err != nil {
		return err
	}
	if err := validateNewName(newName); err != nil {
		return err
	}
	newPath, err := pagePath(siteDir, newName)
	if err != nil {
		return err
	}
	// Renaming an index page away would leave the site (or section) without a landing page.
	if path.Base(oldName) == "index" {
		return fmt.Errorf("%w: %q (the home page cannot be renamed)", ErrInvalidName, oldName)
	}

	info, err := os.Stat(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %q", ErrPageNotFound, oldName)
		}
		return err
	}
	if _, err := os.Lstat(newPath); err == nil {
		return fmt.Errorf("%w: %q", ErrPageExists, newName)
	}
	if err := checkNoConflict(siteDir, newName); err != nil {
		return err
	}

	content, err := os.ReadFile(oldPath) //nolint:gosec // oldPath is validated to stay inside the pages dir above
	if err != nil {
		return err
	}
	content, err = frontmatter.SetField(content, "title", frontmatter.Humanize(newName))
	if err != nil {
		return fmt.Errorf("rename page: %w", err)
	}
	content, err = frontmatter.SetField(content, "updated_at", now.UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("rename page: %w", err)
	}
	// O_EXCL guards against a file appearing at newPath since the check above.
	if err := writeNew(dir, newPath, content, info.Mode().Perm()); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("%w: %q", ErrPageExists, newName)
		}
		return err
	}
	// Keep the original mode exactly (writeNew is subject to umask).
	if err := os.Chmod(newPath, info.Mode().Perm()); err != nil {
		return rollbackRename(dir, newPath, fmt.Errorf("rename page: %w", err))
	}
	if err := os.Remove(oldPath); err != nil {
		return rollbackRename(dir, newPath, err)
	}
	removeEmptyParents(dir, oldPath)
	return nil
}

// rollbackRename removes the half-written newPath so both files aren't left
// behind. If that fails too, the returned error reports both problems.
func rollbackRename(baseDir, newPath string, cause error) error {
	if err := os.Remove(newPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("%w; rollback failed, %q may be left behind: %v", cause, newPath, err)
	}
	removeEmptyParents(baseDir, newPath)
	return cause
}
