// R4-06: Guest middleware — redirects authenticated users away from auth pages
export default defineNuxtRouteMiddleware(async () => {
  const { initAuth, isAuthenticated, isAdmin } = useAuth()
  await initAuth()

  if (isAuthenticated.value) {
    if (isAdmin.value) {
      return navigateTo('/admin')
    }
    return navigateTo('/customer/dashboard')
  }
})
