import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'

test('governance page retains admin gating text', () => {
  const s = fs.readFileSync('app/governance/page.tsx', 'utf8')
  assert.equal(s.includes('admin-only'), true)
})
