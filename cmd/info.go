package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"neocities/api"
	"neocities/config"
)

// RunInfo handles the info command.
func RunInfo(args []string) {
	fs := flag.NewFlagSet("info", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println(`Usage: neocities info [sitename]

Get information about a Neocities site.

If no sitename is provided, shows info for your authenticated site.
You can view any public site's info without authentication.

Examples:
  neocities info          Show your site's info
  neocities info kyledrake    Show another site's info`)
	}

	fs.Parse(args)

	sitename := ""
	if fs.NArg() > 0 {
		sitename = fs.Arg(0)
	}

	var client *api.Client

	// If no sitename provided, we need auth
	if sitename == "" {
		apiKey := config.GetAPIKey()
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "Error: No API key found. Set NEOCITIES_API_KEY or run 'neocities key'")
			fmt.Fprintln(os.Stderr, "       Or provide a sitename to view public info: neocities info <sitename>")
			os.Exit(1)
		}
		client = api.NewClient(apiKey)
	} else {
		// For viewing other sites, we don't need auth
		client = api.NewClient("")
	}

	info, err := client.Info(sitename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Site:       %s\n", info.Sitename)
	fmt.Printf("URL:        https://%s.neocities.org\n", info.Sitename)
	if info.Domain != "" {
		fmt.Printf("Domain:     %s\n", info.Domain)
	}
	fmt.Printf("Hits:       %d\n", info.Hits)
	fmt.Printf("Created:    %s\n", info.CreatedAt)
	fmt.Printf("Updated:    %s\n", info.UpdatedAt)
	if len(info.Tags) > 0 {
		fmt.Printf("Tags:       %s\n", strings.Join(info.Tags, ", "))
	}
}
