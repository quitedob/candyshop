export default defineI18nLocaleDetector((event, config) => {
  const query = getQuery(event)
  if (typeof query.lang === 'string') {
    return query.lang
  }

  const supported = (config.locales || []).map((l: any) => l.code || l)

  const acceptLang = getHeader(event, 'accept-language')
  if (acceptLang) {
    const preferred = acceptLang.split(',')[0]?.split('-')[0]?.trim()
    if (preferred && supported.includes(preferred)) return preferred
  }

  return config.defaultLocale || 'zh'
})
