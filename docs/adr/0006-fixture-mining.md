# Fixtures mined from reverse-engineered format docs

The test pyramid is transport-stub unit tests, fixture-driven contract tests, and live tests gated behind `QUALTRICS_LIVE_TOKEN`. Because the Qualtrics documentation is incomplete, the fixtures for survey definitions and response exports are built from the reverse-engineered QSF format reference (block/flow/options/question element envelopes, ID formats) and the documented CSV export column layout, not from invented shapes.

**Considered Options**: hand-written minimal fixtures (rejected: they encode our guesses about the API, so the tests would pass while the real API disagrees); waiting for API access before testing (rejected: the token's brand currently returns 403 `AuthZ_2.0` on every endpoint, and the client, envelope, pagination, and safety layers are all testable without a live account).

**Consequences**: fixtures live in `internal/qualtrics/testdata/` and are treated as captured truth — when live access arrives, the first job is to diff real responses against them and correct the fixtures, not the tests. The QSF-derived fixture knowledge is also what makes a future `definitions import --from-qsf` feasible.
