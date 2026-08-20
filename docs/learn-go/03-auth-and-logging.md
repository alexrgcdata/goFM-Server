# 3. Authentication and logging

The gateway accepts Authorization: Bearer token. Tokens are compared with
constant-time comparison. Admin metadata uses a separate admin_tokens list;
the test config temporarily uses pass for both.

Every request gets an X-Request-ID. The in-memory log store keeps the newest
configured number of entries, defaulting to 50. It records metadata and failure
status, not authorization headers or cookies.

Optional file persistence uses AES-256-GCM. The key comes from a base64-encoded
32-byte environment variable named by logs.encryption_key_env. It is never
stored in JSON config. Enable response previews only when the data is safe to
retain.
