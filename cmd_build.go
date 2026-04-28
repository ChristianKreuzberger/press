package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ChristianKreuzberger/press/internal/builder"
)

func runBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	outputFlag := fs.String("output", "dist", "output directory for generated HTML files")
	_ = fs.Parse(args)

	siteDir := mustGetwd()

	outputDir := filepath.Join(siteDir, *outputFlag)

	if err := builder.Build(siteDir, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("built site to %s\n", *outputFlag)
}
