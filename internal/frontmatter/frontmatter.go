package frontmatter

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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

// parseField scans the frontmatter block for a line matching "field: value" and
// returns the value with surrounding double-quotes stripped. Returns empty string
// when the field is absent or there is no frontmatter block.
func parseField(content []byte, field string) string {
	s := string(content)
	const delim = "---"
	if !strings.HasPrefix(s, delim+"\n") {
		return ""
	}
	rest := s[len(delim)+1:]
	end := strings.Index(rest, "\n"+delim)
	if end == -1 {
		return ""
	}
	block := rest[:end]
	prefix := field + ":"
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
			if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
				val = val[1 : len(val)-1]
			}
			return val
		}
	}
	return ""
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
	return parseField(content, "draft") == "true"
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

// Strip removes YAML frontmatter from the beginning of a markdown document.
// If the content does not start with "---\n", it is returned unchanged.
func Strip(content string) string {
	const delim = "---"
	if !strings.HasPrefix(content, delim+"\n") {
		return content
	}
	// Find the closing delimiter after the opening one.
	rest := content[len(delim)+1:]
	idx := strings.Index(rest, "\n"+delim)
	if idx == -1 {
		return content
	}
	after := rest[idx+1+len(delim):]
	// Skip optional trailing newline after closing delimiter.
	after = strings.TrimPrefix(after, "\n")
	return after
}
