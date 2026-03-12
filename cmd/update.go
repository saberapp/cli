package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

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

			latest, err := fetchLatestCLIVersion()
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
				fmt.Fprintln(os.Stdout, "Run:  brew upgrade saber")
			} else {
				fmt.Fprintln(os.Stdout, "Run:  curl -sSL https://install.saber.app | sh")
			}
			return nil
		},
	}
}

// fetchLatestCLIVersion queries the GitHub releases API for the latest release.
func fetchLatestCLIVersion() (string, error) {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get("https://api.github.com/repos/saberapp/cli/releases?per_page=20")
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var releases []struct {
		TagName string `json:"tag_name"`
		Draft   bool   `json:"draft"`
		Prerel  bool   `json:"prerelease"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", fmt.Errorf("parse releases: %w", err)
	}

	for _, r := range releases {
		if r.Draft || r.Prerel {
			continue
		}
		if strings.HasPrefix(r.TagName, "v") {
			return strings.TrimPrefix(r.TagName, "v"), nil
		}
	}
	return "", fmt.Errorf("no release found")
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
