# JSON Schema

The stdout contract. Schema version is bumped on any breaking change.

## Success envelope

```json
{
  "ok": true,
  "data": "<command payload>",
  "meta": {
    "command": "surveys.list",
    "profile": "default",
    "duration_ms": 142,
    "schema_version": "2026-09-20",
    "request_id": "uuid-v4",
    "warnings": ["optional"],
    "pagination": { "limit": 100, "offset": 0, "total": 42, "has_more": true }
  }
}
```

`pagination` appears on every list response; omitted fields are omitted from JSON.

## Error envelope

```json
{
  "ok": false,
  "error": {
    "code": "API_ACCESS_FORBIDDEN",
    "message": "token is valid but this user/brand has no access to the API endpoint; ...",
    "category": "api",
    "retryable": false,
    "retry_after_ms": 0
  },
  "meta": { "command": "raw", "profile": "default", "duration_ms": 258, "schema_version": "2026-09-20" }
}
```

## Error codes and exit codes

| Code | Exit | Meaning |
|---|---|---|
| `AUTH_REQUIRED` | 3 | no token / no datacenter configured |
| `AUTH_TOKEN_INVALID` | 3 | Qualtrics did not recognize the token (HTTP 401, `DCD_7`) |
| `API_ACCESS_FORBIDDEN` | 6 | token valid but brand/user lacks API access (HTTP 403, `AuthZ_2.0`) |
| `READ_ONLY_VIOLATION` | 4 | `--read-only` blocked a write |
| `RATE_LIMITED` | 5 | HTTP 429; `retry_after_ms` populated when the server sends it |
| `NETWORK_UNREACHABLE` / `NETWORK_TIMEOUT` | 5 | transport failure |
| `API_ERROR` | 6 | 5xx or unmapped API failure |
| `API_SCHEMA_CHANGED` | 6 | Qualtrics returned a shape the CLI cannot parse |
| `RESOURCE_NOT_FOUND` | 6 | HTTP 404 |
| `VALIDATION_FAILED` | 7 | HTTP 400 (`QVAL_*`) |
| `CONFIRMATION_REQUIRED` | 10 | write attempted without `--confirm` |
| `INVALID_ARGUMENTS` | 2 | bad flags or arguments |
| `INTERNAL_ERROR` | 1 | everything else |

## Dry-run plan

```json
{
  "ok": true,
  "data": { "command": "distributions.create", "resource_id": "EMD_...", "after": { } },
  "meta": { }
}
```

## Guarantees

- stdout carries exactly one JSON document per command in `--json` mode; nothing else.
- Diagnostics and human-mode errors go to stderr.
- `meta.request_id` is a fresh UUID per invocation and also appears in API error messages
  from Qualtrics (their `requestId`), tying local and remote logs together.
