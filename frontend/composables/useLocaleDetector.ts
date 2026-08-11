export default defineI18nLocaleDetector((event, config) => {
  const query = getQuery(event)
  if (typeof query.lang === 'string') {
    return query.lang
  }

  const cfg = config as { locales?: Array<{ code?: string } | string>; defaultLocale?: string }
  const locales = cfg.locales
  const supported = Array.isArray(locales) ? locales.map((l) => (typeof l === 'string' ? l : (l.code ?? ''))) : []

  const userLocale = getCookie(event, 'user-locale')
  if (userLocale && supported.includes(userLocale)) return userLocale

  // Fallback to Accept-Language header
  const acceptLang = getHeader(event, 'accept-language')
  if (acceptLang) {
    const preferred = acceptLang.split(',')[0]?.trim()?.slice(0, 2)
    if (preferred && supported.includes(preferred)) return preferred
  }

  return cfg.defaultLocale || 'zh'
})
