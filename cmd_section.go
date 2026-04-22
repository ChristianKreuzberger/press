package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ChristianKreuzberger/press/internal/section"
)

func runSection(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: press section <command>\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  list                    list all sections\n")
		fmt.Fprintf(os.Stderr, "  create <name>           create a new section\n")
		fmt.Fprintf(os.Stderr, "  delete <name>           delete a section and all its pages\n")
		fmt.Fprintf(os.Stderr, "  update <name>           update a section's index page\n")
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		runSectionList(args[1:])
	case "create":
		runSectionCreate(args[1:])
	case "delete":
		runSectionDelete(args[1:])
	case "update":
		runSectionUpdate(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown section command: %s\n", args[0])
		os.Exit(1)
	}
}

func runSectionList(_ []string) {
	siteDir, _ := os.Getwd()
	sections, err := section.List(siteDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing sections: %v\n", err)
		os.Exit(1)
	}
	if len(sections) == 0 {
		fmt.Println("no sections found")
		return
	}
	for _, s := range sections {
		fmt.Println(s.Name)
	}
}

func runSectionCreate(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: press section create <name> [--file <file.md>]\n")
		os.Exit(1)
	}
	name := args[0]

	fs := flag.NewFlagSet("section create", flag.ExitOnError)
	fileFlag := fs.String("file", "", "markdown file to use as the section index content")
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
	if err := section.Create(siteDir, name, content); err != nil {
		fmt.Fprintf(os.Stderr, "error creating section: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("created section %q\n", name)
}

func runSectionDelete(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: press section delete <name>\n")
		os.Exit(1)
	}
	name := args[0]
	siteDir, _ := os.Getwd()
	if err := section.Delete(siteDir, name); err != nil {
		fmt.Fprintf(os.Stderr, "error deleting section: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("deleted section %q\n", name)
}

func runSectionUpdate(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: press section update <name> --file <file.md>\n")
		os.Exit(1)
	}
	name := args[0]

	fs := flag.NewFlagSet("section update", flag.ExitOnError)
	fileFlag := fs.String("file", "", "markdown file to use as updated section index content")
	_ = fs.Parse(args[1:])

	if *fileFlag == "" {
		fmt.Fprintf(os.Stderr, "press section update requires --file\n")
		os.Exit(1)
	}

	content, err := os.ReadFile(*fileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file %s: %v\n", *fileFlag, err)
		os.Exit(1)
	}

	siteDir, _ := os.Getwd()
	if err := section.Update(siteDir, name, content); err != nil {
		fmt.Fprintf(os.Stderr, "error updating section: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("updated section %q\n", name)
}
