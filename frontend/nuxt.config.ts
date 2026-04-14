import * as process from 'node:process'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  // R4-20: Devtools only in development
  devtools: { enabled: process.env.NODE_ENV !== 'production' },

  // Modules
  modules: [
    '@nuxtjs/i18n',
    '@nuxtjs/tailwindcss',
    '@nuxtjs/sitemap',
    '@nuxt/icon'
  ],

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
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'format-detection', content: 'telephone=no' }
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
    lazy: true,
    langDir: 'locales',
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
      apiBase: process.env.API_BASE_URL || 'http://localhost:8080/api/v1',
      siteUrl: process.env.SITE_URL || 'https://candypro-oem.com',
      whatsappNumber: process.env.WHATSAPP_NUMBER || '1234567890'
    }
  },

  // R4-20: Remove broken sitemap source — /api/__sitemap__/urls doesn't exist
  // sitemap: {
  //   sources: ['/api/__sitemap__/urls']
  // },

  // Nitro server
  nitro: {
    port: 3000,
    host: '0.0.0.0'
  },

  // Disable appManifest to fix Vite pre-transform error with #app-manifest import
  experimental: {
    appManifest: false
  },

  // Vite config
  vite: {
    optimizeDeps: {
      include: []
    }
  },

  // Compatibility
  compatibilityDate: '2026-03-14'
})
