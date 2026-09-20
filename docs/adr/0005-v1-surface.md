# v1 surface: survey lifecycle, responses, distribution, contacts

v1 covers the Survey Platform core — surveys, survey definitions (whole-document plus granular questions/blocks/flow/options CRUD), response export/import, distributions and their links, directories with mailing lists and contacts, event subscriptions — plus a `raw` escape hatch for anything unimplemented. User Management (users, groups, divisions) is deferred to v1.1; Site Intercept, CX/Discover, Employee Experience, and 360 are tracked in the README roadmap, not built.

**Considered Options**: everything at once (rejected: the full Postman workspace spans several products with different conventions; a shallow everything-CLI is worse than a deep core); definitions-only (rejected: pulling responses and sending distributions are the reasons people automate Qualtrics at all).

**Consequences**: an in-repo endpoint catalog (`docs/endpoints.yaml`) records every implemented endpoint with its path, method, safety tier, owning command, and pagination convention; a CI check fails when the catalog and the command tree drift apart. The catalog is our own source of truth because Qualtrics' documentation is known to be incomplete and occasionally inaccurate. The `raw` command exists precisely so that an agent hitting an unimplemented endpoint is never blocked.
