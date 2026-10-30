# Expansion plan

The first version stays database-free: one private config token, one API route,
a documented envelope, and an explicit list of compiled functions.

1. **Add useful functions.** Decode and validate each action's typed arguments,
   keep implementations separate, and test their results and failures. Add
   application-owned fixed upstreams only when there is a concrete need.
2. **Improve operations.** Add structured metadata-only logs, deployment service
   configuration, HTTPS reverse proxy guidance, and bounded request rate limits
   for the actual host. Never log Authorization, config, or payload secrets.
3. **Grow authentication when needed.** A shared token identifies possession,
   not a person. Configured token IDs/scopes can come next. Rotate the current
   token by stopping the server, changing the private token, and restarting;
   callers must update too. Interactive login and expiring tokens need a separate
   design, not a new unprotected endpoint that issues tokens.
4. **Add a browser demo only if useful.** Prefer serving it from the same origin.
   Any cross-origin access needs an explicit allow-list; never embed a server
   token in built frontend assets or store it in browser localStorage.
5. **Consider SQLite for a specific persistence need.** Request history or job
   state may justify it later. Authentication and dummy actions do not need it.
   Define retention, concurrency, backup, and redaction before storing data.

Large jobs, retries, and queues are later decisions. Do not silently turn a
synchronous action into a retryable side effect. Version the contract before
changing params from an array to named arguments or introducing new semantics.

The existing legacy gateway is preserved for reference, not automatically
promoted into the generic server. Review and migrate only the feature required
by the next concrete use case.
