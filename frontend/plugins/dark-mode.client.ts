/** 应用启动时初始化暗色模式（客户端） */
export default defineNuxtPlugin((nuxtApp) => {
  const { isDark, apply, themeCookie } = useDarkMode()

  // 在 app:mounted 钩子中同步主题，避免 composable 内误用 onMounted
  nuxtApp.hook('app:mounted', () => {
    if (themeCookie.value === 'dark' || themeCookie.value === 'light') {
      apply(themeCookie.value === 'dark', false)
      return
    }
    isDark.value = document.documentElement.classList.contains('dark')
  })
})
