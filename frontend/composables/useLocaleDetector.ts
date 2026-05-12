export default defineI18nLocaleDetector((event) => {
  const query = getQuery(event)
  if (typeof query.lang === 'string') {
    return query.lang
  }

  const acceptLang = getHeader(event, 'accept-language')
  if (acceptLang) {
    const preferred = acceptLang.split(',')[0]?.split('-')[0]?.trim()
    if (preferred === 'en') return 'en'
    if (preferred === 'zh') return 'zh'
  }

  return 'zh'
})
