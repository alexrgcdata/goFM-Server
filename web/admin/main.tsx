import { useEffect, useState, type FormEvent } from 'react'
import { createRoot } from 'react-dom/client'
import './style.css'

type Metrics = {
  uptime_seconds: number
  requests: number
  errors: number
  average_response_ms: number
}

function isMetrics(value: unknown): value is Metrics {
  if (typeof value !== 'object' || value === null) {
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
    setLoading(true)
    setMetrics(null)
    setMessage('')

    try {
      // Keep the token in memory and send it only to this Go server.
      const response = await fetch('/admin/metrics', {
        headers: { Authorization: 'Bearer ' + token.trim() },
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
          }} required disabled={loading} />
        <button disabled={loading}>{loading ? 'Loading…' : 'Refresh metrics'}</button>
      </form>
      <p role="status">{message}</p>
      {metrics && (
        <dl>
          <div><dt>Uptime</dt><dd>{metrics.uptime_seconds} s</dd></div>
          <div><dt>API requests</dt><dd>{metrics.requests}</dd></div>
          <div><dt>API errors</dt><dd>{metrics.errors}</dd></div>
          <div><dt>Average response</dt><dd>{metrics.average_response_ms.toFixed(2)} ms</dd></div>
        </dl>
      )}
      <p className="note">Counts include completed API requests and reset when Go restarts.</p>
    </main>
  )
}

const root = document.getElementById('root')
if (root) {
  createRoot(root).render(<App />)
}
