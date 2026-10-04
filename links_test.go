package main

import "testing"

func TestRenameDest(t *testing.T) {
	tests := []struct {
		dest, old, new string
		section        bool
		want           string
		changed        bool
	}{
		{"/about", "about", "about-us", false, "/about-us", true},
		{"/about.html#x", "about", "about-us", false, "/about-us.html#x", true},
		{"/about?q=1", "about", "about-us", false, "/about-us?q=1", true},
		{"/about-more", "about", "about-us", false, "/about-more", false},
		{"/about/x", "about", "about-us", false, "/about/x", false},
		{"about", "about", "about-us", false, "about", false},
		{"//about", "about", "about-us", false, "//about", false},
		{"/blog/a", "blog/a", "blog/b", false, "/blog/b", true},
		{"/blog", "blog", "journal", true, "/journal", true},
		{"/blog/", "blog", "journal", true, "/journal/", true},
		{"/blog/a.html#x", "blog", "journal", true, "/journal/a.html#x", true},
		{"/blogger", "blog", "journal", true, "/blogger", false},
		{"/a", "a", "my page", false, "/my%20page", true},
	}
	for _, tt := range tests {
		got, changed := renameDest(tt.dest, tt.old, tt.new, tt.section)
		if got != tt.want || changed != tt.changed {
			t.Errorf("renameDest(%q, %q, %q, %v) = %q, %v; want %q, %v", tt.dest, tt.old, tt.new, tt.section, got, changed, tt.want, tt.changed)
		}
	}
}

func TestRewriteLinks(t *testing.T) {
	in := "[a](/x \"T\") ![i](/x) `[c](/x)` <a href=\"/x\">h</a>\n```\n[f](/x)\n```\n[b](</x>) [m](/y)\n"
	want := "[a](/z \"T\") ![i](/x) `[c](/x)` <a href=\"/x\">h</a>\n```\n[f](/x)\n```\n[b](</z>) [m](/y)\n"
	got, n := rewriteLinks(in, func(d string) (string, bool) {
		if d == "/x" {
			return "/z", true
		}
		return d, false
	})
	if got != want || n != 2 {
		t.Errorf("rewriteLinks() = %q, %d; want %q, 2", got, n, want)
	}
}

func TestMaskCodeKeepsLength(t *testing.T) {
	in := "a `b` c\r\n~~~\r\nfenced\r\n~~~\r\nend"
	if got := maskCode(in); len(got) != len(in) {
		t.Errorf("maskCode changed the length: %d != %d", len(got), len(in))
	}
}
