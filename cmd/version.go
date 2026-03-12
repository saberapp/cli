package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newVersionCmd(version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			if jsonOutput {
				fmt.Fprintf(os.Stdout, `{"version":%q,"commit":%q,"date":%q}`+"\n", version, commit, date)
				return
			}
			fmt.Fprintf(os.Stdout, "saber version %s (commit %s, built %s)\n", version, commit, date)
		},
	}
}
