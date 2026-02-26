'use client'

import type { CountryAgg } from '../lib/types'

const COLORS: Record<string, string> = {
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

// simplified world silhouette with country regions (geo surface, not grid)
const SHAPES = [
  { code: 'US', d: 'M60,90 L170,85 L190,120 L80,130 Z', label: 'United States' },
  { code: 'GB', d: 'M250,70 L265,68 L268,82 L252,84 Z', label: 'United Kingdom' },
  { code: 'DE', d: 'M275,82 L292,80 L295,95 L278,98 Z', label: 'Germany' },
  { code: 'BR', d: 'M175,150 L220,145 L235,205 L185,220 Z', label: 'Brazil' },
  { code: 'JP', d: 'M445,105 L458,102 L462,120 L448,123 Z', label: 'Japan' },
  { code: 'AU', d: 'M430,210 L485,205 L495,235 L438,242 Z', label: 'Australia' },
]

export default function WorldRiskMap({ data, rows, selectedCountryCode, selectedRegion, onSelectCountry, onSelectRegion }: Props) {
  const normalized: CountryAgg[] = data || (rows || []).map((r: any) => ({
    countryCode: (r.countryCode || r.country || 'XX').toUpperCase(),
    countryName: r.countryName || r.country || 'Unknown',
    incidentCount: r.incidentCount || r.incident_count || 0,
    avgCompositeRisk: r.avgCompositeRisk || r.incident_pressure || 0,
    maxSafetyLevel: r.maxSafetyLevel || 'Stable',
    trustState: r.trustState || r.trust_state || 'healthy',
    topIndustry: r.topIndustry || r.top_sector,
  }))
  const byCode = new Map(normalized.map((d) => [d.countryCode.toUpperCase(), d]))

  if (normalized.length === 0) return <div className='card'><h3>World risk map</h3><div className='empty-state'>No regional aggregates yet. Keep filters broad to populate global heat.</div><aside className='map-legend'>Top regions</aside></div>

  return <div className='card'>
    <h3>World risk map</h3>
    <svg viewBox='0 0 560 280' role='img' aria-label='world risk map'>
      <rect x='0' y='0' width='560' height='280' fill='#0f172a' rx='8' />
      {SHAPES.map((s) => {
        const row = byCode.get(s.code)
        const fill = row ? COLORS[row.maxSafetyLevel] || '#1f9d55' : '#334155'
        return <path
          key={s.code}
          d={s.d}
          fill={fill}
          stroke={(selectedCountryCode === s.code || selectedRegion === s.code) ? '#63b3ed' : '#2d3748'}
          strokeWidth={1}
          onClick={() => { onSelectCountry?.(s.code); onSelectRegion?.(s.code) }}
        >
          <title>{row ? `${row.countryName} | incident count ${row.incidentCount} | avg risk ${row.avgCompositeRisk.toFixed(2)} | ${row.maxSafetyLevel} | ${row.trustState} | ${row.topIndustry || 'n/a'}` : s.label}</title>
        </path>
      })}
    </svg>
    <aside className='map-legend'>Top regions</aside>
  </div>
}
