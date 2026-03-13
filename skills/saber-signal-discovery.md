---
name: saber-signal-discovery
description: >
  Define buying signals that match your ICP — start here before creating signals or building lists.
version: 3
---

# Saber Signal Discovery

Use this skill to help the user define what buying intent looks like for their Ideal Customer Profile (ICP) before activating signal tracking.

## Goal

Produce a clear ICP definition and a set of approved signal definitions in conversation context, ready to be activated via the Saber CLI.

## Workflow

### Step 1 — Establish context

Ask two questions up front:
1. **Where do you work, and what do you sell?** (company name, what the product does)
2. **Who do you sell to?** (the ICP — industry, company size, geography, buyer title)

Once you have those answers, ask follow-up questions to fill in any gaps:
- **Who is the buyer?** (title, seniority, department — e.g. "VP Sales", "Head of RevOps")
- **What makes a company ready to buy?** (triggers — e.g. new sales hire, funding round, tech migration)

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
