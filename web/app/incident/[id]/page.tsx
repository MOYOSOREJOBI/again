'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { api } from '../../../lib/api'
import AppShell, { useGlobalFilters } from '../../../components/AppShell'

export default function IncidentDetail({ params }: { params: { id: string } }) {
  const [row, setRow] = useState<any>(null)
  const [reason, setReason] = useState('Escalation required')
  const [caseId, setCaseId] = useState<number | null>(null)
  const { filters, apply } = useGlobalFilters()
  useEffect(() => { api.incident(params.id).then(setRow) }, [params.id])

  return (
    <AppShell title={`Incident ${params.id}`} subtitle="Why this fired, confidence caveats, and case promotion." filters={filters} setFilters={apply}>
      {!row?.id ? <div className="empty-state">Incident not found or unauthorized.</div> : <>
        <div className="grid-3">
          <div className="card"><h3>Unusualness</h3><p>{row.composite_risk?.toFixed?.(2)}</p></div>
          <div className="card"><h3>Escalation likelihood</h3><p>{row.escalation_probability?.toFixed?.(2)}</p></div>
          <div className="card"><h3>Confidence</h3><p>{row.confidence?.toFixed?.(2)} ({row.trust_state || 'stable'})</p></div>
        </div>
        <div className="grid-2">
          <section className="card">
            <h3>Why this fired</h3>
            <div className="kv"><span>Severity band</span><strong>{row.severity_band}</strong></div>
            <div className="kv"><span>Trust penalty</span><strong>{row.trust_state === 'degraded' ? 'Applied' : 'None'}</strong></div>
            <div className="kv"><span>Top drivers</span><strong>{(row.top_drivers || []).filter(Boolean).join(', ') || 'No deterministic driver breakdown yet'}</strong></div>
            <div className="kv"><span>Delta vs baseline</span><strong>{row.rank_reason || 'no baseline delta payload'}</strong></div>
            <p className="muted" style={{ marginTop: 8 }}>Explanation is deterministic top-driver output. SHAP is not enabled in this build.</p>
          </section>
          <section className="card">
            <h3>Data quality and caveats</h3>
            <ul>
              <li>Stale feed: {row.trust_state === 'degraded' ? 'possible' : 'not detected'}</li>
              <li>Incomplete context: {row.confidence && row.confidence < 0.7 ? 'possible' : 'unlikely'}</li>
              <li>Deterministic fallback used: {row.model_version ? 'no' : 'yes'}</li>
            </ul>
            <p style={{ marginTop: 10 }}><Link href={`/replay/${row.id}`}>Open replay metadata</Link></p>
            <div className="form-grid" style={{ marginTop: 10 }}>
              <input value={reason} onChange={(e) => setReason(e.target.value)} />
              <button onClick={async () => { const out = await api.createCase(row.id, reason).catch(() => null); if (out?.case_id) setCaseId(out.case_id) }}>Promote to case</button>
              {caseId ? <Link href={`/case/${caseId}`}>Open case {caseId}</Link> : null}
            </div>
          </section>
        </div>
      </>}
    </AppShell>
  )
}
