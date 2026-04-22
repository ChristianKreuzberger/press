package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ChristianKreuzberger/press/internal/frontmatter"
	"github.com/ChristianKreuzberger/press/internal/page"
)

func runPageList(_ []string) {
	siteDir, _ := os.Getwd()
	pages, err := page.List(siteDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing pages: %v\n", err)
		os.Exit(1)
	}
	if len(pages) == 0 {
		fmt.Println("no pages found")
		return
	}
	for _, p := range pages {
		fmt.Println(p.Name)
	}
}

func runPageCreate(args []string) {
	// Name is the first positional argument; remaining args may contain flags.
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: press create page <name> [--file <file.md>]\n       name may include sections, e.g. blog/my-post or blog/2026/my-post\n")
		os.Exit(1)
	}
	name := args[0]

	fs := flag.NewFlagSet("create page", flag.ExitOnError)
	fileFlag := fs.String("file", "", "markdown file to use as page content")
	_ = fs.Parse(args[1:])

	var content []byte
	if *fileFlag != "" {
		var err error
		content, err = os.ReadFile(*fileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading file %s: %v\n", *fileFlag, err)
			os.Exit(1)
		}
	} else {
		fm := frontmatter.Generate(name, time.Now())
		content = append(fm, []byte("# "+name+"\n\n")...)
	}

	siteDir, _ := os.Getwd()
	if err := page.Create(siteDir, name, content); err != nil {
		fmt.Fprintf(os.Stderr, "error creating page: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("created page %q\n", name)
}

func runPageDelete(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: press delete page <name>\n")
		os.Exit(1)
	}
	name := args[0]
	siteDir, _ := os.Getwd()
	if err := page.Delete(siteDir, name); err != nil {
		fmt.Fprintf(os.Stderr, "error deleting page: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("deleted page %q\n", name)
}

func runPageUpdate(args []string) {
	// Name is the first positional argument; remaining args may contain flags.
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: press update page <name> --file <file.md>\n")
		os.Exit(1)
	}
	name := args[0]

	fs := flag.NewFlagSet("update page", flag.ExitOnError)
	fileFlag := fs.String("file", "", "markdown file to use as updated page content")
	_ = fs.Parse(args[1:])

	if *fileFlag == "" {
		fmt.Fprintf(os.Stderr, "press update page requires --file\n")
		os.Exit(1)
	}

	content, err := os.ReadFile(*fileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file %s: %v\n", *fileFlag, err)
		os.Exit(1)
	}

	siteDir, _ := os.Getwd()
	if err := page.Update(siteDir, name, content); err != nil {
		fmt.Fprintf(os.Stderr, "error updating page: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("updated page %q\n", name)
}
