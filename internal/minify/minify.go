// Package minify provides HTML minification utilities.
package minify

import (
	"regexp"
	"strings"
)

// reTagWhitespace collapses whitespace runs between adjacent HTML tags.
var reTagWhitespace = regexp.MustCompile(`>\s+<`)

// HTML minifies an HTML string by:
//   - Removing HTML comments (but not IE conditional comments <!--[if ...]>)
//   - Trimming leading/trailing whitespace from lines outside <pre> blocks
//   - Skipping blank lines outside <pre> blocks
//   - Collapsing whitespace between adjacent HTML tags
//
// Content inside <pre> blocks is left untouched to preserve code indentation.
func HTML(s string) string {
	// Remove HTML comments, preserving IE conditionals.
	s = removeComments(s)

	var sb strings.Builder
	preDepth := 0
	lines := strings.Split(s, "\n")

	for _, line := range lines {
		// Count opening and closing <pre> tags in a single pass.
		opens, closes := countPreTags(line)

		if preDepth > 0 {
			// Inside a pre block — write the line verbatim (preserve indentation).
			sb.WriteString(line)
			sb.WriteByte('\n')
		} else {
			// Outside a pre block — trim and skip blanks.
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				sb.WriteString(trimmed)
				sb.WriteByte('\n')
			}
		}

		preDepth += opens - closes
		if preDepth < 0 {
			preDepth = 0
		}
	}

	result := sb.String()
	// Collapse whitespace runs between adjacent HTML tags.
	result = reTagWhitespace.ReplaceAllString(result, "><")
	return strings.TrimSpace(result)
}

// countPreTags counts opening <pre…> and closing </pre> tags in a single
// case-insensitive scan, returning (opens, closes).
// Opening tags are recognised when <pre is followed by '>', whitespace, or EOF.
// Closing tags match </pre> exactly (case-insensitive).
func countPreTags(line string) (opens, closes int) {
	lower := strings.ToLower(line)
	for i := 0; i < len(lower); {
		idx := strings.Index(lower[i:], "<")
		if idx < 0 {
			break
		}
		pos := i + idx
		rest := lower[pos:]
		switch {
		case strings.HasPrefix(rest, "</pre>"):
			closes++
			i = pos + 6
		case strings.HasPrefix(rest, "<pre>") ||
			(len(rest) > 4 && rest[:4] == "<pre" && isSpaceOrEnd(rest[4])):
			opens++
			i = pos + 4
		default:
			i = pos + 1
		}
	}
	return opens, closes
}

// isSpaceOrEnd reports whether b is a whitespace character.
func isSpaceOrEnd(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// removeComments removes HTML comments from s, but preserves IE conditional
// comments of the form <!--[if ...]>.
func removeComments(s string) string {
	var sb strings.Builder
	for {
		start := strings.Index(s, "<!--")
		if start < 0 {
			sb.WriteString(s)
			break
		}
		// Write everything before the comment.
		sb.WriteString(s[:start])
		s = s[start:]

		// Check if this is an IE conditional comment: <!--[if
		if strings.HasPrefix(s, "<!--[if") {
			// Preserve this comment — find its end and emit it.
			end := strings.Index(s, "-->")
			if end < 0 {
				// Unclosed comment — emit as-is.
				sb.WriteString(s)
				break
			}
			sb.WriteString(s[:end+3])
			s = s[end+3:]
			continue
		}

		// Regular comment — find its end and skip it.
		end := strings.Index(s, "-->")
		if end < 0 {
			// Unclosed comment — drop the rest.
			break
		}
		s = s[end+3:]
	}
	return sb.String()
}
