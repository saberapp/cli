package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

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
	cmd.AddCommand(newTemplateExtractCmd())
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

// ── Extract (v1.5 stop-gap) ───────────────────────────────────────────────────

func newTemplateExtractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract reusable templates from historical ad-hoc signals",
		Long: `Cluster historical ad-hoc signal questions into reusable templates so they
can be referenced by scoring rules.

Two-step flow:

  propose → run LLM clustering, get a plan you can review and edit
  apply   → create the templates and back-fill the executions in one transaction

Typical usage:

  saber template extract propose --type company --json > plan.json
  # review and edit plan.json — drop unwanted clusters, tweak names/questions
  saber template extract apply --from-file plan.json

Only score-compatible answer types (boolean, number, percentage, currency,
list) are considered — open-text answers can't be consumed by scoring rules.`,
	}
	cmd.AddCommand(newTemplateExtractProposeCmd())
	cmd.AddCommand(newTemplateExtractApplyCmd())
	return cmd
}

func newTemplateExtractProposeCmd() *cobra.Command {
	var (
		signalType    string
		maxCandidates int
	)
	cmd := &cobra.Command{
		Use:   "propose",
		Short: "Propose template clusters from historical ad-hoc signals",
		Long: `Pulls historical ad-hoc signal_executions for the org and runs LLM clustering
to group equivalent questions into reusable templates. Existing templates are
shown to the model so a cluster matching one is attached rather than duplicated.

The proposal is read-only — no templates are created until you run apply.`,
		Example: `  saber template extract propose --type company
  saber template extract propose --type contact --max-candidates 200 --json > plan.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := parseExtractSignalType(signalType)
			if err != nil {
				return err
			}
			c, ctx := mustClient()
			req := client.ExtractProposeRequest{
				SignalType:    st,
				MaxCandidates: maxCandidates,
			}
			if jsonOutput {
				_, err := c.ProposeExtractTemplates(ctx, req, os.Stdout)
				return err
			}
			p, err := c.ProposeExtractTemplates(ctx, req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintExtractProposal(os.Stdout, p)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&signalType, "type", "", "Signal type to cluster: company or contact")
	cmd.Flags().IntVar(&maxCandidates, "max-candidates", 0, "Max candidates to process (server caps at 500)")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func newTemplateExtractApplyCmd() *cobra.Command {
	var fromFile string
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a reviewed template-extract plan",
		Long: `Send a (possibly edited) plan from extract propose back to the API. Creates
new templates for kind="new" clusters, attaches executions to existing
templates for kind="existing" clusters, in a single transaction.

Re-running the same plan returns 409 — drop already-attached executionIds
before retrying.

The plan can be either a full propose response (the JSON object with a
"clusters" key) or a bare clusters array. Use "-" to read from stdin.`,
		Example: `  saber template extract apply --from-file plan.json
  saber template extract propose --type company --json | saber template extract apply --from-file -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clusters, err := loadExtractClusters(fromFile)
			if err != nil {
				return err
			}
			if len(clusters) == 0 {
				return fmt.Errorf("no clusters to apply")
			}
			c, ctx := mustClient()
			req := client.ExtractApplyRequest{Clusters: clusters}
			if jsonOutput {
				_, err := c.ApplyExtractTemplates(ctx, req, os.Stdout)
				return err
			}
			r, err := c.ApplyExtractTemplates(ctx, req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintExtractApplyResult(os.Stdout, r)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Path to a JSON plan ('-' for stdin)")
	_ = cmd.MarkFlagRequired("from-file")
	return cmd
}

// parseExtractSignalType normalises --type to the API's COMPANY/CONTACT enum.
func parseExtractSignalType(s string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "company":
		return "COMPANY", nil
	case "contact":
		return "CONTACT", nil
	}
	return "", fmt.Errorf("type must be 'company' or 'contact', got %q", s)
}

// loadExtractClusters reads a plan file (or stdin when path == "-") and returns
// the clusters to send to apply. Accepts either the full propose response or a
// bare clusters array, so users can pipe `propose --json` straight in or save
// it to disk and edit before applying.
func loadExtractClusters(path string) ([]client.ExtractCluster, error) {
	var raw []byte
	var err error
	if path == "-" {
		raw, err = io.ReadAll(os.Stdin)
	} else {
		raw, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, fmt.Errorf("read plan: %w", err)
	}

	trimmed := strings.TrimSpace(string(raw))
	if strings.HasPrefix(trimmed, "[") {
		var clusters []client.ExtractCluster
		if err := json.Unmarshal(raw, &clusters); err != nil {
			return nil, fmt.Errorf("parse plan as clusters array: %w", err)
		}
		return clusters, nil
	}

	var wrapper struct {
		Clusters []client.ExtractCluster `json:"clusters"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, fmt.Errorf("parse plan: %w", err)
	}
	return wrapper.Clusters, nil
}
