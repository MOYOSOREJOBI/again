'use client'

import { useEffect, useMemo, useState } from 'react'
import Link from 'next/link'
import { api, QueueRow } from '../../lib/api'

export default function QueuePage() {
  const [rows, setRows] = useState<QueueRow[]>([])
  const [country, setCountry] = useState('')
  const [sector, setSector] = useState('')

  const query = useMemo(() => {
    const p = new URLSearchParams()
    if (country) p.set('country', country)
    if (sector) p.set('sector', sector)
    const s = p.toString()
    return s ? `?${s}` : ''
  }, [country, sector])

  useEffect(() => {
    api.queue(query).then(setRows).catch(() => setRows([]))
    const id = setInterval(() => api.queue(query).then(setRows).catch(() => null), 6000)
    return () => clearInterval(id)
  }, [query])

  return (
    <main style={{ padding: 24 }}>
      <h1>Ranked Incident Queue</h1>
      <div style={{ display: 'flex', gap: 8 }}>
        <input placeholder="country" value={country} onChange={(e) => setCountry(e.target.value)} />
        <input placeholder="sector" value={sector} onChange={(e) => setSector(e.target.value)} />
      </div>
      <table><thead><tr><th>ID</th><th>Symbol</th><th>Severity</th><th>Priority</th><th>Status</th><th>Action</th></tr></thead><tbody>
        {rows.map((r) => <tr key={r.id}><td><Link href={`/incident/${r.id}`}>{r.id}</Link></td><td>{r.symbol}</td><td>{r.severity_band}</td><td>{r.priority_score?.toFixed?.(3)}</td><td>{r.status}</td><td>{r.recommended_action || '-'}</td></tr>)}
      </tbody></table>
    </main>
  )
}
