# goFM Server React client example

This is a minimal JSX-only Vite client for manually exercising a goFM Server.
It is intentionally separate from the Go server and is not built or served by it.

It is a small admin/testing screen, not a login system or a production UI.

## Run locally

1. Start the Go server (normally on `http://localhost:8080`).
2. From this `web` directory, run `npm install` once and `npm run dev`.
3. Vite starts this client at `http://localhost:5173`. Keep that port free:
   the fixed address makes the CORS setting predictable.
4. Configure the Go server to allow the origin `http://localhost:5173` in its
   CORS allow-list. Do not use a wildcard when bearer tokens are involved.
5. Open the displayed browser address. The page repeats the origin that the
   browser is actually using; if it differs, allow that exact value instead.
6. Use **Check health**, then enter a configured bearer token and API endpoint
   such as `/api/echo` to make an authenticated request.

The browser sends requests directly to the Go server. Without the matching CORS
rule, the browser blocks the response even if a command-line `curl` request
works. The health request does not send your token. The API request sends an
`Authorization: Bearer ...` header only when its token field has a value.

To avoid typing the local API address each time, create a file named `.env.local`
inside `web` containing this one line:

```text
VITE_API_BASE_URL=http://localhost:8080
```

Restart the Vite development server after changing the file. This only chooses
the API address; it does not configure CORS on the Go server.

## JSX and functionality notes

This project uses JavaScript with JSX, not TypeScript. JSX lets a component
return HTML-like markup directly from a JavaScript function. For example,
`<button>Check health</button>` is JSX. `className` is used for CSS classes,
and curly braces insert JavaScript values, such as `{window.location.origin}`.

`App.jsx` is the main screen. `useState` remembers what the user typed and the
most recent response. Each input is a controlled input: its `value` comes from
state and its `onChange` handler updates that state. The two buttons call the
same `request` function with different paths and authentication needs.

The response panel shows the requested URL, HTTP status, and JSON response when
available. Network or CORS failures appear as an error message. Tokens remain
only in browser memory and disappear on refresh; do not use production secrets
in this example client.
