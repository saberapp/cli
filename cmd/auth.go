package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}
	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthStatusCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var keyFlag string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Save your Saber API key",
		Long: `Store a Saber API key for CLI use.

Get your key at: https://ai.saber.app → Settings → API Keys`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var key string

			if keyFlag != "" {
				key = strings.TrimSpace(keyFlag)
			} else if !term.IsTerminal(int(os.Stdin.Fd())) {
				return fmt.Errorf("not a TTY — use --key sk_live_... to provide the API key non-interactively")
			} else {
				fmt.Fprint(os.Stderr, "Enter your Saber API key (sk_live_...): ")
				raw, err := term.ReadPassword(int(os.Stdin.Fd()))
				fmt.Fprintln(os.Stderr)
				if err != nil {
					return fmt.Errorf("read API key: %w", err)
				}
				key = strings.TrimSpace(string(raw))
			}

			if err := config.ValidateKeyFormat(key); err != nil {
				return err
			}

			// Validate key against API
			c := client.New(apiURL, key, cliVersion, false, os.Stderr)
			if _, err := c.GetCredits(context.Background(), nil); err != nil {
				if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 401 {
					return fmt.Errorf("invalid API key — authentication failed")
				}
				return fmt.Errorf("validate key: %w", err)
			}

			if err := config.SaveCredentials(key); err != nil {
				return fmt.Errorf("save credentials: %w", err)
			}

			if !quiet {
				fmt.Fprintf(os.Stdout, "Authenticated successfully  %s\n", client.MaskKey(key))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&keyFlag, "key", "", "API key (non-interactive)")
	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.DeleteCredentials(); err != nil {
				return err
			}
			if !quiet {
				fmt.Fprintln(os.Stdout, "Logged out.")
			}
			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			key, err := config.APIKey()
			if err != nil {
				return err
			}
			if key == "" {
				fmt.Fprintln(os.Stdout, "Not authenticated. Run: saber auth login")
				return &config.ErrNotAuthenticated{}
			}
			if jsonOutput {
				fmt.Fprintf(os.Stdout, `{"authenticated":true,"apiKey":%q}`+"\n", key)
				return nil
			}
			fmt.Fprintf(os.Stdout, "Authenticated  %s\n", client.MaskKey(key))
			return nil
		},
	}
}
