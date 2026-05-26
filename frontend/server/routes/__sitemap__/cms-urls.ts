// server/routes/__sitemap__/cms-urls.ts — CMS 动态 URL（避开 /api/** 代理）
export default defineSitemapEventHandler(async () => {
  const config = useRuntimeConfig()
  const base = config.internalApiBase as string

  try {
    const [products, posts, categories] = await Promise.all([
      $fetch<{ data?: Array<{ slug: string; categorySlug?: string; category?: string; updatedAt?: string }> }>(
        `${base}/public/products?limit=10000`
      ).catch(() => ({ data: [] })),
      $fetch<{ data?: Array<{ slug: string; updatedAt?: string; publishedAt?: string }> }>(
        `${base}/public/posts?limit=10000`
      ).catch(() => ({ data: [] })),
      $fetch<Array<{ slug: string; updatedAt?: string }>>(`${base}/public/categories`).catch(() => []),
    ])

    const productUrls = (products?.data || []).map((p) => ({
      loc: `/products/${p.categorySlug || p.category}/${p.slug}`,
      lastmod: p.updatedAt,
    }))

    const postUrls = (posts?.data || []).map((p) => ({
      loc: `/blog/${p.slug}`,
      lastmod: p.updatedAt || p.publishedAt,
    }))

    const categoryUrls = (Array.isArray(categories) ? categories : []).map((c) => ({
      loc: `/products/${c.slug}`,
      lastmod: c.updatedAt,
    }))

    return [...categoryUrls, ...productUrls, ...postUrls]
  } catch {
    return []
  }
})
