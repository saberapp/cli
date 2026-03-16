package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

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

// stripANSI removes ANSI escape sequences so we can measure visual width.
func stripANSI(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if i+1 < len(s) && s[i] == '\x1b' && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			if i < len(s) {
				i++ // skip 'm'
			}
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
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

	logoLines := [4]string{"◢██◢██", "██◤██◤", "◢██◢██", "██◤██◤"}
	infoLines := [4]string{
		versionLabel,
		dim + "Saber Team" + reset,
		dim + cwd + reset,
		dim + "Visit developers.saber.app for more guides and support" + reset,
	}

	// Build each row and measure its visual width.
	type row struct {
		rendered string
		width    int
	}
	rows := make([]row, 4)
	maxWidth := 0
	for i := range logoLines {
		rendered := logoLines[i] + "   " + infoLines[i]
		width := utf8.RuneCountInString(logoLines[i]) + 3 + utf8.RuneCountInString(stripANSI(infoLines[i]))
		rows[i] = row{rendered, width}
		if width > maxWidth {
			maxWidth = width
		}
	}

	const innerPad = 2
	innerWidth := maxWidth + innerPad*2
	border := dim + strings.Repeat("─", innerWidth) + reset

	fmt.Fprintf(w, dim+"┌"+reset+"%s"+dim+"┐"+reset+"\n", border)
	for _, r := range rows {
		rightPad := strings.Repeat(" ", innerWidth-2*innerPad-r.width)
		fmt.Fprintf(w, dim+"│"+reset+"  %s%s  "+dim+"│"+reset+"\n", r.rendered, rightPad)
	}
	fmt.Fprintf(w, dim+"└"+reset+"%s"+dim+"┘"+reset+"\n", border)
	fmt.Fprintln(w)
}

func printCheatSheet(w io.Writer) {
	printLogo(w)
	fmt.Fprintf(w, "\x1b[1mTip:\x1b[0m \x1b[2mUsing Claude Code? Run `saber init-claude` to install Saber Skills and get the most out of the CLI.\x1b[0m\n")
	fmt.Fprintln(w)

	section(w, "COMPANY SIGNALS")
	entry(w, "saber signal -d <domain> -q \"<question>\"", "Run a company signal (sync by default)")
	entry(w, "saber signal get <signalId>", "Fetch a signal result by ID")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  Flags:")
	flag(w, "-d, --domain <domain>", "Company domain (e.g. acme.com)  [required]")
	flag(w, "-q, --question <text>", "Research question  [required]")
	flag(w, "-a, --answer-type <type>", "open_text · boolean · number · list · percentage · currency · url")
	flag(w, "    --force-refresh", "Bypass the 12-hour answer cache")
	flag(w, "    --no-wait", "Async: return signal ID immediately without waiting")
	flag(w, "    --max-wait <seconds>", "Sync timeout sent to server (default 120)")
	flag(w, "    --webhook <url>", "Async: POST result to this URL when complete")
	fmt.Fprintln(w)

	section(w, "COMPANY SEARCH & LISTS")
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

	section(w, "CONTACT SIGNALS")
	entry(w, "saber signal -p <linkedin-url> -q \"<question>\"", "Run a contact signal (sync by default)")
	entry(w, "saber signal get <signalId>", "Fetch a signal result by ID")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  Flags:")
	flag(w, "-p, --profile <url>", "Contact LinkedIn URL  [required]")
	flag(w, "-q, --question <text>", "Research question  [required]")
	flag(w, "-a, --answer-type <type>", "open_text · boolean · number · list · percentage · currency · url")
	flag(w, "    --force-refresh", "Bypass the 12-hour answer cache")
	flag(w, "    --no-wait", "Async: return signal ID immediately without waiting")
	flag(w, "    --max-wait <seconds>", "Sync timeout sent to server (default 120)")
	flag(w, "    --webhook <url>", "Async: POST result to this URL when complete")
	fmt.Fprintln(w)

	section(w, "CONTACT SEARCH & LISTS")
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

	section(w, "ORGANIZATION")
	entry(w, "saber org get", "Show organisation profile")
	entry(w, "saber org update [flags]", "Update organisation profile")
	entry(w, "saber connectors", "List configured connectors and their status")
	entry(w, "saber credits", "Show remaining credit balance")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  Flags (org update):")
	flag(w, "--name <name>", "Organisation name")
	flag(w, "--website <url>", "Organisation website")
	flag(w, "--general <text>", "General description")
	flag(w, "--products <text>", "Products description")
	flag(w, "--use-cases <text>", "Use cases description")
	flag(w, "--value-prop <text>", "Value proposition description")
	fmt.Fprintln(w)

	section(w, "AUTHENTICATION")
	entry(w, "saber auth login [--key sk_live_...]", "Store API key (interactive prompt, or pass via --key)")
	entry(w, "saber auth logout", "Remove stored API key")
	entry(w, "saber auth status", "Show current auth state")
	fmt.Fprintln(w)

	section(w, "UTILITIES")
	entry(w, "saber version", "Print version, commit, and build date")
	entry(w, "saber update", "Check for a newer version and show upgrade instructions")
	entry(w, "saber init-claude", "Initialize CLAUDE.md with Saber context and install Claude Code skills")
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
	fmt.Fprintf(w, "\x1b[1m%s\x1b[0m\n", title)
}

func entry(w io.Writer, usage, desc string) {
	fmt.Fprintf(w, "  %-66s \x1b[2m%s\x1b[0m\n", usage, desc)
}

func flag(w io.Writer, name, desc string) {
	fmt.Fprintf(w, "    %-64s \x1b[2m%s\x1b[0m\n", name, desc)
}
