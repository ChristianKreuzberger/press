package main

import (
	"reflect"
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
