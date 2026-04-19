// Package markdown provides a simple Markdown to HTML converter.
package markdown

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

var (
	reBold       = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reItalic     = regexp.MustCompile(`\*([^*]+)\*`)
	reInlineCode = regexp.MustCompile("`([^`]+)`")
	reLink       = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reImage      = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	reOrdered    = regexp.MustCompile(`^\d+\. (.+)$`)
)

// ToHTML converts Markdown text to an HTML fragment.
func ToHTML(md string) string {
	lines := strings.Split(strings.ReplaceAll(md, "\r\n", "\n"), "\n")
	var buf strings.Builder
	state := "" // "", "ul", "ol", "p", "pre"
	var preLines []string

	closeState := func() {
		switch state {
		case "ul":
			buf.WriteString("</ul>\n")
		case "ol":
			buf.WriteString("</ol>\n")
		case "p":
			buf.WriteString("</p>\n")
		}
		state = ""
	}

	for _, line := range lines {
		// Fenced code block toggle
		if strings.HasPrefix(line, "```") {
			if state == "pre" {
				buf.WriteString(html.EscapeString(strings.Join(preLines, "\n")))
				buf.WriteString("</code></pre>\n")
				preLines = nil
				state = ""
			} else {
				closeState()
				buf.WriteString("<pre><code>")
				state = "pre"
			}
			continue
		}
		if state == "pre" {
			preLines = append(preLines, line)
			continue
		}

		// Blank line closes any open block
		if strings.TrimSpace(line) == "" {
			closeState()
			continue
		}

		// ATX headings: # through ######
		if strings.HasPrefix(line, "#") {
			level := 0
			for level < len(line) && line[level] == '#' {
				level++
			}
			if level <= 6 && level < len(line) && line[level] == ' ' {
				closeState()
				text := applyInline(strings.TrimSpace(line[level+1:]))
				fmt.Fprintf(&buf, "<h%d>%s</h%d>\n", level, text, level)
				continue
			}
		}

		// Horizontal rule
		stripped := strings.TrimSpace(line)
		if stripped == "---" || stripped == "***" || stripped == "___" {
			closeState()
			buf.WriteString("<hr>\n")
			continue
		}

		// Blockquote
		if strings.HasPrefix(line, "> ") {
			closeState()
			text := applyInline(strings.TrimSpace(line[2:]))
			fmt.Fprintf(&buf, "<blockquote><p>%s</p></blockquote>\n", text)
			continue
		}

		// Unordered list
		if len(line) >= 2 && (line[0] == '-' || line[0] == '*' || line[0] == '+') && line[1] == ' ' {
			if state != "ul" {
				closeState()
				buf.WriteString("<ul>\n")
				state = "ul"
			}
			text := applyInline(strings.TrimSpace(line[2:]))
			fmt.Fprintf(&buf, "<li>%s</li>\n", text)
			continue
		}

		// Ordered list
		if m := reOrdered.FindStringSubmatch(line); m != nil {
			if state != "ol" {
				closeState()
				buf.WriteString("<ol>\n")
				state = "ol"
			}
			text := applyInline(strings.TrimSpace(m[1]))
			fmt.Fprintf(&buf, "<li>%s</li>\n", text)
			continue
		}

		// Paragraph
		if state != "p" {
			closeState()
			buf.WriteString("<p>")
			state = "p"
		} else {
			buf.WriteString(" ")
		}
		buf.WriteString(applyInline(line))
	}

	closeState()
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

// applyInline applies inline Markdown formatting to a string.
func applyInline(s string) string {
	// Images must be processed before links (image syntax is a superset).
	s = reImage.ReplaceAllStringFunc(s, func(m string) string {
		sub := reImage.FindStringSubmatch(m)
		return fmt.Sprintf(`<img src="%s" alt="%s">`, html.EscapeString(sub[2]), html.EscapeString(sub[1]))
	})
	s = reLink.ReplaceAllStringFunc(s, func(m string) string {
		sub := reLink.FindStringSubmatch(m)
		return fmt.Sprintf(`<a href="%s">%s</a>`, html.EscapeString(sub[2]), sub[1])
	})
	// Bold before italic so ** is consumed first.
	s = reBold.ReplaceAllString(s, "<strong>$1</strong>")
	s = reItalic.ReplaceAllString(s, "<em>$1</em>")
	s = reInlineCode.ReplaceAllStringFunc(s, func(m string) string {
		sub := reInlineCode.FindStringSubmatch(m)
		return fmt.Sprintf("<code>%s</code>", html.EscapeString(sub[1]))
	})
	return s
}
