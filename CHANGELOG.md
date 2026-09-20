# Changelog

All notable changes to this project are documented in this file. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Repository skeleton: cobra root with `--json`/`--pretty`/`--read-only`/`--dry-run`/
  `--confirm`/`--timeout`/`--profile` persistent flags with `QUALTRICS_*` env fallbacks.
- `internal/qualtrics` REST client: idempotent-method retries with `Retry-After`,
  10MB response cap, redirect rejection, Qualtrics error-envelope mapping
  (`DCD_7` → invalid token, `AuthZ_2.0` → no API access, `QVAL_*` → validation).
- `internal/output` JSON envelope with `meta.pagination`, `internal/errors` taxonomy
  with exit codes, `internal/safety` operation tiers.
- `internal/config` YAML profiles with datacenter/base_url resolution and `env:NAME`
  secret indirection; config store writes at 0600.
- Commands: `auth status|set-token|set-datacenter|logout`, `doctor`, `version`,
  `completion`, `raw`.
- Endpoint catalog (`docs/endpoints.yaml`) with CI drift check, ADRs 0001–0006,
  `CONTEXT.md` glossary.

## [Unreleased] — survey platform core

### Added

- `surveys list|show|create|delete` with `--limit/--offset/--all`, edit and preview
  links in `surveys show`, and typed-ID confirmation for delete.
- `definitions show|export|import` plus granular `questions`, `blocks`, `flow`, and
  `options` CRUD; payloads read from `-f <file>` (`-` for stdin).
- `runList` runner populating `meta.pagination{limit,offset,total,has_more}`.
- Transport-stub contract tests pinning request paths, methods, query strings, and
  bodies for every new endpoint (18 catalog entries now `done`).

### Added — async response jobs

- `responses export start|status|download` with format/label/timezone/breakout options,
  `--wait` polling, and `--extract` unzipping (zip-slip protected).
- `responses import start|status|upload` with multipart file upload.
- Binary download and multipart upload paths in the client, with retry on downloads.
- 24 catalog entries now `done`.
