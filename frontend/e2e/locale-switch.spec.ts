import { test, expect } from '@playwright/test'

test.describe('Locale switching', () => {
  test('legal pages render without raw i18n keys', async ({ page }) => {
    await page.goto('/privacy')
    await expect(page.locator('body')).not.toContainText('legal.privacy')
    await page.goto('/en/privacy')
    await expect(page.locator('body')).not.toContainText('legal.privacy')
  })
})
