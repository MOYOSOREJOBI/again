'use client'

import { useEffect, useState } from 'react'
import { api } from '../../../lib/api'

export default function ReplayPage({ params }: { params: { job: string } }) {
  const [job, setJob] = useState<any>(null)
  useEffect(() => { api.replay(params.job).then(setJob).catch(() => setJob(null)) }, [params.job])
  if (!job) return <main style={{ padding: 24 }}>Replay metadata unavailable.</main>
  return <main style={{ padding: 24 }}><h1>Replay {params.job}</h1><p>Status: {job.status}</p><pre>{JSON.stringify(job, null, 2)}</pre><p>Replay lane rendering is partial in this build; metadata is authoritative.</p></main>
}
