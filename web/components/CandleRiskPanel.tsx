"use client";

import { useMemo } from 'react'

export function CandleRiskPanel({ points }: { points: Array<{ t: string; risk: number }> }) {
  const max = useMemo(() => Math.max(1, ...points.map((p) => p.risk || 0)), [points])
  const width = 640
  const height = 220
  const path = points.map((p, i) => {
    const x = (i / Math.max(1, points.length - 1)) * width
    const y = height - ((p.risk || 0) / max) * (height - 20)
    return `${i === 0 ? 'M' : 'L'}${x},${y}`
  }).join(' ')
  return <div className='card'>
    <h3>Risk trend</h3>
    <svg viewBox={`0 0 ${width} ${height}`} role='img' aria-label='risk trend chart'>
      <rect x='0' y='0' width={width} height={height} fill='#0f172a' rx='8' />
      <path d={path} fill='none' stroke='#38bdf8' strokeWidth='2' />
    </svg>
  </div>
}
