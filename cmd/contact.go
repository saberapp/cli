package cmd

import (
	"os"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/format"
	"github.com/spf13/cobra"
)

func newContactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contact",
		Short: "Search for contacts",
	}
	cmd.AddCommand(newContactSearchCmd())
	return cmd
}

func newContactSearchCmd() *cobra.Command {
	var (
		companyLinkedIn []string
		titles          []string
		keyword         string
		countries       []string
		firstName       string
		lastName        string
	)
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Preview contacts matching filters (without creating a list)",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			req := client.ContactSearchRequest{
				CompanyLinkedInURLs: companyLinkedIn,
				FirstName:           firstName,
				LastName:            lastName,
				JobTitles:           titles,
				Keywords:            keyword,
				Countries:           countries,
			}
			if jsonOutput {
				_, err := c.SearchContacts(ctx, req, os.Stdout)
				return err
			}
			resp, err := c.SearchContacts(ctx, req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintContactSearchResults(os.Stdout, resp.Contacts, resp.Count)
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&companyLinkedIn, "company-linkedin", nil, "Company LinkedIn URL (repeatable)")
	cmd.Flags().StringArrayVar(&titles, "title", nil, "Job title filter (repeatable)")
	cmd.Flags().StringVar(&keyword, "keyword", "", "Keyword filter")
	cmd.Flags().StringArrayVar(&countries, "country", nil, "Country code filter (repeatable)")
	cmd.Flags().StringVar(&firstName, "first-name", "", "First name filter")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name filter")
	return cmd
}
