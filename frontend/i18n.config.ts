export default defineI18nConfig(() => ({
  legacy: true,
  locale: 'zh',
  fallbackLocale: 'en',
  missing: (_locale: string, key: string) => key
}))
