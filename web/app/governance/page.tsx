'use client'

import { useEffect, useState } from 'react'
import { api } from '../../lib/api'

export default function GovernancePage() {
  const [models, setModels] = useState<any[]>([])
  useEffect(() => { api.models().then(setModels).catch(() => setModels([])) }, [])
  return <main style={{ padding: 24 }}><h1>Governance</h1><p>Replay remains partial; this page reflects real model registry state only.</p><ul>{models.map((m) => <li key={m.id}>{m.model_name}:{m.version} ({m.state})</li>)}</ul></main>
}
