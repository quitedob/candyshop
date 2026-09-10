import { $fetch } from 'ofetch'

const NON_DEFAULT_LOCALES = ['en', 'ko', 'ar', 'ja', 'th', 'vi', 'id', 'ms']

const STATIC_PATHS = [
  '/about',
  '/faq',
  '/oem-solutions',
  '/factory-quality',
  '/privacy',
  '/terms',
  '/legal/shipping',
  '/legal/returns',
  '/legal/cookies',
  '/legal/impressum',
  '/legal/data-processing',
  '/legal/translation-disclaimer',
]

const CMS_INDEX_PATHS = ['/', '/products', '/blog', '/cases-clients']
const CMS_CACHE_MAX_AGE_SECONDS = 3600

// Public product/post handlers cap pagination at 50 records per page.
const CMS_PAGE_SIZE = 50
const CMS_REQUEST_TIMEOUT_MS = 10_000
const DEFAULT_INTERNAL_API_BASE = 'http://localhost:8080/api/v1'

interface CmsRouteRecord {
  slug: string
  categorySlug?: string
  category?: string
}

interface CmsPage {
  data: CmsRouteRecord[] | null
  pagination?: { totalPages: number }
}

/** Match Nuxt's runtime override and the legacy deployment variables at build time. */
export function resolveInternalApiBase(environment: Record<string, string | undefined>): string {
  const explicitApiBase = environment.NUXT_INTERNAL_API_BASE || environment.INTERNAL_API_BASE
  if (explicitApiBase) return explicitApiBase.replace(/\/+$/, '')
  if (environment.BACKEND_URL) return `${environment.BACKEND_URL.replace(/\/+$/, '')}/api/v1`
  return DEFAULT_INTERNAL_API_BASE
}

/** Generate each supported locale; Chinese is served without a prefix. */
function localizedPaths(path: string): string[] {
  return [path, ...NON_DEFAULT_LOCALES.map(locale => path === '/' ? `/${locale}` : `/${locale}${path}`)]
}

/** API-backed pages stay dynamic unless the build explicitly opts into CMS snapshots. */
export function createCmsRouteRules(prerenderCms: boolean, isProduction: boolean) {
  const patterns = [...CMS_INDEX_PATHS, '/products/**', '/blog/**', '/cases-clients/**']
  return Object.fromEntries(patterns.flatMap(pattern => localizedPaths(pattern).map(path => [path, {
    prerender: prerenderCms,
    ...(isProduction && !prerenderCms ? { swr: CMS_CACHE_MAX_AGE_SECONDS } : {}),
  }])))
}

async function fetchCmsPages(apiBase: string, collection: 'products' | 'posts'): Promise<CmsRouteRecord[]> {
  const records: CmsRouteRecord[] = []
  let currentPage = 1
  let totalPages = 1
  do {
    const response = await $fetch<CmsPage>(`${apiBase}/public/${collection}`, {
      query: { page: currentPage, limit: CMS_PAGE_SIZE },
      timeout: CMS_REQUEST_TIMEOUT_MS,
      retry: 0,
    })
    if (!response || (response.data !== null && !Array.isArray(response.data))) {
      throw new Error(`Invalid ${collection} pagination response during prerender discovery`)
    }
    records.push(...(response.data || []))
    totalPages = Math.max(1, response.pagination?.totalPages || 1)
    currentPage++
  } while (currentPage <= totalPages)
  return records
}

/** This hook belongs to Nitro's build instance, before its prerender runtime starts. */
export async function addPrerenderRoutes(routes: Set<string>, apiBase: string, prerenderCms = false): Promise<void> {
  for (const path of STATIC_PATHS) {
    for (const localized of localizedPaths(path)) routes.add(localized)
  }

  if (!prerenderCms) return
  for (const path of CMS_INDEX_PATHS) {
    for (const localized of localizedPaths(path)) routes.add(localized)
  }

  // Do not silently publish a partial CMS route set when the build API is unavailable.
  const [products, posts, categories] = await Promise.all([
    fetchCmsPages(apiBase, 'products'),
    fetchCmsPages(apiBase, 'posts'),
    $fetch<CmsRouteRecord[]>(`${apiBase}/public/categories`, { timeout: CMS_REQUEST_TIMEOUT_MS, retry: 0 }),
  ])
  if (!Array.isArray(categories)) throw new Error('Invalid categories response during prerender discovery')

  for (const category of categories) {
    if (!category.slug) continue
    for (const path of localizedPaths(`/products/${category.slug}`)) routes.add(path)
  }
  for (const product of products) {
    const categorySlug = product.categorySlug || product.category
    if (!categorySlug || !product.slug) continue
    for (const path of localizedPaths(`/products/${categorySlug}/${product.slug}`)) routes.add(path)
  }
  for (const post of posts) {
    if (!post.slug) continue
    for (const path of localizedPaths(`/blog/${post.slug}`)) routes.add(path)
  }
}
