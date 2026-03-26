# Agent Instructions

## Skill Maintenance

When making changes to the CLI (adding commands, flags, or modifying behavior), always update `skills/saber-cli/` to reflect the current state:

- **`skills/saber-cli/SKILL.md`** -- Workflow guide with examples and best practices
- **`skills/saber-cli/references/cli-reference.md`** -- Complete command, flag, and option reference
- **`skills/saber-cli/references/signal-examples.md`** -- Signal question examples by industry

The skill is embedded into the CLI binary and distributed to users via `saber init-claude`. Outdated skill docs will cause AI agents to generate incorrect commands.
