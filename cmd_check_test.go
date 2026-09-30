package main

import (
	"reflect"
	"testing"
)

func TestInternalLinks(t *testing.T) {
	tests := []struct {
		name, in string
		want     []string
	}{
		{"plain", "[a](/about)", []string{"/about"}},
		{"title", `[a](/about "T")`, []string{"/about"}},
		{"parens", "[a](/blog/first_(1))", []string{"/blog/first_(1)"}},
		{"angle", "[a](</my page> 'T')", []string{"/my page"}},
		{"image", "![a](/img.png)", nil},
		{"protocol-relative", "[a](//example.com)", nil},
		{"scheme", "[a](https://x.org) [b](mailto:a@b.c)", nil},
		{"relative", "[a](other.md)", nil},
		{"fenced", "```\n[a](/x)\n```\n[b](/y)", []string{"/y"}},
		{"tilde fenced", "~~~\n[a](/x)\n~~~", nil},
		{"inline code", "`[a](/x)` [b](/y)", []string{"/y"}},
		{"unclosed angle", "[a](</x", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := internalLinks(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("internalLinks(%q) = %q, want %q", tt.in, got, tt.want)
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
