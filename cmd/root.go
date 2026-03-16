package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Global flags and state shared across all commands.
var (
	jsonOutput bool
	quiet      bool
	verbose    bool
	yes        bool
	apiURL     string
	cliVersion string
)

func NewRootCmd(version, commit, date string) *cobra.Command {
	cliVersion = version
	root := &cobra.Command{
		Use:   "saber",
		Short: "Saber CLI — run signal research and manage lists from your terminal",
		Long: `Saber CLI lets you run company signal research, manage company and contact
lists, check credits, and more — all from your terminal.

Get an API key at: https://ai.saber.app → Settings → API Keys`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	root.PersistentFlags().BoolVarP(&quiet, "quiet", "Q", false, "Suppress non-error output")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose HTTP logging (stderr)")
	root.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "Skip credit confirmation prompt")
	root.PersistentFlags().StringVar(&apiURL, "api-url", "https://api.saber.app", "Base API URL")

	root.SetErr(os.Stderr)
	root.SetOut(os.Stdout)

	root.AddCommand(newSignalCmd())
	root.AddCommand(newListCmd())
	root.AddCommand(newAuthCmd())
	root.AddCommand(newCreditsCmd())
	root.AddCommand(newConnectorsCmd())
	root.AddCommand(newVersionCmd(version, commit, date))
	root.AddCommand(newSubscriptionCmd())
	root.AddCommand(newUpdateCmd())
	root.AddCommand(newOrgCmd())
	root.AddCommand(newInitClaudeCmd())
	root.SetHelpCommand(newHelpCmd(root))

	return root
}
