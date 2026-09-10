export default defineNuxtPlugin((nuxtApp) => {
  const { initAuth } = useAuth()
  // Public pages must stay interactive during an API outage. Protected routes
  // still await initAuth in auth middleware before rendering private content.
  nuxtApp.hook('app:mounted', () => { void initAuth() })
})
