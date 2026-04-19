package markdown

import (
	"strings"
	"testing"
)

func TestToHTML_Headings(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"# H1", "<h1>H1</h1>\n"},
		{"## H2", "<h2>H2</h2>\n"},
		{"### H3", "<h3>H3</h3>\n"},
		{"###### H6", "<h6>H6</h6>\n"},
	}
	for _, c := range cases {
		got := ToHTML(c.input)
		if got != c.want {
			t.Errorf("ToHTML(%q) = %q; want %q", c.input, got, c.want)
		}
	}
}

func TestToHTML_Paragraph(t *testing.T) {
	got := ToHTML("Hello world")
	if !strings.Contains(got, "<p>") || !strings.Contains(got, "Hello world") {
		t.Errorf("expected paragraph, got %q", got)
	}
}

func TestToHTML_Bold(t *testing.T) {
	got := ToHTML("**bold**")
	if !strings.Contains(got, "<strong>bold</strong>") {
		t.Errorf("expected bold, got %q", got)
	}
}

func TestToHTML_Italic(t *testing.T) {
	got := ToHTML("*italic*")
	if !strings.Contains(got, "<em>italic</em>") {
		t.Errorf("expected italic, got %q", got)
	}
}

func TestToHTML_InlineCode(t *testing.T) {
	got := ToHTML("`code`")
	if !strings.Contains(got, "<code>code</code>") {
		t.Errorf("expected inline code, got %q", got)
	}
}

func TestToHTML_Link(t *testing.T) {
	got := ToHTML("[press](https://example.com)")
	if !strings.Contains(got, `<a href="https://example.com">press</a>`) {
		t.Errorf("expected link, got %q", got)
	}
}

func TestToHTML_Image(t *testing.T) {
	got := ToHTML("![alt](img.png)")
	if !strings.Contains(got, `<img src="img.png" alt="alt">`) {
		t.Errorf("expected image, got %q", got)
	}
}

func TestToHTML_UnorderedList(t *testing.T) {
	md := "- item one\n- item two"
	got := ToHTML(md)
	if !strings.Contains(got, "<ul>") || !strings.Contains(got, "<li>item one</li>") {
		t.Errorf("expected unordered list, got %q", got)
	}
}

func TestToHTML_OrderedList(t *testing.T) {
	md := "1. first\n2. second"
	got := ToHTML(md)
	if !strings.Contains(got, "<ol>") || !strings.Contains(got, "<li>first</li>") {
		t.Errorf("expected ordered list, got %q", got)
	}
}

func TestToHTML_CodeBlock(t *testing.T) {
	md := "```\nfunc main() {}\n```"
	got := ToHTML(md)
	if !strings.Contains(got, "<pre><code>") || !strings.Contains(got, "func main()") {
		t.Errorf("expected code block, got %q", got)
	}
}

func TestToHTML_HorizontalRule(t *testing.T) {
	got := ToHTML("---")
	if !strings.Contains(got, "<hr>") {
		t.Errorf("expected hr, got %q", got)
	}
}

func TestToHTML_Blockquote(t *testing.T) {
	got := ToHTML("> quote text")
	if !strings.Contains(got, "<blockquote>") || !strings.Contains(got, "quote text") {
		t.Errorf("expected blockquote, got %q", got)
	}
}

func TestToHTML_EscapesHTMLInCode(t *testing.T) {
	got := ToHTML("```\n<script>alert(1)</script>\n```")
	if strings.Contains(got, "<script>") {
		t.Errorf("script tag should be escaped in code block, got %q", got)
	}
}

func TestExtractTitle(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"# My Title\n\nSome content.", "My Title"},
		{"## Not H1\n# Actual Title", "Actual Title"},
		{"No heading here", ""},
		{"#NoSpace", ""},
		{"# Title With Extra Spaces  ", "Title With Extra Spaces"},
	}
	for _, c := range cases {
		got := ExtractTitle(c.input)
		if got != c.want {
			t.Errorf("ExtractTitle(%q) = %q; want %q", c.input, got, c.want)
		}
	}
}
