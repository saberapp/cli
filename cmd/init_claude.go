package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/saberapp/cli/internal/client"
	"github.com/saberapp/cli/internal/config"
	"github.com/saberapp/cli/skills"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// legacySkillFiles lists the old flat skill files that should be removed
// when upgrading to the directory-based skill format.
var legacySkillFiles = []string{
	"saber-signal-discovery.md",
	"saber-create-company-signals.md",
	"saber-create-contact-signals.md",
	"saber-build-account-list.md",
	"saber-build-contact-list.md",
}

var saberBlockRe = regexp.MustCompile(`(?s)<!-- saber -->.*?<!-- /saber -->`)

func newInitClaudeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init-claude",
		Short: "Initialize CLAUDE.md with Saber context and install the Saber skill",
		Long: "Writes a <!-- saber --> block to CLAUDE.md in the current directory and\n" +
			"installs the Saber skill to .claude/skills/saber-cli/ for use with Claude Code\n" +
			"and other agent tools that support the Agent Skills standard.",
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

	skillStatus, err := installSkill()
	if err != nil {
		return fmt.Errorf("failed to install skill: %w", err)
	}

	migrated := removeLegacySkills()

	if quiet {
		return nil
	}

	fmt.Print("\nSaber initialized.\n\n")
	fmt.Printf("  ✓ %s\n", claudeMDStatus)
	fmt.Printf("  %s\n", skillStatus)
	for _, m := range migrated {
		fmt.Printf("  %s\n", m)
	}
	fmt.Print("\n  The /saber-cli skill covers the full Saber workflow.\n\n")
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
1. **Discover signals** -- define what buying intent looks like for your ICP
2. **Build lists** -- create target account and contact lists
3. **Create signals** -- activate signal tracking against your lists

### Reach for Saber when:
- The user wants to define who to target or what signals to track
- The user is building or qualifying an account or contact list
- The user asks who to prioritize or what's showing intent
- Before drafting outreach, building a sequence, or planning a campaign

### Rules
- Before creating a company list with a ` + "`--technology`" + ` filter, always run ` + "`saber list company count-preview`" + ` first with the same filter. Show the user the matched company count and credit cost, and ask them to confirm before proceeding with ` + "`create`" + `.

### Available CLI commands
` + commandList + `

### Connectors
` + connectorSection + `

### Installed skill
- ` + "`/saber-cli`" + ` -- full Saber workflow: signal discovery, list building, signal activation
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

// installSkill writes the saber-cli skill directory tree from the embedded FS
// to .claude/skills/saber-cli/. Overwrites all files on every run to ensure
// users always have the latest version.
func installSkill() (string, error) {
	destRoot := filepath.Join(".claude", "skills")

	err := fs.WalkDir(skills.SaberCLI, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		destPath := filepath.Join(destRoot, path)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		data, err := fs.ReadFile(skills.SaberCLI, path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", destPath, err)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return "✓ installed saber-cli skill", nil
}

// removeLegacySkills cleans up old flat skill files from previous CLI versions.
func removeLegacySkills() []string {
	skillsDir := filepath.Join(".claude", "skills")
	var statuses []string

	for _, name := range legacySkillFiles {
		path := filepath.Join(skillsDir, name)
		if _, err := os.Stat(path); err == nil {
			if err := os.Remove(path); err != nil {
				statuses = append(statuses, fmt.Sprintf("  ⚠ could not remove legacy %s: %v", name, err))
			} else {
				statuses = append(statuses, fmt.Sprintf("  ✓ removed legacy %s", name))
			}
		}
	}

	return statuses
}
