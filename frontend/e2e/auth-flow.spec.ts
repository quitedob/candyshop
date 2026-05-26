import { test, expect } from '@playwright/test'

const adminEmail = process.env.E2E_ADMIN_EMAIL || ''
const adminPassword = process.env.E2E_ADMIN_PASSWORD || ''

test.describe('Authenticated admin smoke', () => {
  test.skip(!adminEmail || !adminPassword, 'Set E2E_ADMIN_EMAIL and E2E_ADMIN_PASSWORD to run')

  test('admin can login and view dashboard', async ({ page }) => {
    await page.goto('/auth/login')
    await page.fill('#email', adminEmail)
    await page.fill('#password', adminPassword)
    await page.click('button[type="submit"]')
    await page.waitForURL(/\/admin/, { timeout: 20_000 })
    await expect(page.locator('body')).toContainText(/Dashboard|仪表盘|dashboard/i)
  })
})
