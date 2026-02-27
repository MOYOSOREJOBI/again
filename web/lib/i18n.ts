export type Locale = 'en' | 'fr' | 'es' | 'pt' | 'it' | 'zh-Hans' | 'ja' | 'ar' | 'de' | 'sw' | 'yo' | 'ig' | 'ha' | 'tr' | 'ru' | 'nl' | 'hi'

type Dict = Record<string, string>

const base: Dict = {
  commandCenter: 'Command Center',
  queue: 'Queue',
  cases: 'Cases',
  trust: 'Trust',
  executive: 'Executive',
  governance: 'Governance',
  about: 'About',
  glossary: 'Glossary',
  operatorWorkbench: 'Operator Workbench',
}

const dictionaries: Record<Locale, Dict> = {
  en: base,
  fr: { ...base, commandCenter: 'Centre de commande', queue: 'File priorisée', cases: 'Dossiers', trust: 'Confiance', governance: 'Gouvernance' },
  es: { ...base, commandCenter: 'Centro de mando', queue: 'Cola', cases: 'Casos', trust: 'Confianza', governance: 'Gobernanza' },
  pt: { ...base, commandCenter: 'Centro de comando', queue: 'Fila', cases: 'Casos', trust: 'Confiança', governance: 'Governança', about: 'Sobre' },
  it: { ...base, queue: 'Coda' },
  'zh-Hans': { ...base, commandCenter: '指挥中心', queue: '队列', cases: '案件', trust: '信任' },
  ja: { ...base, commandCenter: 'コマンドセンター', queue: 'キュー', cases: 'ケース', trust: '信頼' },
  ar: { ...base, commandCenter: 'مركز القيادة', queue: 'الطابور', cases: 'الحالات', trust: 'الثقة' },
  de: { ...base, queue: 'Warteschlange' },
  sw: { ...base },
  yo: { ...base },
  ig: { ...base },
  ha: { ...base },
  tr: { ...base, queue: 'Kuyruk' },
  ru: { ...base, queue: 'Очередь' },
  nl: { ...base, queue: 'Wachtrij' },
  hi: { ...base, queue: 'कतार' },
}

const locales: Locale[] = ['en','fr','es','pt','it','zh-Hans','ja','ar','de','sw','yo','ig','ha','tr','ru','nl','hi']

export function applyLocale(locale: string) {
  const safe = (locales.includes(locale as Locale) ? locale : 'en') as Locale
  document.documentElement.lang = safe
  document.documentElement.dir = safe === 'ar' ? 'rtl' : 'ltr'
  localStorage.setItem('sentinel_locale', safe)
  document.cookie = `locale=${safe}; Path=/; Max-Age=31536000; SameSite=Lax`
}

export function localeFromStorage(): Locale {
  if (typeof window === 'undefined') return 'en'
  const raw = (localStorage.getItem('sentinel_locale') || 'en') as Locale
  return locales.includes(raw) ? raw : 'en'
}

export function setLocaleStorage(locale: Locale) {
  if (typeof window === 'undefined') return
  applyLocale(locale)
}

export function t(locale: Locale, key: string, fallback: string): string {
  return dictionaries[locale]?.[key] || fallback
}

export function allLocales(): Locale[] { return locales }
