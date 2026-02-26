import Link from 'next/link'

export default function HomePage() {
  return (
    <main style={{ padding: 24 }}>
      <h1>Sentinel</h1>
      <p>Operator workspace routes.</p>
      <ul>
        <li><Link href="/command-center">Command Center</Link></li>
        <li><Link href="/queue">Queue</Link></li>
        <li><Link href="/trust">Trust</Link></li>
        <li><Link href="/governance">Governance</Link></li>
        <li><Link href="/executive">Executive</Link></li>
      </ul>
    </main>
  )
}
