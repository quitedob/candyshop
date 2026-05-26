import { stripLocalePathPrefix } from '~/utils/stripLocalePathPrefix'

/** 供应商门户已关闭：访问 /supplier 一律重定向到首页（ToB 工厂直销模式） */
export default defineNuxtRouteMiddleware((to) => {
  const pathWithoutLocale = stripLocalePathPrefix(to.path)
  if (!pathWithoutLocale.startsWith('/supplier')) {
    return
  }
  const localePath = useLocalePath()
  return navigateTo(localePath('/'))
})
