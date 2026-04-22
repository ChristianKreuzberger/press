package frontmatter

import (
	"fmt"
	"strings"
	"time"
)

// Generate returns YAML frontmatter bytes for a new markdown file.
// title is used as-is (the page/section name).
// now is the timestamp used for both created_at and updated_at.
func Generate(title string, now time.Time) []byte {
	ts := now.UTC().Format(time.RFC3339)
	s := fmt.Sprintf("---\ntitle: %q\nalias: \"\"\ntags: []\ncreated_at: %q\nupdated_at: %q\n---\n",
		title, ts, ts)
	return []byte(s)
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
