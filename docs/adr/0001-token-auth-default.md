# X-API-TOKEN as the default auth, OAuth deferred

The Qualtrics v3 API authenticates with a per-user `X-API-TOKEN` header and also supports OAuth 2.0 client credentials. We ship token auth first: one static secret per profile, stored in `~/.config/qualtrics-cli/config.yaml` at mode 0600, overridable by `QUALTRICS_TOKEN`. OAuth client-credentials support (client id/secret → cached access token with refresh) is deferred to v1.1.

**Considered Options**: OAuth-only (rejected: most individual users only have a token, and the login dance buys revocation we don't need yet); token-only forever (rejected: org-level automation is the direction Qualtrics is pushing, and the client layer is shared so adding it later is cheap).

**Consequences**: `auth status` must distinguish three failure modes — valid token, unrecognized token (401 `DCD_7`), and recognized-but-unauthorized (403 `AuthZ_2.0`, the state a brand without the API feature lands in). The token is scoped to a single datacenter; the datacenter is part of a profile's identity, not a global setting.
