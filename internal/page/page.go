// Package page provides operations for managing press pages.
package page

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ChristianKreuzberger/press/internal/frontmatter"
)

const pagesDir = "pages"

// ErrPageExists is returned when a page with the given name already exists.
var ErrPageExists = errors.New("page already exists")

// ErrInvalidName is returned when the page name would escape the pages directory.
var ErrInvalidName = errors.New("invalid page name")

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

// Create creates a new page with the given name and content.
// name may contain forward slashes to place the page inside sub-sections
// (e.g. "blog/my-post" or "blog/2026/my-post").
// It returns an error if a page with that name already exists.
func Create(siteDir, name string, content []byte) error {
	dir := PagesDir(siteDir)
	path := filepath.Join(dir, filepath.FromSlash(name)+".md")
	// Prevent path traversal: resolved path must remain inside pages dir.
	cleanDir := filepath.Clean(dir) + string(filepath.Separator)
	if !strings.HasPrefix(filepath.Clean(path), cleanDir) {
		return fmt.Errorf("%w: %q", ErrInvalidName, name)
	}
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("%w: %q", ErrPageExists, name)
	}
	err := writeNew(dir, path, content, 0644)
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
	path := filepath.Join(dir, filepath.FromSlash(name)+".md")
	// Prevent path traversal.
	cleanDir := filepath.Clean(dir) + string(filepath.Separator)
	if !strings.HasPrefix(filepath.Clean(path), cleanDir) {
		return fmt.Errorf("%w: %q", ErrInvalidName, name)
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
	dir := PagesDir(siteDir)
	path := filepath.Join(dir, filepath.FromSlash(name)+".md")
	// Prevent path traversal: resolved path must remain inside pages dir.
	cleanDir := filepath.Clean(dir) + string(filepath.Separator)
	if !strings.HasPrefix(filepath.Clean(path), cleanDir) {
		return fmt.Errorf("%w: %q", ErrInvalidName, name)
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
	cleanDir := filepath.Clean(dir) + string(filepath.Separator)

	oldPath := filepath.Join(dir, filepath.FromSlash(oldName)+".md")
	if !strings.HasPrefix(filepath.Clean(oldPath), cleanDir) {
		return fmt.Errorf("%w: %q", ErrInvalidName, oldName)
	}
	newPath := filepath.Join(dir, filepath.FromSlash(newName)+".md")
	if !strings.HasPrefix(filepath.Clean(newPath), cleanDir) {
		return fmt.Errorf("%w: %q", ErrInvalidName, newName)
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
