package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/format"
	"github.com/spf13/cobra"
)

func newMarketSignalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "market-signal",
		Aliases: []string{"ms"},
		Short:   "Monitor market signals (job posts, LinkedIn posts, fundraising, investments, IPOs)",
		Long: `Create and manage market signal subscriptions that continuously monitor external
data sources and deliver matching signals to your webhook.

Subscription types:
  JOB_POSTS          Monitor new job postings matching your criteria
  LINKEDIN_POST      Monitor LinkedIn posts matching keyword filters
  FUND_RAISED        Monitor fund closings and fundraising events
  RECENT_INVESTMENT  Monitor recent investment and funding rounds
  IPO                Monitor IPO and public listing announcements`,
	}
	cmd.AddCommand(newMarketSignalCreateCmd())
	cmd.AddCommand(newMarketSignalListCmd())
	cmd.AddCommand(newMarketSignalGetCmd())
	cmd.AddCommand(newMarketSignalUpdateCmd())
	cmd.AddCommand(newMarketSignalDeleteCmd())
	cmd.AddCommand(newMarketSignalPauseCmd())
	cmd.AddCommand(newMarketSignalResumeCmd())
	cmd.AddCommand(newMarketSignalTriggerCmd())
	cmd.AddCommand(newMarketSignalSignalsCmd())
	return cmd
}

func newMarketSignalCreateCmd() *cobra.Command {
	var (
		subType             string
		name                string
		prompt              string
		filtersJSON         string
		webhookURL          string
		webhookSecret       string
		interval            string
		intervalSignalLimit int
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a market signal subscription",
		Example: `  # Monitor DevOps job postings
  saber market-signal create --type JOB_POSTS --name "DevOps hiring" \
    --webhook https://example.com/hook \
    --filters '{"titleKeywords":["DevOps Engineer"],"countries":["US"]}'

  # Use a prompt to auto-generate filters
  saber market-signal create --type JOB_POSTS --name "AI startups" \
    --prompt "Find companies hiring ML engineers at Series A startups in Europe" \
    --webhook https://example.com/hook

  # Monitor LinkedIn posts about AI
  saber market-signal create --type LINKEDIN_POST --name "AI posts" \
    --filters '{"keywordsAll":["artificial intelligence"],"keywordsAny":["GPT","LLM"]}' \
    --webhook https://example.com/hook

  # Monitor fundraising events
  saber market-signal create --type FUND_RAISED --name "PE funds" \
    --filters '{"keywords":["private equity"],"maxLookbackDays":30}' \
    --webhook https://example.com/hook --interval weekly`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			req := client.CreateMarketSignalSubscriptionRequest{
				Type:                subType,
				Name:                name,
				Prompt:              prompt,
				WebhookURL:          webhookURL,
				WebhookSecret:       webhookSecret,
				Interval:            interval,
				IntervalSignalLimit: intervalSignalLimit,
			}
			if filtersJSON != "" {
				if err := json.Unmarshal([]byte(filtersJSON), &req.Filters); err != nil {
					return fmt.Errorf("invalid --filters JSON: %w", err)
				}
			}
			if jsonOutput {
				_, err := c.CreateMarketSignalSubscription(ctx, req, os.Stdout)
				return err
			}
			sub, err := c.CreateMarketSignalSubscription(ctx, req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintMarketSignalSubscription(os.Stdout, sub)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&subType, "type", "", "Subscription type: JOB_POSTS, LINKEDIN_POST, FUND_RAISED, RECENT_INVESTMENT, IPO (required)")
	cmd.Flags().StringVar(&name, "name", "", "Display name for the subscription")
	cmd.Flags().StringVar(&prompt, "prompt", "", "Natural language prompt for AI filter generation (JOB_POSTS only)")
	cmd.Flags().StringVar(&filtersJSON, "filters", "", "Filters as JSON string")
	cmd.Flags().StringVar(&webhookURL, "webhook", "", "Webhook URL to receive matched signals (required)")
	cmd.Flags().StringVar(&webhookSecret, "webhook-secret", "", "Secret for HMAC-SHA256 webhook signature verification")
	cmd.Flags().StringVar(&interval, "interval", "", "Polling interval: daily (default) or weekly")
	cmd.Flags().IntVar(&intervalSignalLimit, "signal-limit", 0, "Max signals per polling interval (1-10000, default 500)")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("webhook")
	return cmd
}

func newMarketSignalListCmd() *cobra.Command {
	var (
		limit          int
		offset         int
		includeDeleted bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List market signal subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.ListMarketSignalSubscriptions(ctx, limit, offset, includeDeleted, os.Stdout)
				return err
			}
			resp, err := c.ListMarketSignalSubscriptions(ctx, limit, offset, includeDeleted, nil)
			if err != nil {
				return err
			}
			if quiet {
				return nil
			}
			if len(resp.Items) == 0 {
				fmt.Fprintln(os.Stdout, "No market signal subscriptions found.")
				return nil
			}
			format.PrintMarketSignalSubscriptions(os.Stdout, resp.Items, resp.Total)
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "Max results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset")
	cmd.Flags().BoolVar(&includeDeleted, "include-deleted", false, "Include deleted subscriptions")
	return cmd
}

func newMarketSignalGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <subscriptionId>",
		Short: "Get a market signal subscription by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.GetMarketSignalSubscription(ctx, args[0], os.Stdout)
				return err
			}
			sub, err := c.GetMarketSignalSubscription(ctx, args[0], nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintMarketSignalSubscription(os.Stdout, sub)
			}
			return nil
		},
	}
}

func newMarketSignalUpdateCmd() *cobra.Command {
	var (
		name                string
		prompt              string
		filtersJSON         string
		webhookURL          string
		webhookSecret       string
		interval            string
		intervalSignalLimit int
	)
	cmd := &cobra.Command{
		Use:   "update <subscriptionId>",
		Short: "Update a market signal subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("prompt") &&
				!cmd.Flags().Changed("filters") && !cmd.Flags().Changed("webhook") &&
				!cmd.Flags().Changed("webhook-secret") && !cmd.Flags().Changed("interval") &&
				!cmd.Flags().Changed("signal-limit") {
				return fmt.Errorf("at least one flag is required")
			}
			c, ctx := mustClient()
			req := client.UpdateMarketSignalSubscriptionRequest{}
			if cmd.Flags().Changed("name") {
				req.Name = name
			}
			if cmd.Flags().Changed("prompt") {
				req.Prompt = prompt
			}
			if cmd.Flags().Changed("webhook") {
				req.WebhookURL = webhookURL
			}
			if cmd.Flags().Changed("webhook-secret") {
				req.WebhookSecret = webhookSecret
			}
			if cmd.Flags().Changed("interval") {
				req.Interval = interval
			}
			if cmd.Flags().Changed("signal-limit") {
				req.IntervalSignalLimit = &intervalSignalLimit
			}
			if cmd.Flags().Changed("filters") {
				var filters map[string]any
				if err := json.Unmarshal([]byte(filtersJSON), &filters); err != nil {
					return fmt.Errorf("invalid --filters JSON: %w", err)
				}
				req.Filters = filters
			}
			if jsonOutput {
				_, err := c.UpdateMarketSignalSubscription(ctx, args[0], req, os.Stdout)
				return err
			}
			sub, err := c.UpdateMarketSignalSubscription(ctx, args[0], req, nil)
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintMarketSignalSubscription(os.Stdout, sub)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Display name")
	cmd.Flags().StringVar(&prompt, "prompt", "", "Natural language prompt")
	cmd.Flags().StringVar(&filtersJSON, "filters", "", "Filters as JSON string")
	cmd.Flags().StringVar(&webhookURL, "webhook", "", "Webhook URL")
	cmd.Flags().StringVar(&webhookSecret, "webhook-secret", "", "Webhook secret")
	cmd.Flags().StringVar(&interval, "interval", "", "Polling interval: daily or weekly")
	cmd.Flags().IntVar(&intervalSignalLimit, "signal-limit", 0, "Max signals per interval (1-10000)")
	return cmd
}

func newMarketSignalDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <subscriptionId>",
		Short: "Delete a market signal subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if err := c.DeleteMarketSignalSubscription(ctx, args[0]); err != nil {
				return err
			}
			if !quiet {
				fmt.Fprintf(os.Stdout, "Deleted subscription %s\n", args[0])
			}
			return nil
		},
	}
}

func newMarketSignalPauseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pause <subscriptionId>",
		Short: "Pause an active subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.PauseMarketSignalSubscription(ctx, args[0], os.Stdout)
				return err
			}
			sub, err := c.PauseMarketSignalSubscription(ctx, args[0], nil)
			if err != nil {
				return err
			}
			if !quiet {
				fmt.Fprintf(os.Stdout, "Paused subscription %s (status: %s)\n", sub.ID, sub.Status)
			}
			return nil
		},
	}
}

func newMarketSignalResumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resume <subscriptionId>",
		Short: "Resume a paused subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.ResumeMarketSignalSubscription(ctx, args[0], os.Stdout)
				return err
			}
			sub, err := c.ResumeMarketSignalSubscription(ctx, args[0], nil)
			if err != nil {
				return err
			}
			if !quiet {
				fmt.Fprintf(os.Stdout, "Resumed subscription %s (status: %s)\n", sub.ID, sub.Status)
			}
			return nil
		},
	}
}

func newMarketSignalTriggerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trigger <subscriptionId>",
		Short: "Trigger an immediate run of a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.TriggerMarketSignalSubscription(ctx, args[0], os.Stdout)
				return err
			}
			sub, err := c.TriggerMarketSignalSubscription(ctx, args[0], nil)
			if err != nil {
				return err
			}
			if !quiet {
				fmt.Fprintf(os.Stdout, "Triggered subscription %s (status: %s)\n", sub.ID, sub.Status)
			}
			return nil
		},
	}
}

func newMarketSignalSignalsCmd() *cobra.Command {
	var (
		limit  int
		offset int
	)
	cmd := &cobra.Command{
		Use:   "signals <subscriptionId>",
		Short: "List signals delivered by a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			if jsonOutput {
				_, err := c.ListMarketSignals(ctx, args[0], limit, offset, os.Stdout)
				return err
			}
			resp, err := c.ListMarketSignals(ctx, args[0], limit, offset, nil)
			if err != nil {
				return err
			}
			if quiet {
				return nil
			}
			if len(resp.Items) == 0 {
				fmt.Fprintln(os.Stdout, "No signals found for this subscription.")
				return nil
			}
			format.PrintMarketSignals(os.Stdout, resp.Items, resp.Total)
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "Max results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Offset")
	return cmd
}
