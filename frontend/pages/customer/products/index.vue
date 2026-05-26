<template>
  <div class="catalog">
    <!-- 页头 -->
    <header class="catalog-header">
      <div class="catalog-header__text">
        <h1 class="catalog-header__title">{{ t('customer.products.title') }}</h1>
        <p class="catalog-header__subtitle">{{ t('customer.products.subtitle') }}</p>
      </div>
      <div class="catalog-header__actions">
        <div class="catalog-search">
          <Icon name="heroicons:magnifying-glass" class="catalog-search__icon" aria-hidden="true" />
          <input
            v-model="searchQuery"
            type="search"
            :placeholder="t('customer.products.search')"
            class="catalog-search__input"
          />
        </div>
        <button
          type="button"
          class="catalog-toolbar-btn"
          :class="{ 'catalog-toolbar-btn--active': showFilters }"
          @click="showFilters = !showFilters"
        >
          <Icon name="heroicons:adjustments-horizontal" class="h-5 w-5" aria-hidden="true" />
          <span class="hidden sm:inline">{{ t('customer.products.filters') }}</span>
        </button>
      </div>
    </header>

    <!-- AI 选品 -->
    <section class="catalog-ai">
      <button type="button" class="catalog-ai__toggle" @click="showAIPanel = !showAIPanel">
        <span class="catalog-ai__toggle-left">
          <Icon name="heroicons:sparkles" class="h-5 w-5 text-orange-500" aria-hidden="true" />
          <span class="catalog-ai__toggle-title">{{ t('customer.products.ai_select') }}</span>
          <span class="catalog-ai__toggle-hint">{{ t('customer.products.ai_select_hint') }}</span>
        </span>
        <Icon :name="showAIPanel ? 'heroicons:chevron-up' : 'heroicons:chevron-down'" class="h-5 w-5 text-gray-400" aria-hidden="true" />
      </button>

      <Transition name="catalog-slide">
        <div v-if="showAIPanel" class="catalog-ai__body">
          <div class="catalog-ai__form">
            <textarea
              v-model="aiForm.prompt"
              rows="2"
              :placeholder="t('customer.products.ai_prompt_placeholder')"
              class="catalog-input catalog-input--area"
            />
            <div class="catalog-ai__fields">
              <input
                v-model="aiForm.targetCountry"
                type="text"
                :placeholder="t('customer.products.ai_country_placeholder')"
                class="catalog-input"
              />
              <div class="catalog-ai__row">
                <input v-model.number="aiForm.quantity" type="number" min="1" :placeholder="t('customer.products.ai_qty_placeholder')" class="catalog-input" />
                <input v-model.number="aiForm.budget" type="number" min="0" :placeholder="t('customer.products.ai_budget_placeholder')" class="catalog-input" />
                <input v-model="aiForm.currency" type="text" maxlength="3" :placeholder="currency.cur()" class="catalog-input uppercase" />
              </div>
            </div>
          </div>
          <div class="catalog-ai__actions">
            <button type="button" class="btn-primary" :disabled="aiLoading || !aiForm.prompt.trim()" @click="runAIRecommend">
              <Icon v-if="aiLoading" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" aria-hidden="true" />
              <Icon v-else name="heroicons:sparkles" class="h-4 w-4" aria-hidden="true" />
              {{ aiLoading ? t('customer.products.ai_thinking') : t('customer.products.ai_recommend') }}
            </button>
            <button v-if="aiResults.length" type="button" class="btn-secondary" @click="inquireAllAI">
              <Icon name="heroicons:chat-bubble-left" class="h-4 w-4" aria-hidden="true" />
              {{ t('customer.products.ai_inquire_all', { count: aiResults.length }) }}
            </button>
            <p v-if="aiError" class="catalog-ai__error">{{ aiError }}</p>
          </div>

          <div v-if="aiResults.length" class="catalog-ai__results">
            <p v-if="aiSummary" class="catalog-ai__summary">{{ aiSummary }}</p>
            <div class="catalog-ai__grid">
              <article v-for="item in aiResults" :key="item.id" class="catalog-ai-card">
                <div class="catalog-ai-card__head">
                  <div>
                    <p class="catalog-ai-card__name">{{ tField(item, 'name') }}</p>
                    <p class="catalog-ai-card__cat">{{ tField(item, 'category') }}</p>
                  </div>
                  <div class="catalog-ai-card__badges">
                    <span v-if="item.halalCertified" class="catalog-badge catalog-badge--halal">{{ t('customer.products.halal') }}</span>
                    <span v-if="item.oemAvailable" class="catalog-badge catalog-badge--oem">{{ t('customer.products.oem') }}</span>
                  </div>
                </div>
                <p class="catalog-ai-card__reason">{{ item.reason }}</p>
                <div class="catalog-ai-card__foot">
                  <div class="catalog-qty">
                    <button type="button" class="catalog-qty__btn" @click="item.quantity = Math.max(item.moq || 1, item.quantity - 1)">
                      <Icon name="heroicons:minus" class="h-3 w-3" />
                    </button>
                    <input v-model.number="item.quantity" type="number" :min="item.moq || 1" class="catalog-qty__input" />
                    <button type="button" class="catalog-qty__btn" @click="item.quantity += 1">
                      <Icon name="heroicons:plus" class="h-3 w-3" />
                    </button>
                  </div>
                  <NuxtLink
                    :to="`${localePath('/customer/inquiries/new')}?product=${item.id}&name=${encodeURIComponent(tField(item, 'name'))}`"
                    class="btn-primary btn-primary--sm"
                  >
                    {{ t('customer.products.inquire') }}
                  </NuxtLink>
                </div>
              </article>
            </div>
          </div>
        </div>
      </Transition>
    </section>

    <!-- 主体：筛选 + 列表 -->
    <div class="catalog-body">
      <aside v-show="showFilters" class="catalog-filters">
        <div class="catalog-filters__panel">
          <div class="catalog-filters__head">
            <h2 class="catalog-filters__title">{{ t('customer.products.filters') }}</h2>
            <button type="button" class="catalog-filters__close lg:hidden" @click="showFilters = false">
              <Icon name="heroicons:x-mark" class="h-5 w-5" />
            </button>
          </div>

          <div class="catalog-filters__section">
            <h3 class="catalog-filters__label">{{ t('customer.products.categories') }}</h3>
            <label v-for="cat in categories" :key="cat.slug" class="catalog-check">
              <input v-model="selectedCategories" :value="cat.slug" type="checkbox" />
              <span>{{ tField(cat, 'name') }}</span>
              <span class="catalog-check__count">{{ cat.productCount }}</span>
            </label>
          </div>

          <div class="catalog-filters__section">
            <h3 class="catalog-filters__label">{{ t('customer.products.certifications') }}</h3>
            <label class="catalog-check">
              <input v-model="filters.halal" type="checkbox" />
              <span>{{ t('customer.products.halal_certified') }}</span>
            </label>
            <label class="catalog-check">
              <input v-model="filters.oemOnly" type="checkbox" />
              <span>{{ t('customer.products.oem_available') }}</span>
            </label>
            <label class="catalog-check">
              <input v-model="filters.featuredOnly" type="checkbox" />
              <span>{{ t('customer.products.featured') }}</span>
            </label>
          </div>

          <div class="catalog-filters__section">
            <h3 class="catalog-filters__label">{{ t('customer.products.moq_range') }}</h3>
            <div class="catalog-filters__range">
              <input
                :value="filters.minMoq ?? ''"
                type="number"
                min="1"
                :placeholder="t('customer.products.min')"
                class="catalog-input"
                @input="filters.minMoq = normalizeMoq(($event.target as HTMLInputElement).value)"
              />
              <span>—</span>
              <input
                :value="filters.maxMoq ?? ''"
                type="number"
                min="1"
                :placeholder="t('customer.products.max')"
                class="catalog-input"
                @input="filters.maxMoq = normalizeMoq(($event.target as HTMLInputElement).value)"
              />
            </div>
          </div>

          <div class="catalog-filters__actions">
            <button type="button" class="btn-primary w-full" @click="applyFilters">{{ t('customer.products.apply') }}</button>
            <button type="button" class="btn-ghost w-full" @click="clearFilters">{{ t('customer.products.clear') }}</button>
          </div>
        </div>
      </aside>

      <main class="catalog-main">
        <!-- 加载骨架 -->
        <div v-if="pending" class="catalog-grid">
          <div v-for="i in 8" :key="i" class="catalog-card catalog-card--skeleton">
            <div class="catalog-card__media skeleton" />
            <div class="catalog-card__body">
              <div class="skeleton skeleton--line w-1/3" />
              <div class="skeleton skeleton--line w-3/4" />
              <div class="skeleton skeleton--line w-full" />
            </div>
          </div>
        </div>

        <div v-else-if="error" class="catalog-error">
          <Icon name="heroicons:exclamation-triangle" class="h-10 w-10 text-red-400" aria-hidden="true" />
          <p>{{ error }}</p>
          <button type="button" class="btn-secondary" @click="fetchProducts">{{ t('customer.products.retry') }}</button>
        </div>

        <EmptyState
          v-else-if="products.length === 0"
          icon="heroicons:cube"
          :title="t('customer.products.no_products')"
          class="catalog-empty"
        />

        <template v-else>
          <div class="catalog-toolbar">
            <p class="catalog-toolbar__count">
              {{ t('customer.products.showing', { count: products.length, total: pagination?.total || 0 }) }}
            </p>
            <label class="catalog-sort">
              <span>{{ t('customer.products.sort_by') }}</span>
              <select v-model="sortBy" class="catalog-input catalog-input--select" @change="fetchProducts">
                <option value="popular">{{ t('customer.products.sort_popular') }}</option>
                <option value="name">{{ t('customer.products.sort_name') }}</option>
                <option value="moq_low">{{ t('customer.products.sort_moq_low') }}</option>
                <option value="moq_high">{{ t('customer.products.sort_moq_high') }}</option>
              </select>
            </label>
          </div>

          <div class="catalog-grid">
            <article v-for="product in products" :key="product.id" class="catalog-card">
              <NuxtLink :to="localePath(`/customer/products/${product.slug || product.id}`)" class="catalog-card__media">
                <img v-if="product.thumbnail" :src="product.thumbnail" :alt="tField(product, 'name')" loading="lazy" />
                <div v-else class="catalog-card__placeholder">
                  <Icon name="heroicons:cube" class="h-12 w-12" aria-hidden="true" />
                </div>
                <div class="catalog-card__badges">
                  <span v-if="product.featured" class="catalog-badge catalog-badge--featured">{{ t('customer.products.featured') }}</span>
                  <span v-if="product.halalCertified" class="catalog-badge catalog-badge--halal">{{ t('customer.products.halal') }}</span>
                  <span v-if="product.oemAvailable" class="catalog-badge catalog-badge--oem">{{ t('customer.products.oem') }}</span>
                </div>
              </NuxtLink>

              <div class="catalog-card__body">
                <p class="catalog-card__category">{{ tField(product, 'category') }}</p>
                <h3 class="catalog-card__title">
                  <NuxtLink :to="localePath(`/customer/products/${product.slug || product.id}`)">
                    {{ tField(product, 'name') }}
                  </NuxtLink>
                </h3>
                <p class="catalog-card__summary">{{ tField(product, 'summary') || tField(product, 'description')?.substring(0, 90) }}</p>

                <div class="catalog-card__meta">
                  <span>
                    <Icon name="heroicons:cube" class="h-3.5 w-3.5" aria-hidden="true" />
                    {{ t('customer.products.moq') }} {{ product.moq || 1 }}
                  </span>
                  <span v-if="product.leadTime">
                    <Icon name="heroicons:clock" class="h-3.5 w-3.5" aria-hidden="true" />
                    {{ tField(product, 'leadTime') }}
                  </span>
                </div>

                <div v-if="tArray(product, 'flavors').length || tArray(product, 'shapes').length" class="catalog-card__tags">
                  <span v-for="flavor in tArray(product, 'flavors').slice(0, 2)" :key="flavor">{{ flavor }}</span>
                  <span v-for="shape in tArray(product, 'shapes').slice(0, 1)" :key="shape" class="catalog-card__tag--accent">{{ shape }}</span>
                  <span v-if="tArray(product, 'flavors').length + tArray(product, 'shapes').length > 3" class="catalog-card__tag--muted">{{ t('customer.products.more') }}</span>
                </div>

                <div class="catalog-card__foot">
                  <div>
                    <p class="catalog-card__price-label">{{ t('customer.products.unit_price') }}</p>
                    <p class="catalog-card__price">{{ currency.formatPrice(product.basePrice || 0) }}</p>
                  </div>
                  <div class="catalog-card__actions">
                    <NuxtLink
                      :to="`${localePath('/customer/inquiries/new')}?product=${product.id}&name=${encodeURIComponent(tField(product, 'name'))}`"
                      class="catalog-icon-btn"
                      :title="t('customer.products.inquire')"
                    >
                      <Icon name="heroicons:chat-bubble-left" class="h-5 w-5" />
                    </NuxtLink>
                    <NuxtLink :to="localePath(`/customer/products/${product.slug || product.id}`)" class="btn-primary btn-primary--sm">
                      {{ t('customer.products.view_details') }}
                    </NuxtLink>
                  </div>
                </div>
              </div>
            </article>
          </div>

          <nav v-if="pagination && pagination.totalPages > 1" class="catalog-pagination" aria-label="Pagination">
            <button type="button" class="catalog-page-btn" :disabled="page <= 1" @click="prevPage">
              <Icon name="heroicons:chevron-left" class="h-5 w-5" />
            </button>
            <button
              v-for="p in visiblePages"
              :key="p"
              type="button"
              class="catalog-page-btn"
              :class="{ 'catalog-page-btn--active': p === page }"
              @click="goToPage(p)"
            >
              {{ p }}
            </button>
            <button type="button" class="catalog-page-btn" :disabled="page >= pagination.totalPages" @click="nextPage">
              <Icon name="heroicons:chevron-right" class="h-5 w-5" />
            </button>
          </nav>
        </template>
      </main>
    </div>

    <Transition name="catalog-toast">
      <div v-if="showToast" class="catalog-toast" role="status">
        <Icon name="heroicons:check-circle" class="h-5 w-5" aria-hidden="true" />
        {{ toastMessage }}
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useTranslation } from '~/composables/useTranslation'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const { t } = useI18n()
const { tField, tArray } = useTranslation()
const localePath = useLocalePath()
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

const showAIPanel = ref(false)
const aiLoading = ref(false)
const aiError = ref('')
const aiResults = ref<any[]>([])
const aiSummary = ref('')
const aiForm = reactive({
  prompt: '',
  targetCountry: '',
  quantity: 1,
  budget: 0,
  currency: currency.cur(null),
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
      currency: currency.cur(aiForm.currency).toUpperCase(),
    })
    aiResults.value = (res.recommendations || []).map((item: any) => ({
      ...item,
      quantity: item.quantity || 1,
    }))
    aiSummary.value = res.summary || ''
    if (!aiResults.value.length) aiError.value = t('customer.products.ai_no_results')
  } catch (err: any) {
    aiError.value = err?.message || t('customer.products.ai_error')
  } finally {
    aiLoading.value = false
  }
}

const inquireAllAI = () => {
  const names = aiResults.value.map(item => tField(item, 'name')).join(', ')
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
  maxMoq: null as number | null,
})

/** 将 MOQ 输入规范为有效数字或 null（清空/NaN 视为未设置） */
const normalizeMoq = (value: unknown): number | null => {
  if (value === null || value === undefined || value === '') return null
  const n = typeof value === 'number' ? value : Number(value)
  if (!Number.isFinite(n) || n <= 0) return null
  return Math.floor(n)
}

const sanitizeFilterInputs = () => {
  filters.minMoq = normalizeMoq(filters.minMoq)
  filters.maxMoq = normalizeMoq(filters.maxMoq)
}

const buildProductQuery = (): Parameters<typeof api.getProducts>[0] => {
  sanitizeFilterInputs()
  const query: Parameters<typeof api.getProducts>[0] = {
    page: page.value,
    limit: pageSize,
    sort: sortBy.value as 'popular' | 'name' | 'moq_low' | 'moq_high',
  }
  const search = searchQuery.value.trim()
  if (search) query.search = search
  if (selectedCategories.value.length) query.categories = selectedCategories.value.join(',')
  if (filters.halal) query.halal = true
  if (filters.oemOnly) query.oemOnly = true
  if (filters.featuredOnly) query.featured = true
  if (filters.minMoq != null) query.minMoq = filters.minMoq
  if (filters.maxMoq != null) query.maxMoq = filters.maxMoq
  return query
}

const fetchProducts = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await api.getProducts(buildProductQuery())
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
    categories.value = await api.getCategories() || []
  } catch {
    categories.value = []
  }
}

const visiblePages = computed(() => {
  if (!pagination.value) return []
  const total = pagination.value.totalPages
  const current = page.value
  const pages: number[] = []
  for (let i = Math.max(1, current - 2); i <= Math.min(total, current + 2); i++) pages.push(i)
  return pages
})

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

const prevPage = () => { if (page.value > 1) { page.value -= 1; fetchProducts() } }
const nextPage = () => {
  if (pagination.value && page.value < pagination.value.totalPages) {
    page.value += 1
    fetchProducts()
  }
}
const goToPage = (p: number) => { page.value = p; fetchProducts() }

let searchTimeout: ReturnType<typeof setTimeout>
watch(searchQuery, () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => { page.value = 1; fetchProducts() }, 300)
})

onMounted(() => {
  fetchProducts()
  fetchCategories()
  if (import.meta.client && window.innerWidth < 1024) showFilters.value = false
})
</script>

<style scoped>
.catalog {
  max-width: 1280px;
  margin: 0 auto;
}

/* 页头 */
.catalog-header {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem 1.5rem;
  margin-bottom: 1.5rem;
  padding: 1.25rem 1.5rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
}

.catalog-header__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(1.35rem, 2.5vw, 1.75rem);
  font-weight: 700;
  color: var(--color-primary);
}

.catalog-header__subtitle {
  margin: 0.35rem 0 0;
  font-size: 0.875rem;
  color: var(--color-text-lighter);
  max-width: 36rem;
}

.catalog-header__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
  justify-content: flex-end;
  min-width: min(100%, 280px);
}

.catalog-search {
  position: relative;
  flex: 1;
  min-width: 200px;
  max-width: 320px;
}

.catalog-search__icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  width: 1rem;
  height: 1rem;
  color: var(--color-text-lighter);
  pointer-events: none;
}

.catalog-search__input {
  width: 100%;
  padding: 0.625rem 0.875rem 0.625rem 2.25rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  font-size: 0.875rem;
  background: var(--color-bg-alt);
  transition: border-color 0.15s, box-shadow 0.15s;
}

.catalog-search__input:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(var(--color-highlight-rgb), 0.15);
}

.catalog-toolbar-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.625rem 0.875rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-bg-alt);
  color: var(--color-text);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}

.catalog-toolbar-btn:hover,
.catalog-toolbar-btn--active {
  border-color: var(--color-highlight);
  color: var(--color-highlight);
  background: rgba(var(--color-highlight-rgb), 0.06);
}

/* AI 面板 */
.catalog-ai {
  margin-bottom: 1.5rem;
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-xl);
  background: var(--color-bg);
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

.catalog-ai__toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  gap: 1rem;
  padding: 1rem 1.25rem;
  border: none;
  background: transparent;
  cursor: pointer;
  text-align: left;
}

.catalog-ai__toggle-left {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem 0.75rem;
  min-width: 0;
}

.catalog-ai__toggle-title {
  font-weight: 600;
  color: var(--color-text);
  font-size: 0.9375rem;
}

.catalog-ai__toggle-hint {
  font-size: 0.8125rem;
  color: var(--color-text-lighter);
}

.catalog-ai__body {
  padding: 0 1.25rem 1.25rem;
  border-top: 1px solid var(--color-border-light);
}

.catalog-ai__form {
  display: grid;
  gap: 0.75rem;
  padding-top: 1rem;
}

@media (min-width: 768px) {
  .catalog-ai__form { grid-template-columns: 1.4fr 1fr; }
}

.catalog-ai__fields { display: flex; flex-direction: column; gap: 0.5rem; }
.catalog-ai__row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.5rem; }

.catalog-ai__actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.75rem;
}

.catalog-ai__error { font-size: 0.8125rem; color: var(--color-error); margin: 0; }
.catalog-ai__summary { font-size: 0.8125rem; color: var(--color-text-lighter); margin: 1rem 0 0.75rem; }
.catalog-ai__grid { display: grid; gap: 0.75rem; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); }

.catalog-ai-card {
  padding: 1rem;
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  background: var(--color-bg-alt);
}

.catalog-ai-card__head { display: flex; justify-content: space-between; gap: 0.5rem; }
.catalog-ai-card__name { font-weight: 600; font-size: 0.875rem; margin: 0; }
.catalog-ai-card__cat { font-size: 0.75rem; color: var(--color-text-lighter); margin: 0.125rem 0 0; }
.catalog-ai-card__badges { display: flex; flex-wrap: wrap; gap: 0.25rem; }
.catalog-ai-card__reason { font-size: 0.75rem; color: var(--color-text-lighter); font-style: italic; margin: 0.5rem 0; }
.catalog-ai-card__foot { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; padding-top: 0.75rem; border-top: 1px solid var(--color-border-light); }

/* 布局 */
.catalog-body {
  display: grid;
  gap: 1.25rem;
}

@media (min-width: 1024px) {
  .catalog-body { grid-template-columns: 260px 1fr; align-items: start; }
}

.catalog-filters__panel {
  position: sticky;
  top: 1rem;
  padding: 1rem;
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-xl);
  background: var(--color-bg);
  box-shadow: var(--shadow-sm);
}

.catalog-filters__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}

.catalog-filters__title {
  margin: 0;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text);
}

.catalog-filters__close {
  border: none;
  background: transparent;
  color: var(--color-text-lighter);
  cursor: pointer;
}

.catalog-filters__section { margin-bottom: 1rem; }

.catalog-filters__label {
  margin: 0 0 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-lighter);
}

.catalog-filters__range {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 0.375rem;
  color: var(--color-text-lighter);
}

.catalog-filters__actions { display: flex; flex-direction: column; gap: 0.375rem; }

.catalog-check {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0;
  font-size: 0.8125rem;
  color: var(--color-text);
  cursor: pointer;
}

.catalog-check input { accent-color: var(--color-highlight); }
.catalog-check__count { margin-left: auto; font-size: 0.75rem; color: var(--color-text-lighter); }

/* 工具栏 */
.catalog-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.catalog-toolbar__count { font-size: 0.8125rem; color: var(--color-text-lighter); margin: 0; }

.catalog-sort {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8125rem;
  color: var(--color-text-lighter);
}

/* 产品网格 */
.catalog-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 1rem;
}

.catalog-card {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-xl);
  background: var(--color-bg);
  overflow: hidden;
  transition: box-shadow 0.2s, transform 0.2s;
}

.catalog-card:hover {
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

.catalog-card__media {
  position: relative;
  display: block;
  aspect-ratio: 4 / 3;
  overflow: hidden;
  background: linear-gradient(145deg, rgba(var(--color-highlight-rgb), 0.08), var(--color-bg-alt));
}

.catalog-card__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.35s ease;
}

.catalog-card:hover .catalog-card__media img { transform: scale(1.04); }

.catalog-card__placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(var(--color-highlight-rgb), 0.35);
}

.catalog-card__badges {
  position: absolute;
  top: 0.625rem;
  left: 0.625rem;
  right: 0.625rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.catalog-badge {
  padding: 0.125rem 0.5rem;
  border-radius: var(--radius-full);
  font-size: 0.625rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.catalog-badge--featured { background: #fbbf24; color: #78350f; }
.catalog-badge--halal { background: #d1fae5; color: #065f46; }
.catalog-badge--oem { background: var(--color-highlight); color: white; margin-left: auto; }

.catalog-card__body {
  display: flex;
  flex-direction: column;
  flex: 1;
  padding: 1rem;
  gap: 0.375rem;
}

.catalog-card__category {
  margin: 0;
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-highlight);
}

.catalog-card__title {
  margin: 0;
  font-size: 0.9375rem;
  font-weight: 600;
  line-height: 1.35;
}

.catalog-card__title a {
  color: var(--color-text);
  text-decoration: none;
}

.catalog-card__title a:hover { color: var(--color-highlight); }

.catalog-card__summary {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--color-text-lighter);
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.catalog-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  font-size: 0.75rem;
  color: var(--color-text-lighter);
}

.catalog-card__meta span { display: inline-flex; align-items: center; gap: 0.25rem; }

.catalog-card__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.catalog-card__tags span,
.catalog-card__tag--accent,
.catalog-card__tag--muted {
  padding: 0.125rem 0.5rem;
  border-radius: var(--radius-full);
  font-size: 0.6875rem;
  background: var(--color-bg-alt);
  color: var(--color-text-lighter);
}

.catalog-card__tag--accent { background: rgba(var(--color-highlight-rgb), 0.1); color: var(--color-highlight); }

.catalog-card__foot {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.5rem;
  margin-top: auto;
  padding-top: 0.75rem;
  border-top: 1px solid var(--color-border-light);
}

.catalog-card__price-label { margin: 0; font-size: 0.6875rem; color: var(--color-text-lighter); }
.catalog-card__price { margin: 0; font-size: 1.0625rem; font-weight: 700; color: var(--color-highlight); }

.catalog-card__actions { display: flex; align-items: center; gap: 0.375rem; }

.catalog-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
  color: var(--color-highlight);
  background: rgba(var(--color-highlight-rgb), 0.06);
  transition: background 0.15s;
}

.catalog-icon-btn:hover { background: rgba(var(--color-highlight-rgb), 0.12); }

/* 通用控件 */
.catalog-input {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  background: var(--color-bg-alt);
}

.catalog-input:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 2px rgba(var(--color-highlight-rgb), 0.12);
}

.catalog-input--area { resize: vertical; min-height: 4.5rem; }
.catalog-input--select { width: auto; min-width: 10rem; padding-right: 1.75rem; }

.btn-primary,
.btn-secondary,
.btn-ghost {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  padding: 0.5rem 1rem;
  border-radius: var(--radius-md);
  font-size: 0.8125rem;
  font-weight: 600;
  border: none;
  cursor: pointer;
  transition: background 0.15s, opacity 0.15s;
}

.btn-primary {
  background: var(--color-highlight);
  color: white;
}

.btn-primary:hover:not(:disabled) { filter: brightness(1.05); }
.btn-primary:disabled { opacity: 0.55; cursor: not-allowed; }
.btn-primary--sm { padding: 0.4375rem 0.75rem; font-size: 0.75rem; }

.btn-secondary {
  background: var(--color-bg-alt);
  color: var(--color-text);
  border: 1px solid var(--color-border);
}

.btn-ghost {
  background: transparent;
  color: var(--color-text-lighter);
  border: 1px solid var(--color-border-light);
}

.catalog-qty { display: flex; align-items: center; gap: 0.25rem; }
.catalog-qty__btn {
  width: 1.75rem;
  height: 1.75rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg);
  cursor: pointer;
}

.catalog-qty__input {
  width: 3rem;
  text-align: center;
  font-size: 0.8125rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 0.25rem;
}

/* 分页 / 状态 */
.catalog-pagination {
  display: flex;
  justify-content: center;
  gap: 0.375rem;
  margin-top: 1.5rem;
}

.catalog-page-btn {
  min-width: 2.25rem;
  height: 2.25rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg);
  font-size: 0.8125rem;
  cursor: pointer;
}

.catalog-page-btn--active {
  background: var(--color-highlight);
  border-color: var(--color-highlight);
  color: white;
}

.catalog-page-btn:disabled { opacity: 0.45; cursor: not-allowed; }

.catalog-error,
.catalog-empty {
  padding: 3rem 1rem;
  text-align: center;
}

.catalog-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  color: var(--color-error);
}

.skeleton {
  background: linear-gradient(90deg, var(--color-bg-alt) 25%, var(--color-border-light) 50%, var(--color-bg-alt) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.2s infinite;
  border-radius: var(--radius-sm);
}

.skeleton--line { height: 0.75rem; margin-bottom: 0.5rem; }
.catalog-card--skeleton .catalog-card__body { gap: 0.5rem; }

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.catalog-toast {
  position: fixed;
  bottom: 1.5rem;
  right: 1.5rem;
  z-index: 100;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background: var(--color-primary);
  color: white;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  font-size: 0.875rem;
}

.catalog-slide-enter-active,
.catalog-slide-leave-active,
.catalog-toast-enter-active,
.catalog-toast-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.catalog-slide-enter-from,
.catalog-slide-leave-to,
.catalog-toast-enter-from,
.catalog-toast-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

@media (prefers-reduced-motion: reduce) {
  .catalog-card:hover { transform: none; }
  .skeleton { animation: none; }
}
</style>
