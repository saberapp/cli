package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
				fmt.Fprintf(os.Stdout, "Remaining credits: %s\n", formatCredits(resp.RemainingCredits))
			}
			return nil
		},
	}
}

// formatCredits formats a fractional credit balance with thousands separators.
func formatCredits(n float64) string {
	whole, frac, hasFrac := strings.Cut(strconv.FormatFloat(n, 'f', 4, 64), ".")
	formatted := formatWholeNumber(whole)
	if !hasFrac {
		return formatted
	}
	frac = strings.TrimRight(frac, "0")
	if frac == "" {
		return formatted
	}
	return formatted + "." + frac
}

// formatInt formats an integer with thousands separators.
func formatInt(n int) string {
	return formatWholeNumber(fmt.Sprintf("%d", n))
}

func formatWholeNumber(s string) string {
	sign := ""
	if strings.HasPrefix(s, "-") {
		sign = "-"
		s = strings.TrimPrefix(s, "-")
	}
	if len(s) < 4 {
		return sign + s
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
	return sign + string(out)
}
