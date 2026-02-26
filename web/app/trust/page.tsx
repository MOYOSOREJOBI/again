import TrustStrip from '../../components/TrustStrip'
import { getJSON } from '../../lib/api'

export default async function TrustPage() {
  const trust = await getJSON('/trust').catch(() => ({ degraded: [] }))
  return <main style={{ padding: 24 }}><h1>Trust</h1><TrustStrip degraded={trust.degraded?.length || 0} /><pre>{JSON.stringify(trust, null, 2)}</pre></main>
}
