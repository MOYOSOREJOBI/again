import { getJSON } from '../../../lib/api'

export default async function IncidentDetail({ params }: { params: { id: string } }) {
  const row = await getJSON(`/incident/${params.id}`).catch(() => null)
  if (!row) return <main style={{ padding: 24 }}>Incident not found</main>
  return <main style={{ padding: 24 }}><h1>Incident {row.id}</h1><pre>{JSON.stringify(row, null, 2)}</pre></main>
}
