'use client'

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { summarizeScoreSeverity } from '../lib/risk'
import { navItemsForRole, type Role, type Tab } from '../lib/navigation'

const API_GATEWAY = 'http://localhost:8080'
const API_QUERY = 'http://localhost:8085'
const API_ALERTS = 'http://localhost:8083'
const API_GOV = 'http://localhost:8084'

type Alert = { id: number; symbol: string; status: string; ts: string; created_at?: string }
type Score = { id: number; symbol: string; score: number; severity: string; ts: string }
type Model = { id: number; model_name: string; version: string; state: string; created_at: string }
type User = { user: string; role: Role } | null

export default function Page() {
  const [user, setUser] = useState<User>(null)
  const [tab, setTab] = useState<Tab>('overview')
  const [loading, setLoading] = useState(true)
  const [loggingIn, setLoggingIn] = useState(false)
  const [error, setError] = useState('')
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [scores, setScores] = useState<Score[]>([])
  const [models, setModels] = useState<Model[]>([])
  const [liveAlerts, setLiveAlerts] = useState<string[]>([])
  const [sseState, setSseState] = useState<'connected' | 'reconnecting' | 'failed'>('reconnecting')
  const [retries, setRetries] = useState(0)
  const retryTimer = useRef<number | null>(null)

  const fetchUser = useCallback(async () => {
    try {
      const res = await fetch(`${API_GATEWAY}/me`, { credentials: 'include' })
      setUser(res.ok ? await res.json() : null)
    } catch {
      setUser(null)
    }
  }, [])

  const fetchData = useCallback(async () => {
    const [a, s] = await Promise.all([
      fetch(`${API_QUERY}/alerts`, { credentials: 'include' }),
      fetch(`${API_QUERY}/scores`, { credentials: 'include' }),
    ])
    if (a.ok) setAlerts(await a.json())
    if (s.ok) setScores(await s.json())
  }, [])

  const fetchModels = useCallback(async () => {
    try {
      const r = await fetch(`${API_GOV}/models`, { credentials: 'include' })
      if (r.ok) setModels(await r.json())
    } catch {
      // ignore quietly
    }
  }, [])

  useEffect(() => {
    const init = async () => {
      setLoading(true)
      await fetchUser()
      await fetchData()
      await fetchModels()
      setLoading(false)
    }
    init()
    const id = window.setInterval(() => {
      fetchData()
      if (user) fetchModels()
    }, 6000)
    return () => window.clearInterval(id)
  }, [fetchUser, fetchData, fetchModels, user])

  useEffect(() => {
    if (!user) return
    let es: EventSource | null = null
    let cancelled = false

    const connect = () => {
      if (cancelled) return
      setSseState(retries > 0 ? 'reconnecting' : 'reconnecting')
      es = new EventSource(`${API_ALERTS}/sse/alerts`)
      es.onopen = () => {
        setSseState('connected')
        setRetries(0)
      }
      es.addEventListener('alert', (evt) => {
        setLiveAlerts(prev => [evt.data, ...prev].slice(0, 30))
        fetchData()
      })
      es.onerror = () => {
        es?.close()
        const next = Math.min(30000, 1000 * Math.pow(2, Math.min(retries, 5)))
        setRetries(r => r + 1)
        setSseState(next >= 30000 ? 'failed' : 'reconnecting')
        retryTimer.current = window.setTimeout(connect, next)
      }
    }

    connect()
    return () => {
      cancelled = true
      if (retryTimer.current) window.clearTimeout(retryTimer.current)
      es?.close()
    }
  }, [user, fetchData, retries])

  const navItems = useMemo(() => (user ? navItemsForRole(user.role) : []), [user])

  const doLogin = async (email: string, password: string) => {
    setLoggingIn(true)
    setError('')
    try {
      const res = await fetch(`${API_GATEWAY}/auth/login`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ Email: email, Password: password }),
      })
      if (!res.ok) throw new Error('Invalid credentials')
      await fetchUser()
      await fetchData()
      await fetchModels()
    } catch (e: any) {
      setError(e?.message || 'Login failed')
    }
    setLoggingIn(false)
  }

  const logout = async () => {
    await fetch(`${API_GATEWAY}/auth/logout`, { method: 'POST', credentials: 'include' }).catch(() => null)
    setUser(null)
    setTab('overview')
  }

  const acknowledge = async (id: number) => {
    setError('')
    const res = await fetch(`${API_ALERTS}/alerts/${id}/ack`, { method: 'POST', credentials: 'include' })
    if (!res.ok) {
      setError(`Acknowledge failed (${res.status})`) 
      return
    }
    await fetchData()
  }

  if (loading) return <div className="center-screen">Loading Sentinel...</div>

  if (!user) {
    return <LandingPage onLogin={doLogin} loggingIn={loggingIn} error={error} />
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand"><span className="brand-dot" />Sentinel</div>
        <nav className="nav">
          {navItems.map(n => (
            <button key={n} className={tab === n ? 'nav-btn active' : 'nav-btn'} onClick={() => setTab(n)}>{n}</button>
          ))}
        </nav>
        <div className="user-meta">
          <span className="pill">{user.user}</span>
          <span className="pill role">{user.role}</span>
          <button className="nav-btn" onClick={logout}>Logout</button>
        </div>
      </header>

      {error && <div className="banner banner-error">{error}</div>}
      {(alerts.length > 0 && scores.length === 0) && <div className="banner banner-warn">Pipeline degraded: alerts are present while score persistence is empty.</div>}
      {sseState !== 'connected' && <div className="banner banner-warn">Live stream {sseState}. automatic reconnect is active.</div>}

      {tab === 'overview' && <Overview alerts={alerts} scores={scores} liveAlerts={liveAlerts} />}
      {tab === 'alerts' && <AlertsView alerts={alerts} role={user.role} onAck={acknowledge} />}
      {tab === 'scores' && <ScoresView scores={scores} />}
      {tab === 'governance' && user.role === 'admin' && <GovernanceView models={models} />}

      <footer className="footer-note">Sentinel · Streaming anomaly detection with institutional-grade governance</footer>
    </div>
  )
}

function LandingPage({ onLogin, loggingIn, error }: { onLogin: (e: string, p: string) => Promise<void>; loggingIn: boolean; error: string }) {
  const [email, setEmail] = useState('admin@sentinel.local')
  const [password, setPassword] = useState('Sentinel#123')
  return (
    <div className="landing-wrap">
      <section className="landing-card glass-panel">
        <h1>Sentinel Risk Intelligence</h1>
        <p>Real-time anomaly detection, governed model scoring, and incident investigation for institutional operations.</p>
        <div className="cta-row">
          <a className="nav-btn active" href="#signin">Sign In</a>
          <button className="nav-btn">Request Demo</button>
          <button className="nav-btn">Learn More</button>
        </div>
      </section>
      <section className="landing-card glass-panel" id="signin">
        <h2>Sign In</h2>
        <form onSubmit={async (e) => { e.preventDefault(); await onLogin(email, password) }} className="form-grid">
          <input value={email} onChange={e => setEmail(e.target.value)} type="email" required />
          <input value={password} onChange={e => setPassword(e.target.value)} type="password" required />
          {error && <div className="error-text">{error}</div>}
          <button className="nav-btn active" type="submit" disabled={loggingIn}>{loggingIn ? 'Signing in...' : 'Enter Platform'}</button>
        </form>
      </section>
    </div>
  )
}

function Overview({ alerts, scores, liveAlerts }: { alerts: Alert[]; scores: Score[]; liveAlerts: string[] }) {
  const open = alerts.filter(a => a.status === 'open').length
  const { critical, high } = summarizeScoreSeverity(scores)
  const symbols = new Set(scores.map(s => s.symbol)).size
  return (
    <main className="dashboard">
      {[['Open Alerts', open], ['Critical Scores', critical], ['High Scores', high], ['Symbols Tracked', symbols]].map(([k, v]) => (
        <section key={k as string} className="glass-panel stat-card"><h3>{k}</h3><div className="stat-value">{v as number}</div></section>
      ))}
      <section className="glass-panel span-8">
        <h3>Risk Watchlist</h3>
        <div className="table">
          <div className="row head"><div>Symbol</div><div>Score</div><div>Severity</div><div>Time</div></div>
          {scores.slice(0, 12).map(s => <div className="row" key={s.id}><div className="ticker">{s.symbol}</div><div>{s.score.toFixed(3)}</div><div>{s.severity}</div><div>{new Date(s.ts).toLocaleTimeString()}</div></div>)}
        </div>
      </section>
      <section className="glass-panel span-4">
        <h3>Live Alert Feed</h3>
        <div className="feed-list">
          {liveAlerts.slice(0, 8).map((r, i) => {
            try {
              const p = JSON.parse(r)
              return <div className="feed-item" key={i}><strong>{p.symbol}</strong><small>{p.explanation}</small></div>
            } catch {
              return <div className="feed-item" key={i}><strong>Alert</strong><small>{r}</small></div>
            }
          })}
          {liveAlerts.length === 0 && <div className="muted">Waiting for live alerts...</div>}
        </div>
      </section>
    </main>
  )
}

function AlertsView({ alerts, role, onAck }: { alerts: Alert[]; role: Role; onAck: (id: number) => Promise<void> }) {
  const canAck = role === 'admin' || role === 'analyst'
  return (
    <section className="glass-panel">
      <h3>Alerts</h3>
      <div className="table">
        <div className="row head"><div>Symbol</div><div>Status</div><div>Created</div><div>Action</div></div>
        {alerts.slice(0, 50).map(a => (
          <div key={a.id} className="row">
            <div className="ticker">{a.symbol}</div>
            <div>{a.status}</div>
            <div>{new Date(a.created_at || a.ts).toLocaleString()}</div>
            <div>{a.status === 'open' && canAck ? <button className="nav-btn" onClick={() => onAck(a.id)}>Acknowledge</button> : <span className="muted">—</span>}</div>
          </div>
        ))}
      </div>
    </section>
  )
}

function ScoresView({ scores }: { scores: Score[] }) {
  return (
    <section className="glass-panel">
      <h3>Anomaly Scores</h3>
      <div className="table">
        <div className="row head"><div>Symbol</div><div>Score</div><div>Severity</div><div>Time</div></div>
        {scores.slice(0, 100).map(s => <div key={s.id} className="row"><div className="ticker">{s.symbol}</div><div>{s.score.toFixed(4)}</div><div>{s.severity}</div><div>{new Date(s.ts).toLocaleString()}</div></div>)}
      </div>
    </section>
  )
}

function GovernanceView({ models }: { models: Model[] }) {
  return (
    <section className="glass-panel">
      <h3>Governance</h3>
      <div className="table">
        <div className="row head"><div>Model</div><div>Version</div><div>State</div><div>Created</div></div>
        {models.map(m => <div key={m.id} className="row"><div className="ticker">{m.model_name}</div><div>{m.version}</div><div>{m.state}</div><div>{new Date(m.created_at).toLocaleString()}</div></div>)}
      </div>
    </section>
  )
}
