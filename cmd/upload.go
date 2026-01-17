package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"neocities/api"
	"neocities/config"
)

// RunUpload handles the upload command.
func RunUpload(args []string) {
	fs := flag.NewFlagSet("upload", flag.ExitOnError)
	remotePath := fs.String("path", "", "Remote path prefix for uploaded files")
	fs.Usage = func() {
		fmt.Println(`Usage: neocities upload [options] <files...>

Upload files to your Neocities site.

Options:
  -path string    Remote path prefix for uploaded files

Examples:
  neocities upload index.html
  neocities upload css/style.css js/main.js
  neocities upload -path blog/ post.html`)
	}

	fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	apiKey := config.GetAPIKey()
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: No API key found. Set NEOCITIES_API_KEY or run 'neocities key'")
		os.Exit(1)
	}

	client := api.NewClient(apiKey)
	files := make(map[string]string)

	for _, localPath := range fs.Args() {
		// Check if file exists
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: File not found: %s\n", localPath)
			os.Exit(1)
		}

		// Determine remote path
		remote := filepath.Base(localPath)
		if *remotePath != "" {
			remote = filepath.Join(*remotePath, remote)
		}

		files[remote] = localPath
	}

	fmt.Printf("Uploading %d file(s)...\n", len(files))
	for remote, local := range files {
		fmt.Printf("  %s -> %s\n", local, remote)
	}

	if err := client.Upload(files); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Upload complete!")
}
