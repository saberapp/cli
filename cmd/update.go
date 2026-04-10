package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/saberapp/cli/internal/update"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Check for a newer version of the CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			current := cliVersion
			if current == "dev" {
				fmt.Fprintln(os.Stdout, "Running a dev build — version checks are not available.")
				return nil
			}

			latest, err := update.FetchLatestVersion()
			if err != nil {
				return fmt.Errorf("check for updates: %w", err)
			}

			if jsonOutput {
				fmt.Fprintf(os.Stdout, `{"current":%q,"latest":%q,"upToDate":%v}`+"\n",
					current, latest, current == latest)
				return nil
			}

			if current == latest {
				fmt.Fprintf(os.Stdout, "Already up to date (v%s)\n", current)
				return nil
			}

			fmt.Fprintf(os.Stdout, "Update available: v%s → v%s\n\n", current, latest)
			if isHomebrewInstall() {
				return runBrewUpgrade()
			}
			fmt.Fprintln(os.Stdout, "Run:  curl -sSL https://install.saber.app | sh")
			return nil
		},
	}
}

// runBrewUpgrade runs `brew update && brew upgrade saber`, streaming output to the terminal.
func runBrewUpgrade() error {
	for _, args := range [][]string{{"update"}, {"upgrade", "saber"}} {
		cmd := exec.Command("brew", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("brew %s: %w", strings.Join(args, " "), err)
		}
	}
	return nil
}

// isHomebrewInstall reports whether the running binary lives inside a Homebrew prefix.
func isHomebrewInstall() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	lower := strings.ToLower(exe)
	return strings.Contains(lower, "homebrew") || strings.Contains(lower, "cellar")
}
