<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('admin.invoices.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ t('admin.invoices.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0 flex items-center gap-3">
        <button @click="exportInvoices" class="inline-flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors">
          <Icon name="heroicons:arrow-down-tray" class="h-4 w-4" />
          {{ t('admin.invoices.export') }}
        </button>
        <button @click="openCreateModal" class="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors">
          <Icon name="heroicons:plus" class="h-4 w-4" />
          {{ t('admin.invoices.create_invoice') }}
        </button>
      </div>
    </div>

    <!-- Stats -->
    <div class="mt-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="stat in statsCards" :key="stat.label" class="bg-white rounded-xl shadow-sm border border-gray-200 p-5">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500">{{ stat.label }}</p>
            <p class="mt-1 text-2xl font-bold" :class="stat.color">{{ stat.value }}</p>
          </div>
          <div :class="`h-12 w-12 rounded-lg ${stat.bgColor} flex items-center justify-center`">
            <Icon :name="stat.icon" class="h-6 w-6" :class="stat.color" />
          </div>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 p-4">
      <div class="flex flex-wrap items-center gap-4">
        <div class="flex-1 min-w-[200px]">
          <div class="relative">
            <Icon name="heroicons:magnifying-glass" class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
            <input v-model="searchQuery" type="text" :placeholder="t('admin.invoices.search')" class="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
          </div>
        </div>
        <select v-model="statusFilter" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
          <option value="all">{{ t('admin.invoices.filter_all') }}</option>
          <option value="draft">{{ t('admin.invoices.status_draft') }}</option>
          <option value="sent">{{ t('admin.invoices.status_sent') }}</option>
          <option value="paid">{{ t('admin.invoices.status_paid') }}</option>
          <option value="overdue">{{ t('admin.invoices.status_overdue') }}</option>
          <option value="cancelled">{{ t('admin.invoices.status_cancelled') }}</option>
        </select>
        <input v-model="dateFrom" type="date" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
        <input v-model="dateTo" type="date" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
      </div>
    </div>

    <!-- Invoices Table -->
    <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.invoices.col_invoice') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.invoices.col_order') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.invoices.col_customer') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider text-right">{{ t('admin.invoices.col_amount') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.invoices.col_date') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.invoices.col_due_date') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.invoices.col_status') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.invoices.col_actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white">
            <tr v-if="pending">
              <td colspan="8" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.invoices.loading') }}</td>
            </tr>
            <tr v-else-if="error">
              <td colspan="8" class="px-6 py-10 text-center text-sm text-red-600">{{ error }}</td>
            </tr>
            <tr v-else-if="filteredInvoices.length === 0">
              <td colspan="8" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.invoices.no_data') }}</td>
            </tr>
            <tr v-else v-for="invoice in filteredInvoices" :key="invoice.id" class="hover:bg-gray-50 transition-colors">
              <td class="px-6 py-4">
                <div>
                  <span class="font-mono font-medium text-gray-900">{{ invoice.invoiceNumber }}</span>
                  <p v-if="invoice.type" class="text-xs text-gray-500 mt-0.5 uppercase">{{ invoice.type }}</p>
                </div>
              </td>
              <td class="px-6 py-4 text-sm">
                <NuxtLink v-if="invoice.orderId" :to="`/admin/orders/${invoice.orderId}`" class="text-blue-600 hover:text-blue-900">
                  #{{ invoice.orderNumber || invoice.orderId.substring(0, 8) }}
                </NuxtLink>
                <span v-else class="text-gray-400">-</span>
              </td>
              <td class="px-6 py-4">
                <div class="text-sm font-medium text-gray-900">{{ invoice.customerName }}</div>
                <div class="text-xs text-gray-500">{{ invoice.customerEmail }}</div>
              </td>
              <td class="px-6 py-4 text-right">
                <span class="font-semibold text-gray-900">{{ invoice.currency || 'USD' }} {{ (invoice.totalAmount || 0).toLocaleString() }}</span>
                <p v-if="invoice.paidAmount > 0" class="text-xs text-emerald-600">Paid: {{ invoice.currency || 'USD' }} {{ invoice.paidAmount.toLocaleString() }}</p>
              </td>
              <td class="px-6 py-4 text-sm text-gray-600">
                {{ invoice.invoiceDate ? new Date(invoice.invoiceDate).toLocaleDateString() : '-' }}
              </td>
              <td class="px-6 py-4 text-sm">
                <span :class="isOverdue(invoice) ? 'text-red-600 font-medium' : 'text-gray-600'">
                  {{ invoice.dueDate ? new Date(invoice.dueDate).toLocaleDateString() : '-' }}
                </span>
              </td>
              <td class="px-6 py-4">
                <span :class="statusBadgeClass(invoice.status)" class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold">
                  {{ formatStatus(invoice.status) }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm">
                <button @click="viewInvoice(invoice)" class="text-blue-600 hover:text-blue-900 mr-3">{{ t('admin.invoices.view') }}</button>
                <button @click="openEditModal(invoice)" class="text-gray-600 hover:text-gray-900 mr-3">{{ t('admin.invoices.edit') }}</button>
                <button v-if="invoice.status === 'draft'" @click="sendInvoice(invoice)" class="text-emerald-600 hover:text-emerald-900">{{ t('admin.invoices.send') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="pagination" class="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
        <div class="text-sm text-gray-600">
          {{ t('admin.invoices.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
        </div>
        <div class="flex items-center gap-2">
          <button @click="prevPage" :disabled="page <= 1" class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed">
            {{ t('admin.invoices.previous') }}
          </button>
          <button @click="nextPage" :disabled="page >= pagination.totalPages" class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed">
            {{ t('admin.invoices.next') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="closeModal"></div>
        <div class="inline-block w-full transform overflow-hidden rounded-xl bg-white text-left align-bottom shadow-xl transition-all sm:my-8 sm:max-w-2xl sm:align-middle">
          <div class="bg-gradient-to-r from-blue-600 to-indigo-600 px-6 py-4">
            <h3 class="text-lg font-semibold text-white">{{ editingId ? t('admin.invoices.edit_invoice') : t('admin.invoices.create_invoice') }}</h3>
          </div>
          <form @submit.prevent="saveInvoice" class="p-6 space-y-4">
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.invoice_number') }}</label>
                <input v-model="form.invoiceNumber" type="text" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.type') }}</label>
                <select v-model="form.type" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option value="invoice">Invoice</option>
                  <option value="proforma">Proforma</option>
                  <option value="credit_note">Credit Note</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.order') }}</label>
                <select v-model="form.orderId" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option value="">{{ t('admin.invoices.select_order') }}</option>
                  <option v-for="order in orders" :key="order.id" :value="order.id">
                    #{{ order.orderNumber || order.id.substring(0, 8) }}
                  </option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.currency') }}</label>
                <input v-model="form.currency" type="text" placeholder="USD" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 uppercase" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.total_amount') }}</label>
                <input v-model.number="form.totalAmount" type="number" step="0.01" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.paid_amount') }}</label>
                <input v-model.number="form.paidAmount" type="number" step="0.01" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.invoice_date') }}</label>
                <input v-model="form.invoiceDate" type="date" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.due_date') }}</label>
                <input v-model="form.dueDate" type="date" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.status') }}</label>
                <select v-model="form.status" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option value="draft">Draft</option>
                  <option value="sent">Sent</option>
                  <option value="paid">Paid</option>
                  <option value="overdue">Overdue</option>
                  <option value="cancelled">Cancelled</option>
                </select>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.notes') }}</label>
              <textarea v-model="form.notes" rows="2" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"></textarea>
            </div>
            <div v-if="formError" class="text-sm text-red-600">{{ formError }}</div>
            <div class="flex justify-end gap-3 pt-4 border-t border-gray-100">
              <button type="button" @click="closeModal" class="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100">
                {{ t('admin.invoices.cancel') }}
              </button>
              <button type="submit" :disabled="saving" class="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 flex items-center gap-2">
                <Icon v-if="saving" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" />
                {{ t('admin.invoices.save') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Invoice Preview Modal -->
    <div v-if="showPreviewModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="showPreviewModal = false"></div>
        <div class="inline-block w-full transform overflow-hidden rounded-xl bg-white text-left align-bottom shadow-xl transition-all sm:my-8 sm:max-w-3xl sm:align-middle">
          <div class="bg-gradient-to-r from-gray-700 to-gray-900 px-6 py-4 flex items-center justify-between">
            <h3 class="text-lg font-semibold text-white">{{ t('admin.invoices.invoice_preview') }}</h3>
            <div class="flex items-center gap-2">
              <button @click="printInvoice" class="px-3 py-1.5 bg-white/10 text-white text-sm rounded-lg hover:bg-white/20 transition-colors flex items-center gap-1">
                <Icon name="heroicons:printer" class="h-4 w-4" />
                {{ t('admin.invoices.print') }}
              </button>
              <button @click="showPreviewModal = false" class="text-gray-300 hover:text-white">
                <Icon name="heroicons:x-mark" class="h-5 w-5" />
              </button>
            </div>
          </div>
          <div v-if="previewInvoice" class="p-8 bg-white max-h-[70vh] overflow-y-auto" id="invoice-print-area">
            <!-- Invoice Header -->
            <div class="flex justify-between items-start mb-8">
              <div>
                <h1 class="text-2xl font-bold text-gray-900">INVOICE</h1>
                <p class="text-gray-500 mt-1">{{ previewInvoice.invoiceNumber }}</p>
                <p class="text-sm text-gray-500 mt-1">
                  {{ previewInvoice.type === 'proforma' ? 'PROFORMA INVOICE' : previewInvoice.type === 'credit_note' ? 'CREDIT NOTE' : 'TAX INVOICE' }}
                </p>
              </div>
              <div class="text-right">
                <h2 class="text-xl font-bold text-blue-600">CandyPro OEM</h2>
                <p class="text-sm text-gray-500">B2B Candy Manufacturing</p>
              </div>
            </div>

            <!-- Dates & Status -->
            <div class="grid grid-cols-2 gap-8 mb-8">
              <div>
                <p class="text-sm font-medium text-gray-500 mb-1">Invoice Date</p>
                <p class="text-gray-900">{{ previewInvoice.invoiceDate ? new Date(previewInvoice.invoiceDate).toLocaleDateString() : '-' }}</p>
              </div>
              <div>
                <p class="text-sm font-medium text-gray-500 mb-1">Due Date</p>
                <p class="text-gray-900">{{ previewInvoice.dueDate ? new Date(previewInvoice.dueDate).toLocaleDateString() : '-' }}</p>
              </div>
            </div>

            <!-- Bill To -->
            <div class="mb-8">
              <p class="text-sm font-medium text-gray-500 mb-1">Bill To</p>
              <p class="font-semibold text-gray-900">{{ previewInvoice.customerName }}</p>
              <p class="text-sm text-gray-600">{{ previewInvoice.customerEmail }}</p>
            </div>

            <!-- Amount Summary -->
            <div class="border-t border-b border-gray-200 py-6 mb-8">
              <div class="flex justify-between items-center">
                <span class="text-lg font-medium text-gray-700">Total Amount</span>
                <span class="text-3xl font-bold text-gray-900">{{ previewInvoice.currency || 'USD' }} {{ (previewInvoice.totalAmount || 0).toLocaleString() }}</span>
              </div>
              <div v-if="previewInvoice.paidAmount > 0" class="mt-2 flex justify-between items-center">
                <span class="text-emerald-600">Paid Amount</span>
                <span class="text-lg font-semibold text-emerald-600">{{ previewInvoice.currency || 'USD' }} {{ previewInvoice.paidAmount.toLocaleString() }}</span>
              </div>
            </div>

            <!-- Notes -->
            <div v-if="previewInvoice.notes" class="text-sm text-gray-600">
              <p class="font-medium text-gray-700 mb-1">Notes</p>
              <p>{{ previewInvoice.notes }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()

const invoices = ref<any[]>([])
const orders = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20
const searchQuery = ref('')
const statusFilter = ref('all')
const dateFrom = ref('')
const dateTo = ref('')
const showModal = ref(false)
const editingId = ref('')
const saving = ref(false)
const formError = ref('')
const showPreviewModal = ref(false)
const previewInvoice = ref<any>(null)

const form = reactive({
  invoiceNumber: '', type: 'invoice', orderId: '', currency: 'USD',
  totalAmount: 0, paidAmount: 0, invoiceDate: '', dueDate: '', status: 'draft', notes: ''
})

const statsCards = computed(() => {
  const all = invoices.value.length
  const draft = invoices.value.filter(i => i.status === 'draft').length
  const sent = invoices.value.filter(i => i.status === 'sent').length
  const paid = invoices.value.filter(i => i.status === 'paid').length
  return [
    { label: t('admin.invoices.total'), value: all, icon: 'heroicons:document-text', color: 'text-blue-600', bgColor: 'bg-blue-50' },
    { label: t('admin.invoices.draft'), value: draft, icon: 'heroicons:pencil', color: 'text-gray-600', bgColor: 'bg-gray-50' },
    { label: t('admin.invoices.sent'), value: sent, icon: 'heroicons:paper-airplane', color: 'text-yellow-600', bgColor: 'bg-yellow-50' },
    { label: t('admin.invoices.paid'), value: paid, icon: 'heroicons:check-circle', color: 'text-emerald-600', bgColor: 'bg-emerald-50' }
  ]
})

const filteredInvoices = computed(() => invoices.value.filter(inv => {
  const matchesSearch = !searchQuery.value ||
    inv.invoiceNumber?.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    inv.customerName?.toLowerCase().includes(searchQuery.value.toLowerCase())
  return matchesSearch && (statusFilter.value === 'all' || inv.status === statusFilter.value)
}))

const fetchInvoices = async () => {
  pending.value = true; error.value = ''
  try {
    const res = await api.get<any>('/admin/invoices', { page: page.value, limit: pageSize })
    invoices.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || 'Failed to fetch invoices'
  } finally { pending.value = false }
}

const fetchOrders = async () => {
  try { const res = await api.get<any>('/admin/orders?limit=100'); orders.value = res.data || [] }
  catch { orders.value = [] }
}

const openCreateModal = () => {
  editingId.value = ''
  const today = new Date().toISOString().split('T')[0]
  const dueDate = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
  Object.assign(form, { invoiceNumber: `INV-${Date.now().toString().slice(-8)}`, type: 'invoice', orderId: '', currency: 'USD', totalAmount: 0, paidAmount: 0, invoiceDate: today, dueDate, status: 'draft', notes: '' })
  formError.value = ''; showModal.value = true
}

const openEditModal = (invoice: any) => {
  editingId.value = invoice.id
  Object.assign(form, { invoiceNumber: invoice.invoiceNumber || '', type: invoice.type || 'invoice', orderId: invoice.orderId || '', currency: invoice.currency || 'USD', totalAmount: invoice.totalAmount || 0, paidAmount: invoice.paidAmount || 0, invoiceDate: invoice.invoiceDate ? invoice.invoiceDate.split('T')[0] : '', dueDate: invoice.dueDate ? invoice.dueDate.split('T')[0] : '', status: invoice.status || 'draft', notes: invoice.notes || '' })
  formError.value = ''; showModal.value = true
}

const closeModal = () => { showModal.value = false; saving.value = false }

const saveInvoice = async () => {
  saving.value = true; formError.value = ''
  try {
    if (editingId.value) { await api.put(`/admin/invoices/${editingId.value}`, form) }
    else { await api.post('/admin/invoices', form) }
    closeModal(); await fetchInvoices()
  } catch (err: any) { formError.value = err?.message || 'Failed to save invoice' }
  finally { saving.value = false }
}

const viewInvoice = (invoice: any) => { previewInvoice.value = invoice; showPreviewModal.value = true }

const sendInvoice = async (invoice: any) => {
  try { await api.post(`/admin/invoices/${invoice.id}/send`, {}); await fetchInvoices() }
  catch (err: any) { alert(err?.message || 'Failed to send invoice') }
}

const printInvoice = () => {
  window.print()
}

const exportInvoices = () => {
  const csv = [['Invoice #', 'Order', 'Customer', 'Amount', 'Paid', 'Status', 'Date', 'Due Date'].join(','), ...filteredInvoices.value.map(inv => [inv.invoiceNumber, inv.orderNumber || inv.orderId || '', `"${inv.customerName || ''}"`, inv.totalAmount || 0, inv.paidAmount || 0, inv.status, inv.invoiceDate || '', inv.dueDate || ''].join(','))].join('\n')
  const blob = new Blob([csv], { type: 'text/csv' }); const url = URL.createObjectURL(blob)
  const a = document.createElement('a'); a.href = url; a.download = `invoices-${new Date().toISOString().split('T')[0]}.csv`; a.click()
}

const statusBadgeClass = (status: string) => {
  if (status === 'draft') return 'bg-gray-100 text-gray-800'; if (status === 'sent') return 'bg-yellow-100 text-yellow-800'
  if (status === 'paid') return 'bg-emerald-100 text-emerald-800'; if (status === 'overdue') return 'bg-red-100 text-red-800'
  return 'bg-gray-100 text-gray-800'
}
const formatStatus = (status: string) => status?.charAt(0).toUpperCase() + status?.slice(1) || 'Unknown'
const isOverdue = (invoice: any) => { if (!invoice.dueDate || invoice.status === 'paid' || invoice.status === 'cancelled') return false; return new Date(invoice.dueDate) < new Date() }
const prevPage = () => { if (page.value > 1) { page.value -= 1; fetchInvoices() } }
const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) { page.value += 1; fetchInvoices() } }

onMounted(() => { fetchInvoices(); fetchOrders() })
</script>
