'use client'

type Row = { country: string; region: string; sector: string; incident_count: number }

export default function WorldRiskMap({ rows }: { rows: Row[] }) {
  return (
    <section>
      <h3>World Risk Map (table fallback)</h3>
      <table>
        <thead><tr><th>Country</th><th>Region</th><th>Sector</th><th>Incidents</th></tr></thead>
        <tbody>
          {rows.map((r, i) => <tr key={i}><td>{r.country}</td><td>{r.region}</td><td>{r.sector}</td><td>{r.incident_count}</td></tr>)}
        </tbody>
      </table>
    </section>
  )
}
