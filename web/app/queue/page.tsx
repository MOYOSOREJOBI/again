'use client'

import { useEffect, useMemo, useState } from 'react'
import Link from 'next/link'
import AppShell, { useGlobalFilters } from '../../components/AppShell'
import { api, QueueRow, Role } from '../../lib/api'

export default function QueuePage() {
  const [rows, setRows] = useState<QueueRow[]>([])
  const [role, setRole] = useState<Role>('viewer')
  const { filters, apply } = useGlobalFilters()

  useEffect(() => { api.me().then((m) => setRole(m.role || 'viewer')).catch(() => null) }, [])
  useEffect(() => {
    api.queue(filters).then(setRows)
    const id = setInterval(() => api.queue(filters).then(setRows), 7000)
    return () => clearInterval(id)
  }, [filters])

  const sorted = useMemo(() => [...rows].sort((a, b) => (b.priority_score || 0) - (a.priority_score || 0)), [rows])

  return (
    <AppShell title="Incident Queue" subtitle="Primary triage workbench with rank-stable ordering and role-aware actions." filters={filters} setFilters={apply}>
      <div className="card"><strong>Saved views:</strong> <span className="badge">Critical + degraded trust</span> <span className="badge">New in last 24h</span> <span className="badge">My ownership</span></div>
      <section className="card">
        {sorted.length === 0 ? <div className="empty-state">Queue is clear for current filters. Sentinel is still monitoring new anomalies and trust-degraded feeds.</div> : (
          <table className="table"><thead><tr><th>ID</th><th>Symbol</th><th>Venue/Region</th><th>Severity</th><th>Composite</th><th>Esc.</th><th>Trust</th><th>Drivers</th><th>Owner</th><th>Action</th></tr></thead><tbody>
            {sorted.map((r) => <tr key={r.id}><td><Link href={`/incident/${r.id}`}>{r.id}</Link></td><td>{r.symbol}</td><td>{r.venue || '-'} / {r.country || r.region || '-'}</td><td>{r.severity_band}</td><td>{r.composite_risk?.toFixed?.(2) || '-'}</td><td>{r.escalation_probability?.toFixed?.(2) || '-'}</td><td><span className={`badge ${r.trust_state === 'degraded' ? 'degraded' : 'stable'}`}>{r.trust_state || 'stable'}</span></td><td>{r.rank_reason || '-'}</td><td>{r.owner || 'Unassigned'}</td><td>{role === 'viewer' ? 'Open only' : role === 'analyst' ? 'Assign / close / suppress / promote' : 'Assign / close / suppress / promote / audit'}</td></tr>)}
          </tbody></table>
        )}
      </section>
    </AppShell>
  )
}
