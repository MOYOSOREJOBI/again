import test from 'node:test'
import assert from 'node:assert/strict'
import React from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import Page from '../app/page'

test('root page renders deterministic loading state on initial render', () => {
  const html = renderToStaticMarkup(React.createElement(Page))
  assert.match(html, /Loading Sentinel/)
})
