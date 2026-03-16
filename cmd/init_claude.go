package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/config"
	"github.com/saberapp/cli/skills"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type skillDef struct {
	name    string
	content string
}

var bundledSkills = []skillDef{
	{"saber-signal-discovery", skills.SignalDiscovery},
	{"saber-create-company-signals", skills.CreateCompanySignals},
	{"saber-create-contact-signals", skills.CreateContactSignals},
	{"saber-build-account-list", skills.BuildAccountList},
	{"saber-build-contact-list", skills.BuildContactList},
}

var saberBlockRe = regexp.MustCompile(`(?s)<!-- saber -->.*?<!-- /saber -->`)
var frontmatterVersionRe = regexp.MustCompile(`(?m)^version:\s*(\d+)`)

func newInitClaudeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init-claude",
		Short: "Initialize CLAUDE.md with Saber context and install Claude Code skills",
		Long: "Writes a <!-- saber --> block to CLAUDE.md in the current directory and\n" +
			"installs Saber skill files to .claude/skills/ for use with Claude Code.",
		RunE: runInitClaude,
	}
}

func runInitClaude(cmd *cobra.Command, _ []string) error {
	commandList := generateCommandList(cmd.Root())

	var connectorSection, orgSection string
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); connectorSection = fetchConnectorSection() }()
	go func() { defer wg.Done(); orgSection = fetchOrgSection() }()
	wg.Wait()

	block := buildSaberBlock(commandList, connectorSection, orgSection)

	claudeMDStatus, err := injectClaudeMD(block)
	if err != nil {
		return fmt.Errorf("failed to update CLAUDE.md: %w", err)
	}

	skillStatuses, err := installSkills()
	if err != nil {
		return fmt.Errorf("failed to install skills: %w", err)
	}

	if quiet {
		return nil
	}

	fmt.Print("\nSaber initialized.\n\n")
	fmt.Printf("  ✓ %s\n", claudeMDStatus)
	for _, s := range skillStatuses {
		fmt.Printf("  %s\n", s)
	}
	fmt.Print("\n  Start with: /saber-signal-discovery\n\n")
	return nil
}

// generateCommandList traverses the cobra command tree and returns a formatted list.
func generateCommandList(root *cobra.Command) string {
	var sb strings.Builder
	var visit func(cmd *cobra.Command, path string)
	visit = func(cmd *cobra.Command, path string) {
		for _, sub := range cmd.Commands() {
			name := sub.Name()
			if name == "help" || name == "completion" {
				continue
			}
			fullPath := path + " " + name

			hasSubs := false
			for _, s := range sub.Commands() {
				if s.Name() != "help" && s.Name() != "completion" {
					hasSubs = true
					break
				}
			}

			if sub.Runnable() {
				sb.WriteString(fullPath)
				parts := strings.Fields(sub.Use)
				if len(parts) > 1 {
					sb.WriteString(" ")
					sb.WriteString(strings.Join(parts[1:], " "))
				}
				sb.WriteString(formatFlags(sub))
				sb.WriteString("\n")
			}

			if hasSubs {
				visit(sub, fullPath)
			}
		}
	}
	visit(root, "saber")
	return strings.TrimRight(sb.String(), "\n")
}

// formatFlags returns a compact flag summary for a command: required flags shown
// explicitly (--name <name>), optional flags collapsed to [--options].
func formatFlags(cmd *cobra.Command) string {
	var required []string
	hasOptional := false
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		if f.Annotations != nil {
			if _, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok {
				required = append(required, "--"+f.Name+" <"+f.Name+">")
				return
			}
		}
		hasOptional = true
	})
	var parts []string
	parts = append(parts, required...)
	if hasOptional {
		parts = append(parts, "[--options]")
	}
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ")
}

// fetchOrgSection tries to fetch the organisation profile from the Saber API.
// Returns a placeholder string if not authenticated or on any error.
func fetchOrgSection() string {
	key, err := config.RequireAPIKey()
	if err != nil {
		return "_Run `saber auth login` to populate organisation context._"
	}

	c := client.New(apiURL, key, cliVersion, false, os.Stderr)
	org, err := c.GetOrganisation(context.Background(), nil)
	if err != nil {
		return "_Could not fetch organisation profile (API error). Run `saber org get` to check manually._"
	}

	if org.Name == "" && org.Website == "" && org.Description.General == "" &&
		org.Description.Products == "" && org.Description.UseCases == "" && org.Description.ValueProp == "" {
		return "_No organisation profile set. Run `saber org update` to add context._"
	}

	var sb strings.Builder
	if org.Name != "" {
		fmt.Fprintf(&sb, "**Name:** %s\n", org.Name)
	}
	if org.Website != "" {
		fmt.Fprintf(&sb, "**Website:** %s\n", org.Website)
	}
	if org.Description.General != "" {
		fmt.Fprintf(&sb, "**General:** %s\n", org.Description.General)
	}
	if org.Description.Products != "" {
		fmt.Fprintf(&sb, "**Products:** %s\n", org.Description.Products)
	}
	if org.Description.UseCases != "" {
		fmt.Fprintf(&sb, "**Use cases:** %s\n", org.Description.UseCases)
	}
	if org.Description.ValueProp != "" {
		fmt.Fprintf(&sb, "**Value prop:** %s\n", org.Description.ValueProp)
	}
	return strings.TrimRight(sb.String(), "\n")
}

// fetchConnectorSection tries to fetch connectors from the Saber API.
// Returns a placeholder string if not authenticated or on any error.
func fetchConnectorSection() string {
	key, err := config.RequireAPIKey()
	if err != nil {
		return "_Run `saber auth login` to populate connector status._"
	}

	c := client.New(apiURL, key, cliVersion, false, os.Stderr)
	resp, err := c.ListConnectors(context.Background(), nil)
	if err != nil {
		return "_Could not fetch connector status (API error). Run `saber connectors list` to check manually._"
	}

	if len(resp.Connectors) == 0 {
		return "_No connectors configured. Visit the Saber dashboard to connect integrations._"
	}

	var sb strings.Builder
	for _, conn := range resp.Connectors {
		fmt.Fprintf(&sb, "| %-20s | %-12s |\n", conn.Source, conn.Status)
	}
	return strings.TrimRight(sb.String(), "\n")
}

// buildSaberBlock builds the full <!-- saber --> ... <!-- /saber --> block.
func buildSaberBlock(commandList, connectorSection, orgSection string) string {
	return `<!-- saber -->
## Saber GTM Intelligence

The Saber CLI is available in this project. Use it proactively for any
revenue, prospecting, or signal-related task.

### Your organisation
` + orgSection + `

### The Saber workflow
1. **Discover signals** — define what buying intent looks like for your ICP
2. **Build lists** — create target account and contact lists
3. **Create signals** — activate signal tracking against your lists

### Reach for Saber when:
- The user wants to define who to target or what signals to track
- The user is building or qualifying an account or contact list
- The user asks who to prioritize or what's showing intent
- Before drafting outreach, building a sequence, or planning a campaign

### Available CLI commands
` + commandList + `

### Connectors
` + connectorSection + `

### Installed skills
- ` + "`/saber-signal-discovery`" + ` — define signals that match your ICP (start here)
- ` + "`/saber-create-company-signals`" + ` — activate company-level signal tracking
- ` + "`/saber-create-contact-signals`" + ` — activate contact-level signal tracking
- ` + "`/saber-build-account-list`" + ` — build a target account list and run signals
- ` + "`/saber-build-contact-list`" + ` — build a target contact list and run signals
<!-- /saber -->`
}

func injectClaudeMD(block string) (string, error) {
	existing := ""
	data, err := os.ReadFile("CLAUDE.md")
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err == nil {
		existing = string(data)
	}

	var updated, status string
	if saberBlockRe.MatchString(existing) {
		updated = saberBlockRe.ReplaceAllString(existing, block)
		status = "CLAUDE.md updated (saber block replaced)"
	} else {
		if existing != "" && !strings.HasSuffix(existing, "\n") {
			existing += "\n"
		}
		if existing != "" {
			updated = existing + "\n" + block + "\n"
		} else {
			updated = block + "\n"
		}
		status = "CLAUDE.md updated"
	}

	if err := os.WriteFile("CLAUDE.md", []byte(updated), 0644); err != nil {
		return "", err
	}
	return status, nil
}

func installSkills() ([]string, error) {
	skillsDir := filepath.Join(".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create skills directory: %w", err)
	}

	var statuses []string
	for _, s := range bundledSkills {
		destPath := filepath.Join(skillsDir, s.name+".md")
		bundledVersion := parseSkillVersion(s.content)

		existing, err := os.ReadFile(destPath)
		if err == nil {
			installedVersion := parseSkillVersion(string(existing))
			if installedVersion >= bundledVersion {
				statuses = append(statuses, fmt.Sprintf("↷ skipped %s (already v%d)", s.name, installedVersion))
				continue
			}
			if err := os.WriteFile(destPath, []byte(s.content), 0644); err != nil {
				return nil, fmt.Errorf("failed to write skill %s: %w", s.name, err)
			}
			statuses = append(statuses, fmt.Sprintf("✓ updated %s (v%d → v%d)", s.name, installedVersion, bundledVersion))
			continue
		}

		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read skill %s: %w", s.name, err)
		}

		if err := os.WriteFile(destPath, []byte(s.content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write skill %s: %w", s.name, err)
		}
		statuses = append(statuses, fmt.Sprintf("✓ installed %s (v%d)", s.name, bundledVersion))
	}
	return statuses, nil
}

func parseSkillVersion(content string) int {
	m := frontmatterVersionRe.FindStringSubmatch(content)
	if m == nil {
		return 0
	}
	v, _ := strconv.Atoi(m[1])
	return v
}
