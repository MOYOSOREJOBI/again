export const API_QUERY = process.env.NEXT_PUBLIC_API_QUERY || 'http://localhost:8085'
export const API_ALERTS = process.env.NEXT_PUBLIC_API_ALERTS || 'http://localhost:8083'
export const API_GOV = process.env.NEXT_PUBLIC_API_GOV || 'http://localhost:8084'
export const API_GATEWAY = process.env.NEXT_PUBLIC_API_GATEWAY || 'http://localhost:8080'

export type QueueRow = { id: number; symbol: string; status: string; severity_band: string; priority_score: number; recommended_action?: string; rank_reason?: string; country?: string; region?: string; sector?: string; industry?: string }

function readCSRFCookie(): string {
  if (typeof document === 'undefined') return ''
  const m = document.cookie.split(';').map((x) => x.trim()).find((x) => x.startsWith('sentinel_csrf='))
  return m ? decodeURIComponent(m.slice('sentinel_csrf='.length)) : ''
}

export async function apiFetch(base: string, path: string, init: RequestInit = {}) {
  const headers = new Headers(init.headers || {})
  if (!headers.has('Content-Type') && init.body) headers.set('Content-Type', 'application/json')
  if ((init.method || 'GET').toUpperCase() !== 'GET') {
    const csrf = readCSRFCookie()
    if (csrf) headers.set('X-CSRF-Token', csrf)
  }
  const res = await fetch(`${base}${path}`, { credentials: 'include', cache: 'no-store', ...init, headers })
  if (!res.ok) {
    const txt = await res.text().catch(() => '')
    throw new Error(`${res.status}: ${txt || 'request failed'}`)
  }
  const ct = res.headers.get('content-type') || ''
  if (ct.includes('application/json')) return res.json()
  return null
}

export const api = {
  login: (Email: string, Password: string) => apiFetch(API_GATEWAY, '/auth/login', { method: 'POST', body: JSON.stringify({ Email, Password }) }),
  me: () => apiFetch(API_GATEWAY, '/me'),
  commandCenter: () => apiFetch(API_QUERY, '/command-center'),
  queue: (params = '') => apiFetch(API_QUERY, `/queue${params}`),
  incident: (id: string | number) => apiFetch(API_QUERY, `/incident/${id}`),
  trust: () => apiFetch(API_QUERY, '/trust'),
  worldMap: (params = '') => apiFetch(API_QUERY, `/world-map${params}`),
  case: (id: string | number) => apiFetch(API_QUERY, `/case/${id}`),
  caseDetail: (id: string | number) => apiFetch(API_ALERTS, `/cases/${id}`),
  createCase: (incidentId: number, reason: string) => apiFetch(API_ALERTS, '/cases', { method: 'POST', body: JSON.stringify({ incident_id: incidentId, reason }) }),
  caseStatus: (id: string | number, status: string) => apiFetch(API_ALERTS, `/cases/${id}`, { method: 'PATCH', body: JSON.stringify({ status }) }),
  caseNote: (id: string | number, note: string) => apiFetch(API_ALERTS, `/cases/${id}/notes`, { method: 'POST', body: JSON.stringify({ note }) }),
  caseEvidence: (id: string | number, evidence_type: string, reference_id: string, metadata: any = {}) => apiFetch(API_ALERTS, `/cases/${id}/evidence`, { method: 'POST', body: JSON.stringify({ evidence_type, reference_id, metadata }) }),
  replay: (job: string) => apiFetch(API_QUERY, `/replay/${job}`),
  models: () => apiFetch(API_GOV, '/models'),
}

export type AlertDTO = { id: number; symbol: string; status: string; ts: string; created_at?: string }
export function normalizeAlertCreatedAt(alert: AlertDTO): string { return alert.created_at || alert.ts }
