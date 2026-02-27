'use client'
import { useEffect, useState } from 'react'
import AppShell from '../../components/AppShell'
import TrustStrip from '../../components/TrustStrip'
import WorldRiskMap from '../../components/WorldRiskMap'
import LoadingState from '../../components/LoadingState'
import { api, type GlobalFilters } from '../../lib/api'
import { subscribeCommandCenterPatches } from '../../lib/stream'

export default function CommandCenterPage() {
  const [filters, setFilters] = useState<GlobalFilters>({ window: '24h' })
  const [data, setData] = useState<any>(null)
  const [map, setMap] = useState<any>({ countries: [] })
  useEffect(() => { Promise.all([api.commandCenter(filters), api.worldMap(filters)]).then(([cc, wm]) => { setData(cc); setMap(wm) }) }, [JSON.stringify(filters)])
  useEffect(() => subscribeCommandCenterPatches((p) => setData((d: any) => ({ ...(d || {}), ...p })), () => {}), [])
  if (!data) return <AppShell title="Command Center" subtitle="Real-time risk workspace" filters={filters} setFilters={setFilters}><LoadingState /></AppShell>
  return <AppShell title="Command Center" subtitle="Real-time risk workspace" filters={filters} setFilters={setFilters}>
    <TrustStrip trustState={data?.trust?.state} modelUnavailable={Boolean(data?.trust?.fallback_mode)} dqWarning={data?.trust?.dq_status} />
    <div className="grid-3"><div className="card"><h3>Open incidents</h3><p>{data.openIncidents || data.open_incidents || 0}</p></div><div className="card"><h3>High risk</h3><p>{data.highRiskCount || data.high_risk_count || 0}</p></div><div className="card"><h3>Backlog delta</h3><p>{data.backlogDelta || 0}</p></div></div>
    <div className="muted">incident_pressure_series</div>
    <WorldRiskMap data={map.countries || []} selectedCountryCode={filters.countryCode} onSelectCountry={(c) => setFilters((f) => ({ ...f, countryCode: c }))} />
  </AppShell>
}
