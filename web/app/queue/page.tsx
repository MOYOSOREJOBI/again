'use client'
import { useEffect, useMemo, useState } from 'react'
import AppShell from '../../components/AppShell'
import LoadingState from '../../components/LoadingState'
import EmptyState from '../../components/EmptyState'
import DegradedState from '../../components/DegradedState'
import { api, type GlobalFilters } from '../../lib/api'
import { subscribeQueuePatches } from '../../lib/stream'

export default function QueuePage() {
  const [filters, setFilters] = useState<GlobalFilters>({ time_window: '24h' })
  const [rows, setRows] = useState<any[] | null>(null)
  const [degraded, setDegraded] = useState(false)

  useEffect(() => { api.queue(filters).then(setRows).catch(() => { setRows([]); setDegraded(true) }) }, [JSON.stringify(filters)])
  useEffect(() => subscribeQueuePatches((patch) => {
    setRows((prev) => {
      const arr = [...(prev || [])]
      const idx = arr.findIndex((r) => String(r.id) === String(patch.incident?.id))
      if (patch.type === 'delete' && idx >= 0) arr.splice(idx, 1)
      if (patch.type === 'upsert') idx >= 0 ? arr.splice(idx, 1, patch.incident) : arr.push(patch.incident)
      return arr
    })
  }, () => setDegraded(true)), [])

  const sorted = useMemo(() => (rows || []).slice().sort((a, b) => (b.priorityScore || b.priority_score || 0) - (a.priorityScore || a.priority_score || 0)), [rows])

  return <AppShell title="Queue" subtitle="Ranked incident queue" filters={filters} setFilters={setFilters}>
        {rows === null ? <LoadingState label="Loading queue" /> : sorted.length === 0 ? <EmptyState message="Queue is clear for current filters." /> :
      <div className="card">{sorted.map((r) => <div key={r.id} className="queue-row"><div><strong>{r.symbol}</strong> <span className="severity-chip">{r.severityBand || r.severity}</span></div><div className='muted'>Priority {(r.priorityScore || r.priority_score || 0).toFixed?.(1) || 0}</div><div>{r.recommendedAction || r.recommended_action || 'watch'}</div></div>)}</div>}
    {degraded && <DegradedState />}
  </AppShell>
}
