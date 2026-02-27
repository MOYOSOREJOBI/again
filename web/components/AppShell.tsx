'use client'

import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import { useEffect, useMemo, useState } from 'react'
import { api, GlobalFilters, Role } from '../lib/api'
import { Locale, allLocales, applyLocale, localeFromStorage, setLocaleStorage, t } from '../lib/i18n'

const NAV_ITEMS = [
  { href: '/command-center', label: 'Command Center', key: 'commandCenter' },
  { href: '/queue', label: 'Queue', key: 'queue' },
  { href: '/case', label: 'Cases', key: 'cases' },
  { href: '/trust', label: 'Trust', key: 'trust' },
  { href: '/executive', label: 'Executive', key: 'executive' },
  { href: '/governance', label: 'Governance', key: 'governance', minRole: 'admin' as Role },
  { href: '/about', label: 'About', key: 'about' },
  { href: '/glossary', label: 'Glossary', key: 'glossary' },
]

export function useGlobalFilters() {
  const [filters, setFilters] = useState<GlobalFilters>({ window: '24h' })
  useEffect(() => {
    const raw = localStorage.getItem('sentinel_global_filters')
    if (raw) setFilters({ window: '24h', ...JSON.parse(raw) })
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
  const [locale, setLocale] = useState<Locale>('en')

  useEffect(() => {
    const l = localeFromStorage();
    setLocale(l);
    applyLocale(l)
    api.me().then((m) => setMe({ email: m.email || '', role: m.role || 'viewer' })).catch(() => router.replace('/'))
  }, [router])

  const roleLabel = useMemo(() => me.role.charAt(0).toUpperCase() + me.role.slice(1), [me.role])

  return (
    <main className="shell-root">
      <aside className="left-nav">
        <div className="brand"><span className="brand-dot" />Sentinel</div>
        <p className="muted">{t(locale, 'operatorWorkbench', 'Operator Workbench')}</p>
        <nav className="left-nav-list">
          {NAV_ITEMS.filter((i) => roleAllowed(me.role, i.minRole)).map((item) => (
            <Link key={item.href} href={item.href} className={`left-nav-link ${pathname === item.href ? 'active' : ''}`}>{t(locale, item.key, item.label)}</Link>
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
            <input aria-label="Locale search" placeholder="lang" onChange={(e)=>{ const q=e.target.value.toLowerCase(); const first=allLocales().find((l)=>l.toLowerCase().includes(q)); if(first){ setLocale(first); setLocaleStorage(first);} }} />
            <select aria-label="Locale" value={locale} onChange={(e)=>{ const next=e.target.value as Locale; setLocale(next); setLocaleStorage(next); }}>
              {allLocales().map((l)=><option key={l} value={l}>{l}</option>)}
            </select>
            <span className="pill role">{roleLabel}</span>
            <span className="pill">{me.email || 'signed in'}</span>
          </div>
        </header>
        {setFilters ? (
          <div className="filter-strip sticky-filters">
            <select value={filters?.window || '24h'} onChange={(e) => setFilters({ window: e.target.value as any })}>
              <option value="1h">1h</option><option value="24h">24h</option><option value="7d">7d</option>
            </select>
            <input placeholder="country" value={filters?.countryCode || ''} onChange={(e) => setFilters({ countryCode: e.target.value })} />
            <input placeholder="region" value={filters?.region || ''} onChange={(e) => setFilters({ region: e.target.value })} />
            <input placeholder="sector" value={filters?.sector || ''} onChange={(e) => setFilters({ sector: e.target.value })} />
            <input placeholder="industry" value={filters?.industry || ''} onChange={(e) => setFilters({ industry: e.target.value })} />
            <input placeholder="venue" value={filters?.venue || ''} onChange={(e) => setFilters({ venue: e.target.value })} />
          </div>
        ) : null}
        <div className="page-content">{children}</div>
      </section>
    </main>
  )
}
