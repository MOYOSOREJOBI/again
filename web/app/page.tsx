'use client'

import { memo, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { summarizeScoreSeverity } from '../lib/risk'
import { filterAlerts, groupAlerts, type AlertLike } from '../lib/alerts'
import { hasCapability, navItemsForRole, type Role, type Tab } from '../lib/roles'

const API_GATEWAY = 'http://localhost:8080'
const API_QUERY = 'http://localhost:8085'
const API_ALERTS = 'http://localhost:8083'
const API_GOV = 'http://localhost:8084'

const POLL_INTERVAL = 15_000
const LIVE_FLUSH_INTERVAL = 3_000

type Alert = { id: number; symbol: string; status: string; ts: string; created_at?: string; severity?: string; explanation?: string; score?: number }
type Score = { id: number; symbol: string; score: number; severity: string; ts: string; explanation?: string }
type Model = { id: number; model_name: string; version: string; state: string; created_at: string }
type User = { user: string; role: Role } | null
type AuditState = { ok: boolean; message: string } | null

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
  const [liveChanges, setLiveChanges] = useState(0)
  const [lastUpdated, setLastUpdated] = useState<string>('—')
  const [sseState, setSseState] = useState<'connected' | 'reconnecting' | 'failed'>('reconnecting')
  const [auditState, setAuditState] = useState<AuditState>(null)
  const [governanceNotice, setGovernanceNotice] = useState('')

  const userRef = useRef<User>(null)
  const retriesRef = useRef(0)
  const retryTimer = useRef<number | null>(null)
  const liveBufferRef = useRef<string[]>([])

  const fetchUser = useCallback(async () => {
    try {
      const res = await fetch(`${API_GATEWAY}/me`, { credentials: 'include' })
      const u = res.ok ? await res.json() : null
      userRef.current = u
      setUser(u)
    } catch {
      userRef.current = null
      setUser(null)
    }
  }, [])

  const fetchData = useCallback(async () => {
    try {
      const [a, s] = await Promise.all([
        fetch(`${API_QUERY}/alerts`, { credentials: 'include' }),
        fetch(`${API_QUERY}/scores`, { credentials: 'include' }),
      ])
      if (a.ok) setAlerts(await a.json())
      if (s.ok) setScores(await s.json())
      setLastUpdated(new Date().toLocaleTimeString())
    } catch {
      // network error — silently skip this poll cycle
    }
  }, [])

  const fetchModels = useCallback(async () => {
    try {
      const r = await fetch(`${API_GOV}/models`, { credentials: 'include' })
      if (r.ok) {
        setModels(await r.json())
      } else {
        setGovernanceNotice('Governance data unavailable for this role or environment.')
      }
    } catch {
      setGovernanceNotice('Governance service currently unavailable. Showing cached data if present.')
    }
  }, [])

  const verifyAudit = useCallback(async () => {
    setGovernanceNotice('')
    try {
      const res = await fetch(`${API_GATEWAY}/audit/verify`, { credentials: 'include' })
      if (!res.ok) {
        setGovernanceNotice(`Audit verification unavailable (${res.status}).`)
        return
      }
      const body = await res.json()
      setAuditState(body)
    } catch {
      setGovernanceNotice('Audit verification endpoint unreachable.')
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    const init = async () => {
      setLoading(true)
      await fetchUser()
      await fetchData()
      await fetchModels()
      if (!cancelled) setLoading(false)
    }
    init()
    return () => {
      cancelled = true
    }
  }, [fetchUser, fetchData, fetchModels])

  useEffect(() => {
    if (loading) return
    const id = window.setInterval(() => {
      fetchData()
      if (userRef.current) fetchModels()
    }, POLL_INTERVAL)
    return () => window.clearInterval(id)
  }, [loading, fetchData, fetchModels])

  useEffect(() => {
    const flush = window.setInterval(() => {
      const pending = liveBufferRef.current
      if (!pending.length) return
      setLiveAlerts((prev) => [...pending.reverse(), ...prev].slice(0, 40))
      setLiveChanges((v) => v + pending.length)
      liveBufferRef.current = []
    }, LIVE_FLUSH_INTERVAL)
    return () => window.clearInterval(flush)
  }, [])

  useEffect(() => {
    if (!user) return
    let es: EventSource | null = null
    let cancelled = false

    const connect = () => {
      if (cancelled) return
      setSseState('reconnecting')
      es = new EventSource(`${API_ALERTS}/sse/alerts`)
      es.onopen = () => {
        setSseState('connected')
        retriesRef.current = 0
      }
      es.addEventListener('alert', (evt) => {
        liveBufferRef.current.push(evt.data)
      })
      es.onerror = () => {
        es?.close()
        if (cancelled) return
        const r = retriesRef.current
        const next = Math.min(30000, 1000 * Math.pow(2, Math.min(r, 5)))
        retriesRef.current = r + 1
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
  }, [user])

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
    userRef.current = null
    setUser(null)
    setTab('overview')
  }

  const updateAlertStatus = async (id: number, action: 'ack' | 'resolve' | 'suppress') => {
    setError('')
    const current = alerts.find((a) => a.id === id)
    if (!current) return
    if (action !== 'ack') {
      setError(`${action} is not exposed by backend yet. Control remains disabled for safety.`)
      return
    }
    const res = await fetch(`${API_ALERTS}/alerts/${id}/ack`, { method: 'POST', credentials: 'include' })
    if (!res.ok) {
      setError(`Action failed (${res.status})`)
      return
    }
    await fetchData()
  }

  const deployModel = async (modelName: string, version: string) => {
    setGovernanceNotice('')
    const res = await fetch(`${API_GOV}/models/deploy`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ model_name: modelName, version }),
    })
    if (!res.ok) {
      setGovernanceNotice(`Model deploy failed (${res.status})`)
      return
    }
    setGovernanceNotice(`Model ${modelName}:${version} deployed.`)
    await fetchModels()
  }

  const triggerReplay = async () => {
    setGovernanceNotice('')
    const res = await fetch(`${API_GOV}/replay/start`, { method: 'POST', credentials: 'include' })
    if (!res.ok) {
      setGovernanceNotice(`Replay start failed (${res.status})`)
      return
    }
    const body = await res.json()
    setGovernanceNotice(`Replay run started: ${body.id}`)
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
          {navItems.map((n) => (
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

      {tab === 'overview' && <Overview alerts={alerts} scores={scores} role={user.role} liveAlerts={liveAlerts} liveChanges={liveChanges} lastUpdated={lastUpdated} />}
      {tab === 'alerts' && <AlertsWorkbench alerts={alerts} role={user.role} onAction={updateAlertStatus} />}
      {tab === 'scores' && <ScoresView scores={scores} role={user.role} />}
      {tab === 'governance' && user.role === 'admin' && (
        <GovernanceView
          models={models}
          auditState={auditState}
          notice={governanceNotice}
          onVerifyAudit={verifyAudit}
          onDeploy={deployModel}
          onReplay={triggerReplay}
        />
      )}

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
        <h1>Sentinel Risk Operations</h1>
        <p>Streaming anomaly intake, governed alerts, replay support, and audit verification for risk teams.</p>
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

const Overview = memo(function Overview({ alerts, scores, role, liveAlerts, liveChanges, lastUpdated }: { alerts: Alert[]; scores: Score[]; role: Role; liveAlerts: string[]; liveChanges: number; lastUpdated: string }) {
  const open = alerts.filter((a) => a.status === 'open').length
  const acked = alerts.filter((a) => a.status === 'ack').length
  const { critical, high } = summarizeScoreSeverity(scores)
  const symbols = new Set(scores.map((s) => s.symbol)).size
  const topSymbols = Object.entries(scores.reduce<Record<string, number>>((acc, s) => {
    acc[s.symbol] = (acc[s.symbol] || 0) + 1
    return acc
  }, {})).sort((a, b) => b[1] - a[1]).slice(0, 5)

  return (
    <main className="dashboard">
      {[['Open Alerts', open], ['Acknowledged', acked], ['Critical Scores', critical], ['Symbols Tracked', symbols]].map(([k, v]) => (
        <section key={k as string} className="glass-panel stat-card"><h3>{k}</h3><div className="stat-value">{v as number}</div></section>
      ))}
      <section className="glass-panel span-8">
        <h3>{role === 'viewer' ? 'Risk Watchlist' : 'Priority Queue'}</h3>
        <p className="muted panel-copy">{role === 'viewer' ? 'Read-only monitoring of latest anomalies.' : 'Actionable queue for triage and escalation.'}</p>
        <div className="table">
          <div className="row head"><div>Symbol</div><div>Score</div><div>Severity</div><div>Why this matters</div></div>
          {scores.slice(0, 10).map((s) => (
            <div className="row" key={s.id}><div className="ticker">{s.symbol}</div><div>{s.score.toFixed(3)}</div><div>{s.severity}</div><div>{s.explanation || 'Unusual movement detected against rolling baseline.'}</div></div>
          ))}
          {scores.length === 0 && <div className="empty-state">No scores yet. Streaming pipeline may still be warming up.</div>}
        </div>
      </section>
      <section className="glass-panel span-4">
        <h3>Live Incident Feed</h3>
        <div className="meta-row"><span className="pill">last update {lastUpdated}</span><span className="pill">{liveChanges} changes</span></div>
        <div className="feed-list">
          {liveAlerts.slice(0, 6).map((r, i) => {
            try {
              const p = JSON.parse(r)
              return <div className="feed-item" key={i}><strong>{p.symbol}</strong><small>{p.explanation || 'Alert emitted by anomaly rules.'}</small></div>
            } catch {
              return <div className="feed-item" key={i}><strong>Alert</strong><small>{r}</small></div>
            }
          })}
          {liveAlerts.length === 0 && <div className="muted">Waiting for live alerts...</div>}
        </div>
      </section>
      <section className="glass-panel span-6">
        <h3>Top Affected Symbols</h3>
        <div className="simple-list">{topSymbols.map(([sym, count]) => <div key={sym}><strong>{sym}</strong><span>{count} high-risk records</span></div>)}</div>
      </section>
      <section className="glass-panel span-6">
        <h3>Recommended Actions</h3>
        <ul className="action-list">
          <li>{open > 0 ? `${open} open alerts require triage.` : 'No open alerts currently.'}</li>
          <li>{critical > 0 ? `${critical} critical scores merit immediate review.` : 'No critical scores in current window.'}</li>
          <li>{role === 'viewer' ? 'Viewer access is read-only by policy.' : 'Use alerts workbench to acknowledge high-confidence incidents.'}</li>
          <li>{role === 'admin' ? 'Verify audit chain and run replay after model changes.' : 'Escalate governance tasks to admin role.'}</li>
        </ul>
      </section>
    </main>
  )
})

const AlertsWorkbench = memo(function AlertsWorkbench({ alerts, role, onAction }: { alerts: Alert[]; role: Role; onAction: (id: number, action: 'ack' | 'resolve' | 'suppress') => Promise<void> }) {
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [severity, setSeverity] = useState('all')
  const [status, setStatus] = useState('all')
  const [symbol, setSymbol] = useState('')
  const [groupBy, setGroupBy] = useState<'status' | 'symbol'>('status')

  const enhancedAlerts = useMemo<AlertLike[]>(() => alerts.map((a) => ({ ...a, severity: a.severity || 'unknown', explanation: a.explanation || 'Awaiting expanded context from stream.', score: a.score ?? 0 })), [alerts])
  const filtered = useMemo(() => filterAlerts(enhancedAlerts, { severity, status, symbol }), [enhancedAlerts, severity, status, symbol])
  const grouped = useMemo(() => groupAlerts(filtered, groupBy), [filtered, groupBy])
  const selected = filtered.find((a) => a.id === selectedId) || filtered[0]

  return (
    <section className="glass-panel">
      <h3>Alerts Workbench</h3>
      <div className="toolbar">
        <select value={severity} onChange={(e) => setSeverity(e.target.value)}><option value="all">All severities</option><option value="critical">Critical</option><option value="high">High</option><option value="medium">Medium</option><option value="low">Low</option></select>
        <select value={status} onChange={(e) => setStatus(e.target.value)}><option value="all">All statuses</option><option value="open">Open</option><option value="ack">Ack</option><option value="resolved">Resolved</option><option value="suppressed">Suppressed</option></select>
        <input placeholder="Filter by symbol" value={symbol} onChange={(e) => setSymbol(e.target.value)} />
        <select value={groupBy} onChange={(e) => setGroupBy(e.target.value as 'status' | 'symbol')}><option value="status">Group by status</option><option value="symbol">Group by symbol</option></select>
      </div>
      {Object.entries(grouped).map(([group, items]) => (
        <div key={group} className="group-block">
          <h4>{group} <span className="muted">({items.length})</span></h4>
          <div className="table">
            <div className="row head"><div>Symbol</div><div>Status</div><div>Created</div><div>Action</div></div>
            {items.slice(0, 20).map((a) => (
              <div key={a.id} className={selected?.id === a.id ? 'row selected-row' : 'row'} onClick={() => setSelectedId(a.id)}>
                <div className="ticker">{a.symbol}</div>
                <div>{a.status}</div>
                <div>{new Date((a as Alert).created_at || a.ts).toLocaleString()}</div>
                <div>
                  {a.status === 'open' && hasCapability(role, 'alerts:ack') ? <button className="nav-btn" onClick={(e) => { e.stopPropagation(); onAction(a.id, 'ack') }}>Acknowledge</button> : <span className="muted">read-only</span>}
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}
      {filtered.length === 0 && <div className="empty-state">No alerts match current filters.</div>}
      {selected && <AlertDetailDrawer alert={selected} role={role} onAction={onAction} />}
      {role === 'viewer' && <div className="muted">Viewer role is monitoring-only; triage actions are intentionally hidden.</div>}
    </section>
  )
})

function AlertDetailDrawer({ alert, role, onAction }: { alert: AlertLike; role: Role; onAction: (id: number, action: 'ack' | 'resolve' | 'suppress') => Promise<void> }) {
  return (
    <aside className="detail-drawer">
      <h3>Alert Detail</h3>
      <p><strong>{alert.symbol}</strong> · {alert.status} · {alert.severity}</p>
      <p className="muted">{alert.explanation}</p>
      <div className="action-row">
        {hasCapability(role, 'alerts:ack') && alert.status === 'open' && <button className="nav-btn" onClick={() => onAction(alert.id, 'ack')}>Acknowledge</button>}
        <button className="nav-btn" disabled={!hasCapability(role, 'alerts:resolve') || alert.status !== 'ack'} onClick={() => onAction(alert.id, 'resolve')}>Resolve</button>
        <button className="nav-btn" disabled={!hasCapability(role, 'alerts:suppress') || alert.status === 'suppressed'} onClick={() => onAction(alert.id, 'suppress')}>Suppress</button>
      </div>
      <p className="muted">Resolve/Suppress remain disabled unless backend transition endpoints are introduced.</p>
    </aside>
  )
}

const ScoresView = memo(function ScoresView({ scores, role }: { scores: Score[]; role: Role }) {
  const distribution = scores.reduce<Record<string, number>>((acc, s) => {
    acc[s.severity] = (acc[s.severity] || 0) + 1
    return acc
  }, {})
  const movers = [...scores].sort((a, b) => b.score - a.score).slice(0, 10)

  return (
    <section className="glass-panel">
      <h3>Anomaly Scores</h3>
      <div className="simple-list">
        {Object.entries(distribution).map(([sev, count]) => <div key={sev}><strong>{sev}</strong><span>{count} signals</span></div>)}
      </div>
      <div className="table">
        <div className="row head"><div>Symbol</div><div>Score</div><div>Severity</div><div>Context</div></div>
        {movers.map((s) => <div key={s.id} className="row"><div className="ticker">{s.symbol}</div><div>{s.score.toFixed(4)}</div><div>{s.severity}</div><div>{s.explanation || `${role === 'viewer' ? 'Monitoring' : 'Review'} deviation from baseline behavior.`}</div></div>)}
      </div>
      {scores.length === 0 && <div className="empty-state">No scores available yet.</div>}
    </section>
  )
})

const GovernanceView = memo(function GovernanceView({ models, auditState, notice, onVerifyAudit, onDeploy, onReplay }: { models: Model[]; auditState: AuditState; notice: string; onVerifyAudit: () => Promise<void>; onDeploy: (name: string, version: string) => Promise<void>; onReplay: () => Promise<void> }) {
  const [modelName, setModelName] = useState('risk-anomaly-v2')
  const [version, setVersion] = useState('2.0.0')
  return (
    <section className="glass-panel">
      <h3>Governance Control Plane</h3>
      {notice && <div className="banner banner-warn">{notice}</div>}
      <div className="action-row">
        <button className="nav-btn active" onClick={() => onVerifyAudit()}>Verify Audit Chain</button>
        <button className="nav-btn" onClick={() => onReplay()}>Trigger Replay</button>
      </div>
      {auditState && <div className={auditState.ok ? 'banner banner-ok' : 'banner banner-error'}>Audit status: {auditState.ok ? 'verified' : 'failed'} · {auditState.message}</div>}
      <div className="toolbar">
        <input value={modelName} onChange={(e) => setModelName(e.target.value)} />
        <input value={version} onChange={(e) => setVersion(e.target.value)} />
        <button className="nav-btn" onClick={() => onDeploy(modelName, version)}>Deploy Model</button>
      </div>
      <div className="table">
        <div className="row head"><div>Model</div><div>Version</div><div>State</div><div>Created</div></div>
        {models.map((m) => <div key={m.id} className="row"><div className="ticker">{m.model_name}</div><div>{m.version}</div><div>{m.state}</div><div>{new Date(m.created_at).toLocaleString()}</div></div>)}
      </div>
      {models.length === 0 && <div className="empty-state">No registry records available yet.</div>}
    </section>
  )
})
