import { computed } from 'vue'
import { useLocalePath } from '#i18n'

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

type JwtPayload = {
  sub?: string
  email?: string
  role?: string
  exp?: number
}

const decodeJwtPayload = (token: string): JwtPayload | null => {
  try {
    const parts = token.split('.')
    if (parts.length < 2) {
      return null
    }
    if (typeof atob !== 'function') {
      return null
    }
    const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const json = atob(base64)
    return JSON.parse(json) as JwtPayload
  } catch {
    return null
  }
}

const lazyT = () => {
  try { return useI18n().t } catch { return (key: string) => key }
}

export const useAuth = () => {
  const localePath = useLocalePath()
  const token = useCookie<string | null>('auth_token', {
    maxAge: 60 * 60 * 24 * 7,
    httpOnly: false, // must be false for client-side JS access in SPA; use secure + sameSite instead
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax'
  })
  const refreshToken = useCookie<string | null>('refresh_token', {
    maxAge: 60 * 60 * 24 * 30,
    httpOnly: false,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax'
  })
  const user = useState<User | null>('auth_user', () => null)
  const initialized = useState<boolean>('auth_initialized', () => false)

  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase || '/api/v1'

  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() =>
    isAuthenticated.value && (user.value?.role === 'admin' || user.value?.role === 'superadmin')
  )

  const isPending = computed(() =>
    isAuthenticated.value && user.value?.status === 'pending'
  )

  const clearAuthState = () => {
    token.value = null
    refreshToken.value = null
    user.value = null
  }

  const refreshAccessToken = async () => {
    if (!refreshToken.value) {
      return false
    }

    try {
      const response = await $fetch<any>(`${baseURL}/auth/refresh`, {
        method: 'POST',
        body: { refresh_token: refreshToken.value }
      })
      token.value = response.access_token
      return true
    } catch {
      return false
    }
  }

  const fetchCurrentUser = async () => {
    if (!token.value) {
      return null
    }

    const me = await $fetch<any>(`${baseURL}/auth/me`, {
      headers: {
        Authorization: `Bearer ${token.value}`
      }
    })

    const resolvedRole =
      typeof me.role === 'string' ? me.role : me.role?.name || 'customer'

    const meUser: User = {
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
    user.value = meUser
    return meUser
  }

  const initAuth = async () => {
    if (initialized.value) {
      return
    }

    initialized.value = true
    if (!token.value) {
      user.value = null
      return
    }

    const payload = decodeJwtPayload(token.value)
    const now = Math.floor(Date.now() / 1000)
    const isExpired = !!payload?.exp && payload.exp <= now

    if (isExpired) {
      const refreshed = await refreshAccessToken()
      if (!refreshed) {
        clearAuthState()
        return
      }
    }

    if (user.value) {
      return
    }

    try {
      await fetchCurrentUser()
    } catch {
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

  const login = async (credentials: { email: string; password: string }) => {
    try {
      const response = await $fetch<any>(`${baseURL}/auth/login`, {
        method: 'POST',
        body: credentials
      })

      token.value = response.access_token
      refreshToken.value = response.refresh_token

      const role =
        typeof response.user?.role === 'string'
          ? response.user.role
          : response.user?.role?.name || 'customer'

      user.value = {
        id: response.user.id,
        email: response.user.email,
        firstName: response.user.firstName || '',
        lastName: response.user.lastName || '',
        role,
        company: response.user.company,
        phone: response.user.phone,
        status: response.user.status,
        emailVerified: response.user.emailVerified
      }
      initialized.value = true
      return response
    } catch (error: any) {
      throw new Error(error.data?.message || lazyT()('auth.errors.login_failed'))
    }
  }

  const register = async (userData: Record<string, any>) => {
    try {
      return await $fetch<any>(`${baseURL}/auth/register`, {
        method: 'POST',
        body: userData
      })
    } catch (error: any) {
      throw new Error(error.data?.message || lazyT()('auth.errors.register_failed'))
    }
  }

  const logout = async () => {
    if (refreshToken.value) {
      try {
        await $fetch(`${baseURL}/auth/logout`, {
          method: 'POST',
          body: { refresh_token: refreshToken.value },
          headers: token.value ? { Authorization: `Bearer ${token.value}` } : undefined
        })
      } catch {
        // ignore logout API errors and clear local state anyway
      }
    }

    clearAuthState()
    initialized.value = false
    await navigateTo(localePath('/auth/login'))
  }

  const resendVerificationEmail = async () => {
    try {
      return await $fetch<any>(`${baseURL}/auth/resend-verification`, {
        method: 'POST'
      })
    } catch (error: any) {
      throw new Error(error.data?.message || t('auth.errors.resend_failed'))
    }
  }

  const verifyEmail = async (token: string) => {
    try {
      // R4-06/backend: VerifyEmail endpoint is POST with token in body
      return await $fetch<any>(`${baseURL}/auth/verify-email`, {
        method: 'POST',
        body: { token }
      })
    } catch (error: any) {
      throw new Error(error.data?.message || t('auth.errors.verify_failed'))
    }
  }

  return {
    user,
    token,
    isAuthenticated,
    isAdmin,
    isPending,
    login,
    register,
    logout,
    initAuth,
    resendVerificationEmail,
    verifyEmail
  }
}
