'use client'

import Link from 'next/link'

type Props = {
  trustState?: 'stable' | 'degraded' | 'unavailable'
  modelUnavailable?: boolean
  dqWarning?: string
  drift?: string
  role?: 'viewer' | 'analyst' | 'admin'
}

export default function TrustStrip({ trustState = 'stable', modelUnavailable = false, dqWarning = '', drift = 'stable', role = 'viewer' }: Props) {
  const simple = role === 'viewer'
  const className = `trust-strip ${trustState}`
  const headline = trustState === 'stable'
    ? 'Model healthy • No major DQ degradations'
    : trustState === 'degraded'
      ? 'Degraded trust signal • Review penalties'
      : 'Model unavailable • fallback scoring active'

  return (
    <div className={className} role="status" aria-live="polite">
      <div>
        <strong>Trust: {trustState}</strong>
        <p className="muted">{headline}</p>
        <p className="muted">
          {simple
            ? `Data quality ${dqWarning || 'within normal bounds'}`
            : `Drift: ${drift} • ${modelUnavailable ? 'fallback model active' : 'primary model available'} • DQ: ${dqWarning || 'clean'}`}
        </p>
      </div>
      {(trustState === 'degraded' || trustState === 'unavailable') ? <Link href="/trust"><button>Open Trust</button></Link> : null}
    </div>
  )
}
