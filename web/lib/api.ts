export const API_QUERY = process.env.NEXT_PUBLIC_API_QUERY || 'http://localhost:8085'
export const API_ALERTS = process.env.NEXT_PUBLIC_API_ALERTS || 'http://localhost:8083'
export const API_GOV = process.env.NEXT_PUBLIC_API_GOV || 'http://localhost:8084'
export const API_GATEWAY = process.env.NEXT_PUBLIC_API_GATEWAY || 'http://localhost:8080'

export type Role = 'viewer' | 'analyst' | 'admin'

export type GlobalFilters = {
  time_window?: 'now' | '1h' | '24h' | '7d'
  country?: string
  region?: string
  sector?: string
  industry?: string
  asset_class?: string
  venue?: string
  severity?: string
  trust_state?: string
}

export type QueueRow = {
  id: number
  symbol: string
  status: string
  severity_band: string
  priority_score: number
  recommended_action?: string
  rank_reason?: string
  country?: string
  region?: string
  sector?: string
  industry?: string
  venue?: string
  owner?: string
  composite_risk?: number
  escalation_probability?: number
  trust_state?: string
  age_minutes?: number
}

export type WorldMapRow = {
  country: string
  region: string
  sector: string
  industry?: string
  incident_count: number
  trust_state?: string
  pressure?: number
}

const defaultCommandCenter = {
  open_incidents: 0,
  high_risk_incidents: 0,
  critical_incidents: 0,
  changed_last_24h: 0,
  by_severity: { low: 0, medium: 0, high: 0, critical: 0 },
  top_incidents: [] as QueueRow[],
  recent_transitions: [] as Array<{ id: number; symbol: string; status: string; updated_at: string }>,
}

const defaultTrust = {
  model_unavailable_count: 0,
  degraded: [] as Array<{ symbol: string; alert_count: number }>,
  freshness_seconds: 0,
  missingness_rate: 0,
  duplicate_rate: 0,
  out_of_order_rate: 0,
  drift_indicator: 'stable',
  circuit_breaker_state: 'closed',
}

function readCSRFCookie(): string {
  if (typeof document === 'undefined') return ''
  const m = document.cookie.split(';').map((x) => x.trim()).find((x) => x.startsWith('sentinel_csrf='))
  return m ? decodeURIComponent(m.slice('sentinel_csrf='.length)) : ''
}

function safeArray<T>(value: unknown): T[] {
  return Array.isArray(value) ? (value as T[]) : []
}

function safeObject<T extends object>(value: unknown, fallback: T): T {
  return value && typeof value === 'object' ? ({ ...fallback, ...(value as object) } as T) : fallback
}

export function filtersToQuery(filters: GlobalFilters = {}): string {
  const params = new URLSearchParams()
  for (const [k, v] of Object.entries(filters)) {
    if (v) params.set(k, v)
  }
  const query = params.toString()
  return query ? `?${query}` : ''
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
  me: async () => safeObject(await apiFetch(API_GATEWAY, '/me').catch(() => ({})), { email: '', role: 'viewer' as Role }),
  commandCenter: async (filters: GlobalFilters = {}) => safeObject(await apiFetch(API_QUERY, `/command-center${filtersToQuery(filters)}`).catch(() => ({})), defaultCommandCenter),
  queue: async (filters: GlobalFilters = {}) => safeArray<QueueRow>(await apiFetch(API_QUERY, `/queue${filtersToQuery(filters)}`).catch(() => [])),
  incident: async (id: string | number) => safeObject(await apiFetch(API_QUERY, `/incident/${id}`).catch(() => ({})), { id, top_drivers: [], related_incidents: [] }),
  trust: async (filters: GlobalFilters = {}) => safeObject(await apiFetch(API_QUERY, `/trust${filtersToQuery(filters)}`).catch(() => ({})), defaultTrust),
  worldMap: async (filters: GlobalFilters = {}) => safeArray<WorldMapRow>(await apiFetch(API_QUERY, `/world-map${filtersToQuery(filters)}`).catch(() => [])),
  case: async (id: string | number) => safeObject(await apiFetch(API_QUERY, `/case/${id}`).catch(() => ({})), { id }),
  caseDetail: async (id: string | number) => safeObject(await apiFetch(API_ALERTS, `/cases/${id}`).catch(() => ({})), { id, notes: [], evidence: [], audit_events: [] }),
  createCase: (incidentId: number, reason: string) => apiFetch(API_ALERTS, '/cases', { method: 'POST', body: JSON.stringify({ incident_id: incidentId, reason }) }),
  caseStatus: (id: string | number, status: string) => apiFetch(API_ALERTS, `/cases/${id}`, { method: 'PATCH', body: JSON.stringify({ status }) }),
  caseNote: (id: string | number, note: string) => apiFetch(API_ALERTS, `/cases/${id}/notes`, { method: 'POST', body: JSON.stringify({ note }) }),
  caseEvidence: (id: string | number, evidence_type: string, reference_id: string, metadata: any = {}) => apiFetch(API_ALERTS, `/cases/${id}/evidence`, { method: 'POST', body: JSON.stringify({ evidence_type, reference_id, metadata }) }),
  replay: async (job: string) => safeObject(await apiFetch(API_QUERY, `/replay/${job}`).catch(() => ({})), { id: job, status: 'unknown' }),
  governanceSummary: async () => safeObject(await apiFetch(API_QUERY, '/governance/summary').catch(() => ({})), { deployed_models: [], replay_jobs: [], threshold_changes: [] }),
  executiveSummary: async (filters: GlobalFilters = {}) => safeObject(await apiFetch(API_QUERY, `/executive-summary${filtersToQuery(filters)}`).catch(() => ({})), { top_risks: [], hot_regions: [], trust_summary: {} }),
  models: async () => safeArray(await apiFetch(API_GOV, '/models').catch(() => [])),
}

export type AlertDTO = { id: number; symbol: string; status: string; ts: string; created_at?: string }
export function normalizeAlertCreatedAt(alert: AlertDTO): string { return alert.created_at || alert.ts }
