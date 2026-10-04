package main

import (
	"fmt"
	"os"
	"strings"
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
		fmt.Fprintf(os.Stderr, "error renaming page %q: %v\n", oldName, err)
		os.Exit(exitRuntime)
	}
	fmt.Printf("renamed page %q to %q\n", oldName, newName)
	updateLinks(siteDir, oldName, newName, false)
}

func runSectionRename(args []string) {
	fs := newFlagSet("rename section")
	pos := parseOrExit(fs, args, 2, 2, "press rename section <old-name> <new-name>")
	oldName, newName := pos[0], pos[1]

	siteDir := mustGetwd()
	if err := section.Rename(siteDir, oldName, newName, time.Now()); err != nil {
		fmt.Fprintf(os.Stderr, "error renaming section %q: %v\n", oldName, err)
		os.Exit(exitRuntime)
	}
	fmt.Printf("renamed section %q to %q\n", oldName, newName)
	updateLinks(siteDir, oldName, newName, true)
}

// updateLinks points Markdown links to the old name at the new one after a
// successful rename. If some file cannot be rewritten the rename stays done,
// but the command exits non-zero so the leftover links are not missed.
func updateLinks(siteDir, oldName, newName string, isSection bool) {
	links, files, failed := rewriteSiteLinks(page.PagesDir(siteDir), func(dest string) (string, bool) {
		return renameDest(dest, oldName, newName, isSection)
	})
	if links > 0 {
		fmt.Printf("updated %d link(s) in %d page(s)\n", links, files)
	}
	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, "warning: the rename succeeded, but links could not be updated in: %s\n", strings.Join(failed, ", "))
		os.Exit(exitRuntime)
	}
	fmt.Println("run `press check` to verify the remaining links")
}
