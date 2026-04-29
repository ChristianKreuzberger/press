package minify

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHTML_BasicWhitespaceRemoval(t *testing.T) {
	input := `<html>
  <head>
    <title>Test</title>
  </head>
  <body>
    <p>Hello</p>
  </body>
</html>`
	got := HTML(input)

	// Indentation must be gone.
	if strings.Contains(got, "\n ") || strings.Contains(got, "  ") {
		t.Errorf("expected indentation to be removed, got:\n%s", got)
	}
	// Content must be preserved.
	if !strings.Contains(got, "Test") {
		t.Errorf("expected title content to be preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "Hello") {
		t.Errorf("expected body content to be preserved, got:\n%s", got)
	}
}

func TestHTML_CommentRemoval(t *testing.T) {
	input := `<html><!-- this is a comment -->
<body>
<!-- another comment -->
<p>Content</p>
</body>
</html>`
	got := HTML(input)

	if strings.Contains(got, "<!--") {
		t.Errorf("expected comments to be removed, got:\n%s", got)
	}
	if !strings.Contains(got, "Content") {
		t.Errorf("expected content to be preserved, got:\n%s", got)
	}
}

func TestHTML_IEConditionalCommentPreserved(t *testing.T) {
	input := `<html>
<!--[if lt IE 9]><script src="ie.js"></script><![endif]-->
<body><p>Hi</p></body>
</html>`
	got := HTML(input)

	if !strings.Contains(got, "<!--[if lt IE 9]>") {
		t.Errorf("expected IE conditional comment to be preserved, got:\n%s", got)
	}
}

func TestHTML_PreBlockPreservesIndentation(t *testing.T) {
	input := `<html>
<body>
<pre>
    line one
        line two indented
    line three
</pre>
</body>
</html>`
	got := HTML(input)

	// Indentation inside <pre> must be preserved.
	if !strings.Contains(got, "    line one") {
		t.Errorf("expected indentation inside <pre> to be preserved, got:\n%s", got)
	}
	if !strings.Contains(got, "        line two indented") {
		t.Errorf("expected deeper indentation inside <pre> to be preserved, got:\n%s", got)
	}
}

func TestHTML_Idempotent(t *testing.T) {
	input := `<html><head><title>T</title></head><body><p>Hello</p></body></html>`
	first := HTML(input)
	second := HTML(first)
	if first != second {
		t.Errorf("HTML() is not idempotent:\nfirst:  %s\nsecond: %s", first, second)
	}
}

func TestHTML_EmptyInput(t *testing.T) {
	got := HTML("")
	if got != "" {
		t.Errorf("expected empty string for empty input, got: %q", got)
	}
}

func TestHTML_TrimsLeadingTrailingWhitespace(t *testing.T) {
	input := "  <p>Hello</p>  "
	got := HTML(input)
	if got != "<p>Hello</p>" {
		t.Errorf("expected trimmed output, got: %q", got)
	}
}

// TestHTML_PreTagNotFalselyMatchedByLongerTagName ensures that tags whose names
// start with "pre" (e.g. <presentation>) are NOT treated as <pre> blocks.
func TestHTML_PreTagNotFalselyMatchedByLongerTagName(t *testing.T) {
	input := `<html>
<body>
<presentation>
    this should be trimmed
</presentation>
</body>
</html>`
	got := HTML(input)

	// The content lines should have their indentation stripped because
	// <presentation> must not be confused with <pre>.
	if strings.Contains(got, "    this should be trimmed") {
		t.Errorf("<presentation> was incorrectly treated as a <pre> block; got:\n%s", got)
	}
	if !strings.Contains(got, "this should be trimmed") {
		t.Errorf("content inside <presentation> should be preserved (trimmed); got:\n%s", got)
	}
}

// A space between two inline elements is rendered; removing it glues words together.
func TestHTML_KeepsSpaceBetweenInlineTags(t *testing.T) {
	input := "<p><strong>bold</strong> <em>italic</em></p>"
	if got := HTML(input); got != input {
		t.Errorf("space between inline tags was changed: %q", got)
	}
}

// Whitespace between tags inside <pre> is content and must survive.
func TestHTML_PreKeepsWhitespaceBetweenTags(t *testing.T) {
	input := "<pre><code><span>a</span>   <span>b</span>\n<span>c</span></code></pre>"
	if got := HTML(input); got != input {
		t.Errorf("whitespace inside <pre> was changed: %q", got)
	}
}

func TestFiles(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.html")
	src := "<html>\n  <!-- c -->\n  <p>Hi</p>\n</html>\n"
	if err := os.WriteFile(a, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	before, after, err := Files([]string{a})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(a)
	if string(got) != "<html>\n<p>Hi</p>\n</html>" {
		t.Errorf("unexpected file content: %q", got)
	}
	if before != int64(len(src)) || after != int64(len(got)) || after >= before {
		t.Errorf("bad sizes: before=%d after=%d", before, after)
	}
}

func TestFilesMissingFile(t *testing.T) {
	if _, _, err := Files([]string{filepath.Join(t.TempDir(), "nope.html")}); err == nil {
		t.Fatal("expected error for missing file")
	}
}
