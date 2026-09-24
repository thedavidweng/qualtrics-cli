# 0008: E2E-First Testing

Status: Accepted

Context: Request-shape mirror tests asserted method and path constants. Envelope tests string-matched marshaled JSON. E2E covered help text without artifact except definitions build.

Decision: E2E is default. definitions build plus qsf summary roundtrip is the artifact test. Isolated tests remain only for offline compiler failure modes, zip-slip, file permissions, secret indirection, retry and redirect rejection, and error mapping with exact codes. Coverage stays informational; no codecov gate exists.

Consequences: Deleted envelope, events, directories, distributions, and shape-mirror tests. Exit-code table retained as contract.
