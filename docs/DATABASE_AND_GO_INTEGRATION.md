# Storage options and Go integration

goFM OpenBridge uses Go as the always-on boundary between web clients and
FileMaker. React provides the management screen; Go owns routing,
authentication, request logging, storage, and future FileMaker adapters.

## Current demo storage

The demo uses SQLite through the pure-Go modernc.org/sqlite driver. The file is
configured with storage.db_file, normally data/gofm.sqlite.

SQLite is a strong first choice when one server owns the application. It needs
no database service, keeps data in a portable file, is fast for bounded request
history and route metadata, and deploys cleanly on a Linux VPS.

The current SQLite store holds request-history metadata. Sensitive response
previews remain in the AES-GCM encrypted log file when log encryption is
enabled. Do not put FileMaker passwords or API keys into plaintext SQLite rows.

## Credential storage path

The next security increment should use an encrypted credential vault. Keep the
32-byte master key in a host environment variable or cloud key service, never
in config.json or Git. Encrypt each FileMaker credential with AES-256-GCM,
associate it with a credential ID, decrypt only immediately before an upstream
request, and never return values through admin endpoints, logs, or the browser.

## When to choose another database

| Need | Recommended choice | Why |
| --- | --- | --- |
| Single VPS or demo | SQLite | Simplest deployment and backup. |
| Multiple gateway instances | PostgreSQL | Concurrent writes and shared reporting. |
| Cloud secret management | PostgreSQL plus KMS/Vault | Rotation and audit controls. |
| Expiring cache or rate limits | Redis beside SQLite/PostgreSQL | Fast counters and expiry. |

Start with SQLite. Move to PostgreSQL once several Go processes need shared
state or production reporting and job queues require a centralized database.

## Go and FileMaker model

Web clients call OpenBridge. Go verifies their bearer token, applies route
policy, writes a request ID and log entry, then calls a named FileMaker Data API
or OData target. FileMaker credentials stay server-side. Browser code receives
normalized JSON and never sees a FileMaker password or session token.

## Safe Data API and OData rules

- A route references a named target, never a browser-supplied URL.
- A target fixes its FileMaker host, database, layout, table, and credential.
- Clients submit allowed values, filters, paging, and fields only.
- The server validates fields and operations before building API requests.
- Both transports normalize to records, found_count, offset, and limit.
- Mutations should use idempotency and expected modification IDs.

For the local commands, use LOCAL_TEST_GUIDE.md in this folder.
