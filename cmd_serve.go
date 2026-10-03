package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ChristianKreuzberger/press/internal/builder"
)

// collectFileStates returns the modification time of every file that feeds a
// build: pages/, template.html and the static dir (a name relative to siteDir).
// Everything else (.git, node_modules, editor swap files, the output dir) is
// ignored so it cannot trigger rebuilds. Sources that are missing, or vanish
// while being read (editors replace files on save), are skipped.
func collectFileStates(siteDir, staticDir string) (map[string]time.Time, error) {
	states := make(map[string]time.Time)

	for _, root := range []string{filepath.Join(siteDir, "pages"), filepath.Join(siteDir, staticDir)} {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			info, err := d.Info()
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			if err != nil {
				return err
			}
			states[path] = info.ModTime()
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	tmplPath := filepath.Join(siteDir, "template.html")
	if info, err := os.Stat(tmplPath); err == nil {
		states[tmplPath] = info.ModTime()
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	return states, nil
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

// defaultServeHost keeps the dev server reachable only from this machine.
const defaultServeHost = "127.0.0.1"

// listenAddr builds the address to bind. An empty host falls back to the
// default so we never produce ":port", which would bind all interfaces.
func listenAddr(host string, port int) string {
	host = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(host), "["), "]")
	if host == "" {
		host = defaultServeHost
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}

// isLoopbackHost reports whether host only accepts connections from this machine.
func isLoopbackHost(host string) bool {
	host = strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(host), "["), "]")
	if host == "" || host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	portFlag := fs.Int("port", 8080, "port to serve on")
	outputFlag := fs.String("output", "dist", "output directory for generated HTML files")
	intervalFlag := fs.Duration("interval", time.Second, "polling interval for file changes")
	draftsFlag := fs.Bool("drafts", false, "include draft pages in the build")
	staticFlag := fs.String("static", "static", "name of the static assets directory to copy into the output")
	hostFlag := fs.String("host", defaultServeHost, "host/address to listen on (use 0.0.0.0 to expose to the network)")
	_ = fs.Parse(args)

	siteDir := mustGetwd()

	outputDir := filepath.Join(siteDir, *outputFlag)

	// Initial build.
	fmt.Println("building site...")
	if _, err := builder.Build(siteDir, outputDir, *draftsFlag, *staticFlag); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	warnSkippedPages(siteDir)
	fmt.Printf("built site to %s\n", *outputFlag)

	// Start HTTP file server in the background.
	addr := listenAddr(*hostFlag, *portFlag)
	if !isLoopbackHost(*hostFlag) {
		fmt.Fprintf(os.Stderr, "warning: listening on %s exposes the site to the network\n", addr)
	}
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
	fmt.Printf("serving at http://%s — watching for changes (Ctrl+C to stop)\n", addr)

	// Capture initial file state.
	prev, err := collectFileStates(siteDir, *staticFlag)
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
			curr, err := collectFileStates(siteDir, *staticFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading file states: %v\n", err)
				continue
			}
			if hasChanged(prev, curr) {
				prev = curr
				fmt.Println("change detected — rebuilding...")
				if _, err := builder.Build(siteDir, outputDir, *draftsFlag, *staticFlag); err != nil {
					fmt.Fprintf(os.Stderr, "rebuild failed: %v\n", err)
				} else {
					warnSkippedPages(siteDir)
					fmt.Println("rebuilt successfully")
				}
			}
		}
	}
}
