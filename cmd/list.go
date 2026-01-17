package cmd

import (
	"flag"
	"fmt"
	"os"

	"neocities/api"
	"neocities/config"
)

// RunList handles the list command.
func RunList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	long := fs.Bool("l", false, "Show detailed listing with size and date")
	fs.Usage = func() {
		fmt.Println(`Usage: neocities list [options] [path]

List files on your Neocities site.

Options:
  -l    Show detailed listing with size and date

Examples:
  neocities list
  neocities list -l
  neocities list images/`)
	}

	fs.Parse(args)

	apiKey := config.GetAPIKey()
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: No API key found. Set NEOCITIES_API_KEY or run 'neocities key'")
		os.Exit(1)
	}

	path := ""
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}

	client := api.NewClient(apiKey)
	files, err := client.List(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No files found.")
		return
	}

	for _, f := range files {
		if *long {
			typeChar := "-"
			if f.IsDir {
				typeChar = "d"
			}
			fmt.Printf("%s %10d  %s  %s\n", typeChar, f.Size, f.UpdatedAt, f.Path)
		} else {
			if f.IsDir {
				fmt.Printf("%s/\n", f.Path)
			} else {
				fmt.Println(f.Path)
			}
		}
	}
}
