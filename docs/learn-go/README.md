# Learn Go through goFM Server

Read these lessons in order, then change one small thing and run the tests.

1. `01-request-flow.md` — how a request enters net/http and leaves through a fixed route.
2. `02-types-and-config.md` — structs, JSON tags, slices, maps, and validation.
3. `03-auth-and-logging.md` — bearer auth, request IDs, bounded logs, and AES-GCM persistence.
4. `04-tests-and-tools.md` — gofmt, go test, go vet, httptest, and race testing.
5. `05-filemaker-boundary.md` — interfaces, adapters, normalization, and safe extension.

The executable starts at `cmd/gofm-server/main.go`. Gateway behavior is in
`internal/server/`. The FileMaker integration seam is in `internal/filemaker/`.
