import TrustStrip from '../../components/TrustStrip'
import WorldRiskMap from '../../components/WorldRiskMap'
import { getJSON } from '../../lib/api'

export default async function CommandCenterPage() {
  const trust = await getJSON('/trust').catch(() => ({ degraded: [] }))
  const world = await getJSON('/world-map').catch(() => [])
  return (
    <main style={{ padding: 24 }}>
      <h1>Command Center</h1>
      <TrustStrip degraded={trust.degraded?.length || 0} />
      <WorldRiskMap rows={world} />
    </main>
  )
}
