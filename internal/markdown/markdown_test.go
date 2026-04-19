package markdown

import (
	"strings"
	"testing"
)

func TestToHTML_Headings(t *testing.T) {
	cases := []struct {
		input   string
		wantTag string
		wantText string
	}{
		{"# H1", "h1", "H1"},
		{"## H2", "h2", "H2"},
		{"### H3", "h3", "H3"},
		{"###### H6", "h6", "H6"},
	}
	for _, c := range cases {
		got := ToHTML(c.input)
		open := "<" + c.wantTag
		close := "</" + c.wantTag + ">"
		if !strings.Contains(got, open) || !strings.Contains(got, close) || !strings.Contains(got, c.wantText) {
			t.Errorf("ToHTML(%q) = %q; want %s...%s containing %q", c.input, got, open, close, c.wantText)
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
	if !strings.Contains(got, `href="https://example.com"`) || !strings.Contains(got, ">press<") {
		t.Errorf("expected link, got %q", got)
	}
}

func TestToHTML_Image(t *testing.T) {
	got := ToHTML("![alt](img.png)")
	if !strings.Contains(got, `src="img.png"`) || !strings.Contains(got, `alt="alt"`) {
		t.Errorf("expected image, got %q", got)
	}
}

func TestToHTML_UnorderedList(t *testing.T) {
	md := "- item one\n- item two"
	got := ToHTML(md)
	if !strings.Contains(got, "<ul>") || !strings.Contains(got, "item one") {
		t.Errorf("expected unordered list, got %q", got)
	}
}

func TestToHTML_OrderedList(t *testing.T) {
	md := "1. first\n2. second"
	got := ToHTML(md)
	if !strings.Contains(got, "<ol>") || !strings.Contains(got, "first") {
		t.Errorf("expected ordered list, got %q", got)
	}
}

func TestToHTML_CodeBlock(t *testing.T) {
	md := "```\nfunc main() {}\n```"
	got := ToHTML(md)
	if !strings.Contains(got, "<pre>") || !strings.Contains(got, "func main()") {
		t.Errorf("expected code block, got %q", got)
	}
}

func TestToHTML_HorizontalRule(t *testing.T) {
	got := ToHTML("---")
	if !strings.Contains(got, "<hr") {
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

// GFM-specific features
func TestToHTML_Strikethrough(t *testing.T) {
	got := ToHTML("~~struck~~")
	if !strings.Contains(got, "<del>struck</del>") {
		t.Errorf("expected strikethrough, got %q", got)
	}
}

func TestToHTML_Table(t *testing.T) {
	md := "| A | B |\n|---|---|\n| 1 | 2 |"
	got := ToHTML(md)
	if !strings.Contains(got, "<table>") || !strings.Contains(got, "<td>") {
		t.Errorf("expected GFM table, got %q", got)
	}
}

func TestToHTML_TaskList(t *testing.T) {
	md := "- [x] done\n- [ ] todo"
	got := ToHTML(md)
	if !strings.Contains(got, `type="checkbox"`) {
		t.Errorf("expected GFM task list checkboxes, got %q", got)
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
