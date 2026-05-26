import { useLocalePath } from '#i18n'

// R4-06: Guest middleware — redirects authenticated users away from auth pages
export default defineNuxtRouteMiddleware(async (to) => {
  const { initAuth, isAuthenticated, isAdmin, user } = useAuth()
  await initAuth()

  if (!isAuthenticated.value) {
    return
  }

  // 已登录但未验证邮箱的用户可停留在 check-email 页重发验证邮件
  const checkEmailPath = '/auth/check-email'
  if (to.path.includes(checkEmailPath) && user.value && !user.value.emailVerified) {
    return
  }

  const localePath = useLocalePath()
  if (isAdmin.value) {
    return navigateTo(localePath('/admin'))
  }
  return navigateTo(localePath('/customer/dashboard'))
})
