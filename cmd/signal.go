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
		domain       string
		profile      string
		question     string
		answerType   string
		forceRefresh bool
		webhook      string
		noWait       bool
		maxWait      int
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
  saber signal --profile https://linkedin.com/in/johndoe --question "What is their current role?"
  saber signal --domain acme.com --question "..." --no-wait
  saber signal --domain acme.com --question "..." --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if domain == "" && profile == "" {
				return fmt.Errorf("one of --domain or --profile is required")
			}
			if domain != "" && profile != "" {
				return fmt.Errorf("--domain and --profile are mutually exclusive")
			}

			c, ctx := mustClient()

			if profile != "" {
				return signalContact(c, ctx, profile, question, answerType, forceRefresh, webhook, noWait, maxWait)
			}
			return signalCompany(c, ctx, domain, question, answerType, forceRefresh, webhook, noWait, maxWait)
		},
	}

	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Company domain (e.g. acme.com)")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Contact LinkedIn profile URL")
	cmd.Flags().StringVarP(&question, "question", "q", "", "Research question")
	cmd.Flags().StringVarP(&answerType, "answer-type", "a", "", "Answer type: open_text, boolean, number, list, percentage, currency, url, json_schema")
	cmd.Flags().BoolVar(&forceRefresh, "force-refresh", false, "Bypass 12h cache")
	cmd.Flags().StringVar(&webhook, "webhook", "", "Webhook URL (async, skips waiting)")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "Async: print signal ID and exit immediately")
	cmd.Flags().IntVar(&maxWait, "max-wait", 120, "Max seconds to wait for result (sync mode)")

	_ = cmd.MarkFlagRequired("question")

	cmd.AddCommand(newSignalGetCmd())

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

func signalCompany(c *client.Client, ctx context.Context, domain, question, answerType string, forceRefresh bool, webhook string, noWait bool, maxWait int) error {
	req := client.CreateCompanySignalRequest{
		Domain:       domain,
		Question:     question,
		AnswerType:   answerType,
		ForceRefresh: forceRefresh,
		WebhookURL:   webhook,
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

func signalContact(c *client.Client, ctx context.Context, profile, question, answerType string, forceRefresh bool, webhook string, noWait bool, maxWait int) error {
	req := client.CreateContactSignalRequest{
		ContactProfileURL: profile,
		Question:          question,
		AnswerType:        answerType,
		ForceRefresh:      forceRefresh,
		WebhookURL:        webhook,
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
