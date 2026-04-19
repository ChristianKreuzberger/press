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
		fmt.Fprintf(os.Stderr, "  init            initialise a new site (creates template.html and pages/)\n")
		fmt.Fprintf(os.Stderr, "  page <command>  manage pages (list, create, delete, update)\n")
		fmt.Fprintf(os.Stderr, "  build           build the static site into dist/\n")
		fmt.Fprintf(os.Stderr, "  serve           serve the site locally\n")
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
	case "page":
		runPage(args[1:])
	case "build":
		runBuild(args[1:])
	case "serve":
		fmt.Println("serve: not yet implemented")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		flag.Usage()
		os.Exit(1)
	}
}
