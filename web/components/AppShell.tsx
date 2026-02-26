'use client'

import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import { useEffect, useMemo, useState } from 'react'
import { api, GlobalFilters, Role } from '../lib/api'

const NAV_ITEMS = [
  { href: '/command-center', label: 'Command Center' },
  { href: '/queue', label: 'Queue' },
  { href: '/trust', label: 'Trust' },
  { href: '/executive', label: 'Executive' },
  { href: '/governance', label: 'Governance', minRole: 'admin' as Role },
  { href: '/about', label: 'About' },
  { href: '/glossary', label: 'Glossary' },
]

export function useGlobalFilters() {
  const [filters, setFilters] = useState<GlobalFilters>({ time_window: '24h' })
  useEffect(() => {
    const raw = localStorage.getItem('sentinel_global_filters')
    if (raw) setFilters({ time_window: '24h', ...JSON.parse(raw) })
  }, [])
  const apply = (next: Partial<GlobalFilters>) => {
    setFilters((prev) => {
      const value = { ...prev, ...next }
      localStorage.setItem('sentinel_global_filters', JSON.stringify(value))
      return value
    })
  }
  return { filters, apply }
}

function roleAllowed(userRole: Role, minRole?: Role): boolean {
  if (!minRole) return true
  if (userRole === 'admin') return true
  if (userRole === 'analyst') return minRole !== 'admin'
  return minRole === 'viewer'
}

export default function AppShell({ title, subtitle, children, filters, setFilters }: { title: string; subtitle?: string; children: React.ReactNode; filters?: GlobalFilters; setFilters?: (f: Partial<GlobalFilters>) => void }) {
  const pathname = usePathname()
  const router = useRouter()
  const [me, setMe] = useState<{ email: string; role: Role }>({ email: '', role: 'viewer' })

  useEffect(() => {
    api.me().then((m) => setMe({ email: m.email || '', role: m.role || 'viewer' })).catch(() => router.replace('/'))
  }, [router])

  const roleLabel = useMemo(() => me.role.charAt(0).toUpperCase() + me.role.slice(1), [me.role])

  return (
    <main className="shell-root">
      <aside className="left-nav">
        <div className="brand"><span className="brand-dot" />Sentinel</div>
        <p className="muted">Operator Workbench</p>
        <nav className="left-nav-list">
          {NAV_ITEMS.filter((i) => roleAllowed(me.role, i.minRole)).map((item) => (
            <Link key={item.href} href={item.href} className={`left-nav-link ${pathname === item.href ? 'active' : ''}`}>{item.label}</Link>
          ))}
        </nav>
      </aside>
      <section className="workspace">
        <header className="topbar sticky-top">
          <div>
            <h1>{title}</h1>
            {subtitle ? <p className="muted">{subtitle}</p> : null}
          </div>
          <div className="user-meta">
            <span className="pill role">{roleLabel}</span>
            <span className="pill">{me.email || 'signed in'}</span>
          </div>
        </header>
        {setFilters ? (
          <div className="filter-strip sticky-filters">
            <select value={filters?.time_window || '24h'} onChange={(e) => setFilters({ time_window: e.target.value as any })}>
              <option value="now">now</option><option value="1h">1h</option><option value="24h">24h</option><option value="7d">7d</option>
            </select>
            <input placeholder="country" value={filters?.country || ''} onChange={(e) => setFilters({ country: e.target.value })} />
            <input placeholder="region" value={filters?.region || ''} onChange={(e) => setFilters({ region: e.target.value })} />
            <input placeholder="sector" value={filters?.sector || ''} onChange={(e) => setFilters({ sector: e.target.value })} />
            <input placeholder="venue" value={filters?.venue || ''} onChange={(e) => setFilters({ venue: e.target.value })} />
          </div>
        ) : null}
        <div className="page-content">{children}</div>
      </section>
    </main>
  )
}
