/** 客户端 JWT 自动刷新：在 access token 过期前主动调用 /auth/refresh */
import { refreshAuthSession } from '~/composables/useApi'

export default defineNuxtPlugin((nuxtApp) => {
  if (import.meta.server) return

  const config = useRuntimeConfig()
  const accessMinutes = Number(config.public.jwtAccessMinutes || 15)
  const refreshMs = Math.max(60_000, Math.floor(accessMinutes * 60 * 1000 * 0.8))

  let timer: ReturnType<typeof setInterval> | null = null

  const tryRefresh = async () => {
    const { isAuthenticated } = useAuth()
    if (!isAuthenticated.value) return
    // M1: coalesce through the shared in-flight refresh so the interval + focus/
    // visibility triggers and useApi() 401 retries never issue overlapping
    // /auth/refresh calls (which would rotate each other's tokens out and force
    // spurious logouts).
    await refreshAuthSession()
  }

  const schedule = () => {
    if (timer) clearInterval(timer)
    timer = setInterval(() => {
      void tryRefresh()
    }, refreshMs)
  }

  schedule()

  const onVisibilityChange = () => {
    if (document.visibilityState === 'visible') {
      void tryRefresh()
    }
  }

  const onFocus = () => {
    void tryRefresh()
  }

  document.addEventListener('visibilitychange', onVisibilityChange)
  window.addEventListener('focus', onFocus)

  // 插件内不能使用 onBeforeUnmount，改用 Nuxt 应用生命周期钩子清理
  nuxtApp.hook('app:beforeUnmount' as unknown as Parameters<typeof nuxtApp.hook>[0], () => {
    if (timer) clearInterval(timer)
    document.removeEventListener('visibilitychange', onVisibilityChange)
    window.removeEventListener('focus', onFocus)
  })
})
