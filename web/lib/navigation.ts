export type Role = 'admin' | 'analyst' | 'viewer'
export type Tab = 'overview' | 'alerts' | 'scores' | 'governance'

export function navItemsForRole(role: Role): Tab[] {
  if (role === 'admin') return ['overview', 'alerts', 'scores', 'governance']
  if (role === 'analyst') return ['overview', 'alerts', 'scores']
  return ['overview', 'scores']
}
