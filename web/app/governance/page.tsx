'use client'

import { useEffect, useState } from 'react'
import AppShell, { useGlobalFilters } from '../../components/AppShell'
import { api } from '../../lib/api'

export default function GovernancePage() {
  const [summary, setSummary] = useState<any>({ deployed_models: [], replay_jobs: [], threshold_changes: [] })
  const { filters, apply } = useGlobalFilters()

  useEffect(() => { api.governanceSummary().then(setSummary) }, [])

  return (
    <AppShell title="Governance" subtitle="Admin control plane for model versions, thresholds, and replay approvals." filters={filters} setFilters={apply}>
      <div className="grid-3">
        <div className="card"><h3>Model version</h3><p>{summary.model_version || 'not reported'}</p></div>
        <div className="card"><h3>Feature set</h3><p>{summary.feature_set_version || 'not reported'}</p></div>
        <div className="card"><h3>Approval state</h3><p>{summary.approval_state || 'pending review'}</p></div>
      </div>
      <section className="card"><h3>Replay jobs</h3>{summary.replay_jobs?.length ? <ul>{summary.replay_jobs.map((r: any) => <li key={r.id}>{r.id} · {r.status}</li>)}</ul> : <div className="empty-state">No replay jobs registered yet.</div>}</section>
      <section className="card"><h3>Threshold change log</h3>{summary.threshold_changes?.length ? <ul>{summary.threshold_changes.map((c: any, i: number) => <li key={i}>{c.when}: {c.change}</li>)}</ul> : <div className="empty-state">No threshold modifications found for current window.</div>}</section>
    </AppShell>
  )
}
