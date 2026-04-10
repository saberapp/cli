# CLI Command Reference

Complete reference for every `saber` command, flag, and option.

## Authentication

```bash
saber auth login                       # Interactive prompt for API key (masked input)
saber auth login --key sk_live_...     # Non-interactive (CI-friendly)
saber auth logout                      # Remove stored API key
saber auth status                      # Show current auth state
```

API keys are validated against the API before being saved.
Key format: `sk_live_` + 43 characters (51 chars total).

Credential lookup order:
1. `SABER_API_KEY` environment variable (CI-friendly, never written to disk)
2. `~/.saber/credentials.json` (written by `saber auth login`)

Get an API key at [ai.saber.app > Settings > API Keys](https://ai.saber.app).

## Signals

### Create a company signal

```bash
saber signal --domain <domain> --question "<question>" [flags]
saber signal --domain <domain> --template <templateId> [flags]
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--domain` | `-d` | required | Company domain (e.g. `acme.com`) |
| `--question` | `-q` | required* | Research question (max 500 chars) |
| `--template` | | | Signal template ID (alternative to `--question`) |
| `--answer-type` | `-a` | `open_text` | Response format (see answer types below) |
| `--verification-mode` | | `strict` | `strict` or `lenient` |
| `--force-refresh` | | `false` | Bypass the 12-hour answer cache |
| `--no-wait` | | `false` | Return signal ID immediately without polling |
| `--webhook` | | | POST result to this URL when complete |
| `--max-wait` | | `120` | Max seconds to wait before timing out |

*Required unless `--template` is provided.

### Create a contact signal

```bash
saber signal --profile <linkedin-url> --question "<question>" [flags]
saber signal --profile <linkedin-url> --template <templateId> [flags]
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--profile` | `-p` | required | Contact LinkedIn profile URL |
| `--question` | `-q` | required* | Research question (max 500 chars) |
| `--template` | | | Signal template ID (alternative to `--question`) |
| `--answer-type` | `-a` | `open_text` | Response format (see answer types below) |
| `--verification-mode` | | `strict` | `strict` or `lenient` |
| `--force-refresh` | | `false` | Bypass the 12-hour answer cache |
| `--no-wait` | | `false` | Return signal ID immediately without polling |
| `--webhook` | | | POST result to this URL when complete |
| `--max-wait` | | `120` | Max seconds to wait before timing out |

*Required unless `--template` is provided.

### Retrieve a signal result

```bash
saber signal get <signalId>
```

### List signals

```bash
saber signal list [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--domain` / `-d` | | Filter by company domain |
| `--company-id` | | Filter by company ID |
| `--status` | | Filter by status: `processing`, `completed`, `failed` (repeatable) |
| `--from-date` | | Signals completed after this date (RFC3339) |
| `--to-date` | | Signals completed before this date (RFC3339) |
| `--subscription-id` | | Filter by subscription ID |
| `--limit` | `25` | Max results |
| `--offset` | `0` | Pagination offset |

### Batch signals

Create signals using a Cartesian product of questions x domains.

```bash
saber signal batch --domain <d> --question <q> [flags]
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--domain` | `-d` | required | Company domain (repeatable) |
| `--question` | `-q` | required | Research question (repeatable) |
| `--answer-type` | `-a` | `open_text` | Answer type for all questions |
| `--async` | | `false` | Async mode for large batches (up to 20,000 signals) |
| `--generate-summary` | | `false` | Auto-generate summaries when all signals complete |
| `--webhook` | | | Webhook URL for each signal |
| `--force-refresh` | | `false` | Bypass the 12-hour answer cache |

### Answer types

| Type | Description | Example question |
|---|---|---|
| `open_text` | Free-form text answer (default) | "What is their tech stack?" |
| `boolean` | Yes/No answer | "Are they hiring engineers?" |
| `number` | Numeric answer | "What is their headcount?" |
| `list` | List of items | "What products do they sell?" |
| `percentage` | Percentage value | "What is their YoY growth rate?" |
| `currency` | Monetary amount | "What is their ARR?" |
| `url` | URL answer | "Where is their careers page?" |

## Signal Templates

Reusable templates that define signal questions, answer types, weights, and
qualification criteria.

### Create a template

```bash
saber template create --name "<name>" --question "<question>" [flags]
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--name` | | required | Template name |
| `--question` | `-q` | required | Research question |
| `--description` | | | Optional description |
| `--answer-type` | `-a` | `open_text` | Answer type |
| `--weight` | | | `important`, `nice_to_have`, `not_important` |

### Manage templates

```bash
saber template list [--limit 20] [--offset 0] [--include-deleted]
saber template get <templateId>
saber template update <templateId> [flags]    # PATCH semantics, creates new version
saber template delete <templateId>            # Soft-delete
```

Update flags: `--name`, `--question`, `--description`, `--answer-type`, `--weight`.
At least one flag is required.

## Signal Summaries

AI-powered summaries that consolidate insights from completed signals into
structured data points with qualifications and sources.

### Generate a summary

```bash
saber summary generate --domain <domain>
```

### List summaries

```bash
saber summary list --domain <domain> [--limit 20] [--offset 0]
```

## Company Lists

### Create a company list

```bash
saber list company create --name "<name>" [filters]
```

| Flag | Description | Repeatable |
|---|---|---|
| `--name` | List name (required) | No |
| `--industry` | Industry filter (lowercase, e.g. `software development`) | Yes |
| `--size` | Employee size range (e.g. `51-200`) | Yes |
| `--country` | ISO 3166-1 alpha-2 country code (e.g. `US`, `GB`) | Yes |
| `--technology` | Technology slug (e.g. `stripe`, `hubspot`) | Yes |

**Size range values:** `1-10`, `11-50`, `51-200`, `201-500`, `501-1K`, `1K-5K`, `5K-10K`, `10K+`

### Preview matched companies (free)

```bash
saber list company count-preview [filters]
```

Same filters as `create`. Shows matched company count and credit cost without
charging. Always run this before creating a list with `--technology` filters.

### Search companies (preview without creating)

```bash
saber company search [filters]
```

Same filters as `create`. Returns a preview of matching companies without creating
a list.

### Import from HubSpot

```bash
saber list company import --name "<name>" --property <prop> --operator <op> --value "<val>"
```

| Flag | Default | Description |
|---|---|---|
| `--name` | required | List name |
| `--property` | required | HubSpot property to filter on (e.g. `industry`) |
| `--operator` | `EQ` | `EQ`, `NEQ`, `GT`, `GTE`, `LT`, `LTE`, `HAS_PROPERTY`, `CONTAINS_TOKEN` |
| `--value` | | Filter value |

### Manage company lists

```bash
saber list company list [--limit 50] [--offset 0]    # List all company lists
saber list company get <listId>                       # Get a company list by ID
saber list company update <listId> [filters]          # Update a list (at least one flag required)
saber list company delete <listId>                    # Delete a list
saber list company companies <listId>                 # List companies in a list
```

## Contact Lists

### Create a contact list

```bash
saber list contact create --name "<name>" --company-linkedin <url> [flags]
```

| Flag | Description | Repeatable |
|---|---|---|
| `--name` | List name (required) | No |
| `--company-linkedin` | Company LinkedIn URL (required) | Yes |
| `--title` | Job title filter | Yes |
| `--keyword` | Keyword filter | No |
| `--country` | Country code filter | Yes |

Creating a list runs a live Sales Navigator search and snapshots ~125 contacts.

### Search contacts (preview without creating)

```bash
saber contact search [flags]
```

| Flag | Description | Repeatable |
|---|---|---|
| `--company-linkedin` | Company LinkedIn URL | Yes |
| `--title` | Job title filter | Yes |
| `--keyword` | Keyword filter | No |
| `--country` | Country code filter | Yes |
| `--first-name` | First name filter | No |
| `--last-name` | Last name filter | No |

### Manage contact lists

```bash
saber list contact list [--limit 50] [--offset 0]    # List all contact lists
saber list contact get <listId>                       # Get a contact list by ID
saber list contact update <listId> --name "<name>"    # Rename a contact list
saber list contact delete <listId>                    # Delete a list
saber list contact show <listId>                      # List contacts in a list
```

## Signal Subscriptions

Subscriptions run a signal question against every company in a list on a schedule.

### Create a subscription

```bash
saber subscription create --list <listId> --name "<name>" --question "<question>" [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--list` | required | Company list ID |
| `--name` | required* | Subscription name |
| `--question` | required* | Signal question |
| `--answer-type` | `open_text` | Answer type |
| `--frequency` | required** | `daily`, `weekly`, or `monthly` |
| `--cron` | required** | Custom cron expression (mutually exclusive with `--frequency`) |
| `--timezone` | `UTC` | IANA timezone (e.g. `Europe/Amsterdam`) |
| `--template` | | Signal template ID (alternative to `--name` + `--question`) |
| `--run-once` | `false` | Trigger immediately and stop the schedule |

*Required when not using `--template`. **One of `--frequency` or `--cron` is required.

### Manage subscriptions

```bash
saber subscription list [--limit 50] [--offset 0]    # List subscriptions
saber subscription get <subscriptionId>               # Get subscription details
saber subscription trigger <subscriptionId>           # Run immediately
saber subscription start <subscriptionId>             # Activate the schedule
saber subscription stop <subscriptionId>              # Pause the schedule
```

## Organisation

```bash
saber org get                                         # Show organisation profile
saber org update [flags]                              # Update organisation profile
```

| Flag | Description |
|---|---|
| `--name` | Organisation name |
| `--website` | Organisation website |
| `--general` | General description |
| `--products` | Products description |
| `--use-cases` | Use cases description |
| `--value-prop` | Value proposition description |

At least one flag is required for `update`.

## Market Signals

Market signal subscriptions monitor external data sources (job posts, LinkedIn posts,
fundraising, investments, IPOs) and deliver matching signals to your webhook.

### Create a subscription

```bash
saber market-signal create --type <type> --webhook <url> [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--type` | required | `JOB_POSTS`, `LINKEDIN_POST`, `FUND_RAISED`, `RECENT_INVESTMENT`, `IPO` |
| `--webhook` | required | URL to receive matched signals |
| `--name` | | Display name |
| `--prompt` | | Natural language prompt for AI filter generation (JOB_POSTS only) |
| `--filters` | | Filters as JSON string (schema depends on type) |
| `--webhook-secret` | | Secret for HMAC-SHA256 webhook verification |
| `--interval` | `daily` | Polling interval: `daily` or `weekly` |
| `--signal-limit` | `500` | Max signals per polling interval (1-10000) |

Alias: `saber ms` is shorthand for `saber market-signal`.

### Manage subscriptions

```bash
saber market-signal list [--limit 20] [--offset 0] [--include-deleted]
saber market-signal get <subscriptionId>
saber market-signal update <subscriptionId> [flags]    # PATCH semantics
saber market-signal delete <subscriptionId>            # Soft-delete
saber market-signal pause <subscriptionId>             # Pause polling
saber market-signal resume <subscriptionId>            # Resume polling
saber market-signal trigger <subscriptionId>           # Trigger immediate run
saber market-signal signals <subscriptionId>           # List matched signals
```

Update flags: `--name`, `--prompt`, `--filters`, `--webhook`, `--webhook-secret`,
`--interval`, `--signal-limit`. At least one flag is required.

### Filter schemas by type

**JOB_POSTS:** `titleKeywords`, `excludeTitleKeywords`, `titlePatterns`, `countries`,
`seniority`, `technologySlugs`, `minSalaryUsd`, `maxSalaryUsd`, `remote`,
`companyDomains`, `excludeCompanyDomains`, `minEmployees`, `maxEmployees`,
`fundingStages`, `excludeRecruitingAgencies`, `promptFilter`, `maxLookbackDays`

**LINKEDIN_POST:** `keywordsAll`, `keywordsAny`, `keywordsNone`, `promptFilter`,
`maxLookbackDays`

**FUND_RAISED / RECENT_INVESTMENT / IPO:** `searchQueries`, `keywords`,
`excludeKeywords`, `countries`, `minAmountUsd`, `maxLookbackDays`, `promptFilter`

## Other Commands

```bash
saber credits                                         # Show remaining credit balance
saber connectors                                      # List configured connectors and status
saber version                                         # Print version, commit, build date
saber update                                          # Check for newer version and upgrade
saber init-claude                                     # Install Saber skill for Claude Code
saber help [command]                                  # Show help
```

## Update Checks

The CLI automatically checks for new versions in the background (at most once
every 24 hours). When an update is available, a one-line notice is printed to
stderr after the command finishes:

```
Notice: Update available v0.1.5 -> v0.1.7. Run "saber update" to upgrade.
```

The check is non-blocking: it never slows down command execution. If the network
request does not finish before the command exits, the notice is silently skipped.

The background check is **skipped** when any of these conditions apply:
- `--json` or `--quiet` flag is set
- `SABER_NO_UPDATE_CHECK=1` environment variable is set
- stderr is not a TTY (piped output, CI, AI agents)
- The command is `saber update` (which already performs its own check)
- Running a dev build (`version == "dev"`)

The cached state is stored in `~/.saber/update-check.json`.

## Global Flags

Available on every command:

| Flag | Short | Description |
|---|---|---|
| `--json` | | Output raw API JSON to stdout |
| `--quiet` | `-Q` | Suppress all non-error output |
| `--verbose` | `-v` | Log HTTP method, URL, status, rate-limit headers to stderr |
| `--api-url` | | Override base API URL (default: `https://api.saber.app`) |

## Environment Variables

| Variable | Description |
|---|---|
| `SABER_API_KEY` | API key (overrides `~/.saber/credentials.json`) |
| `SABER_NO_UPDATE_CHECK` | Set to any value to disable the background update check |

## Local Development

Point the CLI at a local Go Platform instance instead of production:

```bash
# Start the local stack (from monorepo root)
docker compose up

# Get a local API key
bun run saber apikey create --org <orgId>

# Run CLI commands against localhost
saber auth login --key sk_live_... --api-url http://localhost:3001/public
saber signal --domain acme.com --question "test" --api-url http://localhost:3001/public --verbose
```
