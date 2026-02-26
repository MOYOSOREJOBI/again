'use client'
import { useEffect, useState } from 'react'
import AppShell from '../../components/AppShell'
import { getQueue, type GlobalFilters } from '../../lib/api'

export default function QueuePage(){
  const [filters,setFilters]=useState<GlobalFilters>({time_window:'24h'}); const [rows,setRows]=useState<any[]>([])
  useEffect(()=>{ let dead=false; const run=()=>getQueue(filters).then((r:any)=>!dead&&setRows(r)); run(); const t=setInterval(run,15000); return ()=>{dead=true; clearInterval(t)} },[JSON.stringify(filters)])
  return <AppShell title="Queue" subtitle="Ranked incident queue" filters={filters} setFilters={setFilters}><div className="card">{!rows.length?<div className="empty-state">Queue is clear for current filters.</div>:rows.map((r)=><div key={r.id} className="queue-row"><div><strong>{r.symbol}</strong> <span className="severity-chip">{r.severity}</span> <span className="badge">{r.trust_label}</span></div><div className="muted">{r.rank_reason}</div><div>{r.recommended_action}</div></div>)}</div></AppShell>
}
