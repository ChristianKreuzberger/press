package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ChristianKreuzberger/press/internal/frontmatter"
	"github.com/ChristianKreuzberger/press/internal/section"
)

func runSectionList(args []string) {
	fs := newFlagSet("list section")
	parseOrExit(fs, args, 0, 0, "press list section")

	siteDir := mustGetwd()
	sections, err := section.List(siteDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing sections: %v\n", err)
		os.Exit(exitRuntime)
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
	fs := newFlagSet("create section")
	fileFlag := fs.String("file", "", "markdown file to use as the section index content")
	pos := parseOrExit(fs, args, 1, 1, "press create section <name> [--file <file.md>]")
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
		content = append(frontmatter.GenerateSection(name, time.Now()), []byte("# "+name+"\n\n")...)
	}

	siteDir := mustGetwd()
	if err := section.Create(siteDir, name, content); err != nil {
		fmt.Fprintf(os.Stderr, "error creating section: %v\n", err)
		os.Exit(exitRuntime)
	}
	fmt.Printf("created section %q\n", name)
}

func runSectionDelete(args []string) {
	fs := newFlagSet("delete section")
	pos := parseOrExit(fs, args, 1, 1, "press delete section <name>")
	name := pos[0]

	siteDir := mustGetwd()
	if err := section.Delete(siteDir, name); err != nil {
		fmt.Fprintf(os.Stderr, "error deleting section: %v\n", err)
		os.Exit(exitRuntime)
	}
	fmt.Printf("deleted section %q\n", name)
}

func runSectionUpdate(args []string) {
	fs := newFlagSet("update section")
	fileFlag := fs.String("file", "", "markdown file to use as updated section index content")
	pos := parseOrExit(fs, args, 1, 1, "press update section <name> --file <file.md>")
	name := pos[0]

	if *fileFlag == "" {
		fmt.Fprintf(os.Stderr, "press update section requires --file\n")
		os.Exit(exitUsage)
	}

	content, err := os.ReadFile(*fileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file %s: %v\n", *fileFlag, err)
		os.Exit(exitRuntime)
	}

	siteDir := mustGetwd()
	if err := section.Update(siteDir, name, content); err != nil {
		fmt.Fprintf(os.Stderr, "error updating section: %v\n", err)
		os.Exit(exitRuntime)
	}
	fmt.Printf("updated section %q\n", name)
}
