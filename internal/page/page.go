// Package page provides operations for managing press pages.
package page

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const pagesDir = "pages"

// ErrPageExists is returned when a page with the given name already exists.
var ErrPageExists = errors.New("page already exists")

// ErrPageNotFound is returned when a page with the given name does not exist.
var ErrPageNotFound = errors.New("page not found")

// Page represents a single content page backed by a Markdown file.
type Page struct {
	Name string // file name without the .md extension
	Path string // absolute path to the .md file
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
			pages = append(pages, Page{
				Name: name,
				Path: filepath.Join(dir, e.Name()),
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
		return fmt.Errorf("invalid page name: %q", name)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: %q", ErrPageExists, name)
	}
	return os.WriteFile(path, content, 0644)
}

// Delete removes the page with the given name.
// name may contain forward slashes (e.g. "blog/my-post").
func Delete(siteDir, name string) error {
	dir := PagesDir(siteDir)
	path := filepath.Join(dir, filepath.FromSlash(name)+".md")
	// Prevent path traversal.
	cleanDir := filepath.Clean(dir) + string(filepath.Separator)
	if !strings.HasPrefix(filepath.Clean(path), cleanDir) {
		return fmt.Errorf("invalid page name: %q", name)
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %q", ErrPageNotFound, name)
		}
		return err
	}
	return nil
}

// Update replaces the content of an existing page.
func Update(siteDir, name string, content []byte) error {
	path := filepath.Join(PagesDir(siteDir), name+".md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%w: %q", ErrPageNotFound, name)
	}
	return os.WriteFile(path, content, 0644)
}
