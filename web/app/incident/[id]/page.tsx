'use client'
import Link from 'next/link'
import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import AppShell from '../../../components/AppShell'
import { api, getIncident } from '../../../lib/api'

export default function IncidentPage(){
  const params=useParams<{id:string}>(); const [data,setData]=useState<any>(null); const [me,setMe]=useState<any>({role:'viewer'}); const [msg,setMsg]=useState('')
  useEffect(()=>{ if(params?.id) getIncident(params.id).then(setData); api.me().then(setMe)},[params?.id])
  const h=data?.score_header||{}
  const canMutate = me?.role === 'analyst' || me?.role === 'admin'
  return <AppShell title={`Incident ${params?.id||''}`} subtitle="Score evidence and explanation"><div className="grid-3">{['anomaly','escalation','confidence','priority','composite_risk','safety_level'].map((k)=><div className="card" key={k}><h3>{k}</h3><p>{String(h[k]??'-')}</p></div>)}</div><section className="card"><h3>Explanation</h3><p>{data?.explanation_text}</p><ul>{(data?.top_drivers||[]).map((d:any,i:number)=><li key={i}>{typeof d==='string'?d:d.driver}</li>)}</ul><p className="muted">{(data?.caveats||[]).join('; ')}</p></section><section className="card"><h3>Model lineage</h3><p>Version: {data?.model_version || 'unknown'} · Fallback: {String(data?.fallback_mode ?? false)}</p><p className="muted">Feature snapshot: {data?.feature_snapshot_hash || 'n/a'}</p></section><section className="card"><h3>Case promotion</h3>{canMutate ? <div className="form-grid"><button onClick={async ()=>{ try { const out = await api.promoteIncidentCase(params?.id || '', 'Analyst promoted from incident detail'); setMsg(`Case ${out.case_id} ready`)} catch(e:any){ setMsg(`Promotion blocked: ${e.message}`)} }}>Promote to Case</button>{msg && <p className="muted">{msg}</p>}</div> : <p className="muted">Viewer mode: case promotion disabled.</p>}<Link href="/case" className="pill">Open Case Workspace</Link></section></AppShell>
}
