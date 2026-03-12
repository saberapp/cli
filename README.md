# Saber CLI

Run company signal research, manage lists, and check credits from your terminal.

```
saber signal --domain acme.com --question "Are they hiring engineers?"
saber list company create --name "US SaaS" --industry "Computer Software" --size "51-200"
saber list contact create --name "VP Sales prospects" --company-linkedin https://linkedin.com/company/acme --title "VP Sales"
saber credits
```

---

## Installation

### curl (recommended)

```sh
curl -sSL https://raw.githubusercontent.com/saberapp/cli/main/scripts/install.sh | sh
```

Installs to `/usr/local/bin/saber` (prompts for sudo if needed). Works on macOS and Linux.

### go install

```sh
go install github.com/saberapp/cli/cmd/saber@latest
```

> Note: `go install` puts the binary in `$GOPATH/bin` (usually `~/go/bin`). Make sure that's in your `PATH`:
> ```sh
> export PATH="$HOME/go/bin:$PATH"
> ```

---

## Authentication

Get an API key at **[ai.saber.app → Settings → API Keys](https://ai.saber.app)**.

```sh
saber auth login                    # prompts for key (masked input)
saber auth login --key sk_live_...  # non-interactive / CI
saber auth status
saber auth logout
```

Keys are validated against the API before being saved. The key format is `sk_live_` + 43 characters (51 chars total).

---

## Commands

### `saber signal`

Run a company signal research query.

```sh
saber signal --domain acme.com --question "Are they hiring engineers?"
saber signal --domain acme.com --question "What is their headcount?" --answer-type number
saber signal --domain acme.com --question "Tech stack?" --json
saber signal --domain acme.com --question "..." --no-wait   # fire and forget, prints ID
saber signal --domain acme.com --question "..." --force-refresh
```

| Flag | Default | Description |
|---|---|---|
| `--domain` / `-d` | required | Company domain |
| `--question` / `-q` | required | Research question (max 500 chars) |
| `--answer-type` / `-a` | `open_text` | `open_text`, `boolean`, `number`, `list`, `percentage`, `currency`, `url`, `json_schema` |
| `--force-refresh` | false | Bypass 12h result cache |
| `--webhook` | — | Webhook URL — skips polling, returns immediately |
| `--no-wait` | false | Return the signal ID immediately without polling |
| `--poll-interval` | `3` | Seconds between poll attempts |
| `--max-wait` | `120` | Max seconds to wait before timing out |

### `saber list company`

Manage company lists backed by Saber's firmographic database.

```sh
saber list company create --name "US SaaS" --industry "Computer Software" --size "51-200" --country US
saber list company list
saber list company get <listId>
saber list company update <listId> --name "Updated name"
saber list company delete <listId>
saber list company companies <listId>          # paginated list of companies
saber list company search --industry "Fintech" # preview without creating
saber list company import --name "From HubSpot" --property industry --operator EQ --value Technology
```

`--industry`, `--size`, and `--country` flags are all repeatable:
```sh
saber list company create --name "Mid-market" \
  --industry "Computer Software" --industry "Internet" \
  --size "51-200" --size "201-500" \
  --country US --country GB
```

### `saber list contact`

Manage contact lists sourced from LinkedIn Sales Navigator.

```sh
saber list contact create --name "VP Sales at Stripe" \
  --company-linkedin https://linkedin.com/company/stripe \
  --title "VP Sales" --title "Head of Sales"

saber list contact list
saber list contact get <listId>
saber list contact update <listId> --name "New name"
saber list contact delete <listId>
saber list contact contacts <listId>    # paginated list of contacts
```

`--company-linkedin` is repeatable to target multiple companies. Creating a list runs a live Sales Navigator search and snapshots ~125 contacts.

### `saber credits`

```sh
saber credits
# Remaining credits: 4,200
```

### `saber connectors`

```sh
saber connectors
# CONNECTOR       STATUS
# salesNavigator  connected
# hubspotApp      disconnected
```

### `saber version`

```sh
saber version
# saber version 0.1.1 (commit abc1234, built 2026-03-12T10:00:00Z)
```

---

## Global flags

Available on every command:

| Flag | Description |
|---|---|
| `--json` | Output raw API response JSON to stdout |
| `--quiet` / `-Q` | Suppress all non-error output |
| `--verbose` / `-v` | Log HTTP method, URL, masked auth header, status, and rate-limit headers to stderr |
| `--api-url` | Override the base API URL (default: `https://api.saber.app`) |

---

## Running locally for testing

Point the CLI at your local Go Platform instance instead of production.

### 1. Start the local stack

From the monorepo root:
```sh
docker compose up   # starts mongo, postgres, redis, temporal, go-platform
```

The Go Platform public routes are served at `http://localhost:3001/public`.

### 2. Get a local API key

```sh
bun run saber apikey create --org <orgId>
```

### 3. Run CLI commands against localhost

```sh
saber auth login --key sk_live_... --api-url http://localhost:3001/public
saber signal --domain acme.com --question "Are they hiring?" --api-url http://localhost:3001/public
saber credits --api-url http://localhost:3001/public
```

### 4. Verbose mode for debugging

```sh
saber signal --domain acme.com --question "test" --verbose --api-url http://localhost:3001/public
# > POST http://localhost:3001/public/v1/companies/signals
# > Authorization: Bearer sk_live_xxx*****xxxx
# < HTTP 201
# < X-Ratelimit-Remaining: 99
```

---

## Development

```sh
make build     # build binary to bin/saber
make test      # go test ./...
make fmt       # gofmt -w .
make lint      # golangci-lint run ./...
make install   # copy to $GOPATH/bin
```

### Project layout

```
├── cmd/
│   ├── saber/
│   │   └── main.go              # Entry point; ldflags inject version/commit/date
│   ├── root.go                  # Persistent flags: --json, --quiet, --verbose, --api-url
│   ├── signal.go                # saber signal
│   ├── list.go                  # saber list company / contact (all subcommands)
│   ├── auth.go                  # saber auth login / logout / status
│   ├── credits.go               # saber credits
│   ├── connectors.go            # saber connectors
│   ├── version.go               # saber version
│   └── helpers.go               # mustClient() — auth check + HTTP client factory
└── internal/
    ├── client/
    │   ├── client.go            # HTTP client; retry/backoff; verbose logging
    │   ├── signals.go           # Company signal types + Create/Get
    │   ├── lists.go             # Company and contact list types + CRUD
    │   ├── credits.go           # CreditsBalance + GetCredits
    │   └── connectors.go        # Connector types + ListConnectors
    ├── config/
    │   └── config.go            # API key load/save/validate; ~/.saber/credentials.json
    └── format/
        ├── signal.go            # PrintSignal (tabwriter)
        ├── list.go              # PrintCompanyList, PrintContactList, etc.
        ├── table.go             # tabwriter helpers
        └── spinner.go           # Poll spinner (suppressed when non-TTY)
```

### Releasing

Tag with `v<semver>` to trigger the release workflow:
```sh
git tag v0.2.0
git push origin v0.2.0
```

goreleaser builds binaries for macOS/Linux/Windows and uploads them to the GitHub release.

---

## How it works

The CLI talks directly to the Saber public API (`https://api.saber.app`), authenticated with an API key stored at `~/.saber/credentials.json`.

**Signal flow** — `saber signal` calls the sync endpoint and waits for the result (use `--no-wait` for fire-and-forget).

**Lists** — company and contact lists are stored server-side. Create a list once, then fetch its companies or contacts as many times as you need.

**Auth** — the API key is looked up in order:
1. `SABER_API_KEY` env var (CI-friendly; never written to disk)
2. `~/.saber/credentials.json` (written by `saber auth login`)
