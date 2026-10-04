package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestPageLinks(t *testing.T) {
	tests := []struct {
		name, in string
		want     []string
	}{
		{"plain", "[a](/about)", []string{"/about"}},
		{"title", `[a](/about "T")`, []string{"/about"}},
		{"parens", "[a](/blog/first_(1))", []string{"/blog/first_(1)"}},
		{"angle", "[a](</my page> 'T')", []string{"/my page"}},
		{"image", "![a](/img.png)", nil},
		{"protocol-relative", "[a](//example.com)", []string{"//example.com"}},
		{"scheme", "[a](https://x.org) [b](mailto:a@b.c)", []string{"https://x.org", "mailto:a@b.c"}},
		{"relative", "[a](other.md)", []string{"other.md"}},
		{"html anchor", `<a href="/a">x</a> <A HREF='b.html'>y</A>`, []string{"/a", "b.html"}},
		{"html in code", "`<a href=\"/a\">`", nil},
		{"fenced", "```\n[a](/x)\n```\n[b](/y)", []string{"/y"}},
		{"tilde fenced", "~~~\n[a](/x)\n~~~", nil},
		{"inline code", "`[a](/x)` [b](/y)", []string{"/y"}},
		{"unclosed angle", "[a](</x", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pageLinks(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pageLinks(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestLinkDestination(t *testing.T) {
	tests := map[string]string{
		"/a)":       "/a",
		"/a \"t\")": "/a",
		"/a_(1))":   "/a_(1)",
		"</a b>)":   "/a b",
		"  /a)":     "/a",
		"/a":        "/a",
	}
	for in, want := range tests {
		if got := linkDestination(in); got != want {
			t.Errorf("linkDestination(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestInternalTarget(t *testing.T) {
	tests := []struct {
		dest, base, want string
		ok               bool
	}{
		{"/about", "/", "/about", true},
		{"/about/", "/", "/about", true},
		{"/", "/", "/index", true},
		{"/a#x", "/", "/a", true},
		{"/a?q=1", "/", "/a", true},
		{"about.html", "/", "/about.html", true},
		{"./about", "/", "/about", true},
		{"post", "/blog/", "/blog/post", true},
		{"../about", "/blog/", "/about", true},
		{"../../x", "/blog/", "/blog/../../x", true},
		{"../x", "/", "/../x", true},
		{"blog/", "/", "/blog", true},
		{"/blog/first_%281%29", "/", "/blog/first_(1)", true},
		{"#top", "/", "", false},
		{"", "/", "", false},
		{"//example.com", "/", "", false},
		{"https://x.org", "/", "", false},
		{"mailto:a@b.c", "/", "", false},
	}
	for _, tt := range tests {
		got, ok := internalTarget(tt.dest, tt.base)
		if got != tt.want || ok != tt.ok {
			t.Errorf("internalTarget(%q, %q) = %q, %v; want %q, %v", tt.dest, tt.base, got, ok, tt.want, tt.ok)
		}
	}
}

func TestCheckPage(t *testing.T) {
	valid := map[string]bool{"/about": true, "/index": true, "/blog": true, "/blog/post": true, "/my page": true}
	drafts := map[string]bool{"/wip": true}
	const fm = "---\ntitle: \"T\"\n---\n"
	tests := []struct {
		name, base, content string
		want                []string // substrings, one per expected issue
	}{
		{"valid link", "/", fm + "[a](/about)", nil},
		{"home link", "/", fm + "[home](/)", nil},
		{"broken link", "/", fm + "[a](/nope)", []string{"broken link → /nope (page not found)"}},
		{"draft link", "/", fm + "[a](/wip)", []string{"/wip (page is a draft"}},
		{"anchor only", "/", fm + "[a](#top)", nil},
		{"fragment and query", "/", fm + "[a](/about?x=1#y)", nil},
		{"external ignored", "/", fm + "[a](https://example.com/nope) [b](mailto:a@b.c) [c](//cdn.example/x)", nil},
		{"code ignored", "/", fm + "`[a](/nope)`\n\n```\n[b](/nope)\n```\ntext", nil},
		{"relative in section", "/blog/", fm + "[a](post)", nil},
		{"relative parent", "/blog/", fm + "[a](../about)", nil},
		{"relative above root", "/", fm + "[a](../about)", []string{"broken link"}},
		{"percent encoded", "/", fm + "[a](/my%20page)", nil},
		{"section link with slash", "/", fm + "[a](/blog/)", nil},
		{"missing title", "/", "---\n---\nbody", []string{"missing title"}},
		{"empty content", "/", "---\ntitle: \"T\"\n---\n  \n", []string{"empty page content"}},
		{"two broken links", "/", fm + "[a](/x) [b](/y)", []string{"/x", "/y"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkPage("pages/p.md", tt.base, []byte(tt.content), valid, drafts)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d issues %q, want %d", len(got), got, len(tt.want))
			}
			for i, w := range tt.want {
				if !strings.Contains(got[i], w) {
					t.Errorf("issue %d = %q, want it to contain %q", i, got[i], w)
				}
				if !strings.HasPrefix(got[i], "pages/p.md: ") {
					t.Errorf("issue %q should be prefixed with the page path", got[i])
				}
			}
		})
	}
}

func TestBuildValidPaths(t *testing.T) {
	site := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		p := filepath.Join(site, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	draft := "---\ntitle: \"D\"\ndraft: true\n---\nbody"
	write("pages/index.md", "# Home")
	write("pages/about.md", "# About")
	write("pages/wip.md", draft)
	write("pages/notes.txt", "not a page")
	write("pages/docs/index.md", "# Docs")
	write("pages/docs/guide.md", "# Guide")
	write("pages/docs/secret.md", draft)
	write("pages/docs/f.pdf", "pdf")
	write("pages/hidden/guide.md", "# no index, not a section")
	write("pages/closed/index.md", draft)
	write("pages/closed/page.md", "# in a draft section")

	valid, drafts := buildValidPaths(site)

	for _, p := range []string{
		"/index", "/index.html", "/about", "/about.html",
		"/docs", "/docs/index.html", "/docs/guide", "/docs/guide.html",
		"/docs/f.pdf", "/notes.txt",
	} {
		if !valid[p] {
			t.Errorf("expected %q to be valid", p)
		}
	}
	for _, p := range []string{"/wip", "/wip.html", "/docs/secret", "/closed", "/closed/page"} {
		if !drafts[p] || valid[p] {
			t.Errorf("expected %q only in drafts (drafts=%v valid=%v)", p, drafts[p], valid[p])
		}
	}
	for _, p := range []string{"/docs/index", "/hidden", "/hidden/guide", "/about.md"} {
		if valid[p] || drafts[p] {
			t.Errorf("did not expect %q to be known", p)
		}
	}
}

func TestBuildValidPathsNoPagesDir(t *testing.T) {
	valid, drafts := buildValidPaths(t.TempDir())
	if len(valid) != 0 || len(drafts) != 0 {
		t.Errorf("expected empty sets, got %v %v", valid, drafts)
	}
}
