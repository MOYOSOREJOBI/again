export type AlertDTO = { id: number; symbol: string; status: string; ts: string; created_at?: string }

export function normalizeAlertCreatedAt(alert: AlertDTO): string {
  return alert.created_at || alert.ts
}
