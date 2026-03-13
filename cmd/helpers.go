package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/config"
	"golang.org/x/term"
)

// confirmCreditAction checks the credit balance and prompts the user to confirm before
// proceeding with a credit-consuming action. It is skipped when --yes or --quiet is set.
func confirmCreditAction(c *client.Client, ctx context.Context) error {
	if yes || quiet {
		return nil
	}
	bal, err := c.GetCredits(ctx, nil)
	if err != nil {
		return fmt.Errorf("check credits: %w", err)
	}
	msg := fmt.Sprintf("This will consume credits (balance: %s). Proceed? [y/N] ",
		formatInt(bal.RemainingCredits))
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("%suse --yes to confirm non-interactively", msg)
	}
	fmt.Fprint(os.Stderr, msg)
	var answer string
	if _, err := fmt.Fscan(os.Stdin, &answer); err != nil && err.Error() != "EOF" {
		return fmt.Errorf("read input: %w", err)
	}
	if strings.ToLower(strings.TrimSpace(answer)) != "y" {
		return fmt.Errorf("aborted")
	}
	return nil
}

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
