package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ChristianKreuzberger/press/internal/builder"
)

// collectFileStates walks dir and returns a map of absolute file path to
// modification time, skipping the excludeDir subtree entirely.
func collectFileStates(dir, excludeDir string) (map[string]time.Time, error) {
	states := make(map[string]time.Time)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path == excludeDir {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			states[path] = info.ModTime()
		}
		return nil
	})
	return states, err
}

// hasChanged reports whether the file state has changed between two snapshots.
// It returns true when a file is added, removed, or modified.
func hasChanged(prev, curr map[string]time.Time) bool {
	if len(prev) != len(curr) {
		return true
	}
	for path, prevMod := range prev {
		currMod, ok := curr[path]
		if !ok || !prevMod.Equal(currMod) {
			return true
		}
	}
	return false
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	portFlag := fs.Int("port", 8080, "port to serve on")
	outputFlag := fs.String("output", "dist", "output directory for generated HTML files")
	intervalFlag := fs.Duration("interval", time.Second, "polling interval for file changes")
	_ = fs.Parse(args)

	siteDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting working directory: %v\n", err)
		os.Exit(1)
	}

	outputDir := filepath.Join(siteDir, *outputFlag)

	// Initial build.
	fmt.Println("building site...")
	if err := builder.Build(siteDir, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("built site to %s\n", *outputFlag)

	// Start HTTP file server in the background.
	addr := fmt.Sprintf(":%d", *portFlag)
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(outputDir)))
	go func() {
		srv := &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		}
		if err := srv.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	}()
	fmt.Printf("serving at http://localhost:%d — watching for changes (Ctrl+C to stop)\n", *portFlag)

	// Capture initial file state.
	prev, err := collectFileStates(siteDir, outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file states: %v\n", err)
		os.Exit(1)
	}

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(*intervalFlag)
	defer ticker.Stop()

	for {
		select {
		case <-quit:
			fmt.Println("\nstopping server")
			return
		case <-ticker.C:
			curr, err := collectFileStates(siteDir, outputDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading file states: %v\n", err)
				continue
			}
			if hasChanged(prev, curr) {
				prev = curr
				fmt.Println("change detected — rebuilding...")
				if err := builder.Build(siteDir, outputDir); err != nil {
					fmt.Fprintf(os.Stderr, "rebuild failed: %v\n", err)
				} else {
					fmt.Println("rebuilt successfully")
				}
			}
		}
	}
}
