<template>
  <div>
    <PageHeader :title="t('admin.products.title')" :description="t('admin.products.description')">
      <template #actions>
        <button type="button" class="inline-flex items-center justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50" @click="openCreateModal">
          {{ t('admin.products.add_product') }}
        </button>
        <button type="button" :disabled="batchTranslating || !products.length" class="inline-flex items-center gap-2 justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 disabled:opacity-50" @click="batchTranslateMissing">
          <Icon name="heroicons:language" class="h-4 w-4" aria-hidden="true" />
          {{ batchTranslating ? t('admin.products.translating') : t('admin.products.batch_translate') }}
        </button>
        <button type="button" class="inline-flex items-center gap-2 justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700" @click="openCreateDrawer">
          <Icon name="heroicons:sparkles" class="h-4 w-4" />
          {{ t('admin.products.ai_import') }}
        </button>
        <AiHelpHint topic="products_import" />
      </template>
    </PageHeader>

    <!-- 搜索与筛选 -->
    <div class="mb-4 flex flex-wrap items-end gap-4">
      <div class="min-w-[200px] flex-1">
        <label for="products-search" class="block text-xs font-medium text-gray-700">{{ t('admin.products.search') }}</label>
        <input
          id="products-search"
          v-model="searchQuery"
          name="search"
          type="search"
          autocomplete="off"
          class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
          :placeholder="t('admin.products.search_placeholder')"
          @keyup.enter="fetchProducts"
        />
      </div>
      <div class="min-w-[140px]">
        <label for="products-status" class="block text-xs font-medium text-gray-700">{{ t('admin.products.status_filter') }}</label>
        <select
          id="products-status"
          v-model="statusFilter"
          class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"
          @change="page = 1; fetchProducts()"
        >
          <option value="">{{ t('admin.products.status_published') }}</option>
          <option value="all">{{ t('admin.products.status_all') }}</option>
          <option value="draft">{{ t('admin.products.status_draft') }}</option>
          <option value="inactive">{{ t('admin.products.status_inactive') }}</option>
        </select>
      </div>
      <button type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700 hover:bg-gray-50" @click="fetchProducts">
        {{ t('admin.products.search') }}
      </button>
    </div>

    <AdminTable
      :columns="columns"
      :rows="products"
      :loading="pending"
      :error="!!error"
      :error-message="error"
      :empty-text="t('admin.products.no_data')"
      @retry="fetchProducts"
    >
      <template #cell-product="{ row }">
        <div class="flex items-center">
          <div class="h-10 w-10 flex-shrink-0">
            <img v-if="row.thumbnail" class="h-10 w-10 rounded-full object-cover" :src="row.thumbnail" alt="" />
            <div v-else class="flex h-10 w-10 items-center justify-center rounded-full bg-gray-200 font-bold text-gray-500">
              {{ tField(row, 'name')?.charAt(0) }}
            </div>
          </div>
          <div class="ms-4">
            <div class="font-medium text-gray-900">{{ tField(row, 'name') }}</div>
            <div class="text-gray-500">{{ row.slug }}</div>
          </div>
        </div>
      </template>

      <template #cell-category="{ row }">
        {{ categoryLabel(row) }}
      </template>

      <template #cell-moq="{ row }">
        {{ row.moq || 0 }}
      </template>

      <template #cell-stock="{ row }">
        <span :class="(row.stockQuantity || 0) > 0 ? 'text-green-600' : 'text-red-600'">
          {{ row.stockQuantity || 0 }}
        </span>
      </template>

      <template #cell-status="{ row }">
        <StatusBadge :status="row.status" type="product" />
      </template>

      <template #cell-actions="{ row }">
        <button type="button" class="text-orange-600 hover:text-orange-900" @click="openEditModal(row)">{{ t('admin.products.edit') }}</button>
        <button type="button" class="ms-4 text-red-600 hover:text-red-900" @click="deleteProduct(row.id)">{{ t('admin.products.delete') }}</button>
      </template>

      <template #bottom>
        <div v-if="pagination" class="flex items-center justify-between">
          <div class="text-sm text-gray-700">
            {{ t('admin.products.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
          </div>
          <div class="flex items-center gap-2">
            <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page <= 1" @click="prevPage">
              {{ t('admin.products.previous') }}
            </button>
            <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page >= pagination.totalPages" @click="nextPage">
              {{ t('admin.products.next') }}
            </button>
          </div>
        </div>
      </template>
    </AdminTable>

    <!-- Editor Modal -->
    <AdminModal :open="showModal" :title="editingId ? t('admin.products.edit_product') : t('admin.products.create_product')" width="xl" @close="closeModal">
      <form @submit.prevent="saveProduct">
        <ProductFormFields
          :form="form"
          :editing-id="editingId"
          :ai-translating="aiTranslating"
          :translation-locale="translationLocale"
          :uploading-image="uploadingImage"
          @trigger-upload="triggerUpload"
          @cancel-upload="cancelImageUpload"
          @ai-translate="aiTranslateAll"
          @update:translation-locale="translationLocale = $event"
        />
        <div v-if="formError" class="mt-4 text-sm text-red-600">{{ formError }}</div>
        <div class="mt-4 flex justify-end gap-3">
          <button type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700" @click="closeModal">
            {{ t('admin.products.cancel') }}
          </button>
          <button type="submit" :disabled="saving" class="rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm text-white disabled:opacity-50">
            {{ saving ? t('admin.products.saving') : (editingId ? t('admin.products.update') : t('admin.products.create')) }}
          </button>
        </div>
      </form>
    </AdminModal>

    <!-- AI Generate Drawer -->
    <Drawer :open="showCreateDrawer" width="2xl" @close="showCreateDrawer = false">
      <template #title>
        <span class="inline-flex items-center">
          {{ t('admin.products.add_product') }} — AI
          <AiHelpHint topic="products_import" size="sm" />
        </span>
      </template>

      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6" style="min-height: 60vh;">
        <div class="lg:col-span-5 space-y-4">
          <div>
            <label for="ai-product-desc" class="block text-sm font-medium text-gray-700">{{ t('admin.products.ai_description_label') }}</label>
            <textarea
              id="ai-product-desc"
              v-model="aiDescription"
              rows="12"
              class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500"
              :placeholder="t('admin.products.ai_description_placeholder')"
            />
          </div>
          <div>
            <label for="ai-product-lang" class="block text-sm font-medium text-gray-700">{{ t('admin.products.ai_language_label') }}</label>
            <select
              id="ai-product-lang"
              v-model="aiLanguage"
              class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500"
            >
              <option v-for="loc in translationLocales" :key="loc" :value="loc">{{ localeTabLabel(loc) }}</option>
            </select>
          </div>
          <button
            type="button"
            :disabled="aiGenerating || !aiDescription.trim()"
            class="inline-flex items-center gap-2 rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50 transition-colors"
            @click="generateProduct"
          >
            <Icon name="heroicons:sparkles" class="h-4 w-4" />
            {{ aiGenerating ? t('admin.products.ai_generating') : t('admin.products.ai_generate') }}
          </button>
          <p v-if="formError" class="text-sm text-red-600">{{ formError }}</p>
        </div>

        <div class="lg:col-span-7 lg:border-l lg:pl-6 overflow-y-auto" style="max-height: 65vh;">
          <div v-if="!aiHasData && !aiGenerating" class="flex flex-col items-center justify-center h-48 text-gray-400 text-sm">
            <Icon name="heroicons:sparkles" class="h-10 w-10 mb-2 text-gray-300" />
            {{ t('admin.products.ai_empty_preview') }}
          </div>

          <div v-if="aiGenerating" class="flex flex-col items-center justify-center h-48 text-gray-500 text-sm gap-2">
            <svg class="animate-spin h-6 w-6 text-orange-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            {{ t('admin.products.ai_generating') }}
          </div>

          <ProductFormFields
            v-if="aiHasData"
            :form="form"
            editing-id=""
            :ai-translating="false"
            :translation-locale="translationLocale"
            :uploading-image="uploadingImage"
            @trigger-upload="triggerUpload"
            @cancel-upload="cancelImageUpload"
            @update:translation-locale="translationLocale = $event"
          />
        </div>
      </div>

      <template #footer>
        <p v-if="formError" class="mb-3 text-sm text-red-600">{{ formError }}</p>
        <div class="flex justify-end gap-3">
          <button type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700 hover:bg-gray-50" @click="showCreateDrawer = false">
            {{ t('admin.products.cancel') }}
          </button>
          <button
            type="button"
            :disabled="saving || !aiHasData"
            class="inline-flex items-center gap-2 rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50"
            @click="saveProduct()"
          >
            <Icon name="heroicons:sparkles" class="h-4 w-4" />
            {{ saving ? t('admin.products.saving') : t('admin.products.ai_save') }}
          </button>
        </div>
      </template>
    </Drawer>

    <p v-if="actionMessage" class="mt-4 text-sm" :class="actionError ? 'text-red-600' : 'text-green-600'">
      {{ actionMessage }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, watch, onMounted } from 'vue'
import { useTranslation } from '~/composables/useTranslation'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const api = useApi()
const route = useRoute()
const { t } = useI18n()
const { tField } = useTranslation()
const { enumLabel } = useDisplay()

const columns = [
  { key: 'product', label: t('admin.products.col_product') },
  { key: 'category', label: t('admin.products.col_category') },
  { key: 'moq', label: t('admin.products.col_moq') },
  { key: 'stock', label: t('admin.products.col_stock') },
  { key: 'status', label: t('admin.products.col_status') },
  { key: 'actions', label: '' },
]

const products = ref<any[]>([])
const categoryCatalog = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20
const searchQuery = ref('')
const statusFilter = ref('')

const showModal = ref(false)
const editingId = ref('')
const saving = ref(false)
const translationLocale = ref(SOURCE_LOCALE)
const aiTranslating = ref(false)
let aiTranslateSlowTimer: ReturnType<typeof setTimeout> | null = null

const categoryBySlug = computed(() => {
  const map = new Map<string, any>()
  for (const cat of categoryCatalog.value) map.set(cat.slug, cat)
  return map
})

const categoryLabel = (row: any) => {
  const fromProduct = tField(row, 'category')
  if (fromProduct) return fromProduct
  const slug = row?.categorySlug
  if (slug && categoryBySlug.value.has(slug)) {
    const cat = categoryBySlug.value.get(slug)
    return tField(cat, 'name') || cat?.name || '-'
  }
  return row?.category || '-'
}

const fetchCategories = async () => {
  try {
    categoryCatalog.value = await api.adminGetCategories()
  } catch {
    categoryCatalog.value = []
  }
}
const batchTranslating = ref(false)
const formError = ref('')
const actionMessage = ref('')
const actionError = ref(false)
const uploadingImage = ref(false)
const uploadError = ref('')

// AI Drawer state
const showCreateDrawer = ref(false)
const aiDescription = ref('')
const aiLanguage = ref('zh')
const aiGenerating = ref(false)
const aiHasData = ref(false)
let uploadAbortController: AbortController | null = null

const triggerUpload = async (field: 'thumbnail' | 'ogImage' | 'images') => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/jpeg,image/png,image/webp,image/gif'
  input.onchange = async (e: any) => {
    const file = e.target.files?.[0]
    if (!file) return
    uploadingImage.value = true
    uploadError.value = ''
    uploadAbortController = new AbortController()
    try {
      const formData = new FormData()
      formData.append('file', file)
      const result = await api.post<any>('/admin/upload/image', formData)
      if (field === 'thumbnail') {
        form.thumbnail = result.url
      } else if (field === 'ogImage') {
        form.ogImage = result.url
      } else {
        const existing = form.imagesInput ? form.imagesInput.split(',').filter(Boolean) : []
        existing.push(result.url)
        form.imagesInput = existing.join(',')
      }
    } catch (err: any) {
      if (err?.name !== 'AbortError') {
        uploadError.value = err?.message || t('errors.api.upload_failed')
      }
    } finally {
      uploadingImage.value = false
      uploadAbortController = null
    }
  }
  input.click()
}

const cancelImageUpload = () => {
  if (uploadAbortController) {
    uploadAbortController.abort()
    uploadAbortController = null
  }
  uploadingImage.value = false
}

// ALL_LOCALES 由 Nuxt 自动导入
const translationLocales = ALL_LOCALES

const translationFieldKeys = ['name', 'alias', 'summary', 'description', 'category', 'categoryAlias', 'ingredients', 'allergens', 'storage', 'leadTime', 'shelfLife'] as const

function emptyTranslationEntry() {
  return {
    name: '', alias: '', summary: '', description: '', category: '', categoryAlias: '',
    ingredients: '', allergens: '', storage: '', leadTime: '', shelfLife: '',
    flavorsInput: '', shapesInput: '',
  }
}

function emptyTranslations() {
  const result: Record<string, ReturnType<typeof emptyTranslationEntry>> = {}
  for (const loc of translationLocales) {
    result[loc] = emptyTranslationEntry()
  }
  return result
}

const form = reactive({
  name: '', slug: '', summary: '', description: '',
  category: '', categorySlug: '', thumbnail: '', ogImage: '',
  moq: 0, basePrice: 0, stockQuantity: 0, leadTime: '',
  oemAvailable: false, halalCertified: false, featured: false, status: 'active',
  imagesInput: '', flavorsInput: '', shapesInput: '', certificationsInput: '',
  ingredients: '', allergens: '', shelfLife: '', storage: '',
  netWeightPerPiece: 0, netWeightPerPack: 0, grossWeightPerCarton: 0,
  piecesPerPack: 0, packsPerCarton: 0,
  productLengthMM: 0, productWidthMM: 0, productHeightMM: 0,
  energyKj: 0, energyKcal: 0, totalFatG: 0, saturatedFatG: 0,
  carbohydratesG: 0, sugarsG: 0, proteinG: 0, saltG: 0, fiberG: 0,
  additivesInput: '', sweetenerType: '', cocoaSolidsPct: 0, milkSolidsPct: 0,
  gmoStatus: '', mayContainInput: '', waterActivity: 0,
  gtin: '', hsCode: '',
  primaryPackaging: '', innerPackConfig: '', palletConfig: '',
  isVegan: false, isGlutenFree: false, isSugarFree: false, isKosher: false, isOrganic: false,
  sampleMOQ: 0, sampleLeadTime: '', samplePrice: 0,
  translations: emptyTranslations()
} as any)

const localeTabLabelMap: Record<string, string> = {
  en: 'admin.products.locale_tab_en', zh: 'admin.products.locale_tab_zh',
  ko: 'admin.products.locale_tab_ko', ar: 'admin.products.locale_tab_ar',
  ja: 'admin.products.locale_tab_ja', th: 'admin.products.locale_tab_th',
  vi: 'admin.products.locale_tab_vi', id: 'admin.products.locale_tab_id',
  ms: 'admin.products.locale_tab_ms',
}

const localeTabLabel = (loc: string) => t(localeTabLabelMap[loc] || loc)

const parseCSV = (value: string) => value.split(',').map(v => v.trim()).filter(Boolean)

/** 将主语言翻译字段同步到标量列（后端搜索/兼容用） */
const syncScalarsFromSourceLocale = () => {
  const src = form.translations[SOURCE_LOCALE]
  if (!src) return
  form.name = src.name?.trim?.() || form.name
  form.summary = src.summary?.trim?.() || form.summary
  form.description = src.description?.trim?.() || form.description
  form.category = src.category?.trim?.() || form.category
  form.ingredients = src.ingredients?.trim?.() || form.ingredients
  form.allergens = src.allergens?.trim?.() || form.allergens
  form.storage = src.storage?.trim?.() || form.storage
  form.shelfLife = src.shelfLife?.trim?.() || form.shelfLife
  form.leadTime = src.leadTime?.trim?.() || form.leadTime
  form.flavorsInput = src.flavorsInput?.trim?.() || form.flavorsInput
  form.shapesInput = src.shapesInput?.trim?.() || form.shapesInput
}

/** 旧单语产品：标量字段仅回填 en，禁止写入 zh（标量多为英文遗留） */
const hydrateTranslationsFromLegacyScalars = () => {
  const hasAnyTranslation = translationLocales.some((loc) => {
    const tr = form.translations[loc]
    if (!tr) return false
    return translationFieldKeys.some((k) => String(tr[k] ?? '').trim() !== '')
      || String(tr.flavorsInput ?? '').trim() !== ''
      || String(tr.shapesInput ?? '').trim() !== ''
  })
  if (hasAnyTranslation) return

  const en = form.translations.en
  if (!en) return
  if (!en.name?.trim() && form.name) en.name = form.name
  if (!en.summary?.trim() && form.summary) en.summary = form.summary
  if (!en.description?.trim() && form.description) en.description = form.description
  if (!en.category?.trim() && form.category) en.category = form.category
  if (!en.ingredients?.trim() && form.ingredients) en.ingredients = form.ingredients
  if (!en.allergens?.trim() && form.allergens) en.allergens = form.allergens
  if (!en.storage?.trim() && form.storage) en.storage = form.storage
  if (!en.shelfLife?.trim() && form.shelfLife) en.shelfLife = form.shelfLife
  if (!en.leadTime?.trim() && form.leadTime) en.leadTime = form.leadTime
  if (!en.flavorsInput?.trim() && form.flavorsInput) en.flavorsInput = form.flavorsInput
  if (!en.shapesInput?.trim() && form.shapesInput) en.shapesInput = form.shapesInput
}

/** 清除 zh 中与 en 完全相同的字段（多为误从 en/标量复制） */
const repairEnglishDuplicatesInZh = () => {
  const zh = form.translations[SOURCE_LOCALE]
  const en = form.translations.en
  if (!zh || !en) return
  for (const key of [...translationFieldKeys, 'flavorsInput', 'shapesInput'] as const) {
    const zhVal = String(zh[key] ?? '').trim()
    const enVal = String(en[key] ?? '').trim()
    if (zhVal && enVal && zhVal === enVal) {
      zh[key] = ''
    }
  }
}

const applyTranslationFields = (loc: string, fields: Record<string, any>) => {
  if (!fields || typeof fields !== 'object' || !form.translations[loc]) return
  for (const key of translationFieldKeys) {
    if (fields[key]) form.translations[loc][key] = fields[key]
  }
  if (Array.isArray(fields.flavors)) {
    form.translations[loc].flavorsInput = fields.flavors.join(', ')
  } else if (typeof fields.flavors === 'string' && fields.flavors) {
    try {
      const arr = JSON.parse(fields.flavors)
      form.translations[loc].flavorsInput = Array.isArray(arr) ? arr.join(', ') : fields.flavors
    } catch {
      form.translations[loc].flavorsInput = fields.flavors
    }
  }
  if (Array.isArray(fields.shapes)) {
    form.translations[loc].shapesInput = fields.shapes.join(', ')
  } else if (typeof fields.shapes === 'string' && fields.shapes) {
    try {
      const arr = JSON.parse(fields.shapes)
      form.translations[loc].shapesInput = Array.isArray(arr) ? arr.join(', ') : fields.shapes
    } catch {
      form.translations[loc].shapesInput = fields.shapes
    }
  }
}

const buildTranslationsPayload = () => {
  const result: Record<string, Record<string, string>> = {}
  for (const loc of translationLocales) {
    const tr = form.translations[loc]
    if (!tr) continue
    const entry: Record<string, string> = {}
    for (const key of translationFieldKeys) {
      const val = tr[key]?.trim?.() ?? tr[key]
      if (typeof val === 'string' && val.trim()) entry[key] = val.trim()
    }
    if (tr.flavorsInput?.trim()) {
      entry.flavors = JSON.stringify(parseCSV(tr.flavorsInput))
    }
    if (tr.shapesInput?.trim()) {
      entry.shapes = JSON.stringify(parseCSV(tr.shapesInput))
    }
    if (Object.keys(entry).length > 0) result[loc] = entry
  }
  return Object.keys(result).length > 0 ? result : null
}

/** 标量列回填到 zh 翻译 tab（DB 中 translations.zh 可能缺失） */
const hydrateSourceLocaleFromScalars = () => {
  const zh = form.translations[SOURCE_LOCALE]
  if (!zh) return
  if (!String(zh.name ?? '').trim() && form.name) zh.name = form.name
  if (!String(zh.summary ?? '').trim() && form.summary) zh.summary = form.summary
  if (!String(zh.description ?? '').trim() && form.description) zh.description = form.description
  if (!String(zh.category ?? '').trim() && form.category) zh.category = form.category
  if (!String(zh.ingredients ?? '').trim() && form.ingredients) zh.ingredients = form.ingredients
  if (!String(zh.allergens ?? '').trim() && form.allergens) zh.allergens = form.allergens
  if (!String(zh.storage ?? '').trim() && form.storage) zh.storage = form.storage
  if (!String(zh.shelfLife ?? '').trim() && form.shelfLife) zh.shelfLife = form.shelfLife
  if (!String(zh.leadTime ?? '').trim() && form.leadTime) zh.leadTime = form.leadTime
  if (!String(zh.flavorsInput ?? '').trim() && form.flavorsInput) zh.flavorsInput = form.flavorsInput
  if (!String(zh.shapesInput ?? '').trim() && form.shapesInput) zh.shapesInput = form.shapesInput
}

/** 从服务端重新加载 translations 并应用到表单（AI 翻译超时兜底） */
const reloadProductTranslations = async (productId: string) => {
  const detail = await api.get<any>(`/admin/products/${productId}`)
  if (!detail?.translations || typeof detail.translations !== 'object') return false
  for (const loc of translationLocales) {
    if (detail.translations[loc] && typeof detail.translations[loc] === 'object') {
      applyTranslationFields(loc, detail.translations[loc])
    }
  }
  hydrateSourceLocaleFromScalars()
  return true
}

const aiTranslateAll = async () => {
  if (!editingId.value || aiTranslating.value) return
  aiTranslating.value = true
  formError.value = ''
  let requestFailed = false
  if (aiTranslateSlowTimer) clearTimeout(aiTranslateSlowTimer)
  aiTranslateSlowTimer = setTimeout(() => {
    if (aiTranslating.value) {
      formError.value = t('admin.products.ai_translate_slow')
    }
  }, 30_000)
  try {
    const targetLocales = translationLocales.filter(l => l !== SOURCE_LOCALE)
    const res = await api.adminAITranslateProduct(editingId.value, {
      productId: editingId.value, targetLocales,
    })
    if (res?.translations && typeof res.translations === 'object') {
      for (const loc of translationLocales) {
        if (res.translations[loc] && typeof res.translations[loc] === 'object') {
          applyTranslationFields(loc, res.translations[loc])
        }
      }
      hydrateSourceLocaleFromScalars()
    }
  } catch (err: any) {
    requestFailed = true
    // 服务端可能已保存但响应超时，尝试从 DB 拉取最新 translations
    try {
      const reloaded = await reloadProductTranslations(editingId.value)
      if (reloaded) {
        actionMessage.value = t('admin.products.ai_translate_reloaded')
        actionError.value = false
        return
      }
    } catch { /* fall through */ }
    formError.value = err?.message || t('errors.api.request_failed')
  } finally {
    if (aiTranslateSlowTimer) {
      clearTimeout(aiTranslateSlowTimer)
      aiTranslateSlowTimer = null
    }
    aiTranslating.value = false
  }
  if (!requestFailed) {
    actionMessage.value = t('admin.products.ai_translate_done')
    actionError.value = false
  }
}

const batchTranslateMissing = async () => {
  if (batchTranslating.value || !products.value.length) return
  batchTranslating.value = true
  formError.value = ''
  try {
    const targetLocales = translationLocales.filter(l => l !== SOURCE_LOCALE)
    for (const row of products.value.slice(0, 20)) {
      await api.adminAITranslateProduct(row.id, {
        productId: row.id, targetLocales,
      })
    }
    await fetchProducts()
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.request_failed')
  } finally {
    batchTranslating.value = false
  }
}

const resetForm = () => {
  form.name = ''; form.slug = ''; form.summary = ''; form.description = ''
  form.category = ''; form.categorySlug = ''; form.thumbnail = ''; form.ogImage = ''
  form.moq = 0; form.basePrice = 0; form.stockQuantity = 0; form.leadTime = ''
  form.oemAvailable = false; form.halalCertified = false; form.featured = false; form.status = 'active'
  form.imagesInput = ''; form.flavorsInput = ''; form.shapesInput = ''; form.certificationsInput = ''
  form.ingredients = ''; form.allergens = ''; form.shelfLife = ''; form.storage = ''
  form.netWeightPerPiece = 0; form.netWeightPerPack = 0; form.grossWeightPerCarton = 0
  form.piecesPerPack = 0; form.packsPerCarton = 0
  form.productLengthMM = 0; form.productWidthMM = 0; form.productHeightMM = 0
  form.energyKj = 0; form.energyKcal = 0; form.totalFatG = 0; form.saturatedFatG = 0
  form.carbohydratesG = 0; form.sugarsG = 0; form.proteinG = 0; form.saltG = 0; form.fiberG = 0
  form.additivesInput = ''; form.sweetenerType = ''; form.cocoaSolidsPct = 0; form.milkSolidsPct = 0
  form.gmoStatus = ''; form.mayContainInput = ''; form.waterActivity = 0
  form.gtin = ''; form.hsCode = ''
  form.primaryPackaging = ''; form.innerPackConfig = ''; form.palletConfig = ''
  form.isVegan = false; form.isGlutenFree = false; form.isSugarFree = false; form.isKosher = false; form.isOrganic = false
  form.sampleMOQ = 0; form.sampleLeadTime = ''; form.samplePrice = 0
  form.translations = emptyTranslations()
}

const fillFormFromProduct = (product: any) => {
  form.name = product.name || ''; form.slug = product.slug || ''
  form.summary = product.summary || ''; form.description = product.description || ''
  form.category = product.category || ''; form.categorySlug = product.categorySlug || ''
  form.thumbnail = product.thumbnail || ''; form.ogImage = product.ogImage || ''; form.moq = product.moq || 0
  form.basePrice = product.basePrice || 0; form.stockQuantity = product.stockQuantity || 0
  form.leadTime = product.leadTime || ''
  form.oemAvailable = Boolean(product.oemAvailable)
  form.halalCertified = Boolean(product.halalCertified)
  form.featured = Boolean(product.featured)
  form.status = product.status || 'active'
  form.imagesInput = Array.isArray(product.images) ? product.images.join(', ') : ''
  form.flavorsInput = Array.isArray(product.flavors) ? product.flavors.join(', ') : ''
  form.shapesInput = Array.isArray(product.shapes) ? product.shapes.join(', ') : ''
  form.certificationsInput = Array.isArray(product.certifications) ? product.certifications.join(', ') : ''
  form.ingredients = product.ingredients || ''; form.allergens = product.allergens || ''
  form.shelfLife = product.shelfLife || ''; form.storage = product.storage || ''
  form.netWeightPerPiece = product.netWeightPerPiece || 0
  form.netWeightPerPack = product.netWeightPerPack || 0
  form.grossWeightPerCarton = product.grossWeightPerCarton || 0
  form.piecesPerPack = product.piecesPerPack || 0
  form.packsPerCarton = product.packsPerCarton || 0
  form.productLengthMM = product.productLengthMM || 0
  form.productWidthMM = product.productWidthMM || 0
  form.productHeightMM = product.productHeightMM || 0
  form.energyKj = product.energyKj || 0; form.energyKcal = product.energyKcal || 0
  form.totalFatG = product.totalFatG || 0; form.saturatedFatG = product.saturatedFatG || 0
  form.carbohydratesG = product.carbohydratesG || 0; form.sugarsG = product.sugarsG || 0
  form.proteinG = product.proteinG || 0; form.saltG = product.saltG || 0
  form.fiberG = product.fiberG || 0
  form.additivesInput = Array.isArray(product.additives) ? product.additives.join(', ') : ''
  form.sweetenerType = product.sweetenerType || ''
  form.cocoaSolidsPct = product.cocoaSolidsPct || 0
  form.milkSolidsPct = product.milkSolidsPct || 0
  form.gmoStatus = product.gmoStatus || ''
  form.mayContainInput = Array.isArray(product.mayContain) ? product.mayContain.join(', ') : ''
  form.waterActivity = product.waterActivity || 0
  form.gtin = product.gtin || ''; form.hsCode = product.hsCode || ''
  form.primaryPackaging = product.primaryPackaging || ''
  form.innerPackConfig = product.innerPackConfig || ''
  form.palletConfig = product.palletConfig || ''
  form.isVegan = Boolean(product.isVegan); form.isGlutenFree = Boolean(product.isGlutenFree)
  form.isSugarFree = Boolean(product.isSugarFree); form.isKosher = Boolean(product.isKosher)
  form.isOrganic = Boolean(product.isOrganic)
  form.sampleMOQ = product.sampleMOQ || 0
  form.sampleLeadTime = product.sampleLeadTime || ''
  form.samplePrice = product.samplePrice || 0
  if (product.translations && typeof product.translations === 'object') {
    form.translations = emptyTranslations()
    const src = product.translations
    for (const loc of translationLocales) {
      if (src[loc] && typeof src[loc] === 'object') {
        for (const key of translationFieldKeys) {
          form.translations[loc][key] = src[loc][key] || ''
        }
        if (src[loc].flavors) {
          try {
            const arr = JSON.parse(src[loc].flavors)
            form.translations[loc].flavorsInput = Array.isArray(arr) ? arr.join(', ') : src[loc].flavors
          } catch { form.translations[loc].flavorsInput = src[loc].flavors }
        }
        if (src[loc].shapes) {
          try {
            const arr = JSON.parse(src[loc].shapes)
            form.translations[loc].shapesInput = Array.isArray(arr) ? arr.join(', ') : src[loc].shapes
          } catch { form.translations[loc].shapesInput = src[loc].shapes }
        }
      }
    }
  } else { form.translations = emptyTranslations() }
  hydrateSourceLocaleFromScalars()
  repairEnglishDuplicatesInZh()
  hydrateTranslationsFromLegacyScalars()
}

const fetchProducts = async () => {
  pending.value = true; error.value = ''
  try {
    const q = searchQuery.value.trim()
    const searchParam = q ? `&search=${encodeURIComponent(q)}` : ''
    const statusParam = statusFilter.value ? `&status=${encodeURIComponent(statusFilter.value)}` : ''
    const res = await api.get<any>(`/admin/products?page=${page.value}&limit=${pageSize}${searchParam}${statusParam}`)
    products.value = res.data || []; pagination.value = res.pagination
  } catch (err: any) { error.value = err?.message || t('errors.api.load_failed') }
  finally { pending.value = false }
}

const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) page.value += 1 }
const prevPage = () => { if (page.value > 1) page.value -= 1 }

const openCreateModal = () => { editingId.value = ''; translationLocale.value = SOURCE_LOCALE; resetForm(); formError.value = ''; actionMessage.value = ''; actionError.value = false; showModal.value = true }

const openCreateDrawer = () => {
  editingId.value = ''; showCreateDrawer.value = true
  aiDescription.value = ''; aiLanguage.value = 'zh'
  aiGenerating.value = false; aiHasData.value = false
  formError.value = ''; translationLocale.value = SOURCE_LOCALE
  resetForm()
}

const generateProduct = async () => {
  if (!aiDescription.value.trim() || aiGenerating.value) return
  aiGenerating.value = true; formError.value = ''; aiHasData.value = false
  try {
    const res = await api.adminAIGenerateProduct({ description: aiDescription.value, language: aiLanguage.value })
    if (res?.parseError) { formError.value = t('admin.products.ai_generate_failed'); return }
    const p = res.product
    if (!p) { formError.value = t('admin.products.ai_generate_failed'); return }
    form.name = p.name || ''; form.slug = p.slug || ''
    form.summary = p.summary || ''; form.description = p.description || ''
    form.category = p.category || ''; form.categorySlug = p.categorySlug || ''
    form.moq = Number(p.moq) || 0; form.basePrice = Math.max(0, Number(p.basePrice) || 0)
    form.stockQuantity = Math.max(0, Number(p.stockQuantity) || 0)
    form.leadTime = p.leadTime || ''
    form.oemAvailable = Boolean(p.oemAvailable); form.halalCertified = Boolean(p.halalCertified)
    form.featured = Boolean(p.featured); form.status = 'active'
    form.flavorsInput = Array.isArray(p.flavors) ? p.flavors.join(', ') : ''
    form.shapesInput = Array.isArray(p.shapes) ? p.shapes.join(', ') : ''
    form.certificationsInput = Array.isArray(p.certifications) ? p.certifications.join(', ') : ''
    form.ingredients = p.ingredients || ''; form.allergens = p.allergens || ''
    form.shelfLife = p.shelfLife || ''; form.storage = p.storage || ''
    form.netWeightPerPiece = Number(p.netWeightPerPiece) || 0
    form.netWeightPerPack = Number(p.netWeightPerPack) || 0
    form.grossWeightPerCarton = Number(p.grossWeightPerCarton) || 0
    form.piecesPerPack = Number(p.piecesPerPack) || 0
    form.packsPerCarton = Number(p.packsPerCarton) || 0
    form.productLengthMM = Number(p.productLengthMM) || 0
    form.productWidthMM = Number(p.productWidthMM) || 0
    form.productHeightMM = Number(p.productHeightMM) || 0
    form.energyKj = Number(p.energyKj) || 0; form.energyKcal = Number(p.energyKcal) || 0
    form.totalFatG = Number(p.totalFatG) || 0; form.saturatedFatG = Number(p.saturatedFatG) || 0
    form.carbohydratesG = Number(p.carbohydratesG) || 0; form.sugarsG = Number(p.sugarsG) || 0
    form.proteinG = Number(p.proteinG) || 0; form.saltG = Number(p.saltG) || 0
    form.fiberG = Number(p.fiberG) || 0
    form.additivesInput = Array.isArray(p.additives) ? p.additives.join(', ') : ''
    form.sweetenerType = p.sweetenerType || ''
    form.cocoaSolidsPct = Number(p.cocoaSolidsPct) || 0
    form.milkSolidsPct = Number(p.milkSolidsPct) || 0
    form.gmoStatus = p.gmoStatus || ''
    form.mayContainInput = Array.isArray(p.mayContain) ? p.mayContain.join(', ') : ''
    form.waterActivity = Number(p.waterActivity) || 0
    form.gtin = p.gtin || ''; form.hsCode = p.hsCode || ''
    form.primaryPackaging = p.primaryPackaging || ''
    form.innerPackConfig = p.innerPackConfig || ''
    form.palletConfig = p.palletConfig || ''
    form.isVegan = Boolean(p.isVegan); form.isGlutenFree = Boolean(p.isGlutenFree)
    form.isSugarFree = Boolean(p.isSugarFree); form.isKosher = Boolean(p.isKosher)
    form.isOrganic = Boolean(p.isOrganic)
    form.sampleMOQ = Number(p.sampleMOQ) || 0
    form.sampleLeadTime = p.sampleLeadTime || ''
    form.samplePrice = Number(p.samplePrice) || 0

    const transMap = res.translations || {}
    form.translations = emptyTranslations()
    const sourceLoc = aiLanguage.value
    applyTranslationFields(sourceLoc, {
      name: p.name,
      summary: p.summary,
      description: p.description,
      category: p.category,
      ingredients: p.ingredients,
      allergens: p.allergens,
      storage: p.storage,
      shelfLife: p.shelfLife,
      leadTime: p.leadTime,
      flavors: p.flavors,
      shapes: p.shapes,
    })
    for (const loc of translationLocales) {
      if (loc === sourceLoc) continue
      if (transMap[loc] && typeof transMap[loc] === 'object') {
        applyTranslationFields(loc, transMap[loc])
      }
    }
    syncScalarsFromSourceLocale()
    aiHasData.value = true
  } catch (err: any) { formError.value = err?.message || t('errors.api.request_failed') }
  finally { aiGenerating.value = false }
}

const openEditModal = async (product: any) => {
  editingId.value = product.id; formError.value = ''; actionMessage.value = ''; actionError.value = false
  translationLocale.value = SOURCE_LOCALE
  try { const detail = await api.get<any>(`/admin/products/${product.id}`); fillFormFromProduct(detail); showModal.value = true }
  catch (err: any) {
    console.warn(`[products] Failed to load detail for product ${product.id}, falling back to list data`)
    fillFormFromProduct(product); showModal.value = true
    actionMessage.value = err?.message || t('errors.api.load_failed')
    actionError.value = true
  }
}

const closeModal = () => { showModal.value = false; saving.value = false; formError.value = ''; uploadingImage.value = false; uploadError.value = '' }

const buildPayload = () => {
  syncScalarsFromSourceLocale()
  return {
  name: form.name, slug: form.slug, summary: form.summary, description: form.description,
  category: form.category, categorySlug: form.categorySlug, thumbnail: form.thumbnail, ogImage: form.ogImage,
  moq: form.moq, basePrice: Math.max(0, Number(form.basePrice) || 0), stockQuantity: Math.max(0, Number(form.stockQuantity) || 0), leadTime: form.leadTime,
  oemAvailable: form.oemAvailable, halalCertified: form.halalCertified, featured: form.featured, status: form.status,
  images: parseCSV(form.imagesInput), flavors: parseCSV(form.flavorsInput), shapes: parseCSV(form.shapesInput),
  certifications: parseCSV(form.certificationsInput), ingredients: form.ingredients, allergens: form.allergens,
  shelfLife: form.shelfLife, storage: form.storage,
  netWeightPerPiece: form.netWeightPerPiece, netWeightPerPack: form.netWeightPerPack, grossWeightPerCarton: form.grossWeightPerCarton,
  piecesPerPack: form.piecesPerPack, packsPerCarton: form.packsPerCarton,
  productLengthMM: form.productLengthMM, productWidthMM: form.productWidthMM, productHeightMM: form.productHeightMM,
  energyKj: form.energyKj, energyKcal: form.energyKcal, totalFatG: form.totalFatG, saturatedFatG: form.saturatedFatG,
  carbohydratesG: form.carbohydratesG, sugarsG: form.sugarsG, proteinG: form.proteinG, saltG: form.saltG, fiberG: form.fiberG,
  additives: parseCSV(form.additivesInput), sweetenerType: form.sweetenerType, cocoaSolidsPct: form.cocoaSolidsPct,
  milkSolidsPct: form.milkSolidsPct, gmoStatus: form.gmoStatus, mayContain: parseCSV(form.mayContainInput),
  waterActivity: form.waterActivity, gtin: form.gtin, hsCode: form.hsCode,
  primaryPackaging: form.primaryPackaging, innerPackConfig: form.innerPackConfig, palletConfig: form.palletConfig,
  isVegan: form.isVegan, isGlutenFree: form.isGlutenFree, isSugarFree: form.isSugarFree, isKosher: form.isKosher, isOrganic: form.isOrganic,
  sampleMOQ: form.sampleMOQ, sampleLeadTime: form.sampleLeadTime, samplePrice: form.samplePrice,
  translations: buildTranslationsPayload()
  }
}

const saveProduct = async () => {
  const sourceName = form.translations[SOURCE_LOCALE]?.name?.trim?.() || ''
  if (!sourceName) { formError.value = t('admin.products.name_required'); return }
  saving.value = true; formError.value = ''; actionMessage.value = ''; actionError.value = false
  const payload = buildPayload()
  try {
    if (editingId.value) { await api.put(`/admin/products/${editingId.value}`, payload); actionMessage.value = t('admin.products.updated_success') }
    else { await api.post('/admin/products', payload); actionMessage.value = t('admin.products.created_success') }
    showModal.value = false; showCreateDrawer.value = false; aiHasData.value = false
    await fetchProducts()
  } catch (err: any) { formError.value = err?.message || t('errors.api.save_failed') }
  finally { saving.value = false }
}

const deleteProduct = async (id: string) => {
  if (!confirm(t('admin.products.confirm_delete'))) return
  actionMessage.value = ''; actionError.value = false
  try { await api.del(`/admin/products/${id}`); actionMessage.value = t('admin.products.deleted_success'); await fetchProducts() }
  catch (err: any) { actionError.value = true; actionMessage.value = err?.message || t('errors.api.delete_failed') }
}

watch(page, fetchProducts)
onMounted(async () => {
  const q = route.query.search as string
  if (q) searchQuery.value = q
  await fetchCategories()
  await fetchProducts()
})
</script>
