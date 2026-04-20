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

	// Running init again should not fail and should skip template.html
	out = run(t, siteDir, "init")
	if !strings.Contains(out, "already exists") {
		t.Errorf("second init should say template.html already exists, got: %s", out)
	}

	// --- press page list (empty) ---
	out = run(t, siteDir, "page", "list")
	if !strings.Contains(out, "no pages") {
		t.Errorf("empty page list should say 'no pages', got: %s", out)
	}

	// --- press page create (from file) ---
	indexMD := filepath.Join(t.TempDir(), "index.md")
	writeFile(t, indexMD, "# Home\n\nWelcome to my site!\n")

	run(t, siteDir, "page", "create", "index", "--file", indexMD)

	if _, err := os.Stat(filepath.Join(pagesDir, "index.md")); err != nil {
		t.Fatal("page create should create pages/index.md")
	}

	// --- press page create (empty, no file flag) ---
	run(t, siteDir, "page", "create", "about")
	if _, err := os.Stat(filepath.Join(pagesDir, "about.md")); err != nil {
		t.Fatal("page create without --file should still create pages/about.md")
	}

	// --- duplicate create should fail ---
	runExpectError(t, siteDir, "page", "create", "index")

	// --- press page list ---
	out = run(t, siteDir, "page", "list")
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

	// --- press page update ---
	updatedMD := filepath.Join(t.TempDir(), "updated.md")
	writeFile(t, updatedMD, "# Home Updated\n\nThis content was updated.\n")
	run(t, siteDir, "page", "update", "index", "--file", updatedMD)

	// Rebuild and verify updated content
	run(t, siteDir, "build")
	content = readFile(t, indexHTML)
	if !strings.Contains(content, "Home Updated") {
		t.Errorf("dist/index.html should contain updated heading, got:\n%s", content)
	}
	if !strings.Contains(content, "This content was updated") {
		t.Errorf("dist/index.html should contain updated body, got:\n%s", content)
	}

	// --- press page delete ---
	run(t, siteDir, "page", "delete", "about")

	if _, err := os.Stat(filepath.Join(pagesDir, "about.md")); !os.IsNotExist(err) {
		t.Fatal("pages/about.md should have been deleted")
	}

	// Verify list no longer contains about
	out = run(t, siteDir, "page", "list")
	if strings.Contains(out, "about") {
		t.Errorf("page list should not contain 'about' after delete, got: %s", out)
	}
	if !strings.Contains(out, "index") {
		t.Errorf("page list should still contain 'index', got: %s", out)
	}

	// --- delete non-existent page should fail ---
	runExpectError(t, siteDir, "page", "delete", "nonexistent")

	// --- update non-existent page should fail ---
	runExpectError(t, siteDir, "page", "update", "nonexistent", "--file", updatedMD)

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

	// Initialise the site first.
	run(t, siteDir, "init")

	// Create a top-level page so navigation has something to link to.
	run(t, siteDir, "page", "create", "index")

	// --- press section list (empty) ---
	out := run(t, siteDir, "section", "list")
	if !strings.Contains(out, "no sections") {
		t.Errorf("empty section list should say 'no sections', got: %s", out)
	}

	// --- press section create (from file) ---
	blogIndexMD := filepath.Join(t.TempDir(), "blog-index.md")
	writeFile(t, blogIndexMD, "# Blog\n\nAll blog posts.\n")

	run(t, siteDir, "section", "create", "blog", "--file", blogIndexMD)

	blogDir := filepath.Join(siteDir, "pages", "blog")
	if _, err := os.Stat(blogDir); err != nil {
		t.Fatal("section create should create pages/blog/")
	}
	if _, err := os.Stat(filepath.Join(blogDir, "index.md")); err != nil {
		t.Fatal("section create should create pages/blog/index.md")
	}

	// --- press section create (empty, no file flag) ---
	run(t, siteDir, "section", "create", "docs")
	if _, err := os.Stat(filepath.Join(siteDir, "pages", "docs", "index.md")); err != nil {
		t.Fatal("section create without --file should still create pages/docs/index.md")
	}

	// --- duplicate section create should fail ---
	runExpectError(t, siteDir, "section", "create", "blog")

	// --- press section list ---
	out = run(t, siteDir, "section", "list")
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

	// --- press section update ---
	updatedBlogMD := filepath.Join(t.TempDir(), "updated-blog.md")
	writeFile(t, updatedBlogMD, "# Blog Updated\n\nUpdated description.\n")
	run(t, siteDir, "section", "update", "blog", "--file", updatedBlogMD)

	// Rebuild and verify updated section index content.
	run(t, siteDir, "build")
	content = readFile(t, blogIndexHTML)
	if !strings.Contains(content, "Blog Updated") {
		t.Errorf("dist/blog/index.html should contain updated heading, got:\n%s", content)
	}
	if !strings.Contains(content, "Updated description.") {
		t.Errorf("dist/blog/index.html should contain updated body, got:\n%s", content)
	}

	// --- press section delete ---
	run(t, siteDir, "section", "delete", "docs")

	if _, err := os.Stat(filepath.Join(siteDir, "pages", "docs")); !os.IsNotExist(err) {
		t.Fatal("pages/docs/ should have been deleted")
	}

	// Verify list no longer contains docs.
	out = run(t, siteDir, "section", "list")
	if strings.Contains(out, "docs") {
		t.Errorf("section list should not contain 'docs' after delete, got: %s", out)
	}
	if !strings.Contains(out, "blog") {
		t.Errorf("section list should still contain 'blog', got: %s", out)
	}

	// --- delete non-existent section should fail ---
	runExpectError(t, siteDir, "section", "delete", "nonexistent")

	// --- update non-existent section should fail ---
	runExpectError(t, siteDir, "section", "update", "nonexistent", "--file", updatedBlogMD)

	// --- section update without --file should fail ---
	runExpectError(t, siteDir, "section", "update", "blog")
}
