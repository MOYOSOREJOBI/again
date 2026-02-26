import test from 'node:test'
import assert from 'node:assert/strict'
import { t } from '../../lib/i18n'

test('i18n dictionary provides translated navigation labels', () => {
  assert.equal(t('fr', 'queue', 'Queue'), 'File priorisée')
  assert.equal(t('es', 'governance', 'Governance'), 'Gobernanza')
  assert.equal(t('pt', 'about', 'About'), 'Sobre')
  assert.equal(t('en', 'commandCenter', 'Command Center'), 'Command Center')
})
