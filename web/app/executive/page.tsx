'use client'

import { useEffect, useState } from 'react'
import AppShell, { useGlobalFilters } from '../../components/AppShell'
import { api } from '../../lib/api'

export default function ExecutivePage() {
  const [summary, setSummary] = useState<any>({ top_risks: [], hot_regions: [], trust_summary: {} })
  const { filters, apply } = useGlobalFilters()
  useEffect(() => { api.executiveSummary(filters).then(setSummary) }, [filters])

  return (
    <AppShell title="Executive Summary" subtitle="Calm, non-technical risk posture for decision-makers." filters={filters} setFilters={apply}>
      <div className="grid-2">
        <section className="card"><h3>Top 5 risks</h3>{summary.top_risks?.length ? <ol>{summary.top_risks.map((r: any) => <li key={r.id}>{r.symbol} — {r.severity_band} ({r.priority_score?.toFixed?.(2)})</li>)}</ol> : <div className="empty-state">No elevated risks in this window.</div>}</section>
        <section className="card"><h3>What changed</h3><p>{summary.what_changed || 'No major ranking shifts.'}</p><p className="muted">Why it matters: concentrated risk can affect liquidity, controls, and regulatory response speed.</p></section>
      </div>
      <section className="card"><h3>Countries / sectors at risk</h3>{summary.hot_regions?.length ? <ul>{summary.hot_regions.map((r: any, i: number) => <li key={i}>{r.country || r.region} · {r.sector} · {r.incident_count}</li>)}</ul> : <div className="empty-state">No concentrated regional risk for the selected filters.</div>}</section>
      <section className="card"><h3>Trust confidence</h3><p>{summary.trust_summary?.label || 'Trust signals stable'}</p></section>
    </AppShell>
  )
}
