package main

import (
	"os"

	"github.com/saberapp/cli/cmd"
)

// Injected by ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// cmd.Execute owns the exit-code mapping and the usage event, so a failed
	// command is still recorded with its real status.
	os.Exit(cmd.Execute(version, commit, date))
}
