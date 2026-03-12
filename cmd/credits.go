package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCreditsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "credits",
		Short: "Show remaining credit balance",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()

			if jsonOutput {
				_, err := c.GetCredits(ctx, os.Stdout)
				return err
			}

			resp, err := c.GetCredits(ctx, nil)
			if err != nil {
				return err
			}

			if !quiet {
				fmt.Fprintf(os.Stdout, "Remaining credits: %s\n", formatInt(resp.RemainingCredits))
			}
			return nil
		},
	}
}

// formatInt formats an integer with thousands separators.
func formatInt(n int) string {
	s := fmt.Sprintf("%d", n)
	if n < 1000 {
		return s
	}
	out := make([]byte, 0, len(s)+len(s)/3)
	mod := len(s) % 3
	if mod == 0 {
		mod = 3
	}
	out = append(out, s[:mod]...)
	for i := mod; i < len(s); i += 3 {
		out = append(out, ',')
		out = append(out, s[i:i+3]...)
	}
	return string(out)
}
