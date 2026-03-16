package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// newHelpCmd overrides Cobra's default help command.
// With no args it prints a full cheat sheet; with args it delegates to the
// standard per-command help so `saber help signal` still works as expected.
func newHelpCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "help [command]",
		Short: "Show help for a command, or print the full command reference",
		Long: `With no arguments, prints a reference of every command and its flags.
With a command name, shows detailed help for that command.

Examples:
  saber help
  saber help signal
  saber help list company`,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				printCheatSheet(os.Stdout)
				return nil
			}
			// Delegate to the target command's own help.
			target, _, err := root.Find(args)
			if target == nil || err != nil {
				return fmt.Errorf("unknown command: %s", args[0])
			}
			return target.Help()
		},
	}
}

func printLogo(w io.Writer) {
	cwd, _ := os.Getwd()
	if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(cwd, home) {
		cwd = "~" + cwd[len(home):]
	}

	const (
		bold  = "\x1b[1m"
		dim   = "\x1b[2m"
		reset = "\x1b[0m"
	)

	versionLabel := bold + "Saber CLI" + reset
	if cliVersion != "" {
		versionLabel += dim + " " + cliVersion + reset
	}

	lines := [4]string{"◢██◢██", "██◤██◤", "◢██◢██", "██◤██◤"}
	info := [4]string{
		versionLabel,
		dim + "Saber Team" + reset,
		dim + cwd + reset,
		dim + "Visit developers.saber.app for more guides and support" + reset,
	}

	for i, line := range lines {
		fmt.Fprintf(w, "%s   %s\n", line, info[i])
	}
	fmt.Fprintln(w)
}

func printCheatSheet(w io.Writer) {
	printLogo(w)
	fmt.Fprintln(w, "Saber CLI — command reference")
	fmt.Fprintln(w)

	section(w, "AUTHENTICATION")
	entry(w, "saber auth login [--key sk_live_...]", "Store API key (interactive prompt, or pass via --key)")
	entry(w, "saber auth logout", "Remove stored API key")
	entry(w, "saber auth status", "Show current auth state")
	fmt.Fprintln(w)

	section(w, "SIGNAL RESEARCH")
	entry(w, "saber signal -d <domain> -q \"<question>\"", "Run a company signal (sync by default)")
	entry(w, "saber signal -p <linkedin-url> -q \"<question>\"", "Run a contact signal")
	entry(w, "saber signal get <signalId>", "Fetch a signal result by ID")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  Flags:")
	flag(w, "-d, --domain <domain>", "Company domain (e.g. acme.com)")
	flag(w, "-p, --profile <url>", "Contact LinkedIn URL")
	flag(w, "-q, --question <text>", "Research question  [required]")
	flag(w, "-a, --answer-type <type>", "open_text · boolean · number · list · percentage · currency · url")
	flag(w, "    --force-refresh", "Bypass the 12-hour answer cache")
	flag(w, "    --no-wait", "Async: return signal ID immediately without waiting")
	flag(w, "    --max-wait <seconds>", "Sync timeout sent to server (default 120)")
	flag(w, "    --webhook <url>", "Async: POST result to this URL when complete")
	fmt.Fprintln(w)

	section(w, "COMPANY LISTS")
	entry(w, "saber list company create --name <name> [filters]", "Create a company list")
	entry(w, "saber list company list [--limit 20] [--offset 0]", "List all company lists")
	entry(w, "saber list company get <listId>", "Get a company list by ID")
	entry(w, "saber list company update <listId> [filters]", "Update a company list (at least one flag required)")
	entry(w, "saber list company delete <listId>", "Delete a company list")
	entry(w, "saber list company companies <listId>", "List companies in a list")
	entry(w, "saber list company search [filters]", "Preview companies matching filters without creating a list")
	entry(w, "saber list company import --name <name> --property <prop>", "Import a list from HubSpot")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  Filters (create / update / search):")
	flag(w, "--name <name>", "List name")
	flag(w, "--industry <value>", "Industry filter — repeatable (e.g. \"software development\")")
	flag(w, "--size <range>", "Employee size filter — repeatable (e.g. \"51-200\")")
	flag(w, "--country <code>", "Country code filter — repeatable (e.g. US, GB)")
	fmt.Fprintln(w, "  Import flags:")
	flag(w, "--property <name>", "HubSpot property to filter on (e.g. industry)  [required]")
	flag(w, "--operator <op>", "EQ · NEQ · GT · GTE · LT · LTE · HAS_PROPERTY · CONTAINS_TOKEN (default EQ)")
	flag(w, "--value <value>", "Filter value")
	fmt.Fprintln(w)

	section(w, "CONTACT LISTS")
	entry(w, "saber list contact create --name <name> --company-linkedin <url>", "Create a contact list via Sales Navigator search")
	entry(w, "saber list contact list [--limit 20] [--offset 0]", "List all contact lists")
	entry(w, "saber list contact get <listId>", "Get a contact list by ID")
	entry(w, "saber list contact update <listId> --name <name>", "Rename a contact list")
	entry(w, "saber list contact delete <listId>", "Delete a contact list")
	entry(w, "saber list contact contacts <listId>", "List contacts in a contact list")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  Flags (create):")
	flag(w, "--name <name>", "List name  [required]")
	flag(w, "--company-linkedin <url>", "Company LinkedIn URL — repeatable  [required]")
	flag(w, "--title <title>", "Job title filter — repeatable")
	flag(w, "--keyword <text>", "Keyword filter")
	flag(w, "--country <code>", "Country code filter — repeatable")
	fmt.Fprintln(w)

	section(w, "CREDITS & CONNECTORS")
	entry(w, "saber credits", "Show remaining credit balance")
	entry(w, "saber connectors", "List configured connectors and their status")
	fmt.Fprintln(w)

	section(w, "UTILITY")
	entry(w, "saber version", "Print version, commit, and build date")
	entry(w, "saber update", "Check for a newer version and show upgrade instructions")
	entry(w, "saber help [command]", "Show this reference, or detailed help for a specific command")
	fmt.Fprintln(w)

	section(w, "GLOBAL FLAGS")
	flag(w, "    --json", "Output raw JSON response body to stdout")
	flag(w, "-Q, --quiet", "Suppress all non-error output")
	flag(w, "-v, --verbose", "Log HTTP method, URL, headers, and status to stderr")
	flag(w, "    --api-url <url>", "Override base API URL (default: https://api.saber.app)")
	fmt.Fprintln(w)

	fmt.Fprintln(w, "  Credentials: SABER_API_KEY env var overrides ~/.saber/credentials.json")
	fmt.Fprintln(w, "  API keys:    https://ai.saber.app → Settings → API Keys")
}

func section(w io.Writer, title string) {
	fmt.Fprintln(w, title)
}

func entry(w io.Writer, usage, desc string) {
	fmt.Fprintf(w, "  %-52s %s\n", usage, desc)
}

func flag(w io.Writer, name, desc string) {
	fmt.Fprintf(w, "    %-48s %s\n", name, desc)
}
