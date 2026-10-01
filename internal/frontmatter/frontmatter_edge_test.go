package frontmatter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseStringFieldQuotingAndComments(t *testing.T) {
	tests := []struct {
		name, line, want string
	}{
		{"escaped double quotes", `title: "My \"Blog\" Post"`, `My "Blog" Post`},
		{"escaped backslash", `title: "a\\b"`, `a\b`},
		{"quoted then comment", `title: "a \"b\"" # note`, `a "b"`},
		{"single quote doubled", `title: 'it''s'`, `it's`},
		{"tab before comment", "title: true\t# wip", "true"},
		{"tab and space before comment", "title: true \t# wip", "true"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseStringField([]byte("---\n"+tc.line+"\n---\n"), "title")
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGenerateRoundTripsSpecialTitle(t *testing.T) {
	title := `My "Blog" \ Post`
	if got := ParseStringField(Generate(title, time.Now()), "title"); got != title {
		t.Errorf("round trip: got %q, want %q", got, title)
	}
}

// draftCases are inputs that both ParseDraft (in-memory) and
// ParseDraftFromFile (file based) must agree on.
var draftCases = []struct {
	name    string
	content string
	want    bool
}{
	{"CRLF draft true", "---\r\ndraft: true\r\n---\r\n# Hi\r\n", true},
	{"CRLF draft false", "---\r\ndraft: false\r\n---\r\n", false},
	{"empty block", "---\n---\n# Hi\n", false},
	{"loose ---- is not a closing delimiter", "---\ntitle: x\n----\ndraft: true\n---\nbody\n", true},
	{"---foo is not a closing delimiter", "---\n---foo\ndraft: true\n---\n", true},
	{"indented key ignored", "---\nmeta:\n  draft: true\n---\n", false},
	{"trailing comment", "---\ndraft: true # wip\n---\n", true},
	{"single quotes", "---\ndraft: 'true'\n---\n", true},
	{"capitalised True", "---\ndraft: True\n---\n", true},
	{"quoted with comment", "---\ndraft: \"true\" # wip\n---\n", true},
	{"long line", "---\ntitle: " + strings.Repeat("a", 200*1024) + "\ndraft: true\n---\n", true},
	{"closing delimiter at EOF without newline", "---\ndraft: true\n---", true},
}

func TestDraftEdgeCases(t *testing.T) {
	for _, tt := range draftCases {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseDraft([]byte(tt.content)); got != tt.want {
				t.Errorf("ParseDraft() = %v, want %v", got, tt.want)
			}
			path := filepath.Join(t.TempDir(), "p.md")
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatal(err)
			}
			got, err := ParseDraftFromFile(path)
			if err != nil {
				t.Fatalf("ParseDraftFromFile() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ParseDraftFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStripEdgeCases(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"CRLF", "---\r\ntitle: x\r\n---\r\n# Body\r\n", "# Body\r\n"},
		{"empty block", "---\n---\n# Body\n", "# Body\n"},
		{"empty block CRLF", "---\r\n---\r\n# Body\r\n", "# Body\r\n"},
		{"loose ---- is not closing", "---\ntitle: x\n----\n---\nBody\n", "Body\n"},
		{"---foo is not closing", "---\n---foo\n---\nBody\n", "Body\n"},
		{"only loose delimiter means unclosed", "---\ntitle: x\n----\nBody\n", "---\ntitle: x\n----\nBody\n"},
		{"closing at EOF", "---\ntitle: x\n---", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Strip(tt.in); got != tt.want {
				t.Errorf("Strip():\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestParseFieldEdgeCases(t *testing.T) {
	t.Run("CRLF title and weight", func(t *testing.T) {
		c := []byte("---\r\ntitle: \"Hello\"\r\nweight: 5\r\n---\r\nbody\r\n")
		if got := ParseStringField(c, "title"); got != "Hello" {
			t.Errorf("title = %q", got)
		}
		if got := ParseWeight(c); got != 5 {
			t.Errorf("weight = %d", got)
		}
	})
	t.Run("indented key is not top-level", func(t *testing.T) {
		c := []byte("---\nmeta:\n  title: nested\n---\n")
		if got := ParseStringField(c, "title"); got != "" {
			t.Errorf("title = %q, want empty", got)
		}
	})
	t.Run("single quoted value", func(t *testing.T) {
		if got := ParseStringField([]byte("---\ntitle: 'Hi'\n---\n"), "title"); got != "Hi" {
			t.Errorf("title = %q", got)
		}
	})
	t.Run("trailing comment stripped, hash inside quotes kept", func(t *testing.T) {
		if got := ParseStringField([]byte("---\ntitle: Hi # c\n---\n"), "title"); got != "Hi" {
			t.Errorf("title = %q", got)
		}
		if got := ParseStringField([]byte("---\ntitle: \"A # B\"\n---\n"), "title"); got != "A # B" {
			t.Errorf("title = %q", got)
		}
	})
	t.Run("weight with comment", func(t *testing.T) {
		if got := ParseWeight([]byte("---\nweight: 3 # first\n---\n")); got != 3 {
			t.Errorf("weight = %d", got)
		}
	})
}

func TestSetFieldEdgeCases(t *testing.T) {
	t.Run("CRLF preserved", func(t *testing.T) {
		in := "---\r\ntitle: \"Old\"\r\nupdated_at: \"x\"\r\n---\r\n# Body\r\n"
		got, err := SetField([]byte(in), "title", "New")
		if err != nil {
			t.Fatal(err)
		}
		want := "---\r\ntitle: \"New\"\r\nupdated_at: \"x\"\r\n---\r\n# Body\r\n"
		if string(got) != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
	t.Run("empty block has no field", func(t *testing.T) {
		if _, err := SetField([]byte("---\n---\n"), "title", "x"); err == nil {
			t.Error("expected ErrFieldNotFound")
		}
	})
	t.Run("loose ---- does not close block", func(t *testing.T) {
		in := "---\ntitle: \"a\"\n----\nupdated_at: \"b\"\n---\n"
		got, err := SetField([]byte(in), "updated_at", "c")
		if err != nil {
			t.Fatal(err)
		}
		if want := "---\ntitle: \"a\"\n----\nupdated_at: \"c\"\n---\n"; string(got) != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
	t.Run("indented key not updated", func(t *testing.T) {
		in := "---\nmeta:\n  title: \"n\"\n---\n"
		if _, err := SetField([]byte(in), "title", "x"); err == nil {
			t.Error("expected ErrFieldNotFound for nested key")
		}
	})
}
