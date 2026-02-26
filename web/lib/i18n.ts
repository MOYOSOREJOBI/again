export type Locale = 'en' | 'fr' | 'es' | 'pt'

type Dict = Record<string, string>

const dictionaries: Record<Locale, Dict> = {
  en: {
    commandCenter: 'Command Center',
    queue: 'Queue',
    cases: 'Cases',
    trust: 'Trust',
    executive: 'Executive',
    governance: 'Governance',
    about: 'About',
    glossary: 'Glossary',
    operatorWorkbench: 'Operator Workbench',
  },
  fr: {
    commandCenter: 'Centre de commande',
    queue: 'File priorisée',
    cases: 'Dossiers',
    trust: 'Confiance',
    executive: 'Direction',
    governance: 'Gouvernance',
    about: 'À propos',
    glossary: 'Glossaire',
    operatorWorkbench: 'Poste opérateur',
  },
  es: {
    commandCenter: 'Centro de mando',
    queue: 'Cola',
    cases: 'Casos',
    trust: 'Confianza',
    executive: 'Ejecutivo',
    governance: 'Gobernanza',
    about: 'Acerca de',
    glossary: 'Glosario',
    operatorWorkbench: 'Mesa operativa',
  },
  pt: {
    commandCenter: 'Centro de comando',
    queue: 'Fila',
    cases: 'Casos',
    trust: 'Confiança',
    executive: 'Executivo',
    governance: 'Governança',
    about: 'Sobre',
    glossary: 'Glossário',
    operatorWorkbench: 'Bancada operacional',
  },
}

export function localeFromStorage(): Locale {
  if (typeof window === 'undefined') return 'en'
  const raw = (localStorage.getItem('sentinel_locale') || 'en').toLowerCase()
  if (raw === 'fr' || raw === 'es' || raw === 'pt') return raw
  return 'en'
}

export function setLocaleStorage(locale: Locale) {
  if (typeof window === 'undefined') return
  localStorage.setItem('sentinel_locale', locale)
}

export function t(locale: Locale, key: string, fallback: string): string {
  return dictionaries[locale]?.[key] || fallback
}

