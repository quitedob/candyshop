export default defineI18nConfig(() => ({
  legacy: false,
  globalInjection: true,
  locale: 'zh',
  fallbackLocale: 'en',
  missingWarn: process.env.NODE_ENV !== 'production',
  fallbackWarn: process.env.NODE_ENV !== 'production',
  availableLocales: ['en', 'zh', 'ko', 'ar', 'ja', 'th', 'vi', 'id', 'ms'],
  modifiers: {
    kebabCase: (str) => typeof str === 'string' ? str.split(' ').join('-').toLowerCase() : str
  },
  numberFormats: {
    en: {
      currency: { style: 'currency', currency: 'USD', notation: 'standard' }
    },
    zh: {
      currency: { style: 'currency', currency: 'CNY', notation: 'standard' }
    },
    ko: {
      currency: { style: 'currency', currency: 'KRW', notation: 'standard' }
    },
    ar: {
      currency: { style: 'currency', currency: 'AED', notation: 'standard' }
    },
    ja: {
      currency: { style: 'currency', currency: 'JPY', notation: 'standard' }
    },
    th: {
      currency: { style: 'currency', currency: 'THB', notation: 'standard' }
    },
    vi: {
      currency: { style: 'currency', currency: 'VND', notation: 'standard' }
    },
    id: {
      currency: { style: 'currency', currency: 'IDR', notation: 'standard' }
    },
    ms: {
      currency: { style: 'currency', currency: 'MYR', notation: 'standard' }
    }
  }
}))
