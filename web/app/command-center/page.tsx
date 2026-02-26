'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import TrustStrip from '../../components/TrustStrip'
import WorldRiskMap from '../../components/WorldRiskMap'
import AppShell, { useGlobalFilters } from '../../components/AppShell'
import { api } from '../../lib/api'

export default function CommandCenterPage() {
  const [summary, setSummary] = useState<any>({ top_incidents: [], by_severity: {} })
  const [trust, setTrust] = useState<any>({ degraded: [] })
  const [world, setWorld] = useState<any[]>([])
  const { filters, apply } = useGlobalFilters()

  useEffect(() => {
    api.commandCenter(filters).then(setSummary)
    api.trust(filters).then(setTrust)
    api.worldMap(filters).then(setWorld)
  }, [filters])

  return (
    <AppShell title="Command Center" subtitle="Operational overview across incidents, trust, and region pressure." filters={filters} setFilters={apply}>
      <TrustStrip trustState={trust.degraded?.length ? 'degraded' : 'stable'} modelUnavailable={Boolean(trust.model_unavailable_count)} dqWarning={`missing=${trust.missingness_rate} dup=${trust.duplicate_rate}`} />
      <div className="grid-3">
        <div className="card"><h3>Open incidents</h3><p>{summary.open_incidents}</p></div>
        <div className="card"><h3>High risk</h3><p>{summary.high_risk_incidents}</p></div>
        <div className="card"><h3>What changed</h3><p>{summary.changed_last_24h || 0} status transitions in 24h.</p></div>
      </div>
      <div className="grid-2">
        <section className="card">
          <h3>Top incidents now</h3>
          {summary.top_incidents?.length ? <table className="table"><thead><tr><th>ID</th><th>Symbol</th><th>Severity</th><th>Composite</th></tr></thead><tbody>{summary.top_incidents.slice(0, 8).map((r: any) => <tr key={r.id}><td><Link href={`/incident/${r.id}`}>{r.id}</Link></td><td>{r.symbol}</td><td>{r.severity_band}</td><td>{r.composite_risk?.toFixed?.(2)}</td></tr>)}</tbody></table> : <div className="empty-state">No open critical incidents right now — here’s what Sentinel is monitoring and how severity is assigned.</div>}
          <p className="muted" style={{ marginTop: 10 }}><Link href="/queue">Open Queue</Link> · <Link href="/trust">Trust</Link> · <Link href="/executive">Executive</Link></p>
        </section>
        <WorldRiskMap rows={world} filters={filters} onFilter={apply} />
      </div>
    </AppShell>
  )
}
