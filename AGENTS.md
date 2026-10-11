# qualtrics-cli

Go/Cobra CLI for the Qualtrics API and offline Markdown-to-QSF survey compilation.

- Before committing, run `mise run check` (format, build, tests, lint, conventions).
- Keep `AGENTS.md` as the sole canonical agent entry point; do not create per-tool instruction files.
- For task-specific rules, read only:
  - [CLI/API contracts](docs/agents/contracts.md) when changing commands, output or endpoints.
  - [Architecture and endpoint integration](docs/agents/architecture.md) when adding features or API methods.
  - [Testing policy](docs/agents/testing.md) when modifying behavior or tests.
  - [Code and decision rules](docs/agents/engineering.md) for Go code, comments, dependencies and ADRs.
