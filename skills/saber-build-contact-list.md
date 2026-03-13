---
name: saber-build-contact-list
description: >
  Build a target contact list using the Saber CLI and run contact signals against it.
version: 4
---

# Saber Build Contact List

Use this skill to build a list of target contacts in Saber and optionally run contact-level signals against them.

## Prerequisites

- Saber CLI is available (`saber --help` works)
- ICP context is available in the conversation (run `/saber-signal-discovery` first if not)
- A target account list ideally exists in Saber (contacts are sourced from target accounts)

## Workflow

### Step 1 — Confirm buyer persona

Summarise the contact criteria from conversation context:
- Job title(s) / seniority
- Department
- Source (which account list to pull from, if applicable)

Ask the user to confirm or adjust before proceeding.

### Step 2 — Build the contact list

Ask the user how they want to supply contacts:

**Option A — Let Saber identify contacts (recommended)**
Use the Saber CLI to create a list filtered by title, sourced from company LinkedIn URLs.
If a target account list exists, get the company LinkedIn URLs from it first (`saber list company companies <listId>`), then:
```bash
saber list contact create --name "<list name>" \
  --company-linkedin <linkedin-url> \
  --company-linkedin <linkedin-url> \
  --title "<title>"
```
`--company-linkedin` and `--title` are both repeatable. `--company-linkedin` is required.

**Option B — Provide contacts directly**
If the user has specific contacts in mind, ask for their LinkedIn URLs or emails and add them manually via the Saber dashboard.

### Step 3 — Review and confirm

Show the list summary and ask the user to confirm before activating signals.

### Step 4 — Sample run (optional but recommended)

Before running signals across the whole list, offer to test a small sample first so the user can validate signal quality without spending credits on every contact.

Ask the user to pick 3–5 contacts to test — ideally ones they already know something about, so they can judge whether the signal results are accurate. Get their LinkedIn profile URLs from the list in the Saber dashboard or from conversation context.

For each contact and each signal question:
```bash
saber signal --profile <linkedin-url> --question "<question>" --answer-type boolean --no-wait
# collect signal IDs
```

Retrieve and review results:
```bash
saber signal get <signalId>
```

Present results and ask:
- Do the signals look accurate for these contacts?
- Any false positives or unexpected answers?
- Do they want to proceed with the rest of the list?

Only continue to Step 5 once the user confirms the sample looks good.

### Step 5 — Run signals on the full list (optional)

If approved contact signals are available in conversation context, offer to run them using `/saber-create-contact-signals`.

### Step 6 — Prioritize

After signals complete, present results ranked by intent so the user can sequence outreach.

## Key commands

```bash
saber list contact create --name "<name>" --company-linkedin <url> [--title] [--keyword] [--country]
saber list contact get <listId>
saber list contact list
```
