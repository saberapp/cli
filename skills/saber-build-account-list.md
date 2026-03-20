---
name: saber-build-account-list
description: >
  Build a target account list using the Saber CLI and run company signals against it.
version: 6
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
- `--technology` values are technology slugs, e.g. `stripe`, `hubspot`, `salesforce`

**Important: if the filter includes `--technology`, always run `count-preview` first** before creating the list. This shows the user how many companies match and how many Saber credits it will cost — without charging anything:
```bash
saber list company count-preview --technology "<slug>" [--industry] [--country] [--size]
```
Present the result to the user ("X companies match, Y credits required") and ask them to confirm before running `create`. Skip this step only if the user has explicitly already confirmed the cost.

**Option B — Provide domains or company names directly**
If the user has a list of domains or company names, create an empty named list and ask them to add companies via the Saber dashboard:
```bash
saber list company create --name "<list name>"
```

**Option C — Import from HubSpot**
If the user wants to pull companies from HubSpot using a property filter:
```bash
saber list company import --name "<list name>" --property <property> --operator EQ --value "<value>"
```
Example: `--property industry --operator EQ --value "Software"`

### Step 3 — Review and confirm

Show the user the list summary (name, company count) and ask if they want to adjust before activating signals.

### Step 4 — Sample run (optional but recommended)

Before running signals across the whole list, offer to test a sample of the first 10 companies so the user can validate signal quality without spending credits on the full list.

Get the first 10 domains from the list:
```bash
saber list company companies <listId> --limit 10
```

For each signal question, fire signals with `--no-wait` so they all run in parallel:
```bash
saber signal --domain <domain> --question "<question>" --answer-type boolean --no-wait
# repeat for each domain — collect the signal IDs printed
```

Then retrieve results:
```bash
saber signal get <signalId>
# repeat for each signal ID
```

Present the 10 results to the user and ask:
- Do the signals look accurate?
- Are there any false positives or unexpected answers?
- Do they want to proceed with the full list?

Only proceed to Step 5 once the user confirms the sample looks good.

### Step 5 — Run signals on the full list (optional)

If approved signals are available in conversation context, offer to run them against the full list now using `/saber-create-company-signals`.

## Key commands

```bash
saber list company count-preview [--technology] [--industry] [--country] [--size]
saber list company create --name "<name>" [--industry] [--country] [--size] [--technology]
saber list company import --name "<name>" --property <property> --operator EQ --value "<value>"
saber list company get <listId>
saber list company list
```
