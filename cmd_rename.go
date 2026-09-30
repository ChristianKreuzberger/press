package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ChristianKreuzberger/press/internal/page"
	"github.com/ChristianKreuzberger/press/internal/section"
)

func runPageRename(args []string) {
	fs := newFlagSet("rename page")
	pos := parseOrExit(fs, args, 2, 2, "press rename page <old-name> <new-name>")
	oldName, newName := pos[0], pos[1]

	siteDir := mustGetwd()
	if err := page.Rename(siteDir, oldName, newName, time.Now()); err != nil {
		fmt.Fprintf(os.Stderr, "error renaming page: %v\n", err)
		os.Exit(exitRuntime)
	}
	fmt.Printf("renamed page %q to %q\n", oldName, newName)
}

func runSectionRename(args []string) {
	fs := newFlagSet("rename section")
	pos := parseOrExit(fs, args, 2, 2, "press rename section <old-name> <new-name>")
	oldName, newName := pos[0], pos[1]

	siteDir := mustGetwd()
	if err := section.Rename(siteDir, oldName, newName, time.Now()); err != nil {
		fmt.Fprintf(os.Stderr, "error renaming section: %v\n", err)
		os.Exit(exitRuntime)
	}
	fmt.Printf("renamed section %q to %q\n", oldName, newName)
}
