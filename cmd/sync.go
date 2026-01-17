package cmd

import (
	"flag"
	"fmt"
	"os"

	"neocities/api"
	"neocities/config"
	"neocities/site"
)

// RunSync handles the sync command.
func RunSync(args []string) {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	deleteRemoved := fs.Bool("delete", false, "Delete remote files not present locally")
	dryRun := fs.Bool("dry-run", false, "Show what would be synced without making changes")
	dir := fs.String("dir", ".", "Directory to sync from")
	fs.Usage = func() {
		fmt.Println(`Usage: neocities sync [options]

Synchronize local site to Neocities.

Compares local files with remote using SHA1 hashes and uploads
only changed or new files.

Options:
  -dir string     Directory to sync from (default ".")
  -delete         Delete remote files not present locally
  -dry-run        Show what would be synced without making changes

Examples:
  neocities sync
  neocities sync -delete
  neocities sync -dir ./public
  neocities sync -dry-run`)
	}

	fs.Parse(args)

	apiKey := config.GetAPIKey()
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: No API key found. Set NEOCITIES_API_KEY or run 'neocities key'")
		os.Exit(1)
	}

	// Load site config for ignore patterns
	siteCfg, err := site.LoadSiteConfig(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load site config: %v\n", err)
		siteCfg = &site.SiteConfig{Ignore: site.DefaultIgnore}
	}

	// Get local files
	fmt.Println("Scanning local files...")
	localFiles, err := site.ListLocalFiles(*dir, siteCfg.Ignore)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning local files: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d local files\n", len(localFiles))

	// Get remote files
	fmt.Println("Fetching remote file list...")
	client := api.NewClient(apiKey)
	remoteFiles, err := client.List("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching remote files: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d remote files\n", len(remoteFiles))

	// Compute sync plan
	plan := site.ComputeSyncPlan(localFiles, remoteFiles, *deleteRemoved)

	if len(plan.Upload) == 0 && len(plan.Delete) == 0 {
		fmt.Println("\nSite is up to date. Nothing to sync.")
		return
	}

	// Show plan
	fmt.Println()
	if len(plan.Upload) > 0 {
		fmt.Printf("Files to upload (%d):\n", len(plan.Upload))
		for _, f := range plan.Upload {
			fmt.Printf("  + %s\n", f.Path)
		}
	}

	if len(plan.Delete) > 0 {
		fmt.Printf("\nFiles to delete (%d):\n", len(plan.Delete))
		for _, f := range plan.Delete {
			fmt.Printf("  - %s\n", f)
		}
	}

	if *dryRun {
		fmt.Println("\n(dry-run mode, no changes made)")
		return
	}

	// Execute sync
	fmt.Println()

	// Upload files
	if len(plan.Upload) > 0 {
		fmt.Println("Uploading files...")
		files := make(map[string]string)
		for _, f := range plan.Upload {
			files[f.Path] = f.FullPath
		}
		if err := client.Upload(files); err != nil {
			fmt.Fprintf(os.Stderr, "Error uploading: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Uploaded %d files\n", len(plan.Upload))
	}

	// Delete files
	if len(plan.Delete) > 0 {
		fmt.Println("Deleting remote files...")
		if err := client.Delete(plan.Delete); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Deleted %d files\n", len(plan.Delete))
	}

	fmt.Println("\nSync complete!")
}
