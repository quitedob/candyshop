export default defineI18nConfig(() => ({
  legacy: false,
  locale: 'zh',
  fallbackLocale: 'en',
  missing: (_locale: string, key: string) => key
}))
