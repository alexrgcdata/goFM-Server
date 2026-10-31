# goFM Server handoff

Updated October 30, 2026. The basic local demo is complete. This resume performed
a lightweight receipt/remote check and documented the state; no runtime changes,
installs, test reruns, builds, servers, or monitors were needed.

## Ownership and source

Work only in `C:\AlexSeidler\projects\goFM-Server`. Read `AGENTS.md` before
editing. The parent portfolio and its deployment belong to the project manager
chat. The sibling Node server and the FileMaker schema explorer are separate
products. Do not copy their code into this checkout.

- [Public repository](https://github.com/alexrgcdata/goFM-Server)
- [Source ZIP](https://github.com/alexrgcdata/goFM-Server/archive/refs/heads/main.zip)
- [Last checked implementation and documentation revision](https://github.com/alexrgcdata/goFM-Server/commit/f8a5092c1e2694323b8e6b0ca95978431a943d88)
- [Completed CI for that revision](https://github.com/alexrgcdata/goFM-Server/actions/runs/41142215794)
- [Published portfolio showcase](https://alexseidler.dev/work/secure-go-api-router/)

At the start of this resume, local `main` and the remote `main` both pointed to
`f8a5092c1e2694323b8e6b0ca95978431a943d88`, and the working tree was clean.
This handoff is a subsequent documentation-only commit. There is no versioned
binary release; the ZIP link downloads source. Preserve visibility, licensing,
and existing history.

## What works

The default command is `cmd/gofm-server`, backed by `internal/api`. It uses Go's
standard library without a runtime database or external service. `GET /health`
is public. `POST /api` authenticates a private config bearer token before JSON
parsing or function dispatch.

Requests contain an `action`, a required `params` array, and optional RFC3339
client `time`. The registered demo actions are `hello`, `success`, `boolean`, and
`echo`. Their results are respectively `"hello world"`, `"successful"`, JSON
`true`, and the supplied parameter array. The response includes success, result,
an error object or null, a request identifier, and server UTC time.

The commented action registry is in `internal/api/api.go`; implementations are
in `internal/api/functions.go`. Request size/type/parameter checks, safe errors,
listener timeouts, and graceful shutdown are implemented. The token generator
does not print the token or overwrite an existing file.

## First run

The module declares Go 1.25; recorded verification used Go 1.26. From the root of
a downloaded or cloned checkout:

```sh
go run ./cmd/gofm-server -init-config
go run ./cmd/gofm-server
```

The first command creates ignored `config.local.json`. Keep it private. The
default listener is `127.0.0.1:8080`. Run `./scripts/demo.ps1` in a second
PowerShell terminal, or check `/health` with a browser/API client. A previously
created config needs no regeneration. Do not assume the pre-shutdown local
process is still running.

For Go, there is no `dist` folder. The documented Windows build produces
`C:\AlexSeidler\projects\goFM-Server\bin\gofm-server.exe`. A Linux cross-build
was also checked. Executables/config are not distributed by this source handoff.
A hosted service would require an approved private application directory outside
the document root and HTTPS; no server destination or deployment is configured.

## Verification already performed

Before the pause, local `go test ./...`, `go vet ./...`, Windows builds, and a
Linux cross-build passed. Real local HTTP checks verified all four demo results
and missing-token rejection with HTTP 401. The demo script passed again during
the README review. GitHub CI for `f8a5092` passed formatting, vet, the full test
suite, dispatcher/CLI race checks, and Linux/Windows builds. The CI conclusion
was confirmed as successful during this resume without rerunning those checks.

See `READINESS.md` and `README.md` for further evidence and commands. The two
images in `docs/images/` present actual captured local API output with synthetic
input and a redacted token. They are designed illustrations of that evidence,
not a shipped dashboard or a public hosted backend.

## Limits and next steps

The current product is a local demo with one shared token. It has no active
FileMaker integration, per-user permissions, persistent jobs, automatic retries,
public Go backend, or independent production-security audit. Legacy gateway,
FileMaker adapters, vault/storage packages, and the React dashboard remain
separate reference work; `cmd/legacy-gateway` uses a different config schema.
Legacy dependencies remain in `go.mod`, but the default dispatcher does not
import them. Do not advertise legacy features as current dispatcher behavior.

Alex's proposed FileMaker scope is small: select a configured OData connection,
deliver a JSON package to the requested permitted database/table/field, then
optionally call a named approved script with the package and record context.
Those choices belong in the request; host credentials stay in private config.
The new default dispatcher does not implement this yet. The retained OData
adapter currently rejects `script_after`; it needs deliberate extension and
protocol tests. Consult current Claris OData documentation and verified metadata
before constructing real entity/key URLs. Distinguish delivery success from
script failure, and do not automatically retry a completed write.

No production database writes, live FileMaker calls, or deployment are authorized
by this handoff. Future implementation should start with synthetic data and a
local test service. There is no remaining work required for the completed generic
demo. The OData delivery proposal is the next separate implementation decision.

Use one worker for this product. Do not launch duplicate servers/watchers or
rerun unchanged suites. Coordinate any expensive checks with the project manager
and record actual machine observations if heavy work is needed. This resume did
not require heavy work or resource monitoring.
