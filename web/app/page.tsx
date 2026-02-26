'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { api } from '../lib/api'

export default function HomePage() {
  const router = useRouter()
  const [email, setEmail] = useState('admin@sentinel.local')
  const [password, setPassword] = useState('Sentinel#123')
  const [error, setError] = useState('')

  useEffect(() => {
    api.me().then(() => router.replace('/command-center')).catch(() => null)
  }, [router])

  return (
    <main style={{ padding: 24, maxWidth: 420 }}>
      <h1>Sentinel Login</h1>
      <p>Sign in to access the operator workspace.</p>
      <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="email" style={{ width: '100%', marginBottom: 8 }} />
      <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="password" style={{ width: '100%', marginBottom: 8 }} />
      <button onClick={async () => { setError(''); try { await api.login(email, password); router.replace('/command-center') } catch (e: any) { setError(e.message) } }}>Login</button>
      {error ? <p style={{ color: 'tomato' }}>{error}</p> : null}
    </main>
  )
}
