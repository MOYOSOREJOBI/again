'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import TrustStrip from '../../components/TrustStrip'
import WorldRiskMap from '../../components/WorldRiskMap'
import { api } from '../../lib/api'

export default function CommandCenterPage() {
  const [summary, setSummary] = useState<any>({ open_incidents: 0, high_risk_incidents: 0 })
  const [trust, setTrust] = useState<any>({ degraded: [] })
  const [world, setWorld] = useState<any[]>([])

  const load = async (params = '') => {
    const [s, t, w] = await Promise.all([api.commandCenter().catch(() => summary), api.trust().catch(() => trust), api.worldMap(params).catch(() => world)])
    setSummary(s); setTrust(t); setWorld(w)
  }

  useEffect(() => { load() }, [])

  return (
    <main style={{ padding: 24 }}>
      <h1>Command Center</h1>
      <TrustStrip trustState={trust.degraded?.length ? 'degraded' : 'stable'} modelUnavailable={Boolean(trust.model_unavailable_count)} dqWarning={trust.degraded?.length ? `${trust.degraded.length} impacted symbols` : ''} />
      <p>Open incidents: {summary.open_incidents} · High risk: {summary.high_risk_incidents}</p>
      <p><Link href="/queue">Open Queue</Link> · <Link href="/trust">Trust</Link> · <Link href="/executive">Executive</Link></p>
      <WorldRiskMap rows={world} onFilter={(f) => {
        const q = new URLSearchParams(f as any).toString()
        load(q ? `?${q}` : '')
      }} />
    </main>
  )
}
