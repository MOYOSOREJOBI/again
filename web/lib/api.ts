export const API_QUERY = process.env.NEXT_PUBLIC_API_QUERY || 'http://localhost:8085'
export const API_ALERTS = process.env.NEXT_PUBLIC_API_ALERTS || 'http://localhost:8083'
export const API_GATEWAY = process.env.NEXT_PUBLIC_API_GATEWAY || 'http://localhost:8080'

export type Role = 'viewer' | 'analyst' | 'admin'
export type GlobalFilters = { time_window?: 'now'|'1h'|'24h'|'7d'; country?: string; region?: string; sector?: string; industry?: string; venue?: string; asset_class?: string; severity?: string; trust_state?: string }

function readCSRFCookie(){ if(typeof document==='undefined') return ''; const m=document.cookie.split(';').map(v=>v.trim()).find(v=>v.startsWith('sentinel_csrf=')); return m?decodeURIComponent(m.split('=')[1]):'' }
export function filtersToQuery(filters: GlobalFilters = {}){ const p=new URLSearchParams(); Object.entries(filters).forEach(([k,v])=>{ if(v) p.set(k,String(v))}); const q=p.toString(); return q?`?${q}`:'' }
async function apiFetch(base:string,path:string,init:RequestInit={}){ const headers=new Headers(init.headers||{}); if((init.method||'GET').toUpperCase()!=='GET'){ const csrf=readCSRFCookie(); if(csrf) headers.set('X-CSRF-Token',csrf); if(!headers.has('Content-Type')) headers.set('Content-Type','application/json') } const res=await fetch(`${base}${path}`,{credentials:'include',cache:'no-store',...init,headers}); if(!res.ok){ throw new Error(`${res.status}`)} const ct=res.headers.get('content-type')||''; return ct.includes('json')?res.json():{} }
const norm=(v:any,d:any)=> (v&&typeof v==='object'?{...d,...v}:d)

export const getCommandCenter = async (f:GlobalFilters={}) => norm(await apiFetch(API_QUERY,`/command-center${filtersToQuery(f)}`).catch(()=>({})), {trust:{state:'stable'},open_incidents:0,high_risk_count:0,top_incidents:[],incident_pressure_series:[],severity_distribution:{}})
export const getQueue = async (f:GlobalFilters={}) => await apiFetch(API_QUERY,`/queue${filtersToQuery(f)}`).catch(()=>([]))
export const getIncident = async (id:string|number,f:GlobalFilters={}) => norm(await apiFetch(API_QUERY,`/incident/${id}${filtersToQuery(f)}`).catch(()=>({id})), {id,score_header:{},top_drivers:[],baseline_deltas:[],trends:{}})
export const getTrust = async (f:GlobalFilters={}) => norm(await apiFetch(API_QUERY,`/trust${filtersToQuery(f)}`).catch(()=>({})), {trust_summary:{state:'stable'}, trends:{missingness:[],duplicates:[],late_events:[]}})
export const getReplay = async (jobId:string) => norm(await apiFetch(API_QUERY,`/replay/${jobId}`).catch(()=>({id:jobId,status:'unknown',mode:'metadata_first'})), {id:jobId,status:'unknown',mode:'metadata_first',timeline_entries:[]})
export const getWorldMap = async (f:GlobalFilters={}) => norm(await apiFetch(API_QUERY,`/world-map${filtersToQuery(f)}`).catch(()=>({rows:[]})), {rows:[],filters:f})
export const getGovernanceSummary = async () => norm(await apiFetch(API_QUERY,'/governance/summary').catch(()=>({})), {model_lineage:[],replay_jobs:[]})
export const getExecutiveSummary = async (f:GlobalFilters={}) => norm(await apiFetch(API_QUERY,`/executive-summary${filtersToQuery(f)}`).catch(()=>({})), {top_risks:[],hot_regions:[]})

export const api = {
  me: async () => await apiFetch(API_GATEWAY, '/me').catch(() => ({email:'', role:'viewer'})),
  login: (Email:string, Password:string) => apiFetch(API_GATEWAY, '/auth/login', {method:'POST', body:JSON.stringify({Email,Password})}),
  commandCenter: getCommandCenter, queue: getQueue, incident: getIncident, trust: getTrust, replay: getReplay, worldMap: getWorldMap, governanceSummary: getGovernanceSummary, executiveSummary: getExecutiveSummary,
  case: async (id:string|number)=> apiFetch(API_QUERY,`/case/${id}`).catch(()=>({id})),
  cases: async ()=> apiFetch(API_QUERY,'/cases').catch(()=>([])),
  caseDetail: async (id:string|number)=> apiFetch(API_ALERTS,`/cases/${id}`).catch(()=>({id,notes:[],evidence:[],actions:[]})),
  createCase: (incidentId:number, reason:string)=> apiFetch(API_ALERTS,'/cases',{method:'POST', body:JSON.stringify({incident_id:incidentId, reason})}),
  promoteIncidentCase: (incidentId:string|number, reason:string)=> apiFetch(API_ALERTS,`/incidents/${incidentId}/promote-case`,{method:'POST', body:JSON.stringify({reason})}),
  caseStatus: (id:string|number, status:string)=> apiFetch(API_ALERTS,`/cases/${id}`,{method:'PATCH', body:JSON.stringify({status})}),
  caseNote: (id:string|number, note:string)=> apiFetch(API_ALERTS,`/cases/${id}/notes`,{method:'POST', body:JSON.stringify({note})}),
  caseEvidence: (id:string|number, evidence_type:string, reference_id:string, metadata:any={})=> apiFetch(API_ALERTS,`/cases/${id}/evidence`,{method:'POST', body:JSON.stringify({evidence_type, reference_id, metadata})}),
  caseDisposition: (id:string|number, status:string, reason:string)=> apiFetch(API_ALERTS,`/cases/${id}/disposition`,{method:'POST', body:JSON.stringify({status, reason})}),
}

export function normalizeAlertCreatedAt(v: any): string { if (!v) return ''; if (typeof v === 'object') return String(v.created_at || v.ts || ''); return String(v) }
