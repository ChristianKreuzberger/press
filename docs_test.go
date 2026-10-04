package main_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestDocsCoverCLI keeps README.md and the agent skill in step with the CLI
// (issue #67): every command and flag the code defines must be mentioned.
func TestDocsCoverCLI(t *testing.T) {
	src := func(name string) string {
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	docs := map[string]string{
		"README.md": src("README.md"),
		"SKILL.md":  src(filepath.Join(".github", "skills", "press", "SKILL.md")),
	}

	// Commands: the cases of the switch in main().
	mainSrc := src("main.go")
	swStart := strings.Index(mainSrc, "switch args[0]")
	if swStart < 0 {
		t.Fatal(`main.go has no "switch args[0]"; has the command switch moved?`)
	}
	sw := mainSrc[swStart:]
	swEnd := strings.Index(sw, "default:")
	if swEnd < 0 {
		t.Fatal(`main.go's command switch has no "default:" case`)
	}
	sw = sw[:swEnd]
	cmds := regexp.MustCompile(`case "(\w+)":`).FindAllStringSubmatch(sw, -1)
	if len(cmds) < 10 {
		t.Fatalf("found only %d commands in main.go; has the switch moved?", len(cmds))
	}

	// Flags: every flag registered on a command's flag set, plus -version.
	flagNames := map[string]bool{"version": true}
	files, _ := filepath.Glob("cmd_*.go")
	flagRe := regexp.MustCompile(`fs\.(?:Bool|String|Duration|Int)\("(\w+)"`)
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		for _, m := range flagRe.FindAllStringSubmatch(src(f), -1) {
			flagNames[m[1]] = true
		}
	}

	for name, doc := range docs {
		for _, m := range cmds {
			if !strings.Contains(doc, "press "+m[1]) {
				t.Errorf("%s does not mention `press %s`", name, m[1])
			}
		}
		for _, noun := range []string{"press rename page", "press rename section"} {
			if !strings.Contains(doc, noun) {
				t.Errorf("%s does not mention `%s`", name, noun)
			}
		}
		for f := range flagNames {
			if !strings.Contains(doc, "--"+f) && !strings.Contains(doc, " -"+f) {
				t.Errorf("%s does not mention the %q flag", name, f)
			}
		}
		if !strings.Contains(doc, "static/") {
			t.Errorf("%s does not mention the static/ directory", name)
		}
	}
}
