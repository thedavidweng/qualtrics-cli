<h1 align="center">qualtrics-cli</h1>

<p align="center">
  Agent-friendly CLI for the Qualtrics Experience Management Platform & offline survey compiler.
</p>

<p align="center">
  <a href="https://github.com/thedavidweng/qualtrics-cli/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/thedavidweng/qualtrics-cli/ci.yml?branch=main&style=flat-square&label=ci" alt="CI"></a>
  <a href="https://github.com/thedavidweng/qualtrics-cli/releases"><img src="https://img.shields.io/github/v/release/thedavidweng/qualtrics-cli?style=flat-square" alt="Release"></a>
  <a href="https://github.com/thedavidweng/qualtrics-cli/blob/main/LICENSE"><img src="https://img.shields.io/github/license/thedavidweng/qualtrics-cli?style=flat-square" alt="License"></a>
  <img src="https://img.shields.io/badge/go-%3E%3D1.26-blue?style=flat-square" alt="Go">
</p>

`qualtrics-cli` gives researchers, data teams, scripts, and AI agents a stable terminal interface for Qualtrics: query surveys, definitions, asynchronous response exports and imports, distributions, contacts, webhooks, compile offline survey specifications to QSF, and call unmapped APIs directly.

## Highlights

- **Agent-first:** stable JSON envelope (`{ok, data, error, meta}`), separated stdout/stderr, request UUIDs, and machine-readable error codes
- **Safety gates:** `--read-only`, `--dry-run`, and `--confirm` gates; destructive operations (deleting surveys and their responses) require typed confirmation
- **Offline survey compiler:** compile plain-text Markdown survey specifications directly to Qualtrics Survey Format (`.qsf`) files (`qualtrics definitions build`) without requiring API access
- **Async jobs as first-class commands:** response exports and imports expose `start`, `status`, and `download`/`upload` with optional `--wait` polling and automated archive extraction
- **Full Survey Platform surface:** complete coverage of surveys, nested definitions, granular questions/blocks/flow/options CRUD, distributions, XM Directory mailing lists & contacts, and event subscriptions
- **Raw escape hatch:** `qualtrics raw <METHOD> <path>` for immediate authenticated access to any endpoint

## Why

Qualtrics is powerful in the browser, but automation often breaks because browser flows are slow and many brand accounts restrict REST API tokens (`AuthZ_2.0`), leaving researchers and autonomous agents blocked.

`qualtrics-cli` provides both:
1. **A complete REST API client** for accounts with API access.
2. **A 100% offline survey-to-QSF compiler** allowing users to describe surveys in Markdown and generate `.qsf` files ready to import via the Qualtrics web UI with zero API permissions needed.

## Quickstart

### Install

Run the following on macOS or Linux:

```shell
curl -fsSL https://raw.githubusercontent.com/thedavidweng/qualtrics-cli/main/install.sh | sh
```

Run the following on Windows (PowerShell):

```shell
powershell -ExecutionPolicy ByPass -c "irm https://raw.githubusercontent.com/thedavidweng/qualtrics-cli/main/install.ps1 | iex"
```

The binary installs to `/usr/local/bin` (or `~/.local/bin`).

<details>
<summary>Other installation methods</summary>

**Go:**

```shell
go install github.com/thedavidweng/qualtrics-cli/cmd/qualtrics@latest
```

**Manual download:** Grab the archive for your operating system from the [latest GitHub Release](https://github.com/thedavidweng/qualtrics-cli/releases/latest), extract it, and place the `qualtrics` binary on your `PATH`.

</details>

### Set up

```shell
qualtrics auth set-token
qualtrics auth set-datacenter pdx1
qualtrics auth status        # validates credentials and checks brand API access
qualtrics doctor             # validates local configuration files
```

Credentials live in `~/.config/qualtrics-cli/config.yaml` (mode 0600). The token can also be supplied via `QUALTRICS_TOKEN` or `env:NAME` secret indirection.

### 1. Offline Workflow: Text to QSF (No API Required)

Create a survey specification file `survey.md`:

```markdown
---
title: Customer Feedback Survey
language: EN
---

# General

## How satisfied are you with our platform? [mc]* @satisfaction
- Very satisfied
- Satisfied
- Neutral
- Unsatisfied

## What can we improve? [text-essay]
```

Compile it to a `.qsf` file and inspect its structure:

```shell
qualtrics definitions build survey.md -o survey.qsf
qualtrics definitions qsf summary survey.qsf
```

Import into Qualtrics: **Create Project → Import a QSF File**.

### 2. Online Workflow: API Operations

```shell
# List all surveys
qualtrics surveys list --json

# Inspect a survey definition
qualtrics definitions show SV_0123456789123

# Export survey responses with polling and automatic extraction
qualtrics responses export start SV_0123456789123 --format csv --wait --extract

# Send a raw API request
qualtrics raw GET /surveys
```

### Uninstall

```shell
rm -f /usr/local/bin/qualtrics ~/.local/bin/qualtrics
rm -rf ~/.config/qualtrics-cli
```

## Commands Overview

| Command Group | Description |
|---|---|
| `auth` | Manage credentials (`status`, `set-token`, `set-datacenter`, `logout`) |
| `doctor` | Check local configuration and environment |
| `definitions build` | Compile plain-text Markdown survey spec to `.qsf` |
| `definitions qsf` | Offline tools for `.qsf` inspection (`summary`) and conversion (`convert`) |
| `surveys` | Survey management (`list`, `show`, `create`, `delete`) |
| `definitions` | Survey definition documents (`show`, `export`, `import`) |
| `definitions questions` | Granular question CRUD (`list`, `show`, `create`, `update`, `delete`) |
| `definitions blocks` | Granular block CRUD (`list`, `show`, `create`, `update`, `delete`) |
| `definitions flow` | Survey flow inspection and update (`show`, `update`) |
| `definitions options` | Survey options inspection and update (`show`, `update`) |
| `responses export` | Asynchronous response export jobs (`start`, `status`, `download`) |
| `responses import` | Asynchronous response import jobs (`start`, `status`, `upload`) |
| `distributions` | Survey distributions and link generation (`list`, `show`, `create`, `delete`) |
| `distributions links` | Distribution link management (`list`, `show`, `create`, `update`) |
| `directories` | XM Directory containers (`list`, `show`, `create`, `update`, `delete`) |
| `directories mailinglists` | Mailing list management (`list`, `show`, `create`, `update`, `delete`) |
| `directories mailinglists contacts` | Contact records (`list`, `show`, `create`, `update`, `delete`) |
| `events subscriptions` | Webhook event subscriptions (`create`, `get`, `delete`) |
| `raw` | Authenticated passthrough for any Qualtrics endpoint |

See [COMMANDS.md](./COMMANDS.md) for the complete flag documentation and [JSON_SCHEMA.md](./JSON_SCHEMA.md) for the stdout envelope contract.

## Configuration

`~/.config/qualtrics-cli/config.yaml`:

```yaml
default_profile: default

profiles:
  default:
    datacenter: pdx1
    auth_method: token
    token: env:QUALTRICS_TOKEN
    timeout: 30s
    read_only: false
```

Configuration precedence: defaults → config file → `QUALTRICS_*` environment variables → flags.

## Safety Model

| Tier | Example Commands | Gate |
|---|---|---|
| `read` | `surveys list`, `definitions show`, `doctor` | None |
| `remote_action` | `responses export start`, `responses import upload` | None |
| `mutation` | `distributions create`, `definitions questions update` | `--confirm` (or `--dry-run`) |
| `destructive` | `surveys delete`, `contacts delete` | Typed confirmation (or `--confirm`) |

`--read-only` blocks all remote mutations and destructive commands. Exit codes:
- `2`: Invalid arguments
- `3`: Authentication required or invalid token
- `4`: Read-only mode violation
- `5`: Rate limit or network timeout
- `6`: API error or endpoint unauthorized (`AuthZ_2.0`)
- `7`: Validation failure
- `10`: Confirmation required

## Roadmap

Postman collections in the [Qualtrics public workspace](https://www.postman.com/qualtrics-public-apis/qualtrics-public-workspace/documentation/bcd3rug/qualtrics-survey-api):

| Collection | Status | Notes |
|---|---|---|
| **Survey Platform (Core)** | ✅ v1.0 | Complete coverage (49 endpoints) |
| **Offline Survey Compiler** | ✅ v1.0 | Pure-Go Markdown → QSF compiler |
| **XM Directory Extended** | 🔜 v1.1 | Contact deduplication, transactions, directory settings |
| **User Management** | 🔜 v1.1 | Users, groups, and division management |
| **CX / Discover** | ⬜ Planned | Dashboards, tickets, sentiment analysis |
| **Employee Experience (EX)** | ⬜ Planned | 360 cycles, participants, engagement pulses |
| **Research Core Panels** | ⬜ Planned | Legacy panel/sample endpoints |
| **Site Intercept** | ⬜ Planned | WRAPI system (`survey.qualtrics.com/WRAPI`) |

Planned capabilities:
- OAuth 2.0 client credentials grant (`auth login`)
- `responses synthetic` — generate synthetic survey data from declared behavioral priors

## Acknowledgments

- [n0g/qualtrics-survey-builder](https://github.com/n0g/qualtrics-survey-builder) — the plain-text survey spec design and reverse-engineered QSF format knowledge that `internal/qsf` builds upon.
- [matomatical/qualtrics](https://github.com/matomatical/qualtrics) — pioneering Python wrapper for the Survey Definitions REST API.
- [ctesta01's QSF gist](https://gist.github.com/ctesta01/d4255959dace01431fb90618d1e8c241) and [sumtxt/qsf](https://github.com/sumtxt/qsf) — QSF reverse-engineering notes.

## Development

```shell
mise run check      # fmt + build + test + lint + conventions
mise run drift      # verify endpoint catalog synchronization
mise run test-live  # live endpoint tests (when working token is configured)
```

Architecture decisions are recorded in [docs/adr/](./docs/adr/).

## License

Apache License 2.0 — see [LICENSE](./LICENSE).
