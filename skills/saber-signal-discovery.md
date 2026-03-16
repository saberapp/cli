---
name: saber-signal-discovery
description: >
  Define buying signals that match your ICP — start here before creating signals or building lists.
version: 4
---

# Saber Signal Discovery

Use this skill to help the user define what buying intent looks like for their Ideal Customer Profile (ICP) before activating signal tracking.

## Goal

Produce a clear ICP definition and a set of approved signal definitions in conversation context, ready to be activated via the Saber CLI.

## Workflow

### Step 1 — Load organisation context

Before asking anything, run: `saber org get`

- If the profile contains name, website, and description fields → use them directly, skip asking.
- If the profile is empty or missing key fields → ask the user to fill in the gaps, then run
  `saber org update --name "..." --general "..." --products "..." --use-cases "..." --value-prop "..."`
  to persist the context for future sessions.

Do not ask questions that are already answered by the org profile.

Once org context is established, you still need ICP specifics — ask these in order:
1. **Who do you sell to?** (industry, company size, geography, buyer title) — required to define the ICP
2. **Who is the buyer?** (title, seniority, department — e.g. "VP Sales", "Head of RevOps")
3. **What makes a company ready to buy?** (triggers — e.g. new sales hire, funding round, tech migration)

Keep this conversational — don't dump all questions at once.

### Step 2 — Generate signal hypotheses

Based on the ICP, propose 3–5 signals that indicate buying intent. For each signal:
- **Signal question** — a natural-language question Saber will answer (e.g. "Is this company actively hiring SDRs?")
- **Signal type** — `company` or `contact`
- **Why it matters** — brief rationale linking the signal to purchase readiness

### Step 3 — Confirm with user

Present the proposed signals and ask for feedback. Adjust as needed.

### Step 4 — Hand off

Keep the agreed ICP and signal definitions in conversation context — do not write them to files.

Once signals are approved, tell the user:
- To build a target list first: use `/saber-build-account-list` or `/saber-build-contact-list`
- To activate company signals: use `/saber-create-company-signals`
- To activate contact signals: use `/saber-create-contact-signals`

## Example signals

```
Signal: Is this company actively hiring in sales or revenue roles?
Type: company
Rationale: Hiring in sales indicates growth mode and budget for new tools.

Signal: Has this company recently raised a funding round?
Type: company
Rationale: Post-funding companies accelerate spend on GTM tooling.

Signal: Is the Head of People at this company posting about frontline retention challenges?
Type: contact
Rationale: Public pain signals from the buyer persona indicate active problem awareness.
```
