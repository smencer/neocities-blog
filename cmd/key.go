package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"neocities/api"
	"neocities/config"

	"golang.org/x/term"
)

// RunKey handles the key command.
func RunKey(args []string) {
	fs := flag.NewFlagSet("key", flag.ExitOnError)
	show := fs.Bool("show", false, "Show current API key")
	fs.Usage = func() {
		fmt.Println(`Usage: neocities key [options]

Retrieve and save your Neocities API key.

If no API key is configured, prompts for username and password
to retrieve your API key from Neocities.

Options:
  -show    Show current API key

Examples:
  neocities key           Retrieve and save API key
  neocities key -show     Show current API key`)
	}

	fs.Parse(args)

	if *show {
		apiKey := config.GetAPIKey()
		if apiKey == "" {
			fmt.Println("No API key configured.")
			fmt.Println("Run 'neocities key' to retrieve your API key.")
		} else {
			fmt.Printf("API Key: %s\n", apiKey)
		}
		return
	}

	// Check if we already have an API key
	existingKey := config.GetAPIKey()
	if existingKey != "" {
		fmt.Println("API key already configured.")
		fmt.Print("Retrieve new key? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return
		}
	}

	// Get username
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	// Get password (hidden input)
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
		os.Exit(1)
	}
	password := string(passwordBytes)

	// Retrieve API key
	fmt.Println("Retrieving API key...")
	client := api.NewClientWithCredentials(username, password)
	apiKey, err := client.Key()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Save API key
	if err := config.SaveAPIKey(apiKey); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving API key: %v\n", err)
		fmt.Printf("Your API key is: %s\n", apiKey)
		fmt.Println("You can set it manually with: export NEOCITIES_API_KEY=<key>")
		os.Exit(1)
	}

	fmt.Println("API key saved to ~/.config/neocities/config.json")
	fmt.Println("You can now use neocities commands without re-authenticating.")
}
