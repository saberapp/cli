package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/format"
	"github.com/spf13/cobra"
)

func newSignalCmd() *cobra.Command {
	var (
		domain           string
		profile          string
		question         string
		answerType       string
		forceRefresh     bool
		webhook          string
		noWait           bool
		maxWait          int
		templateID       string
		verificationMode string
	)

	cmd := &cobra.Command{
		Use:   "signal",
		Short: "Run a signal research query",
		Long: `Run a signal research query against a company domain or a contact LinkedIn profile.

By default the request is synchronous — the command waits for the result and
prints it in one shot. Use --no-wait to fire-and-forget and get back a signal
ID immediately.

Examples:
  saber signal --domain acme.com --question "Are they hiring engineers?"
  saber signal --domain acme.com --question "Headcount?" --answer-type number
  saber signal --domain acme.com --template <templateId>
  saber signal --profile https://linkedin.com/in/johndoe --question "What is their current role?"
  saber signal --domain acme.com --question "..." --verification-mode lenient
  saber signal --domain acme.com --question "..." --no-wait
  saber signal --domain acme.com --question "..." --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if domain == "" && profile == "" {
				return fmt.Errorf("one of --domain or --profile is required")
			}
			if domain != "" && profile != "" {
				return fmt.Errorf("--domain and --profile are mutually exclusive")
			}
			if question == "" && templateID == "" {
				return fmt.Errorf("one of --question or --template is required")
			}

			c, ctx := mustClient()

			if profile != "" {
				return signalContact(c, ctx, profile, question, answerType, forceRefresh, webhook, noWait, maxWait, templateID, verificationMode)
			}
			return signalCompany(c, ctx, domain, question, answerType, forceRefresh, webhook, noWait, maxWait, templateID, verificationMode)
		},
	}

	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Company domain (e.g. acme.com)")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Contact LinkedIn profile URL")
	cmd.Flags().StringVarP(&question, "question", "q", "", "Research question")
	cmd.Flags().StringVarP(&answerType, "answer-type", "a", "", "Answer type: open_text, boolean, number, list, percentage, currency, url, json_schema")
	cmd.Flags().StringVar(&templateID, "template", "", "Signal template ID (alternative to --question)")
	cmd.Flags().StringVar(&verificationMode, "verification-mode", "", "Verification mode: strict (default) or lenient")
	cmd.Flags().BoolVar(&forceRefresh, "force-refresh", false, "Bypass 12h cache")
	cmd.Flags().StringVar(&webhook, "webhook", "", "Webhook URL (async, skips waiting)")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "Async: print signal ID and exit immediately")
	cmd.Flags().IntVar(&maxWait, "max-wait", 120, "Max seconds to wait for result (sync mode)")

	cmd.AddCommand(newSignalGetCmd())
	cmd.AddCommand(newSignalListCmd())
	cmd.AddCommand(newSignalBatchCmd())

	return cmd
}

func newSignalGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <signalId>",
		Short: "Get a signal result by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.GetSignal(ctx, args[0], os.Stdout)
				return err
			}
			sig, err := c.GetSignal(ctx, args[0], nil)
			if err != nil {
				return fmt.Errorf("get signal: %w", err)
			}
			if quiet {
				return nil
			}
			if sig.Status == client.SignalStatusProcessing {
				fmt.Fprintf(os.Stdout, "Signal still processing  ID: %s\n", sig.ID)
				return nil
			}
			format.PrintSignal(os.Stdout, sig)
			return nil
		},
	}
}

func signalCompany(c *client.Client, ctx context.Context, domain, question, answerType string, forceRefresh bool, webhook string, noWait bool, maxWait int, templateID, verificationMode string) error {
	if err := confirmCreditAction(c, ctx); err != nil {
		return err
	}
	req := client.CreateCompanySignalRequest{
		Domain:           domain,
		Question:         question,
		AnswerType:       answerType,
		ForceRefresh:     forceRefresh,
		WebhookURL:       webhook,
		SignalTemplateID: templateID,
		VerificationMode: verificationMode,
	}
	if noWait || webhook != "" {
		sig, err := c.CreateCompanySignal(ctx, req)
		if err != nil {
			return fmt.Errorf("create signal: %w", err)
		}
		return printSignalCreated(sig)
	}
	return runSyncSignal(func() (*client.Signal, error) {
		return c.CreateCompanySignalSync(ctx, req, maxWait, nil)
	}, func() error {
		_, err := c.CreateCompanySignalSync(ctx, req, maxWait, os.Stdout)
		return err
	})
}

func signalContact(c *client.Client, ctx context.Context, profile, question, answerType string, forceRefresh bool, webhook string, noWait bool, maxWait int, templateID, verificationMode string) error {
	if err := confirmCreditAction(c, ctx); err != nil {
		return err
	}
	req := client.CreateContactSignalRequest{
		ContactProfileURL: profile,
		Question:          question,
		AnswerType:        answerType,
		ForceRefresh:      forceRefresh,
		WebhookURL:        webhook,
		SignalTemplateID:  templateID,
		VerificationMode:  verificationMode,
	}
	if noWait || webhook != "" {
		sig, err := c.CreateContactSignal(ctx, req)
		if err != nil {
			return fmt.Errorf("create signal: %w", err)
		}
		return printSignalCreated(sig)
	}
	return runSyncSignal(func() (*client.Signal, error) {
		return c.CreateContactSignalSync(ctx, req, maxWait, nil)
	}, func() error {
		_, err := c.CreateContactSignalSync(ctx, req, maxWait, os.Stdout)
		return err
	})
}

func newSignalListCmd() *cobra.Command {
	var (
		domain         string
		companyID      string
		status         []string
		fromDate       string
		toDate         string
		subscriptionID string
		limit          int
		offset         int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List company signals with optional filters",
		Example: `  saber signal list --domain acme.com
  saber signal list --status completed --limit 10
  saber signal list --domain acme.com --from-date 2024-01-01T00:00:00Z`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			params := client.ListSignalsParams{
				Domain:         domain,
				CompanyID:      companyID,
				Status:         status,
				FromDate:       fromDate,
				ToDate:         toDate,
				SubscriptionID: subscriptionID,
				Limit:          limit,
				Offset:         offset,
			}
			if jsonOutput {
				_, err := c.ListSignals(ctx, params, os.Stdout)
				return err
			}
			resp, err := c.ListSignals(ctx, params, nil)
			if err != nil {
				return err
			}
			if quiet {
				return nil
			}
			if len(resp.Results) == 0 {
				fmt.Fprintln(os.Stdout, "No signals found.")
				return nil
			}
			format.PrintSignalList(os.Stdout, resp.Results, resp.Total)
			return nil
		},
	}
	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Filter by company domain")
	cmd.Flags().StringVar(&companyID, "company-id", "", "Filter by company ID")
	cmd.Flags().StringArrayVar(&status, "status", nil, "Filter by status: processing, completed, failed (repeatable)")
	cmd.Flags().StringVar(&fromDate, "from-date", "", "Filter signals completed after this date (RFC3339)")
	cmd.Flags().StringVar(&toDate, "to-date", "", "Filter signals completed before this date (RFC3339)")
	cmd.Flags().StringVar(&subscriptionID, "subscription-id", "", "Filter by subscription ID")
	cmd.Flags().IntVar(&limit, "limit", 25, "Max results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset")
	return cmd
}

func newSignalBatchCmd() *cobra.Command {
	var (
		domains         []string
		questions       []string
		templateIDs     []string
		answerType      string
		async           bool
		generateSummary bool
		webhook         string
		forceRefresh    bool
	)
	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Create multiple signals in batch (Cartesian product of signals x domains)",
		Long: `Submit multiple signal questions and domains to create signals using a Cartesian
product pattern. Each signal question is combined with each domain.

You can provide inline questions via --question (repeatable), template IDs via
--template (repeatable), or mix both.

By default runs in sync mode (max 100 signals). Use --async for large batches
up to 20,000 signals.`,
		Example: `  saber signal batch --domain acme.com --domain google.com --question "What CRM do they use?" --question "Are they hiring?"
  saber signal batch --domain acme.com --template <id1> --template <id2>
  saber signal batch --domain acme.com --domain google.com --question "Revenue?" --generate-summary
  saber signal batch --domain acme.com --question "..." --async`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(domains) == 0 {
				return fmt.Errorf("at least one --domain is required")
			}
			if len(questions) == 0 && len(templateIDs) == 0 {
				return fmt.Errorf("at least one --question or --template is required")
			}

			c, ctx := mustClient()
			if err := confirmCreditAction(c, ctx); err != nil {
				return err
			}

			var signals []client.BatchSignalItem
			for _, q := range questions {
				signals = append(signals, client.BatchSignalItem{
					Question:     q,
					AnswerType:   answerType,
					WebhookURL:   webhook,
					ForceRefresh: forceRefresh,
				})
			}
			for _, tid := range templateIDs {
				signals = append(signals, client.BatchSignalItem{
					TemplateID:   tid,
					WebhookURL:   webhook,
					ForceRefresh: forceRefresh,
				})
			}

			req := client.CreateSignalBatchRequest{
				Signals:                   signals,
				Domains:                   domains,
				GenerateSummaryOnComplete: generateSummary,
				Async:                     async,
			}

			sp := format.NewSpinner(os.Stderr, jsonOutput || quiet)
			sp.Start("Submitting batch...")

			if jsonOutput {
				sp.Stop()
				_, err := c.CreateSignalBatch(ctx, req, os.Stdout)
				return err
			}

			resp, err := c.CreateSignalBatch(ctx, req, nil)
			sp.Stop()
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintBatchResult(os.Stdout, resp)
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVarP(&domains, "domain", "d", nil, "Company domain (repeatable, required)")
	cmd.Flags().StringArrayVarP(&questions, "question", "q", nil, "Research question (repeatable)")
	cmd.Flags().StringArrayVar(&templateIDs, "template", nil, "Signal template ID (repeatable)")
	cmd.Flags().StringVarP(&answerType, "answer-type", "a", "", "Answer type for inline questions")
	cmd.Flags().BoolVar(&async, "async", false, "Async mode for large batches (up to 20,000 signals)")
	cmd.Flags().BoolVar(&generateSummary, "generate-summary", false, "Auto-generate summaries when all signals complete")
	cmd.Flags().StringVar(&webhook, "webhook", "", "Webhook URL for each signal in the batch")
	cmd.Flags().BoolVar(&forceRefresh, "force-refresh", false, "Bypass the 12-hour answer cache")
	return cmd
}

// runSyncSignal handles the spinner, JSON output, and display for synchronous signal calls.
func runSyncSignal(fetch func() (*client.Signal, error), fetchRaw func() error) error {
	sp := format.NewSpinner(os.Stderr, jsonOutput || quiet)
	sp.Start("Running signal research...")

	if jsonOutput {
		sp.Stop()
		return fetchRaw()
	}

	sig, err := fetch()
	sp.Stop()
	if err != nil {
		return fmt.Errorf("signal: %w", err)
	}
	if quiet {
		return nil
	}
	if sig.Status == client.SignalStatusProcessing {
		fmt.Fprintf(os.Stdout, "Signal still processing  ID: %s\n", sig.ID)
		fmt.Fprintln(os.Stdout, "Re-run later with: saber signal get "+sig.ID)
		return nil
	}
	format.PrintSignal(os.Stdout, sig)
	return nil
}

// printSignalCreated handles --no-wait / --webhook output.
func printSignalCreated(sig *client.Signal) error {
	if jsonOutput {
		fmt.Fprintf(os.Stdout, `{"id":%q,"status":%q}`+"\n", sig.ID, sig.Status)
		return nil
	}
	if !quiet {
		format.PrintSignalCreated(os.Stdout, sig)
	}
	return nil
}
