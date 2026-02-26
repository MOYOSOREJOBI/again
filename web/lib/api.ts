export const API_QUERY = process.env.NEXT_PUBLIC_API_QUERY || 'http://localhost:8085'

export async function getJSON(path: string) {
  const res = await fetch(`${API_QUERY}${path}`, { credentials: 'include', cache: 'no-store' })
  if (!res.ok) throw new Error(`request failed: ${res.status}`)
  return res.json()
}

export type AlertDTO = { id: number; symbol: string; status: string; ts: string; created_at?: string }
export function normalizeAlertCreatedAt(alert: AlertDTO): string { return alert.created_at || alert.ts }
