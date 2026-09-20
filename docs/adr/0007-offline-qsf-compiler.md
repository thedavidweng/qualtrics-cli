# Pure-Go plain-text spec to QSF compiler for offline workflow

To support survey creation without requiring live API access (such as when an organization has not enabled the API feature for a brand's tokens), we include a pure-Go compiler in `internal/qsf`. It compiles plain-text Markdown survey specifications directly into Qualtrics Survey Format (`.qsf`) files ready to import via the web UI (Create Project → Import a QSF File).

**Considered Options**: Python-only external tool (rejected: requires a separate runtime and breaks the single-binary Go CLI experience); waiting for API access (rejected: leaves users blocked when brand tokens lack API authorization); inventing an incompatible syntax (rejected: adopting the established survey markdown DSL maintains compatibility with existing agent skills).

**Consequences**: Surveys can be written in plain-text markdown, compiled locally with `qualtrics definitions build <file.md>`, inspected with `qualtrics definitions qsf summary <file.qsf>`, and converted to API definition payloads with `qualtrics definitions qsf convert <file.qsf>`.
