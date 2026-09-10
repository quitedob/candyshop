import { test, expect } from '@playwright/test'

const existingPost = {
  id: 'existing-post', slug: 'trends-candy-industry-2026', title: 'Top 10 Candy Industry Trends for 2026',
  excerpt: 'Article preview', content: '<p>Article body supplied by the CMS.</p>', category: 'market_insights',
  author: { name: 'CandyPro' }, publishedAt: '2026-03-14T00:00:00Z', readTime: 8, thumbnail: '', tags: [],
}

test.beforeEach(async ({ page }) => {
  // Production navigation may fetch an SSR payload before the component API call.
  // Keep the fixture at the HTTP boundary instead of inheriting live CMS data.
  await page.route('**/_payload.json*', route => route.fulfill({ status: 404, body: '' }))
  await page.route('**/api/v1/auth/**', route => route.fulfill({ status: 401, json: { message: 'Unauthenticated' } }))
  await page.route('**/api/v1/public/posts?**', route => route.fulfill({
    json: { data: [existingPost], pagination: { page: 1, limit: 50, total: 1, totalPages: 1 } },
  }))
  await page.route('**/api/v1/public/posts/*/related?**', route => route.fulfill({ json: [] }))
})

test('a listed article opens with CMS content and localized return links', async ({ page }) => {
  await page.route(`**/api/v1/public/posts/${existingPost.slug}`, route => route.fulfill({ json: existingPost }))
  await page.goto('/en/about')
  await expect(page.locator('.header__theme-toggle')).toBeEnabled()
  await page.locator('.header__nav a[href="/en/blog"]').click()
  await page.locator('.featured-post__inner').click()
  await expect(page.locator('.article__title')).toHaveText(existingPost.title)
  await expect(page.locator('.article__content')).toContainText('Article body supplied by the CMS.')
  await expect(page.locator('.breadcrumb a[href="/en/blog"]')).toBeVisible()
})

test('a CMS outage is not presented as a nonexistent article', async ({ page }) => {
  await page.route(`**/api/v1/public/posts/${existingPost.slug}`, route => route.fulfill({ status: 503, json: { message: 'Service unavailable' } }))
  await page.goto('/en/about')
  await expect(page.locator('.header__theme-toggle')).toBeEnabled()
  await page.locator('.header__nav a[href="/en/blog"]').click()
  await page.locator('.featured-post__inner').click()
  await expect(page.getByText('Unable to Load Post', { exact: true })).toBeVisible()
  await expect(page.getByText('Post Not Found', { exact: true })).toHaveCount(0)
})

test('an actually deleted article still reports not found', async ({ page }) => {
  await page.route(`**/api/v1/public/posts/${existingPost.slug}`, route => route.fulfill({ status: 404, json: { message: 'Post not found' } }))
  await page.goto('/en/about')
  await expect(page.locator('.header__theme-toggle')).toBeEnabled()
  await page.locator('.header__nav a[href="/en/blog"]').click()
  await page.locator('.featured-post__inner').click()
  await expect(page.getByText('Post Not Found', { exact: true })).toBeVisible()
})
