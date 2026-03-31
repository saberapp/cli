package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/saberapp/cli/internal/update"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Global flags and state shared across all commands.
var (
	jsonOutput bool
	quiet      bool
	verbose    bool
	yes        bool
	apiURL     string
	cliVersion string

	// updateCh receives at most one message from the background update
	// checker. It is set in PersistentPreRun and drained in PersistentPostRun.
	updateCh <-chan string
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
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if shouldSkipUpdateCheck(cmd) {
				return
			}
			updateCh = update.RunBackgroundCheck(cliVersion, 24*time.Hour)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if updateCh == nil {
				return
			}
			// Non-blocking read: print notice only if the background
			// check finished before the command completed.
			select {
			case msg, ok := <-updateCh:
				if ok && msg != "" {
					fmt.Fprintln(os.Stderr, msg)
				}
			default:
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			printCheatSheet(os.Stdout)
		},
	}

	root.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	root.PersistentFlags().BoolVarP(&quiet, "quiet", "Q", false, "Suppress non-error output")
	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose HTTP logging (stderr)")
	root.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "Skip credit confirmation prompt")
	root.PersistentFlags().StringVar(&apiURL, "api-url", "https://api.saber.app", "Base API URL")

	root.SetErr(os.Stderr)
	root.SetOut(os.Stdout)

	root.AddCommand(newSignalCmd())
	root.AddCommand(newTemplateCmd())
	root.AddCommand(newSummaryCmd())
	root.AddCommand(newMarketSignalCmd())
	root.AddCommand(newCompanyCmd())
	root.AddCommand(newContactCmd())
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

// shouldSkipUpdateCheck returns true when the passive background version check
// should not run (dev builds, machine-readable output, explicit opt-out, the
// update command itself, or non-TTY stderr).
func shouldSkipUpdateCheck(cmd *cobra.Command) bool {
	if cliVersion == "dev" {
		return true
	}
	if jsonOutput || quiet {
		return true
	}
	if os.Getenv("SABER_NO_UPDATE_CHECK") != "" {
		return true
	}
	// The explicit "update" command already performs its own check.
	if cmd.Name() == "update" {
		return true
	}
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		return true
	}
	return false
}
