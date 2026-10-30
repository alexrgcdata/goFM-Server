# goFM Server — Alex’s practical guide

Use this as the primary beginner guide. The other Markdown files are deeper
reference material.

## Two separate projects

```text
C:\Users\agsei\Desktop\React\goFM-server\
├── web\                 React admin panel source
│   └── dist\            compiled browser files to upload
├── bin\gofm-server      compiled Linux Go program; not public
├── config.example.json   example configuration
├── data\                 SQLite metadata at runtime
├── logs\                 encrypted request log at runtime
└── secrets\              encrypted FileMaker credentials at runtime
```

```text
OpenBridge: https://worksitebuddy.com/openbridge/
goFM panel: https://worksitebuddy.com/openbridge/go/
goFM API:   https://worksitebuddy.com/openbridge/go/api/
```

OpenBridge remains separate. React is public static files; Go is a private
background service. FTP upload does not start a Go executable.

## What works now

- Overview, Routes, Metrics, Logs, and Settings screens.
- FileMaker and Other Services route subnavigation.
- Data API/OData choices and operation permissions.
- AES-256-GCM encrypted FileMaker credential vault.
- Bounded 0–100 transaction logs and response previews.
- SQLite stores metadata; response previews stay in encrypted logs.
- Hooks are metadata only; no scripts or commands execute.
- Go tests, Go vet, Windows/Linux builds, and React production build pass.

## The `?v=redesign` URL

You do not need it. It was only a cache-buster after uploading. Use:

```text
https://worksitebuddy.com/openbridge/go/
```

## Build React first

React is the part that uses pnpm. Run this from the web folder:

```powershell
cd C:\Users\agsei\Desktop\React\goFM-server\web
$env:VITE_APP_BASE_PATH = "/openbridge/go/"
$env:VITE_API_BASE_URL = "/openbridge/go/api"
pnpm install --frozen-lockfile
pnpm run build
Remove-Item Env:VITE_APP_BASE_PATH, Env:VITE_API_BASE_URL
```

Output folder:

```text
C:\Users\agsei\Desktop\React\goFM-server\web\dist\
```

Upload the contents of `dist` to `/openbridge/go/`:

```text
web\dist\index.html  → /openbridge/go/index.html
web\dist\assets\     → /openbridge/go/assets/
```

Never create `/openbridge/go/dist/` or `/openbridge/go/web/dist/`. Press
`Ctrl + Shift + R` once after upload if an old build appears.

## Build Go separately

Go does not use pnpm, npm, or a `dist` folder. It compiles into one executable:

```powershell
cd C:\Users\agsei\Desktop\React\goFM-server
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -o .\bin\gofm-server .\cmd\gofm-server
Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED
```

Linux program:

```text
C:\Users\agsei\Desktop\React\goFM-server\bin\gofm-server
```

Do not upload `gofm-server.exe` to Linux hosting or put Go files in public
`/openbridge/go/`.

## Local test

Terminal 1, from the project root:

```powershell
go run .\cmd\sample-upstream -addr :3001 -name demo-service
```

Terminal 2, from the project root:

```powershell
New-Item -ItemType Directory -Force logs, data, secrets
$env:GOFM_APP_TOKEN = "local-app-token"
$env:GOFM_ADMIN_TOKEN = "local-admin-token"
$env:GOFM_LOG_KEY = "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcYGRobHB0eHyA="
$env:GOFM_VAULT_KEY = "ICEiIyQlJicoKSorLC0uLzAxMjM0NTY3ODk6Ozw9Pj8="
go run .\cmd\gofm-server -config .\config.example.json
```

Terminal 3:

```powershell
cd C:\Users\agsei\Desktop\React\goFM-server\web
pnpm run dev
```

Open `http://localhost:5173/`. In Settings use URL `http://localhost:8080`
and administrator token `local-admin-token`.

## Authentication and credentials

```text
GOFM_APP_TOKEN    normal application requests
GOFM_ADMIN_TOKEN  routes, credentials, logs, and settings
GOFM_VAULT_KEY    encrypts FileMaker credentials
GOFM_LOG_KEY      encrypts request logs
```

Keep application and administrator tokens different. Never put these tokens,
FileMaker passwords, API keys, or FileMaker session tokens in React,
localStorage, query parameters, or public PHP responses.

## Connecting OpenBridge later

```text
OpenBridge browser
    → OpenBridge PHP endpoint
    → Go using GOFM_APP_TOKEN server-side
    → FileMaker using the encrypted vault
    → normalized JSON response
```

OpenBridge PHP will eventually call a specific Go endpoint such as
`/openbridge/go/api/filemaker/execute`. React never receives the Go token or
FileMaker credentials.

## What Go hosting requires

The host must support a persistent Linux Go process, environment variables,
private writable storage, and a reverse proxy. Look for `Application Manager`,
`Setup Go App`, `Passenger Applications`, `Run Background Process`, or `SSH`.

If the account only provides FTP/PHP, it can host React but cannot run Go. Use
a Go-capable host or a small VPS. The private runtime contains:

```text
gofm-server
config.json
data/
logs/
secrets/
```

Bind Go to `127.0.0.1:8080`; HTTPS belongs at the reverse proxy. Configure the
four environment variables in the process manager, not in FTP or config JSON.

## Tomorrow’s checklist

1. Confirm the normal panel URL shows the redesign.
2. Confirm the host can run a persistent Linux Go process.
3. Build `bin\gofm-server` if it can.
4. Create the host-provided private Go application directory.
5. Upload only the Go executable and production `config.json` there.
6. Configure the four environment variables in the process manager.
7. Start Go on `127.0.0.1:8080` and reverse proxy `/openbridge/go/api/`.
8. Test `/health`.
9. Add one read-only FileMaker connection.
10. Connect OpenBridge PHP using `GOFM_APP_TOKEN` server-side.
11. Test one success, one failure, and one encrypted log entry.

## Learn the code

- `cmd/gofm-server/main.go` starts Go.
- `internal/server/server.go` authenticates and routes requests.
- `internal/server/config.go` validates configuration.
- `internal/credentials/vault.go` encrypts credentials.
- `internal/filemaker/dataapi.go` and `odata.go` are adapters.
- `internal/storage/storage.go` stores SQLite metadata.
- `internal/server/logging.go` stores encrypted logs.
- `_test.go` files are executable examples.

Use `gofmt` to format Go, `go test ./...` to run tests, and `go vet ./...` to
check common mistakes.
