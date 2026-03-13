---
name: saber-build-account-list
description: >
  Build a target account list using the Saber CLI and run company signals against it.
version: 4
---

# Saber Build Account List

Use this skill to build a list of target accounts in Saber and optionally run company signals against them.

## Prerequisites

- Saber CLI is available (`saber --help` works)
- ICP context is available in the conversation (run `/saber-signal-discovery` first if not)

## Workflow

### Step 1 — Confirm criteria

Summarise the account criteria from conversation context:
- Industry / vertical
- Company size
- Geography
- Any other filters (e.g. tech stack, business model)

Ask the user to confirm or adjust before proceeding.

### Step 2 — Build the list

Ask the user how they want to supply the accounts:

**Option A — Let Saber identify accounts (recommended)**
Use the Saber CLI to create a list with filter criteria:
```bash
saber list company create --name "<list name>" --industry "<industry>" --country "<country code>" --size "<size range>"
```
Saber will populate the list with matching companies from its database.

Filter value formats:
- `--industry` values are **lowercase**, e.g. `restaurants`, `hospitality`, `food & beverages`, `hotels and motels`. Use `&` not `and`.
- `--size` values use K notation: `1-10`, `11-50`, `51-200`, `201-500`, `501-1K`, `1K-5K`, `5K-10K`, `10K+`
- `--country` values are ISO 3166-1 alpha-2 codes, e.g. `GB`, `DE`, `FR`, `NL`

**Option B — Provide domains or company names directly**
Ask the user to paste a list. Then add them to a named list:
```bash
saber list company create --name "<list name>"
# Saber will prompt to add companies interactively, or use --domain flags
```

**Option C — Import from HubSpot**
If the user wants to pull companies from HubSpot using a property filter:
```bash
saber list company import --name "<list name>" --property <property> --operator EQ --value "<value>"
```
Example: `--property industry --operator EQ --value "Software"`

### Step 3 — Review and confirm

Show the user the list summary (name, company count) and ask if they want to adjust before activating signals.

### Step 4 — Run signals (optional)

If approved signals are available in conversation context, offer to run them against the list now using `/saber-create-company-signals`.

## Key commands

```bash
saber list company create --name "<name>" [--industry] [--country] [--size]
saber list company import --name "<name>" --property <property> --operator EQ --value "<value>"
saber list company get <listId>
saber list company list
```
