import { useEffect, useState, type FormEvent } from 'react'
import { createRoot } from 'react-dom/client'
import './style.css'

type Metrics = {
  uptime_seconds: number
  requests: number
  errors: number
  average_response_ms: number
  recent: TrafficEntry[]
}

type TrafficEntry = {
  time: string
  request_id: string
  action?: string
  status: number
  duration_ms: number
  error_code?: string
}

function isMetrics(value: unknown): value is Metrics {
  if (typeof value !== 'object' || value === null) {
    return false
  }

  const recent: unknown = Reflect.get(value, 'recent')
  if (!Array.isArray(recent) || !recent.every((entry) =>
    entry && typeof entry.time === 'string' && typeof entry.request_id === 'string' &&
    typeof entry.status === 'number' && Number.isFinite(entry.duration_ms) &&
    (entry.action === undefined || typeof entry.action === 'string') &&
    (entry.error_code === undefined || typeof entry.error_code === 'string')
  )) {
    return false
  }

  return ['uptime_seconds', 'requests', 'errors', 'average_response_ms'].every(
    (name) => {
      const field = Reflect.get(value, name)
      return typeof field === 'number' && Number.isFinite(field) && field >= 0
    },
  )
}

function App() {
  const [token, setToken] = useState('')
  const [health, setHealth] = useState('Checking…')
  const [metrics, setMetrics] = useState<Metrics | null>(null)
  const [message, setMessage] = useState('')
  const [loading, setLoading] = useState(false)
  const [sampleResult, setSampleResult] = useState('')

  useEffect(() => {
    const controller = new AbortController()
    fetch('/health', { signal: controller.signal })
      .then((response) => {
        if (!response.ok) {
          throw new Error('Health check failed')
        }

        setHealth('Responding')
      })
      .catch(() => {
        if (!controller.signal.aborted) {
          setHealth('Unavailable')
        }
      })

    return () => controller.abort()
  }, [])

  async function loadMetrics(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    await refresh(false)
  }

  async function refresh(runSample: boolean) {
    setLoading(true)
    setMetrics(null)
    setMessage('')

    try {
      const authorization = 'Bearer ' + token.trim()
      if (runSample) {
        setSampleResult('')
        const sample = await fetch('/api', {
          method: 'POST',
          headers: { Authorization: authorization, 'Content-Type': 'application/json' },
          body: JSON.stringify({ action: 'demo.jobs', params: [] }),
          signal: AbortSignal.timeout(8000),
        })
        if (!sample.ok) {
          throw new Error('Sample request failed (HTTP ' + sample.status + ').')
        }
        const payload = await sample.json()
        setSampleResult(JSON.stringify(payload.result, null, 2))
      }

      // Keep the token in memory and send it only to this Go server.
      const response = await fetch('/admin/metrics', {
        headers: { Authorization: authorization },
        signal: AbortSignal.timeout(8000),
      })

      if (response.status === 401) {
        throw new Error('The server rejected the API token.')
      }
      if (!response.ok) {
        throw new Error('Metrics are unavailable.')
      }

      const data = await response.json()
      if (!isMetrics(data.result)) {
        throw new Error('The server returned an unexpected metrics response.')
      }

      setMetrics(data.result)
      setMessage('Updated at ' + new Date().toLocaleTimeString())
    } catch (error) {
      if (error instanceof Error) {
        setMessage(error.message)
      } else {
        setMessage('Could not load metrics.')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <main>
      <p className="section-name">Workshop / Admin</p>
      <h1>Go API Router</h1>
      <p>Health: <strong>{health}</strong></p>
      <form onSubmit={loadMetrics}>
        <label htmlFor="token">API token</label>
        <input id="token" type="password" autoComplete="off" value={token}
          onChange={(event) => {
            setToken(event.target.value)
            setMetrics(null)
            setMessage('')
            setSampleResult('')
          }} required disabled={loading} />
        <button disabled={loading}>{loading ? 'Loading…' : 'Refresh metrics'}</button>
        <button type="button" disabled={loading || !token.trim()} onClick={() => void refresh(true)}>
          Read sample jobs from PostgreSQL
        </button>
      </form>
      <p role="status">{message}</p>
      {sampleResult && <><h2>PostgreSQL response</h2><pre>{sampleResult}</pre></>}
      {metrics && (
        <dl>
          <div><dt>Uptime</dt><dd>{metrics.uptime_seconds} s</dd></div>
          <div><dt>API requests</dt><dd>{metrics.requests}</dd></div>
          <div><dt>API errors</dt><dd>{metrics.errors}</dd></div>
          <div><dt>Average response</dt><dd>{metrics.average_response_ms.toFixed(2)} ms</dd></div>
        </dl>
      )}
      {metrics && (
        <section>
          <h2>Recent API traffic</h2>
          {metrics.recent.length === 0 ? <p>No API requests yet.</p> : (
            <div className="traffic-table" tabIndex={0} role="region" aria-label="Recent API traffic table">
              <table>
                <thead><tr><th>Time</th><th>Action</th><th>Status</th><th>Duration</th><th>Request ID</th></tr></thead>
                <tbody>
                  {metrics.recent.map((entry) => (
                    <tr key={entry.request_id}>
                      <td>{new Date(entry.time).toLocaleTimeString()}</td>
                      <td>{entry.action || '—'}</td>
                      <td>{entry.status}{entry.error_code && <small>{entry.error_code}</small>}</td>
                      <td>{entry.duration_ms.toFixed(2)} ms</td>
                      <td><code>{entry.request_id}</code></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      )}
      <p className="note">Counts include completed API requests and reset when Go restarts.</p>
    </main>
  )
}

const root = document.getElementById('root')
if (root) {
  createRoot(root).render(<App />)
}
