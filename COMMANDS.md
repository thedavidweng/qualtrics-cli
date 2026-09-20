# Commands

Every command, its flags, and its output shape. `mise run check` enforces that this file
matches the command tree.

## Global flags

| Flag | Env | Description |
|---|---|---|
| `--json` | `QUALTRICS_JSON` | machine-readable envelope on stdout |
| `--pretty` | `QUALTRICS_PRETTY` | pretty-print JSON |
| `--read-only` | `QUALTRICS_READ_ONLY` | block all remote writes |
| `--dry-run` | `QUALTRICS_DRY_RUN` | print the planned write, do not execute |
| `--confirm` | `QUALTRICS_CONFIRM` | execute a remote write |
| `--timeout` | `QUALTRICS_TIMEOUT` | per-request timeout (default 30s) |
| `--profile` | `QUALTRICS_PROFILE` | config profile (default `default`) |
| `--config` | `QUALTRICS_CONFIG` | config file path |

## auth

| Command | Description |
|---|---|
| `auth status` | Probe the API and report token state: `ok` / `invalid_token` / `no_api_access` |
| `auth set-token [--token T]` | Store a token (prompts if absent; falls back to `QUALTRICS_TOKEN`) |
| `auth set-datacenter <id>` | Store the datacenter ID (e.g. `pdx1`) |
| `auth logout` | Remove the stored token |

## doctor

Local checks only (no API call): config file, profile, datacenter, base URL, token source.

## raw

`raw <METHOD> <path> [--data JSON] [--method M]` — authenticated passthrough. Output is the
raw `result` payload inside the standard envelope.

## version / completion

`version` prints build info. `completion bash|zsh|fish|powershell` emits shell scripts.

## Survey Platform (planned — see docs/endpoints.yaml)

| Command | Tier |
|---|---|
| `surveys list\|show\|create\|delete` | read / read / mutation / destructive |
| `definitions show\|export\|import` | read / read / mutation |
| `definitions questions list\|show\|create\|update\|delete` | read…destructive |
| `definitions blocks list\|show\|create\|update\|delete` | read…destructive |
| `definitions flow show\|update` | read / mutation |
| `definitions options show\|update` | read / mutation |
| `responses export start\|status\|download` | remote_action / read / read |
| `responses import start\|status\|upload` | remote_action / read / remote_action |
| `distributions list\|show\|create\|delete` | read…destructive |
| `distributions links list\|create\|show\|update\|delete` | read…destructive |
| `directories list\|create\|show\|update\|delete` | read…destructive |
| `directories mailinglists list\|create\|show\|update\|delete` | read…destructive |
| `directories mailinglists contacts list\|create\|show\|update\|delete` | read…destructive |
| `events subscriptions create\|list\|get\|delete` | mutation / read / read / destructive |

List commands accept `--limit`, `--offset`, and `--all`. Deeply nested payloads default to
a summary view; `--full` prints everything.
