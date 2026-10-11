# Testing policy

- Create the relevant checks before implementing behavior where feasible; do not defer unit/regression tests to after implementation.
- End-to-end user workflows are the default verification target and must produce repeatable evidence.
- Add transport-stub tests only when repeatable E2E cannot verify the relevant request contract or failure path; do not require one automatically for every endpoint. Normal verification must remain runnable without a live Qualtrics token.
- Use focused isolated tests when E2E cannot reach a failure; enumerate failures first. Avoid string-mirror, constant-mirror and trivial getter tests.
- Coverage is diagnostic evidence, not a target to game.
- Real Qualtrics tokens can return 403 `AuthZ_2.0`; run the client, envelope, pagination, safety and QSF tests offline. QSF fixtures must come from established knowledge (ADR 0006), not invented payloads.
- Use `mise run test-live` only when an API-capable token is available. Normal `mise run check` must not depend on that token.
