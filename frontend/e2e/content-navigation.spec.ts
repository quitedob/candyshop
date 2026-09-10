import { test, expect, type Page } from '@playwright/test'
import AxeBuilder from '@axe-core/playwright'

test.beforeEach(async ({ page }) => {
  // These component checks do not depend on a live authentication service.
  await page.route('**/api/v1/auth/**', route => route.fulfill({ status: 401, json: { message: 'Unauthenticated' } }))
})

async function openCatalogWithFixtures(page: Page) {
  // Force client API loading when production navigation would prefetch SSR data.
  await page.route('**/_payload.json*', route => route.fulfill({ status: 404, body: '' }))
  await page.route('**/api/v1/public/categories', route => route.fulfill({ json: [
    {
      slug: 'gummy-candy', name: '软糖', description: '', thumbnail: '/images/categories/gummy-candy.jpg',
      productCount: 2, translations: { en: { name: 'Gummy Candy', description: 'Soft, chewy candies in various fruit flavors and fun shapes.' } },
    },
    {
      slug: 'test-candy-category', name: '测试糖果分类', description: '用于测试的糖果产品分类',
      thumbnail: '', productCount: 0, translations: { zh: { name: '测试糖果分类' } },
    },
  ] }))
  await page.route('**/api/v1/public/products?**', route => route.fulfill({ json: { data: [] } }))
  // Client navigation makes the public API fixture independent of the SSR backend.
  await page.goto('/en/about')
  await expect(page.locator('.header__theme-toggle')).toBeEnabled()
  await page.locator('.header__nav a[href="/en/products"]').click()
  await expect(page.locator('.category-card__title').first()).toHaveText('Gummy Candy')
}

for (const localePrefix of ['', '/en']) {
  for (const route of ['/about', '/factory-quality']) {
    test(`unavailable documents offer localized contact on ${localePrefix}${route}`, async ({ page }) => {
      await page.goto(`${localePrefix}${route}`)
      await expect(page.locator('.header__theme-toggle')).toBeEnabled()
      await expect(page.locator('a[href$=".pdf"], a[download]')).toHaveCount(0)
      const documentActions = page.locator('.cert-item__download, .cert-card__download, .download-item')
      expect(await documentActions.count()).toBeGreaterThan(0)
      for (const documentAction of await documentActions.all()) {
        await expect(documentAction).toHaveAttribute('href', `${localePrefix}/contact`)
      }
      await documentActions.first().click()
      await expect(page).toHaveURL(new RegExp(`${localePrefix}/contact$`))
    })
  }
}

test('category images settle to a placeholder when photographs are unavailable', async ({ page }) => {
  await page.route('**/images/categories/**', route => route.abort())
  await openCatalogWithFixtures(page)
  await expect(page.locator('.header__theme-toggle')).toBeEnabled()
  const categoryCards = page.locator('.products-categories .category-card')
  expect(await categoryCards.count()).toBeGreaterThan(0)
  for (const categoryCard of await categoryCards.all()) {
    await categoryCard.scrollIntoViewIfNeeded()
    await expect(categoryCard.locator('.category-image-placeholder')).toBeVisible()
    await expect(categoryCard.locator('img')).toHaveCount(0)
  }
})

test('catalog hero and category text meet contrast requirements in dark mode', async ({ page }) => {
  await openCatalogWithFixtures(page)
  const themeToggle = page.locator('.header__theme-toggle')
  await expect(themeToggle).toBeEnabled()
  if (!await page.locator('html').evaluate(element => element.classList.contains('dark'))) {
    await themeToggle.click()
  }
  await expect(page.locator('html')).toHaveClass(/dark/)
  await page.locator('.products-page').evaluate(async element => {
    await Promise.all(element.getAnimations({ subtree: true }).map(animation => animation.finished.catch(() => {})))
  })
  const contrastResults = await new AxeBuilder({ page })
    .include('.products-hero')
    .include('.products-categories')
    .withRules(['color-contrast'])
    .analyze()
  expect(contrastResults.violations).toEqual([])
})

for (const portal of ['admin', 'customer']) {
  test(`${portal} mobile navigation traps and restores keyboard focus`, async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 812 })
    // Mock transport only: no real accounts, credentials, or backend writes.
    await page.route('**/api/v1/**', async route => {
      const endpoint = new URL(route.request().url()).pathname
      if (endpoint.endsWith('/auth/me')) {
        await route.fulfill({ json: {
          id: 'sidebar-test-user', email: 'sidebar@example.invalid', firstName: 'Sidebar',
          lastName: 'Test', role: portal, status: 'active', emailVerified: true,
        } })
      } else {
        await route.fulfill({ json: { data: [], items: [] } })
      }
    })
    await page.goto(portal === 'admin' ? '/en/admin' : '/en/customer/help')
    const sidebar = page.locator(`.${portal}-sidebar`)
    const menuButton = page.locator(`.${portal}-mobile-topbar__btn`)
    const closeButton = page.locator(`.${portal}-sidebar__close`)
    await expect(menuButton).toBeVisible()
    await expect(sidebar).toHaveAttribute('inert', '')
    await menuButton.focus()
    await page.keyboard.press('Tab')
    expect(await sidebar.evaluate(element => element.contains(document.activeElement))).toBe(false)

    await menuButton.focus()
    await page.keyboard.press('Enter')
    await expect(sidebar).toHaveAttribute('aria-modal', 'true')
    await expect(closeButton).toBeFocused()
    await expect(page.locator(`.${portal}-main`)).toHaveAttribute('inert', '')

    const firstFocusable = sidebar.locator('a').first()
    const lastFocusable = sidebar.locator(`.${portal}-sidebar__logout`).last()
    await firstFocusable.focus()
    await page.keyboard.press('Shift+Tab')
    await expect(lastFocusable).toBeFocused()
    await page.keyboard.press('Tab')
    await expect(firstFocusable).toBeFocused()
    await page.keyboard.press('Escape')
    await expect(sidebar).toHaveAttribute('inert', '')
    await expect(menuButton).toBeFocused()

    await menuButton.click()
    await closeButton.click()
    await expect(menuButton).toBeFocused()
    await page.setViewportSize({ width: 1280, height: 800 })
    await expect(sidebar).not.toHaveAttribute('inert')
    await expect(page.locator(`.${portal}-main`)).not.toHaveAttribute('inert')
  })
}
