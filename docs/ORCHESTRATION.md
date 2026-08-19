# goFM Server orchestration brief

## Product boundary

`goFM Server` is a small Go HTTP server scaffold. It accepts REST-style HTTP(S)
requests, authenticates protected routes with bearer tokens, and dispatches a
request to a configured HTTP target. It is intentionally not a hosted platform,
database, user-management system, or PHP runtime.

## Near-term architecture

- Go standard library server (`net/http`) with no required third-party runtime.
- JSON configuration for routes and tokens.
- A health endpoint and an authenticated example endpoint.
- Dispatch adapters:
  - `http`: proxy a request to another HTTP service (for a React/Node API).
  - PHP can be placed behind a local web server/FastCGI bridge and used as a
    fixed `http` target. Direct PHP process execution is deferred to avoid
    turning request configuration into command execution.
- A separate JSX React example client. It is not served or built by Go yet.

## Delegated roles

| Role | Scope | Done when |
| --- | --- | --- |
| Backend implementer | `cmd/`, `internal/`, Go configuration and tests | Server starts, routes authenticate, proxy dispatch has tests. |
| React example implementer | `web/` only | JSX client illustrates a bearer-token request and has setup notes. |
| Orchestrator | repository layout, docs, integration, review | Changes are reconciled, formatted, tested, and ready to commit. |

## Working rules

- Keep the dependency surface small and prefer the Go standard library.
- Never store a real token or private key in the repository.
- Add new capabilities as small route targets, not a general plugin system.
- Validate with `go test ./...`, `go vet ./...`, and a curl smoke test.

## Planned increments

1. Bootstrap server, configuration, health route, bearer-token authentication.
2. Add safe dispatch targets (HTTP proxy, PHP handoff) and route tests.
3. Add a React JSX client example and document local development.
4. Create the GitHub repository, commit the scaffold, and push it after the
   GitHub connection is available.
