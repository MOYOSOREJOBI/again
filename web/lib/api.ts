import { filtersToQuery, type GlobalFilters } from './filters'

export const API_QUERY = process.env.NEXT_PUBLIC_API_QUERY || 'http://localhost:8085'
export const API_ALERTS = process.env.NEXT_PUBLIC_API_ALERTS || 'http://localhost:8083'
export const API_GATEWAY = process.env.NEXT_PUBLIC_API_GATEWAY || 'http://localhost:8080'
export type Role = 'viewer' | 'analyst' | 'admin'

function readCSRFCookie() { if (typeof document === 'undefined') return ''; const m = document.cookie.split(';').map(v => v.trim()).find(v => v.startsWith('sentinel_csrf=')); return m ? decodeURIComponent(m.split('=')[1]) : '' }
async function apiFetch(base: string, path: string, init: RequestInit = {}) {
  const headers = new Headers(init.headers || {})
  if ((init.method || 'GET').toUpperCase() !== 'GET') { const csrf = readCSRFCookie(); if (csrf) headers.set('X-CSRF-Token', csrf); if (!headers.has('Content-Type')) headers.set('Content-Type', 'application/json') }
  const res = await fetch(`${base}${path}`, { credentials: 'include', cache: 'no-store', ...init, headers }); if (!res.ok) throw new Error(`${res.status}`)
  const ct = res.headers.get('content-type') || ''; return ct.includes('json') ? res.json() : {}
}

export const api = {
  me: () => apiFetch(API_GATEWAY, '/me').catch(() => ({ role: 'viewer' })),
  login: (Email: string, Password: string) => apiFetch(API_GATEWAY, '/auth/login', { method: 'POST', body: JSON.stringify({ Email, Password }) }),
  commandCenter: (f: GlobalFilters = {}) => apiFetch(API_QUERY, `/command-center${filtersToQuery(f)}`).catch(() => ({ openIncidents: 0, highRiskCount: 0 })),
  queue: (f: GlobalFilters = {}) => apiFetch(API_QUERY, `/queue${filtersToQuery(f)}`).catch(() => ([])),
  incident: (id: string | number) => apiFetch(API_QUERY, `/incident/${id}`).catch(() => ({ id })),
  trust: (f: GlobalFilters = {}) => apiFetch(API_QUERY, `/trust${filtersToQuery(f)}`).catch(() => ({})),
  replay: (job: string) => apiFetch(API_QUERY, `/replay/${job}`).catch(() => ({ id: job, status: 'unknown' })),
  worldMap: (f: GlobalFilters = {}) => apiFetch(API_QUERY, `/world-map${filtersToQuery(f)}`).catch(() => ({ countries: [] })),
  governanceSummary: () => apiFetch(API_QUERY, '/governance/summary').catch(() => ({ modelLineage: [], replayJobs: [] })),
  executiveSummary: (f: GlobalFilters = {}) => apiFetch(API_QUERY, `/executive-summary${filtersToQuery(f)}`).catch(() => ({})),
  case: (id: string | number) => apiFetch(API_QUERY, `/case/${id}`).catch(() => ({ id })),
  cases: () => apiFetch(API_QUERY, '/cases').catch(() => ([])),
  caseDetail: (id: string | number) => apiFetch(API_ALERTS, `/cases/${id}`).catch(() => ({ id, notes: [], evidence: [], actions: [] })),
  caseStatus: (id: string | number, status: string) => apiFetch(API_ALERTS, `/cases/${id}`, { method: 'PATCH', body: JSON.stringify({ status }) }),
  caseNote: (id: string | number, note: string) => apiFetch(API_ALERTS, `/cases/${id}/notes`, { method: 'POST', body: JSON.stringify({ note }) }),
  caseEvidence: (id: string | number, evidence_type: string, reference_id: string, metadata: any = {}) => apiFetch(API_ALERTS, `/cases/${id}/evidence`, { method: 'POST', body: JSON.stringify({ evidence_type, reference_id, metadata }) }),
  caseDisposition: (id: string | number, status: string, reason: string) => apiFetch(API_ALERTS, `/cases/${id}/disposition`, { method: 'POST', body: JSON.stringify({ status, reason }) }),
  promoteIncidentCase: (incidentId: string | number, reason: string) => apiFetch(API_ALERTS, `/incidents/${incidentId}/promote-case`, { method: 'POST', body: JSON.stringify({ reason }) }),
}

export const getCommandCenter = api.commandCenter
export const getQueue = api.queue
export const getIncident = api.incident
export const getTrust = api.trust
export const getReplay = api.replay
export const getWorldMap = api.worldMap
export function normalizeAlertCreatedAt(v: any): string { if (!v) return ''; if (typeof v === 'object') return String(v.created_at || v.ts || ''); return String(v) }
export type { GlobalFilters } from './filters'
