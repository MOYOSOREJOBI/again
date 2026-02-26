'use client'

import { useEffect, useState } from 'react'
import { api } from '../../../lib/api'

export default function CasePage({ params }: { params: { id: string } }) {
  const [row, setRow] = useState<any>(null)
  const [note, setNote] = useState('')
  const [status, setStatus] = useState('investigating')

  const load = () => api.caseDetail(params.id).then(setRow).catch(() => setRow(null))
  useEffect(() => { load() }, [params.id])
  if (!row) return <main style={{ padding: 24 }}>Case unavailable or not found.</main>

  return (
    <main style={{ padding: 24 }}>
      <h1>Case {row.id}</h1>
      <p>Status: {row.status} · Owner: {row.owner || 'Unassigned'} · Incident: {row.incident_id}</p>
      <h3>Notes</h3>
      <ul>{(row.notes || []).map((n: any) => <li key={n.id}>{n.actor}: {n.note}</li>)}</ul>
      <input value={note} onChange={(e) => setNote(e.target.value)} placeholder="add note" />
      <button onClick={async () => { await api.caseNote(params.id, note).catch(() => null); setNote(''); load() }}>Add note</button>
      <h3>Evidence</h3>
      <ul>{(row.evidence || []).map((e: any) => <li key={e.id}>{e.evidence_type} · {e.reference_id}</li>)}</ul>
      <button onClick={async () => { await api.caseEvidence(params.id, 'incident', String(row.incident_id), { source: 'ui' }).catch(() => null); load() }}>Attach incident evidence</button>
      <div>
        <select value={status} onChange={(e) => setStatus(e.target.value)}><option>investigating</option><option>escalated</option><option>closed</option></select>
        <button onClick={async () => { await api.caseStatus(params.id, status).catch(() => null); load() }}>Update status</button>
      </div>
    </main>
  )
}
