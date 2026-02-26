'use client'
import { useEffect, useState } from 'react'
import AppShell, { useGlobalFilters } from '../../components/AppShell'
import LoadingState from '../../components/LoadingState'
import { api } from '../../lib/api'

export default function GovernancePage() {
  const [summary, setSummary] = useState<any>(null)
  const [me, setMe] = useState<any>({ role: 'viewer' })
  const { filters, apply } = useGlobalFilters()
  useEffect(() => { api.governanceSummary().then(setSummary); api.me().then(setMe) }, [])
  if (!summary) return <AppShell title="Governance" subtitle="Admin control plane" filters={filters} setFilters={apply}><LoadingState /></AppShell>
  return <AppShell title="Governance" subtitle="Admin control plane" filters={filters} setFilters={apply}>
    {me?.role !== 'admin' && <div className="empty-state">Governance controls are admin-only.</div>}
    <div className="grid-3">
      <div className="card"><h3>Model Registry</h3>{(summary.modelLineage || []).map((m: any, i: number) => <div key={i} className="kv"><span>{m.model}</span><span>{m.version} ({m.state})</span></div>)}</div>
      <div className="card"><h3>Replay Jobs</h3>{(summary.replayJobs || []).map((j: any) => <div key={j.id} className="kv"><span>{j.id}</span><span>{j.status}</span></div>)}</div>
      <div className="card"><h3>Threshold Changes</h3><p>{(summary.thresholdChanges || []).length} changes</p></div>
    </div>
  </AppShell>
}
