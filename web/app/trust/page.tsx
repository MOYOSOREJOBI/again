'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import AppShell from '../../components/AppShell'
import { getTrust, type GlobalFilters } from '../../lib/api'

export default function TrustPage(){
  const [filters,setFilters]=useState<GlobalFilters>({time_window:'24h'}); const [data,setData]=useState<any>(null)
  useEffect(()=>{ getTrust(filters).then(setData)},[JSON.stringify(filters)])
  return <AppShell title="Trust" subtitle="Model and data quality posture" filters={filters} setFilters={setFilters}><div className="grid-3"><div className="card"><h3>Trust summary</h3><p>{data?.trust_summary?.state || 'stable'}</p></div><div className="card"><h3>Model health</h3><pre>{JSON.stringify(data?.model_health||{},null,2)}</pre></div><div className="card"><h3>DQ metrics</h3><pre>{JSON.stringify(data?.dq_counts||{},null,2)}</pre></div></div><div className="card"><h3>Trends</h3><div className="timeline-lane">missingness {JSON.stringify(data?.trends?.missingness||[])}</div><div className="timeline-lane">duplicates {JSON.stringify(data?.trends?.duplicates||[])}</div><div className="timeline-lane">late {JSON.stringify(data?.trends?.late_events||[])}</div></div><div className="card"><Link href="/queue?trust_state=degraded">View low-trust incidents</Link> · <Link href="/command-center">Back to command center</Link></div></AppShell>
}
