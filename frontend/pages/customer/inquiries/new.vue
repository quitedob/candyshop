<template>
  <div class="inquiry-new">
    <header class="inquiry-new__header">
      <NuxtLink :to="localePath('/customer/inquiries')" class="inquiry-new__back">
        <Icon name="heroicons:arrow-left" class="h-4 w-4" aria-hidden="true" />
        {{ t('customer.inquiries_new.back') }}
      </NuxtLink>
      <div class="inquiry-new__header-main">
        <div>
          <h1 class="inquiry-new__title">{{ t('customer.inquiries_new.title') }}</h1>
          <p class="inquiry-new__subtitle">{{ t('customer.inquiries_new.subtitle') }}</p>
        </div>
      </div>
    </header>

    <form class="inquiry-form" @submit.prevent="handleSubmit">
      <!-- 感兴趣的产品 -->
      <section class="inquiry-section">
        <header class="inquiry-section__head">
          <span class="inquiry-section__icon"><Icon name="heroicons:shopping-bag" class="h-5 w-5" /></span>
          <div class="inquiry-section__head-text">
            <h2 class="inquiry-section__title">
              {{ t('customer.inquiries_new.interested_products') }}
              <span class="inquiry-section__required">*</span>
            </h2>
            <p class="inquiry-section__hint">{{ t('customer.inquiries_new.interested_products_hint') }}</p>
          </div>
        </header>

        <div class="inquiry-products__toolbar">
          <button type="button" class="inquiry-btn inquiry-btn--primary" @click="openPicker">
            <Icon name="heroicons:plus-circle" class="h-4 w-4" aria-hidden="true" />
            {{ t('customer.inquiries_new.pick_from_catalog') }}
          </button>
          <NuxtLink :to="localePath('/customer/products')" class="inquiry-btn inquiry-btn--ghost">
            <Icon name="heroicons:arrow-top-right-on-square" class="h-4 w-4" aria-hidden="true" />
            {{ t('customer.inquiries_new.browse_catalog') }}
          </NuxtLink>
        </div>

        <!-- 已选产品 -->
        <div v-if="selectedProducts.length" class="inquiry-products__list">
          <article v-for="(item, index) in selectedProducts" :key="itemKey(item, index)" class="inquiry-product-chip">
            <div class="inquiry-product-chip__media">
              <img v-if="item.thumbnail" :src="item.thumbnail" :alt="item.name" class="inquiry-product-chip__img" />
              <Icon v-else name="heroicons:cube" class="h-5 w-5 text-gray-400" aria-hidden="true" />
            </div>
            <div class="inquiry-product-chip__body">
              <p class="inquiry-product-chip__name">{{ item.name }}</p>
              <p v-if="item.category" class="inquiry-product-chip__meta">{{ item.category }}</p>
            </div>
            <div class="inquiry-product-chip__actions">
              <NuxtLink
                v-if="item.id"
                :to="localePath(`/customer/products/${item.slug || item.id}`)"
                class="inquiry-product-chip__link"
                target="_blank"
              >
                {{ t('customer.inquiries_new.view_product') }}
                <Icon name="heroicons:arrow-top-right-on-square" class="h-3.5 w-3.5" aria-hidden="true" />
              </NuxtLink>
              <button
                type="button"
                class="inquiry-product-chip__remove"
                :aria-label="t('customer.inquiries_new.remove_product')"
                @click="removeProduct(index)"
              >
                <Icon name="heroicons:x-mark" class="h-4 w-4" />
              </button>
            </div>
          </article>
        </div>

        <div v-else class="inquiry-products__empty">
          <Icon name="heroicons:inbox" class="h-8 w-8 text-gray-300" aria-hidden="true" />
          <p>{{ t('customer.inquiries_new.no_products_selected') }}</p>
        </div>

        <!-- 手动添加 -->
        <div class="inquiry-manual">
          <label for="manual-product" class="inquiry-manual__label">{{ t('customer.inquiries_new.add_manual_product') }}</label>
          <div class="inquiry-manual__row">
            <input
              id="manual-product"
              v-model="manualProductInput"
              type="text"
              class="inquiry-input"
              :placeholder="t('customer.inquiries_new.add_manual_placeholder')"
              @keydown.enter.prevent="addManualProduct"
            />
            <button type="button" class="inquiry-btn inquiry-btn--secondary" @click="addManualProduct">
              {{ t('customer.inquiries_new.add_manual_btn') }}
            </button>
          </div>
        </div>
      </section>

      <!-- 订单详情 -->
      <section class="inquiry-section">
        <header class="inquiry-section__head">
          <span class="inquiry-section__icon"><Icon name="heroicons:clipboard-document-list" class="h-5 w-5" /></span>
          <div>
            <h2 class="inquiry-section__title">{{ t('customer.inquiries_new.section_details') }}</h2>
          </div>
        </header>
        <div class="inquiry-grid">
          <div class="inquiry-field">
            <label for="quantity" class="inquiry-field__label">
              {{ t('customer.inquiries_new.estimated_quantity') }}
              <span class="inquiry-section__required">*</span>
            </label>
            <input
              id="quantity"
              v-model="form.estimatedQuantity"
              type="text"
              class="inquiry-input"
              :placeholder="t('customer.inquiries_new.estimated_quantity_placeholder')"
              required
            />
          </div>
          <div class="inquiry-field">
            <label for="delivery" class="inquiry-field__label">{{ t('customer.inquiries_new.expected_delivery') }}</label>
            <input
              id="delivery"
              v-model="form.expectedDelivery"
              type="text"
              class="inquiry-input"
              :placeholder="t('customer.inquiries_new.expected_delivery_placeholder')"
            />
          </div>
          <div class="inquiry-field inquiry-field--full">
            <span class="inquiry-field__label">{{ t('customer.inquiries_new.oem_question') }}</span>
            <div class="inquiry-radio-group">
              <label class="inquiry-radio">
                <input id="oem-yes" v-model="form.oemNeeded" type="radio" :value="true" />
                <span>{{ t('customer.inquiries_new.yes') }}</span>
              </label>
              <label class="inquiry-radio">
                <input id="oem-no" v-model="form.oemNeeded" type="radio" :value="false" />
                <span>{{ t('customer.inquiries_new.no') }}</span>
              </label>
            </div>
          </div>
          <div class="inquiry-field inquiry-field--full">
            <label for="message" class="inquiry-field__label">{{ t('customer.inquiries_new.additional_details') }}</label>
            <textarea
              id="message"
              v-model="form.message"
              rows="4"
              class="inquiry-input inquiry-input--area"
              :placeholder="t('customer.inquiries_new.additional_details_placeholder')"
            />
          </div>
        </div>
      </section>

      <!-- 附件 -->
      <section class="inquiry-section">
        <header class="inquiry-section__head">
          <span class="inquiry-section__icon"><Icon name="heroicons:paper-clip" class="h-5 w-5" /></span>
          <div>
            <h2 class="inquiry-section__title">{{ t('customer.inquiries_new.attachments') }}</h2>
            <p class="inquiry-section__hint">{{ t('customer.inquiries_new.attachments_hint') }}</p>
          </div>
        </header>
        <InputFile
          id="inquiry-attachments"
          :accept="INQUIRY_ACCEPT_ATTR"
          :max-files="INQUIRY_MAX_FILES"
          :max-size="INQUIRY_MAX_FILE_MB"
          :error="fileError"
          @files-selected="handleFilesSelected"
          @error="fileError = $event"
        />
        <ul v-if="selectedFiles.length" class="inquiry-files">
          <li v-for="(file, index) in selectedFiles" :key="`${file.name}-${index}`" class="inquiry-files__item">
            <Icon name="heroicons:document" class="h-4 w-4 shrink-0 text-gray-400" aria-hidden="true" />
            <span class="inquiry-files__name">{{ file.name }}</span>
            <button
              type="button"
              class="inquiry-files__remove"
              :aria-label="t('customer.inquiries_new.remove_file')"
              @click="removeFile(index)"
            >
              <Icon name="heroicons:x-mark" class="h-4 w-4" />
            </button>
          </li>
        </ul>
      </section>

      <div v-if="error" class="inquiry-alert inquiry-alert--error" role="alert">
        <Icon name="heroicons:exclamation-circle" class="h-5 w-5 shrink-0" aria-hidden="true" />
        {{ error }}
      </div>
      <div v-if="success" class="inquiry-alert inquiry-alert--success" role="status">
        <Icon name="heroicons:check-circle" class="h-5 w-5 shrink-0" aria-hidden="true" />
        {{ t('customer.inquiries_new.success') }}
      </div>

      <footer class="inquiry-form__footer">
        <NuxtLink :to="localePath('/customer/inquiries')" class="inquiry-btn inquiry-btn--ghost">
          {{ t('customer.inquiries_new.cancel') }}
        </NuxtLink>
        <button type="submit" class="inquiry-btn inquiry-btn--primary" :disabled="loading">
          <Icon v-if="loading" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" aria-hidden="true" />
          {{ loading ? t('customer.inquiries_new.submitting') : t('customer.inquiries_new.submit') }}
        </button>
      </footer>
    </form>

    <!-- 产品选择抽屉 -->
    <Drawer :open="pickerOpen" width="3xl" @close="closePicker">
      <template #title>{{ t('customer.inquiries_new.picker_title') }}</template>

      <div class="picker-search">
        <Icon name="heroicons:magnifying-glass" class="picker-search__icon" aria-hidden="true" />
        <input
          v-model="pickerSearch"
          type="search"
          class="picker-search__input"
          :placeholder="t('customer.inquiries_new.picker_search')"
        />
      </div>

      <div v-if="pickerLoading" class="picker-state">{{ t('customer.inquiries_new.picker_loading') }}</div>
      <div v-else-if="!catalogProducts.length" class="picker-state">{{ t('customer.inquiries_new.picker_empty') }}</div>
      <ul v-else class="picker-list" role="listbox" :aria-label="t('customer.inquiries_new.picker_title')">
        <li v-for="product in catalogProducts" :key="product.id">
          <button
            type="button"
            class="picker-item"
            :class="{ 'picker-item--selected': isPickerSelected(product.id) }"
            @click="togglePickerProduct(product)"
          >
            <span class="picker-item__check">
              <Icon v-if="isPickerSelected(product.id)" name="heroicons:check" class="h-4 w-4" aria-hidden="true" />
            </span>
            <span class="picker-item__media">
              <img v-if="product.thumbnail" :src="product.thumbnail" :alt="tField(product, 'name')" class="picker-item__img" />
              <Icon v-else name="heroicons:cube" class="h-6 w-6 text-gray-400" aria-hidden="true" />
            </span>
            <span class="picker-item__body">
              <span class="picker-item__name">{{ tField(product, 'name') }}</span>
              <span class="picker-item__cat">{{ tField(product, 'category') }}</span>
            </span>
          </button>
        </li>
      </ul>

      <template #footer>
        <div class="picker-footer">
          <p class="picker-footer__count">
            {{ t('customer.inquiries_new.picker_selected_count', { count: pickerDraftIds.size }) }}
          </p>
          <div class="picker-footer__actions">
            <button type="button" class="inquiry-btn inquiry-btn--ghost" @click="closePicker">
              {{ t('customer.inquiries_new.cancel') }}
            </button>
            <button type="button" class="inquiry-btn inquiry-btn--primary" @click="confirmPicker">
              {{ t('customer.inquiries_new.picker_add') }}
            </button>
          </div>
        </div>
      </template>
    </Drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { useTranslation } from '~/composables/useTranslation'
import {
  INQUIRY_ACCEPT_ATTR,
  INQUIRY_MAX_FILES,
  INQUIRY_MAX_FILE_MB,
  isInquiryAttachmentAllowed
} from '~/utils/inquiryAttachments'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

/** 已选产品条目 */
interface SelectedProduct {
  id?: string
  slug?: string
  name: string
  thumbnail?: string
  category?: string
}

const { user } = useAuth()
const { t } = useI18n()
const { tField } = useTranslation()
const localePath = useLocalePath()
const route = useRoute()
const { getProducts, submitCustomerInquiry } = useApi()

const form = reactive({
  companyName: user.value?.company || '',
  contactPerson: `${user.value?.firstName || ''} ${user.value?.lastName || ''}`.trim(),
  email: user.value?.email || '',
  estimatedQuantity: '',
  oemNeeded: false,
  expectedDelivery: '',
  message: ''
})

const selectedProducts = ref<SelectedProduct[]>([])
const manualProductInput = ref('')
const loading = ref(false)
const error = ref('')
const success = ref(false)
const fileError = ref('')
const selectedFiles = ref<File[]>([])

const pickerOpen = ref(false)
const pickerSearch = ref('')
const pickerLoading = ref(false)
const catalogProducts = ref<any[]>([])
const pickerDraftIds = ref(new Set<string>())
const pickerDraftMap = ref(new Map<string, SelectedProduct>())

let searchTimer: ReturnType<typeof setTimeout> | null = null

// The backend clamps the page size to 100 (pagination.ParsePagination max), so
// requesting limit: 200 silently truncated the product set. Page through the
// full set at the real page size so enrichment is never silently missing items.
const PRODUCTS_PAGE_SIZE = 100

/** Fetch every catalog product by paging through the backend's real page size. */
const fetchAllProducts = async (): Promise<any[]> => {
  const first = await getProducts({ page: 1, limit: PRODUCTS_PAGE_SIZE })
  const totalPages = first.pagination?.totalPages ?? 1
  if (totalPages <= 1) return first.data || []
  const rest = await Promise.all(
    Array.from({ length: totalPages - 1 }, (_, index) =>
      getProducts({ page: index + 2, limit: PRODUCTS_PAGE_SIZE })
    )
  )
  return first.data.concat(...rest.map((response) => response.data))
}

/** 列表项 key */
const itemKey = (item: SelectedProduct, index: number) => item.id || `manual-${item.name}-${index}`

/** 从 URL 解析预填产品 */
const applyRoutePrefill = () => {
  const next: SelectedProduct[] = []
  const productId = String(route.query.productId || route.query.product || '').trim()
  const name = String(route.query.name || '').trim()
  const productsCsv = String(route.query.products || '').trim()

  if (productId && name) {
    next.push({ id: productId, name })
  } else if (name) {
    next.push({ name })
  }

  if (productsCsv) {
    productsCsv.split(',').map((s) => s.trim()).filter(Boolean).forEach((productName) => {
      if (!next.some((p) => p.name.toLowerCase() === productName.toLowerCase())) {
        next.push({ name: productName })
      }
    })
  }

  if (next.length) {
    selectedProducts.value = next
    enrichSelectedProducts()
  }
}

/** 补全已选产品的 slug / 缩略图等信息 */
const enrichSelectedProducts = async () => {
  const ids = selectedProducts.value.filter((p) => p.id && !p.slug).map((p) => p.id as string)
  if (!ids.length) return
  try {
    const products = await fetchAllProducts()
    const map = new Map(products.map((p: any) => [p.id, p]))
    selectedProducts.value = selectedProducts.value.map((item) => {
      if (!item.id) return item
      const found = map.get(item.id)
      if (!found) return item
      return {
        ...item,
        slug: found.slug,
        thumbnail: found.thumbnail,
        category: tField(found, 'category'),
        name: item.name || tField(found, 'name')
      }
    })
  } catch {
    // 补全失败不影响提交
  }
}

/** 打开产品选择抽屉 */
const openPicker = () => {
  pickerDraftIds.value = new Set(selectedProducts.value.filter((p) => p.id).map((p) => p.id as string))
  pickerDraftMap.value = new Map(
    selectedProducts.value.filter((p) => p.id).map((p) => [p.id as string, p])
  )
  pickerOpen.value = true
  loadCatalogProducts()
}

/** 关闭抽屉 */
const closePicker = () => {
  pickerOpen.value = false
}

/** 加载目录产品（支持搜索） */
const loadCatalogProducts = async () => {
  pickerLoading.value = true
  try {
    const res = await getProducts({
      limit: 50,
      search: pickerSearch.value.trim() || undefined,
      sort: 'name',
      order: 'asc'
    })
    catalogProducts.value = res.data || []
  } catch {
    catalogProducts.value = []
  } finally {
    pickerLoading.value = false
  }
}

watch(pickerSearch, () => {
  if (!pickerOpen.value) return
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(loadCatalogProducts, 300)
})

/** 抽屉内是否已勾选 */
const isPickerSelected = (id: string) => pickerDraftIds.value.has(id)

/** 切换抽屉内勾选 */
const togglePickerProduct = (product: any) => {
  const id = product.id as string
  const nextIds = new Set(pickerDraftIds.value)
  const nextMap = new Map(pickerDraftMap.value)
  if (nextIds.has(id)) {
    nextIds.delete(id)
    nextMap.delete(id)
  } else {
    nextIds.add(id)
    nextMap.set(id, {
      id,
      slug: product.slug,
      name: tField(product, 'name'),
      thumbnail: product.thumbnail,
      category: tField(product, 'category')
    })
  }
  pickerDraftIds.value = nextIds
  pickerDraftMap.value = nextMap
}

/** 确认添加所选产品 */
const confirmPicker = () => {
  const merged = [...selectedProducts.value]
  pickerDraftMap.value.forEach((product) => {
    if (!merged.some((p) => p.id === product.id)) merged.push(product)
  })
  selectedProducts.value = merged
  closePicker()
}

/** 手动添加产品名 */
const addManualProduct = () => {
  const name = manualProductInput.value.trim()
  if (!name) return
  if (selectedProducts.value.some((p) => p.name.toLowerCase() === name.toLowerCase())) {
    manualProductInput.value = ''
    return
  }
  selectedProducts.value.push({ name })
  manualProductInput.value = ''
}

/** 移除已选产品 */
const removeProduct = (index: number) => {
  selectedProducts.value.splice(index, 1)
}

/** 处理选中的附件 */
const handleFilesSelected = (files: FileList) => {
  fileError.value = ''
  for (const file of Array.from(files)) {
    if (selectedFiles.value.length >= INQUIRY_MAX_FILES) {
      fileError.value = t('customer.inquiries_new.max_files')
      break
    }
    if (!isInquiryAttachmentAllowed(file)) {
      fileError.value = t('customer.inquiries_new.file_type_invalid')
      continue
    }
    if (file.size > INQUIRY_MAX_FILE_MB * 1024 * 1024) {
      fileError.value = t('customer.inquiries_new.file_too_large')
      continue
    }
    const dup = selectedFiles.value.some((f) => f.name === file.name && f.size === file.size)
    if (!dup) selectedFiles.value.push(file)
  }
}

/** 移除已选附件 */
const removeFile = (index: number) => {
  selectedFiles.value.splice(index, 1)
  if (selectedFiles.value.length === 0) fileError.value = ''
}

const handleSubmit = async () => {
  if (!selectedProducts.value.length) {
    error.value = t('customer.inquiries_new.no_products_selected')
    return
  }

  loading.value = true
  error.value = ''
  success.value = false
  try {
    const interestedProducts = selectedProducts.value.map((p) => p.name.trim()).filter(Boolean)
    const productIds = selectedProducts.value.map((p) => p.id).filter(Boolean) as string[]
    await submitCustomerInquiry({
      companyName: form.companyName,
      contactPerson: form.contactPerson,
      email: form.email,
      estimatedQuantity: form.estimatedQuantity,
      oemNeeded: form.oemNeeded,
      expectedDelivery: form.expectedDelivery,
      message: form.message,
      interestedProducts,
      productIds: productIds.length ? productIds : undefined,
      files: selectedFiles.value.length ? selectedFiles.value : undefined
    })
    success.value = true
    selectedFiles.value = []
    setTimeout(() => navigateTo(localePath('/customer/inquiries')), 2000)
  } catch (err: any) {
    error.value = err?.message || t('customer.inquiries_new.error')
  } finally {
    loading.value = false
  }
}

onMounted(applyRoutePrefill)
watch(() => route.query, applyRoutePrefill, { deep: true })
</script>

<style scoped>
.inquiry-new {
  width: 100%;
  max-width: 80rem;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1.75rem;
}

.inquiry-new__header { display: flex; flex-direction: column; gap: 1.25rem; }

.inquiry-new__back {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-highlight);
  text-decoration: none;
  width: fit-content;
}

.inquiry-new__header-main {
  padding: 1.75rem 2rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg, 0.75rem);
}

.inquiry-new__title {
  margin: 0;
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.2;
}

.inquiry-new__subtitle {
  margin: 0.5rem 0 0;
  font-size: 1.0625rem;
  color: var(--color-text-lighter);
  line-height: 1.5;
}

.inquiry-form { display: flex; flex-direction: column; gap: 1.5rem; }

.inquiry-section {
  padding: 1.75rem 2rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg, 0.75rem);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.inquiry-section__head {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.25rem;
  padding-bottom: 1.25rem;
  border-bottom: 1px solid var(--color-border-light);
}

.inquiry-section__icon {
  flex-shrink: 0;
  width: 3rem;
  height: 3rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md, 0.5rem);
  background: rgba(var(--color-highlight-rgb), 0.1);
  color: var(--color-highlight);
}

.inquiry-section__icon :deep(svg) {
  width: 1.375rem;
  height: 1.375rem;
}

.inquiry-section__title {
  margin: 0;
  font-size: 1.1875rem;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.35;
}

.inquiry-section__required { color: var(--color-error); margin-left: 0.125rem; }

.inquiry-section__hint {
  margin: 0.375rem 0 0;
  font-size: 0.9375rem;
  color: var(--color-text-lighter);
  line-height: 1.5;
}

.inquiry-products__toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.875rem;
  margin-bottom: 1.25rem;
}

.inquiry-products__list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
}

.inquiry-product-chip {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.125rem;
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-md, 0.5rem);
  background: var(--color-bg-alt);
}

.inquiry-product-chip__media {
  width: 4rem;
  height: 4rem;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm, 0.25rem);
  background: var(--color-bg);
  overflow: hidden;
}

.inquiry-product-chip__img { width: 100%; height: 100%; object-fit: cover; }

.inquiry-product-chip__body { flex: 1; min-width: 0; }

.inquiry-product-chip__name {
  margin: 0;
  font-size: 1.0625rem;
  font-weight: 600;
  color: var(--color-text);
  line-height: 1.35;
}

.inquiry-product-chip__meta {
  margin: 0.25rem 0 0;
  font-size: 0.9375rem;
  color: var(--color-text-lighter);
}

.inquiry-product-chip__actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-shrink: 0;
}

.inquiry-product-chip__link {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-highlight);
  text-decoration: none;
}

.inquiry-product-chip__link:hover { text-decoration: underline; }

.inquiry-product-chip__remove {
  padding: 0.375rem;
  border: none;
  background: transparent;
  color: var(--color-text-lighter);
  cursor: pointer;
  border-radius: var(--radius-sm, 0.25rem);
}

.inquiry-product-chip__remove:hover { color: var(--color-error); }

.inquiry-products__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 2.5rem 1.5rem;
  margin-bottom: 1.25rem;
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-md, 0.5rem);
  font-size: 1rem;
  color: var(--color-text-lighter);
  text-align: center;
}

.inquiry-products__empty :deep(svg) {
  width: 2.5rem;
  height: 2.5rem;
}

.inquiry-manual__label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text);
}

.inquiry-manual__row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.inquiry-manual__row .inquiry-input { flex: 1; min-width: 16rem; }

.inquiry-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.25rem;
}

@media (min-width: 768px) {
  .inquiry-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}

.inquiry-field--full { grid-column: 1 / -1; }

.inquiry-field__label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text);
}

.inquiry-input {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md, 0.5rem);
  font-size: 1rem;
  line-height: 1.5;
  color: var(--color-text);
  background: var(--color-bg-alt);
}

.inquiry-input:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(var(--color-highlight-rgb), 0.12);
  background: var(--color-bg);
}

.inquiry-input--area { resize: vertical; min-height: 8rem; }

.inquiry-radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: 1.25rem;
  margin-top: 0.5rem;
}

.inquiry-radio {
  display: inline-flex;
  align-items: center;
  gap: 0.625rem;
  font-size: 1rem;
  color: var(--color-text);
  cursor: pointer;
}

.inquiry-radio input {
  width: 1.125rem;
  height: 1.125rem;
}

.inquiry-files {
  margin: 1rem 0 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.inquiry-files__item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.75rem 1rem;
  border-radius: var(--radius-md, 0.5rem);
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border-light);
  font-size: 0.9375rem;
}

.inquiry-files__name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.inquiry-files__remove {
  border: none;
  background: transparent;
  color: var(--color-text-lighter);
  cursor: pointer;
  padding: 0.25rem;
}

.inquiry-alert {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  padding: 1rem 1.25rem;
  border-radius: var(--radius-md, 0.5rem);
  font-size: 1rem;
  line-height: 1.5;
}

.inquiry-alert--error {
  background: rgba(var(--color-error-rgb), 0.08);
  color: var(--color-error);
  border: 1px solid rgba(var(--color-error-rgb), 0.2);
}

.inquiry-alert--success {
  background: rgba(var(--color-success-rgb), 0.08);
  color: #065f46;
  border: 1px solid rgba(var(--color-success-rgb), 0.25);
}

.inquiry-form__footer {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 1rem;
  padding: 1.75rem 2rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg, 0.75rem);
}

.inquiry-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  min-height: 2.75rem;
  padding: 0.75rem 1.5rem;
  border-radius: var(--radius-md, 0.5rem);
  font-size: 1rem;
  font-weight: 600;
  text-decoration: none;
  border: none;
  cursor: pointer;
  transition: background 0.15s, filter 0.15s;
}

.inquiry-btn--primary { background: var(--color-highlight); color: white; }
.inquiry-btn--primary:hover:not(:disabled) { filter: brightness(1.05); }
.inquiry-btn--primary:disabled { opacity: 0.55; cursor: not-allowed; }

.inquiry-btn--secondary {
  background: var(--color-bg-alt);
  color: var(--color-text);
  border: 1px solid var(--color-border);
}

.inquiry-btn--ghost {
  background: var(--color-bg);
  color: var(--color-text);
  border: 1px solid var(--color-border);
}

.inquiry-btn--ghost:hover { background: var(--color-bg-alt); }

/* 抽屉内产品列表 */
.picker-search {
  position: relative;
  margin-bottom: 1.25rem;
}

.picker-search__icon {
  position: absolute;
  left: 1rem;
  top: 50%;
  transform: translateY(-50%);
  width: 1.25rem;
  height: 1.25rem;
  color: var(--color-text-lighter);
}

.picker-search__input {
  width: 100%;
  padding: 0.875rem 1rem 0.875rem 2.75rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md, 0.5rem);
  font-size: 1rem;
}

.picker-state {
  padding: 3rem 1rem;
  text-align: center;
  font-size: 1rem;
  color: var(--color-text-lighter);
}

.picker-list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.picker-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.125rem;
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-md, 0.5rem);
  background: var(--color-bg);
  text-align: left;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}

.picker-item:hover { border-color: var(--color-highlight); }

.picker-item--selected {
  border-color: var(--color-highlight);
  background: rgba(var(--color-highlight-rgb), 0.06);
}

.picker-item__check {
  width: 1.5rem;
  height: 1.5rem;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 2px solid var(--color-border);
  border-radius: var(--radius-sm, 0.25rem);
  color: white;
}

.picker-item--selected .picker-item__check {
  background: var(--color-highlight);
  border-color: var(--color-highlight);
}

.picker-item__media {
  width: 3.5rem;
  height: 3.5rem;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm, 0.25rem);
  background: var(--color-bg-alt);
  overflow: hidden;
}

.picker-item__img { width: 100%; height: 100%; object-fit: cover; }

.picker-item__body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.picker-item__name {
  font-size: 1.0625rem;
  font-weight: 600;
  color: var(--color-text);
  line-height: 1.35;
}

.picker-item__cat {
  font-size: 0.9375rem;
  color: var(--color-text-lighter);
}

.picker-footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.picker-footer__count {
  margin: 0;
  font-size: 0.9375rem;
  color: var(--color-text-lighter);
}

.picker-footer__actions {
  display: flex;
  gap: 0.75rem;
  margin-left: auto;
}
</style>
