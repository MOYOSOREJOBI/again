'use client'

type Row = { country: string; region: string; top_sector?: string; incident_count: number; incident_pressure?: number; trust_state?: string }

export default function WorldRiskMap({ rows, selectedRegion, onSelectRegion }: { rows: Row[]; selectedRegion?: string; onSelectRegion?: (region: string) => void }) {
  const max = Math.max(1, ...rows.map((r) => r.incident_count || 0))
  if (!rows.length) return <div className="card"><h3>World risk map</h3><div className="map-grid"><div className="empty-state">No regional aggregates yet. Keep filters broad to populate global heat.</div><aside className="map-legend">Legend ready</aside></div></div>
  return (
    <div className="card">
      <h3>World risk map</h3>
      <div className="map-grid">
        <div className="map-canvas">
          {rows.map((row) => {
            const intensity = Math.min(5, Math.max(1, Math.ceil((row.incident_count / max) * 5)))
            const selected = selectedRegion && selectedRegion === row.region
            return (
              <button key={`${row.country}-${row.region}`} className={`map-cell map-intensity-${intensity} ${selected ? 'selected' : ''}`} title={`${row.country} (${row.region}) incidents ${row.incident_count}`} onClick={() => onSelectRegion?.(row.region)}>
                <strong>{row.country}</strong>
                <div className="muted">{row.region} · {row.incident_count} incidents</div>
                <div className="badge">{row.trust_state || 'stable'}</div>
              </button>
            )
          })}
        </div>
        <aside className="map-legend">
          <h4>Top regions</h4>
          {rows.slice(0, 5).map((r) => <div key={`${r.country}-${r.region}`} className="kv"><span>{r.region}</span><span>{r.incident_count}</span></div>)}
        </aside>
      </div>
    </div>
  )
}
