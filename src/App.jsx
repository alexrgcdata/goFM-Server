import { useState } from 'react'

function joinUrl(baseUrl, path) {
  const normalizedBase = baseUrl.trim().replace(/\/+$/, '')
  const normalizedPath = path.trim().replace(/^\/+/, '')
  return `${normalizedBase}/${normalizedPath}`
}

async function readResponse(response) {
  const body = await response.text()
  if (!body) return null

  try {
    return JSON.parse(body)
  } catch {
    return body
  }
}

export default function App() {
  const [baseUrl, setBaseUrl] = useState(import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080')
  const [token, setToken] = useState('')
  const [endpoint, setEndpoint] = useState('/api/example')
  const [result, setResult] = useState(null)
  const [loading, setLoading] = useState(false)

  async function request(path, needsAuth) {
    if (!baseUrl.trim()) {
      setResult({ error: 'Enter the API base URL first.' })
      return
    }

    setLoading(true)
    setResult(null)
    const url = joinUrl(baseUrl, path)
    const headers = { Accept: 'application/json' }
    if (needsAuth && token.trim()) headers.Authorization = `Bearer ${token.trim()}`

    try {
      const response = await fetch(url, { headers })
      setResult({
        url,
        status: response.status,
        ok: response.ok,
        body: await readResponse(response),
      })
    } catch (error) {
      setResult({ url, error: error.message })
    } finally {
      setLoading(false)
    }
  }

  return (
    <main>
      <section className="card" aria-labelledby="title">
        <p className="eyebrow">Example React client</p>
        <h1 id="title">goFM Server</h1>
        <p className="intro">Use this small client to check the server health endpoint or make a bearer-token API request.</p>
        <p className="cors-note">
          Browser origin: <code>{window.location.origin}</code>. Add this exact origin to the server's allowed CORS origins before calling an API on another origin.
        </p>

        <label>
          API base URL
          <input value={baseUrl} onChange={(event) => setBaseUrl(event.target.value)} placeholder="http://localhost:8080" inputMode="url" />
        </label>
        <label>
          Bearer token <span>(only sent with the API request)</span>
          <input value={token} onChange={(event) => setToken(event.target.value)} placeholder="development token" type="password" autoComplete="off" />
        </label>
        <label>
          API endpoint
          <input value={endpoint} onChange={(event) => setEndpoint(event.target.value)} placeholder="/api/example" />
        </label>

        <div className="actions">
          <button type="button" onClick={() => request('/health', false)} disabled={loading}>Check health</button>
          <button type="button" className="primary" onClick={() => request(endpoint, true)} disabled={loading}>Call API endpoint</button>
        </div>

        <section className="result" aria-live="polite" aria-label="Response">
          <h2>Response</h2>
          <pre>{result ? JSON.stringify(result, null, 2) : 'No request sent yet.'}</pre>
        </section>
      </section>
    </main>
  )
}
