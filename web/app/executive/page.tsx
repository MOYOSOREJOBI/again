'use client'

import { useEffect, useState } from 'react'
import AppShell, { useGlobalFilters } from '../../components/AppShell'
import { api } from '../../lib/api'

export default function ExecutivePage() {
  const [summary, setSummary] = useState<any>({ top_risks: [], hot_regions: [], trust_summary: {} })
  const { filters, apply } = useGlobalFilters()
  useEffect(() => { api.executiveSummary(filters).then(setSummary) }, [filters])

  return (
    <AppShell title="Executive Summary" subtitle="Plain-English current risk posture and concentration overview." filters={filters} setFilters={apply}>
      <div className="grid-3">
        <div className="card"><h3>Top risks</h3><pre>{JSON.stringify(summary.top_risks || [], null, 2)}</pre></div>
        <div className="card"><h3>Hot regions</h3><pre>{JSON.stringify(summary.hot_regions || [], null, 2)}</pre></div>
        <div className="card"><h3>Trust summary</h3><pre>{JSON.stringify(summary.trust_summary || {}, null, 2)}</pre></div>
      </div>
      <section className="card"><h3>Narrative</h3><p>{summary.what_changed || 'Risk concentration is tracked continuously across global venues and sectors.'}</p></section>
    </AppShell>
  )
}
