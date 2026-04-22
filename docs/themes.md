# Themes

A **theme** in press is a single `template.html` file that controls how your site looks. When you run `press build`, press reads `template.html` from the root of your site and applies it to every page.

---

## Choosing a theme at init

Pass `--theme <name>` to `press init` to start with a specific built-in theme:

```bash
press init --theme light
press init --theme terminal
press init --theme dark   # default
```

Available built-in themes:

| Name       | Description |
|------------|-------------|
| `dark`     | Dark developer theme inspired by GitHub dark mode |
| `light`    | Clean editorial theme with serif headings and a light background |
| `terminal` | Retro green-on-black terminal aesthetic with monospace fonts |

If you omit `--theme`, press uses `dark`.

---

## Switching themes

To switch themes after initializing, replace `template.html` with one of the built-in templates. The easiest way is to create a new throwaway site and copy the generated file:

```bash
press init /tmp/theme-preview --theme light
cp /tmp/theme-preview/template.html ./template.html
```

---

## How templates work

`template.html` is a standard [Go `html/template`](https://pkg.go.dev/html/template) file. press renders every Markdown page through this template and passes it a `TemplateData` value with the following fields:

| Field              | Type            | Description |
|--------------------|-----------------|-------------|
| `.Title`           | `string`        | Page title extracted from the first `# heading` or the frontmatter `title` field |
| `.Content`         | `template.HTML` | Rendered HTML body of the page |
| `.Pages`           | `[]PageRef`     | Navigation entries; each has `.Title` (string) and `.URL` (string) |
| `.TableOfContents` | `[]TOCEntry`    | Section index entries (non-empty on section index pages only); each has `.Title` (string) and `.URL` (string) |

### Minimal example

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
</head>
<body>
    <nav>
        {{range .Pages}}<a href="{{.URL}}">{{.Title}}</a> {{end}}
    </nav>
    <main>
        {{.Content}}
        {{if .TableOfContents}}
        <ul>
            {{range .TableOfContents}}<li><a href="{{.URL}}">{{.Title}}</a></li>{{end}}
        </ul>
        {{end}}
    </main>
</body>
</html>
```

Save this as `template.html` in your site root and run `press build`.

---

## Creating your own theme

1. **Copy the minimal example** above into `template.html`.
2. **Add your own CSS.** You can inline styles in a `<style>` block, load a stylesheet from a CDN, or reference a local CSS file in your `dist/` folder (place static assets there manually or via a build step).
3. **Use the template variables** listed above to inject the page title, content, and navigation.
4. **Rebuild** with `press build` (or `press serve` for live reload) after each change.

### Tips

- **CSS variables** make it easy to define a palette once and reuse it everywhere. See the built-in themes in [`internal/themes/themes.go`](../internal/themes/themes.go) for examples.
- **Navigation order** is controlled by the `weight` frontmatter field on pages and sections — you don't need to touch the template to reorder links.
- **Section table of contents** — the `.TableOfContents` list is only populated on section index pages (`pages/<section>/index.md`). Guard it with `{{if .TableOfContents}}` as shown above.
- **External fonts** — load them from a CDN or self-host them in `dist/fonts/`.

### Sharing themes

A theme is just a single HTML file, so sharing is as simple as committing `template.html` to a repository or gist. Others can drop it into their site root and run `press build`.

---

## Fallback behaviour

If `template.html` does not exist when you run `press build`, press falls back to the built-in `dark` theme automatically. No configuration required.
