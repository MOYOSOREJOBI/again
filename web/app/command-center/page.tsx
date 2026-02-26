'use client'
import { useEffect, useState } from 'react'
import AppShell from '../../components/AppShell'
import TrustStrip from '../../components/TrustStrip'
import WorldRiskMap from '../../components/WorldRiskMap'
import { getCommandCenter, getWorldMap, type GlobalFilters } from '../../lib/api'

export default function CommandCenterPage(){
  const [filters,setFilters]=useState<GlobalFilters>({time_window:'24h'})
  const [data,setData]=useState<any>(null); const [map,setMap]=useState<any>({rows:[]})
  useEffect(()=>{ let dead=false; const run=()=>Promise.all([getCommandCenter(filters),getWorldMap(filters)]).then(([cc,wm])=>{ if(!dead){setData(cc); setMap(wm)}}); run(); const t=setInterval(run,15000); return ()=>{dead=true; clearInterval(t)} },[JSON.stringify(filters)])
  return <AppShell title="Command Center" subtitle="Real-time risk workspace" filters={filters} setFilters={setFilters}><TrustStrip trustState={data?.trust?.state} modelUnavailable={Boolean(data?.trust?.fallback_mode)} dqWarning={data?.trust?.dq_status}/><div className="grid-3">{['open_incidents','high_risk_count'].map((k)=><div key={k} className="card"><h3>{k}</h3><p>{data?.[k] ?? 0}</p></div>)}</div><div className="card"><h3>Incident pressure</h3><div className="sparkline">{JSON.stringify(data?.incident_pressure_series || [])}</div></div><div className="card"><h3>Severity distribution</h3><div className="metric-bars">{Object.entries(data?.severity_distribution||{}).map(([k,v]:any)=><div key={k} className="kv"><span>{k}</span><strong>{String(v)}</strong></div>)}</div></div><div className="card"><h3>Top incidents</h3>{(data?.top_incidents||[]).length? data.top_incidents.map((r:any)=><div className="kv" key={r.id}><span>{r.symbol}</span><span>{r.composite_risk?.toFixed?.(2)}</span></div>):<div className="empty-state">No open critical incidents right now.</div>}</div><WorldRiskMap rows={map.rows||[]} selectedRegion={filters.region} onSelectRegion={(region)=>setFilters((f)=>({...f,region}))}/></AppShell>
}
