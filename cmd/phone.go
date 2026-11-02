package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/format"
	"github.com/spf13/cobra"
)

const phoneCreditMessage = "This lookup consumes credits only if a mobile number is found (balance: %s). Proceed? [y/N] "

func newContactFindPhoneCmd() *cobra.Command {
	var req client.FindPhoneRequest
	var maxWait int
	var noWait bool
	cmd := &cobra.Command{
		Use:   "find-phone",
		Short: "Find a verified mobile number for a contact",
		Long: `Find a verified mobile number using a profile URL or a name and company.
Phone lookup must be enabled for your organization.

The server waits up to --max-wait seconds. Use --no-wait to return immediately.
If the lookup is still processing, use get with its run ID to read it later.`,
		Example: `  saber contact find-phone --profile https://www.linkedin.com/in/jane-doe-example
  saber contact find-phone --first-name Jane --last-name Doe --domain example.com
  saber contact find-phone get <runId>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := validatePhoneFlags(req, maxWait, noWait, cmd.Flags().Changed("max-wait"))
			if err != nil {
				return err
			}
			c, ctx := mustClient()
			if err := confirmCreditActionWithMessage(c, ctx, phoneCreditMessage); err != nil {
				return err
			}
			waitSec := maxWait
			if noWait {
				waitSec = 0
			}
			return runPhoneLookup(c, ctx, input, waitSec)
		},
	}
	cmd.Flags().StringVar(&req.ContactProfileURL, "profile", "", "Contact profile URL")
	cmd.Flags().StringVar(&req.FirstName, "first-name", "", "Contact first name")
	cmd.Flags().StringVar(&req.LastName, "last-name", "", "Contact last name")
	cmd.Flags().StringVar(&req.CompanyName, "company", "", "Company name")
	cmd.Flags().StringVar(&req.CompanyDomain, "domain", "", "Company domain")
	cmd.Flags().IntVar(&maxWait, "max-wait", 120, "Max seconds the server waits for a result (1-120)")
	cmd.Flags().BoolVar(&noWait, "no-wait", false, "Return immediately with the run ID or a finished result")
	cmd.AddCommand(newPhoneGetCmd())
	return cmd
}

func validatePhoneFlags(req client.FindPhoneRequest, maxWait int, noWait, maxWaitSet bool) (client.FindPhoneRequest, error) {
	req.ContactProfileURL = strings.TrimSpace(req.ContactProfileURL)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.CompanyName = strings.TrimSpace(req.CompanyName)
	req.CompanyDomain = strings.TrimSpace(req.CompanyDomain)
	if maxWait < 1 || maxWait > 120 {
		return req, fmt.Errorf("--max-wait must be between 1 and 120")
	}
	if noWait && maxWaitSet {
		return req, fmt.Errorf("--no-wait and --max-wait are mutually exclusive")
	}
	if req.ContactProfileURL == "" && (req.FirstName == "" || req.LastName == "" || (req.CompanyName == "" && req.CompanyDomain == "")) {
		return req, fmt.Errorf("provide --profile or --first-name and --last-name with --company or --domain")
	}
	return req, nil
}

func runPhoneLookup(c *client.Client, ctx context.Context, req client.FindPhoneRequest, waitSec int) error {
	sp := format.NewSpinner(os.Stderr, jsonOutput || quiet)
	sp.Start("Looking up mobile number...")
	var rawDst io.Writer
	if jsonOutput && !quiet {
		rawDst = os.Stdout
	}
	run, err := c.FindPhone(ctx, req, waitSec, rawDst)
	sp.Stop()
	if err != nil {
		return phoneLookupError(err, false)
	}
	if !jsonOutput && !quiet {
		format.PrintPhoneRun(os.Stdout, run)
	}
	return nil
}

func newPhoneGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <runId>",
		Short: "Get a phone lookup run by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, ctx := mustClient()
			var rawDst io.Writer
			if jsonOutput && !quiet {
				rawDst = os.Stdout
			}
			run, err := c.GetPhoneRun(ctx, args[0], rawDst)
			if err != nil {
				return phoneLookupError(err, true)
			}
			if !jsonOutput && !quiet {
				format.PrintPhoneRun(os.Stdout, run)
			}
			return nil
		},
	}
}

func phoneLookupError(err error, read bool) error {
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case http.StatusNotFound:
			if read {
				return fmt.Errorf("phone lookup run not found")
			}
			return fmt.Errorf("phone lookup is not enabled for your organization")
		case http.StatusPaymentRequired:
			return fmt.Errorf("not enough credits; check your balance with: saber credits")
		}
	}
	var netErr net.Error
	if !read && errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("%w; run the same command again; it continues the lookup in progress", err)
	}
	return err
}
