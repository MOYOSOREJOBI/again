"use client";

export type CountryAgg = { iso2: string; name: string; risk: number; highRisk: number; p95: number };

export function RiskGlobe({ data, onSelect }: { data: CountryAgg[]; onSelect: (iso2: string) => void }) {
  return <div className='card'>
    <h3>Risk globe</h3>
    <div style={{ width: '100%', height: 320, borderRadius: 12, background: 'radial-gradient(circle at 30% 30%, #334155, #0f172a)' }}>
      <div style={{ padding: 12, color: '#e2e8f0' }}>Interactive globe placeholder (offline-safe build) — click a country:</div>
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, padding: 12 }}>
        {data.slice(0, 24).map((c) => <button key={c.iso2} onClick={() => onSelect(c.iso2)}>{c.name} ({c.iso2})</button>)}
      </div>
    </div>
  </div>
}
