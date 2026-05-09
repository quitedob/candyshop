import { localeCodes } from '#build/i18n.options.mjs'

/**
 * 去掉路径首段的 i18n 语言码（与 @nuxtjs/i18n 的 prefix_except_default 一致），
 * 供路由中间件用「无语言前缀」路径判断 admin / customer / auth 等。
 */
export function stripLocalePathPrefix(path: string): string {
  const segs = path.split('/').filter(Boolean)
  if (!segs.length) {
    return path
  }
  if (!localeCodes.includes(segs[0])) {
    return path
  }
  const tail = segs.slice(1).join('/')
  return tail ? `/${tail}` : '/'
}
