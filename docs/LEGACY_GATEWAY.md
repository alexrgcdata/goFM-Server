# goFM Server

This is a historical description of the older gateway. Start with the
[current README](../README.md) for the supported local dispatcher. The retained
gateway has a separate entry point and configuration; its React dashboard does
not test the current dispatcher. Retained tests do not establish connected
FileMaker acceptance or production readiness.

A deliberately small Go server scaffold for receiving REST/curl requests,
authenticating protected endpoints, and routing them to fixed HTTP targets.
A React JSX example client lives in `web/`.

This project starts with a lean, local-development design. It will not proxy
arbitrary user-provided URLs. To call PHP, place PHP behind a local web server
or FastCGI bridge and configure that bridge as the route's fixed HTTP target.

## Start locally

1. Copy `config.legacy.example.json` to ignored `config.legacy.local.json` and configure its required token/key
   environment variables and approved upstreams. Its `cors_origins` entry
   already permits the bundled React development client at
   `http://localhost:5173`.
2. The separate legacy command is `go run ./cmd/legacy-gateway -config config.legacy.local.json`.
3. Check the public health endpoint:

   ```sh
   curl http://localhost:8080/health
   ```

4. Call a configured endpoint with your token:

   ```sh
   curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/echo
   ```

Configured routes are protected whenever `tokens` contains at least one value.
The server never forwards that bearer token to the configured upstream.

## Project layout

- `cmd/legacy-gateway/`: legacy executable entry point.
- `internal/server/`: configuration, auth, route matching, and HTTP proxy.
- `web/`: standalone Vite + React JSX request client.
- `docs/`: scope and delegation notes.

The React example makes browser requests directly to the server. For local use,
add only the exact browser origin you use to `cors_origins`. The server rejects
wildcards and origins containing paths, queries, credentials, or fragments.

## Learn the project

- [Project guide](docs/PROJECT_GUIDE.md): architecture, PHP/FastCGI bridge
  approach, configuration, and safe extension rules.
- [Go guide](docs/GO_GUIDE.md): packages, structs, methods, handlers, errors,
  concurrency, and tests as used in this codebase.
- [React client guide](web/README.md): JSX syntax, state, browser CORS, and
  the example admin/testing screen.
- [Orchestration brief](docs/ORCHESTRATION.md): deliberately small scope and
  future-agent responsibilities.
