package cmd

import (
	"fmt"
	"os"

	"github.com/saberapp/cli/internal/format"
	"github.com/spf13/cobra"
)

func newConnectorsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "connectors",
		Short: "List configured connectors",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()

			if jsonOutput {
				_, err := c.ListConnectors(ctx, os.Stdout)
				return err
			}

			resp, err := c.ListConnectors(ctx, nil)
			if err != nil {
				return err
			}

			if quiet {
				return nil
			}

			if len(resp.Connectors) == 0 {
				fmt.Fprintln(os.Stdout, "No connectors configured.")
				return nil
			}

			tw := format.NewTabWriter(os.Stdout)
			fmt.Fprintln(tw, "CONNECTOR\tSTATUS")
			for _, conn := range resp.Connectors {
				fmt.Fprintf(tw, "%s\t%s\n", conn.Source, conn.Status)
			}
			format.FlushTable(tw)
			return nil
		},
	}
}
