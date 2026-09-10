import { createHash } from 'node:crypto'
import { test, expect } from '@playwright/test'

test.use({ locale: 'zh-CN' })

// Exercise prerendered, SSR and CSR HTML in both default and prefixed locales.
// Visibility alone passes on a server-rendered page even when Vue never mounts.
for (const route of ['/', '/en/products', '/en/legal/cookies', '/auth/login', '/en/auth/login']) {
  test(`CSP permits hydration on ${route}`, async ({ page }) => {
    const runtimeErrors: string[] = []
    page.on('pageerror', error => runtimeErrors.push(error.message))
    const response = await page.goto(route)
    expect(response?.ok()).toBeTruthy()
    const policy = response!.headers()['content-security-policy'] || ''
    const scriptPolicy = policy.split(';').find(directive => directive.trim().startsWith('script-src ')) || ''
    expect(scriptPolicy).toContain("'self'")
    expect(scriptPolicy).not.toContain("'unsafe-inline'")
    const html = await response!.text()
    for (const script of html.matchAll(/<script\b[^>]*>([\s\S]*?)<\/script\s*>/gi)) {
      if (!script[1]) continue
      const hash = createHash('sha256').update(script[1].replace(/\r\n?/g, '\n')).digest('base64')
      expect(scriptPolicy, `Missing inline hash on ${route}`).toContain(`'sha256-${hash}'`)
    }
    await expect.poll(() => page.evaluate(() => Boolean(
      (window as any).__NUXT__ && (document.querySelector('#__nuxt') as any)?.__vue_app__,
    )), { timeout: 15_000 }).toBe(true)
    if (route.includes('/auth/login')) {
      await expect(page.locator('#password')).toHaveAttribute('type', 'password')
      await page.locator('.auth-password-toggle').click()
      await expect(page.locator('#password')).toHaveAttribute('type', 'text')
    } else {
      const themeToggle = page.locator('.header__theme-toggle')
      await expect(themeToggle).toBeEnabled()
      const wasDark = await page.locator('html').evaluate(element => element.classList.contains('dark'))
      await themeToggle.click()
      await expect.poll(() => page.locator('html').evaluate(element => element.classList.contains('dark'))).toBe(!wasDark)
    }
    expect(runtimeErrors).toEqual([])
  })
}

test('language selection navigates and legal links preserve that locale', async ({ page }) => {
  await page.goto('/en/legal/cookies')
  await expect(page.locator('main a[href="/en/privacy"]')).toHaveCount(1)
  await expect(page.locator('main a[href="/en/terms"]')).toHaveCount(1)
  await expect(page.locator('main a[href="/en/legal/data-processing"]')).toHaveCount(1)
  await page.locator('.header__lang .lang-switcher__trigger').click()
  await expect(page.locator('.lang-switcher__dropdown')).toBeVisible()
  await page.locator('.lang-switcher__option').filter({ hasText: '한국어' }).click()
  await expect(page).toHaveURL(/\/ko\/legal\/cookies/)
  const privacyLink = page.locator('main a[href="/ko/privacy"]').first()
  await privacyLink.click()
  await expect(page).toHaveURL(/\/ko\/privacy/)
})

test('CSP still blocks an unapproved inline script', async ({ page }) => {
  await page.goto('/en/products')
  await expect(page.locator('.header__theme-toggle')).toBeEnabled()
  const executed = await page.evaluate(() => {
    const script = document.createElement('script')
    script.textContent = 'window.unapprovedInlineScriptExecuted = true'
    document.head.append(script)
    return Boolean((window as any).unapprovedInlineScriptExecuted)
  })
  expect(executed).toBe(false)
})

test('public controls hydrate while session discovery is unavailable', async ({ page }) => {
  await page.route('**/api/v1/auth/me', () => new Promise(() => {}))
  await page.goto('/en/legal/cookies')
  await expect(page.locator('.header__theme-toggle')).toBeEnabled()
  await page.locator('.header__lang .lang-switcher__trigger').click()
  await expect(page.locator('.lang-switcher__dropdown')).toBeVisible()
})
