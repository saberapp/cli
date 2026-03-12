package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/saberapp/cli/cmd"
	"github.com/saberapp/cli/internal/config"
)

// Injected by ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	root := cmd.NewRootCmd(version, commit, date)
	if err := root.Execute(); err != nil {
		var notAuth *config.ErrNotAuthenticated
		if errors.As(err, &notAuth) {
			// Message already printed by the command; just exit with the auth error code.
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
