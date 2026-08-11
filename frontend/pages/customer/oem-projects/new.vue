<template>
  <div class="oem-new">
    <!-- 页头 -->
    <header class="oem-new__header">
      <NuxtLink :to="localePath('/customer/oem-projects')" class="oem-new__back">
        <Icon name="heroicons:arrow-left" class="h-4 w-4" aria-hidden="true" />
        {{ t('customer.oemProjects.back') }}
      </NuxtLink>
      <div class="oem-new__header-main">
        <div class="oem-new__header-text">
          <h1 class="oem-new__title">{{ t('customer.oemProjects.new_project') }}</h1>
          <p class="oem-new__subtitle">{{ t('customer.oemProjects.new_project_subtitle') }}</p>
        </div>
        <span v-if="prefilledFromCatalog" class="oem-new__prefill-badge">
          <Icon name="heroicons:sparkles" class="h-4 w-4" aria-hidden="true" />
          {{ t('customer.oemProjects.prefilled_from_catalog') }}
        </span>
      </div>
    </header>

    <div class="oem-new__layout">
      <!-- 侧栏：产品预览 + 流程 -->
      <aside class="oem-new__aside">
        <article class="oem-preview">
          <div class="oem-preview__media">
            <img
              v-if="selectedProduct?.thumbnail"
              :src="selectedProduct.thumbnail"
              :alt="previewName"
              class="oem-preview__img"
            />
            <div v-else class="oem-preview__placeholder">
              <Icon name="heroicons:cube" class="h-10 w-10" aria-hidden="true" />
            </div>
          </div>
          <div class="oem-preview__body">
            <p class="oem-preview__label">{{ t('customer.oemProjects.select_product') }}</p>
            <h2 class="oem-preview__name">{{ previewName || t('customer.oemProjects.select_product_placeholder') }}</h2>
            <p v-if="selectedProduct?.category" class="oem-preview__category">{{ tField(selectedProduct, 'category') }}</p>
            <dl v-if="hasPreviewMeta" class="oem-preview__meta">
              <div v-if="form.moq" class="oem-preview__meta-row">
                <dt>{{ t('customer.oemProjects.moq') }}</dt>
                <dd>{{ formatNumber(form.moq) }}</dd>
              </div>
              <div v-if="form.flavor" class="oem-preview__meta-row">
                <dt>{{ t('customer.oemProjects.flavor') }}</dt>
                <dd>{{ form.flavor }}</dd>
              </div>
              <div v-if="form.shape" class="oem-preview__meta-row">
                <dt>{{ t('customer.oemProjects.shape') }}</dt>
                <dd>{{ form.shape }}</dd>
              </div>
            </dl>
            <div v-if="certificationTags.length" class="oem-preview__tags">
              <span v-for="tag in certificationTags" :key="tag" class="oem-preview__tag">{{ tag }}</span>
            </div>
          </div>
        </article>

        <section class="oem-steps" aria-label="OEM workflow">
          <h3 class="oem-steps__title">{{ t('customer.oemProjects.progress') }}</h3>
          <ol class="oem-steps__list">
            <li v-for="(step, idx) in workflowSteps" :key="step.key" class="oem-steps__item">
              <span class="oem-steps__dot" :class="{ 'oem-steps__dot--active': idx === 0 }">{{ idx + 1 }}</span>
              <span class="oem-steps__label">{{ step.label }}</span>
            </li>
          </ol>
        </section>
      </aside>

      <!-- 主表单 -->
      <div class="oem-new__main">
        <form class="oem-form" @submit.prevent="submitProject">
          <!-- 基础产品 -->
          <section class="oem-section">
            <header class="oem-section__head">
              <span class="oem-section__icon"><Icon name="heroicons:shopping-bag" class="h-5 w-5" /></span>
              <div>
                <h3 class="oem-section__title">{{ t('customer.oemProjects.section_product') }}</h3>
                <p class="oem-section__hint">{{ t('customer.oemProjects.section_product_hint') }}</p>
              </div>
            </header>
            <div class="oem-section__grid oem-section__grid--1">
              <div class="oem-field">
                <label for="oem-product-id" class="oem-field__label">{{ t('customer.oemProjects.select_product') }}</label>
                <select id="oem-product-id" v-model="form.productId" class="oem-input oem-input--select">
                  <option value="">{{ t('customer.oemProjects.select_product_placeholder') }}</option>
                  <option v-for="p in availableProducts" :key="p.id" :value="p.id">
                    {{ tField(p, 'name') }} ({{ tField(p, 'category') }})
                  </option>
                </select>
              </div>
              <div class="oem-field">
                <label for="oem-product-name" class="oem-field__label">
                  {{ t('customer.oemProjects.product_name') }}
                  <span class="oem-field__required">*</span>
                </label>
                <input
                  id="oem-product-name"
                  v-model="form.productName"
                  class="oem-input"
                  :placeholder="t('customer.oemProjects.product_name_placeholder')"
                />
              </div>
            </div>
          </section>

          <!-- 定制规格 -->
          <section class="oem-section">
            <header class="oem-section__head">
              <span class="oem-section__icon"><Icon name="heroicons:beaker" class="h-5 w-5" /></span>
              <div>
                <h3 class="oem-section__title">{{ t('customer.oemProjects.section_specs') }}</h3>
                <p class="oem-section__hint">{{ t('customer.oemProjects.section_specs_hint') }}</p>
              </div>
            </header>
            <div class="oem-section__grid">
              <div class="oem-field">
                <label for="oem-flavor" class="oem-field__label">{{ t('customer.oemProjects.flavor') }}</label>
                <input id="oem-flavor" v-model="form.flavor" class="oem-input" :placeholder="t('customer.oemProjects.flavor_placeholder')" />
              </div>
              <div class="oem-field">
                <label for="oem-shape" class="oem-field__label">{{ t('customer.oemProjects.shape') }}</label>
                <input id="oem-shape" v-model="form.shape" class="oem-input" :placeholder="t('customer.oemProjects.shape_placeholder')" />
              </div>
              <div class="oem-field oem-field--full">
                <label for="oem-packaging" class="oem-field__label">{{ t('customer.oemProjects.packaging') }}</label>
                <input id="oem-packaging" v-model="form.packaging" class="oem-input" :placeholder="t('customer.oemProjects.packaging_placeholder')" />
              </div>
            </div>
          </section>

          <!-- 市场与认证 -->
          <section class="oem-section">
            <header class="oem-section__head">
              <span class="oem-section__icon"><Icon name="heroicons:globe-alt" class="h-5 w-5" /></span>
              <div>
                <h3 class="oem-section__title">{{ t('customer.oemProjects.section_market') }}</h3>
                <p class="oem-section__hint">{{ t('customer.oemProjects.section_market_hint') }}</p>
              </div>
            </header>
            <div class="oem-section__grid">
              <div class="oem-field">
                <label for="oem-market" class="oem-field__label">{{ t('customer.oemProjects.target_market') }}</label>
                <input id="oem-market" v-model="form.targetMarket" class="oem-input" :placeholder="t('customer.oemProjects.target_market_placeholder')" />
              </div>
              <div class="oem-field">
                <label for="oem-moq" class="oem-field__label">{{ t('customer.oemProjects.moq') }}</label>
                <input id="oem-moq" v-model.number="form.moq" type="number" min="1" class="oem-input" :placeholder="t('customer.oemProjects.moq_placeholder')" />
              </div>
              <div class="oem-field oem-field--full">
                <label for="oem-certs" class="oem-field__label">{{ t('customer.oemProjects.certifications') }}</label>
                <input id="oem-certs" v-model="form.certificationsInput" class="oem-input" :placeholder="t('customer.oemProjects.certifications_placeholder')" />
              </div>
            </div>
          </section>

          <!-- 需求与附件 -->
          <section class="oem-section">
            <header class="oem-section__head">
              <span class="oem-section__icon"><Icon name="heroicons:document-text" class="h-5 w-5" /></span>
              <div>
                <h3 class="oem-section__title">{{ t('customer.oemProjects.section_details') }}</h3>
                <p class="oem-section__hint">{{ t('customer.oemProjects.section_details_hint') }}</p>
              </div>
            </header>
            <div class="oem-section__grid oem-section__grid--1">
              <div class="oem-field">
                <label for="oem-requirements" class="oem-field__label">{{ t('customer.oemProjects.requirements') }}</label>
                <textarea
                  id="oem-requirements"
                  v-model="form.requirements"
                  rows="4"
                  class="oem-input oem-input--area"
                  :placeholder="t('customer.oemProjects.requirements_placeholder')"
                />
              </div>
              <div class="oem-field">
                <label class="oem-field__label">{{ t('customer.oemProjects.attachments') }}</label>
                <p class="oem-field__hint">{{ t('customer.oemProjects.attachments_hint') }}</p>
                <InputFile
                  id="oem-project-attachments"
                  :accept="OEM_ACCEPT_ATTR"
                  :max-files="OEM_MAX_FILES"
                  :max-size="OEM_MAX_FILE_MB"
                  :error="fileError"
                  @files-selected="handleFilesSelected"
                  @error="fileError = $event"
                />
                <ul v-if="selectedFiles.length" class="oem-files">
                  <li v-for="(file, index) in selectedFiles" :key="`${file.name}-${index}`" class="oem-files__item">
                    <Icon name="heroicons:paper-clip" class="h-4 w-4 shrink-0 text-gray-400" aria-hidden="true" />
                    <span class="oem-files__name">{{ file.name }}</span>
                    <button
                      type="button"
                      class="oem-files__remove"
                      :aria-label="t('customer.oemProjects.remove_file')"
                      @click="removeFile(index)"
                    >
                      <Icon name="heroicons:x-mark" class="h-4 w-4" />
                    </button>
                  </li>
                </ul>
              </div>
            </div>
          </section>

          <!-- 反馈与操作 -->
          <div v-if="errorMessage" class="oem-alert oem-alert--error" role="alert">
            <Icon name="heroicons:exclamation-circle" class="h-5 w-5 shrink-0" aria-hidden="true" />
            {{ errorMessage }}
          </div>
          <div v-if="successMessage" class="oem-alert oem-alert--success" role="status">
            <Icon name="heroicons:check-circle" class="h-5 w-5 shrink-0" aria-hidden="true" />
            {{ successMessage }}
          </div>

          <footer class="oem-form__footer">
            <NuxtLink :to="localePath('/customer/oem-projects')" class="oem-btn oem-btn--ghost">
              {{ t('customer.oemProjects.cancel') }}
            </NuxtLink>
            <button type="submit" class="oem-btn oem-btn--primary" :disabled="submitting">
              <Icon v-if="submitting" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" aria-hidden="true" />
              {{ submitting ? t('customer.oemProjects.submitting') : t('customer.oemProjects.submit') }}
            </button>
          </footer>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useTranslation } from '~/composables/useTranslation'
import {
  OEM_ACCEPT_ATTR,
  OEM_MAX_FILES,
  OEM_MAX_FILE_MB,
  isOemAttachmentAllowed
} from '~/utils/customerAttachments'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const { t } = useI18n()
const { tField } = useTranslation()
const localePath = useLocalePath()
const route = useRoute()
const { getProducts, getProduct, customerCreateOemProject } = useApi()

const submitting = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const fileError = ref('')
const availableProducts = ref<any[]>([])
const selectedFiles = ref<File[]>([])

// The backend clamps the page size to 100 (pagination.ParsePagination max), so
// requesting limit: 200 silently truncated the product picker. Page through the
// full set at the real page size so the picker shows the complete catalog.
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

const form = reactive({
  productId: '', productName: '', flavor: '', shape: '', packaging: '',
  targetMarket: '', certificationsInput: '', moq: 0, requirements: ''
})

/** OEM 流程步骤（侧栏展示） */
const workflowSteps = computed(() => [
  { key: 'inquiry', label: t('customer.oemProjects.inquiry') },
  { key: 'sampling', label: t('customer.oemProjects.sampling') },
  { key: 'formulation', label: t('customer.oemProjects.formulation') },
  { key: 'quotation', label: t('customer.oemProjects.quotation') },
  { key: 'contract', label: t('customer.oemProjects.contract') },
  { key: 'production', label: t('customer.oemProjects.production') },
  { key: 'delivery', label: t('customer.oemProjects.delivery') },
])

const prefilledFromCatalog = computed(() => Boolean(String(route.query.productId || '').trim()))

const selectedProduct = computed(() =>
  availableProducts.value.find((p) => p.id === form.productId) || null
)

const previewName = computed(() =>
  form.productName.trim() || (selectedProduct.value ? tField(selectedProduct.value, 'name') : '')
)

const certificationTags = computed(() => parseCSV(form.certificationsInput))

const hasPreviewMeta = computed(() =>
  Boolean(form.moq || form.flavor.trim() || form.shape.trim())
)

/** 数字格式化（MOQ 展示） */
const formatNumber = (n: number) => new Intl.NumberFormat().format(n)

/** 将产品数据预填到 OEM 表单 */
const applyProductToForm = (p: any) => {
  if (!p?.id) return
  form.productId = p.id
  if (!form.productName.trim()) form.productName = tField(p, 'name')
  if (!form.moq && p.moq) form.moq = p.moq
  if (!form.flavor.trim()) {
    const flavors = Array.isArray(p.flavors) ? p.flavors : []
    if (flavors[0]) form.flavor = typeof flavors[0] === 'string' ? flavors[0] : ''
  }
  if (!form.shape.trim()) {
    const shapes = Array.isArray(p.shapes) ? p.shapes : []
    if (shapes[0]) form.shape = typeof shapes[0] === 'string' ? shapes[0] : ''
  }
  if (!form.certificationsInput.trim()) {
    const certs = Array.isArray(p.certifications) ? p.certifications : []
    if (certs.length) form.certificationsInput = certs.join(', ')
  }
}

/** 从 URL query 预填（产品详情页跳转） */
const applyQueryToForm = async () => {
  const productId = String(route.query.productId || '').trim()
  const productName = String(route.query.productName || '').trim()
  if (productName) form.productName = productName
  if (route.query.moq) {
    const moq = Number(route.query.moq)
    if (Number.isFinite(moq) && moq > 0) form.moq = moq
  }
  if (route.query.flavor) form.flavor = String(route.query.flavor)
  if (route.query.shape) form.shape = String(route.query.shape)
  if (route.query.certifications) form.certificationsInput = String(route.query.certifications).replace(/,/g, ', ')

  if (!productId) return

  form.productId = productId
  let matched = availableProducts.value.find((p) => p.id === productId)
  if (!matched) {
    try {
      matched = await getProduct(productId)
      if (matched) availableProducts.value.unshift(matched)
    } catch {
      // 单产品加载失败时仍保留 query 中的 productId / productName
    }
  }
  if (matched) applyProductToForm(matched)
}

watch(() => form.productId, (id) => {
  if (!id) return
  const p = availableProducts.value.find((item) => item.id === id)
  if (p) {
    form.productName = tField(p, 'name')
    if (!form.moq && p.moq) form.moq = p.moq
  }
})

onMounted(async () => {
  try {
    availableProducts.value = await fetchAllProducts()
  } catch {
    // 产品列表加载失败不阻塞表单提交
  }
  await applyQueryToForm()
})

const parseCSV = (value: string) => value.split(',').map(v => v.trim()).filter(Boolean)

/** 处理选中的附件 */
const handleFilesSelected = (files: FileList) => {
  fileError.value = ''
  for (const file of Array.from(files)) {
    if (selectedFiles.value.length >= OEM_MAX_FILES) {
      fileError.value = t('customer.oemProjects.max_files')
      break
    }
    if (!isOemAttachmentAllowed(file)) {
      fileError.value = t('customer.oemProjects.file_type_invalid')
      continue
    }
    if (file.size > OEM_MAX_FILE_MB * 1024 * 1024) {
      fileError.value = t('customer.oemProjects.file_too_large')
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

const submitProject = async () => {
  submitting.value = true
  errorMessage.value = ''
  successMessage.value = ''
  try {
    const selected = availableProducts.value.find(p => p.id === form.productId)
    const productName = form.productName.trim() || selected?.name || ''
    if (!productName) {
      errorMessage.value = t('customer.oemProjects.product_name_required')
      submitting.value = false
      return
    }
    await customerCreateOemProject({
      productName,
      notes: form.requirements,
      flavor: form.flavor,
      shape: form.shape,
      packaging: form.packaging,
      targetMarket: form.targetMarket,
      certifications: parseCSV(form.certificationsInput),
      moq: form.moq || 0,
      files: selectedFiles.value.length ? selectedFiles.value : undefined
    })
    successMessage.value = t('customer.oemProjects.project_created')
    selectedFiles.value = []
    setTimeout(() => navigateTo(localePath('/customer/oem-projects')), 1500)
  } catch (err: any) {
    errorMessage.value = err?.message || t('errors.api.save_failed')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.oem-new {
  max-width: 72rem;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.oem-new__header {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.oem-new__back {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-highlight);
  text-decoration: none;
  width: fit-content;
}

.oem-new__back:hover { color: var(--color-highlight-hover); }

.oem-new__header-main {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg, 0.75rem);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.oem-new__title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text);
  line-height: 1.25;
}

.oem-new__subtitle {
  margin: 0.375rem 0 0;
  font-size: 0.875rem;
  color: var(--color-text-lighter);
  max-width: 36rem;
}

.oem-new__prefill-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  border-radius: var(--radius-full, 9999px);
  font-size: 0.75rem;
  font-weight: 600;
  background: rgba(var(--color-highlight-rgb), 0.12);
  color: var(--color-accent);
  white-space: nowrap;
}

.oem-new__layout {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.25rem;
  align-items: start;
}

@media (min-width: 1024px) {
  .oem-new__layout {
    grid-template-columns: minmax(17rem, 20rem) 1fr;
    gap: 1.5rem;
  }
}

.oem-new__aside {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

@media (min-width: 1024px) {
  .oem-new__aside {
    position: sticky;
    top: 1.5rem;
  }
}

/* 产品预览卡片 */
.oem-preview {
  overflow: hidden;
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg, 0.75rem);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.oem-preview__media {
  aspect-ratio: 16 / 10;
  background: var(--color-bg-alt);
  border-bottom: 1px solid var(--color-border-light);
}

.oem-preview__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.oem-preview__placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-lighter);
  opacity: 0.5;
}

.oem-preview__body {
  padding: 1rem 1.125rem 1.125rem;
}

.oem-preview__label {
  margin: 0;
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-highlight);
}

.oem-preview__name {
  margin: 0.25rem 0 0;
  font-size: 1.0625rem;
  font-weight: 700;
  line-height: 1.35;
  color: var(--color-text);
}

.oem-preview__category {
  margin: 0.25rem 0 0;
  font-size: 0.8125rem;
  color: var(--color-text-lighter);
}

.oem-preview__meta {
  margin: 0.875rem 0 0;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.oem-preview__meta-row {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  font-size: 0.8125rem;
}

.oem-preview__meta-row dt { color: var(--color-text-lighter); }
.oem-preview__meta-row dd { margin: 0; font-weight: 600; color: var(--color-text); text-align: right; }

.oem-preview__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
  margin-top: 0.75rem;
}

.oem-preview__tag {
  padding: 0.125rem 0.5rem;
  border-radius: var(--radius-full, 9999px);
  font-size: 0.6875rem;
  font-weight: 600;
  background: rgba(var(--color-success-rgb), 0.12);
  color: #065f46;
}

/* 流程步骤 */
.oem-steps {
  padding: 1rem 1.125rem;
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg, 0.75rem);
}

.oem-steps__title {
  margin: 0 0 0.75rem;
  font-size: 0.8125rem;
  font-weight: 700;
  color: var(--color-text);
}

.oem-steps__list {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.oem-steps__item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  font-size: 0.75rem;
  color: var(--color-text-lighter);
}

.oem-steps__dot {
  flex-shrink: 0;
  width: 1.375rem;
  height: 1.375rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-size: 0.6875rem;
  font-weight: 700;
  background: var(--color-bg-dark);
  color: var(--color-text-lighter);
}

.oem-steps__dot--active {
  background: var(--color-highlight);
  color: white;
}

.oem-steps__label { line-height: 1.3; }

/* 主表单 */
.oem-new__main {
  min-width: 0;
}

.oem-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.oem-section {
  padding: 1.25rem 1.5rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg, 0.75rem);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.oem-section__head {
  display: flex;
  align-items: flex-start;
  gap: 0.875rem;
  margin-bottom: 1.125rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--color-border-light);
}

.oem-section__icon {
  flex-shrink: 0;
  width: 2.5rem;
  height: 2.5rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md, 0.5rem);
  background: rgba(var(--color-highlight-rgb), 0.1);
  color: var(--color-highlight);
}

.oem-section__title {
  margin: 0;
  font-size: 0.9375rem;
  font-weight: 700;
  color: var(--color-text);
}

.oem-section__hint {
  margin: 0.25rem 0 0;
  font-size: 0.8125rem;
  color: var(--color-text-lighter);
}

.oem-section__grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
}

@media (min-width: 640px) {
  .oem-section__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .oem-section__grid--1 { grid-template-columns: 1fr; }
}

.oem-field--full { grid-column: 1 / -1; }

.oem-field__label {
  display: block;
  margin-bottom: 0.375rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text);
}

.oem-field__required { color: var(--color-error); margin-left: 0.125rem; }

.oem-field__hint {
  margin: 0 0 0.5rem;
  font-size: 0.75rem;
  color: var(--color-text-lighter);
  line-height: 1.45;
}

.oem-input {
  width: 100%;
  padding: 0.5625rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md, 0.5rem);
  font-size: 0.875rem;
  color: var(--color-text);
  background: var(--color-bg-alt);
  transition: border-color 0.15s, box-shadow 0.15s;
}

.oem-input:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(var(--color-highlight-rgb), 0.12);
  background: var(--color-bg);
}

.oem-input--area {
  resize: vertical;
  min-height: 6rem;
  line-height: 1.5;
}

.oem-input--select {
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%237e7570'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.625rem center;
  background-size: 1rem;
  padding-right: 2.25rem;
}

/* 附件列表 */
.oem-files {
  margin: 0.75rem 0 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.oem-files__item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-md, 0.5rem);
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border-light);
  font-size: 0.8125rem;
}

.oem-files__name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text);
}

.oem-files__remove {
  flex-shrink: 0;
  padding: 0.125rem;
  border: none;
  background: transparent;
  color: var(--color-text-lighter);
  cursor: pointer;
  border-radius: var(--radius-sm, 0.25rem);
}

.oem-files__remove:hover { color: var(--color-error); }

/* 提示与按钮 */
.oem-alert {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  border-radius: var(--radius-md, 0.5rem);
  font-size: 0.875rem;
}

.oem-alert--error {
  background: rgba(var(--color-error-rgb), 0.08);
  color: var(--color-error);
  border: 1px solid rgba(var(--color-error-rgb), 0.2);
}

.oem-alert--success {
  background: rgba(var(--color-success-rgb), 0.08);
  color: #065f46;
  border: 1px solid rgba(var(--color-success-rgb), 0.25);
}

.oem-form__footer {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1.25rem 1.5rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg, 0.75rem);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.oem-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  padding: 0.5625rem 1.25rem;
  border-radius: var(--radius-md, 0.5rem);
  font-size: 0.875rem;
  font-weight: 600;
  text-decoration: none;
  cursor: pointer;
  transition: background 0.15s, filter 0.15s, opacity 0.15s;
}

.oem-btn--ghost {
  background: var(--color-bg);
  color: var(--color-text);
  border: 1px solid var(--color-border);
}

.oem-btn--ghost:hover { background: var(--color-bg-alt); }

.oem-btn--primary {
  border: none;
  background: var(--color-highlight);
  color: white;
}

.oem-btn--primary:hover:not(:disabled) { filter: brightness(1.05); }
.oem-btn--primary:disabled { opacity: 0.55; cursor: not-allowed; }
</style>
