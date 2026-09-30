package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

// Exit codes: 0 success, 1 runtime error, 2 usage error.
const (
	exitOK      = 0
	exitRuntime = 1
	exitUsage   = 2
)

// newFlagSet returns a flag set that reports errors instead of exiting, so
// callers control the exit code.
func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	return fs
}

// parseArgs parses flags and positionals interspersed in any order. Go's
// flag package stops at the first positional, so we resume after each one.
// A literal "--" ends flag parsing; everything after it is positional.
func parseArgs(fs *flag.FlagSet, args []string) ([]string, error) {
	var positionals []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			return positionals, nil
		}
		// fs.Parse consumes a leading "--", so detect it from the raw input.
		if consumedDoubleDash(args, rest) {
			return append(positionals, rest...), nil
		}
		positionals = append(positionals, rest[0])
		args = rest[1:]
	}
}

// consumedDoubleDash reports whether flag parsing stopped because of "--"
// rather than because of a positional argument.
func consumedDoubleDash(args, rest []string) bool {
	return len(rest) < len(args) && args[len(args)-len(rest)-1] == "--"
}

// parseOrExit parses args and enforces the positional count. It exits 0 for
// -h/--help and 2 for any usage error.
func parseOrExit(fs *flag.FlagSet, args []string, min, max int, usage string) []string {
	pos, err := parseArgs(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(exitOK)
		}
		os.Exit(exitUsage)
	}
	if len(pos) < min || len(pos) > max {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", usage)
		os.Exit(exitUsage)
	}
	return pos
}
