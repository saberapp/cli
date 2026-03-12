---
name: saber-create-company-signals
description: >
  Activate company-level signal tracking using the Saber CLI — creates signals for domains in a company list.
version: 2
---

# Saber Create Company Signals

Use this skill to activate company-level signal tracking against a list of target accounts in Saber.

## Prerequisites

- Approved company signal definitions are available in conversation context (run `/saber-signal-discovery` first if not)
- A target account list exists in Saber (run `/saber-build-account-list` first if not)
- Saber CLI is available (`saber --help` works)

## Workflow

### Step 1 — Confirm signals and list

From conversation context, confirm:
- Which company signals are approved (the questions to run)
- Which account list to run them against (get the list ID if needed: `saber list company list`)

### Step 2 — Activate signals

Run each approved company signal against the account list:

```bash
saber signal company create -q "<signal question>" --list <listId>
```

Each call creates a signal subscription that runs the question against every company in the list.

To check available lists:
```bash
saber list company list
```

### Step 3 — Check signal status

```bash
saber signal company get <signalId>
```

Signals may be `pending`, `processing`, or `complete`. Once complete, results are available per company.

### Step 4 — Review results

```bash
saber signal company results <signalId>
```

Present the results to the user, highlighting companies where the signal fired positively — these are your highest-priority accounts.

## Key commands

```bash
saber signal company create -q "<question>" --list <listId>
saber signal company get <signalId>
saber signal company results <signalId>
saber list company list
```
