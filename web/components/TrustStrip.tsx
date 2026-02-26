'use client'

export default function TrustStrip({ degraded }: { degraded: number }) {
  return <div className="trust-strip">Trust: {degraded === 0 ? 'Stable' : `Degraded (${degraded})`}</div>
}
