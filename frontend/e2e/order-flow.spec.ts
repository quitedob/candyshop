import { test, expect } from '@playwright/test'

test.describe('Public pages smoke', () => {
  test('homepage loads', async ({ page }) => {
    await page.goto('/')
    await expect(page.locator('body')).toBeVisible()
  })

  test('products page loads', async ({ page }) => {
    await page.goto('/products')
    await expect(page.locator('h1, h2').first()).toBeVisible()
  })

  test('faq page loads', async ({ page }) => {
    await page.goto('/faq')
    await expect(page.locator('h1, h2').first()).toBeVisible()
  })
})

test.describe('Auth pages smoke', () => {
  test('login page has email and password fields', async ({ page }) => {
    await page.goto('/auth/login')
    await expect(page.locator('#email')).toBeVisible()
    await expect(page.locator('#password')).toBeVisible()
    await expect(page.locator('button[type="submit"]')).toBeVisible()
  })

  test('register page loads', async ({ page }) => {
    await page.goto('/auth/register')
    await expect(page.locator('form')).toBeVisible()
  })
})

test.describe('Protected routes redirect', () => {
  test('admin dashboard requires auth', async ({ page }) => {
    await page.goto('/admin')
    await page.waitForURL(/\/auth\/login/, { timeout: 15_000 })
    expect(page.url()).toContain('/auth/login')
  })

  test('customer dashboard requires auth', async ({ page }) => {
    await page.goto('/customer/dashboard')
    await page.waitForURL(/\/auth\/login/, { timeout: 15_000 })
    expect(page.url()).toContain('/auth/login')
  })

  test('contact page is available to visitors', async ({ page }) => {
    await page.goto('/contact')
    await expect(page.locator('h1')).toBeVisible()
    await expect(page).toHaveURL(/\/contact$/)
  })
})

test.describe('Order flow pages (unauthenticated)', () => {
  test('quick order page redirects to login', async ({ page }) => {
    await page.goto('/customer/orders/quick')
    await page.waitForURL(/\/auth\/login/, { timeout: 15_000 })
    expect(page.url()).toContain('/auth/login')
  })
})
