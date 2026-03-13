---
name: saber-create-contact-signals
description: >
  Activate contact-level signal tracking using the Saber CLI — creates signals for contacts in a contact list.
version: 4
---

# Saber Create Contact Signals

Use this skill to run contact-level signal research using the Saber CLI.

## Prerequisites

- Approved contact signal definitions are available in conversation context (run `/saber-signal-discovery` first if not)
- Saber CLI is available (`saber --help` works)

## Before running signals

Before executing any `saber signal` or `saber subscription` command that will consume credits:
1. Run `saber credits` and show the user their current balance.
2. Tell the user how many signals will be run (number of contacts × number of questions).
3. Ask the user to confirm they want to proceed before issuing the command.

## Two modes

### Mode A — Spot-check a specific contact

Use `saber signal` with a LinkedIn profile URL to run a question against a specific contact:

```bash
saber signal --profile <linkedin-url> --question "<signal question>"
```

This is synchronous by default — it waits for the result and prints it.

```bash
# Example
saber signal --profile https://linkedin.com/in/janedoe --question "Is this person posting about employee retention challenges?"
```

Use `--no-wait` to fire multiple signals without waiting:

```bash
saber signal --profile <linkedin-url> --question "<question>" --no-wait
# Returns a signal ID; retrieve result later with:
saber signal get <signalId>
```

### Mode B — Run signals across a large contact list (signal subscriptions)

For running signals across a large contact list, use signal subscriptions in the Saber dashboard or via the API. The CLI currently supports per-contact signals only.

## Workflow (spot-check mode)

### Step 1 — Pick contacts to check

Ask the user which contacts they want to prioritise. You'll need LinkedIn profile URLs.

### Step 2 — Run signals

For each contact and each approved signal question:

```bash
saber signal --profile <linkedin-url> --question "<question>" --answer-type boolean
```

### Step 3 — Review and prioritise

Present results to the user. Contacts where the signal fired positively should be sequenced first for outreach.

## Key commands

```bash
saber signal --profile <linkedin-url> --question "<question>" [--answer-type] [--no-wait]
saber signal get <signalId>
```
