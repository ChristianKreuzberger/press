package main

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// linkStartRe matches the start of a Markdown inline link, up to and including
// the "(" before the destination. Group 1 is "!" for image links.
var linkStartRe = regexp.MustCompile(`(!?)\[[^\]]*\]\(`)

// inlineCodeRe matches inline code spans so links inside them can be ignored.
var inlineCodeRe = regexp.MustCompile("`[^`\n]*`")

// htmlHrefRe matches the href of a raw HTML anchor. Group 2 or 3 is the value.
var htmlHrefRe = regexp.MustCompile(`(?i)<a\s[^>]*?href\s*=\s*("([^"]*)"|'([^']*)')`)

// maskCode returns content with fenced code blocks and inline code spans
// overwritten by "x", byte for byte, so offsets found in the result are valid
// in content and links inside code are never matched.
func maskCode(content string) string {
	mask := func(s string) string {
		b := []byte(s)
		for i, c := range b {
			if c != '\r' {
				b[i] = 'x'
			}
		}
		return string(b)
	}
	lines := strings.Split(content, "\n")
	var fence string // the opening fence marker while inside a fenced block
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if fence != "" {
			// A closing fence is at least as long as the opener and has no info string.
			if strings.HasPrefix(trimmed, fence) && strings.Trim(trimmed, fence[:1]) == "" {
				fence = ""
			}
			lines[i] = mask(line)
			continue
		}
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fence = strings.Repeat(trimmed[:1], len(trimmed)-len(strings.TrimLeft(trimmed, trimmed[:1])))
			lines[i] = mask(line)
			continue
		}
		lines[i] = inlineCodeRe.ReplaceAllStringFunc(line, mask)
	}
	return strings.Join(lines, "\n")
}

// linkSpans returns the [start, end) byte ranges of the destinations of all
// non-image Markdown links and raw HTML anchors in content, ignoring code.
// markdownOnly leaves out the HTML anchors.
func linkSpans(content string, markdownOnly bool) [][2]int {
	text := maskCode(content)
	var spans [][2]int
	for _, m := range linkStartRe.FindAllStringSubmatchIndex(text, -1) {
		if text[m[2]:m[3]] == "!" {
			continue // skip image links
		}
		rest := text[m[1]:]
		dest := linkDestination(rest)
		if dest == "" {
			continue
		}
		start := m[1] + strings.Index(rest, dest)
		spans = append(spans, [2]int{start, start + len(dest)})
	}
	if markdownOnly {
		return spans
	}
	for _, m := range htmlHrefRe.FindAllStringSubmatchIndex(text, -1) {
		if m[4] >= 0 {
			spans = append(spans, [2]int{m[4], m[5]})
		} else {
			spans = append(spans, [2]int{m[6], m[7]})
		}
	}
	return spans
}

// pageLinks returns the destinations of all non-image Markdown links and raw
// HTML anchors in content, as written. Links inside fenced code blocks and
// inline code spans are ignored.
func pageLinks(content string) []string {
	var links []string
	for _, sp := range linkSpans(content, false) {
		links = append(links, content[sp[0]:sp[1]])
	}
	return links
}

// rewriteLinks replaces the destination of every non-image Markdown inline
// link in content with fn(dest) when fn reports a change. It returns the new
// content and the number of links changed. Code and raw HTML are left alone.
func rewriteLinks(content string, fn func(dest string) (string, bool)) (string, int) {
	spans := linkSpans(content, true)
	var b strings.Builder
	last, n := 0, 0
	for _, sp := range spans {
		if repl, ok := fn(content[sp[0]:sp[1]]); ok {
			b.WriteString(content[last:sp[0]])
			b.WriteString(repl)
			last = sp[1]
			n++
		}
	}
	b.WriteString(content[last:])
	return b.String(), n
}

// linkDestination parses a link destination from s, which starts right after
// the "(" of an inline link. It accepts "<...>" destinations and bare ones
// with balanced parentheses, and stops at whitespace so an optional title
// ("...") is left out.
func linkDestination(s string) string {
	s = strings.TrimLeft(s, " \t")
	if strings.HasPrefix(s, "<") {
		if end := strings.IndexAny(s, ">\n"); end > 0 && s[end] == '>' {
			return s[1:end]
		}
		return ""
	}
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ' ', '\t', '\n':
			return s[:i]
		case '(':
			depth++
		case ')':
			if depth == 0 {
				return s[:i]
			}
			depth--
		}
	}
	return s
}

// renameDest returns dest rewritten for a rename of the page or section
// oldName to newName, and whether it changed. Only site-absolute links are
// touched. A page ("blog/first") matches "/blog/first" and "/blog/first.html";
// a section matches "/blog" and everything below it. A fragment or query is
// kept.
func renameDest(dest, oldName, newName string, isSection bool) (string, bool) {
	p, suffix := dest, ""
	if i := strings.IndexAny(dest, "#?"); i >= 0 {
		p, suffix = dest[:i], dest[i:]
	}
	if !strings.HasPrefix(p, "/") || strings.HasPrefix(p, "//") {
		return dest, false
	}
	if u, err := url.PathUnescape(p); err == nil {
		p = u
	}
	old, repl := "/"+oldName, "/"+escapePath(newName)
	switch {
	case isSection && (p == old || strings.HasPrefix(p, old+"/")):
		return repl + escapePath(strings.TrimPrefix(p, old)) + suffix, true
	case !isSection && p == old:
		return repl + suffix, true
	case !isSection && p == old+".html":
		return repl + ".html" + suffix, true
	}
	return dest, false
}

// escapePath percent-encodes each "/"-separated segment of p.
func escapePath(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		segs[i] = url.PathEscape(s)
	}
	return strings.Join(segs, "/")
}

// rewriteSiteLinks applies fn to the links of every Markdown file under
// pages/, drafts included. It returns how many links and files changed and
// the files that could not be read or written.
func rewriteSiteLinks(pagesDir string, fn func(dest string) (string, bool)) (links, files int, failed []string) {
	_ = filepath.WalkDir(pagesDir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil //nolint:nilerr // unreadable entries are skipped; the rename itself already succeeded
		}
		rel, _ := filepath.Rel(pagesDir, p)
		rel = path.Clean(filepath.ToSlash(rel))
		content, err := os.ReadFile(p) //nolint:gosec // p comes from walking the site's pages dir
		if err != nil {
			failed = append(failed, rel)
			return nil
		}
		out, n := rewriteLinks(string(content), fn)
		if n == 0 {
			return nil
		}
		if err := os.WriteFile(p, []byte(out), 0644); err != nil { //nolint:gosec // keeps the mode of the existing file
			failed = append(failed, rel)
			return nil
		}
		links += n
		files++
		return nil
	})
	return links, files, failed
}
