'use client'
import { useEffect, useState } from 'react'
import AppShell from '../../components/AppShell'
import LoadingState from '../../components/LoadingState'
import { api, type GlobalFilters } from '../../lib/api'
import { subscribeTrustPatches } from '../../lib/stream'

export default function TrustPage() {
  const [filters, setFilters] = useState<GlobalFilters>({ window: '24h' })
  const [data, setData] = useState<any>(null)
  useEffect(() => { api.trust(filters).then(setData) }, [JSON.stringify(filters)])
  useEffect(() => subscribeTrustPatches((p) => setData((d: any) => ({ ...(d || {}), ...p }))), [])
  return <AppShell title="Trust" subtitle="Model and data quality posture" filters={filters} setFilters={setFilters}>
    {!data ? <LoadingState /> : <>
      <div className="grid-3">
        <div className="card"><h3>DQ metrics</h3><p>Warnings: {data?.dqTotals?.warnings ?? 0}</p></div>
        <div className="card"><h3>Model degraded</h3><p>{data.modelDegradedCount ?? 0}</p></div>
        <div className="card"><h3>Cache health</h3><p>{data.cacheHealth || 'ok'}</p></div>
      </div>
      <div className="timeline-lane">missingness</div>
      <div className="timeline-lane">duplicates</div>
      <div className="timeline-lane">late_events</div>
    </>}
  </AppShell>
}
