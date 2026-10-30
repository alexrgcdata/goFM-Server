# goFM Server readiness receipt

Verified October 30, 2026.

- Working folder: C:\AlexSeidler\projects\goFM-Server
- Public source: https://github.com/alexrgcdata/goFM-Server
- Runtime implementation checked: 2f67c9785045a56a422e998566a2beef354ad684
- Latest source-and-visuals revision checked before this documentation review: 2ac7ec4c16ebef827c38677726bc438ed7f1396e
- Source download: https://github.com/alexrgcdata/goFM-Server/archive/refs/heads/main.zip (or GitHub Code > Download ZIP). No versioned binary release published.
- Local working tree was clean following push.
- Startup from repository root with Go 1.25+: `go run ./cmd/gofm-server -init-config`, then `go run ./cmd/gofm-server`.
- `-init-config` creates ignored config.local.json with a random 256-bit token, loopback listener, and 64 KiB body limit. Does not print token or overwrite an existing file.
- Live local check: `./scripts/demo.ps1`, from the same root, against the running server.
- Runtime: no database, SQLite, React, Node, FileMaker, AI provider, or external Go module is imported by the default dispatcher. Legacy packages retain their module dependencies for their own builds/tests.
- Verified locally: go test ./..., go vet ./..., Windows build and Linux cross-build; compiled Windows executable served real HTTP requests for hello, success, boolean, echo; missing bearer token rejected with 401.
- GitHub CI: https://github.com/alexrgcdata/goFM-Server/actions/runs/41141434827 passed formatting, vet, all tests, dispatcher/CLI race checks, Linux build, Windows cross-build.
- Functions: hello -> string hello world; success -> string successful; boolean -> JSON true; echo -> same dynamic JSON array.
- Contract: POST /api; Authorization Bearer token; required action + params array, optional RFC3339 client time. Auth before body parsing/dispatch. Registry: internal/api/api.go; functions: internal/api/functions.go.
- Public GET /health; typed JSON response envelope; request IDs; safe error messages; size limits; listener timeouts; graceful shutdown.
- A stranger can download public source and run locally using the README and Go toolchain. No real credentials or configured external services required. They generate their own local token.
- Limitations: local demo, not independently audited production security. One shared token, no per-user scopes, durable jobs, active FileMaker integration, or public hosted backend. Loopback by default; remote hosting needs private application storage and HTTPS. No hosting migration/deployment performed.
- Existing legacy gateway/adapters/dashboard preserved under cmd/legacy-gateway and older packages; do not advertise these as features of the new default dispatcher.

Portfolio visuals use actual local HTTP captures with synthetic input. They are designed presentations of request/response evidence, not screenshots of a shipped admin UI. Token is redacted. Capture responses are recorded in the portfolio content/go-router-captures.json.
