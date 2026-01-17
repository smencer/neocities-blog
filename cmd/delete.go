package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"neocities/api"
	"neocities/config"
)

// RunDelete handles the delete command.
func RunDelete(args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	force := fs.Bool("f", false, "Skip confirmation prompt")
	fs.Usage = func() {
		fmt.Println(`Usage: neocities delete [options] <files...>

Delete files from your Neocities site.

WARNING: There is no way to undo a delete!

Options:
  -f    Skip confirmation prompt

Examples:
  neocities delete old-page.html
  neocities delete -f temp.html debug.js`)
	}

	fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	filenames := fs.Args()

	// Check for protected files
	for _, f := range filenames {
		if f == "index.html" {
			fmt.Fprintln(os.Stderr, "Error: Cannot delete index.html")
			os.Exit(1)
		}
	}

	// Confirm unless -f flag is set
	if !*force {
		fmt.Println("The following files will be PERMANENTLY deleted:")
		for _, f := range filenames {
			fmt.Printf("  %s\n", f)
		}
		fmt.Print("\nAre you sure? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			os.Exit(0)
		}
	}

	apiKey := config.GetAPIKey()
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: No API key found. Set NEOCITIES_API_KEY or run 'neocities key'")
		os.Exit(1)
	}

	client := api.NewClient(apiKey)

	fmt.Printf("Deleting %d file(s)...\n", len(filenames))
	if err := client.Delete(filenames); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Delete complete!")
}
