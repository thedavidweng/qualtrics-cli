# Safety tiers, with typed confirmation for survey deletion

Every command maps to one of four tiers — read, remote-action, mutation, destructive — mirroring the sibling monarchmoney-cli design. `--read-only` blocks everything above read. `--dry-run` prints a plan instead of executing. Remote-action, mutation, and destructive operations require `--confirm`. The exception: `surveys delete` requires typing the survey ID, not just `--confirm`.

**Considered Options**: uniform `--confirm` for all writes (rejected: `DELETE /survey-definitions/{id}` also destroys every response ever collected — irreversible data loss that a single flag does not deserve); no gates at all (rejected: this CLI is built for agent use, where an accidental destructive call has no human in the loop).

**Consequences**: the safety check lives in the generic `runMutation` helper, so a new write command cannot forget it. The typed-ID prompt is the only interactive confirmation in the CLI; everything else is flag-driven. In JSON mode, a missing `--confirm` is exit code 10 with a structured error, never a prompt.
