'use client'
import AppShell, { useGlobalFilters } from '../../components/AppShell'

const terms = [
  ['Incident', 'A clustered operational risk event requiring triage.'],
  ['Alert', 'A single scored signal that may link to an incident.'],
  ['Severity', 'Operator-facing risk band from stable to critical.'],
  ['Escalation probability', 'Likelihood that incident handling should intensify.'],
  ['Anomaly', 'Deviation from expected short-window market behavior.'],
  ['Trust', 'Current quality and reliability of score inputs and models.'],
  ['Replay', 'Reconstruction of historical context for audit and review.'],
  ['Case', 'Investigation workspace linked to an incident.'],
  ['Fallback mode', 'Deterministic scoring path used when model artifacts are unavailable.'],
  ['Model registry', 'Source of model versions and artifact lineage metadata.'],
]

export default function GlossaryPage() {
  const { filters, apply } = useGlobalFilters()
  return (
    <AppShell title="Explain / Glossary" subtitle="Plain-English terms for operators and non-experts." filters={filters} setFilters={apply}>
      <section className="card">
        {terms.map(([k, v]) => <div className="kv" key={k}><strong>{k}</strong><span>{v}</span></div>)}
      </section>
    </AppShell>
  )
}
