export default defineI18nLocaleDetector((event, config) => {
  // Detect locale from URL path prefix (e.g., /zh/about → 'zh')
  const path = event.path || ''
  const supportedLocales = (config.locales || []).map((l: any) => l.code || l)
  const segments = path.split('/')
  if (segments.length > 1) {
    const candidate = segments[1]
    if (supportedLocales.includes(candidate) && candidate !== config.defaultLocale) {
      return candidate
    }
  }
  return config.defaultLocale
})
