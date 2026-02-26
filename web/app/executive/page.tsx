'use client'

import { useEffect, useState } from 'react'
import { api } from '../../lib/api'

export default function ExecutivePage() {
  const [queue, setQueue] = useState<any[]>([])
  const [world, setWorld] = useState<any[]>([])
  useEffect(() => { api.queue().then((r) => setQueue(r.slice(0, 5))).catch(() => setQueue([])); api.worldMap().then((r) => setWorld(r.slice(0, 5))).catch(() => setWorld([])) }, [])
  return <main style={{ padding: 24 }}><h1>Executive Summary</h1><h3>Top Risks</h3><ul>{queue.map((r) => <li key={r.id}>{r.symbol} · {r.severity_band} · {r.priority_score?.toFixed?.(2)}</li>)}</ul><h3>Top Countries/Sectors</h3><ul>{world.map((r, i) => <li key={i}>{r.country} · {r.sector} · {r.incident_count}</li>)}</ul></main>
}
