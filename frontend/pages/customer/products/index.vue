<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
    <!-- Header -->
    <div class="bg-white border-b border-orange-100 shadow-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="text-3xl font-bold bg-gradient-to-r from-orange-500 to-amber-600 bg-clip-text text-transparent">
              {{ t('customer.products.title') }}
            </h1>
            <p class="mt-1 text-sm text-gray-600">{{ t('customer.products.subtitle') }}</p>
          </div>
          <div class="flex items-center gap-3">
            <div class="relative">
              <Icon name="heroicons:magnifying-glass" class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
              <input v-model="searchQuery" type="text" :placeholder="t('customer.products.search')" class="pl-10 pr-4 py-2 w-64 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-orange-500 text-sm" />
            </div>
            <button @click="showFilters = !showFilters" class="p-2 border border-gray-200 rounded-xl hover:bg-gray-50 transition-colors">
              <Icon name="heroicons:adjustments" class="h-5 w-5 text-gray-600" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- AI 选品面板 -->
    <div class="bg-gradient-to-r from-orange-500 to-amber-600 text-white">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div class="flex items-center gap-3">
          <button @click="showAIPanel = !showAIPanel"
            class="flex items-center gap-2 px-4 py-2 bg-white/20 hover:bg-white/30 rounded-xl text-sm font-medium transition-colors">
            <Icon name="heroicons:sparkles" class="h-4 w-4" />
            {{ t('customer.products.ai_select') }}
            <Icon :name="showAIPanel ? 'heroicons:chevron-up' : 'heroicons:chevron-down'" class="h-4 w-4" />
          </button>
          <span class="text-blue-200 text-sm">{{ t('customer.products.ai_select_hint') }}</span>
        </div>

        <Transition name="slide-down">
          <div v-if="showAIPanel" class="mt-4 bg-white/10 rounded-2xl p-5 space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div class="md:col-span-2">
                <textarea v-model="aiForm.prompt" rows="2"
                  :placeholder="t('customer.products.ai_prompt_placeholder')"
                  class="w-full px-4 py-2.5 rounded-xl bg-white/20 placeholder-blue-200 text-white border border-white/30 focus:outline-none focus:ring-2 focus:ring-white/50 text-sm resize-none" />
              </div>
              <div class="space-y-2">
                <input v-model="aiForm.targetCountry" type="text"
                  :placeholder="t('customer.products.ai_country_placeholder')"
                  class="w-full px-3 py-2 rounded-xl bg-white/20 placeholder-blue-200 text-white border border-white/30 focus:outline-none focus:ring-2 focus:ring-white/50 text-sm" />
                <div class="flex gap-2">
                  <input v-model.number="aiForm.quantity" type="number" min="1"
                    :placeholder="t('customer.products.ai_qty_placeholder')"
                    class="w-1/3 px-3 py-2 rounded-xl bg-white/20 placeholder-blue-200 text-white border border-white/30 focus:outline-none focus:ring-2 focus:ring-white/50 text-sm" />
                  <input v-model.number="aiForm.budget" type="number" min="0"
                    :placeholder="t('customer.products.ai_budget_placeholder')"
                    class="w-1/3 px-3 py-2 rounded-xl bg-white/20 placeholder-blue-200 text-white border border-white/30 focus:outline-none focus:ring-2 focus:ring-white/50 text-sm" />
                  <input v-model="aiForm.currency" type="text" maxlength="3"
                    :placeholder="cur()"
                    class="w-1/3 px-3 py-2 rounded-xl bg-white/20 placeholder-blue-200 text-white border border-white/30 focus:outline-none focus:ring-2 focus:ring-white/50 text-sm uppercase" />
                </div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <button @click="runAIRecommend" :disabled="aiLoading || !aiForm.prompt.trim()"
                class="px-5 py-2 bg-white text-orange-600 font-semibold rounded-xl hover:bg-orange-50 disabled:opacity-50 transition-colors flex items-center gap-2 text-sm">
                <Icon v-if="aiLoading" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" />
                <Icon v-else name="heroicons:sparkles" class="h-4 w-4" />
                {{ aiLoading ? t('customer.products.ai_thinking') : t('customer.products.ai_recommend') }}
              </button>
              <button v-if="aiResults.length" @click="inquireAllAI"
                class="px-5 py-2 bg-emerald-500 text-white font-semibold rounded-xl hover:bg-emerald-600 transition-colors flex items-center gap-2 text-sm">
                <Icon name="heroicons:chat-bubble-left" class="h-4 w-4" />
                {{ t('customer.products.ai_inquire_all', { count: aiResults.length }) }}
              </button>
              <p v-if="aiError" class="text-red-300 text-sm">{{ aiError }}</p>
            </div>

            <!-- AI 推荐结果 -->
            <div v-if="aiResults.length" class="space-y-3">
              <p class="text-blue-100 text-sm font-medium">{{ aiSummary }}</p>
              <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                <div v-for="item in aiResults" :key="item.id"
                  class="bg-white rounded-xl p-4 text-gray-900 flex flex-col gap-2">
                  <div class="flex items-start justify-between gap-2">
                    <div>
                      <p class="font-semibold text-sm">{{ item.name }}</p>
                      <p class="text-xs text-gray-500">{{ item.category }}</p>
                    </div>
                    <div class="flex gap-1 flex-shrink-0">
                      <span v-if="item.halalCertified" class="px-1.5 py-0.5 bg-emerald-100 text-emerald-700 text-xs rounded-full">Halal</span>
                      <span v-if="item.oemAvailable" class="px-1.5 py-0.5 bg-orange-100 text-orange-700 text-xs rounded-full">OEM</span>
                    </div>
                  </div>
                  <p class="text-xs text-gray-600 italic">{{ item.reason }}</p>
                  <div class="flex items-center gap-2 mt-auto pt-2 border-t border-gray-100">
                    <div class="flex items-center gap-1">
                      <button @click="item.quantity = Math.max(item.moq || 1, item.quantity - 1)"
                        class="w-7 h-7 flex items-center justify-center rounded-lg border border-gray-200 hover:bg-gray-50 text-gray-600">
                        <Icon name="heroicons:minus" class="h-3 w-3" />
                      </button>
                      <input v-model.number="item.quantity" type="number" :min="item.moq || 1"
                        class="w-14 h-7 text-center text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-1 focus:ring-orange-500" />
                      <button @click="item.quantity += 1"
                        class="w-7 h-7 flex items-center justify-center rounded-lg border border-gray-200 hover:bg-gray-50 text-gray-600">
                        <Icon name="heroicons:plus" class="h-3 w-3" />
                      </button>
                    </div>
                    <NuxtLink :to="`${localePath('/customer/inquiries/new')}?product=${item.id}&name=${encodeURIComponent(item.name)}`"
                      class="ml-auto px-3 py-1.5 bg-orange-500 text-white text-xs font-medium rounded-lg hover:bg-orange-600 transition-colors">
                      {{ t('customer.products.inquire') }}
                    </NuxtLink>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Transition>
      </div>
    </div>

    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div class="flex gap-8">
        <!-- Filters Sidebar -->
        <aside v-if="showFilters" class="w-64 flex-shrink-0">
          <div class="bg-white rounded-2xl shadow-sm border border-orange-100 p-5 sticky top-8">
            <div class="flex items-center justify-between mb-4">
              <h3 class="font-semibold text-gray-900">{{ t('customer.products.filters') }}</h3>
              <button @click="showFilters = false" class="text-gray-400 hover:text-gray-600">
                <Icon name="heroicons:x-mark" class="h-5 w-5" />
              </button>
            </div>

            <!-- Categories -->
            <div class="mb-6">
              <h4 class="text-sm font-medium text-gray-700 mb-3">{{ t('customer.products.categories') }}</h4>
              <div class="space-y-2">
                <label v-for="cat in categories" :key="cat.slug" class="flex items-center gap-2 cursor-pointer">
                  <input v-model="selectedCategories" :value="cat.slug" type="checkbox" class="rounded border-gray-300 text-orange-600 focus:ring-orange-500" />
                  <span class="text-sm text-gray-600">{{ cat.name }}</span>
                  <span class="text-xs text-gray-400 ml-auto">({{ cat.productCount }})</span>
                </label>
              </div>
            </div>

            <!-- Certifications -->
            <div class="mb-6">
              <h4 class="text-sm font-medium text-gray-700 mb-3">{{ t('customer.products.certifications') }}</h4>
              <div class="space-y-2">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input v-model="filters.halal" type="checkbox" class="rounded border-gray-300 text-orange-600 focus:ring-orange-500" />
                  <span class="text-sm text-gray-600">Halal Certified</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input v-model="filters.oemOnly" type="checkbox" class="rounded border-gray-300 text-orange-600 focus:ring-orange-500" />
                  <span class="text-sm text-gray-600">OEM Available</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input v-model="filters.featuredOnly" type="checkbox" class="rounded border-gray-300 text-orange-600 focus:ring-orange-500" />
                  <span class="text-sm text-gray-600">Featured</span>
                </label>
              </div>
            </div>

            <!-- MOQ Range -->
            <div class="mb-6">
              <h4 class="text-sm font-medium text-gray-700 mb-3">{{ t('customer.products.moq_range') }}</h4>
              <div class="flex items-center gap-2">
                <input v-model.number="filters.minMoq" type="number" min="0" :placeholder="t('customer.products.min')" class="w-full px-3 py-1.5 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
                <span class="text-gray-400">-</span>
                <input v-model.number="filters.maxMoq" type="number" min="0" :placeholder="t('customer.products.max')" class="w-full px-3 py-1.5 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
              </div>
            </div>

            <!-- Apply/Clear -->
            <div class="flex gap-2">
              <button @click="applyFilters" class="flex-1 py-2 bg-orange-600 text-white text-sm font-medium rounded-lg hover:bg-orange-700 transition-colors">
                {{ t('customer.products.apply') }}
              </button>
              <button @click="clearFilters" class="px-3 py-2 border border-gray-200 text-gray-600 text-sm rounded-lg hover:bg-gray-50 transition-colors">
                {{ t('customer.products.clear') }}
              </button>
            </div>
          </div>
        </aside>

        <!-- Product Grid -->
        <div class="flex-1">
          <div v-if="pending" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            <div v-for="i in 8" :key="i" class="bg-white rounded-2xl shadow-sm border border-orange-100 overflow-hidden animate-pulse">
              <div class="h-48 bg-gray-200"></div>
              <div class="p-4 space-y-3">
                <div class="h-4 bg-gray-200 rounded w-3/4"></div>
                <div class="h-3 bg-gray-200 rounded w-1/2"></div>
                <div class="h-3 bg-gray-200 rounded w-1/4"></div>
              </div>
            </div>
          </div>

          <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-xl p-6 text-center">
            <Icon name="heroicons:exclamation-triangle" class="h-12 w-12 text-red-400 mx-auto mb-3" />
            <p class="text-red-700">{{ error }}</p>
            <button @click="fetchProducts" class="mt-4 px-4 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200">
              {{ t('customer.products.retry') }}
            </button>
          </div>

          <div v-else-if="products.length === 0" class="text-center py-20">
            <Icon name="heroicons:cube" class="h-16 w-16 text-gray-300 mx-auto mb-4" />
            <p class="text-gray-600">{{ t('customer.products.no_products') }}</p>
          </div>

          <div v-else>
            <!-- Results Header -->
            <div class="flex items-center justify-between mb-6">
              <p class="text-sm text-gray-600">
                {{ t('customer.products.showing', { count: products.length, total: pagination?.total || 0 }) }}
              </p>
              <div class="flex items-center gap-2">
                <span class="text-sm text-gray-500">{{ t('customer.products.sort_by') }}</span>
                <select v-model="sortBy" @change="fetchProducts" class="px-3 py-1.5 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500">
                  <option value="popular">{{ t('customer.products.sort_popular') }}</option>
                  <option value="name">{{ t('customer.products.sort_name') }}</option>
                  <option value="moq_low">{{ t('customer.products.sort_moq_low') }}</option>
                  <option value="moq_high">{{ t('customer.products.sort_moq_high') }}</option>
                </select>
              </div>
            </div>

            <!-- Product Grid -->
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
              <div v-for="product in products" :key="product.id" class="group bg-white rounded-2xl shadow-sm border border-orange-100 overflow-hidden hover:shadow-xl hover:border-blue-200 transition-shadow duration-300">
                <!-- Product Image -->
                <div class="relative h-48 bg-gradient-to-br from-orange-50 to-amber-50 overflow-hidden">
                  <img v-if="product.thumbnail" :src="product.thumbnail" :alt="product.name" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" />
                  <div v-else class="w-full h-full flex items-center justify-center">
                    <Icon name="heroicons:cube" class="h-16 w-16 text-blue-200" />
                  </div>
                  <!-- Badges -->
                  <div class="absolute top-3 left-3 flex flex-col gap-2">
                    <span v-if="product.featured" class="px-2 py-1 bg-amber-400 text-amber-900 text-xs font-bold rounded-full">
                      {{ t('customer.products.featured') }}
                    </span>
                    <span v-if="product.halalCertified" class="px-2 py-1 bg-emerald-100 text-emerald-700 text-xs font-medium rounded-full">
                      Halal
                    </span>
                  </div>
                  <!-- OEM Badge -->
                  <div v-if="product.oemAvailable" class="absolute top-3 right-3">
                    <span class="px-2 py-1 bg-orange-600 text-white text-xs font-medium rounded-full flex items-center gap-1">
                      <Icon name="heroicons:sparkles" class="h-3 w-3" />
                      OEM
                    </span>
                  </div>
                </div>

                <!-- Product Info -->
                <div class="p-4">
                  <div class="text-xs text-orange-600 font-medium mb-1">{{ product.category }}</div>
                  <h3 class="font-semibold text-gray-900 mb-1 line-clamp-1">{{ product.name }}</h3>
                  <p class="text-sm text-gray-500 mb-3 line-clamp-2">{{ product.summary || product.description?.substring(0, 80) }}</p>

                  <!-- MOQ & Lead Time -->
                  <div class="flex items-center gap-3 text-xs text-gray-500 mb-3">
                    <span class="flex items-center gap-1">
                      <Icon name="heroicons:cube" class="h-3 w-3" />
                      MOQ: {{ product.moq || 1 }}
                    </span>
                    <span v-if="product.leadTime" class="flex items-center gap-1">
                      <Icon name="heroicons:clock" class="h-3 w-3" />
                      {{ product.leadTime }}
                    </span>
                  </div>

                  <!-- Flavors & Shapes -->
                  <div v-if="product.flavors?.length || product.shapes?.length" class="flex flex-wrap gap-1 mb-3">
                    <span v-for="flavor in product.flavors?.slice(0, 3)" :key="flavor" class="px-2 py-0.5 bg-gray-100 text-gray-600 text-xs rounded-full">
                      {{ flavor }}
                    </span>
                    <span v-for="shape in product.shapes?.slice(0, 2)" :key="shape" class="px-2 py-0.5 bg-orange-50 text-orange-600 text-xs rounded-full">
                      {{ shape }}
                    </span>
                    <span v-if="(product.flavors?.length || 0) > 3 || (product.shapes?.length || 0) > 2" class="px-2 py-0.5 bg-gray-50 text-gray-400 text-xs rounded-full">
                      + more
                    </span>
                  </div>

                  <!-- Price & Action -->
                  <div class="flex items-center justify-between pt-3 border-t border-gray-100">
                    <div>
                      <p class="text-xs text-gray-500">{{ t('customer.products.unit_price') }}</p>
                      <p class="text-lg font-bold text-orange-600">
                        {{ currency.formatPrice(product.basePrice || 0) }}
                      </p>
                    </div>
                    <div class="flex gap-2">
                      <NuxtLink :to="`${localePath('/customer/inquiries/new')}?product=${product.id}&name=${encodeURIComponent(product.name)}`" class="p-2 bg-orange-50 text-orange-600 rounded-lg hover:bg-orange-100 transition-colors" :title="t('customer.products.inquire')">
                        <Icon name="heroicons:chat-bubble-left" class="h-5 w-5" />
                      </NuxtLink>
                      <NuxtLink :to="localePath(`/customer/products/${product.slug || product.id}`)" class="px-4 py-2 bg-orange-600 text-white text-sm font-medium rounded-lg hover:bg-orange-700 transition-colors">
                        {{ t('customer.products.view_details') }}
                      </NuxtLink>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Pagination -->
            <div v-if="pagination && pagination.totalPages > 1" class="mt-8 flex items-center justify-center gap-2">
              <button @click="prevPage" :disabled="page <= 1" class="p-2 border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
                <Icon name="heroicons:chevron-left" class="h-5 w-5" />
              </button>
              <div class="flex items-center gap-1">
                <button v-for="p in visiblePages" :key="p" @click="goToPage(p)" class="w-10 h-10 flex items-center justify-center rounded-lg text-sm font-medium transition-colors" :class="p === page ? 'bg-orange-600 text-white' : 'border border-gray-200 hover:bg-gray-50'">
                  {{ p }}
                </button>
              </div>
              <button @click="nextPage" :disabled="page >= pagination.totalPages" class="p-2 border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
                <Icon name="heroicons:chevron-right" class="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Notification Toast -->
    <Transition name="slide-up">
      <div v-if="showToast" class="fixed bottom-6 right-6 bg-emerald-600 text-white px-6 py-3 rounded-xl shadow-lg flex items-center gap-3 z-50">
        <Icon name="heroicons:check-circle" class="h-5 w-5" />
        {{ toastMessage }}
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'

definePageMeta({
  layout: 'customer',
  middleware: ['auth']
})

const { t } = useI18n()
const localePath = useLocalePath()
const { formatNumber } = useDisplay()
const currency = useCurrency()
const api = useApi()

const products = ref<any[]>([])
const categories = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const showFilters = ref(true)
const showToast = ref(false)
const toastMessage = ref('')

// AI 选品
const showAIPanel = ref(false)
const aiLoading = ref(false)
const aiError = ref('')
const aiResults = ref<any[]>([])
const aiSummary = ref('')
const addingAll = ref(false)
const aiForm = reactive({
  prompt: '',
  targetCountry: '',
  quantity: 1,
  budget: 0,
  currency: cur(null),
})

const runAIRecommend = async () => {
  if (!aiForm.prompt.trim()) return
  aiLoading.value = true
  aiError.value = ''
  aiResults.value = []
  aiSummary.value = ''
  try {
    const res = await api.post<any>('/user/ai/recommend', {
      prompt: aiForm.prompt.trim(),
      targetCountry: aiForm.targetCountry.trim() || 'global',
      quantity: Math.max(1, aiForm.quantity || 1),
      budget: aiForm.budget || 0,
      currency: cur(aiForm.currency).toUpperCase(),
    })
    aiResults.value = (res.recommendations || []).map((item: any) => ({
      ...item,
      quantity: item.quantity || 1,
    }))
    aiSummary.value = res.summary || ''
    if (!aiResults.value.length) {
      aiError.value = t('customer.products.ai_no_results')
    }
  } catch (err: any) {
    aiError.value = err?.message || t('customer.products.ai_error')
  } finally {
    aiLoading.value = false
  }
}

const inquireAllAI = () => {
  const names = aiResults.value.map(item => item.name).join(', ')
  navigateTo(`${localePath('/customer/inquiries/new')}?products=${encodeURIComponent(names)}`)
}

const searchQuery = ref('')
const selectedCategories = ref<string[]>([])
const sortBy = ref('popular')
const page = ref(1)
const pageSize = 24

const filters = reactive({
  halal: false,
  oemOnly: false,
  featuredOnly: false,
  minMoq: null as number | null,
  maxMoq: null as number | null
})

const visiblePages = computed(() => {
  if (!pagination.value) return []
  const total = pagination.value.totalPages
  const current = page.value
  const pages: number[] = []
  for (let i = Math.max(1, current - 2); i <= Math.min(total, current + 2); i++) {
    pages.push(i)
  }
  return pages
})

const fetchProducts = async () => {
  pending.value = true
  error.value = ''
  try {
    const params: Record<string, any> = {
      page: page.value,
      limit: pageSize,
      sort: sortBy.value
    }
    if (searchQuery.value) params.search = searchQuery.value
    if (selectedCategories.value.length) params.categories = selectedCategories.value.join(',')
    if (filters.halal) params.halal = true
    if (filters.oemOnly) params.oemOnly = true
    if (filters.featuredOnly) params.featured = true
    if (filters.minMoq) params.minMoq = filters.minMoq
    if (filters.maxMoq) params.maxMoq = filters.maxMoq

    const res = await api.getProducts(params)
    products.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('customer.products.fetch_error')
  } finally {
    pending.value = false
  }
}

const fetchCategories = async () => {
  try {
    const res = await api.getCategories()
    categories.value = res || []
  } catch {
    categories.value = []
  }
}

const applyFilters = () => {
  page.value = 1
  fetchProducts()
}

const clearFilters = () => {
  searchQuery.value = ''
  selectedCategories.value = []
  filters.halal = false
  filters.oemOnly = false
  filters.featuredOnly = false
  filters.minMoq = null
  filters.maxMoq = null
  sortBy.value = 'popular'
  page.value = 1
  fetchProducts()
}

const prevPage = () => {
  if (page.value > 1) { page.value -= 1; fetchProducts() }
}

const nextPage = () => {
  if (pagination.value && page.value < pagination.value.totalPages) {
    page.value += 1; fetchProducts()
  }
}

const goToPage = (p: number) => {
  page.value = p; fetchProducts()
}

const showNotification = (message: string, isError = false) => {
  toastMessage.value = message
  showToast.value = true
  setTimeout(() => { showToast.value = false }, 3000)
}

// Debounced search
let searchTimeout: NodeJS.Timeout
watch(searchQuery, () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    page.value = 1
    fetchProducts()
  }, 300)
})

onMounted(() => {
  fetchProducts()
  fetchCategories()
})
</script>

<style scoped>
.line-clamp-1 { display: -webkit-box; -webkit-line-clamp: 1; -webkit-box-orient: vertical; overflow: hidden; }
.line-clamp-2 { display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

.slide-up-enter-active, .slide-up-leave-active { transition: opacity 0.3s ease, transform 0.3s ease; }
.slide-up-enter-from, .slide-up-leave-to { opacity: 0; transform: translateY(20px); }

.slide-down-enter-active, .slide-down-leave-active { transition: opacity 0.3s ease, transform 0.3s ease; }
.slide-down-enter-from, .slide-down-leave-to { opacity: 0; transform: translateY(-8px); }
</style>
