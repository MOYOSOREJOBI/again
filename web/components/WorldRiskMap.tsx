'use client'

import { useMemo, useState } from 'react'
import { GlobalFilters, WorldMapRow } from '../lib/api'

function intensity(count: number, max: number) {
  if (!max) return 1
  const ratio = count / max
  if (ratio > 0.8) return 5
  if (ratio > 0.6) return 4
  if (ratio > 0.4) return 3
  if (ratio > 0.2) return 2
  return 1
}

export default function WorldRiskMap({ rows, filters, onFilter }: { rows: WorldMapRow[]; filters?: GlobalFilters; onFilter?: (f: Partial<GlobalFilters>) => void }) {
  const [hover, setHover] = useState<WorldMapRow | null>(null)
  const top = useMemo(() => rows.slice(0, 18), [rows])
  const max = useMemo(() => Math.max(0, ...rows.map((r) => r.incident_count)), [rows])

  return (
    <section className="card">
      <div className="topbar" style={{ alignItems: 'center' }}>
        <h3>World Risk Map (2D-first)</h3>
        <select value={filters?.time_window || '24h'} onChange={(e) => onFilter?.({ time_window: e.target.value as any })}>
          <option value="now">now</option><option value="1h">1h</option><option value="24h">24h</option><option value="7d">7d</option>
        </select>
      </div>
      {rows.length === 0 ? <div className="empty-state">No regional aggregates yet for the selected filters.</div> : (
        <div className="map-grid">
          <div className="map-canvas">
            {top.map((r, i) => (
              <div key={`${r.country}-${i}`} className={`map-cell map-intensity-${intensity(r.incident_count, max)}`} onMouseEnter={() => setHover(r)} onClick={() => onFilter?.({ country: r.country, region: r.region, sector: r.sector, industry: r.industry })}>
                <strong>{r.country || r.region}</strong>
                <p>{r.incident_count} incidents</p>
                <small>{r.sector}</small>
              </div>
            ))}
          </div>
          <div className="card">
            <h3>Hover context</h3>
            {hover ? (
              <>
                <div className="kv"><span>Region</span><strong>{hover.region}</strong></div>
                <div className="kv"><span>Country</span><strong>{hover.country}</strong></div>
                <div className="kv"><span>Incidents</span><strong>{hover.incident_count}</strong></div>
                <div className="kv"><span>Top sector</span><strong>{hover.sector}</strong></div>
                <div className="kv"><span>Trust</span><strong>{hover.trust_state || 'stable'}</strong></div>
              </>
            ) : <div className="empty-state">Hover a region to inspect trust and pressure.</div>}
          </div>
        </div>
      )}
      {rows.length > 0 ? <table className="table"><thead><tr><th>Country</th><th>Region</th><th>Sector</th><th>Incidents</th></tr></thead><tbody>{rows.slice(0, 6).map((r, i) => <tr key={i}><td>{r.country}</td><td>{r.region}</td><td>{r.sector}</td><td>{r.incident_count}</td></tr>)}</tbody></table> : null}
    </section>
  )
}
