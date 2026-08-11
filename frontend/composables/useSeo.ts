/**
 * SEO Composable
 * Provides SEO helpers with structured data support
 */
import { toValue, type MaybeRef } from '#imports'

interface SEOOptions {
  title?: MaybeRef<string>
  description?: MaybeRef<string>
  ogImage?: MaybeRef<string>
  ogType?: 'website' | 'product' | 'article'
  canonical?: string
  noindex?: boolean
  schema?: Record<string, unknown>
  twitterCard?: 'summary' | 'summary_large_image'
}

interface BreadcrumbItem {
  name: string
  item?: string
}

interface ProductSchema {
  name: string
  description: string
  image: string[]
  brand?: string
  manufacturer?: string
  category?: string
  offers?: {
    priceCurrency?: string
    availability?: string
  }
}

interface ArticleSchema {
  headline: string
  description: string
  image: string[]
  author: string
  publishedTime: string
  modifiedTime?: string
  publisher?: {
    name: string
    logo?: string
  }
}

interface OrganizationSchema {
  name: string
  url: string
  logo: string
  description: string
  address?: {
    streetAddress: string
    addressLocality: string
    addressCountry: string
  }
  contactPoint?: {
    contactType: string
    telephone: string
    email: string
  }
  sameAs?: string[]
}

export const useSeo = (options: SEOOptions = {}) => {
  const { t } = useI18n()
  const config = useRuntimeConfig()
  const route = useRoute()

  const siteUrl = config.public.siteUrl || 'https://candyfactory.example.com'
  const defaultTitle = t('seo.default_title')
  const defaultDescription = t('seo.default_description')

  const title = toValue(options.title) || defaultTitle
  const description = toValue(options.description) || defaultDescription
  const ogImage = toValue(options.ogImage)
  const resolvedOgImage = ogImage
    ? (ogImage.startsWith('http') ? ogImage : new URL(ogImage, siteUrl).href)
    : ''

  const pageUrl = new URL(route.path, siteUrl).href

  useHead({
    title,
    meta: [
      { name: 'description', content: description },
      { property: 'og:title', content: title },
      { property: 'og:description', content: description },
      { property: 'og:type', content: options.ogType || 'website' },
      { property: 'og:url', content: options.canonical || pageUrl },
      { property: 'og:site_name', content: 'CandyPro OEM' },
      ...(resolvedOgImage ? [{ property: 'og:image', content: resolvedOgImage }] : []),
      { name: 'twitter:card', content: options.twitterCard || 'summary_large_image' },
      { name: 'twitter:title', content: title },
      { name: 'twitter:description', content: description },
      ...(resolvedOgImage ? [{ name: 'twitter:image', content: resolvedOgImage }] : []),
      ...(options.noindex ? [{ name: 'robots', content: 'noindex, nofollow' }] : []),
    ],
    ...(options.schema ? {
      script: [{
        type: 'application/ld+json',
        innerHTML: JSON.stringify(options.schema),
        tagPosition: 'head',
      }],
    } : {}),
  })
}

/**
 * Create product structured data
 */
export const useProductSchema = (product: {
  name: string
  slug: string
  description: string
  images: string[]
  category: string
  oemAvailable?: boolean
}) => {
  const config = useRuntimeConfig()
  const siteUrl = config.public.siteUrl

  return {
    '@context': 'https://schema.org',
    '@type': 'Product',
    name: product.name,
    description: product.description,
    image: product.images,
    category: product.category,
    url: new URL(`/products/${product.slug}`, siteUrl).href,
    offers: {
      '@type': 'AggregateOffer',
      priceCurrency: 'USD',
      availability: 'https://schema.org/InStock',
      seller: {
        '@type': 'Organization',
        name: 'CandyPro OEM'
      }
    }
  }
}

/**
 * Create breadcrumb structured data
 */
export const useBreadcrumbSchema = (items: BreadcrumbItem[]) => {
  return {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: items.map((item, index) => ({
      '@type': 'ListItem',
      position: index + 1,
      name: item.name,
      ...(item.item ? { item } : {})
    }))
  }
}

/**
 * Create article/blog structured data
 */
export const useArticleSchema = (article: {
  title: string
  slug: string
  excerpt: string
  thumbnail?: string
  publishedAt: string
  updatedAt?: string
  author?: { name: string }
  category?: string
}) => {
  const { t } = useI18n()
  const config = useRuntimeConfig()
  const siteUrl = config.public.siteUrl

  return {
    '@context': 'https://schema.org',
    '@type': 'Article',
    headline: article.title,
    description: article.excerpt,
    image: article.thumbnail ? [article.thumbnail] : [],
    author: {
      '@type': 'Person',
      name: article.author?.name || t('seo.default_author')
    },
    publisher: {
      '@type': 'Organization',
      name: 'CandyPro OEM',
      logo: {
        '@type': 'ImageObject',
        url: new URL('/logo.png', siteUrl).href
      }
    },
    datePublished: article.publishedAt,
    ...(article.updatedAt ? { dateModified: article.updatedAt } : {}),
    mainEntityOfPage: {
      '@type': 'WebPage',
      '@id': new URL(`/blog/${article.slug}`, siteUrl).href
    }
  }
}

/**
 * Create organization structured data
 */
export const useOrganizationSchema = () => {
  const config = useRuntimeConfig()
  const siteUrl = config.public.siteUrl

  return {
    '@context': 'https://schema.org',
    '@type': 'Organization',
    name: 'CandyPro OEM',
    url: siteUrl,
    logo: new URL('/logo.png', siteUrl).href,
    description: 'Professional OEM candy manufacturer providing private label gummy, hard candy, aerated candy, toffee & compound chocolate for global brands.',
    address: {
      '@type': 'PostalAddress',
      streetAddress: '123 Industrial Zone',
      addressLocality: 'Sweet City',
      addressCountry: 'CN'
    },
    contactPoint: {
      '@type': 'ContactPoint',
      contactType: 'sales',
      telephone: '+86-123-456-7890',
      email: 'info@candypro.com',
      availableLanguage: ['English', 'Chinese']
    },
    sameAs: [
      'https://facebook.com/candypro',
      'https://linkedin.com/company/candypro',
      'https://instagram.com/candypro'
    ]
  }
}

/**
 * Create local business structured data
 */
export const useLocalBusinessSchema = () => {
  const config = useRuntimeConfig()
  const siteUrl = config.public.siteUrl

  return {
    '@context': 'https://schema.org',
    '@type': 'FoodEstablishment',
    name: 'CandyPro OEM Manufacturing',
    image: new URL('/factory.jpg', siteUrl).href,
    url: siteUrl,
    // Real company data — keep in sync with pages/contact.vue + Footer. The
    // previous fabricated telephone/address ("123 Industrial Zone", "Sweet City")
    // contradicted the contact page and misled search engines.
    telephone: '+86 21 6731 0088',
    email: 'sales@candypro.com',
    address: {
      '@type': 'PostalAddress',
      streetAddress: 'No. 88 Shipin Road, Jinshan District',
      addressLocality: 'Shanghai',
      addressRegion: 'SH',
      postalCode: '201500',
      addressCountry: 'CN'
    },
    geo: {
      '@type': 'GeoCoordinates',
      latitude: 31.2304,
      longitude: 121.4737
    },
    openingHoursSpecification: {
      '@type': 'OpeningHoursSpecification',
      dayOfWeek: [
        'Monday',
        'Tuesday',
        'Wednesday',
        'Thursday',
        'Friday'
      ],
      opens: '09:00',
      closes: '18:00'
    },
    priceRange: '$$'
  }
}

/**
 * Create FAQ structured data
 */
export const useFAQSchema = (faqs: Array<{
  question: string
  answer: string
}>) => {
  return {
    '@context': 'https://schema.org',
    '@type': 'FAQPage',
    mainEntity: faqs.map(faq => ({
      '@type': 'Question',
      name: faq.question,
      acceptedAnswer: {
        '@type': 'Answer',
        text: faq.answer
      }
    }))
  }
}

/**
 * Create video structured data
 */
export const useVideoSchema = (video: {
  name: string
  description: string
  thumbnailUrl: string
  uploadDate: string
  duration?: string
  embedUrl?: string
}) => {
  return {
    '@context': 'https://schema.org',
    '@type': 'VideoObject',
    name: video.name,
    description: video.description,
    thumbnailUrl: video.thumbnailUrl,
    uploadDate: video.uploadDate,
    ...(video.duration ? { duration: video.duration } : {}),
    ...(video.embedUrl ? { embedUrl: video.embedUrl } : {})
  }
}

/**
 * Combine multiple schema types
 */
export const useCombinedSchema = (schemas: Record<string, unknown>[]) => {
  return schemas
}

/**
 * Set up global organization schema (call once in app.vue)
 */
export const useGlobalSchema = () => {
  const orgSchema = useOrganizationSchema()
  const localSchema = useLocalBusinessSchema()

  useHead({
    script: [
      {
        type: 'application/ld+json',
        innerHTML: JSON.stringify([orgSchema, localSchema]),
        tagPosition: 'head'
      }
    ]
  })
}

/**
 * Helper to generate page title with template
 */
export const usePageTitle = (title: string) => {
  const { t } = useI18n()
  const template = t('nav.home')

  return `${title} | ${template}`
}
