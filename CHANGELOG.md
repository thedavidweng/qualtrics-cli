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
