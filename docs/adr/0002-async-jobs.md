# Async jobs are first-class: start / status / download

Response exports and response imports are asynchronous in the Qualtrics API: a POST returns a progress ID, progress is polled, and the result is a separate file download. Rather than hiding this behind a blocking call, the CLI exposes the job as three commands (`responses export start|status|download`, `responses import start|status|upload`) plus a `--wait` flag on `start` that polls to completion and then downloads.

**Considered Options**: blocking-only (rejected: exports of large surveys run for many minutes and would hold a terminal with no progress signal); poll-and-hide entirely (rejected: scripts and agents want the progress ID so they can re-attach later).

**Consequences**: every async endpoint gets one shared `Job` type (progress ID, status, percent complete, file ID). Polling uses a fixed interval (`--interval`, default 5s) rather than exponential backoff, because the server reports `percentComplete` and there is nothing to back off from. The `--wait` path is the only place the CLI sleeps on purpose; client retry backoff is the only other sleep.
