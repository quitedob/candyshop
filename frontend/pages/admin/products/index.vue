<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.products.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.products.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0 flex items-center gap-3">
        <button type="button" class="inline-flex items-center justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50" @click="openCreateModal">
          {{ t('admin.products.add_product') }}
        </button>
        <button type="button" class="inline-flex items-center gap-2 justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700" @click="openCreateDrawer">
          <Icon name="heroicons:sparkles" class="h-4 w-4" />
          {{ t('admin.products.ai_import') }}
        </button>
      </div>
    </div>

    <div class="mt-8 overflow-hidden rounded-lg bg-white shadow ring-1 ring-black ring-opacity-5">
      <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
          <tr>
            <th class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{{ t('admin.products.col_product') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.products.col_category') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.products.col_moq') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.products.col_stock') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.products.col_status') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <tr v-if="pending">
            <td colspan="6" class="py-5 text-center text-sm text-gray-500">{{ t('admin.products.loading') }}</td>
          </tr>
          <tr v-else-if="error">
            <td colspan="6" class="py-5 text-center text-sm text-red-600">{{ error }}</td>
          </tr>
          <tr v-else-if="products.length === 0">
            <td colspan="6" class="py-5 text-center text-sm text-gray-500">{{ t('admin.products.no_data') }}</td>
          </tr>
          <tr v-else v-for="product in products" :key="product.id">
            <td class="py-4 pl-4 pr-3 text-sm sm:pl-6">
              <div class="flex items-center">
                <div class="h-10 w-10 flex-shrink-0">
                  <img v-if="product.thumbnail" class="h-10 w-10 rounded-full object-cover" :src="product.thumbnail" alt="" />
                  <div v-else class="flex h-10 w-10 items-center justify-center rounded-full bg-gray-200 font-bold text-gray-500">
                    {{ product.name?.charAt(0) }}
                  </div>
                </div>
                <div class="ml-4">
                  <div class="font-medium text-gray-900">{{ product.name }}</div>
                  <div class="text-gray-500">{{ product.slug }}</div>
                </div>
              </div>
            </td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ product.category || '-' }}</td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ product.moq || 0 }}</td>
            <td class="whitespace-nowrap px-3 py-4 text-sm" :class="(product.stockQuantity || 0) > 0 ? 'text-light' : 'text-error'">
              {{ product.stockQuantity || 0 }}
            </td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-light">
              <span class="badge" :class="product.status === 'active' ? 'badge-success' : 'badge-default'">
                {{ enumLabel('product_status', product.status, 'active') }}
              </span>
            </td>
            <td class="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
              <button type="button" class="text-orange-600 hover:text-orange-900" @click="openEditModal(product)">{{ t('admin.products.edit') }}</button>
              <button type="button" class="ml-4 text-red-600 hover:text-red-900" @click="deleteProduct(product.id)">{{ t('admin.products.delete') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="pagination" class="mt-4 flex items-center justify-between rounded-lg border-t border-gray-200 bg-white px-4 py-3 shadow sm:px-6">
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

    <div v-if="showModal" class="fixed inset-0 z-10 overflow-y-auto" role="dialog" aria-modal="true">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <button type="button" class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity w-full border-0 cursor-pointer" @click="closeModal" :aria-label="t('close')"></button>
        <span class="hidden sm:inline-block sm:h-screen sm:align-middle" aria-hidden="true">&#8203;</span>
        <div class="inline-block w-full transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left align-bottom shadow-xl sm:my-8 sm:max-w-3xl sm:p-6 sm:align-middle">
          <h3 class="text-lg font-medium leading-6 text-gray-900">{{ editingId ? t('admin.products.edit_product') : t('admin.products.create_product') }}</h3>

          <form class="mt-4" @submit.prevent="saveProduct">
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
        </div>
      </div>
    </div>

    <!-- AI Generate Drawer -->
    <Drawer :open="showCreateDrawer" width="2xl" @close="showCreateDrawer = false">
      <template #title>{{ t('admin.products.add_product') }} — AI</template>

      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6" style="min-height: 60vh;">
        <!-- LEFT: Input panel -->
        <div class="lg:col-span-5 space-y-4">
          <div>
            <label for="ai-product-desc" class="block text-sm font-medium text-gray-700">{{ t('admin.products.ai_description_label') }}</label>
            <textarea
              id="ai-product-desc"
              v-model="aiDescription"
              rows="12"
              class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-orange-500 focus:outline-none focus:ring-1 focus:ring-orange-500"
              :placeholder="t('admin.products.ai_description_placeholder')"
            ></textarea>
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

        <!-- RIGHT: Existing form template auto-filled by AI -->
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
import { reactive, ref, watch, onMounted } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const api = useApi()
const { t } = useI18n()
const { enumLabel } = useDisplay()

const products = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20

const showModal = ref(false)
const editingId = ref('')
const saving = ref(false)
const translationLocale = ref('en')
const aiTranslating = ref(false)
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

const triggerUpload = async (field: 'thumbnail' | 'images') => {
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

const translationLocales = ['en', 'zh', 'ko', 'ar', 'ja', 'th', 'vi', 'id', 'ms']

function emptyTranslations() {
  const result: Record<string, { name: string; summary: string; description: string }> = {}
  for (const loc of translationLocales) {
    result[loc] = { name: '', summary: '', description: '' }
  }
  return result
}

const form = reactive({
  name: '',
  slug: '',
  summary: '',
  description: '',
  category: '',
  categorySlug: '',
  thumbnail: '',
  moq: 0,
  basePrice: 0,
  stockQuantity: 0,
  leadTime: '',
  oemAvailable: false,
  halalCertified: false,
  featured: false,
  status: 'active',
  imagesInput: '',
  flavorsInput: '',
  shapesInput: '',
  certificationsInput: '',
  ingredients: '',
  allergens: '',
  shelfLife: '',
  storage: '',
  // Weight & Measurement
  netWeightPerPiece: 0,
  netWeightPerPack: 0,
  grossWeightPerCarton: 0,
  piecesPerPack: 0,
  packsPerCarton: 0,
  // Dimensions
  productLengthMM: 0,
  productWidthMM: 0,
  productHeightMM: 0,
  // Nutrition
  energyKj: 0,
  energyKcal: 0,
  totalFatG: 0,
  saturatedFatG: 0,
  carbohydratesG: 0,
  sugarsG: 0,
  proteinG: 0,
  saltG: 0,
  fiberG: 0,
  // Ingredient Compliance
  additivesInput: '',
  sweetenerType: '',
  cocoaSolidsPct: 0,
  milkSolidsPct: 0,
  gmoStatus: '',
  mayContainInput: '',
  waterActivity: 0,
  // Trade & Barcode
  gtin: '',
  hsCode: '',
  // Packaging
  primaryPackaging: '',
  innerPackConfig: '',
  palletConfig: '',
  // Dietary
  isVegan: false,
  isGlutenFree: false,
  isSugarFree: false,
  isKosher: false,
  isOrganic: false,
  // Sample Specs
  sampleMOQ: 0,
  sampleLeadTime: '',
  samplePrice: 0,
  translations: emptyTranslations()
} as any)

const localeTabLabelMap: Record<string, string> = {
  en: 'admin.products.locale_tab_en',
  zh: 'admin.products.locale_tab_zh',
  ko: 'admin.products.locale_tab_ko',
  ar: 'admin.products.locale_tab_ar',
  ja: 'admin.products.locale_tab_ja',
  th: 'admin.products.locale_tab_th',
  vi: 'admin.products.locale_tab_vi',
  id: 'admin.products.locale_tab_id',
  ms: 'admin.products.locale_tab_ms',
}

const localeTabLabel = (loc: string) => t(localeTabLabelMap[loc] || loc)

const parseCSV = (value: string) => value.split(',').map(v => v.trim()).filter(Boolean)

const buildTranslationsPayload = () => {
  const result: Record<string, Record<string, string>> = {}
  for (const loc of translationLocales) {
    const t = form.translations[loc]
    if (!t) continue
    const entry: Record<string, string> = {}
    if (t.name?.trim()) entry.name = t.name.trim()
    if (t.summary?.trim()) entry.summary = t.summary.trim()
    if (t.description?.trim()) entry.description = t.description.trim()
    if (Object.keys(entry).length > 0) result[loc] = entry
  }
  return Object.keys(result).length > 0 ? result : null
}

const aiTranslateAll = async () => {
  if (!editingId.value || aiTranslating.value) return
  aiTranslating.value = true
  try {
    const targetLocales = translationLocales.filter(l => l !== 'zh')
    const res = await api.post<any>(`/admin/products/${editingId.value}/ai-translate`, {
      productId: editingId.value,
      targetLocales,
    })
    if (res?.translations && typeof res.translations === 'object') {
      for (const loc of targetLocales) {
        const fields = res.translations[loc]
        if (fields && typeof fields === 'object') {
          if (fields.name) form.translations[loc].name = fields.name
          if (fields.summary) form.translations[loc].summary = fields.summary
          if (fields.description) form.translations[loc].description = fields.description
        }
      }
    }
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.request_failed')
  } finally {
    aiTranslating.value = false
  }
}

const resetForm = () => {
  form.name = ''
  form.slug = ''
  form.summary = ''
  form.description = ''
  form.category = ''
  form.categorySlug = ''
  form.thumbnail = ''
  form.moq = 0
  form.basePrice = 0
  form.stockQuantity = 0
  form.leadTime = ''
  form.oemAvailable = false
  form.halalCertified = false
  form.featured = false
  form.status = 'active'
  form.imagesInput = ''
  form.flavorsInput = ''
  form.shapesInput = ''
  form.certificationsInput = ''
  form.ingredients = ''
  form.allergens = ''
  form.shelfLife = ''
  form.storage = ''
  form.netWeightPerPiece = 0
  form.netWeightPerPack = 0
  form.grossWeightPerCarton = 0
  form.piecesPerPack = 0
  form.packsPerCarton = 0
  form.productLengthMM = 0
  form.productWidthMM = 0
  form.productHeightMM = 0
  form.energyKj = 0
  form.energyKcal = 0
  form.totalFatG = 0
  form.saturatedFatG = 0
  form.carbohydratesG = 0
  form.sugarsG = 0
  form.proteinG = 0
  form.saltG = 0
  form.fiberG = 0
  form.additivesInput = ''
  form.sweetenerType = ''
  form.cocoaSolidsPct = 0
  form.milkSolidsPct = 0
  form.gmoStatus = ''
  form.mayContainInput = ''
  form.waterActivity = 0
  form.gtin = ''
  form.hsCode = ''
  form.primaryPackaging = ''
  form.innerPackConfig = ''
  form.palletConfig = ''
  form.isVegan = false
  form.isGlutenFree = false
  form.isSugarFree = false
  form.isKosher = false
  form.isOrganic = false
  form.sampleMOQ = 0
  form.sampleLeadTime = ''
  form.samplePrice = 0
  form.translations = emptyTranslations()
}

const fillFormFromProduct = (product: any) => {
  form.name = product.name || ''
  form.slug = product.slug || ''
  form.summary = product.summary || ''
  form.description = product.description || ''
  form.category = product.category || ''
  form.categorySlug = product.categorySlug || ''
  form.thumbnail = product.thumbnail || ''
  form.moq = product.moq || 0
  form.basePrice = product.basePrice || 0
  form.stockQuantity = product.stockQuantity || 0
  form.leadTime = product.leadTime || ''
  form.oemAvailable = Boolean(product.oemAvailable)
  form.halalCertified = Boolean(product.halalCertified)
  form.featured = Boolean(product.featured)
  form.status = product.status || 'active'
  form.imagesInput = Array.isArray(product.images) ? product.images.join(', ') : ''
  form.flavorsInput = Array.isArray(product.flavors) ? product.flavors.join(', ') : ''
  form.shapesInput = Array.isArray(product.shapes) ? product.shapes.join(', ') : ''
  form.certificationsInput = Array.isArray(product.certifications) ? product.certifications.join(', ') : ''
  form.ingredients = product.ingredients || ''
  form.allergens = product.allergens || ''
  form.shelfLife = product.shelfLife || ''
  form.storage = product.storage || ''
  form.netWeightPerPiece = product.netWeightPerPiece || 0
  form.netWeightPerPack = product.netWeightPerPack || 0
  form.grossWeightPerCarton = product.grossWeightPerCarton || 0
  form.piecesPerPack = product.piecesPerPack || 0
  form.packsPerCarton = product.packsPerCarton || 0
  form.productLengthMM = product.productLengthMM || 0
  form.productWidthMM = product.productWidthMM || 0
  form.productHeightMM = product.productHeightMM || 0
  form.energyKj = product.energyKj || 0
  form.energyKcal = product.energyKcal || 0
  form.totalFatG = product.totalFatG || 0
  form.saturatedFatG = product.saturatedFatG || 0
  form.carbohydratesG = product.carbohydratesG || 0
  form.sugarsG = product.sugarsG || 0
  form.proteinG = product.proteinG || 0
  form.saltG = product.saltG || 0
  form.fiberG = product.fiberG || 0
  form.additivesInput = Array.isArray(product.additives) ? product.additives.join(', ') : ''
  form.sweetenerType = product.sweetenerType || ''
  form.cocoaSolidsPct = product.cocoaSolidsPct || 0
  form.milkSolidsPct = product.milkSolidsPct || 0
  form.gmoStatus = product.gmoStatus || ''
  form.mayContainInput = Array.isArray(product.mayContain) ? product.mayContain.join(', ') : ''
  form.waterActivity = product.waterActivity || 0
  form.gtin = product.gtin || ''
  form.hsCode = product.hsCode || ''
  form.primaryPackaging = product.primaryPackaging || ''
  form.innerPackConfig = product.innerPackConfig || ''
  form.palletConfig = product.palletConfig || ''
  form.isVegan = Boolean(product.isVegan)
  form.isGlutenFree = Boolean(product.isGlutenFree)
  form.isSugarFree = Boolean(product.isSugarFree)
  form.isKosher = Boolean(product.isKosher)
  form.isOrganic = Boolean(product.isOrganic)
  form.sampleMOQ = product.sampleMOQ || 0
  form.sampleLeadTime = product.sampleLeadTime || ''
  form.samplePrice = product.samplePrice || 0
  if (product.translations && typeof product.translations === 'object') {
    const src = product.translations
    form.translations = emptyTranslations()
    for (const loc of translationLocales) {
      if (src[loc] && typeof src[loc] === 'object') {
        form.translations[loc].name = src[loc].name || ''
        form.translations[loc].summary = src[loc].summary || ''
        form.translations[loc].description = src[loc].description || ''
      }
    }
  } else {
    form.translations = emptyTranslations()
  }
}

const fetchProducts = async () => {
  pending.value = true; error.value = ''
  try {
    const res = await api.get<any>(`/admin/products?page=${page.value}&limit=${pageSize}`)
    products.value = res.data || []; pagination.value = res.pagination
  } catch (err: any) { error.value = err?.message || t('errors.api.load_failed') }
  finally { pending.value = false }
}

const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) page.value += 1 }
const prevPage = () => { if (page.value > 1) page.value -= 1 }

const openCreateModal = () => { editingId.value = ''; translationLocale.value = 'en'; resetForm(); formError.value = ''; actionMessage.value = ''; actionError.value = false; showModal.value = true }

const openCreateDrawer = () => {
  showCreateDrawer.value = true
  aiDescription.value = ''
  aiLanguage.value = 'zh'
  aiGenerating.value = false
  aiHasData.value = false
  formError.value = ''
  translationLocale.value = 'en'
  resetForm()
}

const generateProduct = async () => {
  if (!aiDescription.value.trim() || aiGenerating.value) return
  aiGenerating.value = true
  formError.value = ''
  try {
    const res = await api.adminAIGenerateProduct({ description: aiDescription.value, language: aiLanguage.value })
    if (res?.parseError) {
      formError.value = t('admin.products.ai_generate_failed')
      return
    }
    const p = res.product
    if (!p) {
      formError.value = t('admin.products.ai_generate_failed')
      return
    }
    // Fill form from AI result
    form.name = p.name || ''
    form.slug = p.slug || ''
    form.summary = p.summary || ''
    form.description = p.description || ''
    form.category = p.category || ''
    form.categorySlug = p.categorySlug || ''
    form.moq = Number(p.moq) || 0
    form.basePrice = Math.max(0, Number(p.basePrice) || 0)
    form.stockQuantity = Math.max(0, Number(p.stockQuantity) || 0)
    form.leadTime = p.leadTime || ''
    form.oemAvailable = Boolean(p.oemAvailable)
    form.halalCertified = Boolean(p.halalCertified)
    form.featured = Boolean(p.featured)
    form.status = 'active'
    form.flavorsInput = Array.isArray(p.flavors) ? p.flavors.join(', ') : ''
    form.shapesInput = Array.isArray(p.shapes) ? p.shapes.join(', ') : ''
    form.certificationsInput = Array.isArray(p.certifications) ? p.certifications.join(', ') : ''
    form.ingredients = p.ingredients || ''
    form.allergens = p.allergens || ''
    form.shelfLife = p.shelfLife || ''
    form.storage = p.storage || ''
    form.netWeightPerPiece = Number(p.netWeightPerPiece) || 0
    form.netWeightPerPack = Number(p.netWeightPerPack) || 0
    form.grossWeightPerCarton = Number(p.grossWeightPerCarton) || 0
    form.piecesPerPack = Number(p.piecesPerPack) || 0
    form.packsPerCarton = Number(p.packsPerCarton) || 0
    form.productLengthMM = Number(p.productLengthMM) || 0
    form.productWidthMM = Number(p.productWidthMM) || 0
    form.productHeightMM = Number(p.productHeightMM) || 0
    form.energyKj = Number(p.energyKj) || 0
    form.energyKcal = Number(p.energyKcal) || 0
    form.totalFatG = Number(p.totalFatG) || 0
    form.saturatedFatG = Number(p.saturatedFatG) || 0
    form.carbohydratesG = Number(p.carbohydratesG) || 0
    form.sugarsG = Number(p.sugarsG) || 0
    form.proteinG = Number(p.proteinG) || 0
    form.saltG = Number(p.saltG) || 0
    form.fiberG = Number(p.fiberG) || 0
    form.additivesInput = Array.isArray(p.additives) ? p.additives.join(', ') : ''
    form.sweetenerType = p.sweetenerType || ''
    form.cocoaSolidsPct = Number(p.cocoaSolidsPct) || 0
    form.milkSolidsPct = Number(p.milkSolidsPct) || 0
    form.gmoStatus = p.gmoStatus || ''
    form.mayContainInput = Array.isArray(p.mayContain) ? p.mayContain.join(', ') : ''
    form.waterActivity = Number(p.waterActivity) || 0
    form.gtin = p.gtin || ''
    form.hsCode = p.hsCode || ''
    form.primaryPackaging = p.primaryPackaging || ''
    form.innerPackConfig = p.innerPackConfig || ''
    form.palletConfig = p.palletConfig || ''
    form.isVegan = Boolean(p.isVegan)
    form.isGlutenFree = Boolean(p.isGlutenFree)
    form.isSugarFree = Boolean(p.isSugarFree)
    form.isKosher = Boolean(p.isKosher)
    form.isOrganic = Boolean(p.isOrganic)
    form.sampleMOQ = Number(p.sampleMOQ) || 0
    form.sampleLeadTime = p.sampleLeadTime || ''
    form.samplePrice = Number(p.samplePrice) || 0

    // Fill AI translations into form.translations
    const transMap = res.translations || {}
    form.translations = emptyTranslations()
    // Populate source locale from the generated content
    const sourceLoc = aiLanguage.value
    if (sourceLoc && form.translations[sourceLoc]) {
      form.translations[sourceLoc].name = form.name
      form.translations[sourceLoc].summary = form.summary
      form.translations[sourceLoc].description = form.description
    }
    // Populate translations from backend
    for (const loc of translationLocales) {
      if (transMap[loc] && typeof transMap[loc] === 'object') {
        form.translations[loc].name = transMap[loc].name || form.translations[loc].name
        form.translations[loc].summary = transMap[loc].summary || form.translations[loc].summary
        form.translations[loc].description = transMap[loc].description || form.translations[loc].description
      }
    }

    aiHasData.value = true
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.request_failed')
  } finally {
    aiGenerating.value = false
  }
}

const openEditModal = async (product: any) => {
  editingId.value = product.id; formError.value = ''; actionMessage.value = ''; actionError.value = false
  try { const detail = await api.get<any>(`/admin/products/${product.id}`); fillFormFromProduct(detail); showModal.value = true }
  catch { fillFormFromProduct(product); showModal.value = true }
}

const closeModal = () => { showModal.value = false; saving.value = false; formError.value = '' }

const buildPayload = () => ({
  name: form.name, slug: form.slug, summary: form.summary, description: form.description,
  category: form.category, categorySlug: form.categorySlug, thumbnail: form.thumbnail,
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
})

const saveProduct = async () => {
  if (!form.name.trim()) { formError.value = t('admin.products.name_required'); return }
  saving.value = true; formError.value = ''; actionMessage.value = ''; actionError.value = false
  const payload = buildPayload()
  try {
    if (editingId.value) { await api.put(`/admin/products/${editingId.value}`, payload); actionMessage.value = t('admin.products.updated_success') }
    else { await api.post('/admin/products', payload); actionMessage.value = t('admin.products.created_success') }
    showModal.value = false
    showCreateDrawer.value = false
    aiHasData.value = false
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
onMounted(fetchProducts)
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-xl);
}

.page-title {
  font-size: var(--text-2xl);
  font-weight: 600;
  color: var(--color-primary);
}

.page-subtitle {
  margin-top: var(--spacing-xs);
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.table-container {
  overflow: hidden;
  border-radius: var(--radius-lg);
  background: white;
  box-shadow: var(--shadow-md);
  margin-top: var(--spacing-xl);
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  padding: var(--spacing-md);
  text-align: left;
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-bg-alt);
  border-bottom: 1px solid var(--color-border);
}

.data-table td {
  padding: var(--spacing-md);
  font-size: var(--text-sm);
  color: var(--color-text);
  border-bottom: 1px solid var(--color-border);
}

.data-table tbody tr:hover {
  background: var(--color-bg);
}

.data-table tbody tr:last-child td {
  border-bottom: none;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: var(--spacing-lg);
  padding: var(--spacing-md) var(--spacing-lg);
  background: white;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
}

.link {
  color: var(--color-highlight);
  font-weight: 500;
  transition: color var(--transition-fast);
}

.link:hover {
  color: var(--color-highlight-hover);
}

.link-danger {
  color: var(--color-error);
}

.link-danger:hover {
  color: var(--color-error);
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-lg);
  margin-top: var(--spacing-lg);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.form-group.col-span-2 {
  grid-column: span 2;
}

.form-label {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
}

.form-input {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-base);
  color: var(--color-text);
  background: var(--color-bg);
  border: 1.5px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(var(--color-highlight-rgb), 0.1);
}

.form-textarea {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-base);
  color: var(--color-text);
  background: var(--color-bg);
  border: 1.5px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  resize: vertical;
}

.form-textarea:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(var(--color-highlight-rgb), 0.1);
}

.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl);
}

.modal-container {
  position: relative;
  width: 100%;
  max-width: 800px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
}

.modal-content {
  position: relative;
  background: white;
  border-radius: var(--radius-xl);
  padding: var(--spacing-xl);
  box-shadow: var(--shadow-xl);
  animation: modalIn 0.3s ease forwards;
}

@keyframes modalIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.modal-title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-primary);
}
</style>
