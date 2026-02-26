'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import TrustStrip from '../../components/TrustStrip'
import { api } from '../../lib/api'
import AppShell, { useGlobalFilters } from '../../components/AppShell'

export default function TrustPage() {
  const [trust, setTrust] = useState<any>(null)
  const { filters, apply } = useGlobalFilters()
  useEffect(() => { api.trust(filters).then(setTrust) }, [filters])

  return (
    <AppShell title="Trust" subtitle="Data quality, model health, and circuit-breaker posture." filters={filters} setFilters={apply}>
      {!trust ? <div className="empty-state">Trust data unavailable.</div> : <>
        <TrustStrip trustState={trust.degraded?.length ? 'degraded' : 'stable'} modelUnavailable={Boolean(trust.model_unavailable_count)} dqWarning={`missing=${trust.missingness_rate} dup=${trust.duplicate_rate}`} drift={trust.drift_indicator} role="analyst" />
        <div className="grid-3">
          <div className="card"><h3>Feed freshness</h3><p>{trust.freshness_seconds}s</p></div>
          <div className="card"><h3>Missing / dupes</h3><p>{trust.missingness_rate} / {trust.duplicate_rate}</p></div>
          <div className="card"><h3>Circuit breaker</h3><p>{trust.circuit_breaker_state || 'closed'}</p></div>
        </div>
        <section className="card">
          <h3>Top degraded symbols/venues</h3>
          {(trust.degraded || []).length ? <table className="table"><thead><tr><th>Symbol</th><th>Alerts</th></tr></thead><tbody>{trust.degraded.map((d: any, i: number) => <tr key={i}><td>{d.symbol}</td><td>{d.alert_count}</td></tr>)}</tbody></table> : <div className="empty-state">No degraded feeds right now — here’s how trust penalties work: degraded data lowers confidence and can downgrade incident rank.</div>}
          <p className="muted" style={{ marginTop: 10 }}><Link href="/queue">Show DQ-penalized incidents</Link> · <Link href="/queue">Back to Queue</Link></p>
        </section>
      </>}
    </AppShell>
  )
}
