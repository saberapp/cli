package cmd

import (
	"fmt"
	"os"

	"github.com/saberapp/cli/internal/client"
	"github.com/spf13/cobra"
)

func newOrgCmd() *cobra.Command {
	org := &cobra.Command{
		Use:   "org",
		Short: "Manage your organisation profile",
	}
	org.AddCommand(newOrgGetCmd())
	org.AddCommand(newOrgUpdateCmd())
	return org
}

func newOrgGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Show the organisation profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()

			if jsonOutput {
				_, err := c.GetOrganisation(ctx, os.Stdout)
				return err
			}

			resp, err := c.GetOrganisation(ctx, nil)
			if err != nil {
				return err
			}

			if quiet {
				return nil
			}

			printOrg(resp)
			return nil
		},
	}
}

func newOrgUpdateCmd() *cobra.Command {
	var name, website, general, products, useCases, valueProp string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the organisation profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			anyChanged := cmd.Flags().Changed("name") || cmd.Flags().Changed("website") ||
				cmd.Flags().Changed("general") || cmd.Flags().Changed("products") ||
				cmd.Flags().Changed("use-cases") || cmd.Flags().Changed("value-prop")
			if !anyChanged {
				return fmt.Errorf("at least one flag must be provided; see --help")
			}

			c, ctx := mustClient()

			req := client.UpdateOrganisationRequest{}
			if cmd.Flags().Changed("name") {
				req.Name = name
			}
			if cmd.Flags().Changed("website") {
				req.Website = website
			}

			descChanged := cmd.Flags().Changed("general") ||
				cmd.Flags().Changed("products") ||
				cmd.Flags().Changed("use-cases") ||
				cmd.Flags().Changed("value-prop")

			if descChanged {
				// Fetch the current profile first so we send a fully-merged description
				// object — avoids clearing sibling fields the user didn't touch.
				current, err := c.GetOrganisation(ctx, nil)
				if err != nil {
					return fmt.Errorf("could not fetch current org profile to merge description: %w", err)
				}
				merged := current.Description
				if cmd.Flags().Changed("general") {
					merged.General = general
				}
				if cmd.Flags().Changed("products") {
					merged.Products = products
				}
				if cmd.Flags().Changed("use-cases") {
					merged.UseCases = useCases
				}
				if cmd.Flags().Changed("value-prop") {
					merged.ValueProp = valueProp
				}
				req.Description = &merged
			}

			if jsonOutput {
				_, err := c.UpdateOrganisation(ctx, req, os.Stdout)
				return err
			}

			resp, err := c.UpdateOrganisation(ctx, req, nil)
			if err != nil {
				return err
			}

			if quiet {
				return nil
			}

			printOrg(resp)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Organisation name")
	cmd.Flags().StringVar(&website, "website", "", "Organisation website")
	cmd.Flags().StringVar(&general, "general", "", "General description")
	cmd.Flags().StringVar(&products, "products", "", "Products description")
	cmd.Flags().StringVar(&useCases, "use-cases", "", "Use cases description")
	cmd.Flags().StringVar(&valueProp, "value-prop", "", "Value proposition description")

	return cmd
}

func printOrg(o *client.Organisation) {
	if o.Name != "" {
		fmt.Fprintf(os.Stdout, "Name:       %s\n", o.Name)
	}
	if o.Website != "" {
		fmt.Fprintf(os.Stdout, "Website:    %s\n", o.Website)
	}
	if o.Description.General != "" {
		fmt.Fprintf(os.Stdout, "General:    %s\n", o.Description.General)
	}
	if o.Description.Products != "" {
		fmt.Fprintf(os.Stdout, "Products:   %s\n", o.Description.Products)
	}
	if o.Description.UseCases != "" {
		fmt.Fprintf(os.Stdout, "Use cases:  %s\n", o.Description.UseCases)
	}
	if o.Description.ValueProp != "" {
		fmt.Fprintf(os.Stdout, "Value prop: %s\n", o.Description.ValueProp)
	}
}
