package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ChristianKreuzberger/press/internal/page"
	"github.com/ChristianKreuzberger/press/internal/section"
)

func runPageRename(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: press rename page <old-name> <new-name>\n")
		os.Exit(1)
	}
	oldName, newName := args[0], args[1]
	siteDir := mustGetwd()
	if err := page.Rename(siteDir, oldName, newName, time.Now()); err != nil {
		fmt.Fprintf(os.Stderr, "error renaming page: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("renamed page %q to %q\n", oldName, newName)
}

func runSectionRename(args []string) {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: press rename section <old-name> <new-name>\n")
		os.Exit(1)
	}
	oldName, newName := args[0], args[1]
	siteDir := mustGetwd()
	if err := section.Rename(siteDir, oldName, newName, time.Now()); err != nil {
		fmt.Fprintf(os.Stderr, "error renaming section: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("renamed section %q to %q\n", oldName, newName)
}
