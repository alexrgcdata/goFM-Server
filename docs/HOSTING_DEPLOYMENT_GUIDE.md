# goFM hosting and FTP deployment guide

This guide deploys the standalone Go middleware without changing or replacing
OpenBridge.

## Start here: the exact answer for the current WorksiteBuddy FTP account

Your local project is:

```text
C:\Users\agsei\Desktop\React\goFM-server\
```

Your existing server folder is:

```text
/openbridge/
```

The Go demo dashboard belongs in a new folder inside it:

```text
/openbridge/go/
```

The public URLs are therefore:

```text
Existing OpenBridge: https://worksitebuddy.com/openbridge/
Go demo dashboard:   https://worksitebuddy.com/openbridge/go/
```

### Build the two parts

The Go backend has no `dist` folder. It compiles to one Linux executable:

```powershell
cd C:\Users\agsei\Desktop\React\goFM-server

$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -o .\bin\gofm-server .\cmd\gofm-server
Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED
```

The React demo dashboard does have a `dist` folder:

```powershell
cd C:\Users\agsei\Desktop\React\goFM-server\web

$env:VITE_APP_BASE_PATH = "/openbridge/go/"
$env:VITE_API_BASE_URL = "/openbridge/go"
pnpm install --frozen-lockfile
pnpm run build
Remove-Item Env:VITE_APP_BASE_PATH, Env:VITE_API_BASE_URL
```

### What to FTP now

In the FTP program, open this local folder:

```text
C:\Users\agsei\Desktop\React\goFM-server\web\dist\
```

On the server, open `/openbridge/` and create a folder named `go`.

Upload the **contents** of the local `dist` folder into `/openbridge/go/`:

```text
Local                                                     Server
-----                                                     ------
web\dist\index.html               -> /openbridge/go/index.html
web\dist\assets\                  -> /openbridge/go/assets/
```

Do not upload the `dist` folder itself. The result must not be
`/openbridge/go/dist/index.html`.

After uploading, this page should open:

```text
https://worksitebuddy.com/openbridge/go/
```

This publishes the dashboard only.

### What not to FTP into `/openbridge/go/`

Do not upload these into the public dashboard folder:

```text
C:\Users\agsei\Desktop\React\goFM-server\bin\gofm-server
C:\Users\agsei\Desktop\React\goFM-server\config.json
```

The Linux executable will not start merely because it was uploaded by FTP. A
hosting application manager or persistent process runner must start it. Until
that exists, the uploaded dashboard will display, but its live health, route,
and log requests cannot reach a hosted Go backend.

### The decision point for the backend

Look in the hosting control panel for one of these features:

```text
Application Manager
Setup Application
Setup Go App
Passenger Applications
Run Background Process
```

If none exists, ask hosting support the question in section 2. If the account
does not support persistent Go processes, the backend needs a small VPS or
Go-capable host. There is no additional FTP upload that can make a normal PHP
shared-hosting account execute the Go binary.

## 1. The two separate locations

OpenBridge remains in its existing public directory:

```text
/public_html/openbridge/
```

The Go application belongs in a private directory outside `public_html`:

```text
/home/HOSTING_USERNAME/private/gofm-openbridge/
```

The intended public URL is:

```text
https://worksitebuddy.com/openbridge/go/
```

That URL is created by a reverse-proxy rule. It is not a folder containing the
Go executable. Do not upload the Go executable or its configuration into
`/public_html/openbridge/go/`.

## 2. Confirm the host can run Go before uploading

FTP alone cannot run a Go server. The hosting account must provide:

- a persistent application/process runner;
- SSH or a control-panel application manager;
- persistent environment variables/secrets;
- a reverse proxy from `/openbridge/go/` to `127.0.0.1:8080`;
- private writable storage outside `public_html`.

Ask hosting support this exact question:

```text
Can my account run a persistent Linux x86-64 Go executable bound to
127.0.0.1:8080, configure persistent environment variables for it, and reverse
proxy https://worksitebuddy.com/openbridge/go/ to that local process while
preserving the /openbridge/go path prefix?
```

If the answer is no, do not upload the private Go application to `public_html`.
Use a small VPS or Go-capable application host instead. OpenBridge can remain on
the current host and its PHP API can call the separate Go host over HTTPS.

## 3. Build the Linux executable locally

Open PowerShell in the project root:

```powershell
cd C:\Users\agsei\Desktop\React\goFM-server

$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

go build -o .\bin\gofm-server .\cmd\gofm-server

Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED
```

The deployment executable is:

```text
C:\Users\agsei\Desktop\React\goFM-server\bin\gofm-server
```

Do not upload `gofm-server.exe`; that is the Windows build.

## 4. Prepare the production configuration

Copy the hosted example locally:

```powershell
Copy-Item `
  .\config.worksitebuddy.example.json `
  .\bin\config.json
```

Edit `bin/config.json` when adding named FileMaker connections. Keep these
hosted settings:

```json
{
  "address": "127.0.0.1:8080",
  "base_path": "/openbridge/go",
  "token_env": "GOFM_APP_TOKEN",
  "admin_token_env": "GOFM_ADMIN_TOKEN"
}
```

Do not type token values, encryption keys, or FileMaker credentials into
`config.json`.

## 5. Upload exactly these two files

Using SFTP or the hosting file manager, create:

```text
/home/HOSTING_USERNAME/private/gofm-openbridge/
```

Upload:

```text
Local                                                    Server
-----                                                    ------
bin/gofm-server       -> /home/HOSTING_USERNAME/private/gofm-openbridge/gofm-server
bin/config.json       -> /home/HOSTING_USERNAME/private/gofm-openbridge/config.json
```

Create these empty server directories:

```text
/home/HOSTING_USERNAME/private/gofm-openbridge/data/
/home/HOSTING_USERNAME/private/gofm-openbridge/logs/
/home/HOSTING_USERNAME/private/gofm-openbridge/secrets/
```

Recommended permissions:

```text
gofm-server       700
config.json       600
data/             700
logs/             700
secrets/          700
```

Do not upload anything from this Go project into
`/public_html/openbridge/`.

## 6. The four required server environment variables

These are secret values configured in the hosting process manager. They are
not files to FTP and they do not belong in `config.json`.

```text
GOFM_APP_TOKEN
GOFM_ADMIN_TOKEN
GOFM_VAULT_KEY
GOFM_LOG_KEY
```

### What each value does

`GOFM_APP_TOKEN`

- Authenticates normal API requests.
- OpenBridge's PHP backend will eventually use it when calling Go.
- Never place it in React, browser JavaScript, localStorage, or a public PHP
  response.

`GOFM_ADMIN_TOKEN`

- Authenticates private administrative endpoints such as storing FileMaker
  credentials and viewing administrative logs.
- Must be different from `GOFM_APP_TOKEN`.
- Use it only from localhost, SSH, a protected server-side tool, or an SSH
  tunnel. Do not put it in a public admin webpage.

`GOFM_VAULT_KEY`

- A base64-encoded 32-byte AES key.
- Encrypts and decrypts `secrets/credentials.enc`.
- Losing it makes the credential vault unreadable.
- Changing it requires recreating the credential vault.

`GOFM_LOG_KEY`

- A separate base64-encoded 32-byte AES key.
- Encrypts and decrypts `logs/requests.enc`.
- Must not reuse the vault key.

## 7. Generate production secrets

Generate them in PowerShell. Store them immediately in a password manager or
the hosting secret manager. Do not paste them into chat, email, Git, or public
files.

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

$productionSecrets = [ordered]@{
    GOFM_APP_TOKEN   = New-SecureBase64Key
    GOFM_ADMIN_TOKEN = New-SecureBase64Key
    GOFM_VAULT_KEY   = New-SecureBase64Key
    GOFM_LOG_KEY     = New-SecureBase64Key
}
```

Copy each value directly into the matching field in the hosting service's
environment-variable or secret settings. Do not save `$productionSecrets` to a
file inside the project.

The values generated earlier for `go run` exist only in that PowerShell
process. Production needs its own persistent copies in the server process
manager.

## 8. Where to set the variables on the server

Use the first option your host supports.

### Hosting control-panel application manager

Open the Go application's environment-variable settings and add all four names
and values. The application command should be:

```text
/home/HOSTING_USERNAME/private/gofm-openbridge/gofm-server
```

Arguments:

```text
-config /home/HOSTING_USERNAME/private/gofm-openbridge/config.json
```

Working directory:

```text
/home/HOSTING_USERNAME/private/gofm-openbridge
```

Restart the application after adding or changing environment variables.

### systemd on a VPS

Store the four variables in a root/service-readable environment file outside
`public_html`, for example:

```text
/etc/gofm-openbridge.env
```

The file contains:

```text
GOFM_APP_TOKEN=REPLACE_WITH_SECRET
GOFM_ADMIN_TOKEN=REPLACE_WITH_DIFFERENT_SECRET
GOFM_VAULT_KEY=REPLACE_WITH_BASE64_32_BYTE_KEY
GOFM_LOG_KEY=REPLACE_WITH_DIFFERENT_BASE64_32_BYTE_KEY
```

Set its permissions to `600`. A systemd unit can use:

```ini
[Unit]
Description=goFM OpenBridge middleware
After=network-online.target

[Service]
Type=simple
User=HOSTING_USERNAME
WorkingDirectory=/home/HOSTING_USERNAME/private/gofm-openbridge
EnvironmentFile=/etc/gofm-openbridge.env
ExecStart=/home/HOSTING_USERNAME/private/gofm-openbridge/gofm-server -config /home/HOSTING_USERNAME/private/gofm-openbridge/config.json
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

This option requires SSH and usually administrator/sudo access.

## 9. Reverse proxy

The hosted configuration requires the proxy to preserve the path prefix:

```text
Public:  https://worksitebuddy.com/openbridge/go/health
Private: http://127.0.0.1:8080/openbridge/go/health
```

Ask the host to block public access to:

```text
/openbridge/go/__gofm/*
```

Administrative calls should be performed through localhost or an SSH tunnel.
Normal application calls use `/openbridge/go/api/...`.

Do not expose port `8080` directly to the internet.

## 10. Start and test on the server

After the process manager starts Go, test the public health endpoint:

```text
https://worksitebuddy.com/openbridge/go/health
```

Expected response:

```json
{"status":"ok"}
```

From an SSH terminal on the server, test an authenticated endpoint without
putting the token in a URL:

```bash
curl -H "Authorization: Bearer $GOFM_APP_TOKEN" \
  http://127.0.0.1:8080/openbridge/go/api/echo
```

The example hosted configuration has no generic HTTP routes, so add only the
specific routes or named FileMaker connections you intend to expose.

## 11. Store FileMaker credentials after deployment

From the server itself, call the administrator endpoint over loopback:

```bash
curl -X PUT \
  -H "Authorization: Bearer $GOFM_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  --data '{"username":"FILEMAKER_API_USER","password":"FILEMAKER_API_PASSWORD"}' \
  http://127.0.0.1:8080/openbridge/go/__gofm/credentials/primary-fm
```

Go encrypts the credential into:

```text
/home/HOSTING_USERNAME/private/gofm-openbridge/secrets/credentials.enc
```

The username and password are not returned to the browser and should never be
placed in OpenBridge JavaScript.

## 12. OpenBridge connection boundary

Do not add the Go application token to React.

The future connection should be:

```text
OpenBridge browser
    -> authenticated OpenBridge PHP endpoint
    -> Go using GOFM_APP_TOKEN server-side
    -> FileMaker using the encrypted credential vault
```

That PHP connector is a later integration step. The standalone Go middleware
can be deployed and tested independently first.
