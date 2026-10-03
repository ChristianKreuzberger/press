// Package markdown provides a Markdown to HTML converter backed by goldmark
// with GitHub-Flavored Markdown (GFM) extensions.
package markdown

import (
	"bytes"
	"fmt"
	stdhtml "html"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// ytShortcode matches !youtube[VIDEO_ID] where VIDEO_ID is an 11-character
// YouTube video identifier consisting of alphanumeric characters, hyphens, and underscores.
var ytShortcode = regexp.MustCompile(`^[ \t]*!youtube\[([a-zA-Z0-9_-]{11})\][ \t]*$`)

// fenceMarker matches the opening of a fenced code block (``` or ~~~).
var fenceMarker = regexp.MustCompile("^[ \t]*(`{3,}|~{3,})")

// expandYouTube replaces !youtube[VIDEO_ID] shortcodes with a responsive iframe embed.
// It skips expansion inside fenced code blocks (``` or ~~~).
func expandYouTube(md string) string {
	lines := strings.Split(md, "\n")
	inFence := false
	var fenceChar byte
	var fenceLen int

	for i, line := range lines {
		sub := fenceMarker.FindStringSubmatch(line)
		if sub != nil {
			run := sub[1]
			if !inFence {
				inFence = true
				fenceChar = run[0]
				fenceLen = len(run)
				continue
			}
			// Close only when same character and length >= opening length.
			if run[0] == fenceChar && len(run) >= fenceLen {
				inFence = false
				fenceChar = 0
				fenceLen = 0
			}
			continue
		}
		if inFence {
			continue
		}
		if m := ytShortcode.FindStringSubmatch(line); m != nil {
			id := m[1]
			lines[i] = fmt.Sprintf(
				`<iframe style="width:100%%;aspect-ratio:16/9;" `+
					`src="https://www.youtube-nocookie.com/embed/%s" `+
					`title="YouTube video player" `+
					`allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" `+
					`allowfullscreen></iframe>`,
				id,
			)
		}
	}
	return strings.Join(lines, "\n")
}

var gm = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.DefinitionList,
		extension.Footnote,
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
		// html.WithUnsafe is enabled because press renders author-controlled content only.
		// All Markdown is written by the site owner — never by untrusted third-party users.
		// This is required for the !youtube shortcode iframe to pass through the renderer.
		// SSGs such as Hugo and Zola apply the same trust model.
		html.WithUnsafe(),
	),
)

// ToHTML converts Markdown text to an HTML fragment using goldmark with
// GitHub-Flavored Markdown extensions (tables, task lists, strikethrough, etc.).
// It also expands !youtube[VIDEO_ID] shortcodes into embedded iframes.
func ToHTML(md string) string {
	md = expandYouTube(md)
	var buf bytes.Buffer
	if err := gm.Convert([]byte(md), &buf); err != nil {
		// Fallback: return escaped source on unexpected errors.
		return "<p>" + stdhtml.EscapeString(md) + "</p>"
	}
	return buf.String()
}

// ExtractTitle returns the plain text of the first level-1 heading in the
// Markdown, or an empty string if none is found. It parses the document, so
// "# comment" lines inside code blocks are ignored, and inline formatting,
// entities and raw HTML are reduced to text (the template escapes the result).
func ExtractTitle(md string) string {
	src := []byte(md)
	doc := gm.Parser().Parse(text.NewReader(src))
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if h, ok := n.(*ast.Heading); ok && h.Level == 1 {
			var b strings.Builder
			plainText(&b, h, src)
			return strings.Join(strings.Fields(b.String()), " ")
		}
	}
	return ""
}

// plainText appends the visible text under n to b.
func plainText(b *strings.Builder, n ast.Node, src []byte) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch v := c.(type) {
		case *ast.Text:
			// goldmark keeps entities raw in Text segments, so decode them
			// after resolving backslash escapes. (An escaped "\&amp;" ends up
			// as "&"; accepted as a rare edge case.)
			b.WriteString(stdhtml.UnescapeString(string(util.UnescapePunctuations(v.Segment.Value(src)))))
			if v.SoftLineBreak() || v.HardLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(v.Value)
		case *ast.AutoLink:
			// Autolinks have no child nodes; their text is the label.
			b.Write(v.Label(src))
		case *ast.RawHTML:
			// Tags carry no visible text.
		default:
			plainText(b, c, src)
		}
	}
}
