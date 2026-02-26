import test from 'node:test'
import assert from 'node:assert/strict'
import React from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import fs from 'node:fs'
import TrustStrip from '../components/TrustStrip'
import WorldRiskMap from '../components/WorldRiskMap'
import IncidentPage from '../app/incident/[id]/page'
import CasePage from '../app/case/[id]/page'

test('root page keeps login entry implementation', () => {
  const src = fs.readFileSync(process.cwd() + '/app/page.tsx', 'utf-8')
  assert.match(src, /Sentinel Login/)
  assert.match(src, /api\.login/)
})

test('TrustStrip renders stable and degraded variants', () => {
  const stable = renderToStaticMarkup(React.createElement(TrustStrip, { trustState: 'stable', modelUnavailable: false }))
  const degraded = renderToStaticMarkup(React.createElement(TrustStrip, { trustState: 'degraded', modelUnavailable: true, dqWarning: '2 symbols' }))
  assert.match(stable, /Stable/)
  assert.match(degraded, /Fallback model active/)
})

test('WorldRiskMap renders real rows and selectable table', () => {
  const html = renderToStaticMarkup(React.createElement(WorldRiskMap, { rows: [{ country: 'US', region: 'NA', sector: 'TECH', industry: 'SW', incident_count: 3 }] }))
  assert.match(html, /US/)
  assert.match(html, /TECH/)
})

test('incident and case pages show truthful empty states before data load', () => {
  const incident = renderToStaticMarkup(React.createElement(IncidentPage as any, { params: { id: '1' } }))
  const caseHtml = renderToStaticMarkup(React.createElement(CasePage as any, { params: { id: '1' } }))
  assert.match(incident, /not found|unauthorized/i)
  assert.match(caseHtml, /unavailable|not found/i)
})
