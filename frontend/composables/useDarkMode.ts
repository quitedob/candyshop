/** 暗色模式切换 composable — cookie 持久化，SSR 可读取用户偏好 */
export function useDarkMode() {
  const themeCookie = useCookie<'dark' | 'light' | ''>('theme', {
    default: () => '',
    maxAge: 60 * 60 * 24 * 365,
    sameSite: 'lax',
  })

  // 服务端从 cookie 初始化；客户端无 cookie 时读取 inline script 已设置的 html class
  const isDark = useState('dark-mode', () => {
    if (themeCookie.value === 'dark') return true
    if (themeCookie.value === 'light') return false
    if (import.meta.client) {
      return document.documentElement.classList.contains('dark')
    }
    return false
  })

  const apply = (dark: boolean, persist = true) => {
    isDark.value = dark
    if (persist) {
      themeCookie.value = dark ? 'dark' : 'light'
    }
    if (import.meta.client) {
      document.documentElement.classList.toggle('dark', dark)
    }
  }

  const toggle = () => apply(!isDark.value)

  // 全局只注册一次 html class，供 SSR 输出
  const headReady = useState('dark-mode-head', () => false)
  if (!headReady.value) {
    headReady.value = true
    useHead({
      htmlAttrs: {
        class: computed(() => (isDark.value ? 'dark' : '')),
      },
    })
  }

  return { isDark, toggle, apply, themeCookie }
}
