import { detectBrowserLanguage } from '#i18n'

export default defineI18nLocaleDetector((event) => {
  const query = getQuery(event)
  if (typeof query.lang === 'string') {
    return query.lang
  }

  const accept = getHeader(event, 'accept-language')
  if (accept) {
    return detectBrowserLanguage(accept, { localeCodes: ['zh', 'en'] }) || 'zh'
  }

  return 'zh'
})
