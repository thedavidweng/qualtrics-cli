# Architecture and adding endpoints

`cmd/qualtrics/` is the thin entry point. `internal/cli/` hosts Cobra domains; `internal/qualtrics/` contains typed HTTP methods and test fixtures; `internal/config/` handles YAML profiles and `env:NAME` secrets; `internal/output/` owns JSON envelopes and rendering; `internal/errors/` owns exit codes; `internal/safety/` owns operation tiers; `internal/qsf/` is the offline spec compiler. `internal/auth/` is planned OAuth support, not evidence that it exists.

For a new endpoint:
1. Add a typed method in `internal/qualtrics/<domain>.go`.
2. Add a Cobra command in `internal/cli/<domain>.go`, using `run[T]` for reads or `runMutation` for writes; do not instantiate clients ad hoc in commands.
3. Add `docs/endpoints.yaml` with tier and pagination.
4. Verify the new endpoint through a repeatable offline E2E scenario. If E2E cannot exercise critical request details (method, path, headers, body) or failure paths, add a focused transport-stub or isolated test for those gaps.
5. Update `COMMANDS.md` and `JSON_SCHEMA.md`.
