'use client'
import { useEffect, useMemo, useState } from 'react'
import dynamic from 'next/dynamic'
import AppShell from '../../components/AppShell'
import TrustStrip from '../../components/TrustStrip'
import WorldRiskMap from '../../components/WorldRiskMap'
import { CandleRiskPanel } from '../../components/CandleRiskPanel'
import LoadingState from '../../components/LoadingState'
import { api, type GlobalFilters } from '../../lib/api'
import { subscribeCommandCenterPatches } from '../../lib/stream'
import { useSSE } from '../../lib/useSSE'

const RiskGlobe = dynamic(() => import('../../components/RiskGlobe').then(m => m.RiskGlobe), { ssr: false })

export default function CommandCenterPage() {
  const [filters, setFilters] = useState<GlobalFilters>({ window: '24h' })
  const [data, setData] = useState<any>(null)
  const [map, setMap] = useState<any>({ countries: [] })
  useEffect(() => { Promise.all([api.commandCenter(filters), api.worldMap(filters)]).then(([cc, wm]) => { setData(cc); setMap(wm) }) }, [JSON.stringify(filters)])
  useEffect(() => subscribeCommandCenterPatches((p) => setData((d: any) => ({ ...(d || {}), ...p })), () => {}), [])

  const [events, setEvents] = useState<any[]>([])
  const { status: sseStatus } = useSSE<any>('/api/sse/alerts', (evt) => {
    setEvents((prev) => [evt, ...prev].slice(0, 8))
  })
  const points = useMemo(() => (map?.countries || []).slice(0, 30).map((c: any, i: number) => ({ t: String(i), risk: Number(c.avgCompositeRisk || c.avg_risk || c.incident_pressure || 0) })), [map])
  const globeData = useMemo(() => (map?.countries || []).map((c: any) => ({ iso2: (c.countryCode || c.country_code || 'XX').toUpperCase(), name: c.countryName || c.country_name || 'Unknown', risk: Number(c.avgCompositeRisk || c.avg_risk || 0), highRisk: Number(c.incidentCount || c.incident_count || 0), p95: Number(c.avgCompositeRisk || c.avg_risk || 0) })), [map])
  if (!data) return <AppShell title="Command Center" subtitle="Real-time risk workspace" filters={filters} setFilters={setFilters}><LoadingState /></AppShell>
  return <AppShell title="Command Center" subtitle="Real-time risk workspace" filters={filters} setFilters={setFilters}>
    <TrustStrip trustState={data?.trust?.state} modelUnavailable={Boolean(data?.trust?.fallback_mode)} dqWarning={data?.trust?.dq_status} />
    <div className="grid-3"><div className="card"><h3>Open incidents</h3><p>{data.openIncidents || data.open_incidents || 0}</p></div><div className="card"><h3>High risk</h3><p>{data.highRiskCount || data.high_risk_count || 0}</p></div><div className="card"><h3>Backlog delta</h3><p>{data.backlogDelta || 0}</p></div></div>
    <CandleRiskPanel points={points} />
    <RiskGlobe data={globeData} onSelect={(iso2) => setFilters((f) => ({ ...f, countryCode: iso2 }))} />
    <WorldRiskMap data={map.countries || []} selectedCountryCode={filters.countryCode} onSelectCountry={(c) => setFilters((f) => ({ ...f, countryCode: c }))} />
    <div className='card'>
      <h3>Activity feed</h3>
      <div className='muted'>SSE status: {sseStatus}</div>
      <div className='muted'>Latest event: {events[0]?.message || 'waiting'}</div>
      <ul>{events.map((e, i) => <li key={`${e.seq || 'evt'}-${i}`}>{e.message || JSON.stringify(e)}</li>)}</ul>
    </div>
  </AppShell>
}
