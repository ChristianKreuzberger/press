package themes

import (
	"html/template"
	"strings"
	"testing"
)

func TestAllHasEntries(t *testing.T) {
	if len(All) == 0 {
		t.Fatal("All must contain at least one theme")
	}
}

func TestDefaultReturnsFirst(t *testing.T) {
	d := Default()
	if d.Name != All[0].Name {
		t.Errorf("Default() returned %q; want %q", d.Name, All[0].Name)
	}
}

func TestGetKnownThemes(t *testing.T) {
	for _, theme := range All {
		got, ok := Get(theme.Name)
		if !ok {
			t.Errorf("Get(%q) returned false", theme.Name)
			continue
		}
		if got.Name != theme.Name {
			t.Errorf("Get(%q).Name = %q", theme.Name, got.Name)
		}
		if got.Template == "" {
			t.Errorf("Get(%q).Template is empty", theme.Name)
		}
		if got.Description == "" {
			t.Errorf("Get(%q).Description is empty", theme.Name)
		}
	}
}

func TestGetUnknownTheme(t *testing.T) {
	_, ok := Get("does-not-exist")
	if ok {
		t.Error("Get(unknown) should return false")
	}
}

func TestNames(t *testing.T) {
	names := Names()
	if len(names) != len(All) {
		t.Errorf("Names() returned %d names; want %d", len(names), len(All))
	}
	for i, name := range names {
		if name != All[i].Name {
			t.Errorf("Names()[%d] = %q; want %q", i, name, All[i].Name)
		}
	}
}

func TestBuiltinThemeNames(t *testing.T) {
	want := []string{"dark", "light", "terminal"}
	names := Names()
	for _, w := range want {
		found := false
		for _, n := range names {
			if n == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected theme %q in Names(), got %v", w, names)
		}
	}
}

func TestTemplatesAreValidGoTemplates(t *testing.T) {
	for _, theme := range All {
		_, err := template.New(theme.Name).Parse(theme.Template)
		if err != nil {
			t.Errorf("theme %q has invalid Go template: %v", theme.Name, err)
		}
	}
}

func TestTemplatesContainRequiredPlaceholders(t *testing.T) {
	required := []string{"{{.Title}}", "{{.Content}}", "{{range .Pages}}"}
	for _, theme := range All {
		for _, placeholder := range required {
			if !strings.Contains(theme.Template, placeholder) {
				t.Errorf("theme %q template missing placeholder %q", theme.Name, placeholder)
			}
		}
	}
}
