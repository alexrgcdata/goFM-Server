# goFM Middleware: implementation, deployment, and developer guide

## 1. The permanent project boundary

OpenBridge and goFM are separate applications.

- **OpenBridge** is the existing frontend and PHP project at
  `https://worksitebuddy.com/openbridge/`.
- **goFM** is a standalone Go middleware service. It authenticates callers,
  keeps FileMaker credentials server-side, calls FileMaker, and returns safe,
  normalized JSON.
- OpenBridge may call goFM, but neither project is built into or copied over
  the other.

The public path `/openbridge/go/` is only a URL routed by the web server. It
does not mean the private Go executable belongs inside the OpenBridge files.

```text
Browser -> authenticated OpenBridge PHP endpoint
      |
      | server-to-server HTTPS with application bearer token
      v
https://worksitebuddy.com/openbridge/go/
      |
      | reverse proxy (Apache, LiteSpeed, or Nginx)
      v
goFM on 127.0.0.1:8080
      |
      | server-side FileMaker credentials/session token
      v
FileMaker Data API or OData
```

OpenBridge user authentication remains in OpenBridge/PHP. PHP holds the
long-lived Go application token and calls goFM server-to-server. Never embed
that token in React or browser JavaScript. Direct browser access would require
a future short-lived, user-bound token issuer and is outside this basic release.

## 2. Final server directory layout

Public OpenBridge files remain exactly where they already are:

```text
/public_html/openbridge/
    index.html
    assets/
    api/
        openbridge-api.php
```

The private Go deployment lives outside `public_html`:

```text
/home/HOSTING_ACCOUNT/private/gofm-openbridge/
    gofm-server
    config.json
    data/
        gofm.sqlite
    logs/
        requests.enc
    secrets/
        credentials.enc
```

Only a reverse-proxy mapping joins the public URL to the private process. The
proxy must preserve the `/openbridge/go` prefix because the hosted config uses
that value as its `base_path`:

```text
/openbridge/go/*  ->  http://127.0.0.1:8080/openbridge/go/*
```

Never upload the Go executable, configuration, database, logs, credential
vault, source code, or Node modules into `/public_html/openbridge/`.

## 3. What Go owns

goFM is the security boundary. It owns:

- application and administrator authentication;
- route and operation authorization;
- encrypted FileMaker credentials;
- FileMaker Data API login and session-token caching;
- OData authentication;
- request validation and size limits;
- FileMaker request construction;
- JSON response normalization;
- bounded request logging and encrypted log files;
- timeouts and rate limiting;
- optional, allow-listed FileMaker script execution.

The browser never receives a FileMaker password or FileMaker session token.
The caller chooses a configured connection name and an allowed operation—not a
hostname, credential, command, executable, PHP filename, or arbitrary URL.

## 4. Authentication layers

There are three distinct secrets:

1. **Application token**: OpenBridge or another approved client presents this
   as `Authorization: Bearer ...` when calling normal middleware routes.
2. **Administrator token**: used only for private administration endpoints,
   including storing credentials. It must differ from the application token.
3. **FileMaker credential**: stored encrypted in goFM's private vault and used
   only by the selected FileMaker adapter.

Production tokens are loaded from environment variables. JSON configuration
contains environment-variable names, never production token values. Legacy
plaintext tokens are supported only for local migration and should be removed.

Generate three independent 32-byte values in PowerShell:

```powershell
function New-SecureBase64Key {
    $keyBytes = New-Object byte[] 32
    $generator = [Security.Cryptography.RandomNumberGenerator]::Create()
    try {
        $generator.GetBytes($keyBytes)
    }
    finally {
        $generator.Dispose()
    }
    [Convert]::ToBase64String($keyBytes)
}

$env:GOFM_APP_TOKEN = New-SecureBase64Key
$env:GOFM_ADMIN_TOKEN = New-SecureBase64Key
$env:GOFM_VAULT_KEY = New-SecureBase64Key
$env:GOFM_LOG_KEY = New-SecureBase64Key
```

For a hosted service, set these in the service manager or hosting secret
facility. Do not put them in shell history, Git, React, or public PHP files.

## 5. Encrypted credential storage

Credentials are added through an administrator-only endpoint. They are
encrypted with AES-256-GCM before being written to `secrets/credentials.enc`.
The 32-byte encryption key comes from `GOFM_VAULT_KEY` and is never written next
to the vault.

A credential has a stable name such as `primary-fm`. Routes reference that
name. Normal API callers cannot submit or replace usernames and passwords.

Example administrative request from the server itself:

```powershell
$headers = @{
    Authorization = "Bearer $env:GOFM_ADMIN_TOKEN"
    "Content-Type" = "application/json"
}
$body = @{
    username = "FILEMAKER_API_USER"
    password = "FILEMAKER_API_PASSWORD"
} | ConvertTo-Json

Invoke-RestMethod -Method Put `
  -Uri http://127.0.0.1:8080/__gofm/credentials/primary-fm `
  -Headers $headers -Body $body
```

The response reports the credential name but never echoes secret values.

## 6. Configured FileMaker connections

`config.json` defines approved connections. A connection fixes the FileMaker
host and references a credential by name:

```json
{
  "filemaker_connections": [
    {
      "name": "demo-data-api",
      "adapter": "dataapi",
      "base_url": "https://filemaker.example.com",
      "credential": "primary-fm",
      "default_database": "OpenBridgeDemo",
      "allowed_databases": ["OpenBridgeDemo"],
      "allowed_layouts": ["Customers", "Orders"],
      "allowed_tables": []
    },
    {
      "name": "demo-odata",
      "adapter": "odata",
      "base_url": "https://filemaker.example.com",
      "credential": "primary-fm",
      "default_database": "OpenBridgeDemo",
      "allowed_databases": ["OpenBridgeDemo"],
      "allowed_layouts": [],
      "allowed_tables": ["Customers", "Orders"]
    }
  ]
}
```

The request can select `dataapi` or `odata`, but only through one of these
named connections. An empty allow-list means the value must be fixed by the
connection/route; it does not mean unrestricted access.

## 7. FileMaker request API

The general endpoint is:

```text
POST /api/filemaker/execute
Authorization: Bearer APPLICATION_TOKEN
Content-Type: application/json
```

Example find request:

```json
{
  "connection": "demo-data-api",
  "operation": "find",
  "database": "OpenBridgeDemo",
  "layout": "Customers",
  "query": [
    {"field": "Status", "op": "eq", "value": "Active"}
  ],
  "limit": 50,
  "offset": 1
}
```

Example OData get request:

```json
{
  "connection": "demo-odata",
  "operation": "get",
  "database": "OpenBridgeDemo",
  "table": "Customers",
  "record_id": "10"
}
```

Open-ended `fields` and query values are allowed because FileMaker records vary
by solution. Hosts, connection names, databases, layouts, tables, operations,
and mutation permissions remain allow-listed server-side.

## 8. Normalized response

Both Data API and OData return this application-level shape:

```json
{
  "records": [],
  "found_count": 0,
  "offset": 1,
  "limit": 50,
  "meta": {
    "adapter": "dataapi",
    "request_id": "..."
  }
}
```

Go maps FileMaker response envelopes into ordinary JSON objects. The original
FileMaker envelope is not returned unless a future administrator-only debug
mode explicitly enables safe metadata.

## 9. Optional FileMaker scripts

Script execution is limited to FileMaker scripts listed in configuration. A
request may ask for an approved script after a successful operation:

```json
{
  "connection": "demo-data-api",
  "operation": "find",
  "layout": "Customers",
  "script_after": {
    "name": "Rebuild Customer Cache",
    "parameter": "active"
  }
}
```

This means “run a named FileMaker script through the Data API.” It never means
execute an operating-system command, PHP file, uploaded program, or URL. The
existing `fireGoHookBefore` and `fireGoHookAfter` route fields remain inert
metadata.

## 10. Logging and local storage

The SQLite database keeps bounded request metadata. The encrypted log file uses
AES-256-GCM when `GOFM_LOG_KEY` is configured.

Logs may include request ID, route, operation, duration, status, and safe error
code. Logs must not include:

- Authorization or Cookie headers;
- FileMaker usernames or passwords;
- FileMaker session tokens;
- credential-vault contents;
- full request or response bodies by default.

Response previews are a local debugging option and should be disabled on a
server handling private FileMaker records.

## 11. Local development on Windows

Open PowerShell—not Command Prompt—and run:

```powershell
cd C:\Users\agsei\Desktop\React\goFM-server

New-Item -ItemType Directory -Force data, logs, secrets

function New-SecureBase64Key {
    $keyBytes = New-Object byte[] 32
    $generator = [Security.Cryptography.RandomNumberGenerator]::Create()
    try {
        $generator.GetBytes($keyBytes)
    }
    finally {
        $generator.Dispose()
    }
    [Convert]::ToBase64String($keyBytes)
}

$env:GOFM_APP_TOKEN = New-SecureBase64Key
$env:GOFM_ADMIN_TOKEN = New-SecureBase64Key
$env:GOFM_VAULT_KEY = New-SecureBase64Key
$env:GOFM_LOG_KEY = New-SecureBase64Key

go run .\cmd\gofm-server -config .\config.json
```

The local `echo` route also needs the included sample upstream. Start it in a
second PowerShell window:

```powershell
cd C:\Users\agsei\Desktop\React\goFM-server
go run .\cmd\sample-upstream -addr 127.0.0.1:3001 -name echo
```

Send test requests from another PowerShell window that has the same application
token value:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/health

$headers = @{ Authorization = "Bearer $env:GOFM_APP_TOKEN" }
Invoke-RestMethod -Headers $headers `
  http://127.0.0.1:8080/api/echo
```

Environment variables are scoped to their PowerShell process. For testing from
a second terminal, set the same temporary values there or start requests from
the first terminal after launching Go as a background job.

## 12. Tests and build commands

From the Go project root:

```powershell
go test ./...
go vet ./...
go build -o .\bin\gofm-server.exe .\cmd\gofm-server
```

Build a Linux x64 server binary from PowerShell:

```powershell
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -o .\bin\gofm-server .\cmd\gofm-server
Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED
```

Upload only these Go deployment artifacts to the private directory:

```text
bin/gofm-server                   -> /home/ACCOUNT/private/gofm-openbridge/gofm-server
config.worksitebuddy.example.json -> /home/ACCOUNT/private/gofm-openbridge/config.json
```

Create `data`, `logs`, and `secrets` on the server with permissions limited to
the service account. The program creates its SQLite schema and encrypted files.

## 13. Running on a VPS

The service command is:

```text
/home/ACCOUNT/private/gofm-openbridge/gofm-server \
  -config /home/ACCOUNT/private/gofm-openbridge/config.json
```

Use systemd or the host's process manager to set the four secret environment
variables and restart the process after deployment. Configure the reverse proxy
to send `/openbridge/go/` to `http://127.0.0.1:8080/openbridge/go/`, preserving
the path prefix and HTTPS-facing request information.

Do not expose `8080` in the firewall.

## 14. Shared-hosting limitation

FTP uploads cannot start a persistent Go process. LiteSpeed, Apache, PHP, and
FTP availability do not automatically mean the account supports Go.

The host must provide all of the following:

- a way to run a persistent user process;
- environment-variable or secret configuration;
- reverse-proxy support;
- private writable directories;
- restart controls or SSH access.

If WorksiteBuddy's current plan does not provide those capabilities, keep
OpenBridge there and run goFM on a small VPS or Go-capable application host.
OpenBridge's PHP API can then make server-to-server HTTPS calls to goFM.

## 15. Go syntax map for this project

Go source is organized into packages:

```text
cmd/gofm-server/       program entry point (`package main`)
internal/server/       HTTP, authentication, routing, limits, and logs
internal/filemaker/    FileMaker contracts and adapters
internal/storage/      SQLite request history
internal/credentials/  encrypted credential vault
```

Important syntax used here:

```go
type Request struct {
    Operation string `json:"operation"`
    Fields map[string]any `json:"fields,omitempty"`
}
```

`type` defines a named type. `struct` groups fields. The text in backticks is a
JSON tag: it maps Go's `Operation` to JSON's `operation`; `omitempty` leaves an
empty field out of output.

```go
func (r Request) Validate() error { ... }
```

`func` declares a function. `(r Request)` is a receiver, so `Validate` is a
method on `Request`. An `error` return is `nil` on success and non-nil on
failure.

```go
type Adapter interface {
    Execute(context.Context, Request) (Response, error)
}
```

An interface describes behavior. Data API and OData adapters can have different
internals while both satisfying this contract.

```go
if err := request.Validate(); err != nil {
    return Response{}, err
}
```

This initializes `err`, checks it, and returns early. Explicit error handling is
a central Go pattern.

```go
ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
defer cancel()
```

`context` carries cancellation and deadlines. `defer` schedules cleanup when
the current function returns.

```go
s.mu.RLock()
defer s.mu.RUnlock()
```

HTTP requests run concurrently. A mutex protects shared maps and slices from
simultaneous reads and writes.

## 16. Why Go instead of a single PHP page

PHP can communicate with FileMaker securely when credentials remain on the
server. Go is not automatically secure merely because it is compiled. The
advantage comes from this service design:

- one long-running process can reuse FileMaker sessions and HTTP connections;
- static types make route, request, and response contracts easier to audit;
- interfaces allow Data API and OData to share one normalized API;
- concurrency and cancellation are built into the standard library;
- one compiled binary has a small production dependency surface;
- centralized limits, logging, redaction, and authorization are easier to keep
  consistent across many websites.

PHP remains useful as OpenBridge's same-origin session boundary. Go becomes the
reusable FileMaker capability server behind it.

## 17. Future sample FileMaker file

A FileMaker `.fmp12` file can be offered as a download from a website, but it
cannot run as a FileMaker server merely by uploading it to web hosting. Live
Data API/OData tests require FileMaker Server or FileMaker Cloud with the sample
file hosted and the relevant APIs enabled. Building that sample is a separate
future project.
