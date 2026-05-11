import * as process from 'node:process'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  // R4-20: Devtools only in development
  devtools: { enabled: process.env.NODE_ENV !== 'production' },

  // Modules
  modules: [
    '@nuxtjs/seo',
    '@nuxtjs/i18n',
    '@nuxtjs/tailwindcss',
    '@nuxt/icon'
  ],

  // Site config for @nuxtjs/seo
  site: {
    url: process.env.SITE_URL || 'https://candypro-oem.com',
    name: 'CandyPro — Professional OEM Candy Manufacturer',
    defaultLocale: 'zh',
  },

  // Components Configuration
  components: [
    {
      path: '~/components',
      pathPrefix: false,
    },
  ],

  // App config
  app: {
    head: {
      // R4-20: lang is set dynamically by i18n — don't hardcode 'en'
      // htmlAttrs.lang is managed by @nuxtjs/i18n automatically
      htmlAttrs: {
        style: 'color-scheme: light dark;'
      },
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1, viewport-fit=cover' },
        { name: 'format-detection', content: 'telephone=no' },
        { name: 'theme-color', content: '#ffffff', media: '(prefers-color-scheme: light)' },
        { name: 'theme-color', content: '#1c1917', media: '(prefers-color-scheme: dark)' }
      ],
      link: [
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Outfit:wght@500;600;700;800&family=Noto+Sans+SC:wght@400;500;600;700&display=swap' }
      ]
    },
    pageTransition: { name: 'page', mode: 'out-in' }
  },

  // CSS
  css: ['~/assets/css/main.css'],

  // i18n Configuration
  i18n: {
    locales: [
      {
        code: 'en',
        iso: 'en-US',
        files: [
          'en/common.json',
          'en/nav.json',
          'en/home.json',
          'en/products.json',
          'en/factory.json',
          'en/oem.json',
          'en/cases.json',
          'en/blog.json',
          'en/form.json',
          'en/auth.json',
          'en/admin.json',
          'en/customer.json',
          'en/legal.json',
          'en/seo.json',
          'en/about.json'
        ],
        name: 'English'
      },
      {
        code: 'zh',
        iso: 'zh-CN',
        files: [
          'zh/common.json',
          'zh/nav.json',
          'zh/home.json',
          'zh/products.json',
          'zh/factory.json',
          'zh/oem.json',
          'zh/cases.json',
          'zh/blog.json',
          'zh/form.json',
          'zh/auth.json',
          'zh/admin.json',
          'zh/customer.json',
          'zh/legal.json',
          'zh/seo.json',
          'zh/about.json'
        ],
        name: '中文'
      }
    ],
    lazy: false,
    langDir: '.',
    defaultLocale: 'zh',
    strategy: 'prefix_except_default',
    seo: false,
    detectBrowserLanguage: false,
    experimental: {
      localeDetector: './composables/useLocaleDetector.ts'
    },
    vueI18n: './i18n.config.ts'
  },

  // Runtime config
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE_URL || '/api/v1',
      siteUrl: process.env.SITE_URL || 'https://candypro-oem.com',
      whatsappNumber: process.env.WHATSAPP_NUMBER || '1234567890'
    },
    // Server-only: used during SSR so $fetch bypasses Nitro localFetch (which ignores devProxy)
    internalApiBase: 'http://localhost:8080/api/v1'
  },

  // Sitemap
  sitemap: {
    sources: ['/api/__sitemap__/urls'],
    exclude: [
      '/admin/**',
      '/customer/**',
      '/auth/**',
      '/login',
      '/register',
      '/forgot-password',
      '/reset-password',
    ],
    autoLastmod: true,
  },

  // Robots.txt (via nuxt-simple-robots, bundled in @nuxtjs/seo)
  robots: {
    disallow: [
      '/admin/',
      '/customer/',
      '/auth/',
    ],
    allow: [
      '/',
      '/products/**',
      '/blog/**',
      '/faq',
      '/about',
      '/contact',
      '/oem-solutions',
      '/factory-quality',
      '/cases-clients/**',
    ],
  },

  // Nitro server
  nitro: {
    port: 3000,
    host: '0.0.0.0',
    devProxy: {
      '/api': {
        target: 'http://localhost:8080/api',
        changeOrigin: true
      }
    },
    prerender: {
      routes: ['/', '/about', '/faq', '/contact', '/factory-quality', '/oem-solutions', '/privacy', '/terms', '/products', '/blog'],
    },
    alias: {
      // Fix nuxt-og-image unenv v2 incompatibility — the module references
      // "unenv/runtime/mock/*" which resolves to dist/runtime/runtime/mock/* (double)
      // under unenv v2's wildcard export map. Correct to the canonical paths.
      'unenv/runtime/mock/empty': 'unenv/mock/empty',
      'unenv/runtime/mock/proxy-cjs': 'unenv/mock/proxy-cjs',
    },
  },

  // Hybrid rendering strategy
  routeRules: {
    // SSG — public content, max crawl budget
    '/': { prerender: true },
    '/about': { prerender: true },
    '/faq': { prerender: true },
    '/contact': { prerender: true },
    '/factory-quality': { prerender: true },
    '/oem-solutions': { prerender: true },
    '/privacy': { prerender: true },
    '/terms': { prerender: true },
    '/blog/**': { prerender: true },
    '/products/**': { prerender: true },
    // CSR only — authenticated, zero SEO value
    '/admin/**': { ssr: false },
    '/customer/**': { ssr: false },
    '/auth/**': { ssr: false },
    // SSR + ISR — dynamic content refreshed hourly
    '/cases-clients/**': { swr: 3600 },
  },

  // Disable appManifest to fix Vite pre-transform error with #app-manifest import
  experimental: {
    appManifest: false
  },

  // Vite config
  vite: {
    optimizeDeps: {
      include: [],
      exclude: ['unenv', 'nuxt-og-image']
    },
    resolve: {
      alias: {
        // Fix unenv v2 double-runtime path issue (unenv/dist/runtime/runtime/mock/empty.mjs)
        'unenv/runtime/mock/empty.mjs': 'unenv/dist/runtime/mock/empty.mjs'
      }
    }
  },

  // Compatibility
  compatibilityDate: '2026-03-14'
})
