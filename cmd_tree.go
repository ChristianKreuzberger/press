package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/ChristianKreuzberger/press/internal/page"
	"github.com/ChristianKreuzberger/press/internal/section"
)

func runTree(_ []string) {
	siteDir := mustGetwd()

	pages, err := page.List(siteDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing pages: %v\n", err)
		os.Exit(1)
	}

	sections, err := section.List(siteDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing sections: %v\n", err)
		os.Exit(1)
	}

	if len(pages) == 0 && len(sections) == 0 {
		fmt.Println("no pages or sections found")
		return
	}

	type item struct {
		name      string
		isSection bool
	}

	items := make([]item, 0, len(pages)+len(sections))
	for _, p := range pages {
		items = append(items, item{name: p.Name})
	}
	for _, s := range sections {
		items = append(items, item{name: s.Name, isSection: true})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].name < items[j].name
	})

	fmt.Println("pages/")
	for i, it := range items {
		isLast := i == len(items)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		if it.isSection {
			fmt.Printf("%s%s/\n", connector, it.name)

			sectionPages, err := section.ListPages(siteDir, it.name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error listing pages in section %q: %v\n", it.name, err)
				os.Exit(1)
			}

			childIndent := "│   "
			if isLast {
				childIndent = "    "
			}
			for j, sp := range sectionPages {
				childConnector := "├── "
				if j == len(sectionPages)-1 {
					childConnector = "└── "
				}
				fmt.Printf("%s%s%s\n", childIndent, childConnector, sp.Name)
			}
		} else {
			fmt.Printf("%s%s\n", connector, it.name)
		}
	}
}
