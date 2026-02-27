import { test, expect } from '@playwright/test'

test('login -> command center -> globe renders', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('button', { name: /sign in/i }).click()
  await expect(page).toHaveURL(/command-center|executive/)
  await expect(page.getByText(/Risk globe/i)).toBeVisible()
  await expect(page.locator('svg[aria-label="world risk map"]')).toBeVisible()
})

test('globe/map country click updates filters and queue request scope', async ({ page }) => {
  await page.goto('/command-center')
  await page.locator('svg[aria-label="world risk map"] path').first().click()
  const countryInput = page.getByPlaceholder('country')
  await expect(countryInput).toHaveValue(/US|GB|DE|BR|JP|AU/)
  const chosen = await countryInput.inputValue()

  const reqPromise = page.waitForRequest((r) => r.url().includes('http://localhost:8085/queue') && r.url().includes(`countryCode=${encodeURIComponent(chosen)}`))
  await page.goto('/queue')
  await reqPromise
})

test('switch locale to Arabic sets rtl html direction', async ({ page }) => {
  await page.goto('/command-center')
  await page.getByLabel('Locale').selectOption('ar')
  await expect(page.locator('html')).toHaveAttribute('dir', 'rtl')
})

test('SSE event arrives and activity feed updates', async ({ page }) => {
  await page.goto('/command-center')
  await expect(page.getByText(/SSE status/i)).toBeVisible()
  await expect(page.getByText(/SSE status:\s*open/i)).toBeVisible({ timeout: 15000 })
  await expect(page.getByText(/Latest event/i)).toBeVisible({ timeout: 15000 })
})

test('viewer cannot mutate case workflow endpoint', async ({ request }) => {
  const res = await request.post('http://localhost:8083/incidents/1/promote-case', { data: { reason: 'e2e-check' } })
  expect([401, 403]).toContain(res.status())
})
