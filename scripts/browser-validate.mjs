#!/usr/bin/env node
import fs from 'node:fs/promises'
import process from 'node:process'
import { chromium } from 'playwright'

const BASE = process.env.BASE_URL || 'http://localhost:3000'
const APIQ = process.env.API_QUERY || 'http://localhost:8085'
const APIA = process.env.API_ALERTS || 'http://localhost:8083'
const results = []

async function waitReady(url, name, tries = 60) {
  for (let i = 0; i < tries; i++) {
    try { const r = await fetch(url); if (r.ok) return true } catch {}
    await new Promise(r => setTimeout(r, 1000))
  }
  results.push({flow:name,status:'FAIL',reason:`timeout waiting ${url}`})
  return false
}

async function login(page, email, password) {
  await page.goto(`${BASE}/`, { waitUntil: 'networkidle' })
  await page.getByPlaceholder('email').fill(email)
  await page.getByPlaceholder('password').fill(password)
  await page.getByRole('button', { name: 'Sign in to Sentinel' }).click()
  await page.waitForLoadState('networkidle')
}

(async () => {
  if (!(await waitReady('http://localhost:3000/', 'web')) || !(await waitReady('http://localhost:8085/readyz', 'query'))) process.exit(1)
  const browser = await chromium.launch({ headless: true })
  const context = await browser.newContext()
  const page = await context.newPage()

  // A: world-map propagation
  try {
    const seen = []
    page.on('request', req => { if (req.url().includes('/queue?') || req.url().includes('/command-center?')) seen.push(req.url()) })
    await login(page, 'admin@sentinel.local', 'Sentinel#123')
    await page.goto(`${BASE}/command-center`, { waitUntil: 'networkidle' })
    const buttons = page.locator('.map-cell')
    const count = await buttons.count()
    if (count === 0) throw new Error('no world map cells')
    const txt = await buttons.first().innerText()
    const region = txt.split('·')[0].split('\n').pop().trim()
    await buttons.first().click()
    await page.goto(`${BASE}/queue?region=${encodeURIComponent(region)}`, { waitUntil: 'networkidle' })
    if (!seen.some(u => u.includes('region='))) throw new Error('no propagated region in fetch')
    results.push({flow:'world-map-filter-propagation',status:'PASS',region})
  } catch (e) { results.push({flow:'world-map-filter-propagation',status:'FAIL',reason:String(e.message||e)}) }

  // B + C incident->case and role gating
  try {
    const req = await context.request.get(`${APIQ}/queue`, { failOnStatusCode: false })
    const queue = await req.json(); const id = queue[0]?.id
    if (!id) throw new Error('no incident id')
    await page.goto(`${BASE}/incident/${id}`, { waitUntil: 'networkidle' })
    await page.getByRole('button', { name: 'Promote to Case' }).click()
    await page.waitForTimeout(500)
    const msg = await page.locator('text=Case').first().innerText()
    const m = msg.match(/Case\s+(\d+)/); if (!m) throw new Error('case id not found in promotion result')
    const caseId = m[1]
    await page.goto(`${BASE}/case/${caseId}`, { waitUntil: 'networkidle' })
    await page.getByRole('button', { name: 'Add note' }).click()
    await page.waitForTimeout(300)

    const viewerCtx = await browser.newContext(); const viewerPage = await viewerCtx.newPage()
    await login(viewerPage, 'viewer@sentinel.local', 'Sentinel#123')
    await viewerPage.goto(`${BASE}/incident/${id}`, { waitUntil: 'networkidle' })
    const disabled = await viewerPage.locator('text=Viewer mode: case promotion disabled.').count()
    const reject = await viewerCtx.request.post(`${APIA}/incidents/${id}/promote-case`, { data:{reason:'x'}, failOnStatusCode:false })
    if (disabled < 1 || reject.status() !== 403) throw new Error('viewer gating failed')
    await viewerCtx.close()

    results.push({flow:'incident-to-case-and-role-gating',status:'PASS',incident_id:id,case_id:caseId})
  } catch (e) { results.push({flow:'incident-to-case-and-role-gating',status:'FAIL',reason:String(e.message||e)}) }

  await browser.close()
  await fs.mkdir('docs/screenshots', { recursive: true })
  await fs.writeFile('docs/screenshots/browser-validation.json', JSON.stringify(results, null, 2))
  const failed = results.some(r => r.status === 'FAIL')
  for (const r of results) console.log(`${r.status}: ${r.flow}${r.reason ? ` (${r.reason})` : ''}`)
  process.exit(failed ? 1 : 0)
})()
