# Why Go belongs between FileMaker and the web

FileMaker is remarkably good at modeling real-world work. It is fast to build
in, approachable for teams, and deeply effective when a business needs custom
software rather than another generic SaaS product. The web changes the shape of
the problem: browsers, WordPress sites, mobile clients, AI agents, and external
services all expect durable HTTP APIs, precise authentication, and predictable
JSON.

That is where Go makes an unusually strong companion for FileMaker.

## The missing middle layer

It is tempting to let browser code call FileMaker directly. The result is often
an exposed API address, credentials that are hard to rotate, inconsistent
response shapes, and business rules spread across JavaScript, PHP, FileMaker
scripts, and server configuration.

A small Go service creates a clean middle layer. Web clients call one gateway.
The gateway authenticates the caller, applies route policy, records a request
ID, and sends the request to a known FileMaker Data API or OData target. The
browser gets clean JSON back without learning how FileMaker is configured.

## Why Go

Go compiles to a single server binary with a small operational footprint. There
is no Node runtime required on the production API server after compilation. Its
standard HTTP library is mature, concurrency is built in, and deployment is
straightforward on a VPS or container platform.

For a FileMaker team, that matters. The middleware should be boring in the best
way: start one process, load a controlled configuration, handle many requests,
write clear logs, and stay out of the way of the FileMaker solution.

## One secure doorway

OpenBridge uses bearer authentication at the edge. A WordPress page, React
application, mobile client, or future AI web viewer talks to the same doorway.
The gateway verifies the caller token and never forwards that token to
FileMaker. FileMaker credentials stay server-side in an encrypted vault.

That single doorway gives a team one place to answer practical questions:

- Which endpoint was called?
- Which target received it?
- Did FileMaker respond or fail?
- How long did it take?
- Was an after-hook requested?
- What did the normalized response look like?

## Data API and OData without two client architectures

FileMaker Data API and OData have different URLs and response envelopes. A
middleware layer can normalize both into records, found_count, offset, and
limit. The web team works with one predictable JSON shape while the FileMaker
team can choose the appropriate transport per layout or table.

## Ready for AI-assisted workflows

AI agents and web viewers need guardrails. They should not receive a raw
database URL and permission to invent filters or execute scripts. A well-built
gateway exposes named, allow-listed operations: find an active customer, fetch
an order, create a support record, or request a cache rebuild. Each operation
has an allowed input shape, authentication rule, and audit trail.

That turns FileMaker from an isolated application into a secure capability layer
for future agentic interfaces without giving those interfaces unchecked access
to production data.

## Start small, grow deliberately

The first version does not need a complicated platform. It needs fixed routes,
safe authentication, logs, encrypted secrets, and a small number of useful
FileMaker operations. SQLite is a good first store. PostgreSQL and a dedicated
secrets manager become valuable when multiple gateway instances, job queues, or
team-wide audit reporting arrive.

Go does not replace FileMaker. It gives FileMaker a dependable web-facing
boundary: one that is easy to deploy, easy to reason about, and designed to
grow with the business.

By Alex Seidler · agseidler@gmail.com
