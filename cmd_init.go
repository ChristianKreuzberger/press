package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ChristianKreuzberger/press/internal/builder"
)

func runInit(args []string) {
	siteDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting working directory: %v\n", err)
		os.Exit(1)
	}
	if len(args) > 0 {
		siteDir = args[0]
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

	// Create template.html (do not overwrite if it already exists)
	tmplPath := filepath.Join(siteDir, "template.html")
	if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
		if err := os.WriteFile(tmplPath, []byte(builder.DefaultTemplate), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error creating template.html: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("created template.html")
	} else {
		fmt.Println("template.html already exists, skipping")
	}

	fmt.Println("site initialised")
}
