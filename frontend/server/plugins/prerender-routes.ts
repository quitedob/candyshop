// 构建时动态注入预渲染路由（9 语种静态页 + 分类/产品/博客）
const NON_DEFAULT_LOCALES = ['en', 'ko', 'ar', 'ja', 'th', 'vi', 'id', 'ms']

const STATIC_PATHS = [
  '/',
  '/about',
  '/faq',
  '/products',
  '/oem-solutions',
  '/factory-quality',
  '/privacy',
  '/terms',
  '/blog',
  '/cases-clients',
  '/legal/shipping',
  '/legal/returns',
  '/legal/cookies',
  '/legal/impressum',
  '/legal/data-processing',
  '/legal/translation-disclaimer',
]

/** 为路径生成所有语种变体（zh 无前缀） */
function localizedPaths(path: string): string[] {
  const routes = [path]
  for (const locale of NON_DEFAULT_LOCALES) {
    routes.push(path === '/' ? `/${locale}` : `/${locale}${path}`)
  }
  return routes
}

/** 为 CMS 路径生成所有语种变体 */
function localizedCmsPaths(path: string): string[] {
  const routes = [path]
  for (const locale of NON_DEFAULT_LOCALES) {
    routes.push(`/${locale}${path}`)
  }
  return routes
}

export default defineNitroPlugin((nitroApp) => {
  const hook = nitroApp.hooks.hook as unknown as (name: string, handler: (routes: Set<string>) => Promise<void>) => void
  hook('prerender:routes', async (routes: Set<string>) => {
    for (const path of STATIC_PATHS) {
      for (const localized of localizedPaths(path)) {
        routes.add(localized)
      }
    }

    const apiBase = process.env.INTERNAL_API_BASE || 'http://localhost:8080/api/v1'

    try {
      const [productsRes, postsRes, categoriesRes] = await Promise.all([
        $fetch<{ data?: Array<{ slug: string; categorySlug?: string; category?: string }> }>(
          `${apiBase}/public/products?limit=50&sort=created&order=desc`
        ).catch(() => ({ data: [] })),
        $fetch<{ data?: Array<{ slug: string }> }>(
          `${apiBase}/public/posts?limit=20&sort=created&order=desc`
        ).catch(() => ({ data: [] })),
        $fetch<Array<{ slug: string }>>(`${apiBase}/public/categories`).catch(() => []),
      ])

      const products = productsRes?.data || []
      const posts = postsRes?.data || []
      const categories = Array.isArray(categoriesRes) ? categoriesRes : []

      for (const cat of categories) {
        if (!cat.slug) continue
        for (const path of localizedCmsPaths(`/products/${cat.slug}`)) {
          routes.add(path)
        }
      }

      for (const product of products) {
        const cat = product.categorySlug || product.category
        if (!cat || !product.slug) continue
        for (const path of localizedCmsPaths(`/products/${cat}/${product.slug}`)) {
          routes.add(path)
        }
      }

      for (const post of posts) {
        if (!post.slug) continue
        for (const path of localizedCmsPaths(`/blog/${post.slug}`)) {
          routes.add(path)
        }
      }

      console.log(`[prerender-routes] Added ${routes.size} routes (static + CMS)`)
    } catch (err) {
      console.warn('[prerender-routes] CMS fetch failed, static routes only:', err)
    }
  })
})
