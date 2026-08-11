import * as process from 'node:process'

// SWR/ISR 仅用于生产；开发模式走 Vite 实时 SSR，避免 stale dist 与 _payload.json 404
const isProduction = process.env.NODE_ENV === 'production'
const cacheRouteRules = isProduction
  ? {
      '/products/*': { swr: 3600 },
      '/products/*/*': { swr: 3600 },
      '/blog/**': { swr: 3600 },
      '/cases-clients/**': { swr: 3600 },
    }
  : {}

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  // R4-20: Devtools only in development
  devtools: { enabled: process.env.NODE_ENV !== 'production' },
  ssr: true,

  // Modules
  modules: [
    '@nuxtjs/seo',
    '@nuxtjs/i18n',
    '@nuxtjs/tailwindcss',
    '@nuxt/icon'
  ],

  // Hybrid rendering: marketing SSR/ISR, admin/customer CSR
  routeRules: {
    '/api/**': {
      proxy: {
        to: (process.env.BACKEND_URL || 'http://localhost:8080') + '/api/**',
        fetchOptions: { timeout: 660_000 },
      },
    },
    '/**': {
      headers: {
        'Content-Security-Policy': "default-src 'self'; script-src 'self' 'unsafe-inline'; worker-src 'self' blob:; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https: blob:; connect-src 'self' https:; object-src 'none'; base-uri 'self'; frame-src 'self' https://www.openstreetmap.org https://www.google.com; form-action 'self'; frame-ancestors 'none'"
      }
    },
    '/admin/**': { ssr: false },
    '/customer/**': { ssr: false },
    '/auth/**': { ssr: false },
    '/supplier/**': { redirect: '/' },
    '/': { prerender: true },
    '/about': { prerender: true },
    '/faq': { prerender: true },
    '/products': { prerender: true },
    '/oem-solutions': { prerender: true },
    '/factory-quality': { prerender: true },
    '/privacy': { prerender: true },
    '/terms': { prerender: true },
    '/legal/**': { prerender: true },
    '/blog': { prerender: true },
    '/cases-clients': { prerender: true },
    ...cacheRouteRules,
    '/sitemap.xml': { prerender: true },
    '/__og-image__/**': { prerender: false, index: false },
  },

  ogImage: {
    enabled: process.env.NUXT_OG_IMAGE !== 'false',
    defaults: {
      component: 'Default',
      width: 1200,
      height: 630,
    },
    compatibility: {
      prerender: { chromium: false, sharp: false },
    },
  },

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
        { name: 'theme-color', content: '#1c1b1b', media: '(prefers-color-scheme: dark)' }
      ],
      link: [
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: 'anonymous' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Noto+Serif:ital,wght@0,400;0,500;0,600;0,700;1,400&family=Noto+Serif+SC:wght@400;500;600;700&family=Work+Sans:wght@400;500;600;700&family=Noto+Sans+SC:wght@400;500;600;700&display=swap' }
      ],
      // 首屏前应用主题，避免闪烁；配合 useDarkMode cookie 逻辑
      script: [
        {
          innerHTML: `(function(){try{var m=document.cookie.match(/(?:^|;\\s*)theme=([^;]*)/);var t=m?decodeURIComponent(m[1]):'';var d=t==='dark'||(!t&&window.matchMedia('(prefers-color-scheme: dark)').matches);document.documentElement.classList.toggle('dark',d)}catch(e){}})();`,
          tagPosition: 'head',
        },
      ],
    },
    pageTransition: { name: 'page', mode: 'out-in' }
  },

  // CSS
  css: ['~/assets/css/main.css', '~/assets/css/admin-tokens.css'],

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
    // 禁止 lazy 模式预加载 fallback 语言包（避免缺失键时显示英语）
    fallbackLocale: false as const,
    baseUrl: process.env.SITE_URL || 'https://candypro-oem.com',
    detectBrowserLanguage: {
      useCookie: true,
      cookieKey: 'user-locale',
      alwaysRedirect: false,
      fallbackLocale: 'zh',
    },
    experimental: {
      strictSeo: true,
      localeDetector: './composables/useLocaleDetector.ts',
    },
    vueI18n: './i18n.config.ts'
  },

  // Icon：禁用 server bundle，避免 prerender 阶段 createRequire(file:///_entry.js) 崩溃
  icon: {
    provider: 'iconify',
    clientBundle: {
      scan: true,
    },
  },

  // Runtime config
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE_URL || '/api/v1',
      siteUrl: process.env.SITE_URL || 'https://candypro-oem.com',
      defaultOgImage: '/og-default.png',
      // No fake default: an unset number yields '' so no WhatsApp button links
      // to a placeholder phone. Set WHATSAPP_NUMBER in env.
      whatsappNumber: process.env.WHATSAPP_NUMBER || '',
      googleMapsApiKey: process.env.NUXT_PUBLIC_GOOGLE_MAPS_API_KEY || '',
      enableMultiWarehouse: process.env.NUXT_PUBLIC_ENABLE_MULTI_WAREHOUSE === 'true',
      jwtAccessMinutes: Number(process.env.NUXT_PUBLIC_JWT_ACCESS_MINUTES || 15),
    },
    // Server-only: used during SSR so $fetch bypasses Nitro localFetch (which ignores devProxy)
    internalApiBase: process.env.INTERNAL_API_BASE || 'http://localhost:8080/api/v1'
  },

  // Sitemap — i18n 多 sitemap 模式下顶层 sources 会被忽略，须写入各 locale sitemap
  sitemap: {
    sitemaps: {
      'en-US': { sources: ['/__sitemap__/cms-urls'], includeAppSources: true },
      'zh-CN': { sources: ['/__sitemap__/cms-urls'], includeAppSources: true },
      'ko-KR': { sources: ['/__sitemap__/cms-urls'], includeAppSources: true },
      'ar-SA': { sources: ['/__sitemap__/cms-urls'], includeAppSources: true },
      'ja-JP': { sources: ['/__sitemap__/cms-urls'], includeAppSources: true },
      'th-TH': { sources: ['/__sitemap__/cms-urls'], includeAppSources: true },
      'vi-VN': { sources: ['/__sitemap__/cms-urls'], includeAppSources: true },
      'id-ID': { sources: ['/__sitemap__/cms-urls'], includeAppSources: true },
      'ms-MY': { sources: ['/__sitemap__/cms-urls'], includeAppSources: true },
    },
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
      '/contact',
      '/api/',
    ],
    allow: [
      '/',
      '/products/**',
      '/blog/**',
      '/faq',
      '/about',
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
        target: process.env.API_PROXY_TARGET || 'http://localhost:8080/api',
        changeOrigin: true
      }
    },
    prerender: {
      failOnError: false,
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
      // 预打包核心依赖，减少 dev 模式下 optimizeDeps 变更触发的 HMR 双 Vue 实例问题
      include: [
        'vue',
        'vue-router',
        'vue-chartjs',
        'chart.js',
      ],
      exclude: ['unenv', 'nuxt-og-image']
    },
    resolve: {
      // 避免 Vite 热更新后加载多份 Vue，导致 inject(route) / ref.value 报错
      dedupe: [
        'vue',
        'vue-router',
        '@vue/runtime-core',
        '@vue/runtime-dom',
        '@vue/reactivity',
      ],
      alias: {
        // Fix unenv v2 double-runtime path issue (unenv/dist/runtime/runtime/mock/empty.mjs)
        'unenv/runtime/mock/empty.mjs': 'unenv/dist/runtime/mock/empty.mjs'
      }
    }
  },

  // Compatibility
  compatibilityDate: '2026-03-14'
})
