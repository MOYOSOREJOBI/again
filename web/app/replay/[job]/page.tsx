'use client'

import { useEffect, useState } from 'react'
import { api } from '../../../lib/api'
import AppShell, { useGlobalFilters } from '../../../components/AppShell'

export default function ReplayPage({ params }: { params: { job: string } }) {
  const [job, setJob] = useState<any>(null)
  const { filters, apply } = useGlobalFilters()
  useEffect(() => { api.replay(params.job).then(setJob) }, [params.job])

  return (
    <AppShell title={`Replay ${params.job}`} subtitle="Metadata-first replay workspace" filters={filters} setFilters={apply}>
      {!job ? <div className="empty-state">Replay metadata unavailable.</div> : <>
        <div className="grid-3">
          <div className="card"><h3>Status</h3><p>{job.status}</p></div>
          <div className="card"><h3>Window</h3><p>{job.started_at || 'not available'} → {job.completed_at || 'running'}</p></div>
          <div className="card"><h3>Versions</h3><p>Model {job.model_version || 'captured in metadata'}</p></div>
        </div>
        <section className="card">
          <h3>Reproducibility scope</h3>
          <p>Sentinel currently provides metadata-first replay. You can reproduce configuration, version, and timing context. Full lane recompute is not yet enabled in this build.</p>
          <pre>{JSON.stringify(job.diff_summary || { note: 'No diff summary available yet.' }, null, 2)}</pre>
        </section>
      </>}
    </AppShell>
  )
}
