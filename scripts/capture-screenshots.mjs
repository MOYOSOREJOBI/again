#!/usr/bin/env node
import fs from 'node:fs/promises'
import process from 'node:process'
import { chromium } from 'playwright'
const OUT = process.env.OUT_DIR || 'docs/screenshots'
const BASE = process.env.BASE_URL || 'http://localhost:3000'
await fs.mkdir(OUT, { recursive: true })
const manifest = []
const routes = [['/command-center','command-center'],['/queue','queue'],['/incident/{incidentId}','incident'],['/case','case'],['/case/{caseId}','case-detail'],['/trust','trust'],['/replay/{jobId}','replay'],['/governance','governance'],['/executive','executive'],['/about','about'],['/glossary','glossary']]
const browser = await chromium.launch({ headless: true })
const context = await browser.newContext({ viewport: { width: 1600, height: 1000 } })
const page = await context.newPage()
try {
  await page.goto(`${BASE}/`, { waitUntil: 'networkidle' })
  await page.getByPlaceholder('email').fill('admin@sentinel.local')
  await page.getByPlaceholder('password').fill('Sentinel#123')
  await page.getByRole('button', { name: 'Sign in to Sentinel' }).click()
  await page.waitForLoadState('networkidle')
} catch (e) {
  manifest.push({ route: 'login', status: 'FAIL', reason: String(e.message || e) })
  await fs.writeFile(`${OUT}/manifest.json`, JSON.stringify(manifest, null, 2))
  await browser.close(); process.exit(1)
}
const queueResp = await context.request.get('http://localhost:8085/queue', { failOnStatusCode: false })
const queue = queueResp.ok() ? await queueResp.json() : []
const casesResp = await context.request.get('http://localhost:8085/cases', { failOnStatusCode: false })
const caseRows = casesResp.ok() ? await casesResp.json() : []
const incidentId = queue[0]?.id || 1
const caseId = caseRows[0]?.id
for (const [tpl, name] of routes) {
  if (tpl.includes('{caseId}') && !caseId) { manifest.push({ route: tpl, status: 'SKIP', reason: 'no case id available' }); continue }
  const route = tpl.replace('{incidentId}', incidentId).replace('{caseId}', caseId || '').replace('{jobId}', 'demo')
  const output = `${OUT}/${name}.png`
  try { await page.goto(`${BASE}${route}`, { waitUntil: 'networkidle', timeout: 30000 }); await page.screenshot({ path: output, fullPage: true }); manifest.push({ route, status: 'PASS', output }); console.log(`PASS: ${route}`) }
  catch (e) { manifest.push({ route, status: 'FAIL', reason: String(e.message || e) }); console.log(`FAIL: ${route}`) }
}
await fs.writeFile(`${OUT}/manifest.json`, JSON.stringify(manifest, null, 2))
await browser.close()
if (manifest.some(m => m.status === 'FAIL')) process.exit(1)
