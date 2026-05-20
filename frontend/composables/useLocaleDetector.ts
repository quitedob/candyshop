export default defineI18nLocaleDetector((event, config) => {
  const query = getQuery(event)
  if (typeof query.lang === 'string') {
    return query.lang
  }

  const supported = (config.locales || []).map((l: any) => l.code || l)

  const userLocale = getCookie(event, 'user-locale')
  if (userLocale && supported.includes(userLocale)) return userLocale

  // Fallback to Accept-Language header
  const acceptLang = getHeader(event, 'accept-language')
  if (acceptLang) {
    const preferred = acceptLang.split(',')[0]?.trim()?.slice(0, 2)
    if (preferred && supported.includes(preferred)) return preferred
  }

  return config.defaultLocale || 'zh'
})
