'use client'

import type { CountryAgg } from '../lib/types'

const colorBySafety: Record<string, string> = {
  Stable: '#1f9d55',
  Elevated: '#d69e2e',
  'High Risk': '#dd6b20',
  Critical: '#e53e3e',
  'Data Unreliable': '#718096',
}

type Props = {
  data?: CountryAgg[]
  rows?: any[]
  selectedCountryCode?: string
  selectedRegion?: string
  onSelectCountry?: (code?: string) => void
  onSelectRegion?: (region: string) => void
  timeWindow?: '1h' | '24h' | '7d' | 'custom'
  industry?: string
}

const GEO: { code: string; name: string; x: number; y: number; w: number; h: number }[] = [
  { code: 'US', name: 'United States', x: 110, y: 110, w: 70, h: 25 },
  { code: 'GB', name: 'United Kingdom', x: 250, y: 85, w: 16, h: 14 },
  { code: 'DE', name: 'Germany', x: 270, y: 95, w: 16, h: 14 },
  { code: 'JP', name: 'Japan', x: 430, y: 115, w: 16, h: 14 },
  { code: 'BR', name: 'Brazil', x: 185, y: 180, w: 24, h: 28 },
  { code: 'AU', name: 'Australia', x: 430, y: 205, w: 32, h: 20 },
]

export default function WorldRiskMap({ data, rows, selectedCountryCode, selectedRegion, onSelectCountry, onSelectRegion }: Props) {
  const normalized: CountryAgg[] = (data || (rows || []).map((r: any) => ({
    countryCode: r.countryCode || r.country || 'XX',
    countryName: r.countryName || r.country || 'Unknown',
    incidentCount: r.incidentCount || r.incident_count || 0,
    avgCompositeRisk: r.avgCompositeRisk || r.incident_pressure || 0,
    maxSafetyLevel: r.maxSafetyLevel || 'Stable',
    trustState: r.trustState || r.trust_state || 'healthy',
    topIndustry: r.topIndustry || r.top_sector,
  })))
  const byCode = new Map(normalized.map((d) => [d.countryCode.toUpperCase(), d]))
  if (normalized.length === 0) return <div className='card'><h3>World risk map</h3><div className='empty-state'>No regional aggregates yet. Keep filters broad to populate global heat.</div><aside className='map-legend'>Top regions</aside></div>
  return <div className='card'>
    <h3>World risk map</h3>
    <svg viewBox='0 0 560 280' role='img' aria-label='world risk map'>
      <rect x='0' y='0' width='560' height='280' fill='#0f172a' rx='8' />
      {GEO.map((g) => {
        const row = byCode.get(g.code)
        const fill = row ? colorBySafety[row.maxSafetyLevel] || '#1f9d55' : '#334155'
        return <g key={g.code}>
          <rect x={g.x} y={g.y} width={g.w} height={g.h} fill={fill} stroke={(selectedCountryCode === g.code || selectedRegion === g.code) ? '#63b3ed' : '#2d3748'} onClick={() => { onSelectCountry?.(g.code); onSelectRegion?.(g.code) }} />
          <title>{row ? `${row.countryName} | incident count ${row.incidentCount} | avg risk ${row.avgCompositeRisk.toFixed(2)} | ${row.maxSafetyLevel} | ${row.trustState} | ${row.topIndustry || 'n/a'}` : g.name}</title>
        </g>
      })}
    </svg><aside className="map-legend">Top regions</aside>
  </div>
}
