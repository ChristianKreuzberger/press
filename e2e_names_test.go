package main_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestE2EInvalidAndCollidingNames(t *testing.T) {
	siteDir := t.TempDir()
	run(t, siteDir, "init")
	run(t, siteDir, "create", "page", "about-us")
	run(t, siteDir, "create", "section", "blog")

	for _, name := range []string{"", "blog/", ".", "..", "/tmp/zzz", `a\b`} {
		runExpectError(t, siteDir, "create", "page", name)
		runExpectError(t, siteDir, "update", "page", name, "--file", "pages/about-us.md")
		runExpectError(t, siteDir, "delete", "page", name)
		runExpectError(t, siteDir, "rename", "page", "about-us", name)
	}

	runExpectError(t, siteDir, "create", "page", "blog")          // section exists
	runExpectError(t, siteDir, "create", "section", "about-us")   // page exists
	runExpectError(t, siteDir, "create", "page", "about-us/team") // page file in the way
	runExpectError(t, siteDir, "rename", "page", "about-us", "blog")
	runExpectError(t, siteDir, "rename", "section", "blog", "about-us")
	runExpectError(t, siteDir, "rename", "page", "index", "home")

	for _, p := range []string{"blog.md", "about-us", "about-us/team.md", ".md", "tmp", "home.md"} {
		if _, err := os.Stat(filepath.Join(siteDir, "pages", p)); err == nil {
			t.Errorf("%s should not exist", p)
		}
	}
	for _, p := range []string{"index.md", "about-us.md", "blog/index.md"} {
		if _, err := os.Stat(filepath.Join(siteDir, "pages", p)); err != nil {
			t.Errorf("%s should still exist: %v", p, err)
		}
	}
}
