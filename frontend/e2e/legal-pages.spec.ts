import { test, expect } from '@playwright/test'

test.describe('Legal pages i18n smoke', () => {
  test('zh privacy page renders policy sections', async ({ page }) => {
    await page.goto('/privacy')
    await expect(page.locator('h1, h2').first()).toBeVisible()
    await expect(page.locator('body')).toContainText(/隐私|Privacy/i)
  })

  test('ko legal cookies page renders translated title', async ({ page }) => {
    await page.goto('/ko/legal/cookies')
    await expect(page.locator('h1')).toBeVisible()
    await expect(page.locator('body')).not.toContainText('cookie_title')
    await expect(page.locator('body')).toContainText(/쿠키|Cookie/i)
  })

  test('en returns policy page loads', async ({ page }) => {
    await page.goto('/en/legal/returns')
    await expect(page.locator('h1')).toBeVisible()
    await expect(page.locator('body')).toContainText(/Return|Refund/i)
  })

  test('id shipping policy page has no raw i18n keys', async ({ page }) => {
    await page.goto('/id/legal/shipping')
    await expect(page.locator('h1')).toBeVisible()
    await expect(page.locator('body')).not.toContainText('shipping_methods_title')
  })
})
