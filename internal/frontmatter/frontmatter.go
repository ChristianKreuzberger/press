package frontmatter

import (
	"fmt"
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
