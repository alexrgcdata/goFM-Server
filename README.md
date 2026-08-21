<<<<<<< HEAD
# goFM Server

A deliberately small Go server scaffold for receiving REST/curl requests,
authenticating protected endpoints, and routing them to fixed HTTP targets.
A React JSX example client lives in `web/`.

This project starts with a lean, local-development design. It will not proxy
arbitrary user-provided URLs. To call PHP, place PHP behind a local web server
or FastCGI bridge and configure that bridge as the route's fixed HTTP target.

## Start locally

1. Copy `config.example.json` to `config.json`, replace its sample token, and
   point the example routes at services you control. Its `cors_origins` entry
   already permits the bundled React development client at
   `http://localhost:5173`.
2. Start the server with `go run ./cmd/gofm-server -config config.json`.
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

- `cmd/gofm-server/`: executable entry point.
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
=======
# goFM Server React client example

This is a minimal JSX-only Vite client for manually exercising a goFM Server.
It is intentionally separate from the Go server and is not built or served by it.

It is a small admin/testing screen, not a login system or a production UI.

## Run locally

1. Start the Go server (normally on `http://localhost:8080`).
2. From this `web` directory, run `npm install` once and `npm run dev`.
3. Vite starts this client at `http://localhost:5173`. Keep that port free:
   the fixed address makes the CORS setting predictable.
4. Configure the Go server to allow the origin `http://localhost:5173` in its
   CORS allow-list. Do not use a wildcard when bearer tokens are involved.
5. Open the displayed browser address. The page repeats the origin that the
   browser is actually using; if it differs, allow that exact value instead.
6. Use **Check health**, then enter a configured bearer token and API endpoint
   such as `/api/echo` to make an authenticated request.

The browser sends requests directly to the Go server. Without the matching CORS
rule, the browser blocks the response even if a command-line `curl` request
works. The health request does not send your token. The API request sends an
`Authorization: Bearer ...` header only when its token field has a value.

To avoid typing the local API address each time, create a file named `.env.local`
inside `web` containing this one line:

```text
VITE_API_BASE_URL=http://localhost:8080
```

Restart the Vite development server after changing the file. This only chooses
the API address; it does not configure CORS on the Go server.

## JSX and functionality notes

This project uses JavaScript with JSX, not TypeScript. JSX lets a component
return HTML-like markup directly from a JavaScript function. For example,
`<button>Check health</button>` is JSX. `className` is used for CSS classes,
and curly braces insert JavaScript values, such as `{window.location.origin}`.

`App.jsx` is the main screen. `useState` remembers what the user typed and the
most recent response. Each input is a controlled input: its `value` comes from
state and its `onChange` handler updates that state. The two buttons call the
same `request` function with different paths and authentication needs.

The response panel shows the requested URL, HTTP status, and JSON response when
available. Network or CORS failures appear as an error message. Tokens remain
only in browser memory and disappear on refresh; do not use production secrets
in this example client.
>>>>>>> 0d8871589256cb66840deaca1805331e2759ccc6
