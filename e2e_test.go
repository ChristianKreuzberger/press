package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// pressBinary holds the path to the compiled press binary, built once by TestMain.
var pressBinary string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "press-build-*")
	if err != nil {
		panic("could not create temp dir for binary: " + err.Error())
	}

	pressBinary = filepath.Join(tmp, "press")
	cmd := exec.Command("go", "build", "-o", pressBinary, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = os.RemoveAll(tmp)
		panic("could not build press binary: " + string(out))
	}

	code := m.Run()
	_ = os.RemoveAll(tmp)
	os.Exit(code)
}

// run executes press with the given arguments inside siteDir and returns
// the combined stdout+stderr output.
func run(t *testing.T, siteDir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(pressBinary, args...)
	cmd.Dir = siteDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("press %v failed: %v\n%s", args, err, out)
	}
	return string(out)
}

// runExpectError executes press and asserts it exits with a non-zero status.
func runExpectError(t *testing.T, siteDir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(pressBinary, args...)
	cmd.Dir = siteDir
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("press %v expected failure but succeeded; output:\n%s", args, out)
	}
	return string(out)
}

func TestE2E(t *testing.T) {
	siteDir := t.TempDir()

	// --- press init ---
	out := run(t, siteDir, "init")
	if !strings.Contains(out, "initialised") {
		t.Errorf("init output should mention initialised, got: %s", out)
	}

	tmplPath := filepath.Join(siteDir, "template.html")
	if _, err := os.Stat(tmplPath); err != nil {
		t.Fatal("press init should create template.html")
	}
	pagesDir := filepath.Join(siteDir, "pages")
	if _, err := os.Stat(pagesDir); err != nil {
		t.Fatal("press init should create pages/")
	}
	if _, err := os.Stat(filepath.Join(pagesDir, "index.md")); err != nil {
		t.Fatal("press init should create pages/index.md")
	}

	// Running init again should not fail and should skip existing files
	out = run(t, siteDir, "init")
	if !strings.Contains(out, "already exists") {
		t.Errorf("second init should say files already exist, got: %s", out)
	}

	// --- press list page (has index from init) ---
	out = run(t, siteDir, "list", "page")
	if !strings.Contains(out, "index") {
		t.Errorf("page list should contain 'index' after init, got: %s", out)
	}

	// --- press update page index (from file) ---
	indexMD := filepath.Join(t.TempDir(), "index.md")
	writeFile(t, indexMD, "# Home\n\nWelcome to my site!\n")

	run(t, siteDir, "update", "page", "index", "--file", indexMD)

	// --- press create page (empty, no file flag) ---
	run(t, siteDir, "create", "page", "about")
	if _, err := os.Stat(filepath.Join(pagesDir, "about.md")); err != nil {
		t.Fatal("page create without --file should still create pages/about.md")
	}

	// --- duplicate create should fail ---
	runExpectError(t, siteDir, "create", "page", "index")

	// --- press list page ---
	out = run(t, siteDir, "list", "page")
	if !strings.Contains(out, "index") {
		t.Errorf("page list should contain 'index', got: %s", out)
	}
	if !strings.Contains(out, "about") {
		t.Errorf("page list should contain 'about', got: %s", out)
	}

	// --- press build ---
	run(t, siteDir, "build")

	distDir := filepath.Join(siteDir, "dist")
	indexHTML := filepath.Join(distDir, "index.html")
	aboutHTML := filepath.Join(distDir, "about.html")

	if _, err := os.Stat(indexHTML); err != nil {
		t.Fatal("build should produce dist/index.html")
	}
	if _, err := os.Stat(aboutHTML); err != nil {
		t.Fatal("build should produce dist/about.html")
	}

	// Check index.html content
	content := readFile(t, indexHTML)
	if !strings.Contains(content, "<h1") || !strings.Contains(content, ">Home</h1>") {
		t.Errorf("dist/index.html should contain <h1>Home</h1>, got:\n%s", content)
	}
	if !strings.Contains(content, "Welcome to my site") {
		t.Errorf("dist/index.html should contain page body, got:\n%s", content)
	}
	if !strings.Contains(content, "<title>Home</title>") {
		t.Errorf("dist/index.html should have <title>Home</title>, got:\n%s", content)
	}
	// Navigation links
	if !strings.Contains(content, "about.html") {
		t.Errorf("dist/index.html should link to about.html, got:\n%s", content)
	}
	if !strings.Contains(content, "index.html") {
		t.Errorf("dist/index.html should link to index.html, got:\n%s", content)
	}

	// --- press update page ---
	updatedMD := filepath.Join(t.TempDir(), "updated.md")
	writeFile(t, updatedMD, "# Home Updated\n\nThis content was updated.\n")
	run(t, siteDir, "update", "page", "index", "--file", updatedMD)

	// Rebuild and verify updated content
	run(t, siteDir, "build")
	content = readFile(t, indexHTML)
	if !strings.Contains(content, "Home Updated") {
		t.Errorf("dist/index.html should contain updated heading, got:\n%s", content)
	}
	if !strings.Contains(content, "This content was updated") {
		t.Errorf("dist/index.html should contain updated body, got:\n%s", content)
	}

	// --- press delete page ---
	run(t, siteDir, "delete", "page", "about")

	if _, err := os.Stat(filepath.Join(pagesDir, "about.md")); !os.IsNotExist(err) {
		t.Fatal("pages/about.md should have been deleted")
	}

	// Verify list no longer contains about
	out = run(t, siteDir, "list", "page")
	if strings.Contains(out, "about") {
		t.Errorf("page list should not contain 'about' after delete, got: %s", out)
	}
	if !strings.Contains(out, "index") {
		t.Errorf("page list should still contain 'index', got: %s", out)
	}

	// --- delete non-existent page should fail ---
	runExpectError(t, siteDir, "delete", "page", "nonexistent")

	// --- update non-existent page should fail ---
	runExpectError(t, siteDir, "update", "page", "nonexistent", "--file", updatedMD)

	// --- press build --output ---
	customOut := filepath.Join(siteDir, "public")
	run(t, siteDir, "build", "--output", "public")
	if _, err := os.Stat(filepath.Join(customOut, "index.html")); err != nil {
		t.Fatal("build --output public should produce public/index.html")
	}

	// --- press --version ---
	cmd := exec.Command(pressBinary, "--version")
	cmd.Dir = siteDir
	vOut, _ := cmd.CombinedOutput()
	if strings.TrimSpace(string(vOut)) == "" {
		t.Error("--version should print a version string")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestE2ESection(t *testing.T) {
	siteDir := t.TempDir()

	// Initialise the site first (also creates pages/index.md).
	run(t, siteDir, "init")

	// --- press list section (empty) ---
	out := run(t, siteDir, "list", "section")
	if !strings.Contains(out, "no sections") {
		t.Errorf("empty section list should say 'no sections', got: %s", out)
	}

	// --- press create section (from file) ---
	blogIndexMD := filepath.Join(t.TempDir(), "blog-index.md")
	writeFile(t, blogIndexMD, "# Blog\n\nAll blog posts.\n")

	run(t, siteDir, "create", "section", "blog", "--file", blogIndexMD)

	blogDir := filepath.Join(siteDir, "pages", "blog")
	if _, err := os.Stat(blogDir); err != nil {
		t.Fatal("section create should create pages/blog/")
	}
	if _, err := os.Stat(filepath.Join(blogDir, "index.md")); err != nil {
		t.Fatal("section create should create pages/blog/index.md")
	}

	// --- press create section (empty, no file flag) ---
	run(t, siteDir, "create", "section", "docs")
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "docs", "index.md")); err != nil {
		t.Fatal("section create without --file should still create pages/docs/index.md")
	}

	// --- duplicate section create should fail ---
	runExpectError(t, siteDir, "create", "section", "blog")

	// --- press list section ---
	out = run(t, siteDir, "list", "section")
	if !strings.Contains(out, "blog") {
		t.Errorf("section list should contain 'blog', got: %s", out)
	}
	if !strings.Contains(out, "docs") {
		t.Errorf("section list should contain 'docs', got: %s", out)
	}

	// Add a non-index page to the blog section manually.
	writeFile(t, filepath.Join(blogDir, "first-post.md"), "# First Post\n\nHello world!\n")

	// --- press build (with sections) ---
	run(t, siteDir, "build")

	distDir := filepath.Join(siteDir, "dist")

	// Section index should be generated.
	blogIndexHTML := filepath.Join(distDir, "blog", "index.html")
	if _, err := os.Stat(blogIndexHTML); err != nil {
		t.Fatal("build should produce dist/blog/index.html")
	}

	// Section sub-page should be generated.
	firstPostHTML := filepath.Join(distDir, "blog", "first-post.html")
	if _, err := os.Stat(firstPostHTML); err != nil {
		t.Fatal("build should produce dist/blog/first-post.html")
	}

	// Check blog index content.
	content := readFile(t, blogIndexHTML)
	if !strings.Contains(content, "All blog posts.") {
		t.Errorf("dist/blog/index.html should contain blog body, got:\n%s", content)
	}
	if !strings.Contains(content, "<title>Blog</title>") {
		t.Errorf("dist/blog/index.html should have <title>Blog</title>, got:\n%s", content)
	}

	// Navigation in blog/index.html should use relative links prefixed with "../".
	if !strings.Contains(content, "../index.html") {
		t.Errorf("dist/blog/index.html nav should link to ../index.html, got:\n%s", content)
	}
	if !strings.Contains(content, "../blog/index.html") {
		t.Errorf("dist/blog/index.html nav should link to ../blog/index.html, got:\n%s", content)
	}

	// Navigation in top-level index.html should include the section link.
	topContent := readFile(t, filepath.Join(distDir, "index.html"))
	if !strings.Contains(topContent, "blog/index.html") {
		t.Errorf("dist/index.html nav should link to blog/index.html, got:\n%s", topContent)
	}

	// Check first post content.
	postContent := readFile(t, firstPostHTML)
	if !strings.Contains(postContent, "Hello world!") {
		t.Errorf("dist/blog/first-post.html should contain post body, got:\n%s", postContent)
	}

	// --- press update section ---
	updatedBlogMD := filepath.Join(t.TempDir(), "updated-blog.md")
	writeFile(t, updatedBlogMD, "# Blog Updated\n\nUpdated description.\n")
	run(t, siteDir, "update", "section", "blog", "--file", updatedBlogMD)

	// Rebuild and verify updated section index content.
	run(t, siteDir, "build")
	content = readFile(t, blogIndexHTML)
	if !strings.Contains(content, "Blog Updated") {
		t.Errorf("dist/blog/index.html should contain updated heading, got:\n%s", content)
	}
	if !strings.Contains(content, "Updated description.") {
		t.Errorf("dist/blog/index.html should contain updated body, got:\n%s", content)
	}

	// --- press delete section ---
	run(t, siteDir, "delete", "section", "docs")

	if _, err := os.Stat(filepath.Join(siteDir, "pages", "docs")); !os.IsNotExist(err) {
		t.Fatal("pages/docs/ should have been deleted")
	}

	// Verify list no longer contains docs.
	out = run(t, siteDir, "list", "section")
	if strings.Contains(out, "docs") {
		t.Errorf("section list should not contain 'docs' after delete, got: %s", out)
	}
	if !strings.Contains(out, "blog") {
		t.Errorf("section list should still contain 'blog', got: %s", out)
	}

	// --- delete non-existent section should fail ---
	runExpectError(t, siteDir, "delete", "section", "nonexistent")

	// --- update non-existent section should fail ---
	runExpectError(t, siteDir, "update", "section", "nonexistent", "--file", updatedBlogMD)

	// --- section update without --file should fail ---
	runExpectError(t, siteDir, "update", "section", "blog")
}

func TestE2ESectionTOC(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")

	// Create a section with toc_sort by title.
	blogIndexMD := filepath.Join(t.TempDir(), "blog-index.md")
	writeFile(t, blogIndexMD, "---\ntitle: \"Blog\"\ntoc_sort: \"title\"\ntoc_order: \"asc\"\n---\n# Blog\n\nAll posts.\n")
	run(t, siteDir, "create", "section", "blog", "--file", blogIndexMD)

	blogDir := filepath.Join(siteDir, "pages", "blog")
	writeFile(t, filepath.Join(blogDir, "zebra-post.md"), "# Zebra Post\n\nZ content.\n")
	writeFile(t, filepath.Join(blogDir, "apple-post.md"), "# Apple Post\n\nA content.\n")

	run(t, siteDir, "build")

	distDir := filepath.Join(siteDir, "dist")
	blogIndexHTML := filepath.Join(distDir, "blog", "index.html")

	content := readFile(t, blogIndexHTML)

	// TOC section must be present.
	if !strings.Contains(content, "Contents") {
		t.Errorf("blog/index.html should contain a TOC, got:\n%s", content)
	}

	// apple-post should come before zebra-post (title asc).
	applePos := strings.Index(content, "apple-post.html")
	zebraPos := strings.Index(content, "zebra-post.html")
	if applePos == -1 || zebraPos == -1 {
		t.Fatalf("expected both TOC entries in blog/index.html, got:\n%s", content)
	}
	if applePos > zebraPos {
		t.Errorf("expected apple-post.html before zebra-post.html (title asc)")
	}

	// Child pages themselves should not have a TOC.
	postContent := readFile(t, filepath.Join(distDir, "blog", "apple-post.html"))
	if strings.Contains(postContent, "class=\"toc\"") {
		t.Errorf("child page should not have a TOC section, got:\n%s", postContent)
	}
}


func TestE2ECheck(t *testing.T) {
	siteDir := t.TempDir()
	pagesDir := filepath.Join(siteDir, "pages")

	// --- clean site after init should have no issues ---
	run(t, siteDir, "init")

	out := run(t, siteDir, "check")
	if !strings.Contains(out, "pages checked") {
		t.Errorf("check output should mention pages checked, got: %s", out)
	}

	// --- page with missing title in frontmatter ---
	writeFile(t, filepath.Join(pagesDir, "no-title.md"), "---\ntitle: \"\"\n---\n# No Title\n\nContent here.\n")
	out = runExpectError(t, siteDir, "check")
	if !strings.Contains(out, "no-title.md: missing title") {
		t.Errorf("check should report missing title, got: %s", out)
	}
	if err := os.Remove(filepath.Join(pagesDir, "no-title.md")); err != nil {
		t.Fatal(err)
	}

	// --- page with empty content ---
	writeFile(t, filepath.Join(pagesDir, "empty.md"), "---\ntitle: \"Empty\"\n---\n")
	out = runExpectError(t, siteDir, "check")
	if !strings.Contains(out, "empty.md: empty page content") {
		t.Errorf("check should report empty page content, got: %s", out)
	}
	if err := os.Remove(filepath.Join(pagesDir, "empty.md")); err != nil {
		t.Fatal(err)
	}

	// --- section directory without index.md ---
	bareDir := filepath.Join(pagesDir, "bare")
	if err := os.MkdirAll(bareDir, 0755); err != nil {
		t.Fatal(err)
	}
	out = runExpectError(t, siteDir, "check")
	if !strings.Contains(out, "bare/: section has no index.md") {
		t.Errorf("check should report section without index.md, got: %s", out)
	}
	if err := os.RemoveAll(bareDir); err != nil {
		t.Fatal(err)
	}

	// --- broken internal link ---
	writeFile(t, filepath.Join(pagesDir, "with-link.md"), "---\ntitle: \"With Link\"\n---\n# With Link\n\nSee [team](/team) for more.\n")
	out = runExpectError(t, siteDir, "check")
	if !strings.Contains(out, "with-link.md: broken link → /team (page not found)") {
		t.Errorf("check should report broken internal link, got: %s", out)
	}

	// Creating the linked page should fix the broken link.
	writeFile(t, filepath.Join(pagesDir, "team.md"), "---\ntitle: \"Team\"\n---\n# Team\n\nMeet the team.\n")
	out = run(t, siteDir, "check")
	if strings.Contains(out, "broken link") {
		t.Errorf("check should not report broken link after creating the target page, got: %s", out)
	}

	// --- exit code 0 when all clean ---
	if err := os.Remove(filepath.Join(pagesDir, "with-link.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(pagesDir, "team.md")); err != nil {
		t.Fatal(err)
	}
	out = run(t, siteDir, "check")
	if strings.Contains(out, "issue(s) found") {
		t.Errorf("check should report no issues on clean site, got: %s", out)
	}
}

func TestE2ETree(t *testing.T) {
	siteDir := t.TempDir()

	// --- empty site: no pages or sections ---
	out := run(t, siteDir, "tree")
	if !strings.Contains(out, "no pages or sections") {
		t.Errorf("tree on empty site should say 'no pages or sections', got: %s", out)
	}

	// --- after init: only index page ---
	run(t, siteDir, "init")
	out = run(t, siteDir, "tree")
	if !strings.Contains(out, "pages/") {
		t.Errorf("tree should show 'pages/' header, got: %s", out)
	}
	if !strings.Contains(out, "index") {
		t.Errorf("tree should show 'index' page, got: %s", out)
	}

	// --- add a page and a section with a page ---
	run(t, siteDir, "create", "page", "about")
	run(t, siteDir, "create", "section", "blog")
	blogDir := filepath.Join(siteDir, "pages", "blog")
	writeFile(t, filepath.Join(blogDir, "first-post.md"), "# First Post\n\nHello world!\n")

	out = run(t, siteDir, "tree")

	// pages/ header
	if !strings.Contains(out, "pages/") {
		t.Errorf("tree should show 'pages/' header, got: %s", out)
	}
	// top-level pages
	if !strings.Contains(out, "about") {
		t.Errorf("tree should show 'about' page, got: %s", out)
	}
	if !strings.Contains(out, "index") {
		t.Errorf("tree should show 'index' page, got: %s", out)
	}
	// section
	if !strings.Contains(out, "blog/") {
		t.Errorf("tree should show 'blog/' section, got: %s", out)
	}
	// section pages
	if !strings.Contains(out, "first-post") {
		t.Errorf("tree should show 'first-post' under blog, got: %s", out)
	}
}
