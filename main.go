package main

import (
	"flag"
	"fmt"
	"os"
)

var version = "dev"

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "press - simple static site generator for agents\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  press [flags] <command>\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  init                    initialise a new site (creates template.html and pages/)\n")
		fmt.Fprintf(os.Stderr, "  build                   build the static site into dist/\n")
		fmt.Fprintf(os.Stderr, "  serve                   serve the site locally\n")
		fmt.Fprintf(os.Stderr, "  create <page|section>   create a page or section\n")
		fmt.Fprintf(os.Stderr, "  list <page|section>     list pages or sections\n")
		fmt.Fprintf(os.Stderr, "  update <page|section>   update a page or section\n")
		fmt.Fprintf(os.Stderr, "  delete <page|section>   delete a page or section\n")
		fmt.Fprintf(os.Stderr, "  tree                    show a tree of all pages and sections\n")
	}
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	switch args[0] {
	case "init":
		runInit(args[1:])
	case "build":
		runBuild(args[1:])
	case "serve":
		runServe(args[1:])
	case "create":
		runCreate(args[1:])
	case "list":
		runList(args[1:])
	case "update":
		runUpdate(args[1:])
	case "delete":
		runDelete(args[1:])
	case "tree":
		runTree(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		flag.Usage()
		os.Exit(1)
	}
}

func nounArg(verb string, args []string) string {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: press %s <page|section> [args]\n", verb)
		os.Exit(1)
	}
	return args[0]
}

func runCreate(args []string) {
	switch nounArg("create", args) {
	case "page":
		runPageCreate(args[1:])
	case "section":
		runSectionCreate(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown noun: %s (expected: page, section)\n", args[0])
		os.Exit(1)
	}
}

func runList(args []string) {
	switch nounArg("list", args) {
	case "page":
		runPageList(args[1:])
	case "section":
		runSectionList(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown noun: %s (expected: page, section)\n", args[0])
		os.Exit(1)
	}
}

func runUpdate(args []string) {
	switch nounArg("update", args) {
	case "page":
		runPageUpdate(args[1:])
	case "section":
		runSectionUpdate(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown noun: %s (expected: page, section)\n", args[0])
		os.Exit(1)
	}
}

func runDelete(args []string) {
	switch nounArg("delete", args) {
	case "page":
		runPageDelete(args[1:])
	case "section":
		runSectionDelete(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown noun: %s (expected: page, section)\n", args[0])
		os.Exit(1)
	}
}
