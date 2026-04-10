package cmd

import (
	"fmt"
	"os"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/format"
	"github.com/spf13/cobra"
)

func newSummaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Generate and view AI-powered signal summaries for a domain",
		Long: `Generate an AI summary consolidating insights from all completed signals for
a company domain. Summaries pull together key data points, qualifications,
and sources from individual signals into a structured overview.`,
	}
	cmd.AddCommand(newSummaryGenerateCmd())
	cmd.AddCommand(newSummaryListCmd())
	return cmd
}

func newSummaryGenerateCmd() *cobra.Command {
	var domain string
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate an AI summary for a domain",
		Long: `Generate a new AI-powered summary of all completed signals for a given domain.
If no completed signals exist, returns empty. If a summary is already in progress,
returns a conflict error. If no new signals since the last summary, returns the
existing summary.`,
		Example: `  saber summary generate --domain acme.com
  saber summary generate --domain acme.com --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			req := client.GenerateSummaryRequest{Domain: domain}

			sp := format.NewSpinner(os.Stderr, jsonOutput || quiet)
			sp.Start("Generating summary...")

			if jsonOutput {
				sp.Stop()
				_, err := c.GenerateSummary(ctx, req, os.Stdout)
				return err
			}

			resp, err := c.GenerateSummary(ctx, req, nil)
			sp.Stop()
			if err != nil {
				return err
			}
			if quiet {
				return nil
			}
			if len(resp.Summary) == 0 {
				fmt.Fprintln(os.Stdout, "No completed signals available for summary.")
				return nil
			}
			format.PrintSummary(os.Stdout, resp.Summary)
			return nil
		},
	}
	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Company domain (e.g. acme.com)")
	_ = cmd.MarkFlagRequired("domain")
	return cmd
}

func newSummaryListCmd() *cobra.Command {
	var (
		domain string
		limit  int
		offset int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List summaries for a domain",
		Example: `  saber summary list --domain acme.com
  saber summary list --domain acme.com --limit 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.ListSummaries(ctx, domain, limit, offset, os.Stdout)
				return err
			}
			resp, err := c.ListSummaries(ctx, domain, limit, offset, nil)
			if err != nil {
				return err
			}
			if quiet {
				return nil
			}
			if len(resp.Results) == 0 {
				fmt.Fprintln(os.Stdout, "No summaries found for this domain.")
				return nil
			}
			format.PrintSummaryRecords(os.Stdout, resp.Results, resp.Total)
			return nil
		},
	}
	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Company domain (e.g. acme.com)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Max results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset")
	_ = cmd.MarkFlagRequired("domain")
	return cmd
}
