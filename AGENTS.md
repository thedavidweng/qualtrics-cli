# Agent Instructions

This repository is developed primarily by coding agents. Read this before making changes.

## Non-negotiables

- **One agent-instructions file.** This file. Never add `CLAUDE.md`, `.cursorrules`,
  `.windsurfrules`, `.clinerules`, or `GEMINI.md`.
- **Docs move with code.** Any change to a command, flag, or output shape updates
  `COMMANDS.md` and `JSON_SCHEMA.md` in the same commit. New endpoints update
  `docs/endpoints.yaml` in the same commit.
- **Comment budget.** Comments (excluding `//go:` and `//nolint`) must stay under 5% of
  non-test code lines in `cmd/` and `internal/`. Rationale belongs in `docs/adr/`, not in code.
- **`mise run check` must pass** before any commit: fmt, build, test, lint, conventions.

## Architecture

```
cmd/qualtrics/          entry point only
internal/cli/           cobra command layer; one file per domain
internal/qualtrics/     HTTP client + typed domain methods + testdata fixtures
internal/config/        YAML profiles, env:NAME secrets, paths
internal/auth/          (planned) oauth token exchange
internal/output/        JSON envelope + renderer (the stdout contract)
internal/errors/        error taxonomy + exit codes
internal/safety/        operation tiers + gates
internal/qsf/           (planned) plain-text spec → QSF compiler
```

### Adding an endpoint

1. Add the typed method in `internal/qualtrics/<domain>.go`.
2. Add the cobra command in `internal/cli/<domain>.go` using `run[T]` (reads) or
   `runMutation` (writes). Never construct a client by hand in a command.
3. Add the entry to `docs/endpoints.yaml` with the correct tier and pagination convention.
4. Add a transport-stub test pinning the request path, method, headers, and body.
5. Update `COMMANDS.md` and `JSON_SCHEMA.md`.

### Conventions

- Commands are named by dotted strings (`surveys.list`, `responses.export.start`) that flow
  into `meta.command`, error payloads, and the endpoint catalog. One identifier everywhere.
- Human output is a summary by default; `--full` prints the complete payload. stdout stays
  parseable in JSON mode; diagnostics go to stderr.
- The client retries idempotent methods (GET/PUT/DELETE) on 429 and 5xx with `Retry-After`
  honored; POST is never retried (creating an export job twice is worse than failing once).
- Response bodies are capped at 10MB; redirects are rejected outright.

## Testing without API access

The brand's token may lack API access (403 `AuthZ_2.0`). The client, envelope, pagination,
safety, and QSF layers are all tested without live access. Fixtures for survey definitions
and response exports come from reverse-engineered QSF knowledge (see ADR-0006), not from
invented shapes. `mise run test-live` runs against the real API when a working token exists.
