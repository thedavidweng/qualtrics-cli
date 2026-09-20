<p align="center">
  <strong>qualtrics-cli</strong><br>
  An agent-friendly CLI for the Qualtrics API
</p>

<p align="center">
  <a href="https://github.com/thedavidweng/qualtrics-cli/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/thedavidweng/qualtrics-cli/ci.yml?style=flat-square" alt="CI"></a>
  <a href="https://github.com/thedavidweng/qualtrics-cli/releases"><img src="https://img.shields.io/github/v/release/thedavidweng/qualtrics-cli?style=flat-square" alt="Release"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/github/license/thedavidweng/qualtrics-cli?style=flat-square" alt="License"></a>
  <img src="https://img.shields.io/badge/go-1.26.5-00ADD8?style=flat-square&logo=go" alt="Go">
</p>

---

A single-binary CLI for the [Qualtrics v3 REST API](https://api.qualtrics.com/) — surveys,
survey definitions, response exports and imports, distributions, directories,
mailing lists, contacts, and event subscriptions. Built for terminals, scripts,
and local agents.

## Highlights

- **Agent-first JSON** — every command emits a stable envelope
  (`{ok, data, error, meta}`) with `meta.pagination`, exit codes, and machine-readable error codes.
- **Safety gates** — `--read-only`, `--dry-run`, `--confirm`; destructive operations
  (e.g. deleting a survey, which also deletes its responses) require typed confirmation.
- **Async jobs as first-class commands** — response exports and imports expose
  `start | status | download` with an optional `--wait`.
- **Offline survey toolchain** — build QSF survey files from a plain-text spec with
  `definitions build`, no API access required.
- **Escape hatch** — `qualtrics raw GET /any/endpoint` for anything not yet wrapped.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/thedavidweng/qualtrics-cli/main/install.sh | bash
```

Or from source:

```bash
go install github.com/thedavidweng/qualtrics-cli/cmd/qualtrics@latest
```

## Quickstart

Generate an API token in Qualtrics (Account Settings → Qualtrics IDs → Generate Token),
find your datacenter ID on the same page (e.g. `pdx1`, `iad1`, `fra1`), then:

```bash
qualtrics auth set-token
qualtrics auth set-datacenter pdx1
qualtrics auth status        # verifies the token against the API
qualtrics doctor             # local configuration checks
```

Credentials live in `~/.config/qualtrics-cli/config.yaml` (mode 0600). The token may
be the literal value or `env:VAR_NAME` to read it from the environment. Multiple
brands are supported via profiles (`--profile`, `QUALTRICS_PROFILE`).

## Commands

```
qualtrics auth status|set-token|set-datacenter|logout
qualtrics doctor
qualtrics raw <METHOD> <path> [-d body]

qualtrics surveys list|show|create|delete
qualtrics definitions show|export|import
qualtrics definitions questions  list|show|create|update|delete
qualtrics definitions blocks     list|show|create|update|delete
qualtrics definitions flow       show|update
qualtrics definitions options    show|update
qualtrics definitions build <spec> [-o out.qsf]

qualtrics responses export start|status|download
qualtrics responses import start|status|upload

qualtrics distributions list|show|create|delete
qualtrics distributions links list|create|show|update|delete

qualtrics directories list|create|show|update|delete
qualtrics directories mailinglists list|create|show|update|delete
qualtrics directories mailinglists contacts list|create|show|update|delete

qualtrics events subscriptions create|list|get|delete
```

See [COMMANDS.md](./COMMANDS.md) for the full flag reference and
[JSON_SCHEMA.md](./JSON_SCHEMA.md) for the output contract.

## Configuration

`~/.config/qualtrics-cli/config.yaml`:

```yaml
default_profile: default
profiles:
  default:
    datacenter: pdx1          # or base_url: https://pdx1.qualtrics.com/API/v3
    auth_method: token        # oauth arrives in a later release
    token: env:QUALTRICS_TOKEN
```

Precedence: defaults → config file → `QUALTRICS_*` environment variables → flags.

## Safety

| Tier | Examples | Gate |
|---|---|---|
| read | `surveys list`, `definitions show` | none |
| remote_action | `responses export start` | none |
| mutation | `distributions create`, `questions update` | `--confirm` (or `--dry-run`) |
| destructive | `surveys delete`, `contacts delete` | typed-ID confirmation |

`--read-only` blocks everything above read. Exit codes: 2 args, 3 auth, 4 read-only,
5 rate-limit/network, 6 API, 7 validation, 10 confirmation required.

## Roadmap

Postman collections in the [Qualtrics public workspace](https://www.postman.com/qualtrics-public-apis/qualtrics-public-workspace/documentation/bcd3rug/qualtrics-survey-api)
that are **not yet covered**:

| Collection | Status | Notes |
|---|---|---|
| Survey Platform (core) | ✅ v1 | surveys, definitions, responses, distributions, directories, events |
| XM Directory | 🔜 v1.1 | remainder of directories/mailinglists/contacts surface |
| User Management | 🔜 v1.1 | users, groups, divisions |
| CX / Discover | ⬜ planned | dashboards, tickets, CX responses |
| Employee Experience | ⬜ planned | engagement surveys, participants |
| Qualtrics 360 | ⬜ planned | review cycles, subjects, reviewees |
| Research Core | ⬜ planned | panel/sample management |
| Site Intercept | ⬜ planned | separate WRAPI system (`survey.qualtrics.com/WRAPI`) |

Additional planned work:

- OAuth 2.0 client-credentials auth (`auth login`)
- `definitions import --from-qsf` — direct QSF upload via the API (needs live verification)
- `responses synthetic` — generate synthetic response data from a declared behavioral model

## Acknowledgments

- [n0g/qualtrics-survey-builder](https://github.com/n0g/qualtrics-survey-builder) — the plain-text
  survey spec design and reverse-engineered QSF format knowledge that `definitions build` builds on.
- [matomatical/qualtrics](https://github.com/matomatical/qualtrics) — early mapping of the
  survey-definitions REST surface.
- [ctesta01's QSF gist](https://gist.github.com/ctesta01/d4255959dace01431fb90618d1e8c241) and
  [sumtxt/qsf](https://github.com/sumtxt/qsf) — QSF format references.

## Development

```bash
mise run check     # fmt + build + test + lint + conventions
mise run test-live # live endpoint checks (needs a token with API access)
```

Decisions are recorded in [docs/adr/](./docs/adr/); the endpoint catalog is
[docs/endpoints.yaml](./docs/endpoints.yaml) and is checked for drift in CI.

## License

Apache License 2.0 — see [LICENSE](./LICENSE).
