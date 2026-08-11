import { test, expect } from '@playwright/test'
import AxeBuilder from '@axe-core/playwright'

// Public routes that render without authentication. Product/customer/admin pages
// are covered separately once the shared-component debt is cleared.
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
]

// Mute intentional patterns that are not real defects:
// - autocomplete=off is disabled globally for privacy/spam reasons.
const RULES_TO_SKIP = ['autocomplete-valid']

test.describe('a11y — WCAG 2 AA on public routes', () => {
  test.setTimeout(120_000)
  for (const route of ROUTES) {
    test(`${route} has no serious/critical violations`, async ({ page }) => {
      await page.goto(route, { waitUntil: 'load' })
      // Let the client app hydrate before scanning.
      await page.waitForTimeout(1500)

      const results = await new AxeBuilder({ page })
        .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
        .disableRules(RULES_TO_SKIP)
        .analyze()

      const serious = results.violations.filter((v) =>
        ['serious', 'critical'].includes(v.impact ?? '')
      )

      // Surface the failures for the sweep: id, impact, help + target selectors.
      if (serious.length > 0) {
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
})
