package cmd

import (
	"fmt"
	"os"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/format"
	"github.com/spf13/cobra"
)

func newTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Manage reusable signal templates",
		Long: `Create and manage reusable signal templates that define standard research
questions, answer types, and qualification criteria. Templates can be referenced
when creating signals, batches, or subscriptions.`,
	}
	cmd.AddCommand(newTemplateCreateCmd())
	cmd.AddCommand(newTemplateListCmd())
	cmd.AddCommand(newTemplateGetCmd())
	cmd.AddCommand(newTemplateUpdateCmd())
	cmd.AddCommand(newTemplateDeleteCmd())
	return cmd
}

func newTemplateCreateCmd() *cobra.Command {
	var (
		name        string
		question    string
		description string
		answerType  string
		weight      string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a reusable signal template",
		Example: `  saber template create --name "CRM Detection" --question "Which CRM are they using?" --answer-type list
  saber template create --name "Hiring check" --question "Are they hiring engineers?" --answer-type boolean --weight important`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			req := client.CreateSignalTemplateRequest{
				Name:        name,
				Question:    question,
				Description: description,
				AnswerType:  answerType,
				Weight:      weight,
			}
			if jsonOutput {
				_, err := c.CreateSignalTemplate(ctx, req, os.Stdout)
				return err
			}
			tmpl, err := c.CreateSignalTemplate(ctx, req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintSignalTemplate(os.Stdout, tmpl)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Template name (required)")
	cmd.Flags().StringVarP(&question, "question", "q", "", "Research question (required)")
	cmd.Flags().StringVar(&description, "description", "", "Optional description")
	cmd.Flags().StringVarP(&answerType, "answer-type", "a", "", "Answer type: open_text, boolean, number, list, percentage, currency, url, json_schema")
	cmd.Flags().StringVar(&weight, "weight", "", "Signal importance: important, nice_to_have, not_important")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("question")
	return cmd
}

func newTemplateListCmd() *cobra.Command {
	var (
		limit          int
		offset         int
		includeDeleted bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List signal templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.ListSignalTemplates(ctx, limit, offset, includeDeleted, os.Stdout)
				return err
			}
			resp, err := c.ListSignalTemplates(ctx, limit, offset, includeDeleted, nil)
			if err != nil {
				return err
			}
			if quiet {
				return nil
			}
			if len(resp.Items) == 0 {
				fmt.Fprintln(os.Stdout, "No templates found.")
				return nil
			}
			format.PrintSignalTemplates(os.Stdout, resp.Items, resp.Total)
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "Max results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset")
	cmd.Flags().BoolVar(&includeDeleted, "include-deleted", false, "Include deleted templates")
	return cmd
}

func newTemplateGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <templateId>",
		Short: "Get a signal template by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.GetSignalTemplate(ctx, args[0], os.Stdout)
				return err
			}
			tmpl, err := c.GetSignalTemplate(ctx, args[0], nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintSignalTemplate(os.Stdout, tmpl)
			}
			return nil
		},
	}
}

func newTemplateUpdateCmd() *cobra.Command {
	var (
		name        string
		question    string
		description string
		answerType  string
		weight      string
	)
	cmd := &cobra.Command{
		Use:   "update <templateId>",
		Short: "Update a signal template (creates a new version)",
		Long: `Update a signal template using PATCH semantics. Only provided fields are
updated; omitted fields retain their previous values. The template ID stays
the same but the version number is incremented.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("question") &&
				!cmd.Flags().Changed("description") && !cmd.Flags().Changed("answer-type") &&
				!cmd.Flags().Changed("weight") {
				return fmt.Errorf("at least one of --name, --question, --description, --answer-type, --weight is required")
			}
			c, ctx := mustClient()
			req := client.UpdateSignalTemplateRequest{}
			if cmd.Flags().Changed("name") {
				req.Name = name
			}
			if cmd.Flags().Changed("question") {
				req.Question = question
			}
			if cmd.Flags().Changed("description") {
				req.Description = description
			}
			if cmd.Flags().Changed("answer-type") {
				req.AnswerType = answerType
			}
			if cmd.Flags().Changed("weight") {
				req.Weight = weight
			}
			if jsonOutput {
				_, err := c.UpdateSignalTemplate(ctx, args[0], req, os.Stdout)
				return err
			}
			tmpl, err := c.UpdateSignalTemplate(ctx, args[0], req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintSignalTemplate(os.Stdout, tmpl)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Template name")
	cmd.Flags().StringVarP(&question, "question", "q", "", "Research question")
	cmd.Flags().StringVar(&description, "description", "", "Description")
	cmd.Flags().StringVarP(&answerType, "answer-type", "a", "", "Answer type")
	cmd.Flags().StringVar(&weight, "weight", "", "Signal importance: important, nice_to_have, not_important")
	return cmd
}

func newTemplateDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <templateId>",
		Short: "Delete a signal template (soft-delete)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if err := c.DeleteSignalTemplate(ctx, args[0]); err != nil {
				return err
			}
			if !quiet {
				fmt.Fprintf(os.Stdout, "Deleted template %s\n", args[0])
			}
			return nil
		},
	}
}
