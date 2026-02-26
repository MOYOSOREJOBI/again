'use client'

import { useEffect, useState } from 'react'
import TrustStrip from '../../components/TrustStrip'
import { api } from '../../lib/api'

export default function TrustPage() {
  const [trust, setTrust] = useState<any>(null)
  useEffect(() => { api.trust().then(setTrust).catch(() => setTrust(null)) }, [])
  if (!trust) return <main style={{ padding: 24 }}>Trust data unavailable.</main>
  return (
    <main style={{ padding: 24 }}>
      <h1>Trust</h1>
      <TrustStrip trustState={trust.degraded?.length ? 'degraded' : 'stable'} modelUnavailable={Boolean(trust.model_unavailable_count)} dqWarning={`missing=${trust.missingness_rate} dup=${trust.duplicate_rate}`} />
      <p>Freshness (sec): {trust.freshness_seconds} · Out-of-order: {trust.out_of_order_rate}</p>
      <ul>{(trust.degraded || []).map((d: any, i: number) => <li key={i}>{d.symbol}: {d.alert_count}</li>)}</ul>
    </main>
  )
}
