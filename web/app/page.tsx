'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { api } from '../lib/api'

export default function HomePage() {
  const router = useRouter()
  const [email, setEmail] = useState('admin@sentinel.local')
  const [password, setPassword] = useState('Sentinel#123')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    api.me().then((me) => router.replace(me.role === 'viewer' ? '/executive' : '/command-center')).catch(() => null)
  }, [router])

  const submit = async () => {
    setError('')
    setLoading(true)
    try {
      await api.login(email, password)
      const me = await api.me()
      router.replace(me.role === 'viewer' ? '/executive' : '/command-center')
    } catch (e: any) {
      setError(e.message || 'Authentication failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="login-root">
      <section className="login-card">
        <h1>Sentinel</h1>
        <p>Risk intelligence operator workbench for ranked incidents, governed cases, trust monitoring, and replay metadata.</p>
        <div className="form-grid">
          <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="email" />
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="password" />
          <button disabled={loading} onClick={submit}>{loading ? 'Signing in…' : 'Sign in to Sentinel'}</button>
          {error ? <p className="error-text">{error}</p> : null}
        </div>
      </section>
    </main>
  )
}
