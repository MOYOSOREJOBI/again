'use client'

import AppShell, { useGlobalFilters } from '../../components/AppShell'

export default function AboutPage() {
  const { filters, apply } = useGlobalFilters()
  return (
    <AppShell title="About Sentinel" subtitle="What Sentinel does and how trust + replay fit the workflow." filters={filters} setFilters={apply}>
      <section className="card">
        <p>Sentinel turns streaming market and telemetry data into ranked incidents, analyst queue actions, and governed cases.</p>
        <p>Trust signals explain whether data quality reduces confidence and ranking.</p>
        <p>Replay is currently metadata-first: version, timing, and diff context are reproducible while full lane recompute remains partial.</p>
      </section>
    </AppShell>
  )
}
