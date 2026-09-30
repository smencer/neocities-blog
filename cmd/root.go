package cmd

import (
	"fmt"
	"os"
)

const version = "0.1.0"

// Usage prints the help message.
func Usage() {
	fmt.Println(`neocities - Neocities site management utility

Usage:
  neocities <command> [options] [arguments]

Commands:
  init <dir>           Scaffold a new site with standard structure
  upload <files...>    Upload files to your Neocities site
  delete <files...>    Delete files from your Neocities site
  list [path]          List files on your Neocities site
  info [sitename]      Get information about a site
  sync                 Synchronize local site to Neocities
  key                  Retrieve and save your API key
  blog <subcommand>    Manage blog posts (init, build, new)
  version              Show version information
  help                 Show this help message

Authentication:
  Set NEOCITIES_API_KEY environment variable, or run 'neocities key'
  to save your API key to ~/.config/neocities/config.json

Examples:
  neocities init mysite           Create a new site scaffold
  neocities upload index.html     Upload a single file
  neocities sync                  Sync current directory to Neocities
  neocities info                  Show your site's information
  neocities info someguy          Show another user's site info`)
}

// Run executes the appropriate command based on arguments.
func Run() {
	if len(os.Args) < 2 {
		Usage()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "init":
		RunInit(os.Args[2:])
	case "upload":
		RunUpload(os.Args[2:])
	case "delete":
		RunDelete(os.Args[2:])
	case "list":
		RunList(os.Args[2:])
	case "info":
		RunInfo(os.Args[2:])
	case "sync":
		RunSync(os.Args[2:])
	case "key":
		RunKey(os.Args[2:])
	case "blog":
		RunBlog(os.Args[2:])
	case "version":
		fmt.Printf("neocities version %s\n", version)
	case "help", "-h", "--help":
		Usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		Usage()
		os.Exit(1)
	}
}
