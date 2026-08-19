# goFM Server project guide

## What this scaffold does

goFM Server is a small authentication boundary and request router. A caller
uses curl, another HTTP client, or the separate React example client to make a
request. The Go service verifies a bearer token and forwards the request to a
known local service.

```text
curl / React client
        |
        v
goFM Server -- verifies token, matches method + path
        |
        v
fixed local HTTP service (Node, PHP through FastCGI/web server, etc.)
```

The configured target is fixed in `config.json`. It must never be taken from a
query parameter, header, or request body.

## PHP and FastCGI

Go is intentionally not launching a `php` process per request. Use a local
PHP-compatible web server or FastCGI bridge, then configure its address as an
`http` route target. This keeps process ownership, timeouts, and PHP file
selection outside a public request router.

For example, a local bridge listening at `http://127.0.0.1:9001` can receive
the following safely fixed route:

```json
{
  "path": "/api/php-example",
  "methods": ["GET", "POST"],
  "target": { "type": "http", "url": "http://127.0.0.1:9001" }
}
```

The Go server forwards `/api/php-example` to that local bridge. The bridge,
not the caller, decides which PHP application code handles it.

## Configuration principles

- Keep real `config.json` out of Git; start with `config.example.json`.
- Use a long random token, not the sample value.
- Keep upstream services on loopback/private networking where possible.
- Add an exact browser origin only when the React admin client needs it.
- Prefer narrowly scoped routes such as `/api/orders` over a catch-all route.

## Instructions for future agents

1. Preserve the fixed-target routing model and JSON error shape.
2. Add a focused unit test for each auth, CORS, or route behavior change.
3. Do not add a database, identity provider, or user-registration flow unless
   the project owner explicitly requests it.
4. Do not execute caller-provided commands, scripts, or URLs.
5. Run `gofmt`, `go vet ./...`, and `go test ./...` before publishing changes.
