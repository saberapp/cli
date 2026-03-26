---
name: saber-cli
description: >
  GTM intelligence via the Saber CLI. Use when the user wants to define their ICP
  and buying signals, build target account or contact lists, run company or contact
  signal research, manage signal subscriptions, or prioritize outreach. Also use
  when the user mentions prospecting, lead scoring, intent data, buying intent,
  firmographic filters, or anything related to sales intelligence. Triggers include
  "find companies", "who should I target", "track hiring signals", "build a list",
  "run signals", "check credits", or any GTM workflow involving the saber command.
---

# Saber CLI

The Saber CLI runs company and contact signal research, builds target lists, and
manages signal subscriptions from the terminal. It talks to the Saber API,
authenticated with an API key stored at `~/.saber/credentials.json`.

## Core Workflow

Every Saber engagement follows three phases:

1. **Discover** -- define ICP and buying signals
2. **Build** -- create target account and contact lists
3. **Activate** -- run signals against those lists

```bash
# Phase 1: Understand the org
saber org get

# Phase 2: Build a company list
saber list company create --name "Mid-market SaaS" --industry "software development" --size "51-200" --country US

# Phase 3: Run a signal across the list
saber subscription create --list <listId> --name "Hiring in sales" \
  --question "Is this company actively hiring in sales or revenue roles?" \
  --answer-type boolean --frequency monthly --run-once
```

## Phase 1 -- Signal Discovery

Help the user define what buying intent looks like for their Ideal Customer Profile
before activating any signal tracking.

### Step 1 -- Load organisation context

Run `saber org get` first. If the profile has name, website, and description fields,
use them directly. If the profile is empty, ask the user to fill the gaps, then persist:

```bash
saber org update --name "Acme Corp" --general "B2B sales intelligence platform" \
  --products "Signal tracking, list building" \
  --use-cases "Outbound prospecting, ABM campaigns" \
  --value-prop "Find buying intent before your competitors do"
```

### Step 2 -- Gather ICP details

Once org context is established, ask these in order (conversationally, not all at once):

1. **Who do you sell to?** -- industry, company size, geography, buyer title
2. **Who is the buyer?** -- title, seniority, department (e.g. "VP Sales", "Head of RevOps")
3. **What makes a company ready to buy?** -- triggers (new hire, funding, tech migration)

### Step 3 -- Generate signal hypotheses

Propose 3-5 signals based on the ICP. For each:
- **Signal question** -- natural language question Saber will answer
- **Signal type** -- `company` or `contact`
- **Why it matters** -- brief rationale linking the signal to purchase readiness

Example signals:

```
Signal: Is this company actively hiring in sales or revenue roles?
Type: company
Rationale: Hiring in sales indicates growth mode and budget for new tools.

Signal: Has this company recently raised a funding round?
Type: company
Rationale: Post-funding companies accelerate spend on GTM tooling.

Signal: Is the Head of People posting about frontline retention challenges?
Type: contact
Rationale: Public pain signals from the buyer persona indicate active problem awareness.
```

### Step 4 -- Confirm and hand off

Present proposed signals, adjust based on feedback, then keep the agreed ICP and signal
definitions in conversation context. Next steps: build lists (Phase 2), then activate
signals (Phase 3).

## Phase 2 -- Build Target Lists

### Company lists

Ask how the user wants to supply accounts:

**Option A -- Let Saber find companies (recommended)**

```bash
saber list company create --name "US Mid-market SaaS" \
  --industry "software development" --size "51-200" --country US
```

Filter formats:
- `--industry` values are lowercase: `restaurants`, `hospitality`, `food & beverages`
- `--size` uses K notation: `1-10`, `11-50`, `51-200`, `201-500`, `501-1K`, `1K-5K`, `5K-10K`, `10K+`
- `--country` uses ISO 3166-1 alpha-2: `US`, `GB`, `DE`, `NL`
- `--technology` uses slugs: `stripe`, `hubspot`, `salesforce`

**Technology filters consume credits per matched company.** Always run `count-preview`
first so the user can see the cost before committing:

```bash
saber list company count-preview --technology "salesforce" --size "51-200" --country US
# Shows: X companies matched, Y credits required
# Ask user to confirm before running create
```

**Option B -- Import from HubSpot**

```bash
saber list company import --name "From HubSpot" \
  --property industry --operator EQ --value "Technology"
```

**Preview without creating:**

```bash
saber company search --industry "software development" --size "51-200" --country US
```

### Contact lists

Contacts are sourced from company LinkedIn URLs. If a company list exists, get URLs
from it first:

```bash
saber list company companies <listId> --limit 10
```

Then create the contact list:

```bash
saber list contact create --name "VP Sales at target accounts" \
  --company-linkedin https://linkedin.com/company/acme \
  --company-linkedin https://linkedin.com/company/globex \
  --title "VP Sales" --title "Head of Sales"
```

Both `--company-linkedin` and `--title` are repeatable. `--company-linkedin` is required.

**Preview without creating:**

```bash
saber contact search --company-linkedin https://linkedin.com/company/acme --title "VP Sales"
```

## Phase 3 -- Activate Signal Tracking

### Before running any signals

Before any `saber signal` or `saber subscription` command that consumes credits:
1. Run `saber credits` and show the current balance
2. State how many signals will run (companies x questions, or contacts x questions)
3. Ask the user to confirm before proceeding

### Signal templates

Create reusable templates for frequently used signal questions. Templates store the
question, answer type, weight, and qualification criteria so they can be referenced
by ID instead of repeating the full definition.

```bash
# Create a template
saber template create --name "CRM Detection" --question "Which CRM are they using?" \
  --answer-type list --weight important

# List templates
saber template list

# Get template details
saber template get <templateId>

# Update a template (creates a new version)
saber template update <templateId> --name "CRM Detection v2" --weight nice_to_have

# Delete a template (soft-delete)
saber template delete <templateId>
```

### Company signals -- spot check

Run a question against a single domain:

```bash
saber signal --domain acme.com --question "Are they hiring engineers?" --answer-type boolean
```

Use a template instead of inline question:

```bash
saber signal --domain acme.com --template <templateId>
```

Use lenient verification mode for broader answers:

```bash
saber signal --domain acme.com --question "What is their tech stack?" --verification-mode lenient
```

Fire multiple in parallel with `--no-wait`:

```bash
saber signal --domain acme.com --question "Are they hiring?" --no-wait
saber signal --domain globex.com --question "Are they hiring?" --no-wait
# Collect signal IDs, then retrieve:
saber signal get <signalId>
```

### Company signals -- batch

Run multiple questions across multiple domains in one call. Creates a Cartesian
product (each question x each domain).

```bash
# 2 questions x 3 domains = 6 signals
saber signal batch \
  --domain acme.com --domain globex.com --domain initech.com \
  --question "Are they hiring engineers?" \
  --question "What CRM do they use?" \
  --answer-type boolean

# Auto-generate summaries when all signals complete
saber signal batch --domain acme.com --domain globex.com \
  --question "Revenue?" --question "Headcount?" --generate-summary

# Async mode for large batches (up to 20,000 signals)
saber signal batch --domain acme.com --question "..." --async
```

### Company signals -- listing and filtering

Browse and filter existing signal results:

```bash
saber signal list                                          # all signals
saber signal list --domain acme.com --status completed     # filter by domain and status
saber signal list --from-date 2024-01-01T00:00:00Z         # filter by date
saber signal list --subscription-id <id>                   # filter by subscription
saber signal list --limit 10 --offset 20                   # pagination
```

### Signal summaries

Generate AI-powered summaries that consolidate insights from all completed signals
for a domain into structured data points with qualifications and sources.

```bash
# Generate a new summary
saber summary generate --domain acme.com

# List historical summaries for a domain
saber summary list --domain acme.com
```

### Company signals -- full list (subscriptions)

For running a signal across all companies in a list, use subscriptions.

**One-off run (recommended for getting started):**

```bash
saber subscription create --list <listId> \
  --name "Hiring signal" \
  --question "Is this company actively hiring in sales or revenue roles?" \
  --answer-type boolean \
  --frequency monthly --run-once
```

`--run-once` triggers immediately, then stops the schedule. One subscription per signal question.

**Recurring schedule:**

```bash
saber subscription create --list <listId> \
  --name "Hiring signal" \
  --question "Is this company actively hiring in sales or revenue roles?" \
  --answer-type boolean --frequency weekly
```

Manage subscriptions:

```bash
saber subscription list
saber subscription get <subscriptionId>
saber subscription trigger <subscriptionId>    # run immediately
saber subscription start <subscriptionId>      # activate schedule
saber subscription stop <subscriptionId>       # pause schedule
```

### Contact signals

Run a question against a LinkedIn profile:

```bash
saber signal --profile https://linkedin.com/in/janedoe \
  --question "Is this person posting about employee retention challenges?"
```

Use a template:

```bash
saber signal --profile https://linkedin.com/in/janedoe --template <templateId>
```

Fire multiple with `--no-wait`:

```bash
saber signal --profile <linkedin-url> --question "<question>" --no-wait
saber signal get <signalId>
```

### Sample run (recommended before full activation)

Before running signals across a full list, test a sample of 5-10 items first so the
user can validate quality without spending credits on everything.

For company signals:

```bash
saber list company companies <listId> --limit 10
# For each domain, fire signals with --no-wait
saber signal --domain <domain> --question "<question>" --answer-type boolean --no-wait
# Collect IDs, then retrieve
saber signal get <signalId>
```

Or use batch for a quicker multi-domain sample:

```bash
saber signal batch --domain acme.com --domain globex.com --domain initech.com \
  --question "Are they hiring?" --answer-type boolean
```

Present results and ask:
- Do the signals look accurate?
- Any false positives or unexpected answers?
- Proceed with the full list?

For contact signals, pick 3-5 contacts the user already knows about (so they can judge
accuracy), run signals, review, then confirm.

## Signal Flags

| Flag | Default | Description |
|---|---|---|
| `--domain` / `-d` | -- | Company domain (company signals) |
| `--profile` / `-p` | -- | LinkedIn profile URL (contact signals) |
| `--question` / `-q` | required* | Research question (max 500 chars) |
| `--template` | -- | Signal template ID (alternative to `--question`) |
| `--answer-type` / `-a` | `open_text` | `open_text`, `boolean`, `number`, `list`, `percentage`, `currency`, `url` |
| `--verification-mode` | `strict` | `strict` or `lenient` |
| `--force-refresh` | false | Bypass 12h result cache |
| `--no-wait` | false | Return signal ID immediately without polling |
| `--webhook` | -- | POST result to this URL when complete |
| `--max-wait` | `120` | Max seconds to wait before timing out |

*Required unless `--template` is provided.

## Common Patterns

### End-to-end ICP activation

```bash
# 1. Set up org context
saber org update --name "Acme" --general "We sell HR software to mid-market"

# 2. Build target list
saber list company create --name "Mid-market US" \
  --industry "software development" --size "51-200" --country US

# 3. Sample test (10 companies)
saber list company companies <listId> --limit 10
saber signal --domain <domain> --question "Are they hiring in HR?" --no-wait

# 4. Full activation
saber subscription create --list <listId> --name "HR hiring" \
  --question "Is this company actively hiring in HR roles?" \
  --answer-type boolean --frequency monthly --run-once

# 5. Review results
saber subscription get <subscriptionId>
```

### Multi-signal activation

Create one subscription per signal question against the same list:

```bash
saber subscription create --list <listId> --name "Hiring signal" \
  --question "Is this company actively hiring?" --answer-type boolean --frequency monthly --run-once
saber subscription create --list <listId> --name "Funding signal" \
  --question "Has this company raised funding recently?" --answer-type boolean --frequency monthly --run-once
saber subscription create --list <listId> --name "Tech migration" \
  --question "Is this company migrating their tech stack?" --answer-type boolean --frequency monthly --run-once
```

Or use batch for a quick multi-question run across specific domains:

```bash
saber signal batch \
  --domain acme.com --domain globex.com \
  --question "Is this company actively hiring?" \
  --question "Has this company raised funding recently?" \
  --question "Is this company migrating their tech stack?" \
  --answer-type boolean --generate-summary
```

### Template-based workflow

Create templates once, reuse across signals and subscriptions:

```bash
# Create templates
saber template create --name "Hiring" --question "Is this company actively hiring?" --answer-type boolean
saber template create --name "Funding" --question "Has this company raised funding recently?" --answer-type boolean

# Run signals using templates
saber signal --domain acme.com --template <templateId>

# Use templates in subscriptions
saber subscription create --list <listId> --template <templateId> --frequency weekly
```

### Check account status

```bash
saber credits                    # remaining credit balance
saber connectors                 # configured integrations and status
saber auth status                # current auth state
```

## Global Flags

Available on every command:

| Flag | Description |
|---|---|
| `--json` | Output raw API JSON to stdout |
| `--quiet` / `-Q` | Suppress all non-error output |
| `--verbose` / `-v` | Log HTTP method, URL, status, rate-limit headers to stderr |
| `--api-url` | Override base API URL (default: `https://api.saber.app`) |

## Deep-Dive Documentation

| Reference | When to Use |
|---|---|
| [references/cli-reference.md](references/cli-reference.md) | Full command reference with all flags and options |
| [references/signal-examples.md](references/signal-examples.md) | Signal question examples by industry and use case |
