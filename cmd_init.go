package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ChristianKreuzberger/press/internal/frontmatter"
	"github.com/ChristianKreuzberger/press/internal/themes"
)

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	themeName := fs.String("theme", themes.Default().Name,
		"theme to use (available: "+strings.Join(themes.Names(), ", ")+")")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: press init [dir] [flags]\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	theme, ok := themes.Get(*themeName)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown theme %q; available themes: %s\n", *themeName, strings.Join(themes.Names(), ", "))
		os.Exit(1)
	}

	siteDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting working directory: %v\n", err)
		os.Exit(1)
	}
	if fs.NArg() > 0 {
		siteDir = fs.Arg(0)
	}

	// Create pages directory
	pagesDir := filepath.Join(siteDir, "pages")
	if _, err := os.Stat(pagesDir); os.IsNotExist(err) {
		if err := os.MkdirAll(pagesDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "error creating pages directory: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("created pages/")
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "error checking pages directory: %v\n", err)
		os.Exit(1)
	}

	// Create pages/index.md (do not overwrite if it already exists)
	indexPath := filepath.Join(pagesDir, "index.md")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		indexContent := append(frontmatter.Generate("Home", time.Now()), []byte("# Home\n\nWelcome to my site.\n")...)
		if err := os.WriteFile(indexPath, indexContent, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error creating pages/index.md: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("created pages/index.md")
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "error checking pages/index.md: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Println("pages/index.md already exists, skipping")
	}

	// Create template.html (do not overwrite if it already exists)
	tmplPath := filepath.Join(siteDir, "template.html")
	if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
		if err := os.WriteFile(tmplPath, []byte(theme.Template), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error creating template.html: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("created template.html (theme: %s)\n", theme.Name)
	} else {
		fmt.Println("template.html already exists, skipping")
	}

	fmt.Println("site initialised")
}
