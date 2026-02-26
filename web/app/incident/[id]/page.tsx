'use client'
import { useEffect, useState } from 'react'
import { useParams, useRouter } from 'next/navigation'
import AppShell from '../../../components/AppShell'
import WhyThisFiredPanel from '../../../components/WhyThisFiredPanel'
import { api } from '../../../lib/api'

export default function IncidentPage() {
  const params = useParams<{ id: string }>()
  const router = useRouter()
  const [data, setData] = useState<any>(null)
  const [promoteState, setPromoteState] = useState('')

  useEffect(() => { if (params?.id) api.incident(params.id).then(setData) }, [params?.id])

  const promote = async () => {
    if (!params?.id) return
    setPromoteState('Promoting…')
    try {
      const created = await api.promoteIncidentCase(params.id, 'promoted from incident detail')
      if (created?.case_id) {
        setPromoteState(`Promoted as case #${created.case_id}`)
        router.push(`/case/${created.case_id}`)
        return
      }
      setPromoteState('Promoted')
    } catch {
      setPromoteState('Promotion failed or forbidden')
    }
  }

  return <AppShell title={`Incident ${params?.id || ''}`} subtitle="Score evidence and explanation">
    <div className="card"><h3>Score header</h3><p>Priority {data?.score_header?.priority ?? '-'}</p></div>
    <section className="card"><h3>Explanation</h3><p>{data?.explanation_text || '-'}</p></section>
    <section className="card"><h3>Model lineage</h3><p>{data?.model_version || 'unknown'}</p></section>
    <button onClick={promote}>Promote to Case</button>
    {promoteState ? <p className="muted">{promoteState}</p> : null}
    <WhyThisFiredPanel drivers={data?.top_drivers || []} caveats={data?.caveats || []} />
  </AppShell>
}
