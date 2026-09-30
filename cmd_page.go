package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ChristianKreuzberger/press/internal/frontmatter"
	"github.com/ChristianKreuzberger/press/internal/page"
)

func runPageList(args []string) {
	fs := newFlagSet("list page")
	parseOrExit(fs, args, 0, 0, "press list page")

	siteDir := mustGetwd()
	pages, err := page.List(siteDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing pages: %v\n", err)
		os.Exit(exitRuntime)
	}
	if len(pages) == 0 {
		fmt.Println("no pages found")
		return
	}
	for _, p := range pages {
		if p.Draft {
			fmt.Println(p.Name + " [draft]")
		} else {
			fmt.Println(p.Name)
		}
	}
}

func runPageCreate(args []string) {
	fs := newFlagSet("create page")
	fileFlag := fs.String("file", "", "markdown file to use as page content")
	pos := parseOrExit(fs, args, 1, 1, "press create page <name> [--file <file.md>]\n       name may include sections, e.g. blog/my-post or blog/2026/my-post")
	name := pos[0]

	var content []byte
	if *fileFlag != "" {
		var err error
		content, err = os.ReadFile(*fileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file %s: %v\n", *fileFlag, err)
			os.Exit(exitRuntime)
		}
	} else {
		content = append(frontmatter.Generate(name, time.Now()), []byte("# "+name+"\n\n")...)
	}

	siteDir := mustGetwd()
	if err := page.Create(siteDir, name, content); err != nil {
		fmt.Fprintf(os.Stderr, "error creating page: %v\n", err)
		os.Exit(exitRuntime)
	}
	fmt.Printf("created page %q\n", name)
}

func runPageDelete(args []string) {
	fs := newFlagSet("delete page")
	pos := parseOrExit(fs, args, 1, 1, "press delete page <name>")
	name := pos[0]

	siteDir := mustGetwd()
	if err := page.Delete(siteDir, name); err != nil {
		fmt.Fprintf(os.Stderr, "error deleting page: %v\n", err)
		os.Exit(exitRuntime)
	}
	fmt.Printf("deleted page %q\n", name)
}

func runPageUpdate(args []string) {
	fs := newFlagSet("update page")
	fileFlag := fs.String("file", "", "markdown file to use as updated page content")
	pos := parseOrExit(fs, args, 1, 1, "press update page <name> --file <file.md>")
	name := pos[0]

	if *fileFlag == "" {
		fmt.Fprintf(os.Stderr, "press update page requires --file\n")
		os.Exit(exitUsage)
	}

	content, err := os.ReadFile(*fileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file %s: %v\n", *fileFlag, err)
		os.Exit(exitRuntime)
	}

	siteDir := mustGetwd()
	if err := page.Update(siteDir, name, content); err != nil {
		fmt.Fprintf(os.Stderr, "error updating page: %v\n", err)
		os.Exit(exitRuntime)
	}
	fmt.Printf("updated page %q\n", name)
}
