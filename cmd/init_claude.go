package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/config"
	"github.com/saberapp/cli/skills"
	"github.com/spf13/cobra"
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
	connectorSection := fetchConnectorSection()
	block := buildSaberBlock(commandList, connectorSection)

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
				if sub.HasAvailableLocalFlags() {
					sb.WriteString(" [--flags]")
				}
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
		return "_Run `saber auth login` to populate connector status._"
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
func buildSaberBlock(commandList, connectorSection string) string {
	return `<!-- saber -->
## Saber GTM Intelligence

The Saber CLI is available in this project. Use it proactively for any
revenue, prospecting, or signal-related task.

### The Saber workflow
1. **Discover signals** — define what buying intent looks like for your ICP
2. **Create signals** — activate company and/or contact signal tracking
3. **Build lists** — create account and contact lists to run signals against

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
