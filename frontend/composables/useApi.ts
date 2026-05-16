/**
 * API Client Composable
 * Provides a typed wrapper for all API calls with error handling
 */

interface ProductQuery {
  category?: string
  page?: number
  limit?: number
  sort?: 'name' | 'created' | 'popular'
  order?: 'asc' | 'desc'
  search?: string
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

export const useApi = () => {
  const config = useRuntimeConfig()
  // During SSR, $fetch uses Nitro's localFetch which bypasses devProxy.
  // Use the direct backend URL on the server so requests don't fall through to Vue Router.
  const baseURL = import.meta.server
    ? config.internalApiBase
    : config.public.apiBase || '/api/v1'
  const publicBaseURL = `${baseURL}/public`
  const { t, locale } = useI18n()
  const authToken = useCookie<string | null>('auth_token')

  /**
   * Generic fetch wrapper with error handling
   */
  const fetchApi = async <T>(
    endpoint: string,
    options?: Record<string, unknown>
  ): Promise<T> => {
    try {
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        'Accept-Language': locale.value,
        ...(options?.headers as Record<string, string> ?? {})
      }

      if (authToken.value) {
        headers['Authorization'] = `Bearer ${authToken.value}`
      }

      const response = await $fetch<T>(`${baseURL}${endpoint}`, {
        ...options,
        headers
      })
      return response as T
    } catch (err: unknown) {
      const error = err as { statusCode?: number; data?: { message?: string } }
      const apiError: ApiError = {
        message: t('errors.default'),
        statusCode: error?.statusCode
      }

      if (error?.data?.message) {
        apiError.message = error.data.message
      }
      if (error?.data?.details) {
        apiError.details = error.data.details
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
    if (params.page) queryParams.append('page', params.page.toString())
    if (params.limit) queryParams.append('limit', params.limit.toString())
    if (params.sort) queryParams.append('sort', params.sort)
    if (params.order) queryParams.append('order', params.order)
    if (params.search) queryParams.append('search', params.search)
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
    // Use FormData to support file attachments
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
    return $fetch<InquiryResponse>(`${baseURL}/user/inquiries`, {
      method: 'POST',
      body: formData,
      headers: authToken.value ? { Authorization: `Bearer ${authToken.value}` } : {}
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
    return fetchApi<any>(`/user/orders/${id}`)
  }

  const createOrder = async (data: {
    items: { productId: string; quantity: number; unitPrice: number; specifications?: string }[]
    currency?: string
    taxAmount?: number
    shippingAmount?: number
    shippingAddress: { street: string; city: string; state?: string; zipCode?: string; country: string }
    inquiryId?: string
  }): Promise<any> => {
    return fetchApi<any>('/user/orders', { method: 'POST', body: JSON.stringify(data) })
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
    return fetchApi<any>('/user/orders/ai-assist', { method: 'POST', body: JSON.stringify(data) })
  }

  const confirmOrder = async (id: string, complianceAck: boolean): Promise<any> => {
    return fetchApi<any>(`/user/orders/${id}/confirm`, { method: 'POST', body: JSON.stringify({ complianceAck }) })
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

  const POST = <T>(endpoint: string, data?: any): Promise<T> => {
    return fetchApi<T>(endpoint, { method: 'POST', body: JSON.stringify(data) })
  }

  const PUT = <T>(endpoint: string, data?: any): Promise<T> => {
    return fetchApi<T>(endpoint, { method: 'PUT', body: JSON.stringify(data) })
  }

  const DELETE = <T>(endpoint: string): Promise<T> => {
    return fetchApi<T>(endpoint, { method: 'DELETE' })
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

  // Admin - Inquiries
  const adminGetInquiry = (id: string) => GET<any>(`/admin/inquiries/${id}`)
  const adminConvertInquiryToOrder = (id: string) => POST<any>(`/admin/inquiries/${id}/convert-to-order`)
  const adminConfirmInquiry = (id: string, data: any) => PUT<any>(`/admin/inquiries/${id}/confirm`, data)

  // Customer - Inquiries
  const customerUploadInquiryAttachment = (inquiryId: string, formData: FormData) => {
    return $fetch<any>(`${baseURL}/user/inquiries/${inquiryId}/attachments`, {
      method: 'POST',
      body: formData,
      headers: authToken.value ? { Authorization: `Bearer ${authToken.value}` } : {}
    })
  }
  const customerConfirmInquiry = (inquiryId: string, data: any) => POST<any>(`/user/inquiries/${inquiryId}/confirm`, data)

  // Customer - OEM Projects
  const customerGetOemProjects = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/user/oem-projects', params)
  const customerCreateOemProject = (data: any) => POST<any>('/user/oem-projects', data)
  const customerGetOemProject = (id: string) => GET<any>(`/user/oem-projects/${id}`)

  // Customer - Payments
  const customerUploadPaymentProof = (orderId: string, formData: FormData) => {
    return $fetch<any>(`${baseURL}/user/orders/${orderId}/payments`, {
      method: 'POST',
      body: formData,
      headers: authToken.value ? { Authorization: `Bearer ${authToken.value}` } : {}
    })
  }

  // Customer - Cart
  const getCart = () => GET<any>('/user/cart')
  const addToCart = (productId: string, data: { quantity: number; unitPrice: number; specifications?: string }) => POST<any>('/user/cart/items', { productId, ...data })
  const updateCartItem = (itemId: string, data: { quantity: number }) => PUT<any>(`/user/cart/items/${itemId}`, data)
  const removeCartItem = (itemId: string) => DELETE<any>(`/user/cart/items/${itemId}`)
  const clearCart = () => DELETE<any>('/user/cart')
  const checkoutCart = (data: { shippingAddress?: any }) => POST<any>('/user/cart/checkout', data)

  // Customer - Notifications
  const markNotificationRead = (id: number) => PUT<any>(`/user/notifications/${id}/read`, {})
  const markAllNotificationsRead = () => POST<any>('/user/notifications/mark-all-read', {})

  // Admin - Inventory
  const getInventory = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/admin/inventory', params)
  const updateInventory = (productId: string, data: { stockQuantity: number; reason?: string; notes?: string }) => PUT<any>(`/admin/inventory/${productId}`, data)
  const exportInventoryXlsx = (ids: string[]) => {
    return fetchApi<Blob>('/admin/inventory/export-xlsx', { method: 'POST', body: JSON.stringify({ ids }) }).then(async (_res) => {
      // Actually fetch as blob for download
      const resp = await $fetch<Blob>(`${baseURL}/admin/inventory/export-xlsx`, {
        method: 'POST',
        body: { ids },
        headers: authToken.value ? { Authorization: `Bearer ${authToken.value}` } : {},
        responseType: 'blob'
      })
      return resp
    })
  }
  const importInventoryXlsx = (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return $fetch<any>(`${baseURL}/admin/inventory/import-xlsx`, {
      method: 'POST',
      body: formData,
      headers: authToken.value ? { Authorization: `Bearer ${authToken.value}` } : {}
    })
  }
  const applyInventoryImport = (data: { rows: any[]; imageColumns: string[] }) => {
    return fetchApi<any>('/admin/inventory/import-xlsx/apply', { method: 'POST', body: JSON.stringify(data) })
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

  // Admin - Content
  const adminGetContent = (params?: Record<string, any>) => GET<PaginatedResponse<any>>('/admin/content', params)
  const adminGetContentById = (id: string, type: string) => GET<any>(`/admin/content/${id}?type=${type}`)
  const adminCreateContent = (data: any) => POST<any>('/admin/content', data)
  const adminUpdateContent = (id: string, data: any) => PUT<any>(`/admin/content/${id}`, data)
  const adminDeleteContent = (id: string, type: string) => DELETE<any>(`/admin/content/${id}?type=${type}`)
  const adminAIGenerateContent = (data: { topic: string; type: string; language: string }) => POST<any>('/admin/content/ai-generate', data)

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

    // Admin - Inquiries
    adminGetInquiry,
    adminConvertInquiryToOrder,
    adminConfirmInquiry,
    customerUploadInquiryAttachment,
    customerConfirmInquiry,

    // Customer - OEM Projects
    customerGetOemProjects,
    customerCreateOemProject,
    customerGetOemProject,

    // Customer - Payments
    customerUploadPaymentProof,

    // Customer - Cart
    getCart,
    addToCart,
    updateCartItem,
    removeCartItem,
    clearCart,
    checkoutCart,

    // Customer - Notifications
    markNotificationRead,
    markAllNotificationsRead,

    // Admin - Inventory
    getInventory,
    updateInventory,
    exportInventoryXlsx,
    importInventoryXlsx,
    applyInventoryImport,
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

    // Admin - Content
    adminGetContent,
    adminGetContentById,
    adminCreateContent,
    adminUpdateContent,
    adminDeleteContent,
    adminAIGenerateContent,

    // Generic HTTP helpers
    fetchApi,
    get: GET,
    post: POST,
    put: PUT,
    del: DELETE
  }
}
