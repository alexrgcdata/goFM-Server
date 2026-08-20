# Local test guide

The fastest local test uses three terminals. Run all commands from the
repository root unless a `cd` command says otherwise.

Before the first run, download Go dependencies once:

`go mod download`

## Terminal 1: sample upstream A

`go run ./cmd/sample-upstream -addr :3001 -name inventory-service`

## Terminal 2: sample upstream B

`go run ./cmd/sample-upstream -addr :3002 -name crm-service`

## Terminal 3: goFM gateway and storage

PowerShell first creates an encrypted log key and a local log directory:

`New-Item -ItemType Directory -Force logs`

`$env:GOFM_LOG_KEY = [Convert]::ToBase64String((1..32 | ForEach-Object { [byte](Get-Random -Maximum 256) }))`

`go run ./cmd/gofm-server -config ./config.json`

The local config writes encrypted request history to `logs/requests.enc` when
`GOFM_LOG_KEY` remains set in that terminal.

It also creates the SQLite database at `data/gofm.sqlite`. SQLite stores
request-history metadata for the demo; credentials are not stored there yet.

## Terminal 4: frontend

`cd web`

`npm install`

`npm run dev`

Open `http://localhost:5173`, enter bearer token `pass`, and use:

- `/api/echo` — routes to sample upstream A;
- `/api/inventory` — routes to sample upstream B;
- `/api/failure` — deliberately routes to an unavailable port to test failures.

The frontend also includes a Route builder. It creates reviewed JSON drafts with
authentication and before/after hook names. In this demo it does not mutate the
running config; copy reviewed route JSON into `config.json`, then restart Go.

The OpenBridge coding-help panel has Data API and OData tabs. Each tab has Find
records, Get record, and Create record starter buttons. The code window uses
dummy values and comments so a developer can adapt it safely.

Use `POST` with JSON such as `{"sku":"DEMO-001","fireGoHookAfter":true}`
to verify that the hook request flag is captured. It is logged but does not
execute anything.

The admin endpoints are protected with the same local token `pass`:

- `http://localhost:8080/__gofm/overview`
- `http://localhost:8080/__gofm/routes`
- `http://localhost:8080/__gofm/logs`
- `http://localhost:8080/__gofm/storage/logs`

## Stop and reset

Press `Ctrl+C` in each Go terminal. To clear the test history, stop the gateway
and remove `logs/requests.enc`; it will be recreated on the next request.
To reset persistent request history, stop Go and remove `data/gofm.sqlite`.

## Deployment note

React only needs the output of `npm run build` uploaded to a public web folder.
The Go binary must run as a persistent process on a VPS or a host with Go app
support. Shared hosting that only supports FTP/PHP cannot run the gateway.
