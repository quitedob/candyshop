import { createServer, type Server } from 'node:http'
import { test, expect } from '@playwright/test'
import { addPrerenderRoutes, createCmsRouteRules, resolveInternalApiBase } from '../build/prerender-routes'

test('build API resolution matches runtime and deployment overrides', () => {
  expect(resolveInternalApiBase({})).toBe('http://localhost:8080/api/v1')
  expect(resolveInternalApiBase({ BACKEND_URL: 'http://api:8080///' })).toBe('http://api:8080/api/v1')
  expect(resolveInternalApiBase({ BACKEND_URL: 'http://api:8080', INTERNAL_API_BASE: 'http://build-api/api/v1/' }))
    .toBe('http://build-api/api/v1')
  expect(resolveInternalApiBase({ INTERNAL_API_BASE: 'http://build-api/api/v1', NUXT_INTERNAL_API_BASE: 'http://runtime-api/api/v1/' }))
    .toBe('http://runtime-api/api/v1')
})

test('offline image builds prerender static locales while CMS pages stay dynamic', async () => {
  const routes = new Set<string>(['/existing-route'])
  await addPrerenderRoutes(routes, 'http://unreachable.example.invalid/api/v1')
  expect(routes.has('/existing-route')).toBe(true)
  for (const prefix of ['', '/en', '/ko', '/ar', '/ja', '/th', '/vi', '/id', '/ms']) {
    expect(routes.has(`${prefix}/about`)).toBe(true)
    expect(routes.has(`${prefix}/legal/cookies`)).toBe(true)
    expect(routes.has(`${prefix}/blog`)).toBe(false)
    expect(routes.has(`${prefix}/products`)).toBe(false)
  }
  const routeRules = createCmsRouteRules(false, true)
  expect(routeRules['/blog/**']).toEqual({ prerender: false, swr: 3600 })
  expect(routeRules['/en/products/**']).toEqual({ prerender: false, swr: 3600 })
  expect(routeRules['/']).toEqual({ prerender: false, swr: 3600 })
  expect(createCmsRouteRules(false, false)['/en/blog/**']).toEqual({ prerender: false })
  expect(createCmsRouteRules(true, true)['/en/blog/**']).toEqual({ prerender: true })
})

test('CMS opt-in enumerates every API page and every locale', async () => {
  const requestedPages: string[] = []
  const fixtureServer = createServer((request, response) => {
    const requestUrl = new URL(request.url || '/', 'http://fixture.invalid')
    requestedPages.push(`${requestUrl.pathname}${requestUrl.search}`)
    response.setHeader('Content-Type', 'application/json')
    if (requestUrl.pathname.endsWith('/categories')) {
      response.end(JSON.stringify([{ slug: 'gummy-candy' }]))
      return
    }
    const currentPage = Number(requestUrl.searchParams.get('page'))
    const isProductCollection = requestUrl.pathname.endsWith('/products')
    const records = currentPage === 1
      ? Array.from({ length: 50 }, (_, index) => ({ slug: `record-${index + 1}`, categorySlug: 'gummy-candy' }))
      : [{ slug: isProductCollection ? 'last-product' : 'last-post', categorySlug: 'gummy-candy' }]
    response.end(JSON.stringify({ data: records, pagination: { totalPages: 2 } }))
  })
  await new Promise<void>(resolve => fixtureServer.listen(0, '127.0.0.1', resolve))
  const address = fixtureServer.address()
  if (!address || typeof address === 'string') throw new Error('Fixture server did not bind a TCP port')
  try {
    const routes = new Set<string>()
    await addPrerenderRoutes(routes, `http://127.0.0.1:${address.port}/api/v1`, true)
    for (const prefix of ['', '/en', '/ko', '/ar', '/ja', '/th', '/vi', '/id', '/ms']) {
      expect(routes.has(`${prefix}/blog/last-post`)).toBe(true)
      expect(routes.has(`${prefix}/products/gummy-candy/last-product`)).toBe(true)
      expect(routes.has(`${prefix}/products/gummy-candy`)).toBe(true)
      expect(routes.has(`${prefix}/blog`)).toBe(true)
    }
    expect(requestedPages.filter(path => path.includes('page=2'))).toHaveLength(2)
    expect(requestedPages.filter(path => path.includes('limit=50'))).toHaveLength(4)
  } finally {
    await closeFixtureServer(fixtureServer)
  }
})

test('CMS opt-in fails visibly when the build API is unavailable', async () => {
  const fixtureServer = createServer((_request, response) => {
    response.writeHead(503, { 'Content-Type': 'application/json' })
    response.end(JSON.stringify({ message: 'Service unavailable' }))
  })
  await new Promise<void>(resolve => fixtureServer.listen(0, '127.0.0.1', resolve))
  const address = fixtureServer.address()
  if (!address || typeof address === 'string') throw new Error('Fixture server did not bind a TCP port')
  try {
    await expect(addPrerenderRoutes(new Set(), `http://127.0.0.1:${address.port}/api/v1`, true)).rejects.toThrow('503')
  } finally {
    await closeFixtureServer(fixtureServer)
  }
})

async function closeFixtureServer(server: Server): Promise<void> {
  server.closeAllConnections()
  await new Promise<void>((resolve, reject) => server.close(error => error ? reject(error) : resolve()))
}
