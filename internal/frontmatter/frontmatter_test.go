package frontmatter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateSection(t *testing.T) {
	now := time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)
	got := string(GenerateSection("Blog", now))

	if !strings.Contains(got, `toc_sort: "weight"`) {
		t.Errorf("expected toc_sort field, got: %s", got)
	}
	if !strings.Contains(got, `toc_order: "asc"`) {
		t.Errorf("expected toc_order field, got: %s", got)
	}
	if !strings.Contains(got, `title: "Blog"`) {
		t.Errorf("expected title field, got: %s", got)
	}
	if !strings.HasPrefix(got, "---\n") {
		t.Error("expected frontmatter to start with ---")
	}
}

func TestGenerateSection_DoesNotContainTOCFields(t *testing.T) {
	// Regular Generate should NOT contain toc_sort or toc_order.
	now := time.Now().UTC()
	got := string(Generate("Page", now))

	if strings.Contains(got, "toc_sort") {
		t.Errorf("Generate() should not contain toc_sort, got: %s", got)
	}
	if strings.Contains(got, "toc_order") {
		t.Errorf("Generate() should not contain toc_order, got: %s", got)
	}
}

func TestParseStringField(t *testing.T) {
	tests := []struct {
		name    string
		content string
		field   string
		want    string
	}{
		{
			name:    "parses toc_sort title",
			content: "---\ntitle: \"Blog\"\ntoc_sort: \"title\"\n---\n",
			field:   "toc_sort",
			want:    "title",
		},
		{
			name:    "parses toc_order desc",
			content: "---\ntoc_order: \"desc\"\n---\n",
			field:   "toc_order",
			want:    "desc",
		},
		{
			name:    "parses toc_sort weight",
			content: "---\ntoc_sort: \"weight\"\n---\n",
			field:   "toc_sort",
			want:    "weight",
		},
		{
			name:    "returns empty when absent",
			content: "---\ntitle: \"Test\"\n---\n",
			field:   "toc_sort",
			want:    "",
		},
		{
			name:    "returns empty when no frontmatter",
			content: "# Hello\n",
			field:   "toc_sort",
			want:    "",
		},
		{
			name:    "strips surrounding quotes",
			content: "---\ntoc_sort: \"created_at\"\n---\n",
			field:   "toc_sort",
			want:    "created_at",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseStringField([]byte(tt.content), tt.field)
			if got != tt.want {
				t.Errorf("ParseStringField(%q) = %q; want %q", tt.field, got, tt.want)
			}
		})
	}
}

func TestParseTimeField(t *testing.T) {
	wantTime := time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)
	content := "---\ncreated_at: \"2026-04-22T10:00:00Z\"\n---\n"
	got := ParseTimeField([]byte(content), "created_at")
	if !got.Equal(wantTime) {
		t.Errorf("ParseTimeField(created_at) = %v; want %v", got, wantTime)
	}
}

func TestParseTimeField_UpdatedAt(t *testing.T) {
	wantTime := time.Date(2025, 1, 15, 8, 30, 0, 0, time.UTC)
	content := "---\nupdated_at: \"2025-01-15T08:30:00Z\"\n---\n"
	got := ParseTimeField([]byte(content), "updated_at")
	if !got.Equal(wantTime) {
		t.Errorf("ParseTimeField(updated_at) = %v; want %v", got, wantTime)
	}
}

func TestParseTimeField_AbsentReturnsZero(t *testing.T) {
	content := "---\ntitle: \"Test\"\n---\n"
	got := ParseTimeField([]byte(content), "created_at")
	if !got.IsZero() {
		t.Errorf("ParseTimeField() should return zero time when absent, got %v", got)
	}
}

func TestParseTimeField_NoFrontmatterReturnsZero(t *testing.T) {
	content := "# Hello\n"
	got := ParseTimeField([]byte(content), "created_at")
	if !got.IsZero() {
		t.Errorf("ParseTimeField() should return zero time without frontmatter, got %v", got)
	}
}

func TestParseTimeField_InvalidFormatReturnsZero(t *testing.T) {
	content := "---\ncreated_at: \"not-a-date\"\n---\n"
	got := ParseTimeField([]byte(content), "created_at")
	if !got.IsZero() {
		t.Errorf("ParseTimeField() should return zero time for invalid format, got %v", got)
	}
}

func TestGenerateSection_TimestampFormat(t *testing.T) {
	now := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
	got := string(GenerateSection("test", now))

	if !strings.Contains(got, "2026-12-31T23:59:59Z") {
		t.Errorf("expected RFC3339 timestamp in section frontmatter, got: %s", got)
	}
}

func TestGenerate(t *testing.T) {
	now := time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC)
	got := string(Generate("Home", now))

	want := "---\ntitle: \"Home\"\nalias: \"\"\ntags: []\nweight: 0\ncreated_at: \"2026-04-22T10:00:00Z\"\nupdated_at: \"2026-04-22T10:00:00Z\"\n---\n"
	if got != want {
		t.Errorf("Generate():\ngot:  %q\nwant: %q", got, want)
	}
}

func TestGenerate_TitleWithSpecialChars(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	got := string(Generate(`My "Blog" Post`, now))

	if !strings.Contains(got, "title:") {
		t.Error("expected frontmatter to contain title field")
	}
	if !strings.HasPrefix(got, "---\n") {
		t.Error("expected frontmatter to start with ---")
	}
	if !strings.Contains(got, "---\n") {
		t.Error("expected frontmatter to contain closing ---")
	}
}

func TestStrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "strips frontmatter",
			input: "---\ntitle: \"Hello\"\n---\n# Hello\n\nBody text.\n",
			want:  "# Hello\n\nBody text.\n",
		},
		{
			name:  "no frontmatter returned unchanged",
			input: "# Hello\n\nBody text.\n",
			want:  "# Hello\n\nBody text.\n",
		},
		{
			name:  "unclosed frontmatter returned unchanged",
			input: "---\ntitle: \"Hello\"\n# Hello\n",
			want:  "---\ntitle: \"Hello\"\n# Hello\n",
		},
		{
			name:  "empty body after frontmatter",
			input: "---\ntitle: \"Hello\"\n---\n",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Strip(tt.input)
			if got != tt.want {
				t.Errorf("Strip():\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestGenerate_CreatedAndUpdatedMatch(t *testing.T) {
	now := time.Now().UTC()
	got := string(Generate("page", now))

	ts := now.Format(time.RFC3339)
	if strings.Count(got, ts) != 2 {
		t.Errorf("expected created_at and updated_at to both equal %q\ngot: %s", ts, got)
	}
}

func TestGenerate_EmptyAliasAndTags(t *testing.T) {
	now := time.Now().UTC()
	got := string(Generate("test", now))

	if !strings.Contains(got, `alias: ""`) {
		t.Errorf("expected empty alias field, got: %s", got)
	}
	if !strings.Contains(got, "tags: []") {
		t.Errorf("expected empty tags field, got: %s", got)
	}
}

func TestGenerate_TimestampFormat(t *testing.T) {
	now := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
	got := string(Generate("test", now))

	if !strings.Contains(got, "2026-12-31T23:59:59Z") {
		t.Errorf("expected RFC3339 timestamp in output, got: %s", got)
	}
}

func TestGenerate_ContainsWeightField(t *testing.T) {
	now := time.Now().UTC()
	got := string(Generate("test", now))

	if !strings.Contains(got, "weight: 0") {
		t.Errorf("expected weight field in generated frontmatter, got: %s", got)
	}
}

func TestParseWeight(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
	}{
		{
			name:    "parses positive weight",
			content: "---\ntitle: \"Test\"\nweight: 5\n---\n# Content\n",
			want:    5,
		},
		{
			name:    "returns 0 when weight absent",
			content: "---\ntitle: \"Test\"\n---\n# Content\n",
			want:    0,
		},
		{
			name:    "returns 0 when no frontmatter",
			content: "# Content\n",
			want:    0,
		},
		{
			name:    "returns 0 for explicit weight 0",
			content: "---\ntitle: \"Test\"\nweight: 0\n---\n# Content\n",
			want:    0,
		},
		{
			name:    "returns 0 for non-integer weight",
			content: "---\ntitle: \"Test\"\nweight: abc\n---\n# Content\n",
			want:    0,
		},
		{
			name:    "parses weight 1",
			content: "---\ntitle: \"Test\"\nweight: 1\n---\n# Content\n",
			want:    1,
		},
		{
			name:    "parses generated frontmatter weight",
			content: "---\ntitle: \"Home\"\nalias: \"\"\ntags: []\nweight: 0\ncreated_at: \"2026-01-01T00:00:00Z\"\nupdated_at: \"2026-01-01T00:00:00Z\"\n---\n# Home\n",
			want:    0,
		},
		{
			name:    "parses large weight value",
			content: "---\nweight: 999\ntitle: \"Heavy\"\n---\n",
			want:    999,
		},
		{
			// Quoted integer: the old implementation returned 0 because strconv.Atoi
			// failed on `"7"`. The new implementation strips quotes via parseField,
			// so weight: "7" correctly returns 7.
			name:    "parses quoted integer",
			content: "---\nweight: \"7\"\n---\n",
			want:    7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseWeight([]byte(tt.content))
			if got != tt.want {
				t.Errorf("ParseWeight() = %d, want %d", got, tt.want)
			}
		})
	}
}
func TestParseStringField_UnclosedFrontmatter(t *testing.T) {
	// Frontmatter that starts with --- but has no closing delimiter returns "".
	content := "---\ntitle: \"Hello\"\n"
	got := ParseStringField([]byte(content), "title")
	if got != "" {
		t.Errorf("expected empty string for unclosed frontmatter, got %q", got)
	}
}

func TestParseDraft(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "draft true unquoted",
			content: "---\ndraft: true\n---\n",
			want:    true,
		},
		{
			name:    "draft true quoted",
			content: "---\ndraft: \"true\"\n---\n",
			want:    true,
		},
		{
			name:    "draft false unquoted",
			content: "---\ndraft: false\n---\n",
			want:    false,
		},
		{
			name:    "draft absent",
			content: "---\ntitle: \"Hello\"\n---\n",
			want:    false,
		},
		{
			name:    "no frontmatter",
			content: "# Hello\n",
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDraft([]byte(tt.content))
			if got != tt.want {
				t.Errorf("ParseDraft() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDraftFromFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "draft true unquoted",
			content: "---\ndraft: true\n---\n# Page\n\nLong body that should not be read.\n",
			want:    true,
		},
		{
			name:    "draft true quoted",
			content: "---\ndraft: \"true\"\n---\n",
			want:    true,
		},
		{
			name:    "draft false",
			content: "---\ndraft: false\n---\n",
			want:    false,
		},
		{
			name:    "draft absent",
			content: "---\ntitle: \"Hello\"\n---\n",
			want:    false,
		},
		{
			name:    "no frontmatter",
			content: "# Hello\n",
			want:    false,
		},
		{
			name:    "unclosed frontmatter",
			content: "---\ndraft: true\n",
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "page.md")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
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

func TestParseDraftFromFile_NotFound(t *testing.T) {
	_, err := ParseDraftFromFile("/nonexistent/path/page.md")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestHumanize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"about-us", "About Us"},
		{"my_cool_page", "My Cool Page"},
		{"blog", "Blog"},
		{"my-blog-post", "My Blog Post"},
		{"section_one_two", "Section One Two"},
		{"UPPER", "Upper"},
		{"mixed-Case_word", "Mixed Case Word"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Humanize(tt.input)
			if got != tt.want {
				t.Errorf("Humanize(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSetField(t *testing.T) {
	t.Run("updates title", func(t *testing.T) {
		content := "---\ntitle: \"Old Title\"\nupdated_at: \"2026-01-01T00:00:00Z\"\n---\n# Body\n"
		got, err := SetField([]byte(content), "title", "New Title")
		if err != nil {
			t.Fatalf("SetField() error: %v", err)
		}
		if !strings.Contains(string(got), `title: "New Title"`) {
			t.Errorf("expected updated title, got: %s", got)
		}
		if !strings.Contains(string(got), "# Body\n") {
			t.Errorf("expected body preserved, got: %s", got)
		}
	})

	t.Run("updates updated_at", func(t *testing.T) {
		content := "---\ntitle: \"Page\"\nupdated_at: \"2026-01-01T00:00:00Z\"\n---\n"
		got, err := SetField([]byte(content), "updated_at", "2027-06-15T12:00:00Z")
		if err != nil {
			t.Fatalf("SetField() error: %v", err)
		}
		if !strings.Contains(string(got), `updated_at: "2027-06-15T12:00:00Z"`) {
			t.Errorf("expected updated timestamp, got: %s", got)
		}
	})

	t.Run("error when no frontmatter", func(t *testing.T) {
		content := "# Hello\n"
		_, err := SetField([]byte(content), "title", "New")
		if err == nil {
			t.Error("expected error for content without frontmatter, got nil")
		}
	})

	t.Run("error when field absent", func(t *testing.T) {
		content := "---\ntitle: \"Page\"\n---\n"
		_, err := SetField([]byte(content), "updated_at", "2027-01-01T00:00:00Z")
		if err == nil {
			t.Error("expected error when field is absent, got nil")
		}
	})
}
