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
	minifyFlag := fs.Bool("minify", false, "minify HTML output to reduce file size")
	_ = fs.Parse(args)

	siteDir := mustGetwd()
	outputDir := filepath.Join(siteDir, *outputFlag)

	opts := builder.Options{Minify: *minifyFlag}
	stats, err := builder.BuildWithOptions(siteDir, outputDir, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	if *minifyFlag && stats.InputSize > 0 {
		saved := stats.InputSize - stats.OutputSize
		fmt.Printf("built %d page(s) to %s (%s, saved %s by minification)\n",
			stats.Pages, *outputFlag, formatBytes(stats.OutputSize), formatBytes(saved))
	} else {
		fmt.Printf("built %d page(s) to %s (%s)\n",
			stats.Pages, *outputFlag, formatBytes(stats.OutputSize))
	}
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for q := n / unit; q >= unit; q /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
