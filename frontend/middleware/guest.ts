import { useLocalePath } from '#i18n'

// R4-06: Guest middleware — redirects authenticated users away from auth pages
export default defineNuxtRouteMiddleware(async () => {
  const { initAuth, isAuthenticated, isAdmin } = useAuth()
  await initAuth()

  if (isAuthenticated.value) {
    const localePath = useLocalePath()
    if (isAdmin.value) {
      return navigateTo(localePath('/admin'))
    }
    return navigateTo(localePath('/customer/dashboard'))
  }
})
