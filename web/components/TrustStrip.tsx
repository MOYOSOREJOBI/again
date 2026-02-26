'use client'

type Props = {
  trustState?: 'stable' | 'degraded' | 'unavailable'
  modelUnavailable?: boolean
  dqWarning?: string
}

export default function TrustStrip({ trustState = 'stable', modelUnavailable = false, dqWarning = '' }: Props) {
  const label = trustState === 'stable' ? 'Stable' : trustState === 'degraded' ? 'Degraded' : 'Unavailable'
  return (
    <div className="trust-strip" role="status" aria-live="polite">
      <strong>Trust:</strong> {label}
      {modelUnavailable ? ' · Fallback model active' : ' · Model healthy'}
      {dqWarning ? ` · DQ: ${dqWarning}` : ''}
    </div>
  )
}
