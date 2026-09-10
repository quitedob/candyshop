import { computed } from 'vue'
import { useLocalePath } from '#i18n'

// Session discovery must finish even when the API or its proxy is unavailable.
const SESSION_REQUEST_TIMEOUT_MS = 5_000

export interface User {
  id: string
  email: string
  firstName: string
  lastName: string
  role: string
  company?: string
  phone?: string
  status?: string
  emailVerified?: boolean
}

const lazyT = (): ((key: string) => string) => {
  try { return useI18n().t } catch { return (key: string) => key }
}

/** 带 Cookie 凭证的认证请求（HttpOnly JWT） */
export const authFetch = <T>(url: string, options: Record<string, unknown> = {}): Promise<T> => {
  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string> ?? {})
  }
  if (import.meta.server) {
    const incoming = useRequestHeaders(['cookie'])
    if (incoming.cookie) {
      headers.cookie = incoming.cookie
    }
  }
  return $fetch<T>(url, {
    ...options,
    credentials: 'include',
    headers
  }) as Promise<T>
}

export const useAuth = () => {
  const localePath = useLocalePath()
  const user = useState<User | null>('auth_user', () => null)
  const initialized = useState<boolean>('auth_initialized', () => false)
  const sessionActive = useState<boolean>('auth_session', () => false)

  const config = useRuntimeConfig()
  const baseURL = import.meta.server
    ? config.internalApiBase
    : (config.public.apiBase || '/api/v1')

  const isAuthenticated = computed(() => !!user.value && sessionActive.value)
  const isAdmin = computed(() =>
    isAuthenticated.value && (user.value?.role === 'admin' || user.value?.role === 'superadmin')
  )
  const isActive = computed(() =>
    isAuthenticated.value && user.value?.status === 'active'
  )
  const isPending = computed(() =>
    isAuthenticated.value && user.value?.status === 'pending'
  )

  const clearAuthState = () => {
    user.value = null
    sessionActive.value = false
  }

  const mapMeToUser = (me: any): User => {
    const resolvedRole =
      typeof me.role === 'string' ? me.role : me.role?.name || 'customer'
    return {
      id: me.id,
      email: me.email,
      firstName: me.firstName || '',
      lastName: me.lastName || '',
      role: resolvedRole,
      company: me.company,
      phone: me.phone,
      status: me.status,
      emailVerified: me.emailVerified
    }
  }

  const refreshAccessToken = async () => {
    try {
      await authFetch(`${baseURL}/auth/refresh`, { method: 'POST', body: {}, timeout: SESSION_REQUEST_TIMEOUT_MS, retry: 0 })
      sessionActive.value = true
      return true
    } catch {
      return false
    }
  }

  const fetchCurrentUser = async () => {
    const me = await authFetch<any>(`${baseURL}/auth/me`, { timeout: SESSION_REQUEST_TIMEOUT_MS, retry: 0 })
    user.value = mapMeToUser(me)
    sessionActive.value = true
    return user.value
  }

  const initAuth = async () => {
    if (initialized.value && user.value) {
      return
    }

    initialized.value = true
    try {
      await fetchCurrentUser()
    } catch (error: any) {
      // Only expired/missing access credentials can be repaired by a refresh.
      // Retrying outages would hold protected navigation open for another request.
      if ((error?.response?.status ?? error?.statusCode) !== 401) {
        clearAuthState()
        return
      }
      const refreshed = await refreshAccessToken()
      if (!refreshed) {
        clearAuthState()
        return
      }
      try {
        await fetchCurrentUser()
      } catch {
        clearAuthState()
      }
    }
  }

  const login = async (credentials: { email: string; password: string; remember?: boolean }) => {
    try {
      const response = await authFetch<any>(`${baseURL}/auth/login`, {
        method: 'POST',
        body: credentials
      })

      sessionActive.value = true
      if (response.user) {
        user.value = mapMeToUser(response.user)
      } else {
        await fetchCurrentUser()
      }
      initialized.value = true
      return response
    } catch (error: any) {
      throw new Error(error.data?.message || lazyT()('auth.errors.login_failed'))
    }
  }

  const register = async (userData: Record<string, any>) => {
    try {
      const response = await authFetch<any>(`${baseURL}/auth/register`, {
        method: 'POST',
        body: userData
      })
      sessionActive.value = true
      if (response.user) {
        user.value = mapMeToUser(response.user)
      } else {
        await fetchCurrentUser()
      }
      initialized.value = true
      return response
    } catch (error: any) {
      throw new Error(error.data?.message || lazyT()('auth.errors.register_failed'))
    }
  }

  const logout = async () => {
    try {
      await authFetch(`${baseURL}/auth/logout`, { method: 'POST', body: {} })
    } catch {
      // ignore logout API errors
    }
    clearAuthState()
    initialized.value = false
    await navigateTo(localePath('/auth/login'))
  }

  const resendVerificationEmail = async (email?: string) => {
    try {
      if (sessionActive.value && user.value) {
        return await authFetch<any>(`${baseURL}/auth/resend-verification`, {
          method: 'POST'
        })
      }
      if (!email?.trim()) {
        throw new Error(lazyT()('auth.errors.resend_failed'))
      }
      return await authFetch<any>(`${baseURL}/auth/resend-verification-public`, {
        method: 'POST',
        body: { email: email.trim() }
      })
    } catch (error: any) {
      throw new Error(error.data?.message || lazyT()('auth.errors.resend_failed'))
    }
  }

  const verifyEmail = async (token: string) => {
    try {
      return await authFetch<any>(`${baseURL}/auth/verify-email`, {
        method: 'POST',
        body: { token }
      })
    } catch (error: any) {
      throw new Error(error.data?.message || lazyT()('auth.errors.verify_failed'))
    }
  }

  /** @deprecated HttpOnly cookie 模式下无客户端 token，SSE 请用 credentials:'include' */
  const token = computed(() => null as string | null)

  return {
    user,
    token,
    isAuthenticated,
    isAdmin,
    isActive,
    isPending,
    login,
    register,
    logout,
    initAuth,
    refreshAccessToken,
    resendVerificationEmail,
    verifyEmail,
    authFetch
  }
}
