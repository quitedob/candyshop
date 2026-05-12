export default defineI18nConfig(() => ({
  locale: 'zh',
  fallbackLocale: 'en',
  missing: (_locale: string, key: string) => key
}))
