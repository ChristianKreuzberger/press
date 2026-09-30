package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runCode executes press and returns its exit code and combined output.
func runCode(t *testing.T, siteDir string, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command(pressBinary, args...)
	cmd.Dir = siteDir
	out, err := cmd.CombinedOutput()
	if err == nil {
		return 0, string(out)
	}
	ee, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("press %v: %v", args, err)
	}
	return ee.ExitCode(), string(out)
}

func assertExit(t *testing.T, siteDir string, want int, args ...string) string {
	t.Helper()
	code, out := runCode(t, siteDir, args...)
	if code != want {
		t.Errorf("press %v: exit %d, want %d; output:\n%s", args, code, want, out)
	}
	return out
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestE2EArgParsing(t *testing.T) {
	site := t.TempDir()
	run(t, site, "init")
	src := filepath.Join(site, "x.md")
	writeFile(t, src, "---\ntitle: X\n---\n# X\n")
	pages := filepath.Join(site, "pages")

	for _, noun := range []string{"page", "section"} {
		t.Run(noun, func(t *testing.T) {
			// Documented order and flag-first order both work.
			assertExit(t, site, 0, "create", noun, "a-"+noun, "--file", src)
			assertExit(t, site, 0, "create", noun, "--file", src, "b-"+noun)
			for _, n := range []string{"a-", "b-"} {
				p := filepath.Join(pages, n+noun)
				if noun == "page" {
					p += ".md"
				}
				if !exists(p) {
					t.Errorf("%s not created", p)
				}
			}
			assertExit(t, site, 0, "update", noun, "a-"+noun, "--file", src)
			assertExit(t, site, 0, "update", noun, "--file", src, "b-"+noun)

			assertExit(t, site, 2, "create", noun, "x1", "x2")
			assertExit(t, site, 2, "create", noun)
			assertExit(t, site, 2, "create", noun, "--bogus", "x3")
			assertExit(t, site, 2, "update", noun, "a-"+noun)
			assertExit(t, site, 2, "delete", noun, "a-"+noun, "extra")
			assertExit(t, site, 2, "delete", noun)
			assertExit(t, site, 2, "list", noun, "junk")
			assertExit(t, site, 2, "rename", noun, "a", "b", "c")
			assertExit(t, site, 2, "rename", noun, "a")
			assertExit(t, site, 1, "update", noun, "nonexistent", "--file", src)
		})
	}
	for _, n := range []string{"--file.md", "x1.md", "x1", "x2.md", "x2", "x3.md", "x3"} {
		if exists(filepath.Join(pages, n)) {
			t.Errorf("stray %s created by invalid invocation", n)
		}
	}
}

func TestE2EInitArgs(t *testing.T) {
	base := t.TempDir()
	d1, d2 := filepath.Join(base, "d1"), filepath.Join(base, "d2")
	assertExit(t, base, 0, "init", d1, "--theme", "light")
	assertExit(t, base, 0, "init", "--theme", "light", d2)
	for _, d := range []string{d1, d2} {
		if !exists(filepath.Join(d, "template.html")) {
			t.Errorf("%s not initialised", d)
		}
	}
	assertExit(t, base, 2, "init", "a", "b")
	assertExit(t, base, 2, "init", "--theme", "nope")
	assertExit(t, base, 2, "init", filepath.Join(base, "d3"), "--theme", "nope")
	if exists(filepath.Join(base, "d3")) {
		t.Error("invalid theme must not create the directory")
	}
}

func TestE2EUsageExitCodes(t *testing.T) {
	site := t.TempDir()
	run(t, site, "init")
	assertExit(t, site, 2)
	assertExit(t, site, 2, "frobnicate")
	out := assertExit(t, site, 2, "create", "widget", "x")
	if !strings.Contains(out, "widget") {
		t.Errorf("unknown noun message should name the noun, got: %s", out)
	}
	assertExit(t, site, 2, "create")
	assertExit(t, site, 2, "build", "--version")
	assertExit(t, site, 2, "build", "extra")
	assertExit(t, site, 2, "tree", "extra")
	assertExit(t, site, 2, "check", "extra")
	assertExit(t, site, 0, "build", "--help")
	assertExit(t, site, 0, "build")
}
