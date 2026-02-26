'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '../../../lib/api'

export default function IncidentDetail({ params }: { params: { id: string } }) {
  const [row, setRow] = useState<any>(null)
  const [reason, setReason] = useState('Escalation required')
  const [caseId, setCaseId] = useState<number | null>(null)
  useEffect(() => { api.incident(params.id).then(setRow).catch(() => setRow(null)) }, [params.id])
  if (!row) return <main style={{ padding: 24 }}>Incident not found or unauthorized.</main>
  return (
    <main style={{ padding: 24 }}>
      <h1>Incident {row.id}</h1>
      <p>Anomaly/composite risk: {row.composite_risk} · escalation: {row.escalation_probability} · priority: {row.priority_score}</p>
      <p>Confidence: {row.confidence} · trust: {row.trust_state}</p>
      <p>Top drivers: {(row.top_drivers || []).filter(Boolean).join(', ') || 'Unavailable'}</p>
      <p>Replay: <Link href={`/replay/${row.id}`}>Open replay metadata</Link></p>
      <div>
        <input value={reason} onChange={(e) => setReason(e.target.value)} />
        <button onClick={async () => { const out = await api.createCase(row.id, reason).catch(() => null); if (out?.case_id) setCaseId(out.case_id) }}>Promote to case</button>
        {caseId ? <Link href={`/case/${caseId}`}>Open case {caseId}</Link> : null}
      </div>
    </main>
  )
}
