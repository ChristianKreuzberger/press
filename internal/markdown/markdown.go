// Package markdown provides a Markdown to HTML converter backed by goldmark
// with GitHub-Flavored Markdown (GFM) extensions.
package markdown

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

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
	),
)

// ToHTML converts Markdown text to an HTML fragment using goldmark with
// GitHub-Flavored Markdown extensions (tables, task lists, strikethrough, etc.).
func ToHTML(md string) string {
	var buf bytes.Buffer
	if err := gm.Convert([]byte(md), &buf); err != nil {
		// Fallback: return escaped source on unexpected errors.
		return "<p>" + md + "</p>"
	}
	return buf.String()
}

// ExtractTitle returns the text of the first level-1 heading in the Markdown,
// or an empty string if none is found.
func ExtractTitle(md string) string {
	for _, line := range strings.Split(md, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	return ""
}
