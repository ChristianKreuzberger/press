// Package section provides operations for managing press sections.
// A section is a subdirectory under pages/ that groups related pages together.
// Every section must contain an index.md file that acts as its landing page.
package section

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

// ErrSectionExists is returned when a section with the given name already exists.
var ErrSectionExists = errors.New("section already exists")

// ErrSectionNotFound is returned when a section with the given name does not exist.
var ErrSectionNotFound = errors.New("section not found")

// ErrInvalidName is returned when a section name contains illegal characters.
var ErrInvalidName = errors.New("invalid section name")

// Section represents a group of pages backed by a subdirectory under pages/.
type Section struct {
	Name      string // directory name (no slashes)
	Path      string // absolute path to the section directory
	IndexPath string // absolute path to the section's index.md
}

// Page represents a page within a section.
type Page struct {
	Name  string // file name without the .md extension
	Path  string // absolute path to the .md file
	Draft bool   // true when the page has draft: true in its frontmatter
}

// validateName rejects names that are empty, equal to "." or "..", or that
// contain a path separator — any of which could cause filesystem operations to
// escape the pages/ directory.
func validateName(name string) error {
	if name == "" || name == "." || name == ".." {
		return fmt.Errorf("%w: %q", ErrInvalidName, name)
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("%w: %q (must not contain path separators)", ErrInvalidName, name)
	}
	return nil
}

// sectionsBaseDir returns the pages/ directory within siteDir.
func sectionsBaseDir(siteDir string) string {
	return filepath.Join(siteDir, pagesDir)
}

// sectionDir returns the path to a specific section directory.
func sectionDir(siteDir, name string) string {
	return filepath.Join(sectionsBaseDir(siteDir), name)
}

// List returns all sections found in the pages/ directory of siteDir.
// A section is a subdirectory that contains an index.md file.
// It returns nil (not an error) when the pages directory does not exist yet.
func List(siteDir string) ([]Section, error) {
	base := sectionsBaseDir(siteDir)
	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var sections []Section
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(base, e.Name())
		indexPath := filepath.Join(dir, "index.md")
		if _, err := os.Stat(indexPath); err != nil {
			if os.IsNotExist(err) {
				// Directory without index.md is not a valid section.
				continue
			}
			return nil, err
		}
		sections = append(sections, Section{
			Name:      e.Name(),
			Path:      dir,
			IndexPath: indexPath,
		})
	}
	return sections, nil
}

// Create creates a new section with the given name and index content.
// It returns an error if a section with that name already exists.
func Create(siteDir, name string, content []byte) error {
	if err := validateName(name); err != nil {
		return err
	}
	dir := sectionDir(siteDir, name)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("%w: %q", ErrSectionExists, name)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "index.md"), content, 0644)
}

// Delete removes the section directory and all its contents.
func Delete(siteDir, name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	dir := sectionDir(siteDir, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("%w: %q", ErrSectionNotFound, name)
	}
	return os.RemoveAll(dir)
}

// Update replaces the content of an existing section's index.md.
func Update(siteDir, name string, content []byte) error {
	if err := validateName(name); err != nil {
		return err
	}
	dir := sectionDir(siteDir, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("%w: %q", ErrSectionNotFound, name)
	}
	indexPath := filepath.Join(dir, "index.md")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return fmt.Errorf("%w: %q", ErrSectionNotFound, name)
	}
	return os.WriteFile(indexPath, content, 0644)
}

// Rename renames the section from oldName to newName.
// It updates the title and updated_at in the section's index.md.
func Rename(siteDir, oldName, newName string, now time.Time) error {
	if err := validateName(oldName); err != nil {
		return err
	}
	if err := validateName(newName); err != nil {
		return err
	}

	oldDir := sectionDir(siteDir, oldName)
	newDir := sectionDir(siteDir, newName)

	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		return fmt.Errorf("%w: %q", ErrSectionNotFound, oldName)
	}
	if _, err := os.Stat(newDir); err == nil {
		return fmt.Errorf("%w: %q", ErrSectionExists, newName)
	}

	indexPath := filepath.Join(oldDir, "index.md")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}
	content, err = frontmatter.SetField(content, "title", frontmatter.Humanize(newName))
	if err != nil {
		return fmt.Errorf("rename section: %w", err)
	}
	content, err = frontmatter.SetField(content, "updated_at", now.UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("rename section: %w", err)
	}
	if err := os.Rename(oldDir, newDir); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(newDir, "index.md"), content, 0644); err != nil {
		// Roll back the directory rename to avoid a half-applied state.
		_ = os.Rename(newDir, oldDir)
		return err
	}
	return nil
}

// ListPages returns all pages found inside a section directory, including index.md.
// Pages are returned in directory order.
func ListPages(siteDir, sectionName string) ([]Page, error) {
	if err := validateName(sectionName); err != nil {
		return nil, err
	}
	dir := sectionDir(siteDir, sectionName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %q", ErrSectionNotFound, sectionName)
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
