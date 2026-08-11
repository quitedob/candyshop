/**
 * API Client Composable
 * Provides a typed wrapper for all API calls with error handling
 */

interface ProductQuery {
  category?: string
  categories?: string
  page?: number
  limit?: number
  sort?: 'name' | 'created' | 'popular' | 'moq_low' | 'moq_high'
  order?: 'asc' | 'desc'
  search?: string
  halal?: boolean
  oemOnly?: boolean
  featured?: boolean
  minMoq?: number
  maxMoq?: number
  filters?: Record<string, string[]>
}

interface Product {
  id: string
  slug: string
  name: string
  summary: string
  description: string
  category: string
  categorySlug: string
  thumbnail: string
  ogImage?: string
  images: string[]
  oemAvailable: boolean
  halalCertified: boolean
  certifications: string[]
  moq: number
  leadTime: string
  featured?: boolean
  flavors?: string[]
  shapes?: string[]
  ingredients?: string
  allergens?: string
  shelfLife?: string
  storage?: string
  translations?: Record<string, Record<string, string>>
  // Weight & Measurement
  netWeightPerPiece?: number
  netWeightPerPack?: number
  grossWeightPerCarton?: number
  piecesPerPack?: number
  packsPerCarton?: number
  // Dimensions
  productLengthMM?: number
  productWidthMM?: number
  productHeightMM?: number
  // Nutrition (per 100g)
  energyKj?: number
  energyKcal?: number
  totalFatG?: number
  saturatedFatG?: number
  carbohydratesG?: number
  sugarsG?: number
  proteinG?: number
  saltG?: number
  fiberG?: number
  // Ingredient Compliance
  additives?: string[]
  sweetenerType?: string
  cocoaSolidsPct?: number
  milkSolidsPct?: number
  gmoStatus?: string
  mayContain?: string[]
  waterActivity?: number
  // Trade & Barcode
  gtin?: string
  hsCode?: string
  // Packaging
  primaryPackaging?: string
  innerPackConfig?: string
  palletConfig?: string
  // Dietary
  isVegan?: boolean
  isGlutenFree?: boolean
  isSugarFree?: boolean
  isKosher?: boolean
  isOrganic?: boolean
  // Certification Details
  certificationDetails?: CertificationDetail[]
  // Sample Specs
  sampleMOQ?: number
  sampleLeadTime?: string
  samplePrice?: number
  // Meta
  status?: string
  basePrice?: number
  stockQuantity?: number
  viewCount?: number
  createdBy?: string
  updatedBy?: string
  createdAt?: string
  updatedAt?: string
}

interface CertificationDetail {
  name: string
  abbrev?: string
  issuedBy?: string
  certNumber?: string
  validUntil?: string
  docUrl?: string
}

interface Category {
  slug: string
  name: string
  description: string
  thumbnail: string
  icon?: string
  productCount: number
}

interface OEMFlow {
  id: string
  title: string
  description: string
  type: 'quick_odm' | 'full_oem'
  steps: OEMStep[]
  timeline: string
  moq: number
}

interface OEMStep {
  order: number
  title: string
  description: string
  duration: string
  deliverables: string[]
}

interface OEMSolution {
  id: string
  slug: string
  title: string
  description: string
  thumbnail: string
  images: string[]
  category: string
  moq: number
  applications: string[]
}

interface FactoryInfo {
  name: string
  founded: number
  description: string
  address: string
  capacity: {
    daily: string
    monthly: string
  }
  area: string
  employees: number
  markets: string[]
  images: string[]
}

interface Certification {
  id: string
  name: string
  abbreviation: string
  description: string
  issuer: string
  validUntil: string
  certificateUrl?: string
  badgeUrl: string
}

interface ProcessControl {
  id: string
  name: string
  description: string
  stage: 'raw_material' | 'production' | 'finished_product'
  standards: string[]
}

interface QualityTimeline {
  year: number
  milestone: string
  description: string
}

interface BlogPost {
  id: string
  slug: string
  title: string
  excerpt: string
  content: string
  category: string
  author: {
    name: string
    avatar?: string
    title?: string
    bio?: string
  }
  publishedAt: string
  thumbnail?: string
  readTime: number
  tags: string[]
}

interface CaseStudy {
  id: string
  slug: string
  title: string
  client: string
  industry: string
  location: string
  thumbnail: string
  images: string[]
  challenge: string
  solution: string
  result: string
  timeline: string
  services: string[]
}

interface InquiryData {
  companyName: string
  contactPerson: string
  email: string
  whatsapp?: string
  targetCountry?: string
  estimatedQuantity?: string
  interestedProducts?: string[]
  productIds?: string[]
  packagingRequirements?: string
  flavorRequirements?: string
  oemNeeded?: boolean
  expectedDelivery?: string
  message?: string
  files?: File[]
  recaptchaToken?: string
}

interface InquiryResponse {
  success: boolean
  message: string
  inquiryId?: string
}

interface PaginatedResponse<T> {
  data: T[]
  pagination: {
    total: number
    page: number
    limit: number
    totalPages: number
  }
}

interface ApiError {
  message: string
  statusCode?: number
  details?: unknown
}

// M1: Module-scoped in-flight refresh promise.
//
// Without this, every useApi() instance carries its own refreshPromise and the
// auth-auto-refresh plugin's interval + focus/visibility handlers call
// refreshAccessToken() directly. Two overlapping /auth/refresh calls each carry
// the same refresh-token cookie; the backend rotates the token on the first
// success, so the second (racing) call fails and triggers a spurious forced
// logout. Routing every trigger through one promise coalesces them into a
// single refresh call.
let sharedRefreshPromise: Promise<boolean> | null = null

/**
 * Trigger a token refresh, coalescing concurrent callers (useApi() 401 retries,
 * the auth-auto-refresh plugin, focus/visibility handlers) into a single
 * in-flight /auth/refresh request. On the server, SSR requests are isolated per
 * render and must not share a module-level promise (cookies differ per request),
 * so only the client path coalesces.
 */
export const refreshAuthSession = (): Promise<boolean> => {
  const { refreshAccessToken } = useAuth()
  if (import.meta.server) {
    return refreshAccessToken()
  }
  if (!sharedRefreshPromise) {
    sharedRefreshPromise = refreshAccessToken().finally(() => {
      sharedRefreshPromise = null
    })
  }
  return sharedRefreshPromise
}

export const useApi = () => {
  const config = useRuntimeConfig()
  // During SSR, $fetch uses Nitro's localFetch which bypasses devProxy.
  // Use the direct backend URL on the server so requests don't fall through to Vue Router.
  const baseURL = import.meta.server
    ? config.internalApiBase
    : config.public.apiBase || '/api/v1'
  const publicBaseURL = `${baseURL}/public`
  const { t, locale } = useI18n()
  const auth = useAuth()
  const requestHeaders = import.meta.server ? useRequestHeaders(['cookie']) : { cookie: undefined as string | undefined }

  /**
   * Generic fetch wrapper with error handling and automatic 401 retry
   */
  const isFormData = (body: unknown): body is FormData => body instanceof FormData

  const fetchApi = async <T>(
    endpoint: string,
    options?: Record<string, unknown>
  ): Promise<T> => {
    const doFetch = async (): Promise<T> => {
      const body = options?.body
      const headers: Record<string, string> = {
        'Accept-Language': locale.value,
        ...(options?.headers as Record<string, string> ?? {})
      }

      if (!isFormData(body)) {
        headers['Content-Type'] = 'application/json'
      }

      if (import.meta.server && requestHeaders.cookie) {
        headers.cookie = requestHeaders.cookie
      }

      return await $fetch<T>(`${baseURL}${endpoint}`, {
        ...options,
        credentials: 'include',
        headers
      }) as T
    }

    try {
      return await doFetch()
    } catch (err: unknown) {
      let error = err as { statusCode?: number; data?: unknown }

      if (error?.statusCode === 401) {
        const refreshed = await refreshAuthSession()
        if (refreshed) {
          try {
            return await doFetch()
          } catch (retryErr: unknown) {
            error = retryErr as { statusCode?: number; data?: unknown }
          }
        } else {
          auth.logout()
          throw { message: t('errors.session_expired'), statusCode: 401 }
        }
      }

      const apiError: ApiError = {
        message: t('errors.default'),
        statusCode: error?.statusCode
      }

      // M1: the error body may be a non-object (e.g. a proxy error page or a
      // plain-text response), in which case the old `'details' in error.data`
      // check threw a TypeError. Guard the shape before normalizing.
      const errorData = error?.data
      if (typeof errorData === 'object' && errorData !== null) {
        const data = errorData as Record<string, unknown>
        if (typeof data.message === 'string' && data.message) {
          apiError.message = data.message
        }
        if ('details' in data) {
          apiError.details = data.details
        }
      }

      throw apiError
    }
  }

  const fetchPublicApi = async <T>(
    endpoint: string,
    options?: Record<string, unknown>
  ): Promise<T> => {
    return fetchApi<T>(`/public${endpoint}`, options)
  }

  /**
   * Products
   */
  const getCategories = async (): Promise<Category[]> => {
    return fetchPublicApi<Category[]>('/categories')
  }

  const getCategory = async (slug: string): Promise<Category & { products?: Product[] }> => {
    return fetchPublicApi<Category & { products?: Product[] }>(`/categories/${slug}`)
  }

  const getProducts = async (params: ProductQuery = {}): Promise<PaginatedResponse<Product>> => {
    const queryParams = new URLSearchParams()

    if (params.category) queryParams.append('category', params.category)
    if (params.categories) queryParams.append('categories', params.categories)
    if (params.page) queryParams.append('page', params.page.toString())
    if (params.limit) queryParams.append('limit', params.limit.toString())
    if (params.sort) queryParams.append('sort', params.sort)
    if (params.order) queryParams.append('order', params.order)
    if (params.search) queryParams.append('search', params.search)
    if (params.halal) queryParams.append('halal', 'true')
    if (params.oemOnly) queryParams.append('oemOnly', 'true')
    if (params.featured) queryParams.append('featured', 'true')
    if (params.minMoq != null && Number.isFinite(params.minMoq) && params.minMoq > 0) {
      queryParams.append('minMoq', String(params.minMoq))
    }
    if (params.maxMoq != null && Number.isFinite(params.maxMoq) && params.maxMoq > 0) {
      queryParams.append('maxMoq', String(params.maxMoq))
    }
    if (params.filters) {
      Object.entries(params.filters).forEach(([key, values]) => {
        values.forEach(value => queryParams.append(`filter[${key}]`, value))
      })
    }

    const queryString = queryParams.toString()
    return fetchPublicApi<PaginatedResponse<Product>>(`/products${queryString ? `?${queryString}` : ''}`)
  }

  const getProduct = async (slug: string): Promise<Product> => {
    return fetchPublicApi<Product>(`/products/${slug}`)
  }

  const getFeaturedProducts = async (limit = 8): Promise<Product[]> => {
    return fetchPublicApi<Product[]>(`/products/featured?limit=${limit}`)
  }

  const getRelatedProducts = async (slug: string, limit = 4): Promise<Product[]> => {
    return fetchPublicApi<Product[]>(`/products/${slug}/related?limit=${limit}`)
  }

  /**
   * OEM
   */
  const getOEMFlows = async (): Promise<OEMFlow[]> => {
    return fetchPublicApi<OEMFlow[]>('/oem/flows')
  }

  const getOEMFlow = async (id: string): Promise<OEMFlow> => {
    return fetchPublicApi<OEMFlow>(`/oem/flows/${id}`)
  }

  const getOEMSolutions = async (): Promise<OEMSolution[]> => {
    return fetchPublicApi<OEMSolution[]>('/oem/solutions')
  }

  const getOEMSolution = async (slug: string): Promise<OEMSolution> => {
    return fetchPublicApi<OEMSolution>(`/oem/solutions/${slug}`)
  }

  /**
   * Factory
   */
  const getFactoryInfo = async (): Promise<FactoryInfo> => {
    return fetchPublicApi<FactoryInfo>('/factory')
  }

  const getCertifications = async (): Promise<Certification[]> => {
    return fetchPublicApi<Certification[]>('/certifications')
  }

  const getCertification = async (id: string): Promise<Certification> => {
    return fetchPublicApi<Certification>(`/certifications/${id}`)
  }

  const getProcessControls = async (): Promise<ProcessControl[]> => {
    return fetchPublicApi<ProcessControl[]>('/factory/quality-controls')
  }

  const getQualityTimeline = async (): Promise<QualityTimeline[]> => {
    return fetchPublicApi<QualityTimeline[]>('/factory/timeline')
  }

  /**
   * Content (Blog & Cases)
   */
  const getPosts = async (params: {
    category?: string
    page?: number
    limit?: number
  } = {}): Promise<PaginatedResponse<BlogPost>> => {
    const queryParams = new URLSearchParams()

    if (params.category) queryParams.append('category', params.category)
    if (params.page) queryParams.append('page', params.page.toString())
    if (params.limit) queryParams.append('limit', params.limit.toString())

    const queryString = queryParams.toString()
    return fetchPublicApi<PaginatedResponse<BlogPost>>(`/posts${queryString ? `?${queryString}` : ''}`)
  }

  const getPost = async (slug: string): Promise<BlogPost> => {
    return fetchPublicApi<BlogPost>(`/posts/${slug}`)
  }

  const getRelatedPosts = async (slug: string, limit = 3): Promise<BlogPost[]> => {
    return fetchPublicApi<BlogPost[]>(`/posts/${slug}/related?limit=${limit}`)
  }

  const getCases = async (params: {
    industry?: string
    page?: number
    limit?: number
  } = {}): Promise<PaginatedResponse<CaseStudy>> => {
    const queryParams = new URLSearchParams()

    if (params.industry) queryParams.append('industry', params.industry)
    if (params.page) queryParams.append('page', params.page.toString())
    if (params.limit) queryParams.append('limit', params.limit.toString())

    const queryString = queryParams.toString()
    return fetchPublicApi<PaginatedResponse<CaseStudy>>(`/cases${queryString ? `?${queryString}` : ''}`)
  }

  const getCase = async (slug: string): Promise<CaseStudy> => {
    return fetchPublicApi<CaseStudy>(`/cases/${slug}`)
  }

  const getRelatedCases = async (slug: string, limit = 3): Promise<CaseStudy[]> => {
    return fetchPublicApi<CaseStudy[]>(`/cases/${slug}/related?limit=${limit}`)
  }

  /**
   * Inquiry Submission
   */
  const submitInquiry = async (data: InquiryData): Promise<InquiryResponse> => {
    // Create FormData for file upload
    const formData = new FormData()

    Object.entries(data).forEach(([key, value]) => {
      if (key === 'files' && value instanceof Array) {
        value.forEach((file) => {
          formData.append('files', file)
        })
      } else if (value instanceof Array) {
        value.forEach((item) => {
          formData.append(key, String(item))
        })
      } else if (value !== undefined && value !== null) {
        formData.append(key, String(value))
      }
    })

    try {
      return await $fetch<InquiryResponse>(`${publicBaseURL}/inquiry`, {
        method: 'POST',
        body: formData
      })
    } catch (err: unknown) {
      const error = err as { statusCode?: number; data?: { message?: string } }
      const apiError: ApiError = {
        message: t('form.error'),
        statusCode: error?.statusCode
      }

      if (error?.data?.message) {
        apiError.message = error.data.message
      }

      throw apiError
    }
  }

  const submitCustomerInquiry = async (data: InquiryData): Promise<InquiryResponse> => {
    const formData = new FormData()
    Object.entries(data).forEach(([key, value]) => {
      if (key === 'files' && Array.isArray(value)) {
        value.forEach((file) => formData.append('files', file))
      } else if (Array.isArray(value)) {
        value.forEach((item) => formData.append(key, String(item)))
      } else if (value !== undefined && value !== null) {
        formData.append(key, String(value))
      }
    })
    return fetchApi<InquiryResponse>('/user/inquiries', {
      method: 'POST',
      body: formData
    })
  }

  /**
   * Customer Orders
   */
  const getOrders = async (params: { page?: number; limit?: number } = {}): Promise<PaginatedResponse<any>> => {
    const queryParams = new URLSearchParams()
    if (params.page) queryParams.append('page', params.page.toString())
    if (params.limit) queryParams.append('limit', (params.limit || 20).toString())
    const queryString = queryParams.toString()
    return fetchApi<PaginatedResponse<any>>(`/user/orders${queryString ? `?${queryString}` : ''}`)
  }

  const getOrder = async (id: string): Promise<any> => {
    const res = await fetchApi<any>(`/user/orders/${id}`)
    if (res && typeof res === 'object' && 'order' in res) {
      return { ...res.order, canApprove: res.canApprove, isOwner: res.isOwner }
    }
    return res
  }

  // R2 B-4: critical customer mutations are routed through the POST helper so
  // they pick up the auto-generated Idempotency-Key. Direct fetchApi POSTs
  // were duplicate-prone on retries (network blip during checkout) — the
  // backend M-17 middleware deduplicates only when this header is present.
  const approveOrder = async (id: string, comment = '') =>
    POST<any>(`/user/orders/${id}/approve`, { comment })

  const rejectOrder = async (id: string, comment = '') =>
    POST<any>(`/user/orders/${id}/reject`, { comment })

  const createOrder = async (data: {
    items: { productId: string; quantity: number; unitPrice: number; specifications?: string }[]
    currency?: string
    taxAmount?: number
    shippingAmount?: number
    incoterms?: string
    estimatedWeightKg?: number
    shippingAddress: { street: string; city: string; state?: string; zipCode?: string; country: string }
    inquiryId?: string
  }): Promise<any> => {
    return POST<any>('/user/orders', data)
  }

  const createAIAssistOrder = async (data: {
    prompt: string
    targetCountry: string
    quantity: number
    budget?: number
    currency?: string
    taxAmount?: number
    shippingAmount?: number
    shippingAddress: { street: string; city: string; state?: string; zipCode?: string; country: string }
    additionalRequirements?: string
    inquiryId?: string
  }): Promise<any> => {
    return POST<any>('/user/orders/ai-assist', data)
  }

  const confirmOrder = async (id: string, complianceAck: boolean): Promise<any> => {
    return POST<any>(`/user/orders/${id}/confirm`, { complianceAck })
  }

  /**
   * Search
   */
  const search = async (query: string, filters?: {
    type?: 'products' | 'posts' | 'cases' | 'all'
    limit?: number
  }): Promise<{
    products: Product[]
    posts: BlogPost[]
    cases: CaseStudy[]
  }> => {
    const params = new URLSearchParams()
    params.append('q', query)
    if (filters?.type) params.append('type', filters.type)
    if (filters?.limit) params.append('limit', filters.limit.toString())

    return fetchPublicApi(`/search?${params.toString()}`)
  }

  // Helper methods
  const GET = <T>(endpoint: string, params?: Record<string, any>): Promise<T> => {
    const queryString = params ? '?' + new URLSearchParams(params).toString() : ''
    return fetchApi<T>(`${endpoint}${queryString}`)
  }

  // R2 B-4: auto-generate an Idempotency-Key for every POST unless the
  // caller has already supplied one (or explicitly opted out via the
  // skipIdempotency option). This complements the backend M-17 middleware:
  // when a customer's browser retries a checkout/payment/inquiry mutation
  // (network blip, double-click, page reload during request), the second
  // call hits the same key and the backend returns the original response
  // instead of creating a duplicate row. PUT/DELETE remain idempotent at
  // the HTTP-method level so we don't add the header there.
  const generateIdempotencyKey = (): string => {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
      return crypto.randomUUID()
    }
    return `idem-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
  }

  const POST = <T>(endpoint: string, data?: any, options?: Record<string, unknown>): Promise<T> => {
    const body = isFormData(data) ? data : JSON.stringify(data)
    const skip = options?.skipIdempotency === true
    const callerHeaders = (options?.headers as Record<string, string> | undefined) ?? {}
    const hasKey = Object.keys(callerHeaders).some((h) => h.toLowerCase() === 'idempotency-key')
    const finalOptions: Record<string, unknown> = { ...options, method: 'POST', body }
    if (!skip && !hasKey) {
      finalOptions.headers = { ...callerHeaders, 'Idempotency-Key': generateIdempotencyKey() }
    }
    delete finalOptions.skipIdempotency
    return fetchApi<T>(endpoint, finalOptions)
  }

  const PUT = <T>(endpoint: string, data?: any): Promise<T> => {
    const body = isFormData(data) ? data : JSON.stringify(data)
    return fetchApi<T>(endpoint, { method: 'PUT', body })
  }

  const DELETE = <T>(endpoint: string, data?: any): Promise<T> => {
    const opts: Record<string, unknown> = { method: 'DELETE' }
    if (data !== undefined) {
      opts.body = JSON.stringify(data)
    }
    return fetchApi<T>(endpoint, opts)
  }

  // Admin - Companies
  const adminGetCompanies = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/admin/companies', params)
  const adminGetCompany = (id: string) => GET<any>(`/admin/companies/${id}`)
  const adminUpdateCompany = (id: string, data: any) => PUT<any>(`/admin/companies/${id}`, data)
  const adminVerifyCompany = (id: string, status: string) => PUT<any>(`/admin/companies/${id}/verify`, { status })

  // Admin - Order Payments
  const adminGetOrderPayments = (orderId: string) => GET<any[]>(`/admin/orders/${orderId}/payments`)
  const adminCreatePayment = (orderId: string, data: any) => POST<any>(`/admin/orders/${orderId}/payments`, data)
  const adminConfirmPayment = (orderId: string, paymentId: string, data?: any) => PUT<any>(`/admin/orders/${orderId}/payments/${paymentId}/confirm`, data)
  const adminRefundPayment = (orderId: string, paymentId: string, data?: any) => PUT<any>(`/admin/orders/${orderId}/payments/${paymentId}/refund`, data)

  // Admin - Trades
  const adminGetTrades = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/admin/trades', params)
  const adminGetTrade = (id: string) => GET<any>(`/admin/trades/${id}`)
  const adminUpdateTradeStatus = (id: string, data: any) => PUT<any>(`/admin/trades/${id}/status`, data)
  const adminGetTradeDocuments = (id: string) => GET<any[]>(`/admin/trades/${id}/documents`)
  const adminUpdateTradeDocument = (tradeId: string, docId: string, data: any) => PUT<any>(`/admin/trades/${tradeId}/documents/${docId}`, data)

  // Admin - OEM Projects
  const adminGetOemProjects = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/admin/oem-projects', params)
  const adminGetOemProject = (id: string) => GET<any>(`/admin/oem-projects/${id}`)
  const adminUpdateOemProject = (id: string, data: any) => PUT<any>(`/admin/oem-projects/${id}`, data)
  const adminUpdateOemStatus = (id: string, data: any) => PUT<any>(`/admin/oem-projects/${id}/status`, data)
  // OEM sample lifecycle (P0.2 / G-OEM-1)
  const adminAddOemSample = (id: string, data: { name: string }) => POST<any>(`/admin/oem-projects/${id}/samples`, data)
  const adminUpdateOemSample = (id: string, sampleId: string, data: { status: string; feedback?: string }) =>
    PUT<any>(`/admin/oem-projects/${id}/samples/${sampleId}`, data)
  // OEM project → order conversion (P0.2 / G-OEM-2)
  const adminConvertOemProjectToOrder = (id: string, data: { productId?: string; quantity?: number; unitPrice?: number; currency?: string; incoterms?: string; shippingAddress?: any }) =>
    POST<any>(`/admin/oem-projects/${id}/convert-to-order`, data)
  // OEM inventory holds (P0.2 / G-OEM-3)
  const adminCreateOemInventoryHold = (id: string, data: { productId: string; quantity: number; notes?: string }) =>
    POST<any>(`/admin/oem-projects/${id}/inventory-holds`, data)
  const adminListOemInventoryHolds = (id: string) => GET<any>(`/admin/oem-projects/${id}/inventory-holds`)
  const adminReleaseOemInventoryHold = (id: string, holdId: string | number) =>
    DELETE<any>(`/admin/oem-projects/${id}/inventory-holds/${holdId}`)

  // Admin - OEM Flows
  const adminGetOemFlows = () => GET<any[]>('/admin/oem-flows')
  const adminGetOemFlow = (id: string) => GET<any>(`/admin/oem-flows/${id}`)
  const adminCreateOemFlow = (data: any) => POST<any>('/admin/oem-flows', data)
  const adminUpdateOemFlow = (id: string, data: any) => PUT<any>(`/admin/oem-flows/${id}`, data)
  const adminDeleteOemFlow = (id: string) => DELETE<any>(`/admin/oem-flows/${id}`)

  // Admin - OEM Solutions
  const adminGetOemSolutions = () => GET<any[]>('/admin/oem-solutions')
  const adminGetOemSolution = (id: string) => GET<any>(`/admin/oem-solutions/${id}`)
  const adminCreateOemSolution = (data: any) => POST<any>('/admin/oem-solutions', data)
  const adminUpdateOemSolution = (id: string, data: any) => PUT<any>(`/admin/oem-solutions/${id}`, data)
  const adminDeleteOemSolution = (id: string) => DELETE<any>(`/admin/oem-solutions/${id}`)

  // Admin - Price Lists
  const adminGetPriceLists = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/admin/price-lists', params)
  const adminCreatePriceList = (data: any) => POST<any>('/admin/price-lists', data)
  const adminUpdatePriceList = (id: string, data: any) => PUT<any>(`/admin/price-lists/${id}`, data)
  const adminDeletePriceList = (id: string) => DELETE<any>(`/admin/price-lists/${id}`)
  const adminGetProductPrices = (productId: string) => GET<any[]>(`/admin/products/${productId}/prices`)
  const adminSetProductPrice = (productId: string, data: any) => POST<any>(`/admin/products/${productId}/prices`, data)
  const adminAIGenerateProduct = (data: { description: string; language: string }) =>
    POST<any>('/admin/products/ai-generate', data, { timeout: 660_000 })

  /** AI 批量翻译可能耗时数分钟，超时与后端 WRITE_TIMEOUT 对齐 */
  const adminAITranslateProduct = (productId: string, data: { productId: string; targetLocales: string[] }) =>
    POST<any>(`/admin/products/${productId}/ai-translate`, data, { timeout: 660_000 })

  const adminAITranslateContent = (contentId: string, data: { contentId: string; contentType: string; targetLocales: string[] }) =>
    POST<any>(`/admin/content/${contentId}/ai-translate`, data, { timeout: 660_000 })

  // Admin - Inquiries
  const adminGetInquiry = (id: string) => GET<any>(`/admin/inquiries/${id}`)
  const adminConvertInquiryToOrder = (id: string) => POST<any>(`/admin/inquiries/${id}/convert-to-order`)
  const adminConfirmInquiry = (id: string, data: any) => PUT<any>(`/admin/inquiries/${id}/confirm`, data)
  const adminQuoteInquiry = (id: string, data: any) => POST<any>(`/admin/inquiries/${id}/quote`, data)
  const adminCreateNegotiation = (inquiryId: string, data: any) => POST<any>(`/admin/inquiries/${inquiryId}/negotiations`, data)
  const adminAcceptNegotiation = (inquiryId: string, offerId: string) => POST<any>(`/admin/inquiries/${inquiryId}/negotiations/${offerId}/accept`)
  const adminRejectNegotiation = (inquiryId: string, offerId: string) => POST<any>(`/admin/inquiries/${inquiryId}/negotiations/${offerId}/reject`)

  // Admin - Users
  const adminGetUser = (id: string) => GET<any>(`/admin/users/${id}`)
  const adminUpdateUserStatus = (id: string, status: string) => PUT<any>(`/admin/users/${id}/status`, { status })
  const adminUpdateUserRole = (id: string, role: string) => PUT<any>(`/admin/users/${id}/role`, { roleName: role })
  const adminUpdateUser = (id: string, data: any) => PUT<any>(`/admin/users/${id}`, data)
  const adminDeleteUser = (id: string) => DELETE<any>(`/admin/users/${id}`)

  // Customer - Inquiries
  const customerUploadInquiryAttachment = (inquiryId: string, formData: FormData) => {
    return fetchApi<any>(`/user/inquiries/${inquiryId}/attachments`, {
      method: 'POST',
      body: formData
    })
  }
  const customerConfirmInquiry = (inquiryId: string, data: any) => POST<any>(`/user/inquiries/${inquiryId}/confirm`, data)

  // Customer - OEM Projects
  const customerGetOemProjects = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/user/oem-projects', params)
  const customerCreateOemProject = (data: {
    productName: string
    inquiryId?: string
    notes?: string
    flavor?: string
    shape?: string
    packaging?: string
    targetMarket?: string
    certifications?: string[]
    moq?: number
    files?: File[]
  }) => {
    if (data.files?.length) {
      const formData = new FormData()
      formData.append('productName', data.productName)
      if (data.inquiryId) formData.append('inquiryId', data.inquiryId)
      if (data.notes) formData.append('notes', data.notes)
      if (data.flavor) formData.append('flavor', data.flavor)
      if (data.shape) formData.append('shape', data.shape)
      if (data.packaging) formData.append('packaging', data.packaging)
      if (data.targetMarket) formData.append('targetMarket', data.targetMarket)
      if (data.certifications?.length) formData.append('certifications', data.certifications.join(','))
      if (data.moq) formData.append('moq', String(data.moq))
      data.files.forEach((file) => formData.append('files', file))
      return fetchApi<any>('/user/oem-projects', { method: 'POST', body: formData })
    }
    return POST<any>('/user/oem-projects', {
      productName: data.productName,
      inquiryId: data.inquiryId,
      notes: data.notes,
      requirements: {
        flavor: data.flavor,
        shape: data.shape,
        packaging: data.packaging,
        targetMarket: data.targetMarket,
        certifications: data.certifications,
        moq: data.moq || 0
      }
    })
  }
  const customerUploadOemProjectAttachment = (projectId: string, formData: FormData) =>
    fetchApi<any>(`/user/oem-projects/${projectId}/attachments`, { method: 'POST', body: formData })
  const customerGetOemProject = (id: string) => GET<any>(`/user/oem-projects/${id}`)

  // Customer - Shipments
  const customerNudgeShipment = (tradeId: number | string, shipmentId: number | string) =>
    POST<any>(`/user/trades/${tradeId}/shipments/${shipmentId}/nudge`, {})
  const customerUploadShipmentAttachment = (
    tradeId: number | string,
    shipmentId: number | string,
    formData: FormData
  ) =>
    fetchApi<any>(`/user/trades/${tradeId}/shipments/${shipmentId}/attachments`, {
      method: 'POST',
      body: formData,
    })
  const customerGetShipmentTimeline = (tradeId: number | string, shipmentId: number | string) =>
    GET<any>(`/user/trades/${tradeId}/shipments/${shipmentId}/timeline`)

  // Customer - Payments
  const customerUploadPaymentProof = (orderId: string, formData: FormData) => {
    return fetchApi<any>(`/user/orders/${orderId}/payments`, {
      method: 'POST',
      body: formData
    })
  }

  const customerCreateGatewayPayment = (orderId: string, body: { method: string; amount?: number }) =>
    POST<any>(`/user/orders/${orderId}/payments/gateway`, body)

  /** 发送订单消息（支持纯文本 JSON 或 multipart 带附件） */
  const customerSendOrderMessage = (
    orderId: string,
    payload: { message?: string; files?: File[] }
  ) => {
    if (payload.files?.length) {
      const formData = new FormData()
      if (payload.message?.trim()) formData.append('message', payload.message.trim())
      payload.files.forEach((f) => formData.append('files', f))
      return fetchApi<any>(`/user/orders/${orderId}/messages`, { method: 'POST', body: formData })
    }
    return POST<any>(`/user/orders/${orderId}/messages`, { message: payload.message?.trim() || '' })
  }

  // Customer - Cart
  const getCart = (params?: { destination?: string; region?: string; incoterms?: string }) =>
    GET<any>('/user/cart', params)
  const addToCart = (productId: string, data: { quantity: number; unitPrice: number; specifications?: string }) => POST<any>('/user/cart/items', { productId, ...data })
  const updateCartItem = (itemId: string, data: { quantity: number }) => PUT<any>(`/user/cart/items/${itemId}`, data)
  const removeCartItem = (itemId: string) => DELETE<any>(`/user/cart/items/${itemId}`)
  const clearCart = () => DELETE<any>('/user/cart')
  const checkoutCart = (data: { shippingAddress?: any; couponCode?: string; incoterms?: string }) => POST<any>('/user/cart/checkout', data)
  const validateCartCoupon = (data: { code: string; subtotal?: number }) => POST<any>('/user/cart/coupon/validate', data)
  const applyCartCoupon = (data: { orderId: string; code: string }) => POST<any>('/user/cart/coupon', data)
  const removeCartCoupon = (data: { orderId: string }) => DELETE<any>('/user/cart/coupon', data)

  const getCustomerInvoices = () => GET<{ data: any[]; total: number }>('/user/invoices')

  // Customer - Notifications
  const markNotificationRead = (id: number) => PUT<any>(`/user/notifications/${id}/read`, {})
  const markAllNotificationsRead = () => POST<any>('/user/notifications/mark-all-read', {})

  // Admin - Inventory
  const getInventory = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/admin/inventory', params)
  const updateInventory = (productId: string, data: { stockQuantity: number; reason?: string; notes?: string }) => PUT<any>(`/admin/inventory/${productId}`, data)
  const exportInventoryXlsx = (ids: string[]) => {
    return fetchApi<Blob>('/admin/inventory/export-xlsx', {
      method: 'POST',
      body: JSON.stringify({ ids }),
      responseType: 'blob'
    })
  }
  const importInventoryXlsx = (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return fetchApi<any>('/admin/inventory/import-xlsx', {
      method: 'POST',
      body: formData
    })
  }
  const applyInventoryImport = (data: { rows: any[]; imageColumns: string[] }) => {
    return fetchApi<any>('/admin/inventory/import-xlsx/apply', { method: 'POST', body: JSON.stringify(data) })
  }

  // Admin - XLSX export & translate
  const exportAdminXlsx = (type: 'products' | 'orders' | 'revenue' | 'trades' | 'customers', params?: Record<string, string>) => {
    const qs = params ? `?${new URLSearchParams(params).toString()}` : ''
    return fetchApi<Blob>(`/admin/xlsx/export/${type}${qs}`, { responseType: 'blob' })
  }
  const translateAdminXlsx = (file: File, opts: { targetLang?: string; targetLangs?: string[]; sourceLang?: string; batch?: boolean }) => {
    const formData = new FormData()
    formData.append('file', file)
    if (opts.sourceLang) formData.append('sourceLang', opts.sourceLang)
    if (opts.batch && opts.targetLangs?.length) {
      formData.append('targetLangs', opts.targetLangs.join(','))
      return fetchApi<Blob>('/admin/xlsx/translate-batch', { method: 'POST', body: formData, responseType: 'blob', timeout: 660_000 })
    }
    if (opts.targetLang) formData.append('targetLang', opts.targetLang)
    return fetchApi<Blob>('/admin/xlsx/translate', { method: 'POST', body: formData, responseType: 'blob', timeout: 660_000 })
  }

  const batchUpdateInventory = (ids: string[], updates: Record<string, any>) => {
    return fetchApi<any>('/admin/inventory/batch-update', { method: 'POST', body: JSON.stringify({ ids, updates }) })
  }
  const batchDeleteInventory = (ids: string[]) => {
    return fetchApi<any>('/admin/inventory/batch-delete', { method: 'POST', body: JSON.stringify({ ids }) })
  }

  // Admin - Shipments
  const getShipments = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/admin/shipments', params)
  const createShipment = (data: any) => POST<any>('/admin/shipments', data)
  const updateShipment = (id: string, data: any) => PUT<any>(`/admin/shipments/${id}`, data)

  // Admin - Invoices
  const getInvoices = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/admin/invoices', params)
  const createInvoice = (data: any) => POST<any>('/admin/invoices', data)
  const updateInvoice = (id: string, data: any) => PUT<any>(`/admin/invoices/${id}`, data)

  // Admin - Translations
  const getTranslations = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/admin/translations', params)
  const getTranslation = (id: string) => GET<any>(`/admin/translations/${id}`)
  const createTranslation = (data: any) => POST<any>('/admin/translations', data)
  const updateTranslation = (id: string, data: any) => PUT<any>(`/admin/translations/${id}`, data)
  const deleteTranslation = (id: string) => DELETE<any>(`/admin/translations/${id}`)
  const getTranslationGroups = () => GET<any[]>('/admin/translations/groups')
  const importTranslations = (data: any) => POST<any>('/admin/translations/import', data)
  const exportTranslations = () => GET<any[]>('/admin/translations/export')

  // Admin - Certifications
  const adminGetCertifications = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/admin/certifications', params)
  const adminGetCertification = (id: string) => GET<any>(`/admin/certifications/${id}`)
  const adminCreateCertification = (data: any) => POST<any>('/admin/certifications', data)
  const adminUpdateCertification = (id: string, data: any) => PUT<any>(`/admin/certifications/${id}`, data)
  const adminDeleteCertification = (id: string) => DELETE<any>(`/admin/certifications/${id}`)

  const adminGetCategories = () => GET<any[]>('/admin/categories')
  const adminGetCategory = (slug: string) => GET<any>(`/admin/categories/${slug}`)
  const adminCreateCategory = (data: any) => POST<any>('/admin/categories', data)
  const adminUpdateCategory = (slug: string, data: any) => PUT<any>(`/admin/categories/${slug}`, data)
  const adminDeleteCategory = (slug: string) => DELETE<any>(`/admin/categories/${slug}`)

  // Admin - Content
  const adminGetContent = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/admin/content', params)
  const adminGetContentById = (id: string, type: string) => GET<any>(`/admin/content/${id}?type=${type}`)
  const adminCreateContent = (data: any) => POST<any>('/admin/content', data)
  const adminUpdateContent = (id: string, data: any) => PUT<any>(`/admin/content/${id}`, data)
  const adminDeleteContent = (id: string, type: string) => DELETE<any>(`/admin/content/${id}?type=${type}`)
  const adminAIGenerateContent = (data: { topic: string; type: string; language?: string; translate?: boolean }) => POST<any>('/admin/content/ai-generate', data)
  const adminAITranslateContentFields = (data: {
    contentType: string
    sourceLocale?: string
    targetLocales: string[]
    fields: Record<string, string>
  }) => POST<any>('/admin/content/ai-translate-fields', data, { timeout: 660_000 })
  const adminAIInlineEditContent = (data: {
    instruction: string
    selectedText: string
    selectedHtml?: string
    contextBefore?: string
    contextAfter?: string
    fieldType: string
    language?: string
  }) => POST<any>('/admin/content/ai-inline-edit', data)
  const adminAIReviseContent = (data: {
    type: string
    language?: string
    instruction: string
    selectedText?: string
    focusFields?: string[]
    current: Record<string, string>
  }) => POST<any>('/admin/content/ai-revise', data)

  return {
    // Products
    getCategories,
    getCategory,
    getProducts,
    getProduct,
    getFeaturedProducts,
    getRelatedProducts,

    // OEM
    getOEMFlows,
    getOEMFlow,
    getOEMSolutions,
    getOEMSolution,

    // Factory
    getFactoryInfo,
    getCertifications,
    getCertification,
    getProcessControls,
    getQualityTimeline,

    // Content
    getPosts,
    getPost,
    getRelatedPosts,
    getCases,
    getCase,
    getRelatedCases,

    // Actions
    submitInquiry,
    submitCustomerInquiry,

    // Customer Orders
    getOrders,
    getOrder,
    approveOrder,
    rejectOrder,
    createOrder,
    createAIAssistOrder,
    confirmOrder,

    // Search
    search,

    // Admin - Companies
    adminGetCompanies,
    adminGetCompany,
    adminUpdateCompany,
    adminVerifyCompany,

    // Admin - Order Payments
    adminGetOrderPayments,
    adminCreatePayment,
    adminConfirmPayment,
    adminRefundPayment,

    // Admin - Trades
    adminGetTrades,
    adminGetTrade,
    adminUpdateTradeStatus,
    adminGetTradeDocuments,
    adminUpdateTradeDocument,

    // Admin - OEM Projects
    adminGetOemProjects,
    adminGetOemProject,
    adminUpdateOemProject,
    adminUpdateOemStatus,
    adminAddOemSample,
    adminUpdateOemSample,
    adminConvertOemProjectToOrder,
    adminCreateOemInventoryHold,
    adminListOemInventoryHolds,
    adminReleaseOemInventoryHold,

    // Admin - OEM Flows
    adminGetOemFlows,
    adminGetOemFlow,
    adminCreateOemFlow,
    adminUpdateOemFlow,
    adminDeleteOemFlow,

    // Admin - OEM Solutions
    adminGetOemSolutions,
    adminGetOemSolution,
    adminCreateOemSolution,
    adminUpdateOemSolution,
    adminDeleteOemSolution,

    // Admin - Price Lists
    adminGetPriceLists,
    adminCreatePriceList,
    adminUpdatePriceList,
    adminDeletePriceList,
    adminGetProductPrices,
    adminSetProductPrice,
    adminAIGenerateProduct,
    adminAITranslateProduct,
    adminAITranslateContent,

    // Admin - Inquiries
    adminGetInquiry,
    adminConvertInquiryToOrder,
    adminConfirmInquiry,
    adminQuoteInquiry,
    adminCreateNegotiation,
    adminAcceptNegotiation,
    adminRejectNegotiation,
    adminGetUser,
    adminUpdateUserStatus,
    adminUpdateUserRole,
    adminUpdateUser,
    adminDeleteUser,
    customerUploadInquiryAttachment,
    customerConfirmInquiry,

    // Customer - OEM Projects
    customerGetOemProjects,
    customerCreateOemProject,
    customerUploadOemProjectAttachment,
    customerGetOemProject,

    // Customer - Shipments
    customerNudgeShipment,
    customerUploadShipmentAttachment,
    customerGetShipmentTimeline,

    // Customer - Payments
    customerUploadPaymentProof,
    customerCreateGatewayPayment,
    customerSendOrderMessage,

    // Customer - Cart
    getCart,
    addToCart,
    updateCartItem,
    removeCartItem,
    clearCart,
    checkoutCart,
    validateCartCoupon,
    applyCartCoupon,
    removeCartCoupon,
    getCustomerInvoices,

    // Customer - Notifications
    markNotificationRead,
    markAllNotificationsRead,

    // Admin - Inventory
    getInventory,
    updateInventory,
    exportInventoryXlsx,
    importInventoryXlsx,
    applyInventoryImport,
    exportAdminXlsx,
    translateAdminXlsx,
    batchUpdateInventory,
    batchDeleteInventory,

    // Admin - Shipments
    getShipments,
    createShipment,
    updateShipment,

    // Admin - Invoices
    getInvoices,
    createInvoice,
    updateInvoice,

    // Admin - Translations
    getTranslations,
    getTranslation,
    createTranslation,
    updateTranslation,
    deleteTranslation,
    getTranslationGroups,
    importTranslations,
    exportTranslations,

    // Admin - Certifications
    adminGetCertifications,
    adminGetCertification,
    adminCreateCertification,
    adminUpdateCertification,
    adminDeleteCertification,

    // Admin - Categories
    adminGetCategories,
    adminGetCategory,
    adminCreateCategory,
    adminUpdateCategory,
    adminDeleteCategory,

    // Admin - Content
    adminGetContent,
    adminGetContentById,
    adminCreateContent,
    adminUpdateContent,
    adminDeleteContent,
    adminAIGenerateContent,
    adminAITranslateContentFields,
    adminAIInlineEditContent,
    adminAIReviseContent,

    // Generic HTTP helpers
    fetchApi,
    get: GET,
    post: POST,
    put: PUT,
    del: DELETE
  }
}
