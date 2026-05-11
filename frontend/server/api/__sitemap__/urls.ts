// Returns CMS-driven URLs for products, categories, and published blog posts
export default defineEventHandler(async () => {
  try {
    const [products, posts, categories] = await Promise.all([
      $fetch('/api/v1/products?limit=10000').catch(() => ({ data: [] })),
      $fetch('/api/v1/posts?limit=10000').catch(() => ({ data: [] })),
      $fetch('/api/v1/categories').catch(() => []),
    ])

    const productUrls = ((products as any).data || []).map((p: any) => ({
      loc: `/products/${p.categorySlug || p.category}/${p.slug}`,
      lastmod: p.updatedAt || p.updated_at,
    }))

    const postUrls = ((posts as any).data || []).map((p: any) => ({
      loc: `/blog/${p.slug}`,
      lastmod: p.updatedAt || p.updated_at || p.publishedAt || p.published_at,
    }))

    const categoryUrls = (categories as any[]).map((c: any) => ({
      loc: `/products/${c.slug}`,
      lastmod: c.updatedAt || c.updated_at,
    }))

    return [...productUrls, ...postUrls, ...categoryUrls]
  } catch {
    return []
  }
})
