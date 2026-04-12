export default defineNuxtRouteMiddleware(async (to) => {
  const { initAuth, isAuthenticated, isAdmin } = useAuth()
  await initAuth()

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

  const isPublicRoute = publicRoutes.has(to.path)
  const requiresProtectedArea = to.path.startsWith('/admin') || to.path.startsWith('/customer')

  if (!isAuthenticated.value && requiresProtectedArea) {
    // R4-15: Only allow relative redirects to prevent open redirect attacks
    const redirectPath = to.fullPath.startsWith('/') ? to.fullPath : '/customer/dashboard'
    return navigateTo({
      path: '/auth/login',
      query: { redirect: redirectPath }
    })
  }

  if (isAuthenticated.value && isPublicRoute) {
    if (isAdmin.value) {
      return navigateTo('/admin')
    }
    return navigateTo('/customer/dashboard')
  }

  if (to.path.startsWith('/admin') && !isAdmin.value) {
    return navigateTo('/customer/dashboard')
  }
})

