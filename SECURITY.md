# Security Policy

## Reporting a Vulnerability

Report security issues privately to the repository maintainer via GitHub Security
Advisories. Do not open a public issue for anything involving credential handling.

## Credential handling

- API tokens are stored only in `~/.config/qualtrics-cli/config.yaml` (mode 0600,
  directory 0700). Tokens are never logged, never included in error payloads, and
  never written to the audit trail.
- `env:VAR_NAME` indirection keeps secrets out of the config file entirely for CI use.
- `qualtrics auth status` reports only whether a token is set and whether the API
  accepts it — never the token value.

## Threat model notes

- The CLI is built for unattended agent use. All remote writes are gated
  (`--confirm` / typed-ID confirmation) and `--read-only` blocks them wholesale.
- The HTTP client rejects redirects, caps response bodies at 10MB, and retries only
  idempotent methods so a retry storm cannot duplicate a write.
- Qualtrics tokens act as the user. Treat the config file as equivalent to the
  account password: anyone who reads it can read and modify all surveys and
  response data.
