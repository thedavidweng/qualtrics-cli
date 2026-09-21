# Changelog

All notable changes to this project are documented in this file. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- Spec conformance pass: `meta.request_id` is now present on error envelopes;
  `RESOURCE_NOT_FOUND` exits 6; `API_SCHEMA_CHANGED` is documented in
  `JSON_SCHEMA.md`.
- `raw` write methods go through the safety gates (POST/PUT need `--confirm`,
  DELETE is destructive); JSON-mode destructive commands never prompt (exit 10
  without `--confirm`); `--confirm` satisfies non-typed destructives; only
  `surveys delete` requires typing the ID when interactive.
- `--full` is honored by every human summary; `--dry-run` prints the plan in
  human mode; progress, prompts, and CLI errors go to stderr; `-o -` with
  `--json` is rejected instead of mixing bytes with the envelope.
- `distributions links delete` implemented (typed client, command, catalog);
  `definitions questions show` added to the endpoint catalog.
- Every list command accepts `--all` (mutually exclusive with `--offset`);
  client honors the HTTP `Retry-After` header; 401 maps to invalid-token only
  for `DCD_7`/empty codes; `responses import start` supports `--wait`.
- Async jobs share one `Job` type; response fixtures live in
  `internal/qualtrics/testdata/`; `convert` skips Trash blocks like `summary`.

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

### Added — distributions, directories, and events

- `distributions list|show|create|delete` and `distributions links list|show|create|update`.
- `directories list|show|create|update|delete`.
- `directories mailinglists list|show|create|update|delete`.
- `directories mailinglists contacts list|show|create|update|delete`.
- `events subscriptions create|get|delete`.
- All 49 survey platform endpoints in `docs/endpoints.yaml` are now implemented and tested.

### Added — offline survey toolchain (qsf compiler)

- `internal/qsf`: pure-Go markdown-to-QSF compiler (`ParseSurvey`, `BuildQSF`,
  `SummarizeQSF`, `ConvertQSFToDefinition`). Supports blocks, page breaks, 9 question
  types, matrix questions with scale/scale-translations, branch-if, loop-from,
  carry-from, show-if display logic, skip-if skip logic, recode values, variable
  names, and inline text entry.
- `definitions build <spec.md> [-o out.qsf]` compiles survey markdown to QSF offline.
- `definitions qsf summary <file.qsf>` inspects QSF structure and question counts.
- `definitions qsf convert <file.qsf> [-o out.json]` extracts survey definition JSON.
- ADR-0007 documents the offline compiler architecture.

### Added — e2e tests and release scaffolding

- `tests/e2e/binary_test.go`: builds real binary, tests `--help` command discovery
  against `requiredCommands` golden list, tests `version --json` envelope contract,
  `doctor`, and `definitions build` + `qsf summary`.
- Installer scripts: `install.sh` and `install.ps1`.
- Release automation: `.goreleaser.yaml`, `.github/workflows/release.yml`, and
  `release-please.yml`.
