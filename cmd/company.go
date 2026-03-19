package cmd

import (
	"os"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/format"
	"github.com/spf13/cobra"
)

func newCompanyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "company",
		Short: "Search for companies",
	}
	cmd.AddCommand(newCompanySearchCmd())
	return cmd
}

func newCompanySearchCmd() *cobra.Command {
	var (
		industries   []string
		sizes        []string
		countryCodes []string
	)
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Preview companies matching filters (without creating a list)",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			req := client.CompanySearchRequest{
				Filter: client.CompanyListFilter{
					Industries: industries,
					Sizes:      sizes,
				},
			}
			if len(countryCodes) > 0 {
				req.Filter.Location = &client.CompanyListLocation{CountryCodes: countryCodes}
			}
			if jsonOutput {
				_, err := c.SearchCompanies(ctx, req, os.Stdout)
				return err
			}
			resp, err := c.SearchCompanies(ctx, req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintCompanySearchResults(os.Stdout, resp.Companies, resp.Total)
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&industries, "industry", nil, "Industry filter (repeatable)")
	cmd.Flags().StringArrayVar(&sizes, "size", nil, "Employee size filter (repeatable)")
	cmd.Flags().StringArrayVar(&countryCodes, "country", nil, "Country code filter (repeatable)")
	return cmd
}
