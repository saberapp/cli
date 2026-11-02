package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/saberapp/cli/internal/client"
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
			// Recorded for the telemetry event written after the command ends.
			executedCmd = cmd
			startTelemetry(cmd)
			if shouldSkipUpdateCheck(cmd) {
				return
			}
			updateCh = update.RunBackgroundCheck(cliVersion, 24*time.Hour)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			// The minimum-version warning replaces the generic update
			// notice, so the user sees one line, not two.
			if printMinVersionWarning() {
				return
			}
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
	root.AddCommand(newScoringCmd())
	root.AddCommand(newTelemetryCmd())
	root.AddCommand(newInitClaudeCmd())
	root.SetHelpCommand(newHelpCmd(root))

	return root
}

// printMinVersionWarning prints one stderr line when the API reported a minimum
// CLI version that this build is below, or has reached while a newer release
// is available. It makes no network call: the minimum comes from this
// command's API responses and the latest release from the update-check cache.
// It reports whether it printed.
func printMinVersionWarning() bool {
	if cliVersion == "dev" || jsonOutput || quiet || !term.IsTerminal(int(os.Stderr.Fd())) {
		return false
	}
	minimum := client.MinSupportedVersion()
	if minimum == "" {
		return false
	}
	msg := minVersionWarning(cliVersion, minimum, update.LoadState().LatestVersion)
	if msg == "" {
		return false
	}
	fmt.Fprintln(os.Stderr, msg)
	return true
}

// minVersionWarning returns the warning for current against the API minimum and
// the latest known release, or "" when no warning applies.
func minVersionWarning(current, minimum, latest string) string {
	if minimum == "" {
		return ""
	}
	if update.IsBelow(current, minimum) {
		return update.FormatMinVersionWarning(current, minimum)
	}
	atFloor := !update.IsBelow(minimum, current)
	if atFloor && update.IsBelow(current, latest) {
		return update.FormatMinVersionFloorWarning(current, latest)
	}
	return ""
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
