'use client'

import AppShell, { useGlobalFilters } from '../../components/AppShell'

export default function GlossaryPage() {
  const { filters, apply } = useGlobalFilters()
  return (
    <AppShell title="Glossary" subtitle="Plain-language definitions and ranking impact." filters={filters} setFilters={apply}>
      <section className="card">
        <div className="kv"><span>Anomaly score</span><span>How unusual behavior is vs baseline; higher tends to raise ranking.</span></div>
        <div className="kv"><span>Z-score</span><span>Standardized distance from average; helps normalize unusualness.</span></div>
        <div className="kv"><span>EWMA volatility</span><span>Smoothed volatility trend; spikes can increase rank pressure.</span></div>
        <div className="kv"><span>Escalation probability</span><span>Likelihood incident becomes actionable.</span></div>
        <div className="kv"><span>Trust penalty</span><span>Confidence reduction from degraded data quality; lowers certainty and can change rank.</span></div>
        <div className="kv"><span>Incident</span><span>Ranked anomaly cluster ready for triage.</span></div>
        <div className="kv"><span>Case</span><span>Governed investigation record promoted from incident.</span></div>
        <div className="kv"><span>Replay</span><span>Historical reconstruction metadata; full recompute is partial in current release.</span></div>
      </section>
    </AppShell>
  )
}
