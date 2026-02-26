'use client'
import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import AppShell from '../../../components/AppShell'
import { getReplay } from '../../../lib/api'

export default function ReplayPage(){
  const params=useParams<{job:string}>(); const [job,setJob]=useState<any>(null)
  useEffect(()=>{ if(params?.job) getReplay(params.job).then(setJob)},[params?.job])
  return <AppShell title={`Replay ${params?.job||''}`} subtitle="Metadata-first replay workspace">{!job?<div className="empty-state">Replay metadata unavailable.</div>:<><div className="grid-3"><div className="card"><h3>Status</h3><p>{job.status}</p></div><div className="card"><h3>Window</h3><p>{job.time_window}</p></div><div className="card"><h3>Mode</h3><p>{job.mode}</p></div></div><div className="card"><h3>Lane summary</h3><pre>{JSON.stringify(job.lane_series||{},null,2)}</pre><p className="muted">metadata-first mode: full recompute is not yet enabled.</p></div><div className="card"><h3>Deterministic timeline</h3>{(job.timeline_entries||[]).length ? <div>{job.timeline_entries.map((e:any,idx:number)=><div key={idx} className="timeline-lane"><strong>{e.kind}</strong> · {e.symbol} · {e.status} · {e.severity}</div>)}</div> : <div className="empty-state">No timeline entries available for this replay window.</div>}</div></>}</AppShell>
}
