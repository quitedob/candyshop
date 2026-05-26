/**
 * usePageOgImage — OG 图片优先级：管理员 ogImage > thumbnail > 动态模板
 */
import { toValue, type MaybeRef } from '#imports'

interface PageOgImageOptions {
  title: MaybeRef<string>
  description?: MaybeRef<string>
  ogImage?: MaybeRef<string | undefined>
  thumbnail?: MaybeRef<string | undefined>
  images?: MaybeRef<string[] | undefined>
  ogType?: 'website' | 'product' | 'article'
  schema?: Record<string, unknown>
  noindex?: boolean
  twitterCard?: 'summary' | 'summary_large_image'
}

/** 解析为绝对 URL */
function resolveAbsoluteUrl(path: string, siteUrl: string): string {
  if (!path) return ''
  return path.startsWith('http') ? path : new URL(path, siteUrl).href
}

export function usePageOgImage(options: PageOgImageOptions) {
  const config = useRuntimeConfig()
  const siteUrl = config.public.siteUrl || 'https://candypro-oem.com'

  const title = toValue(options.title)
  const description = toValue(options.description)
  const customOg = toValue(options.ogImage)
  const thumb = toValue(options.thumbnail)
  const imgs = toValue(options.images)
  const entityImage = customOg || thumb || imgs?.[0]

  let seoOgImage: string | undefined

  if (entityImage) {
    seoOgImage = resolveAbsoluteUrl(entityImage, siteUrl)
  } else {
    // prerender 阶段跳过 defineOgImageComponent，避免 Nitro createRequire 崩溃
    if (!import.meta.prerender) {
      defineOgImageComponent('Default', {
        title,
        description: description || '',
      })
    }
    const defaultPath = config.public.defaultOgImage || '/og-default.png'
    seoOgImage = resolveAbsoluteUrl(defaultPath, siteUrl)
  }

  useSeo({
    title,
    description,
    ogImage: seoOgImage,
    ogType: options.ogType,
    schema: options.schema,
    noindex: options.noindex,
    twitterCard: options.twitterCard,
  })
}
