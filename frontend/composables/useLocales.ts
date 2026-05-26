/** 与 nuxt.config i18n.locales 保持一致的语言元数据 */
export interface LocaleMeta {
  code: string
  name: string
}

export const LOCALE_META: LocaleMeta[] = [
  { code: 'en', name: 'English' },
  { code: 'zh', name: '中文' },
  { code: 'ko', name: '한국어' },
  { code: 'ar', name: 'العربية' },
  { code: 'ja', name: '日本語' },
  { code: 'th', name: 'ไทย' },
  { code: 'vi', name: 'Tiếng Việt' },
  { code: 'id', name: 'Bahasa Indonesia' },
  { code: 'ms', name: 'Bahasa Melayu' },
]

export const ALL_LOCALES = LOCALE_META.map((l) => l.code)

/** 管理端内容/产品的主语言（标量字段同步来源） */
export const SOURCE_LOCALE = 'zh'

/** 供页面/组件通过 useLocales() 获取语言列表 */
export function useLocales() {
  return { LOCALE_META, ALL_LOCALES, SOURCE_LOCALE }
}
