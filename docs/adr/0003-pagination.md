# One paginator for three wire conventions

Qualtrics list endpoints do not share a pagination convention: some take `offset`/`limit`, some return a `continuationToken`, some return a `nextPage` URL. The CLI normalizes all three behind a generic `Page[T]` / `Paginate[T]` pair in the client package. User-facing list commands expose `--limit`/`--offset` and an `--all` flag that transparently walks every page; token/page-URL endpoints expose only `--all`.

**Considered Options**: per-endpoint hand-rolled loops (rejected: the same bug would be re-implemented a dozen times, and the three conventions are easy to confuse); cursor-only abstraction (rejected: it would force the offset endpoints into an unnatural shape).

**Consequences**: the JSON envelope carries `meta.pagination{limit,offset,total,has_more}` on every list response, populated from whichever convention the endpoint actually used. `--all` never mixes with `--offset` (validation error). The envelope's pagination block is the contract scripts rely on; the wire conventions stay an implementation detail of the client package.
