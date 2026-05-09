<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.products.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.products.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0">
        <button type="button" class="inline-flex items-center justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700" @click="openCreateModal">
          {{ t('admin.products.add_product') }}
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
            <th class="relative py-3.5 pl-3 pr-4 sm:pr-6"><span class="sr-only">Actions</span></th>
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
                {{ product.status || 'active' }}
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
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="closeModal"></div>
        <span class="hidden sm:inline-block sm:h-screen sm:align-middle" aria-hidden="true">&#8203;</span>
        <div class="inline-block w-full transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left align-bottom shadow-xl transition-all sm:my-8 sm:max-w-3xl sm:p-6 sm:align-middle">
          <h3 class="text-lg font-medium leading-6 text-gray-900">{{ editingId ? t('admin.products.edit_product') : t('admin.products.create_product') }}</h3>

          <form class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2" @submit.prevent="saveProduct">
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.name') }}</label>
              <input v-model="form.name" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.slug') }}</label>
              <input v-model="form.slug" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.category') }}</label>
              <input v-model="form.category" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.category_slug') }}</label>
              <input v-model="form.categorySlug" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.moq') }}</label>
              <input v-model.number="form.moq" type="number" min="0" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.lead_time') }}</label>
              <input v-model="form.leadTime" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.stock_quantity') }}</label>
              <input v-model.number="form.stockQuantity" type="number" min="0" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.status') }}</label>
              <select v-model="form.status" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option value="active">active</option>
                <option value="draft">draft</option>
                <option value="inactive">inactive</option>
              </select>
            </div>
            <div class="flex items-center gap-4 pt-7">
              <label class="inline-flex items-center">
                <input v-model="form.oemAvailable" type="checkbox" class="rounded border-gray-300" />
                <span class="ml-2 text-sm text-gray-700">{{ t('admin.products.oem') }}</span>
              </label>
              <label class="inline-flex items-center">
                <input v-model="form.halalCertified" type="checkbox" class="rounded border-gray-300" />
                <span class="ml-2 text-sm text-gray-700">{{ t('admin.products.halal') }}</span>
              </label>
              <label class="inline-flex items-center">
                <input v-model="form.featured" type="checkbox" class="rounded border-gray-300" />
                <span class="ml-2 text-sm text-gray-700">{{ t('admin.products.featured') }}</span>
              </label>
            </div>
            <div class="sm:col-span-2">
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.summary') }}</label>
              <textarea v-model="form.summary" rows="2" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div class="sm:col-span-2">
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.description_label') }}</label>
              <textarea v-model="form.description" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div class="sm:col-span-2">
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.thumbnail') }}</label>
              <div class="mt-1 flex items-center gap-2">
                <input v-model="form.thumbnail" class="block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
                <button type="button" class="rounded-md bg-orange-50 px-3 py-2 text-xs font-medium text-orange-600 hover:bg-orange-500" @click="triggerUpload('thumbnail')">
                  {{ t('admin.products.upload_image') }}
                </button>
              </div>
            </div>
            <div class="sm:col-span-2">
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.images') }}</label>
              <div class="mt-1 flex items-center gap-2">
                <input v-model="form.imagesInput" class="block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
                <button type="button" class="rounded-md bg-orange-50 px-3 py-2 text-xs font-medium text-orange-600 hover:text-orange-500" @click="triggerUpload('images')">
                  {{ t('admin.products.upload_images') }}
                </button>
                <div v-if="uploadingImage" class="mt-1 text-xs text-gray-500">
                  <button type="button" class="rounded-md bg-red-50 px-3 py-2 text-xs font-medium text-red-600 hover:text-red-500" @click="cancelImageUpload">
                    {{ t('admin.products.cancel') }}
                  </button>
                </div>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.flavors') }}</label>
              <input v-model="form.flavorsInput" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.shapes') }}</label>
              <input v-model="form.shapesInput" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div class="sm:col-span-2">
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.certifications') }}</label>
              <input v-model="form.certificationsInput" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.ingredients') }}</label>
              <textarea v-model="form.ingredients" rows="2" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.allergens') }}</label>
              <textarea v-model="form.allergens" rows="2" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.shelf_life') }}</label>
              <input v-model="form.shelfLife" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.products.storage') }}</label>
              <input v-model="form.storage" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>

            <div v-if="formError" class="sm:col-span-2 text-sm text-red-600">{{ formError }}</div>
            <div class="sm:col-span-2 mt-2 flex justify-end gap-3">
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

const products = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20

const showModal = ref(false)
const editingId = ref('')
const saving = ref(false)
const formError = ref('')
const actionMessage = ref('')
const actionError = ref(false)
const uploadingImage = ref(false)
const uploadError = ref('')
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

const form = reactive({
  name: '',
  slug: '',
  summary: '',
  description: '',
  category: '',
  categorySlug: '',
  thumbnail: '',
  moq: 0,
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
  storage: ''
})

const parseCSV = (value: string) => value.split(',').map(v => v.trim()).filter(Boolean)

const resetForm = () => {
  form.name = ''
  form.slug = ''
  form.summary = ''
  form.description = ''
  form.category = ''
  form.categorySlug = ''
  form.thumbnail = ''
  form.moq = 0
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

const openCreateModal = () => { editingId.value = ''; resetForm(); formError.value = ''; actionMessage.value = ''; actionError.value = false; showModal.value = true }

const openEditModal = async (product: any) => {
  editingId.value = product.id; formError.value = ''; actionMessage.value = ''; actionError.value = false
  try { const detail = await api.get<any>(`/admin/products/${product.id}`); fillFormFromProduct(detail); showModal.value = true }
  catch { fillFormFromProduct(product); showModal.value = true }
}

const closeModal = () => { showModal.value = false; saving.value = false; formError.value = '' }

const buildPayload = () => ({
  name: form.name, slug: form.slug, summary: form.summary, description: form.description,
  category: form.category, categorySlug: form.categorySlug, thumbnail: form.thumbnail,
  moq: form.moq, stockQuantity: Math.max(0, Number(form.stockQuantity) || 0), leadTime: form.leadTime,
  oemAvailable: form.oemAvailable, halalCertified: form.halalCertified, featured: form.featured, status: form.status,
  images: parseCSV(form.imagesInput), flavors: parseCSV(form.flavorsInput), shapes: parseCSV(form.shapesInput),
  certifications: parseCSV(form.certificationsInput), ingredients: form.ingredients, allergens: form.allergens,
  shelfLife: form.shelfLife, storage: form.storage
})

const saveProduct = async () => {
  if (!form.name.trim()) { formError.value = t('admin.products.name_required'); return }
  saving.value = true; formError.value = ''; actionMessage.value = ''; actionError.value = false
  const payload = buildPayload()
  try {
    if (editingId.value) { await api.put(`/admin/products/${editingId.value}`, payload); actionMessage.value = t('admin.products.updated_success') }
    else { await api.post('/admin/products', payload); actionMessage.value = t('admin.products.created_success') }
    closeModal(); await fetchProducts()
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
  color: #dc2626;
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
  transition: all var(--transition-fast);
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
  transition: all var(--transition-fast);
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
