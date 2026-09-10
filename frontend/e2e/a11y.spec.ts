import { test, expect } from '@playwright/test'
import AxeBuilder from '@axe-core/playwright'

// Public routes that render without authentication. Product detail and portal
// states have separate flow checks.
const ROUTES = [
  '/',
  '/about',
  '/faq',
  '/products',
  '/blog',
  '/cases-clients',
  '/oem-solutions',
  '/factory-quality',
  '/privacy',
  '/terms',
  '/contact',
]

// Mute intentional patterns that are not real defects:
// - autocomplete=off is disabled globally for privacy/spam reasons.
const RULES_TO_SKIP = ['autocomplete-valid']

test.describe('a11y — WCAG 2 AA on public routes', () => {
  test.setTimeout(120_000)
  for (const theme of ['light', 'dark']) {
    for (const route of ROUTES) {
      test(`${route} in ${theme} mode has no serious/critical violations`, async ({ page }) => {
        await page.emulateMedia({ reducedMotion: 'reduce' })
        await page.goto(route, { waitUntil: 'load' })
        const themeToggle = page.locator('.header__theme-toggle')
        await expect(themeToggle).toBeEnabled({ timeout: 15_000 })
        const isDark = await page.locator('html').evaluate(element => element.classList.contains('dark'))
        if (isDark !== (theme === 'dark')) await themeToggle.click()
        await page.locator('html').evaluate(async element => {
          await Promise.all(element.getAnimations({ subtree: true })
            .filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity)
            .map(animation => animation.finished.catch(() => {})))
        })

        const results = await new AxeBuilder({ page })
          .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
          .disableRules(RULES_TO_SKIP)
          .analyze()

        const serious = results.violations.filter((v) =>
          ['serious', 'critical'].includes(v.impact ?? '')
        )

        // Surface the failures for the sweep: id, impact, help + target selectors.
        if (serious.length > 0) {
          await test.info().attach('accessibility-violations', {
            body: JSON.stringify(serious),
            contentType: 'application/json',
          })
          const summary = serious
            .map((v) => `${v.impact} [${v.id}] ${v.help} @ ${v.nodes[0]?.target.join(' ')}`)
            .join('\n')
          expect.soft(serious, `Serious/critical a11y violations on ${route}:\n${summary}`).toHaveLength(0)
        }

        // Also count moderate for the tracking report (non-blocking).
        const moderate = results.violations.filter((v) => v.impact === 'moderate').length
        test.info().annotations.push({
          type: 'moderate-count',
          description: `${moderate} moderate violations on ${route}`,
        })
      })
    }
  }
})
