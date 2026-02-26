'use client'

import { useMemo, useState } from 'react'

type Row = { country: string; region: string; sector: string; industry?: string; incident_count: number }

export default function WorldRiskMap({ rows, onFilter }: { rows: Row[]; onFilter?: (f: { country?: string; region?: string; sector?: string; industry?: string; time_window?: string }) => void }) {
  const [hover, setHover] = useState<Row | null>(null)
  const [timeWindow, setTimeWindow] = useState('24h')
  const top = useMemo(() => rows.slice(0, 20), [rows])

  return (
    <section>
      <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
        <h3 style={{ margin: 0 }}>World Risk Map (2D fallback)</h3>
        <select value={timeWindow} onChange={(e) => { setTimeWindow(e.target.value); onFilter?.({ time_window: e.target.value }) }}>
          <option value="1h">1h</option>
          <option value="24h">24h</option>
          <option value="7d">7d</option>
        </select>
      </div>
      {hover ? <p>Hover: {hover.country} · {hover.region} · {hover.sector} · incidents={hover.incident_count}</p> : <p>Hover a row to inspect risk context.</p>}
      <table>
        <thead><tr><th>Country</th><th>Region</th><th>Sector</th><th>Industry</th><th>Incidents</th></tr></thead>
        <tbody>
          {top.map((r, i) => (
            <tr key={`${r.country}-${r.sector}-${i}`} onMouseEnter={() => setHover(r)} onClick={() => onFilter?.({ country: r.country, region: r.region, sector: r.sector, industry: r.industry, time_window: timeWindow })} style={{ cursor: 'pointer' }}>
              <td>{r.country}</td><td>{r.region}</td><td>{r.sector}</td><td>{r.industry || '-'}</td><td>{r.incident_count}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  )
}
