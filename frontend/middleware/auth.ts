import { useLocalePath } from '#i18n'
import { stripLocalePathPrefix } from '~/utils/stripLocalePathPrefix'

export default defineNuxtRouteMiddleware(async (to) => {
  const { initAuth, isAuthenticated, isAdmin } = useAuth()
  await initAuth()

  const localePath = useLocalePath()
  const pathWithoutLocale = stripLocalePathPrefix(to.path)

  const publicRoutes = new Set([
    '/login',
    '/register',
    '/forgot-password',
    '/reset-password',
    '/auth/login',
    '/auth/register',
    '/auth/forgot-password',
    '/auth/reset-password',
    '/auth/verify-email'
  ])

  const isPublicRoute = publicRoutes.has(pathWithoutLocale)
  const requiresProtectedArea =
    pathWithoutLocale.startsWith('/admin') || pathWithoutLocale.startsWith('/customer')

  if (!isAuthenticated.value && requiresProtectedArea) {
    // R4-15: Only allow relative redirects to prevent open redirect attacks
    const redirectPath = to.fullPath.startsWith('/') ? to.fullPath : localePath('/customer/dashboard')
    return navigateTo({
      path: localePath('/auth/login'),
      query: { redirect: redirectPath }
    })
  }

  if (isAuthenticated.value && isPublicRoute) {
    if (isAdmin.value) {
      return navigateTo(localePath('/admin'))
    }
    return navigateTo(localePath('/customer/dashboard'))
  }

  if (pathWithoutLocale.startsWith('/admin') && !isAdmin.value) {
    return navigateTo(localePath('/customer/dashboard'))
  }
})

