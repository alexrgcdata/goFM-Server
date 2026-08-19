# Go guide: reading goFM Server

This is a practical companion to the server code, written for someone brushing
up on Go while making changes here.

## Where the code starts

`cmd/gofm-server/main.go` is the executable program. Every executable Go
program begins in a package named `main`, with a `main()` function as its entry
point. This program reads a JSON configuration file, constructs the server,
and passes it to Go's built-in HTTP listener.

```go
if err := http.ListenAndServe(config.Address, server.New(config)); err != nil {
    log.Printf("server stopped: %v", err)
    os.Exit(1)
}
```

The `if err := ...; err != nil` form creates `err` only for that `if` block.
It is idiomatic Go: operations return an error explicitly rather than throwing
an exception. `%v` formats the error value for the log.

## Packages and visibility

Go code is grouped into packages, usually one package per folder. The package
under `internal/server` is intentionally private to this repository: Go
prevents code outside the project tree from importing an `internal` package.

An identifier beginning with a capital letter is exported for other packages:

```go
type Config struct { /* ... */ } // exported
func LoadConfig(path string) {}  // exported
func writeError() {}             // private to package server
```

## Structs and JSON tags

`Config`, `Route`, and `Target` are structs: named groups of typed fields.
The quoted tag after a field tells `encoding/json` how to map a JSON key.

```go
type Route struct {
    Path    string   `json:"path"`
    Methods []string `json:"methods"`
    Target  Target   `json:"target"`
}
```

`[]string` is a growable list of strings. `Target` is nested directly rather
than being a pointer because every configured route requires a target.

## Methods and HTTP handlers

`Server` has methods. The receiver `(s *Server)` means the method operates on
a pointer to a `Server`, avoiding a copy of its HTTP client and configuration.

```go
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

This exact method makes `*Server` satisfy Go's `http.Handler` interface. Go
uses structural interfaces: there is no separate `implements` declaration.
`r` is the incoming request, and `w` is used to set headers, status, and body.

## Request flow

1. `ServeHTTP` allows `GET /health` without a token.
2. It matches an exact configured method and path.
3. It checks the `Authorization: Bearer ...` header when tokens are configured.
4. It creates a new request for the fixed configured HTTP target.
5. It copies safe headers and streams the upstream response back.

The route list, not client input, determines the upstream URL. That distinction
prevents the server becoming an open proxy.

## Concurrency and safety

`net/http` serves requests concurrently. This scaffold keeps its configuration
read-only after startup, so it is safe to share. The token comparison uses
`subtle.ConstantTimeCompare`, reducing timing clues during comparison.

Always close an HTTP response body:

```go
response, err := s.client.Do(request)
if err != nil { /* handle it */ }
defer response.Body.Close()
```

`defer` schedules the close when the current function returns, including early
returns later in the function.

## Tests

Tests live next to the package they verify and use the `_test.go` suffix. A
test function begins with `Test` and accepts `*testing.T`:

```go
func TestHealthIsPublic(t *testing.T) {
    // Arrange, act, assert.
}
```

Run the project checks with:

```sh
go test ./...
go vet ./...
```

Use `gofmt` after editing Go. It is the standard formatter and removes style
debates from code review.
