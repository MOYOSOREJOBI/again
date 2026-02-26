import { getJSON } from '../../lib/api'
import Link from 'next/link'

export default async function QueuePage() {
  const rows = await getJSON('/queue').catch(() => [])
  return (
    <main style={{ padding: 24 }}>
      <h1>Ranked Incident Queue</h1>
      <table><thead><tr><th>ID</th><th>Symbol</th><th>Severity</th><th>Priority</th><th>Status</th></tr></thead><tbody>
      {rows.map((r: any) => <tr key={r.id}><td><Link href={`/incident/${r.id}`}>{r.id}</Link></td><td>{r.symbol}</td><td>{r.severity_band}</td><td>{r.priority_score?.toFixed?.(3) || r.priority_score}</td><td>{r.status}</td></tr>)}
      </tbody></table>
    </main>
  )
}
