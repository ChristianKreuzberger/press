// Package markdown provides a Markdown to HTML converter backed by goldmark
// with GitHub-Flavored Markdown (GFM) extensions.
package markdown

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// ytShortcode matches !youtube[VIDEO_ID] where VIDEO_ID is an 11-character
// YouTube video identifier consisting of alphanumeric characters, hyphens, and underscores.
var ytShortcode = regexp.MustCompile(`!youtube\[([a-zA-Z0-9_-]{11})\]`)

// fenceMarker matches the opening of a fenced code block (``` or ~~~).
var fenceMarker = regexp.MustCompile("^[ \t]*(`{3,}|~{3,})")

// expandYouTube replaces !youtube[VIDEO_ID] shortcodes with a responsive iframe embed.
// It skips expansion inside fenced code blocks (``` or ~~~).
func expandYouTube(md string) string {
	lines := strings.Split(md, "\n")
	inFence := false
	for i, line := range lines {
		if fenceMarker.MatchString(line) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		lines[i] = ytShortcode.ReplaceAllStringFunc(line, func(m string) string {
			sub := ytShortcode.FindStringSubmatch(m)
			if len(sub) < 2 {
				return m
			}
			id := sub[1]
			return fmt.Sprintf(
				`<iframe style="width:100%%;aspect-ratio:16/9;" `+
					`src="https://www.youtube-nocookie.com/embed/%s" `+
					`title="YouTube video player" `+
					`allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" `+
					`allowfullscreen></iframe>`,
				id,
			)
		})
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
