# 1. Request flow

The executable loads JSON, constructs a server.Server, and gives that handler
to net/http.

browser / fetch / curl -> ServeHTTP -> request ID and log -> health or admin ->
fixed method/path match -> bearer authentication -> fixed HTTP target.

The caller supplies a path and values, but never the upstream destination.
Targets come from validated startup configuration.
