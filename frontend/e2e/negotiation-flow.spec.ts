import { test, expect, type APIRequestContext } from '@playwright/test'

const adminEmail = process.env.E2E_ADMIN_EMAIL || ''
const adminPassword = process.env.E2E_ADMIN_PASSWORD || ''
const customerEmail = process.env.E2E_CUSTOMER_EMAIL || ''
const customerPassword = process.env.E2E_CUSTOMER_PASSWORD || ''
const apiBase = process.env.PLAYWRIGHT_API_BASE || 'http://localhost:8080/api/v1'

const hasCreds = Boolean(adminEmail && adminPassword && customerEmail && customerPassword)

async function loginApi(request: APIRequestContext, email: string, password: string) {
  const res = await request.post(`${apiBase}/auth/login`, { data: { email, password } })
  expect(res.ok()).toBeTruthy()
  const body = await res.json()
  return body.access_token as string
}

async function createInquiry(request: APIRequestContext, token: string, email: string) {
  const res = await request.post(`${apiBase}/user/inquiries`, {
    headers: { Authorization: `Bearer ${token}` },
    multipart: {
      companyName: 'E2E Test Co',
      contactPerson: 'E2E Buyer',
      email,
      targetCountry: 'US',
      estimatedQuantity: '1000',
      message: `Negotiation E2E ${Date.now()}`,
    },
  })
  expect(res.ok()).toBeTruthy()
  const body = await res.json()
  return body.inquiryId as string
}

test.describe('Inquiry negotiation flow', () => {
  test.skip(!hasCreds, 'Set E2E_ADMIN_EMAIL/PASSWORD and E2E_CUSTOMER_EMAIL/PASSWORD')

  test('buyer offer → admin counter → buyer accepts', async ({ request, page }) => {
    const customerToken = await loginApi(request, customerEmail, customerPassword)
    const inquiryId = await createInquiry(request, customerToken, customerEmail)

    const buyerOfferRes = await request.post(`${apiBase}/user/inquiries/${inquiryId}/negotiations`, {
      headers: { Authorization: `Bearer ${customerToken}` },
      data: { unitPrice: 1.5, quantity: 1000, totalAmount: 1500, currency: 'USD', message: 'Buyer opening offer' },
    })
    expect(buyerOfferRes.ok()).toBeTruthy()

    const adminToken = await loginApi(request, adminEmail, adminPassword)
    const adminCounterRes = await request.post(`${apiBase}/admin/inquiries/${inquiryId}/negotiations`, {
      headers: { Authorization: `Bearer ${adminToken}` },
      data: { unitPrice: 1.8, quantity: 1000, totalAmount: 1800, currency: 'USD', message: 'Admin counter-offer' },
    })
    expect(adminCounterRes.ok()).toBeTruthy()
    const adminOffer = await adminCounterRes.json()
    const adminOfferId = adminOffer.offer?.id as string
    expect(adminOfferId).toBeTruthy()

    const acceptRes = await request.post(
      `${apiBase}/user/inquiries/${inquiryId}/negotiations/${adminOfferId}/accept`,
      { headers: { Authorization: `Bearer ${customerToken}` }, data: {} },
    )
    expect(acceptRes.ok()).toBeTruthy()

    await page.goto('/auth/login')
    await page.fill('#email', adminEmail)
    await page.fill('#password', adminPassword)
    await page.click('button[type="submit"]')
    await page.waitForURL(/\/admin/, { timeout: 20_000 })
    await page.goto(`/admin/inquiries/${inquiryId}`)
    await expect(page.locator('body')).toContainText(/accepted|已接受|offer_accepted/i)
  })
})
