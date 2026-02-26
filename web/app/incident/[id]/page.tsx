'use client'
import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import AppShell from '../../../components/AppShell'
import WhyThisFiredPanel from '../../../components/WhyThisFiredPanel'
import { api } from '../../../lib/api'

export default function IncidentPage() {
  const params = useParams<{ id: string }>()
  const [data, setData] = useState<any>(null)
  useEffect(() => { if (params?.id) api.incident(params.id).then(setData) }, [params?.id])
  return <AppShell title={`Incident ${params?.id || ''}`} subtitle="Score evidence and explanation">
    <div className="card"><h3>Score header</h3><p>Priority {data?.score_header?.priority ?? '-'}</p></div>
    <section className="card"><h3>Explanation</h3><p>{data?.explanation_text||"-"}</p></section>
    <section className="card"><h3>Model lineage</h3><p>{data?.model_version||"unknown"}</p></section>
    <button>Promote to Case</button>
    <WhyThisFiredPanel drivers={data?.top_drivers || []} caveats={data?.caveats || []} />
  </AppShell>
}
