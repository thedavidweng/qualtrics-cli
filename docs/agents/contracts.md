# CLI and API contracts

- Update `COMMANDS.md` and `JSON_SCHEMA.md` in the same commit as command, flag or output-shape changes; new endpoints also update `docs/endpoints.yaml`.
- Dotted command IDs (for example `surveys.list`) must match across `meta.command`, errors and the endpoint catalog.
- Human output is summarized by default; `--full` prints the full payload. JSON-mode stdout is machine-readable; diagnostics use stderr.
- Client retries idempotent GET/PUT/DELETE requests on 429, 5xx and transient network errors, honoring `Retry-After`; POST is not retried.
- Limit API response bodies to 10 MB and downloads to 200 MB; reject redirects.
- Preserve the existing read-only, dry-run, confirm and operation-tier safety contracts.
