package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ChristianKreuzberger/press/internal/page"
)

func runPage(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: press page <command>\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  list              list all pages\n")
		fmt.Fprintf(os.Stderr, "  create <name>     create a new page\n")
		fmt.Fprintf(os.Stderr, "  delete <name>     delete a page\n")
		fmt.Fprintf(os.Stderr, "  update <name>     update a page\n")
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		runPageList(args[1:])
	case "create":
		runPageCreate(args[1:])
	case "delete":
		runPageDelete(args[1:])
	case "update":
		runPageUpdate(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown page command: %s\n", args[0])
		os.Exit(1)
	}
}

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
		fmt.Fprintf(os.Stderr, "Usage: press page create <name> [--file <file.md>]\n")
		os.Exit(1)
	}
	name := args[0]

	fs := flag.NewFlagSet("page create", flag.ExitOnError)
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
		content = []byte("# " + name + "\n\n")
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
		fmt.Fprintf(os.Stderr, "Usage: press page delete <name>\n")
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
		fmt.Fprintf(os.Stderr, "Usage: press page update <name> --file <file.md>\n")
		os.Exit(1)
	}
	name := args[0]

	fs := flag.NewFlagSet("page update", flag.ExitOnError)
	fileFlag := fs.String("file", "", "markdown file to use as updated page content")
	_ = fs.Parse(args[1:])

	if *fileFlag == "" {
		fmt.Fprintf(os.Stderr, "press page update requires --file\n")
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
