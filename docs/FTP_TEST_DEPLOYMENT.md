# WorksiteBuddy secondary OpenBridge gateway

This gateway is deliberately separate from the existing OpenBridge website.
It is mounted only below `/openbridge/api` and must never replace the existing
`/openbridge/index.html`.

## Public URLs

```text
https://worksitebuddy.com/openbridge/             existing site: unchanged
https://worksitebuddy.com/openbridge/gofm-admin/  this project's admin panel
https://worksitebuddy.com/openbridge/api/         Go gateway only
```

## Upload only these files

Build the admin panel with `pnpm build` from `web`. Upload the *contents* of
`web/dist` to `public_html/openbridge/gofm-admin/`. Its index.html is separate
from the existing OpenBridge index.html.

Build the Linux Go binary from the project root with `GOOS=linux`,
`GOARCH=amd64`, `CGO_ENABLED=0`, then `go build -o gofm-server ./cmd/gofm-server`.

Upload `gofm-server` and a private copy of `config.worksitebuddy.example.json`
renamed as `config.json` to `/home/YOUR-USER/gofm-openbridge/`, outside
`public_html`. Do not upload source code, node_modules, data, or logs.

## Required proxy rule

The host must forward `/openbridge/api` to the long-running Go program on
`127.0.0.1:8080` **without removing the path prefix**. The server's
`base_path` setting keeps every other WorksiteBuddy URL out of the gateway.
FTP alone cannot start or proxy a Go program; SSH plus a service manager and
reverse proxy (or a VPS) is required.
