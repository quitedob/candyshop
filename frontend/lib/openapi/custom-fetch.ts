/**
 * Orval 自定义 fetch：与 useApi 一致（Cookie 认证、SSR cookie 转发、Accept-Language）
 * 返回 { status, data, headers } 以匹配 orval fetch 客户端约定
 */
/** Orval fetch 客户端标准响应结构 */
export type OrvalFetchResponse<TData> = {
  data: TData
  status: number
  headers: Headers
}

const resolveBaseURL = (): string => {
  try {
    const config = useRuntimeConfig()
    if (import.meta.server) {
      return config.internalApiBase as string
    }
    return (config.public.apiBase as string) || '/api/v1'
  } catch {
    return '/api/v1'
  }
}

const resolveLocale = (): string => {
  try {
    return useI18n().locale.value
  } catch {
    return 'zh'
  }
}

const buildURL = (path: string, base: string): string => {
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path
  }
  const normalized = path.startsWith('/') ? path : `/${path}`
  return `${base.replace(/\/$/, '')}${normalized}`
}

const parseBody = async (response: Response): Promise<unknown> => {
  if (response.status === 204) {
    return undefined
  }
  const contentType = response.headers.get('content-type') || ''
  if (contentType.includes('application/json')) {
    return response.json()
  }
  return response.text()
}

/** customFetch 供 orval 生成的各 tag 客户端调用 */
export const customFetch = async <T>(
  url: string,
  options?: RequestInit,
): Promise<T> => {
  const init = options ?? {}
  const base = resolveBaseURL()
  const fullURL = buildURL(url, base)

  const headers = new Headers(init.headers)
  if (!headers.has('Accept-Language')) {
    headers.set('Accept-Language', resolveLocale())
  }
  const body = init.body
  if (body && !(body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  if (import.meta.server) {
    try {
      const incoming = useRequestHeaders(['cookie'])
      if (incoming.cookie && !headers.has('cookie')) {
        headers.set('cookie', incoming.cookie)
      }
    } catch {
      // 非 Nuxt 上下文（如脚本）忽略
    }
  }

  const response = await fetch(fullURL, {
    ...init,
    headers,
    credentials: 'include',
  })

  const data = await parseBody(response)

  if (!response.ok) {
    const err = new Error(`API ${response.status}: ${fullURL}`) as Error & {
      status: number
      data: unknown
    }
    err.status = response.status
    err.data = data
    throw err
  }

  return {
    data,
    status: response.status,
    headers: response.headers,
  } as T
}

export default customFetch
