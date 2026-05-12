<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ t('admin.inquiries.title') }}</h1>
        <p class="page-subtitle">{{ t('admin.inquiries.description') }}</p>
      </div>
      <div>
        <button type="button" class="btn btn-highlight" @click="openCreateModal">
          {{ t('admin.inquiries.new_inquiry') }}
        </button>
      </div>
    </div>

    <div class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th>{{ t('admin.inquiries.col_date') }}</th>
            <th>{{ t('admin.inquiries.col_company_contact') }}</th>
            <th>{{ t('admin.inquiries.col_products') }}</th>
            <th>{{ t('admin.inquiries.col_status') }}</th>
            <th><span class="sr-only">Actions</span></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="pending">
            <td colspan="5" class="text-center py-5">{{ t('admin.inquiries.loading') }}</td>
          </tr>
          <tr v-else-if="error">
            <td colspan="5" class="text-center py-5 text-error">{{ error }}</td>
          </tr>
          <tr v-else-if="inquiries.length === 0">
            <td colspan="5" class="text-center py-5">{{ t('admin.inquiries.no_data') }}</td>
          </tr>
          <tr v-else v-for="inquiry in inquiries" :key="inquiry.id">
            <td>
              {{ formatDate(inquiry.createdAt) }}
            </td>
            <td>
              <div class="font-medium">{{ inquiry.companyName }}</div>
              <div class="text-light">{{ inquiry.contactPerson }} ({{ inquiry.email }})</div>
            </td>
            <td>{{ formatProducts(inquiry.interestedProducts) }}</td>
            <td>
              <span class="badge" :class="statusClass(inquiry.status)">
                {{ inquiry.status }}
              </span>
            </td>
            <td class="text-right">
              <NuxtLink :to="localePath(`/admin/inquiries/${inquiry.id}`)" class="link mr-3">{{ t('admin.inquiries.view') }}</NuxtLink>
              <button type="button" class="link mr-3" @click="openEditModal(inquiry.id)">{{ t('admin.inquiries.edit') }}</button>
              <button type="button" class="link link-danger" @click="deleteInquiry(inquiry.id)">{{ t('admin.inquiries.delete') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="pagination" class="pagination">
      <div class="text-sm text-light">
        {{ t('admin.inquiries.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
      </div>
      <div class="flex items-center gap-2">
        <button type="button" class="btn btn-outline btn-sm" :disabled="page <= 1" @click="prevPage">
          {{ t('common.previous') }}
        </button>
        <button type="button" class="btn btn-outline btn-sm" :disabled="page >= pagination.totalPages" @click="nextPage">
          {{ t('common.next') }}
        </button>
      </div>
    </div>

    <div v-if="showModal" class="modal-overlay" role="dialog" aria-modal="true">
      <div class="modal-container">
        <button type="button" class="modal-backdrop w-full border-0 cursor-pointer" @click="closeModal" :aria-label="t('close')"></button>
        <div class="modal-content">
          <h3 class="modal-title">{{ editingId ? t('admin.inquiries.edit_inquiry') : t('admin.inquiries.create_inquiry') }}</h3>

          <form class="form-grid" @submit.prevent="saveInquiry">
            <div class="form-group">
              <label for="inquiry-userId" class="form-label">{{ t('admin.inquiries.user_id') }}</label>
              <input id="inquiry-userId" v-model="form.userId" name="userId" autocomplete="off" class="form-input" />
            </div>
            <div class="form-group">
              <label for="inquiry-companyName" class="form-label">{{ t('admin.inquiries.company_name') }}</label>
              <input id="inquiry-companyName" v-model="form.companyName" name="companyName" autocomplete="organization" required class="form-input" />
            </div>
            <div class="form-group">
              <label for="inquiry-contactPerson" class="form-label">{{ t('admin.inquiries.contact_person') }}</label>
              <input id="inquiry-contactPerson" v-model="form.contactPerson" name="contactPerson" autocomplete="name" required class="form-input" />
            </div>
            <div class="form-group">
              <label for="inquiry-email" class="form-label">{{ t('admin.inquiries.email') }}</label>
              <input id="inquiry-email" v-model="form.email" name="email" type="email" autocomplete="email" required class="form-input" />
            </div>
            <div class="form-group">
              <label for="inquiry-whatsApp" class="form-label">{{ t('admin.inquiries.whatsapp') }}</label>
              <input id="inquiry-whatsApp" v-model="form.whatsApp" name="whatsApp" autocomplete="tel" class="form-input" />
            </div>
            <div class="form-group">
              <label for="inquiry-targetCountry" class="form-label">{{ t('admin.inquiries.target_country') }}</label>
              <input id="inquiry-targetCountry" v-model="form.targetCountry" name="targetCountry" autocomplete="country-name" class="form-input" />
            </div>
            <div class="form-group">
              <label for="inquiry-estimatedQuantity" class="form-label">{{ t('admin.inquiries.estimated_quantity') }}</label>
              <input id="inquiry-estimatedQuantity" v-model="form.estimatedQuantity" name="estimatedQuantity" autocomplete="off" class="form-input" />
            </div>
            <div class="form-group">
              <label for="inquiry-expectedDelivery" class="form-label">{{ t('admin.inquiries.expected_delivery') }}</label>
              <input id="inquiry-expectedDelivery" v-model="form.expectedDelivery" name="expectedDelivery" autocomplete="off" class="form-input" />
            </div>
            <div class="form-group">
              <label for="inquiry-status" class="form-label">{{ t('admin.inquiries.status') }}</label>
              <select id="inquiry-status" v-model="form.status" name="status" class="form-input">
                <option value="pending">{{ enumLabel('inquiry_status', 'pending') }}</option>
                <option value="contacted">{{ enumLabel('inquiry_status', 'contacted') }}</option>
                <option value="quoted">{{ enumLabel('inquiry_status', 'quoted') }}</option>
                <option value="negotiating">{{ enumLabel('inquiry_status', 'negotiating') }}</option>
                <option value="won">{{ enumLabel('inquiry_status', 'won') }}</option>
                <option value="lost">{{ enumLabel('inquiry_status', 'lost') }}</option>
                <option value="closed">{{ enumLabel('inquiry_status', 'closed') }}</option>
              </select>
            </div>
            <div class="form-group">
              <label for="inquiry-priority" class="form-label">{{ t('admin.inquiries.priority') }}</label>
              <select id="inquiry-priority" v-model="form.priority" name="priority" class="form-input">
                <option value="low">{{ enumLabel('inquiry_priority', 'low') }}</option>
                <option value="normal">{{ enumLabel('inquiry_priority', 'normal') }}</option>
                <option value="high">{{ enumLabel('inquiry_priority', 'high') }}</option>
                <option value="urgent">{{ enumLabel('inquiry_priority', 'urgent') }}</option>
              </select>
            </div>
            <div class="form-group col-span-2">
              <label for="inquiry-interestedProductsInput" class="form-label">{{ t('admin.inquiries.interested_products') }}</label>
              <input id="inquiry-interestedProductsInput" v-model="form.interestedProductsInput" name="interestedProductsInput" autocomplete="off" class="form-input" />
            </div>
            <div class="form-group col-span-2">
              <label class="inline-flex items-center gap-2">
                <input id="inquiry-oemNeeded" v-model="form.oemNeeded" name="oemNeeded" type="checkbox" class="form-checkbox" />
                <span class="text-sm">{{ t('admin.inquiries.oem_needed') }}</span>
              </label>
            </div>
            <div class="form-group col-span-2">
              <label for="inquiry-message" class="form-label">{{ t('admin.inquiries.message') }}</label>
              <textarea id="inquiry-message" v-model="form.message" name="message" rows="3" class="form-textarea"></textarea>
            </div>
            <div class="form-group col-span-2">
              <label for="inquiry-customerNotes" class="form-label">{{ t('admin.inquiries.customer_notes') }}</label>
              <textarea id="inquiry-customerNotes" v-model="form.customerNotes" name="customerNotes" rows="3" class="form-textarea"></textarea>
            </div>

            <div v-if="editingId" class="col-span-2 form-section">
              <h4 class="form-section-title">{{ t('admin.inquiries.assignment') }}</h4>
              <div class="grid grid-cols-3 gap-3">
                <input id="inquiry-assignedTo" v-model="assignment.assignedTo" name="assignedTo" autocomplete="off" :placeholder="t('admin.inquiries.assignee_placeholder')" class="form-input" />
                <select id="inquiry-assignmentPriority" v-model="assignment.priority" name="assignmentPriority" class="form-input">
                  <option value="low">{{ enumLabel('inquiry_priority', 'low') }}</option>
                  <option value="normal">{{ enumLabel('inquiry_priority', 'normal') }}</option>
                  <option value="high">{{ enumLabel('inquiry_priority', 'high') }}</option>
                  <option value="urgent">{{ enumLabel('inquiry_priority', 'urgent') }}</option>
                </select>
                <button type="button" class="btn btn-primary" :disabled="assigning" @click="assignInquiry">
                  {{ assigning ? t('admin.inquiries.assigning') : t('admin.inquiries.assign') }}
                </button>
              </div>
            </div>

            <div v-if="editingId" class="col-span-2 form-section">
              <div class="flex items-center justify-between">
                <h4 class="form-section-title">{{ t('admin.inquiries.ai_analysis') }}</h4>
                <button type="button" class="btn btn-accent" :disabled="analyzing" @click="analyzeInquiry">
                  {{ analyzing ? t('admin.inquiries.analyzing') : t('admin.inquiries.analyze') }}
                </button>
              </div>
              <div v-if="analysisResult" class="mt-3 p-3 bg-accent/10 rounded-md text-sm">
                <p><strong>{{ t('admin.inquiries.risk') }}:</strong> {{ analysisResult.riskLevel }}</p>
                <p><strong>{{ t('admin.inquiries.action') }}:</strong> {{ analysisResult.recommendedAction }}</p>
                <p><strong>{{ t('admin.inquiries.summary') }}:</strong> {{ analysisResult.summary }}</p>
              </div>
            </div>

            <div v-if="editingId" class="col-span-2 form-section">
              <h4 class="form-section-title">{{ t('admin.inquiries.quote') }}</h4>
              <div class="grid grid-cols-2 gap-3">
                <input id="inquiry-quotedAmount" v-model.number="quote.quotedAmount" name="quotedAmount" type="number" min="0" step="0.01" :placeholder="t('admin.inquiries.quoted_amount_placeholder')" class="form-input" />
                <input id="inquiry-validUntil" v-model="quote.validUntil" name="validUntil" type="datetime-local" class="form-input" />
                <input id="inquiry-quoteProducts" v-model="quote.products" name="quoteProducts" :placeholder="t('admin.inquiries.products_placeholder')" class="col-span-2 form-input" />
                <textarea id="inquiry-quoteCustomerNotes" v-model="quote.customerNotes" name="quoteCustomerNotes" rows="2" :placeholder="t('admin.inquiries.notes_placeholder')" class="col-span-2 form-textarea"></textarea>
                <button type="button" class="btn btn-highlight col-span-2" :disabled="quoting" @click="submitQuote">
                  {{ quoting ? t('admin.inquiries.submitting') : t('admin.inquiries.submit_quote') }}
                </button>
              </div>
            </div>

            <div v-if="formError" class="col-span-2 text-sm text-error">{{ formError }}</div>
            <div class="col-span-2 flex justify-end gap-3">
              <button type="button" class="btn btn-outline" @click="closeModal">
                {{ t('admin.inquiries.cancel') }}
              </button>
              <button type="submit" :disabled="saving" class="btn btn-highlight">
                {{ saving ? t('admin.inquiries.saving') : (editingId ? t('admin.inquiries.update') : t('admin.inquiries.create')) }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <p v-if="actionMessage" class="mt-4 text-sm" :class="actionError ? 'text-error' : 'text-success'">
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

const { token } = useAuth()
const { t } = useI18n()
const { enumLabel, formatNumber, formatDate } = useDisplay()
const localePath = useLocalePath()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'

const inquiries = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20

const showModal = ref(false)
const editingId = ref('')
const saving = ref(false)
const assigning = ref(false)
const analyzing = ref(false)
const quoting = ref(false)
const formError = ref('')
const actionMessage = ref('')
const actionError = ref(false)
const analysisResult = ref<any>(null)

const form = reactive({
  userId: '',
  companyName: '',
  contactPerson: '',
  email: '',
  whatsApp: '',
  targetCountry: '',
  estimatedQuantity: '',
  interestedProductsInput: '',
  expectedDelivery: '',
  oemNeeded: false,
  message: '',
  status: 'pending',
  priority: 'normal',
  customerNotes: ''
})

const assignment = reactive({
  assignedTo: '',
  priority: 'normal'
})

const quote = reactive({
  quotedAmount: 0,
  validUntil: '',
  products: '',
  customerNotes: ''
})

const parseCSV = (value: string) => value.split(',').map(v => v.trim()).filter(Boolean)

const statusClass = (status: string) => {
  if (status === 'pending') return 'badge-warning'
  if (status === 'contacted' || status === 'quoted') return 'badge-info'
  if (status === 'negotiating') return 'badge-primary'
  if (status === 'won') return 'badge-success'
  return 'badge-default'
}

const formatProducts = (products: unknown) => {
  if (Array.isArray(products)) {
    return products.length ? products.join(', ') : '-'
  }
  return '-'
}

const fetchInquiries = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await $fetch<any>(`${baseURL}/admin/inquiries?page=${page.value}&limit=${pageSize}`, {
      headers: { Authorization: `Bearer ${token.value}` }
    })
    inquiries.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.data?.message || err.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const nextPage = () => {
  if (pagination.value && page.value < pagination.value.totalPages) {
    page.value += 1
  }
}

const prevPage = () => {
  if (page.value > 1) {
    page.value -= 1
  }
}

const resetForm = () => {
  form.userId = ''
  form.companyName = ''
  form.contactPerson = ''
  form.email = ''
  form.whatsApp = ''
  form.targetCountry = ''
  form.estimatedQuantity = ''
  form.interestedProductsInput = ''
  form.expectedDelivery = ''
  form.oemNeeded = false
  form.message = ''
  form.status = 'pending'
  form.priority = 'normal'
  form.customerNotes = ''

  assignment.assignedTo = ''
  assignment.priority = 'normal'

  quote.quotedAmount = 0
  quote.validUntil = ''
  quote.products = ''
  quote.customerNotes = ''

  analysisResult.value = null
}

const fillForm = (inquiry: any) => {
  form.userId = inquiry.userId || ''
  form.companyName = inquiry.companyName || ''
  form.contactPerson = inquiry.contactPerson || ''
  form.email = inquiry.email || ''
  form.whatsApp = inquiry.whatsapp || ''
  form.targetCountry = inquiry.targetCountry || ''
  form.estimatedQuantity = inquiry.estimatedQuantity || ''
  form.interestedProductsInput = Array.isArray(inquiry.interestedProducts) ? inquiry.interestedProducts.join(', ') : ''
  form.expectedDelivery = inquiry.expectedDelivery || ''
  form.oemNeeded = Boolean(inquiry.oemNeeded)
  form.message = inquiry.message || ''
  form.status = inquiry.status || 'pending'
  form.priority = inquiry.priority || 'normal'
  form.customerNotes = inquiry.customerNotes || ''

  assignment.assignedTo = inquiry.assignedTo || ''
  assignment.priority = inquiry.priority || 'normal'

  quote.quotedAmount = inquiry.quotedAmount || 0
  quote.validUntil = inquiry.validUntil ? new Date(inquiry.validUntil).toISOString().slice(0, 16) : ''
  quote.products = Array.isArray(inquiry.products) ? inquiry.products.join(', ') : ''
  quote.customerNotes = inquiry.customerNotes || ''
}

const openCreateModal = () => {
  editingId.value = ''
  resetForm()
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false
  showModal.value = true
}

const openEditModal = async (id: string) => {
  editingId.value = id
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false
  try {
    const inquiry = await $fetch<any>(`${baseURL}/admin/inquiries/${id}`, {
      headers: { Authorization: `Bearer ${token.value}` }
    })
    fillForm(inquiry)
    showModal.value = true
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.data?.message || err.message || t('errors.api.load_failed')
  }
}

const closeModal = () => {
  showModal.value = false
  saving.value = false
  formError.value = ''
}

const buildPayload = () => {
  const payload: Record<string, any> = {
    companyName: form.companyName,
    contactPerson: form.contactPerson,
    email: form.email,
    whatsapp: form.whatsApp,
    targetCountry: form.targetCountry,
    estimatedQuantity: form.estimatedQuantity,
    interestedProducts: parseCSV(form.interestedProductsInput),
    expectedDelivery: form.expectedDelivery,
    oemNeeded: form.oemNeeded,
    message: form.message,
    status: form.status,
    priority: form.priority,
    customerNotes: form.customerNotes
  }
  if (form.userId.trim()) {
    payload.userId = form.userId.trim()
  } else if (editingId.value) {
    payload.userId = ''
  }
  return payload
}

const saveInquiry = async () => {
  if (!form.companyName.trim() || !form.contactPerson.trim() || !form.email.trim()) {
    formError.value = t('admin.inquiries.validation_required')
    return
  }

  saving.value = true
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false

  try {
    const payload = buildPayload()
    if (editingId.value) {
      await $fetch(`${baseURL}/admin/inquiries/${editingId.value}`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${token.value}` },
        body: payload
      })
      actionMessage.value = t('admin.inquiries.updated_success')
    } else {
      await $fetch(`${baseURL}/admin/inquiries`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token.value}` },
        body: payload
      })
      actionMessage.value = t('admin.inquiries.created_success')
    }
    closeModal()
    await fetchInquiries()
  } catch (err: any) {
    formError.value = err?.data?.message || err.message || t('errors.api.save_failed')
  } finally {
    saving.value = false
  }
}

const deleteInquiry = async (id: string) => {
  if (!confirm(t('admin.inquiries.confirm_delete'))) return

  actionMessage.value = ''
  actionError.value = false
  try {
    await $fetch(`${baseURL}/admin/inquiries/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token.value}` }
    })
    actionMessage.value = t('admin.inquiries.deleted_success')
    await fetchInquiries()
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.data?.message || err.message || t('errors.api.delete_failed')
  }
}

const assignInquiry = async () => {
  if (!editingId.value) return
  if (!assignment.assignedTo.trim()) {
    formError.value = t('admin.inquiries.validation_assignee')
    return
  }

  assigning.value = true
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false
  try {
    await $fetch(`${baseURL}/admin/inquiries/${editingId.value}/assign`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${token.value}` },
      body: {
        assignedTo: assignment.assignedTo.trim(),
        priority: assignment.priority
      }
    })
    form.priority = assignment.priority
    actionMessage.value = t('admin.inquiries.assigned_success')
    await fetchInquiries()
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.data?.message || err.message || t('errors.api.assign_failed')
  } finally {
    assigning.value = false
  }
}

const analyzeInquiry = async () => {
  if (!editingId.value) return
  analyzing.value = true
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false
  try {
    const res = await $fetch<any>(`${baseURL}/admin/inquiries/${editingId.value}/analyze`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token.value}` }
    })
    analysisResult.value = res.analysis || null
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.data?.message || err.message || t('errors.api.analyze_failed')
  } finally {
    analyzing.value = false
  }
}

const submitQuote = async () => {
  if (!editingId.value) return
  if (quote.quotedAmount <= 0) {
    formError.value = t('admin.inquiries.validation_quote_amount')
    return
  }

  quoting.value = true
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false
  try {
    await $fetch(`${baseURL}/admin/inquiries/${editingId.value}/quote`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token.value}` },
      body: {
        quotedAmount: quote.quotedAmount,
        validUntil: quote.validUntil ? new Date(quote.validUntil).toISOString() : '',
        products: parseCSV(quote.products),
        customerNotes: quote.customerNotes
      }
    })
    form.status = 'quoted'
    actionMessage.value = t('admin.inquiries.quote_submitted')
    await fetchInquiries()
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.data?.message || err.message || t('errors.api.quote_failed')
  } finally {
    quoting.value = false
  }
}

watch(page, fetchInquiries)
onMounted(fetchInquiries)
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

.form-checkbox {
  width: 18px;
  height: 18px;
  border: 2px solid var(--color-border);
  border-radius: var(--radius-sm);
  cursor: pointer;
  accent-color: var(--color-highlight);
}

.form-section {
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border);
}

.form-section-title {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: var(--spacing-sm);
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
