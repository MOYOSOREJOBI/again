import { test, expect } from '@playwright/test'

test('login loads command center and globe', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('button', { name: /sign in/i }).click()
  await expect(page).toHaveURL(/command-center|executive/)
  await expect(page.getByText(/Risk globe/i)).toBeVisible()
})

test('country click updates filter path', async ({ page }) => {
  await page.goto('/command-center')
  await page.getByRole('button').first().click()
  await expect(page.getByText(/World risk map/i)).toBeVisible()
})

test('arabic locale sets rtl', async ({ page }) => {
  await page.goto('/command-center')
  await page.getByLabel('Locale').selectOption('ar')
  await expect(page.locator('html')).toHaveAttribute('dir', 'rtl')
})

test('sse endpoint exists', async ({ request }) => {
  const res = await request.get('/api/sse/alerts')
  expect(res.status()).toBe(200)
})

test('viewer cannot mutate case', async ({ request }) => {
  const res = await request.post('http://localhost:8083/incidents/1/promote-case', { data: {} })
  expect([401,403]).toContain(res.status())
})
