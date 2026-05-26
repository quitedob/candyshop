<template>
  <div>
    <PageHeader :title="t('admin.inquiries.title')" :description="t('admin.inquiries.description')">
      <template #actions>
        <button type="button" class="btn btn-highlight" @click="openCreateModal">
          {{ t('admin.inquiries.new_inquiry') }}
        </button>
      </template>
    </PageHeader>

    <AdminTable
      :columns="columns"
      :rows="inquiries"
      :loading="pending"
      :error="!!error"
      :error-message="error"
      :empty-text="t('admin.inquiries.no_data')"
      @retry="fetchInquiries"
    >
      <template #cell-createdAt="{ row }">
        {{ formatDate(row.createdAt) }}
      </template>

      <template #cell-company="{ row }">
        <div class="font-medium">{{ row.companyName }}</div>
        <div class="text-light">{{ row.contactPerson }} ({{ row.email }})</div>
      </template>

      <template #cell-products="{ row }">
        {{ formatProducts(row.interestedProducts) }}
        <span v-if="row.productIds?.length" class="block text-xs text-gray-400 font-mono mt-0.5">
          ID: {{ row.productIds.join(', ') }}
        </span>
      </template>

      <template #cell-status="{ row }">
        <StatusBadge :status="row.status" type="inquiry" />
      </template>

      <template #cell-actions="{ row }">
        <NuxtLink :to="localePath(`/admin/inquiries/${row.id}`)" class="text-orange-600 hover:text-orange-900 mr-3">{{ t('admin.inquiries.view') }}</NuxtLink>
        <button type="button" class="text-orange-600 hover:text-orange-900 mr-3" @click="openEditModal(row.id)">{{ t('admin.inquiries.edit') }}</button>
        <button type="button" class="text-red-600 hover:text-red-900" @click="deleteInquiry(row.id)">{{ t('admin.inquiries.delete') }}</button>
      </template>

      <template #bottom>
        <div v-if="pagination" class="flex items-center justify-between">
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
      </template>
    </AdminTable>

    <!-- Editor Modal -->
    <AdminModal :open="showModal" :title="editingId ? t('admin.inquiries.edit_inquiry') : t('admin.inquiries.create_inquiry')" width="xl" @close="closeModal">
      <form class="grid grid-cols-1 gap-4 sm:grid-cols-2" @submit.prevent="saveInquiry">
        <div>
          <label for="inquiry-userId" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.user_id') }}</label>
          <input id="inquiry-userId" v-model="form.userId" name="userId" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label for="inquiry-companyName" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.company_name') }}</label>
          <input id="inquiry-companyName" v-model="form.companyName" name="companyName" autocomplete="organization" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label for="inquiry-contactPerson" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.contact_person') }}</label>
          <input id="inquiry-contactPerson" v-model="form.contactPerson" name="contactPerson" autocomplete="name" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label for="inquiry-email" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.email') }}</label>
          <input id="inquiry-email" v-model="form.email" name="email" type="email" autocomplete="email" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label for="inquiry-whatsApp" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.whatsapp') }}</label>
          <input id="inquiry-whatsApp" v-model="form.whatsApp" name="whatsApp" autocomplete="tel" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label for="inquiry-targetCountry" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.target_country') }}</label>
          <input id="inquiry-targetCountry" v-model="form.targetCountry" name="targetCountry" autocomplete="country-name" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label for="inquiry-estimatedQuantity" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.estimated_quantity') }}</label>
          <input id="inquiry-estimatedQuantity" v-model="form.estimatedQuantity" name="estimatedQuantity" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label for="inquiry-expectedDelivery" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.expected_delivery') }}</label>
          <input id="inquiry-expectedDelivery" v-model="form.expectedDelivery" name="expectedDelivery" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label for="inquiry-status" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.status') }}</label>
          <select id="inquiry-status" v-model="form.status" name="status" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
            <option value="pending">{{ enumLabel('inquiry_status', 'pending') }}</option>
            <option value="contacted">{{ enumLabel('inquiry_status', 'contacted') }}</option>
            <option value="quoted">{{ enumLabel('inquiry_status', 'quoted') }}</option>
            <option value="negotiating">{{ enumLabel('inquiry_status', 'negotiating') }}</option>
            <option value="won">{{ enumLabel('inquiry_status', 'won') }}</option>
            <option value="lost">{{ enumLabel('inquiry_status', 'lost') }}</option>
            <option value="closed">{{ enumLabel('inquiry_status', 'closed') }}</option>
          </select>
        </div>
        <div>
          <label for="inquiry-priority" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.priority') }}</label>
          <select id="inquiry-priority" v-model="form.priority" name="priority" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
            <option value="low">{{ enumLabel('inquiry_priority', 'low') }}</option>
            <option value="normal">{{ enumLabel('inquiry_priority', 'normal') }}</option>
            <option value="high">{{ enumLabel('inquiry_priority', 'high') }}</option>
            <option value="urgent">{{ enumLabel('inquiry_priority', 'urgent') }}</option>
          </select>
        </div>
        <div class="sm:col-span-2">
          <label for="inquiry-interestedProductsInput" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.interested_products') }}</label>
          <input id="inquiry-interestedProductsInput" v-model="form.interestedProductsInput" name="interestedProductsInput" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div class="sm:col-span-2">
          <label for="inquiry-productIdsInput" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.product_ids') }}</label>
          <input id="inquiry-productIdsInput" v-model="form.productIdsInput" name="productIdsInput" autocomplete="off" :placeholder="t('admin.inquiries.product_ids_placeholder')" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm font-mono text-xs" />
          <p class="mt-1 text-xs text-gray-500">{{ t('admin.inquiries.product_ids_hint') }}</p>
        </div>
        <div class="sm:col-span-2">
          <label class="inline-flex items-center gap-2">
            <input id="inquiry-oemNeeded" v-model="form.oemNeeded" name="oemNeeded" type="checkbox" class="rounded border-gray-300 text-orange-600 focus:ring-orange-500" />
            <span class="text-sm">{{ t('admin.inquiries.oem_needed') }}</span>
          </label>
        </div>
        <div class="sm:col-span-2">
          <label for="inquiry-message" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.message') }}</label>
          <textarea id="inquiry-message" v-model="form.message" name="message" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
        </div>
        <div class="sm:col-span-2">
          <label for="inquiry-customerNotes" class="block text-sm font-medium text-gray-700">{{ t('admin.inquiries.customer_notes') }}</label>
          <textarea id="inquiry-customerNotes" v-model="form.customerNotes" name="customerNotes" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
        </div>

        <!-- Assignment section (edit only) -->
        <template v-if="editingId">
          <div class="sm:col-span-2 pt-4 border-t border-gray-200">
            <h4 class="text-sm font-semibold text-gray-900 mb-3">{{ t('admin.inquiries.assignment') }}</h4>
            <div class="grid grid-cols-3 gap-3">
              <input id="inquiry-assignedTo" v-model="assignment.assignedTo" name="assignedTo" autocomplete="off" :placeholder="t('admin.inquiries.assignee_placeholder')" class="rounded-md border border-gray-300 px-3 py-2 text-sm" />
              <select id="inquiry-assignmentPriority" v-model="assignment.priority" name="assignmentPriority" class="rounded-md border border-gray-300 px-3 py-2 text-sm">
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

          <!-- AI Analysis section -->
          <div class="sm:col-span-2 pt-4 border-t border-gray-200">
            <div class="flex items-center justify-between">
              <h4 class="text-sm font-semibold text-gray-900">{{ t('admin.inquiries.ai_analysis') }}</h4>
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

          <!-- Quote section -->
          <div class="sm:col-span-2 pt-4 border-t border-gray-200">
            <h4 class="text-sm font-semibold text-gray-900 mb-3">{{ t('admin.inquiries.quote') }}</h4>
            <div class="grid grid-cols-2 gap-3">
              <input id="inquiry-quotedAmount" v-model.number="quote.quotedAmount" name="quotedAmount" type="number" min="0" step="0.01" :placeholder="t('admin.inquiries.quoted_amount_placeholder')" class="rounded-md border border-gray-300 px-3 py-2 text-sm" />
              <input id="inquiry-validUntil" v-model="quote.validUntil" name="validUntil" type="datetime-local" class="rounded-md border border-gray-300 px-3 py-2 text-sm" />
              <input id="inquiry-quoteProducts" v-model="quote.products" name="quoteProducts" :placeholder="t('admin.inquiries.products_placeholder')" class="col-span-2 rounded-md border border-gray-300 px-3 py-2 text-sm" />
              <textarea id="inquiry-quoteCustomerNotes" v-model="quote.customerNotes" name="quoteCustomerNotes" rows="2" :placeholder="t('admin.inquiries.notes_placeholder')" class="col-span-2 rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
              <button type="button" class="btn btn-highlight col-span-2" :disabled="quoting" @click="submitQuote">
                {{ quoting ? t('admin.inquiries.submitting') : t('admin.inquiries.submit_quote') }}
              </button>
            </div>
          </div>
        </template>

        <div v-if="formError" class="sm:col-span-2 text-sm text-red-600">{{ formError }}</div>
        <div class="sm:col-span-2 flex justify-end gap-3">
          <button type="button" class="btn btn-outline" @click="closeModal">
            {{ t('admin.inquiries.cancel') }}
          </button>
          <button type="submit" :disabled="saving" class="btn btn-highlight">
            {{ saving ? t('admin.inquiries.saving') : (editingId ? t('admin.inquiries.update') : t('admin.inquiries.create')) }}
          </button>
        </div>
      </form>
    </AdminModal>

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
const { enumLabel, formatDate } = useDisplay()
const localePath = useLocalePath()

const columns = [
  { key: 'createdAt', label: t('admin.inquiries.col_date') },
  { key: 'company', label: t('admin.inquiries.col_company_contact') },
  { key: 'products', label: t('admin.inquiries.col_products') },
  { key: 'status', label: t('admin.inquiries.col_status') },
  { key: 'actions', label: '' },
]

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
  productIdsInput: '',
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

const formatProducts = (products: unknown) => {
  if (Array.isArray(products)) return products.length ? products.join(', ') : '-'
  return '-'
}

const fetchInquiries = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await api.get<any>(`/admin/inquiries?page=${page.value}&limit=${pageSize}`)
    inquiries.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const nextPage = () => {
  if (pagination.value && page.value < pagination.value.totalPages) page.value += 1
}

const prevPage = () => {
  if (page.value > 1) page.value -= 1
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
  form.productIdsInput = ''
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
  form.whatsApp = inquiry.whatsapp || inquiry.whatsApp || ''
  form.targetCountry = inquiry.targetCountry || ''
  form.estimatedQuantity = inquiry.estimatedQuantity || ''
  form.interestedProductsInput = Array.isArray(inquiry.interestedProducts) ? inquiry.interestedProducts.join(', ') : ''
  form.productIdsInput = Array.isArray(inquiry.productIds) ? inquiry.productIds.join(', ') : ''
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
    const inquiry = await api.get<any>(`/admin/inquiries/${id}`)
    fillForm(inquiry)
    showModal.value = true
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('errors.api.load_failed')
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
    productIds: parseCSV(form.productIdsInput),
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
      await api.put(`/admin/inquiries/${editingId.value}`, payload)
      actionMessage.value = t('admin.inquiries.updated_success')
    } else {
      await api.post('/admin/inquiries', payload)
      actionMessage.value = t('admin.inquiries.created_success')
    }
    closeModal()
    await fetchInquiries()
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.save_failed')
  } finally {
    saving.value = false
  }
}

const deleteInquiry = async (id: string) => {
  if (!confirm(t('admin.inquiries.confirm_delete'))) return

  actionMessage.value = ''
  actionError.value = false
  try {
    await api.del(`/admin/inquiries/${id}`)
    actionMessage.value = t('admin.inquiries.deleted_success')
    await fetchInquiries()
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('errors.api.delete_failed')
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
    await api.put(`/admin/inquiries/${editingId.value}/assign`, {
      assignedTo: assignment.assignedTo.trim(),
      priority: assignment.priority
    })
    form.priority = assignment.priority
    actionMessage.value = t('admin.inquiries.assigned_success')
    await fetchInquiries()
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('errors.api.assign_failed')
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
    const res = await api.post<any>(`/admin/inquiries/${editingId.value}/analyze`)
    analysisResult.value = res.analysis || null
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('errors.api.analyze_failed')
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
    await api.post(`/admin/inquiries/${editingId.value}/quote`, {
      quotedAmount: quote.quotedAmount,
      validUntil: quote.validUntil ? new Date(quote.validUntil).toISOString() : '',
      products: parseCSV(quote.products),
      customerNotes: quote.customerNotes
    })
    form.status = 'quoted'
    actionMessage.value = t('admin.inquiries.quote_submitted')
    await fetchInquiries()
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('errors.api.quote_failed')
  } finally {
    quoting.value = false
  }
}

watch(page, fetchInquiries)
onMounted(fetchInquiries)
</script>
