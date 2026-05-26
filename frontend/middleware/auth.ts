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
    '/auth/verify-email',
    '/auth/check-email'
  ])

  const protectedRoutes = new Set([
    // /contact intentionally NOT listed here — it's a public marketing page.
    // Anything customer-only lives under /customer/*; supplier portal pages
    // are explicitly listed below.
    '/supplier',
    '/supplier/profile',
    '/supplier/purchase-orders'
  ])

  const isPublicRoute = publicRoutes.has(pathWithoutLocale)
  const requiresProtectedArea =
    pathWithoutLocale.startsWith('/admin') ||
    pathWithoutLocale.startsWith('/customer') ||
    protectedRoutes.has(pathWithoutLocale)

  if (!isAuthenticated.value && requiresProtectedArea) {
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
