# Workshop admin dashboard

The dashboard shows Go health, uptime, completed `/api` requests, HTTP errors,
and average response time. Counters are held in memory and reset on restart.
Health checks and metrics reads do not affect the counters.

From the repo root, run `npm --prefix web ci` and
`npm --prefix web run build:admin`. Start the router with:

```sh
go run ./cmd/gofm-server -config config.local.json -dashboard web/dist/admin
```

Open `http://127.0.0.1:8080/admin/dashboard/`. Enter the token from your private
config to load metrics. The page keeps it in memory until you close or reload it.
Use an SSH tunnel for Workshop. Keep `/admin/` out of the public Nginx routes.
The dashboard directory should contain only the generated static files.

Start with `internal/api/metrics.go` for the Go counters and
`web/admin/main.tsx` for the React screen.
