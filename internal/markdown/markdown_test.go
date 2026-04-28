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
		closeTag := "</" + c.wantTag + ">"
		if !strings.Contains(got, open) || !strings.Contains(got, closeTag) || !strings.Contains(got, c.wantText) {
			t.Errorf("ToHTML(%q) = %q; want %s...%s containing %q", c.input, got, open, closeTag, c.wantText)
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

// GFM-specific features.
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

func TestToHTML_YouTubeShortcode(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantSrc string
		wantIn  bool
	}{
		{
			name:    "basic shortcode",
			input:   "!youtube[dQw4w9WgXcQ]",
			wantSrc: "https://www.youtube-nocookie.com/embed/dQw4w9WgXcQ",
			wantIn:  true,
		},
		{
			name:    "shortcode with surrounding text",
			input:   "Watch this:\n\n!youtube[dQw4w9WgXcQ]\n\nEnd.",
			wantSrc: "https://www.youtube-nocookie.com/embed/dQw4w9WgXcQ",
			wantIn:  true,
		},
		{
			name:    "invalid id too short",
			input:   "!youtube[short]",
			wantSrc: "youtube-nocookie.com/embed/short",
			wantIn:  false,
		},
		{
			name:    "invalid id too long",
			input:   "!youtube[toolongvideoidstring]",
			wantSrc: "youtube-nocookie.com/embed/toolongvideoidstring",
			wantIn:  false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ToHTML(c.input)
			if c.wantIn && !strings.Contains(got, c.wantSrc) {
				t.Errorf("ToHTML(%q) = %q; expected to contain %q", c.input, got, c.wantSrc)
			}
			if !c.wantIn && strings.Contains(got, c.wantSrc) {
				t.Errorf("ToHTML(%q) = %q; expected NOT to contain %q", c.input, got, c.wantSrc)
			}
		})
	}
}

func TestToHTML_YouTubeShortcode_IframeAttributes(t *testing.T) {
	got := ToHTML("!youtube[dQw4w9WgXcQ]")
	checks := []string{
		"<iframe",
		"allowfullscreen",
		"youtube-nocookie.com/embed/dQw4w9WgXcQ",
		"aspect-ratio:16/9",
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("ToHTML youtube shortcode missing %q in output: %q", want, got)
		}
	}
	if strings.Contains(got, "frameborder") {
		t.Errorf("ToHTML youtube shortcode should not contain deprecated frameborder attribute, got %q", got)
	}
	if strings.Contains(got, `width="560"`) {
		t.Errorf("ToHTML youtube shortcode should not contain hardcoded width, got %q", got)
	}
}

func TestToHTML_YouTubeShortcode_SkipsCodeBlock(t *testing.T) {
	md := "```\n!youtube[dQw4w9WgXcQ]\n```"
	got := ToHTML(md)
	if strings.Contains(got, "<iframe") {
		t.Errorf("ToHTML should not expand youtube shortcode inside fenced code block, got %q", got)
	}
	if !strings.Contains(got, "!youtube[dQw4w9WgXcQ]") {
		t.Errorf("ToHTML should preserve shortcode text inside fenced code block, got %q", got)
	}
}

func TestExpandYouTube(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		wantIn    string
		wantNotIn string
	}{
		{
			name:   "replaces valid shortcode",
			input:  "!youtube[dQw4w9WgXcQ]",
			wantIn: "dQw4w9WgXcQ",
		},
		{
			name:   "leaves non-shortcode text unchanged",
			input:  "just text",
			wantIn: "just text",
		},
		{
			// abc-DEF_123 is exactly 11 characters: a,b,c,-,D,E,F,_,1,2,3
			name:   "handles hyphens and underscores in id",
			input:  "!youtube[abc-DEF_123]",
			wantIn: "youtube-nocookie.com/embed/abc-DEF_123",
		},
		{
			name:      "skips shortcode inside fenced code block",
			input:     "```\n!youtube[dQw4w9WgXcQ]\n```",
			wantNotIn: "<iframe",
			wantIn:    "!youtube[dQw4w9WgXcQ]",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := expandYouTube(c.input)
			if c.wantIn != "" && !strings.Contains(got, c.wantIn) {
				t.Errorf("expandYouTube(%q) = %q; expected to contain %q", c.input, got, c.wantIn)
			}
			if c.wantNotIn != "" && strings.Contains(got, c.wantNotIn) {
				t.Errorf("expandYouTube(%q) = %q; expected NOT to contain %q", c.input, got, c.wantNotIn)
			}
		})
	}
}

func TestToHTML_YouTubeShortcode_SkipsInlineCode(t *testing.T) {
	got := ToHTML("`!youtube[dQw4w9WgXcQ]`")
	if strings.Contains(got, "<iframe") {
		t.Errorf("shortcode inside inline code must not expand, got %q", got)
	}
	if !strings.Contains(got, "!youtube[dQw4w9WgXcQ]") {
		t.Errorf("shortcode text must be preserved inside inline code, got %q", got)
	}
}

func TestExpandYouTube_FenceCharMismatch(t *testing.T) {
	// ~~~ fence opened; ``` line must NOT close it.
	md := "~~~\n!youtube[dQw4w9WgXcQ]\n```\n!youtube[dQw4w9WgXcQ]\n~~~"
	got := expandYouTube(md)
	if strings.Contains(got, "<iframe") {
		t.Errorf("shortcode inside ~~~-fenced block must not expand, got %q", got)
	}
}

func TestExpandYouTube_FenceLengthMismatch(t *testing.T) {
	// ```` fence (4 backticks); ``` closer (3 backticks) must NOT close it.
	md := "````\n!youtube[dQw4w9WgXcQ]\n```\nstill inside\n````"
	got := expandYouTube(md)
	if strings.Contains(got, "<iframe") {
		t.Errorf("shortcode inside 4-backtick fence must not expand, got %q", got)
	}
}

func TestExpandYouTube_MidParagraph(t *testing.T) {
	// Shortcode mid-sentence must not expand.
	md := "Watch !youtube[dQw4w9WgXcQ] here."
	got := expandYouTube(md)
	if strings.Contains(got, "<iframe") {
		t.Errorf("mid-paragraph shortcode must not expand, got %q", got)
	}
	if !strings.Contains(got, "!youtube[dQw4w9WgXcQ]") {
		t.Errorf("original text must be preserved, got %q", got)
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
