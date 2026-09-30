package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ChristianKreuzberger/press/internal/builder"
)

func runBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	outputFlag := fs.String("output", "dist", "output directory for generated HTML files")
	draftsFlag := fs.Bool("drafts", false, "include draft pages in the build")
	verboseFlag := fs.Bool("verbose", false, "print each built page")
	staticFlag := fs.String("static", "static", "name of the static assets directory to copy into the output")
	_ = fs.Parse(args)

	siteDir := mustGetwd()

	outputDir := filepath.Join(siteDir, *outputFlag)

	start := time.Now()
	built, err := builder.Build(siteDir, outputDir, *draftsFlag, *staticFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	if *verboseFlag {
		for _, path := range built {
			rel, err := filepath.Rel(siteDir, path)
			if err != nil {
				rel = path
			}
			fmt.Printf("  %s\n", rel)
		}
	}
	fmt.Printf("✓ Built %d pages in %v → %s/\n", len(built), time.Since(start).Round(time.Millisecond), *outputFlag)
}
