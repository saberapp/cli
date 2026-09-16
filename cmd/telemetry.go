package cmd

import (
	"fmt"
	"os"

	"github.com/saberapp/cli/internal/telemetry"
	"github.com/spf13/cobra"
)

func newTelemetryCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "telemetry",
		Short: "Show or change anonymous usage data collection",
		Long: `Saber CLI can send anonymous usage data: the command name, which flag names
you used, the CLI version, your OS, and how long the command took.

It never sends your API key, your arguments, flag values, or any result data.

Turn it off for one run with SABER_TELEMETRY=0, or everywhere with
DO_NOT_TRACK=1.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTelemetryStatus()
		},
	}
	c.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show whether usage data is sent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTelemetryStatus()
		},
	})
	c.AddCommand(&cobra.Command{
		Use:   "enable",
		Short: "Send anonymous usage data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setTelemetry(true)
		},
	})
	c.AddCommand(&cobra.Command{
		Use:   "disable",
		Short: "Stop sending anonymous usage data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setTelemetry(false)
		},
	})
	return c
}

// setTelemetry persists an explicit answer from the user.
func setTelemetry(enabled bool) error {
	state := telemetry.SaveConsent(telemetry.LoadState(), enabled, cliVersion, telemetry.ConsentModeCommand)
	if jsonOutput {
		return printTelemetryJSON(state)
	}
	if quiet {
		return nil
	}
	if enabled {
		fmt.Fprintln(os.Stdout, "Usage data is on.")
	} else {
		fmt.Fprintln(os.Stdout, "Usage data is off.")
	}
	printEnvOverrideNotice(state)
	return nil
}

func runTelemetryStatus() error {
	state := telemetry.LoadState()
	if jsonOutput {
		return printTelemetryJSON(state)
	}
	if quiet {
		return nil
	}
	switch telemetry.Resolve(state) {
	case telemetry.DecisionEnabled:
		fmt.Fprintln(os.Stdout, "Usage data: on")
	case telemetry.DecisionDisabled:
		fmt.Fprintln(os.Stdout, "Usage data: off")
	case telemetry.DecisionNeedsConsent:
		fmt.Fprintln(os.Stdout, "Usage data: not answered yet (the next command asks)")
	}
	if !telemetry.Configured() {
		fmt.Fprintln(os.Stdout, "This build has no analytics key, so it sends nothing.")
	}
	printEnvOverrideNotice(state)
	return nil
}

// printEnvOverrideNotice tells the user when an environment variable decides
// the outcome, so a saved answer that has no effect is not confusing.
func printEnvOverrideNotice(state *telemetry.State) {
	if os.Getenv("DO_NOT_TRACK") != "" {
		fmt.Fprintln(os.Stdout, "DO_NOT_TRACK is set, so nothing is sent.")
		return
	}
	if os.Getenv("SABER_TELEMETRY") != "" {
		fmt.Fprintf(os.Stdout, "SABER_TELEMETRY=%s overrides the saved answer.\n", os.Getenv("SABER_TELEMETRY"))
	}
	_ = state
}

func printTelemetryJSON(state *telemetry.State) error {
	enabled := telemetry.Resolve(state) == telemetry.DecisionEnabled
	fmt.Fprintf(os.Stdout,
		`{"enabled":%t,"answered":%t,"consentMode":%q,"configured":%t,"organizationKnown":%t}`+"\n",
		enabled, state.Answered(), state.ConsentMode, telemetry.Configured(), state.OrganizationID != "")
	return nil
}
