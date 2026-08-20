# 5. FileMaker boundary

internal/filemaker defines transport-neutral request and response types plus an
Adapter interface. A future Data API adapter and OData adapter can both return
the same normalized shape: records, found_count, offset, and limit.

Routes should reference a named, validated target rather than an arbitrary URL.
Credentials should come from an encrypted server-side store or environment
reference, never from query parameters. Mutations should eventually use an
idempotency key and expected modification ID so stale updates return a conflict
instead of overwriting newer FileMaker data.
