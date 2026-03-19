package cmd

import (
	"fmt"
	"os"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/format"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	list := &cobra.Command{
		Use:   "list",
		Short: "Manage company and contact lists",
	}
	list.AddCommand(newListCompanyCmd())
	list.AddCommand(newListContactCmd())
	return list
}

// ── Company Lists ─────────────────────────────────────────────────────────────

func newListCompanyCmd() *cobra.Command {
	company := &cobra.Command{
		Use:   "company",
		Short: "Manage company lists",
	}
	company.AddCommand(newListCompanyCreateCmd())
	company.AddCommand(newListCompanyListCmd())
	company.AddCommand(newListCompanyGetCmd())
	company.AddCommand(newListCompanyUpdateCmd())
	company.AddCommand(newListCompanyDeleteCmd())
	company.AddCommand(newListCompanyCompaniesCmd())
	company.AddCommand(newListCompanyImportCmd())
	return company
}

func newListCompanyCreateCmd() *cobra.Command {
	var (
		name         string
		industries   []string
		sizes        []string
		countryCodes []string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a company list",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			req := client.CreateCompanyListRequest{
				Name: name,
				Filter: client.CompanyListFilter{
					Industries: industries,
					Sizes:      sizes,
				},
			}
			if len(countryCodes) > 0 {
				req.Filter.Location = &client.CompanyListLocation{CountryCodes: countryCodes}
			}
			if jsonOutput {
				_, err := c.CreateCompanyList(ctx, req, os.Stdout)
				return err
			}
			list, err := c.CreateCompanyList(ctx, req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintCompanyList(os.Stdout, list)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "List name")
	cmd.Flags().StringArrayVar(&industries, "industry", nil, "Industry filter (repeatable)")
	cmd.Flags().StringArrayVar(&sizes, "size", nil, "Employee size filter (repeatable, e.g. '51-200')")
	cmd.Flags().StringArrayVar(&countryCodes, "country", nil, "Country code filter (repeatable, e.g. 'US')")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newListCompanyListCmd() *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all company lists",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.ListCompanyLists(ctx, limit, offset, os.Stdout)
				return err
			}
			resp, err := c.ListCompanyLists(ctx, limit, offset, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintCompanyLists(os.Stdout, resp.Items, resp.Total)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "Max results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset")
	return cmd
}

func newListCompanyGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <listId>",
		Short: "Get a company list by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.GetCompanyList(ctx, args[0], os.Stdout)
				return err
			}
			list, err := c.GetCompanyList(ctx, args[0], nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintCompanyList(os.Stdout, list)
			}
			return nil
		},
	}
}

func newListCompanyUpdateCmd() *cobra.Command {
	var (
		name         string
		industries   []string
		sizes        []string
		countryCodes []string
	)
	cmd := &cobra.Command{
		Use:   "update <listId>",
		Short: "Update a company list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("industry") &&
				!cmd.Flags().Changed("size") && !cmd.Flags().Changed("country") {
				return fmt.Errorf("at least one of --name, --industry, --size, --country is required")
			}
			c, ctx := mustClient()
			req := client.UpdateCompanyListRequest{
				Name: name,
				Filter: client.CompanyListFilter{
					Industries: industries,
					Sizes:      sizes,
				},
			}
			if len(countryCodes) > 0 {
				req.Filter.Location = &client.CompanyListLocation{CountryCodes: countryCodes}
			}
			if jsonOutput {
				_, err := c.UpdateCompanyList(ctx, args[0], req, os.Stdout)
				return err
			}
			list, err := c.UpdateCompanyList(ctx, args[0], req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintCompanyList(os.Stdout, list)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New list name")
	cmd.Flags().StringArrayVar(&industries, "industry", nil, "Industry filter (repeatable)")
	cmd.Flags().StringArrayVar(&sizes, "size", nil, "Employee size filter (repeatable)")
	cmd.Flags().StringArrayVar(&countryCodes, "country", nil, "Country code filter (repeatable)")
	return cmd
}

func newListCompanyDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <listId>",
		Short: "Delete a company list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if err := c.DeleteCompanyList(ctx, args[0]); err != nil {
				return err
			}
			if !quiet {
				fmt.Fprintf(os.Stdout, "Deleted list %s\n", args[0])
			}
			return nil
		},
	}
}

func newListCompanyCompaniesCmd() *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "companies <listId>",
		Short: "List companies in a list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.GetCompaniesInList(ctx, args[0], limit, offset, os.Stdout)
				return err
			}
			resp, err := c.GetCompaniesInList(ctx, args[0], limit, offset, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintCompanies(os.Stdout, resp.Items, resp.Total)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "Max results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset")
	return cmd
}

func newListCompanyImportCmd() *cobra.Command {
	var (
		name         string
		propertyName string
		operator     string
		value        string
	)
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import a company list from HubSpot",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			req := client.ImportCompanyListRequest{
				Name: name,
				Source: client.ImportCompanyListSource{
					Type: "hubspot",
					Filter: client.HubSpotPropertyFilter{
						PropertyName: propertyName,
						Operator:     operator,
						Value:        value,
					},
				},
			}
			if jsonOutput {
				_, err := c.ImportCompanyList(ctx, req, os.Stdout)
				return err
			}
			list, err := c.ImportCompanyList(ctx, req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintCompanyList(os.Stdout, list)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "List name")
	cmd.Flags().StringVar(&propertyName, "property", "", "HubSpot property name to filter on (e.g. 'industry')")
	cmd.Flags().StringVar(&operator, "operator", "EQ", "Filter operator: EQ, NEQ, GT, GTE, LT, LTE, HAS_PROPERTY, NOT_HAS_PROPERTY, CONTAINS_TOKEN, NOT_CONTAINS_TOKEN")
	cmd.Flags().StringVar(&value, "value", "", "Filter value")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("property")
	return cmd
}

// ── Contact Lists ─────────────────────────────────────────────────────────────

func newListContactCmd() *cobra.Command {
	contact := &cobra.Command{
		Use:   "contact",
		Short: "Manage contact lists",
	}
	contact.AddCommand(newListContactCreateCmd())
	contact.AddCommand(newListContactListCmd())
	contact.AddCommand(newListContactGetCmd())
	contact.AddCommand(newListContactUpdateCmd())
	contact.AddCommand(newListContactDeleteCmd())
	contact.AddCommand(newListContactShowCmd())
	return contact
}

func newListContactCreateCmd() *cobra.Command {
	var (
		name            string
		companyLinkedIn []string
		titles          []string
		keywords        string
		countries       []string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a contact list via Sales Navigator search",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			req := client.CreateContactListRequest{
				Name: name,
				Filters: client.ContactListFilters{
					CompanyLinkedInURLs: companyLinkedIn,
					JobTitles:           titles,
					Keywords:            keywords,
					Countries:           countries,
				},
			}
			if jsonOutput {
				_, err := c.CreateContactList(ctx, req, os.Stdout)
				return err
			}
			list, err := c.CreateContactList(ctx, req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintContactList(os.Stdout, list)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "List name")
	cmd.Flags().StringArrayVar(&companyLinkedIn, "company-linkedin", nil, "Company LinkedIn URL (repeatable)")
	cmd.Flags().StringArrayVar(&titles, "title", nil, "Job title filter (repeatable)")
	cmd.Flags().StringVar(&keywords, "keyword", "", "Keywords")
	cmd.Flags().StringArrayVar(&countries, "country", nil, "Country code filter (repeatable)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("company-linkedin")
	return cmd
}

func newListContactListCmd() *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all contact lists",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.ListContactLists(ctx, limit, offset, os.Stdout)
				return err
			}
			resp, err := c.ListContactLists(ctx, limit, offset, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintContactLists(os.Stdout, resp.Items, resp.Total)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "Max results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset")
	return cmd
}

func newListContactGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <listId>",
		Short: "Get a contact list by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.GetContactList(ctx, args[0], os.Stdout)
				return err
			}
			list, err := c.GetContactList(ctx, args[0], nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintContactList(os.Stdout, list)
			}
			return nil
		},
	}
}

func newListContactUpdateCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "update <listId>",
		Short: "Update a contact list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			req := client.UpdateContactListRequest{Name: name}
			if jsonOutput {
				_, err := c.UpdateContactList(ctx, args[0], req, os.Stdout)
				return err
			}
			list, err := c.UpdateContactList(ctx, args[0], req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintContactList(os.Stdout, list)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New list name")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newListContactDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <listId>",
		Short: "Delete a contact list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if err := c.DeleteContactList(ctx, args[0]); err != nil {
				return err
			}
			if !quiet {
				fmt.Fprintf(os.Stdout, "Deleted list %s\n", args[0])
			}
			return nil
		},
	}
}

func newListContactShowCmd() *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "show <listId>",
		Short: "List contacts in a contact list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.GetContactsInList(ctx, args[0], limit, offset, os.Stdout)
				return err
			}
			resp, err := c.GetContactsInList(ctx, args[0], limit, offset, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintContacts(os.Stdout, resp.Items, resp.Total)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "Max results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset")
	return cmd
}
