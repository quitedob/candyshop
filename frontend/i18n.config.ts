export default defineI18nConfig(() => ({
  locale: 'zh',
  fallbackLocale: 'en',
  missing: (_locale: string, key: string) => {
    if (import.meta.dev) {
      console.warn(`[i18n] Missing translation: ${key}`)
    }
    return key
  }
}))
