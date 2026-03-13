package cmd

import (
	"fmt"
	"os"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/format"
	"github.com/spf13/cobra"
)

func newSubscriptionCmd() *cobra.Command {
	sub := &cobra.Command{
		Use:   "subscription",
		Short: "Manage signal subscriptions — run signals against a company list on a schedule",
	}
	sub.AddCommand(newSubscriptionCreateCmd())
	sub.AddCommand(newSubscriptionListCmd())
	sub.AddCommand(newSubscriptionGetCmd())
	sub.AddCommand(newSubscriptionStartCmd())
	sub.AddCommand(newSubscriptionStopCmd())
	sub.AddCommand(newSubscriptionTriggerCmd())
	return sub
}

func newSubscriptionCreateCmd() *cobra.Command {
	var (
		listID     string
		name       string
		question   string
		answerType string
		frequency  string
		cronExpr   string
		timezone   string
		templateID string
		runOnce    bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a signal subscription for a company list",
		Long: `Create a subscription that runs a signal question against every company in a list.

The subscription is created in a stopped state. Use 'saber subscription start <id>'
to activate the schedule, or 'saber subscription trigger <id>' to run it immediately.

Use --run-once to trigger immediately and stop the schedule after — useful for
one-off runs without committing to a recurring schedule.

Either --template or --name + --question is required.
Either --frequency or --cron is required (use --frequency monthly with --run-once
if you don't intend to run on a schedule).`,
		Example: `  saber subscription create --list <listId> --name "Hiring signal" --question "Is this company actively hiring in HR?" --frequency weekly
  saber subscription create --list <listId> --name "One-off check" --question "Opening new locations?" --frequency monthly --run-once
  saber subscription create --list <listId> --template <templateId> --frequency daily`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if templateID == "" && (name == "" || question == "") {
				return fmt.Errorf("either --template or both --name and --question are required")
			}
			if frequency == "" && cronExpr == "" {
				return fmt.Errorf("either --frequency (daily|weekly|monthly) or --cron is required")
			}

			c, ctx := mustClient()
			req := client.CreateSubscriptionRequest{
				SignalTemplateID: templateID,
				Name:             name,
				Question:         question,
				AnswerType:       answerType,
				Frequency:        frequency,
				CronExpression:   cronExpr,
				Timezone:         timezone,
				ListID:           listID,
			}

			sub, err := c.CreateSubscription(ctx, req, nil)
			if err != nil {
				return err
			}

			if runOnce {
				if _, err := c.TriggerSubscription(ctx, sub.ID); err != nil {
					return fmt.Errorf("created subscription %s but failed to trigger: %w", sub.ID, err)
				}
				if _, err := c.StopSubscription(ctx, sub.ID); err != nil {
					return fmt.Errorf("created and triggered subscription %s but failed to stop schedule: %w", sub.ID, err)
				}
				sub.Status = "stopped"
			}

			if jsonOutput {
				_, err := c.GetSubscription(ctx, sub.ID, os.Stdout)
				return err
			}

			if !quiet {
				format.PrintSubscription(os.Stdout, sub)
				if !runOnce {
					fmt.Fprintf(os.Stdout, "\nRun immediately:   saber subscription trigger %s\n", sub.ID)
					fmt.Fprintf(os.Stdout, "Activate schedule: saber subscription start %s\n", sub.ID)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&listID, "list", "", "Company list ID to run signals against (required)")
	cmd.Flags().StringVar(&name, "name", "", "Subscription name (required when not using --template)")
	cmd.Flags().StringVar(&question, "question", "", "Signal question (required when not using --template)")
	cmd.Flags().StringVarP(&answerType, "answer-type", "a", "", "Answer type: boolean, open_text, number, list, percentage, currency, url")
	cmd.Flags().StringVar(&frequency, "frequency", "", "Schedule frequency: daily, weekly, or monthly")
	cmd.Flags().StringVar(&cronExpr, "cron", "", "Custom cron expression, e.g. \"0 9 * * 1\" (mutually exclusive with --frequency)")
	cmd.Flags().StringVar(&timezone, "timezone", "UTC", "IANA timezone for scheduling, e.g. Europe/Amsterdam")
	cmd.Flags().StringVar(&templateID, "template", "", "Existing signal template ID (alternative to --name + --question)")
	cmd.Flags().BoolVar(&runOnce, "run-once", false, "Trigger immediately and stop the schedule — for one-off runs")
	_ = cmd.MarkFlagRequired("list")
	return cmd
}

func newSubscriptionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List signal subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()

			if jsonOutput {
				_, err := c.ListSubscriptions(ctx, 50, 0, os.Stdout)
				return err
			}

			resp, err := c.ListSubscriptions(ctx, 50, 0, nil)
			if err != nil {
				return err
			}

			if quiet {
				return nil
			}

			if len(resp.Items) == 0 {
				fmt.Fprintln(os.Stdout, "No subscriptions found.")
				return nil
			}

			format.PrintSubscriptions(os.Stdout, resp.Items, resp.Total)
			return nil
		},
	}
}

func newSubscriptionGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <subscriptionId>",
		Short: "Get a signal subscription by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()

			if jsonOutput {
				_, err := c.GetSubscription(ctx, args[0], os.Stdout)
				return err
			}

			sub, err := c.GetSubscription(ctx, args[0], nil)
			if err != nil {
				return err
			}

			if !quiet {
				format.PrintSubscription(os.Stdout, sub)
			}
			return nil
		},
	}
}

func newSubscriptionStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <subscriptionId>",
		Short: "Activate the schedule for a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			sub, err := c.StartSubscription(ctx, args[0])
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintSubscription(os.Stdout, sub)
			}
			return nil
		},
	}
}

func newSubscriptionStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <subscriptionId>",
		Short: "Pause the schedule for a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			sub, err := c.StopSubscription(ctx, args[0])
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintSubscription(os.Stdout, sub)
			}
			return nil
		},
	}
}

func newSubscriptionTriggerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trigger <subscriptionId>",
		Short: "Run a subscription immediately, regardless of its schedule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			sub, err := c.TriggerSubscription(ctx, args[0])
			if err != nil {
				return err
			}
			if !quiet {
				format.PrintSubscription(os.Stdout, sub)
			}
			return nil
		},
	}
}
