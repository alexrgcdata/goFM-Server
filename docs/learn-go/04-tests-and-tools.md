# 4. Tests and tools

Run these commands from the repository root:

gofmt -w (Get-ChildItem internal,cmd -Filter *.go -Recurse | ForEach-Object { $_.FullName })
go test ./...
go vet ./...

Tests use request recorders so auth, CORS, admin routes, proxy behavior,
bounded logs, and encrypted persistence can be checked without a real host.
On a machine with cgo enabled, also run go test -race ./...
