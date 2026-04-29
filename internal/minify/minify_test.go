package minify

import (
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

	// Tags should be adjacent after collapsing inter-tag whitespace.
	if strings.Contains(got, ">\n<") || strings.Contains(got, ">  <") {
		t.Errorf("expected whitespace between tags to be collapsed, got:\n%s", got)
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
