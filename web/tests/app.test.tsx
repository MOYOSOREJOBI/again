import test from 'node:test'
import assert from 'node:assert/strict'
import React from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import fs from 'node:fs'
import TrustStrip from '../components/TrustStrip'
import WorldRiskMap from '../components/WorldRiskMap'

test('login page keeps branded entry copy', () => {
  const src = fs.readFileSync(process.cwd() + '/app/page.tsx', 'utf-8')
  assert.match(src, /Sign in to Sentinel/)
  assert.match(src, /Risk intelligence operator workbench/)
})

test('TrustStrip renders healthy and degraded variants with richer copy', () => {
  const stable = renderToStaticMarkup(React.createElement(TrustStrip, { trustState: 'stable', modelUnavailable: false }))
  const degraded = renderToStaticMarkup(React.createElement(TrustStrip, { trustState: 'degraded', modelUnavailable: true, dqWarning: '2 symbols' }))
  assert.match(stable, /Model healthy/)
  assert.match(degraded, /Open Trust/)
})

test('WorldRiskMap renders truthful empty state', () => {
  const html = renderToStaticMarkup(React.createElement(WorldRiskMap, { rows: [] }))
  assert.match(html, /No regional aggregates yet/) 
})

test('WorldRiskMap renders interactive region cells', () => {
  const html = renderToStaticMarkup(React.createElement(WorldRiskMap, { rows: [{ country: 'US', region: 'NA', sector: 'TECH', industry: 'SW', incident_count: 3 }] }))
  assert.match(html, /US/)
  assert.match(html, /incidents/)
})

test('shell and role nav structure exists in AppShell source', () => {
  const src = fs.readFileSync(process.cwd() + '/components/AppShell.tsx', 'utf-8')
  assert.match(src, /left-nav/)
  assert.match(src, /Governance/)
  assert.match(src, /roleAllowed/)
})

test('command center and queue include non-empty state product copy', () => {
  const cc = fs.readFileSync(process.cwd() + '/app/command-center/page.tsx', 'utf-8')
  const queue = fs.readFileSync(process.cwd() + '/app/queue/page.tsx', 'utf-8')
  assert.match(cc, /No open critical incidents right now/)
  assert.match(queue, /Queue is clear for current filters/)
})
