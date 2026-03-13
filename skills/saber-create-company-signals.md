---
name: saber-create-company-signals
description: >
  Activate company-level signal tracking using the Saber CLI — creates signals for domains in a company list.
version: 4
---

# Saber Create Company Signals

Use this skill to run company-level signal research using the Saber CLI.

## Prerequisites

- Approved company signal definitions are available in conversation context (run `/saber-signal-discovery` first if not)
- Saber CLI is available (`saber --help` works)

## Two modes

### Mode A — Run signals against a full list (signal subscriptions)

For running a signal across all companies in a list, use signal subscriptions.

**Option 1 — Run once (recommended for getting started)**

Use `--run-once` to trigger immediately and stop the schedule automatically:
```bash
saber subscription create \
  --list <listId> \
  --name "<signal name>" \
  --question "<signal question>" \
  --answer-type boolean \
  --frequency monthly \
  --run-once
```

**Option 2 — Recurring schedule**

Omit `--run-once` to keep the subscription active on a schedule:
```bash
saber subscription create \
  --list <listId> \
  --name "<signal name>" \
  --question "<signal question>" \
  --answer-type boolean \
  --frequency weekly
```

Then trigger manually when needed:
```bash
saber subscription trigger <subscriptionId>
```

Create one subscription per signal question.

### Mode B — Spot-check a specific company

Use `saber signal` to run a question against a single domain:

```bash
saber signal --domain <domain> --question "<signal question>" --answer-type boolean
```

Use `--no-wait` to fire multiple signals without waiting for each result:
```bash
saber signal --domain <domain> --question "<question>" --no-wait
saber signal get <signalId>
```

## Workflow (list mode)

### Step 1 — Confirm signals and list ID

From conversation context, confirm:
- The approved signal questions to activate
- The account list to run them against

To find a list ID:
```bash
saber list company list
```

### Step 2 — Create and trigger subscriptions

Create one subscription per signal question using `--run-once` to trigger immediately:
```bash
saber subscription create --list <listId> --name "<name>" --question "<question>" --answer-type boolean --frequency monthly --run-once
```

This creates the subscription, runs it immediately, and stops the schedule automatically.

### Step 3 — Review results

```bash
saber subscription get <subscriptionId>
```

When complete, present results to the user. Companies where signals fired positively are highest priority for outreach.

## Key commands

```bash
saber subscription create --list <listId> --name "<name>" --question "<question>" [--answer-type] [--frequency daily|weekly|monthly] [--cron] [--timezone]
saber subscription trigger <subscriptionId>
saber subscription get <subscriptionId>
saber subscription list
saber subscription start <subscriptionId>
saber subscription stop <subscriptionId>
saber signal --domain <domain> --question "<question>" [--answer-type] [--no-wait]
saber list company list
