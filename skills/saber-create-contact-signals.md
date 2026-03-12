---
name: saber-create-contact-signals
description: >
  Activate contact-level signal tracking using the Saber CLI — creates signals for contacts in a contact list.
version: 2
---

# Saber Create Contact Signals

Use this skill to activate contact-level signal tracking against a list of target prospects in Saber.

## Prerequisites

- Approved contact signal definitions are available in conversation context (run `/saber-signal-discovery` first if not)
- A target contact list exists in Saber (run `/saber-build-contact-list` first if not)
- Saber CLI is available (`saber --help` works)

## Workflow

### Step 1 — Confirm signals and list

From conversation context, confirm:
- Which contact signals are approved (the questions to run)
- Which contact list to run them against (get the list ID if needed: `saber list contact list`)

### Step 2 — Activate signals

Run each approved contact signal against the contact list:

```bash
saber signal contact create -q "<signal question>" --list <listId>
```

Each call creates a signal subscription that runs the question against every contact in the list.

To check available lists:
```bash
saber list contact list
```

### Step 3 — Check signal status

```bash
saber signal contact get <signalId>
```

Signals may be `pending`, `processing`, or `complete`.

### Step 4 — Review and prioritize

```bash
saber signal contact results <signalId>
```

Present results ranked by signal strength. Contacts where the signal fired positively should be sequenced first for outreach.

## Key commands

```bash
saber signal contact create -q "<question>" --list <listId>
saber signal contact get <signalId>
saber signal contact results <signalId>
saber list contact list
```
