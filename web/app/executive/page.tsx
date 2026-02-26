'use client'
import { useEffect, useState } from 'react'
import AppShell, { useGlobalFilters } from '../../components/AppShell'
import { api } from '../../lib/api'

export default function ExecutivePage() {
  const [summary, setSummary] = useState<any>({})
  const { filters, apply } = useGlobalFilters()
  useEffect(() => { api.executiveSummary(filters).then(setSummary) }, [JSON.stringify(filters)])
  return <AppShell title="Executive Summary" subtitle="Plain-English current risk posture" filters={filters} setFilters={apply}>
    <div className="grid-3"><div className="card"><h3>Top 5 risks</h3><p>{(summary.topRisks || []).length}</p></div><div className="card"><h3>Country concentration</h3><p>{(summary.countryConcentration || []).length}</p></div><div className="card"><h3>Industry concentration</h3><p>{(summary.industryConcentration || []).length}</p></div></div>
    <section className="card"><h3>Risk memo</h3><p>{summary.riskMemo || 'No material change.'}</p></section>
  </AppShell>
}
