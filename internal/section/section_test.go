package section

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestListEmpty(t *testing.T) {
	dir := t.TempDir()
	sections, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 0 {
		t.Errorf("expected 0 sections, got %d", len(sections))
	}
}

func TestListNoPages(t *testing.T) {
	// pages/ directory does not exist yet
	dir := t.TempDir()
	sections, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if sections != nil {
		t.Errorf("expected nil when pages/ absent, got %v", sections)
	}
}

func TestListSkipsDirectoriesWithoutIndex(t *testing.T) {
	dir := t.TempDir()
	// Create a subdirectory without index.md — should not appear as section.
	noIndex := filepath.Join(dir, "pages", "noindex")
	if err := os.MkdirAll(noIndex, 0755); err != nil {
		t.Fatal(err)
	}

	sections, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 0 {
		t.Errorf("directory without index.md should not be listed as section, got %v", sections)
	}
}

func TestCreateAndList(t *testing.T) {
	dir := t.TempDir()

	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "docs", []byte("# Docs\n")); err != nil {
		t.Fatal(err)
	}

	sections, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	names := map[string]bool{}
	for _, s := range sections {
		names[s.Name] = true
	}
	if !names["blog"] || !names["docs"] {
		t.Errorf("unexpected section names: %v", names)
	}
}

func TestCreateDuplicate(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "blog", []byte("dup")); err == nil {
		t.Error("expected error creating duplicate section, got nil")
	}
}

func TestCreateWritesIndexMD(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	indexPath := filepath.Join(dir, "pages", "blog", "index.md")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("expected index.md to exist: %v", err)
	}
	if string(content) != "# Blog\n" {
		t.Errorf("unexpected index.md content: %q", content)
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	if err := Delete(dir, "blog"); err != nil {
		t.Fatal(err)
	}
	sections, _ := List(dir)
	if len(sections) != 0 {
		t.Errorf("expected 0 sections after delete, got %d", len(sections))
	}
}

func TestDeleteNotFound(t *testing.T) {
	dir := t.TempDir()
	if err := Delete(dir, "missing"); err == nil {
		t.Error("expected error deleting non-existent section, got nil")
	}
}

func TestUpdate(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Old\n")); err != nil {
		t.Fatal(err)
	}
	if err := Update(dir, "blog", []byte("# New\n")); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(filepath.Join(dir, "pages", "blog", "index.md"))
	if string(content) != "# New\n" {
		t.Errorf("update did not change content: %q", content)
	}
}

func TestUpdateNotFound(t *testing.T) {
	dir := t.TempDir()
	if err := Update(dir, "missing", []byte("x")); err == nil {
		t.Error("expected error updating non-existent section, got nil")
	}
}

func TestListPages(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	// Add extra pages to the section directory.
	blogDir := filepath.Join(dir, "pages", "blog")
	if err := os.WriteFile(filepath.Join(blogDir, "post-one.md"), []byte("# Post One\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "post-two.md"), []byte("# Post Two\n"), 0644); err != nil {
		t.Fatal(err)
	}

	pages, err := ListPages(dir, "blog")
	if err != nil {
		t.Fatal(err)
	}
	// Should include index, post-one, post-two.
	if len(pages) != 3 {
		t.Fatalf("expected 3 pages, got %d", len(pages))
	}
	names := map[string]bool{}
	for _, p := range pages {
		names[p.Name] = true
	}
	if !names["index"] || !names["post-one"] || !names["post-two"] {
		t.Errorf("unexpected page names: %v", names)
	}
}

func TestListPagesNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := ListPages(dir, "missing")
	if err == nil {
		t.Error("expected error listing pages of non-existent section, got nil")
	}
}

func TestValidateName(t *testing.T) {
	invalidNames := []string{
		"",
		".",
		"..",
		"sub/dir",
		"sub\\dir",
		"a/b/c",
	}
	for _, name := range invalidNames {
		if err := validateName(name); err == nil {
			t.Errorf("validateName(%q): expected error, got nil", name)
		}
	}

	validNames := []string{"blog", "my-section", "docs2", "section_name"}
	for _, name := range validNames {
		if err := validateName(name); err != nil {
			t.Errorf("validateName(%q): expected nil, got %v", name, err)
		}
	}
}

func TestCreateRejectsInvalidName(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"", "..", "a/b"} {
		if err := Create(dir, name, []byte("x")); err == nil {
			t.Errorf("Create with name %q should fail, got nil", name)
		}
	}
}

func TestDeleteRejectsInvalidName(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"", "..", "a/b"} {
		if err := Delete(dir, name); err == nil {
			t.Errorf("Delete with name %q should fail, got nil", name)
		}
	}
}

func TestUpdateRejectsInvalidName(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"", "..", "a/b"} {
		if err := Update(dir, name, []byte("x")); err == nil {
			t.Errorf("Update with name %q should fail, got nil", name)
		}
	}
}

func TestListPagesRejectsInvalidName(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"", "..", "a/b"} {
		if _, err := ListPages(dir, name); err == nil {
			t.Errorf("ListPages with name %q should fail, got nil", name)
		}
	}
}

func TestListWithFileInPagesDir(t *testing.T) {
	dir := t.TempDir()
	// A plain file (not a directory) inside pages/ should be silently skipped.
	if err := os.MkdirAll(filepath.Join(dir, "pages"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pages", "about.md"), []byte("# About\n"), 0644); err != nil {
		t.Fatal(err)
	}

	sections, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 0 {
		t.Errorf("expected 0 sections when pages/ only contains files, got %d", len(sections))
	}
}

func TestUpdateMissingIndexMD(t *testing.T) {
	dir := t.TempDir()
	// Create the section directory manually without an index.md.
	sectionPath := filepath.Join(dir, "pages", "blog")
	if err := os.MkdirAll(sectionPath, 0755); err != nil {
		t.Fatal(err)
	}

	err := Update(dir, "blog", []byte("# Blog\n"))
	if !errors.Is(err, ErrSectionNotFound) {
		t.Errorf("expected ErrSectionNotFound when index.md is absent, got %v", err)
	}
}

func TestListPagesDraftField(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	blogDir := filepath.Join(dir, "pages", "blog")
	if err := os.WriteFile(filepath.Join(blogDir, "post.md"), []byte("# Post\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blogDir, "wip.md"), []byte("---\ndraft: true\n---\n# WIP\n"), 0644); err != nil {
		t.Fatal(err)
	}

	pages, err := ListPages(dir, "blog")
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 3 {
		t.Fatalf("expected 3 pages, got %d", len(pages))
	}
	byName := map[string]Page{}
	for _, p := range pages {
		byName[p.Name] = p
	}
	post, ok := byName["post"]
	if !ok {
		t.Fatal("expected post page to be returned")
	}
	if post.Draft {
		t.Error("expected post to have Draft=false")
	}
	wip, ok := byName["wip"]
	if !ok {
		t.Fatal("expected wip page to be returned")
	}
	if !wip.Draft {
		t.Error("expected wip to have Draft=true")
	}
	index, ok := byName["index"]
	if !ok {
		t.Fatal("expected index page to be returned")
	}
	if index.Draft {
		t.Error("expected index to have Draft=false")
	}
}

func TestRename(t *testing.T) {
	dir := t.TempDir()
	content := []byte("---\ntitle: \"Blog\"\nalias: \"\"\ntags: []\nweight: 0\ncreated_at: \"2026-01-01T00:00:00Z\"\nupdated_at: \"2026-01-01T00:00:00Z\"\ntoc_sort: \"weight\"\ntoc_order: \"asc\"\n---\n# Blog\n")
	if err := Create(dir, "blog", content); err != nil {
		t.Fatal(err)
	}
	newNow := time.Date(2027, 6, 15, 12, 0, 0, 0, time.UTC)
	if err := Rename(dir, "blog", "articles", newNow); err != nil {
		t.Fatalf("Rename() error: %v", err)
	}
	// Old directory should be gone.
	if _, err := os.Stat(filepath.Join(dir, "pages", "blog")); err == nil {
		t.Error("expected old directory to be removed")
	}
	// New directory should exist with updated index.md.
	newIndexPath := filepath.Join(dir, "pages", "articles", "index.md")
	newContent, err := os.ReadFile(newIndexPath)
	if err != nil {
		t.Fatalf("expected new index.md to exist: %v", err)
	}
	s := string(newContent)
	if !strings.Contains(s, `title: "Articles"`) {
		t.Errorf("expected updated title in frontmatter, got: %s", s)
	}
	if !strings.Contains(s, newNow.UTC().Format(time.RFC3339)) {
		t.Errorf("expected updated_at = %s in frontmatter, got: %s", newNow.UTC().Format(time.RFC3339), s)
	}
}

func TestRenameNotFound(t *testing.T) {
	dir := t.TempDir()
	err := Rename(dir, "missing", "new-name", time.Now())
	if !errors.Is(err, ErrSectionNotFound) {
		t.Errorf("expected ErrSectionNotFound, got %v", err)
	}
}

func TestRenameTargetExists(t *testing.T) {
	dir := t.TempDir()
	content := []byte("---\ntitle: \"Blog\"\nalias: \"\"\ntags: []\nweight: 0\ncreated_at: \"2026-01-01T00:00:00Z\"\nupdated_at: \"2026-01-01T00:00:00Z\"\ntoc_sort: \"weight\"\ntoc_order: \"asc\"\n---\n# Blog\n")
	if err := Create(dir, "blog", content); err != nil {
		t.Fatal(err)
	}
	if err := Create(dir, "articles", content); err != nil {
		t.Fatal(err)
	}
	err := Rename(dir, "blog", "articles", time.Now())
	if !errors.Is(err, ErrSectionExists) {
		t.Errorf("expected ErrSectionExists, got %v", err)
	}
}

// assetDir creates pages/assets/i.png with no index.md, i.e. a plain
// directory that must not be treated as a section.
func assetDir(t *testing.T) (siteDir, file string) {
	t.Helper()
	siteDir = t.TempDir()
	d := filepath.Join(siteDir, "pages", "assets")
	if err := os.MkdirAll(d, 0755); err != nil {
		t.Fatal(err)
	}
	file = filepath.Join(d, "i.png")
	if err := os.WriteFile(file, []byte("png"), 0644); err != nil {
		t.Fatal(err)
	}
	return siteDir, file
}

func assertIntact(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("directory contents should be left intact: %v", err)
	}
}

func TestDeleteWithoutIndexMD(t *testing.T) {
	dir, file := assetDir(t)
	if err := Delete(dir, "assets"); !errors.Is(err, ErrSectionNotFound) {
		t.Errorf("expected ErrSectionNotFound, got %v", err)
	}
	assertIntact(t, file)
}

func TestUpdateWithoutIndexMD(t *testing.T) {
	dir, file := assetDir(t)
	if err := Update(dir, "assets", []byte("x")); !errors.Is(err, ErrSectionNotFound) {
		t.Errorf("expected ErrSectionNotFound, got %v", err)
	}
	assertIntact(t, file)
	if _, err := os.Stat(filepath.Join(dir, "pages", "assets", "index.md")); err == nil {
		t.Error("index.md must not be created")
	}
}

func TestRenameWithoutIndexMD(t *testing.T) {
	dir, file := assetDir(t)
	if err := Rename(dir, "assets", "media", time.Now()); !errors.Is(err, ErrSectionNotFound) {
		t.Errorf("expected ErrSectionNotFound, got %v", err)
	}
	assertIntact(t, file)
}

func TestDeleteIndexIsDirectory(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, "pages", "odd", "index.md")
	if err := os.MkdirAll(d, 0755); err != nil {
		t.Fatal(err)
	}
	if err := Delete(dir, "odd"); !errors.Is(err, ErrSectionNotFound) {
		t.Errorf("expected ErrSectionNotFound, got %v", err)
	}
	assertIntact(t, d)
}

func TestIndexIsDirectoryIsNotASection(t *testing.T) {
	dir := t.TempDir()
	d := filepath.Join(dir, "pages", "odd", "index.md")
	if err := os.MkdirAll(d, 0755); err != nil {
		t.Fatal(err)
	}
	if err := Update(dir, "odd", []byte("x")); !errors.Is(err, ErrSectionNotFound) {
		t.Errorf("Update: expected ErrSectionNotFound, got %v", err)
	}
	if err := Rename(dir, "odd", "even", time.Now()); !errors.Is(err, ErrSectionNotFound) {
		t.Errorf("Rename: expected ErrSectionNotFound, got %v", err)
	}
	assertIntact(t, d)
	sections, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 0 {
		t.Errorf("List should skip a directory whose index.md is a directory, got %v", sections)
	}
}

func TestRenameMinimalFrontmatter(t *testing.T) {
	now := time.Date(2027, 1, 2, 3, 4, 5, 0, time.UTC)
	for name, in := range map[string]string{
		"no frontmatter": "# Blog\n",
		"title only":     "---\ntitle: \"blog\"\n---\n# Blog\n",
		"no updated_at":  "---\ntitle: \"Blog\"\n---\n# Blog\n",
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if err := Create(dir, "blog", []byte(in)); err != nil {
				t.Fatal(err)
			}
			if err := Rename(dir, "blog", "my-journal", now); err != nil {
				t.Fatalf("Rename() error: %v", err)
			}
			b, err := os.ReadFile(filepath.Join(dir, "pages", "my-journal", "index.md"))
			if err != nil {
				t.Fatal(err)
			}
			s := string(b)
			if !strings.Contains(s, `title: "My Journal"`) || !strings.Contains(s, `updated_at: "2027-01-02T03:04:05Z"`) {
				t.Errorf("unexpected content: %q", s)
			}
		})
	}
}

func TestRenameKeepsCustomTitle(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("---\ntitle: \"Thoughts\"\n---\n# Blog\n")); err != nil {
		t.Fatal(err)
	}
	if err := Rename(dir, "blog", "journal", time.Now()); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "pages", "journal", "index.md"))
	if !strings.Contains(string(b), `title: "Thoughts"`) {
		t.Errorf("custom title should be kept, got: %s", b)
	}
}

func TestRenameInvalidNames(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	if err := Rename(dir, "../evil", "ok", time.Now()); !errors.Is(err, ErrInvalidName) {
		t.Errorf("invalid old name: expected ErrInvalidName, got %v", err)
	}
	if err := Rename(dir, "blog", "../evil", time.Now()); !errors.Is(err, ErrInvalidName) {
		t.Errorf("invalid new name: expected ErrInvalidName, got %v", err)
	}
}

func TestRenameNameConflictWithPage(t *testing.T) {
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	pageFile := filepath.Join(dir, "pages", "about.md")
	if err := os.WriteFile(pageFile, []byte("# About\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Rename(dir, "blog", "about", time.Now()); !errors.Is(err, ErrNameConflict) {
		t.Errorf("expected ErrNameConflict, got %v", err)
	}
	assertIntact(t, pageFile)
}

func TestRenameUnreadableIndex(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission errors cannot be provoked as root")
	}
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	idx := filepath.Join(dir, "pages", "blog", "index.md")
	if err := os.Chmod(idx, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(idx, 0644) })
	if err := Rename(dir, "blog", "news", time.Now()); err == nil {
		t.Fatal("expected an error for an unreadable index.md")
	}
	if _, err := os.Stat(filepath.Join(dir, "pages", "news")); err == nil {
		t.Error("section must not be renamed")
	}
}

func TestRenameRenameFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission errors cannot be provoked as root")
	}
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	pages := filepath.Join(dir, "pages")
	if err := os.Chmod(pages, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(pages, 0755) })
	if err := Rename(dir, "blog", "news", time.Now()); err == nil {
		t.Fatal("expected an error when pages/ is read-only")
	}
	if _, err := os.Stat(filepath.Join(pages, "blog", "index.md")); err != nil {
		t.Errorf("section must be intact: %v", err)
	}
}

func TestRenameRollsBackWhenIndexCannotBeWritten(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission errors cannot be provoked as root")
	}
	dir := t.TempDir()
	if err := Create(dir, "blog", []byte("# Blog\n")); err != nil {
		t.Fatal(err)
	}
	idx := filepath.Join(dir, "pages", "blog", "index.md")
	if err := os.Chmod(idx, 0444); err != nil {
		t.Fatal(err)
	}
	if err := Rename(dir, "blog", "news", time.Now()); err == nil {
		t.Fatal("expected an error when index.md is read-only")
	}
	if _, err := os.Stat(idx); err != nil {
		t.Errorf("directory rename must be rolled back: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "pages", "news")); err == nil {
		t.Error("new section directory must not remain")
	}
}
