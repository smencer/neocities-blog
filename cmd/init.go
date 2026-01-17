package cmd

import (
	"flag"
	"fmt"
	"os"

	"neocities/site"
)

// RunInit handles the init command.
func RunInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println(`Usage: neocities init <directory>

Scaffold a new Neocities site with a standard directory structure.

Creates:
  <directory>/
  ├── index.html
  ├── css/style.css
  ├── js/main.js
  ├── images/
  └── .neocities.json`)
	}

	fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	dir := fs.Arg(0)

	if err := site.Init(dir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created new site scaffold in %s\n", dir)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", dir)
	fmt.Println("  neocities sync")
}
