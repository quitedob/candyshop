import { stripLocalePathPrefix } from '~/utils/stripLocalePathPrefix'

const activeRequiredRoutes = new Set([
  '/customer/inquiries/new',
  '/customer/oem-projects/new',
  '/customer/orders/new',
  '/customer/orders/quick'
])

export default defineNuxtRouteMiddleware(async (to) => {
  const path = stripLocalePathPrefix(to.path)
  if (!activeRequiredRoutes.has(path)) {
    return
  }

  const { initAuth, isAuthenticated, isAdmin, isActive } = useAuth()
  await initAuth()
  if (!isAuthenticated.value || isAdmin.value || isActive.value) {
    return
  }

  const localePath = useLocalePath()
  return navigateTo({
    path: localePath('/customer/company'),
    query: { kyb: 'required' }
  })
})
