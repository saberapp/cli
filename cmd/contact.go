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
		Short: "Search for contacts and find verified emails",
	}
	cmd.AddCommand(newContactSearchCmd())
	cmd.AddCommand(newContactFindEmailCmd())
	return cmd
}

func newContactFindEmailCmd() *cobra.Command {
	var (
		fullName string
		domain   string
	)
	cmd := &cobra.Command{
		Use:   "find-email",
		Short: "Find a verified email for a contact at a company domain",
		Long: `Find a verified email address for a contact, given their full name and
the domain of the company they work for.

Returns an email plus verification metadata (state, score, accept_all),
or a "no email found" message on a clean miss. Names with European
tussenvoegsels (van, van der, etc.) are handled.

Examples:
  saber contact find-email --full-name "Joey van Ommen" --domain saber.app
  saber contact find-email --full-name "Tim Cook" --domain apple.com --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			req := client.FindEmailRequest{
				FullName: fullName,
				Domain:   domain,
			}
			if jsonOutput {
				_, err := c.FindEmail(ctx, req, os.Stdout)
				return err
			}
			resp, err := c.FindEmail(ctx, req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintFindEmailResult(os.Stdout, fullName, domain, resp)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&fullName, "full-name", "", "Full name of the contact (required)")
	cmd.Flags().StringVar(&domain, "domain", "", "Company mail domain, e.g. acme.com (required)")
	_ = cmd.MarkFlagRequired("full-name")
	_ = cmd.MarkFlagRequired("domain")
	return cmd
}

func newContactSearchCmd() *cobra.Command {
	var (
		companyLinkedIn []string
		titles          []string
		keyword         string
		countries       []string
		departments     []string
		seniorities     []string
		firstName       string
		lastName        string
		limit           int
		offset          int
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
				Departments:         departments,
				SeniorityLevels:     seniorities,
				Limit:               limit,
				Offset:              offset,
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
				format.PrintContactSearchResults(os.Stdout, resp.Items, resp.Total, resp.HasMore, resp.SalesNavConnected)
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&companyLinkedIn, "company-linkedin", nil, "Company LinkedIn URL (repeatable)")
	cmd.Flags().StringArrayVar(&titles, "title", nil, "Job title filter (repeatable)")
	cmd.Flags().StringVar(&keyword, "keyword", "", "Keyword filter")
	cmd.Flags().StringArrayVar(&countries, "country", nil, "Country code filter (repeatable)")
	cmd.Flags().StringArrayVar(&departments, "department", nil, "Department filter (repeatable)")
	cmd.Flags().StringArrayVar(&seniorities, "seniority", nil, "Seniority level filter (repeatable)")
	cmd.Flags().StringVar(&firstName, "first-name", "", "First name filter")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name filter")
	cmd.Flags().IntVar(&limit, "limit", 0, "Max results per page (1-100, API default 25)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Zero-based offset for pagination")
	return cmd
}
