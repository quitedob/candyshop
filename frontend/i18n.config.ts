import type { VueMessageType } from 'vue-i18n'

export default defineI18nConfig(() => ({
  legacy: false,
  globalInjection: true,
  locale: 'zh',
  fallbackLocale: false,
  missingWarn: false,
  fallbackWarn: false,
  availableLocales: ['en', 'zh', 'ko', 'ar', 'ja', 'th', 'vi', 'id', 'ms'],
  modifiers: {
    kebabCase: (str: VueMessageType) => typeof str === 'string' ? str.split(' ').join('-').toLowerCase() : str
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
  },
  missing: (_locale: string, key: string) => {
    if (import.meta.dev) {
      console.warn(`[i18n] Missing translation: ${key}`)
    }
    // R2 D-2: previously this returned `key.split('.').pop()` — the trailing
    // segment of the dotted key. That surfaced raw snake_case like
    // "time_just_now" or "activity_inquiry" in the UI when a translation was
    // missing. In production we'd rather render an empty string than a
    // half-readable identifier; in dev we keep the key visible to make the
    // gap obvious during testing.
    return import.meta.dev ? key : ''
  }
}))
