import * as process from 'node:process'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  // R4-20: Devtools only in development
  devtools: { enabled: process.env.NODE_ENV !== 'production' },
  ssr: false,

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
      },
      {
        code: 'ko',
        iso: 'ko-KR',
        files: [
          'ko/common.json', 'ko/nav.json', 'ko/home.json', 'ko/products.json',
          'ko/factory.json', 'ko/oem.json', 'ko/cases.json', 'ko/blog.json',
          'ko/form.json', 'ko/auth.json', 'ko/admin.json', 'ko/customer.json',
          'ko/legal.json', 'ko/seo.json', 'ko/about.json'
        ],
        name: '한국어'
      },
      {
        code: 'ar',
        iso: 'ar-SA',
        dir: 'rtl',
        files: [
          'ar/common.json', 'ar/nav.json', 'ar/home.json', 'ar/products.json',
          'ar/factory.json', 'ar/oem.json', 'ar/cases.json', 'ar/blog.json',
          'ar/form.json', 'ar/auth.json', 'ar/admin.json', 'ar/customer.json',
          'ar/legal.json', 'ar/seo.json', 'ar/about.json'
        ],
        name: 'العربية'
      },
      {
        code: 'ja',
        iso: 'ja-JP',
        files: [
          'ja/common.json', 'ja/nav.json', 'ja/home.json', 'ja/products.json',
          'ja/factory.json', 'ja/oem.json', 'ja/cases.json', 'ja/blog.json',
          'ja/form.json', 'ja/auth.json', 'ja/admin.json', 'ja/customer.json',
          'ja/legal.json', 'ja/seo.json', 'ja/about.json'
        ],
        name: '日本語'
      },
      {
        code: 'th',
        iso: 'th-TH',
        files: [
          'th/common.json', 'th/nav.json', 'th/home.json', 'th/products.json',
          'th/factory.json', 'th/oem.json', 'th/cases.json', 'th/blog.json',
          'th/form.json', 'th/auth.json', 'th/admin.json', 'th/customer.json',
          'th/legal.json', 'th/seo.json', 'th/about.json'
        ],
        name: 'ไทย'
      },
      {
        code: 'vi',
        iso: 'vi-VN',
        files: [
          'vi/common.json', 'vi/nav.json', 'vi/home.json', 'vi/products.json',
          'vi/factory.json', 'vi/oem.json', 'vi/cases.json', 'vi/blog.json',
          'vi/form.json', 'vi/auth.json', 'vi/admin.json', 'vi/customer.json',
          'vi/legal.json', 'vi/seo.json', 'vi/about.json'
        ],
        name: 'Tiếng Việt'
      },
      {
        code: 'id',
        iso: 'id-ID',
        files: [
          'id/common.json', 'id/nav.json', 'id/home.json', 'id/products.json',
          'id/factory.json', 'id/oem.json', 'id/cases.json', 'id/blog.json',
          'id/form.json', 'id/auth.json', 'id/admin.json', 'id/customer.json',
          'id/legal.json', 'id/seo.json', 'id/about.json'
        ],
        name: 'Bahasa Indonesia'
      },
      {
        code: 'ms',
        iso: 'ms-MY',
        files: [
          'ms/common.json', 'ms/nav.json', 'ms/home.json', 'ms/products.json',
          'ms/factory.json', 'ms/oem.json', 'ms/cases.json', 'ms/blog.json',
          'ms/form.json', 'ms/auth.json', 'ms/admin.json', 'ms/customer.json',
          'ms/legal.json', 'ms/seo.json', 'ms/about.json'
        ],
        name: 'Bahasa Melayu'
      }
    ],
    lazy: true,
    langDir: 'i18n',
    restructureDir: false,
    defaultLocale: 'zh',
    strategy: 'prefix_except_default',
    seo: false,
    detectBrowserLanguage: false,
    experimental: {
      localeDetector: './composables/useLocaleDetector.ts'
    },
    vueI18n: './i18n.config.ts'
  },

  // Icon configuration (serve from local bundle, avoid CDN 404s)
  icon: {
    serverBundle: 'local'
  },

  // Runtime config
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE_URL || '/api/v1',
      siteUrl: process.env.SITE_URL || 'https://candypro-oem.com',
      whatsappNumber: process.env.WHATSAPP_NUMBER || '1234567890',
      googleMapsApiKey: process.env.NUXT_PUBLIC_GOOGLE_MAPS_API_KEY || ''
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
