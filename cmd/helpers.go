package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/config"
)

// mustClient loads the API key and builds a client. Exits with code 2 if not authenticated.
func mustClient() (*client.Client, context.Context) {
	key, err := config.RequireAPIKey()
	if err != nil {
		if _, ok := err.(*config.ErrNotAuthenticated); ok {
			fmt.Fprintln(os.Stderr, "Not authenticated. Run: saber auth login")
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	c := client.New(apiURL, key, cliVersion, verbose, os.Stderr)
	return c, context.Background()
}
