'use client'
import AppShell, { useGlobalFilters } from '../../components/AppShell'

export default function AboutPage() {
  const { filters, apply } = useGlobalFilters()
  return (
    <AppShell title="About Sentinel" subtitle="Production-minded market and risk operator workspace." filters={filters} setFilters={apply}>
      <section className="card"><h3>What Sentinel is</h3><p>Sentinel connects streaming market signals to incidents and investigation workflow with explicit trust and fallback labeling.</p></section>
      <section className="card"><h3>Who it is for</h3><p>Analysts and operators triaging incidents, escalating cases, and reviewing governance posture.</p></section>
      <section className="card"><h3>Pipeline</h3><p>Ticks → candles → features → scoring → alerts → incidents → cases → replay/governance views.</p></section>
      <section className="card"><h3>Honest scope</h3><p>This build is production-minded but simplified: fallback scoring is deterministic, replay is metadata-first with derived timeline, and destructive governance actions are intentionally omitted.</p></section>
    </AppShell>
  )
}
