export default defineI18nConfig(() => ({
  legacy: false,
  globalInjection: true,
  locale: 'zh',
  fallbackLocale: 'en',
  availableLocales: ['en', 'zh'],
  modifiers: {
    kebabCase: (str) => typeof str === 'string' ? str.split(' ').join('-').toLowerCase() : str
  },
  numberFormats: {
    en: {
      currency: {
        style: 'currency',
        currency: 'USD',
        notation: 'standard'
      }
    },
    zh: {
      currency: {
        style: 'currency',
        currency: 'CNY',
        notation: 'standard'
      }
    }
  }
}))
