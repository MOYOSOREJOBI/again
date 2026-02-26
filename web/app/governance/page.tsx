'use client'

import { useEffect, useState } from 'react'
import AppShell, { useGlobalFilters } from '../../components/AppShell'
import { api } from '../../lib/api'

export default function GovernancePage() {
  const [summary, setSummary] = useState<any>({ model_lineage: [], replay_jobs: [], threshold_changes: [] })
  const [me, setMe] = useState<any>({ role: 'viewer' })
  const { filters, apply } = useGlobalFilters()

  useEffect(() => { api.governanceSummary().then(setSummary); api.me().then(setMe) }, [])
  const admin = me?.role === 'admin'

  return (
    <AppShell title="Governance" subtitle="Admin control plane for model status, trust posture, and audit lineage." filters={filters} setFilters={apply}>
      {!admin ? <div className="empty-state">Governance controls are admin-only. Viewer/Analyst can review read-only trust surfaces.</div> : null}
      <div className="grid-3">
        <div className="card"><h3>Model lineage</h3><pre>{JSON.stringify(summary.model_lineage || [], null, 2)}</pre></div>
        <div className="card"><h3>Replay jobs</h3><pre>{JSON.stringify(summary.replay_jobs || [], null, 2)}</pre></div>
        <div className="card"><h3>Threshold changes</h3><pre>{JSON.stringify(summary.threshold_changes || [], null, 2)}</pre></div>
      </div>
      <section className="card"><h3>Trust warnings</h3><p>Fallback vs model-backed mode is surfaced from live services. No destructive governance action is exposed in this build.</p></section>
    </AppShell>
  )
}
