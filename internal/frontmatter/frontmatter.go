// Package frontmatter parses and edits YAML-style frontmatter in markdown files.
package frontmatter

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// ErrNoFrontmatter is returned by SetField when the content has no frontmatter block.
var ErrNoFrontmatter = errors.New("frontmatter: no frontmatter block found")

// ErrFieldNotFound is returned by SetField when the named field is absent from the frontmatter.
var ErrFieldNotFound = errors.New("frontmatter: field not found")

// Generate returns YAML frontmatter bytes for a new markdown file.
// title is used as-is (the page/section name).
// now is the timestamp used for both created_at and updated_at.
func Generate(title string, now time.Time) []byte {
	ts := now.UTC().Format(time.RFC3339)
	s := fmt.Sprintf("---\ntitle: %q\nalias: \"\"\ntags: []\nweight: 0\ncreated_at: %q\nupdated_at: %q\n---\n",
		title, ts, ts)
	return []byte(s)
}

// GenerateSection returns YAML frontmatter bytes for a new section index file.
// It includes toc_sort and toc_order fields in addition to the standard fields.
func GenerateSection(title string, now time.Time) []byte {
	ts := now.UTC().Format(time.RFC3339)
	s := fmt.Sprintf("---\ntitle: %q\nalias: \"\"\ntags: []\nweight: 0\ncreated_at: %q\nupdated_at: %q\ntoc_sort: \"weight\"\ntoc_order: \"asc\"\n---\n",
		title, ts, ts)
	return []byte(s)
}

// delim is the line that opens and closes a frontmatter block.
const delim = "---"

// isDelimLine reports whether line (without its "\n") is exactly "---",
// tolerating a trailing "\r" from CRLF files.
func isDelimLine(line string) bool {
	return strings.TrimSuffix(line, "\r") == delim
}

// split locates the frontmatter block in s. s[blockStart:blockEnd] is the text
// between the delimiter lines, and bodyStart is the offset of the first byte
// after the closing delimiter line. ok is false when s does not start with a
// delimiter line or the block is never closed. Both "\n" and "\r\n" line
// endings are accepted and the closing delimiter must be a whole line.
func split(s string) (blockStart, blockEnd, bodyStart int, ok bool) {
	nl := strings.IndexByte(s, '\n')
	if nl == -1 || !isDelimLine(s[:nl]) {
		return 0, 0, 0, false
	}
	blockStart = nl + 1
	for pos := blockStart; pos < len(s); {
		line, next := s[pos:], len(s)
		if end := strings.IndexByte(s[pos:], '\n'); end != -1 {
			line, next = s[pos:pos+end], pos+end+1
		}
		if isDelimLine(line) {
			return blockStart, pos, next, true
		}
		pos = next
	}
	return 0, 0, 0, false
}

// fieldValue returns the value of a top-level "field: value" line in block.
func fieldValue(block, field string) string {
	prefix := field + ":"
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSuffix(line, "\r")
		// Column 0 only, so nested keys never match as top-level ones.
		if strings.HasPrefix(line, prefix) {
			return cleanValue(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

// cleanValue trims a raw YAML scalar: surrounding quotes (single or double)
// are removed, and an unquoted value loses any trailing " # comment".
func cleanValue(val string) string {
	val = strings.TrimSpace(val)
	if val == "" {
		return ""
	}
	switch val[0] {
	case '"':
		// Generate and SetField write strconv-quoted values, so the closing
		// quote must be found escape-aware and the content unescaped.
		for i := 1; i < len(val); i++ {
			switch val[i] {
			case '\\':
				i++
			case '"':
				if s, err := strconv.Unquote(val[:i+1]); err == nil {
					return s
				}
				return val[1:i]
			}
		}
		return val
	case '\'':
		// In single-quoted YAML a doubled quote is a literal quote.
		var b strings.Builder
		for i := 1; i < len(val); i++ {
			switch {
			case val[i] != '\'':
				b.WriteByte(val[i])
			case i+1 < len(val) && val[i+1] == '\'':
				b.WriteByte('\'')
				i++
			default:
				return b.String()
			}
		}
		return val
	}
	// A comment starts at '#' preceded by a space or tab.
	for i := 0; i+1 < len(val); i++ {
		if (val[i] == ' ' || val[i] == '\t') && val[i+1] == '#' {
			val = val[:i]
			break
		}
	}
	return strings.TrimSpace(val)
}

// parseField returns the value of a top-level frontmatter field with quotes and
// trailing comments removed. Returns empty string when the field is absent or
// there is no frontmatter block.
func parseField(content []byte, field string) string {
	s := string(content)
	start, end, _, ok := split(s)
	if !ok {
		return ""
	}
	return fieldValue(s[start:end], field)
}

// ParseStringField extracts the string value of a named field from YAML frontmatter.
// Returns empty string when the field is absent or there is no frontmatter.
func ParseStringField(content []byte, field string) string {
	return parseField(content, field)
}

// ParseTimeField extracts an RFC3339 timestamp from a named field in YAML frontmatter.
// Returns the zero time.Time when the field is absent, empty, or unparseable.
func ParseTimeField(content []byte, field string) time.Time {
	val := parseField(content, field)
	if val == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return time.Time{}
	}
	return t
}

// ParseDraft reports whether the frontmatter contains "draft: true".
// Both unquoted (draft: true) and quoted (draft: "true") values are accepted.
// Returns false when the field is absent or set to any other value.
func ParseDraft(content []byte) bool {
	return isTrue(parseField(content, "draft"))
}

// isTrue reports whether a cleaned scalar is a YAML-style true ("true", "True", "TRUE").
func isTrue(val string) bool { return strings.EqualFold(val, "true") }

// ParseDraftFromFile opens the file at path and reads only the frontmatter
// block (up to and including the closing "---" delimiter) to determine
// whether "draft: true" is set. This avoids loading the full file into memory
// when only the draft flag is needed.
// Returns false (and no error) when the file has no frontmatter.
func ParseDraftFromFile(path string) (bool, error) {
	f, err := os.Open(path) //nolint:gosec // path is a page file discovered under the site directory
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }() // read-only file, close error is not actionable

	// bufio.Reader (unlike Scanner) has no line length limit.
	r := bufio.NewReader(f)
	var sb strings.Builder
	for first := true; ; first = false {
		line, err := r.ReadString('\n')
		sb.WriteString(line)
		isDelim := isDelimLine(strings.TrimSuffix(line, "\n"))
		if first && !isDelim {
			return false, nil
		}
		if !first && isDelim {
			// Closing delimiter reached; reuse the shared splitter.
			return isTrue(parseField([]byte(sb.String()), "draft")), nil
		}
		if err == io.EOF {
			// No closing delimiter found: treat as no frontmatter.
			return false, nil
		}
		if err != nil {
			return false, err
		}
	}
}

// ParseWeight extracts the weight field value from YAML frontmatter.
// Returns 0 if the field is absent, unparseable, or no frontmatter block is found.
// Quoted integers (e.g. weight: "5") are accepted and return 5.
func ParseWeight(content []byte) int {
	val := parseField(content, "weight")
	if val == "" {
		return 0
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return n
}

// Humanize converts a slug-style name to a human-readable title.
// Hyphens and underscores are replaced with spaces, and each word is title-cased.
func Humanize(name string) string {
	r := strings.NewReplacer("-", " ", "_", " ")
	words := strings.Fields(r.Replace(name))
	for i, w := range words {
		runes := []rune(w)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		for j := 1; j < len(runes); j++ {
			runes[j] = unicode.ToLower(runes[j])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

// SetField updates the value of a named top-level field in the YAML
// frontmatter block. The field must already exist in the frontmatter; the new
// value is written as a double-quoted string and the file's line endings
// (LF or CRLF) are preserved. Returns an error if there is no frontmatter or
// the field is absent.
func SetField(content []byte, field, value string) ([]byte, error) {
	s := string(content)
	start, end, _, ok := split(s)
	if !ok {
		return nil, ErrNoFrontmatter
	}
	prefix := field + ":"
	found := false
	lines := strings.Split(s[start:end], "\n")
	for i, line := range lines {
		if !found && strings.HasPrefix(line, prefix) {
			lines[i] = field + ": " + strconv.Quote(value)
			if strings.HasSuffix(line, "\r") {
				lines[i] += "\r"
			}
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("%w: %q", ErrFieldNotFound, field)
	}
	return []byte(s[:start] + strings.Join(lines, "\n") + s[end:]), nil
}

// Strip removes YAML frontmatter from the beginning of a markdown document.
// If the content has no complete frontmatter block, it is returned unchanged.
func Strip(content string) string {
	_, _, bodyStart, ok := split(content)
	if !ok {
		return content
	}
	return content[bodyStart:]
}
