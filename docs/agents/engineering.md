# Engineering and decisions

- Keep the `mise.toml` comment-budget gate: comments in non-test `cmd/` and `internal/` code (except `//go:` and `//nolint`) remain below 5% of lines.
- Put enduring design rationale in sequentially numbered `docs/adr/` records; use concise code comments only for constraints that cannot be expressed in code.
- Do not add `CLAUDE.md`, `.cursorrules`, `.windsurfrules`, `.clinerules` or `GEMINI.md`; the conventions task enforces this.
- Use existing Go/Cobra patterns rather than new command infrastructure.
