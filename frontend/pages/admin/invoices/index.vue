<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('admin.invoices.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ t('admin.invoices.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0 flex items-center gap-3">
        <button @click="openFromOrderModal" class="inline-flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors">
          <Icon name="heroicons:document-plus" class="h-4 w-4" aria-hidden="true" />
          {{ t('admin.invoices.from_order') }}
        </button>
        <button @click="exportInvoices" class="inline-flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors">
          <Icon name="heroicons:arrow-down-tray" class="h-4 w-4" aria-hidden="true" />
          {{ t('admin.invoices.export') }}
        </button>
        <button @click="openCreateModal" class="inline-flex items-center gap-2 px-4 py-2 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 transition-colors">
          <Icon name="heroicons:plus" class="h-4 w-4" aria-hidden="true" />
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
            <Icon :name="stat.icon" class="h-6 w-6" :class="stat.color" aria-hidden="true" />
          </div>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 p-4">
      <div class="flex flex-wrap items-center gap-4">
        <div class="flex-1 min-w-[200px]">
          <div class="relative">
            <label for="invoice-search" class="sr-only">{{ t('admin.invoices.search') }}</label>
            <Icon name="heroicons:magnifying-glass" class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" aria-hidden="true" />
            <input id="invoice-search" v-model="searchQuery" name="search" type="text" :placeholder="t('admin.invoices.search')" class="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
          </div>
        </div>
        <label for="invoice-filter-status" class="sr-only">{{ t('admin.invoices.filter_status') }}</label>
        <select id="invoice-filter-status" v-model="statusFilter" name="statusFilter" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500">
          <option value="all">{{ t('admin.invoices.filter_all') }}</option>
          <option value="draft">{{ t('admin.invoices.status_draft') }}</option>
          <option value="sent">{{ t('admin.invoices.status_sent') }}</option>
          <option value="paid">{{ t('admin.invoices.status_paid') }}</option>
          <option value="overdue">{{ t('admin.invoices.status_overdue') }}</option>
          <option value="voided">{{ t('admin.invoices.status_voided') }}</option>
        </select>
        <label for="invoice-dateFrom" class="sr-only">{{ t('admin.invoices.date_from') }}</label>
        <input id="invoice-dateFrom" v-model="dateFrom" name="dateFrom" type="date" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
        <label for="invoice-dateTo" class="sr-only">{{ t('admin.invoices.date_to') }}</label>
        <input id="invoice-dateTo" v-model="dateTo" name="dateTo" type="date" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
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
                  <p v-if="invoice.type" class="text-xs text-gray-500 mt-0.5 uppercase">{{ enumLabel('invoice_type', invoice.type) }}</p>
                </div>
              </td>
              <td class="px-6 py-4 text-sm">
                <NuxtLink v-if="invoice.orderId" :to="localePath(`/admin/orders/${invoice.orderId}`)" class="text-orange-600 hover:text-orange-900">
                  #{{ invoice.orderNumber || invoice.orderId.substring(0, 8) }}
                </NuxtLink>
                <span v-else class="text-gray-400">{{ cell(null) }}</span>
              </td>
              <td class="px-6 py-4">
                <div class="text-sm font-medium text-gray-900">{{ invoice.customerName }}</div>
                <div class="text-xs text-gray-500">{{ invoice.customerEmail }}</div>
              </td>
              <td class="px-6 py-4 text-right">
                <span class="font-semibold text-gray-900">{{ cur(invoice.currency) }} {{ formatNumber(invoice.totalAmount || 0) }}</span>
                <p v-if="invoice.paidAmount > 0" class="text-xs text-emerald-600">{{ t('admin.invoices.paid_amount_line', { currency: cur(invoice.currency), amount: formatNumber(invoice.paidAmount) }) }}</p>
              </td>
              <td class="px-6 py-4 text-sm text-gray-600">
                {{ formatDate(invoice.invoiceDate) }}
              </td>
              <td class="px-6 py-4 text-sm">
                <span :class="isOverdue(invoice) ? 'text-red-600 font-medium' : 'text-gray-600'">
                  {{ formatDate(invoice.dueDate) }}
                </span>
              </td>
              <td class="px-6 py-4">
                <span :class="statusBadgeClass(invoice.status)" class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold">
                  {{ formatStatus(invoice.status) }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm">
                <button @click="viewInvoice(invoice)" class="text-orange-600 hover:text-orange-900 mr-3">{{ t('admin.invoices.view') }}</button>
                <button @click="openEditModal(invoice)" class="text-gray-600 hover:text-gray-900 mr-3">{{ t('admin.invoices.edit') }}</button>
                <button @click="exportInvoiceDoc(invoice)" class="text-gray-600 hover:text-gray-900 mr-3">{{ t('admin.invoices.export_doc') }}</button>
                <button v-if="invoice.status === 'draft'" @click="sendInvoice(invoice)" class="text-emerald-600 hover:text-emerald-900 mr-3">{{ t('admin.invoices.send') }}</button>
                <button v-if="invoice.status === 'draft'" @click="deleteInvoice(invoice)" class="text-red-600 hover:text-red-900">{{ t('admin.invoices.delete') }}</button>
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
        <button type="button" class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity w-full border-0 cursor-pointer" @click="closeModal" :aria-label="t('close')"></button>
        <div class="inline-block w-full transform overflow-hidden rounded-xl bg-white text-left align-bottom shadow-xl sm:my-8 sm:max-w-2xl sm:align-middle">
          <div class="bg-gradient-to-r from-orange-500 to-amber-600 px-6 py-4">
            <h3 class="text-lg font-semibold text-white">{{ editingId ? t('admin.invoices.edit_invoice') : t('admin.invoices.create_invoice') }}</h3>
          </div>
          <form @submit.prevent="saveInvoice" class="p-6 space-y-4">
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label for="invoice-invoiceNumber" class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.invoice_number') }}</label>
                <input id="invoice-invoiceNumber" v-model="form.invoiceNumber" name="invoiceNumber" type="text" autocomplete="off" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
              </div>
              <div>
                <label for="invoice-type" class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.type') }}</label>
                <select id="invoice-type" v-model="form.type" name="type" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500">
                  <option value="invoice">{{ enumLabel('invoice_type', 'invoice') }}</option>
                  <option value="proforma">{{ enumLabel('invoice_type', 'proforma') }}</option>
                  <option value="commercial">{{ enumLabel('invoice_type', 'commercial') }}</option>
                  <option value="credit_note">{{ enumLabel('invoice_type', 'credit_note') }}</option>
                </select>
              </div>
              <div>
                <label for="invoice-orderId" class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.order') }}</label>
                <select id="invoice-orderId" v-model="form.orderId" name="orderId" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500">
                  <option value="">{{ t('admin.invoices.select_order') }}</option>
                  <option v-for="order in orders" :key="order.id" :value="order.id">
                    #{{ order.orderNumber || String(order.id).substring(0, 8) }}
                  </option>
                </select>
              </div>
              <div>
                <label for="invoice-currency" class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.currency') }}</label>
                <input id="invoice-currency" v-model="form.currency" name="currency" type="text" autocomplete="off" :placeholder="t('defaults.currency')" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500 uppercase" />
              </div>
              <div>
                <label for="invoice-totalAmount" class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.total_amount') }}</label>
                <input id="invoice-totalAmount" v-model.number="form.totalAmount" name="totalAmount" type="number" step="0.01" autocomplete="off" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
              </div>
              <div>
                <label for="invoice-paidAmount" class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.paid_amount') }}</label>
                <input id="invoice-paidAmount" v-model.number="form.paidAmount" name="paidAmount" type="number" step="0.01" autocomplete="off" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
              </div>
              <div>
                <label for="invoice-invoiceDate" class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.invoice_date') }}</label>
                <input id="invoice-invoiceDate" v-model="form.invoiceDate" name="invoiceDate" type="date" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
              </div>
              <div>
                <label for="invoice-dueDate" class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.due_date') }}</label>
                <input id="invoice-dueDate" v-model="form.dueDate" name="dueDate" type="date" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
              </div>
              <div>
                <label for="invoice-status" class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.status') }}</label>
                <select id="invoice-status" v-model="form.status" name="status" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500">
                  <option value="draft">{{ t('admin.status_options.draft') }}</option>
                  <option value="sent">{{ t('admin.status_options.sent') }}</option>
                  <option value="paid">{{ t('admin.status_options.paid') }}</option>
                  <option value="overdue">{{ t('admin.status_options.overdue') }}</option>
                  <option value="voided">{{ t('admin.invoices.status_voided') }}</option>
                </select>
              </div>
            </div>
            <div>
              <label for="invoice-notes" class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.notes') }}</label>
              <textarea id="invoice-notes" v-model="form.notes" name="notes" rows="2" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"></textarea>
            </div>
            <div v-if="formError" class="text-sm text-red-600">{{ formError }}</div>
            <div class="flex justify-end gap-3 pt-4 border-t border-gray-100">
              <button type="button" @click="closeModal" class="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100">
                {{ t('admin.invoices.cancel') }}
              </button>
              <button type="submit" :disabled="saving" class="px-4 py-2 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 disabled:opacity-50 flex items-center gap-2">
                <Icon v-if="saving" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" aria-hidden="true" />
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
        <button type="button" class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity w-full border-0 cursor-pointer" @click="showPreviewModal = false" :aria-label="t('close')"></button>
        <div class="inline-block w-full transform overflow-hidden rounded-xl bg-white text-left align-bottom shadow-xl sm:my-8 sm:max-w-3xl sm:align-middle">
          <div class="bg-gradient-to-r from-gray-700 to-gray-900 px-6 py-4 flex items-center justify-between">
            <h3 class="text-lg font-semibold text-white">{{ t('admin.invoices.invoice_preview') }}</h3>
            <div class="flex items-center gap-2">
              <button @click="printInvoice" class="px-3 py-1.5 bg-white/10 text-white text-sm rounded-lg hover:bg-white/20 transition-colors flex items-center gap-1">
                <Icon name="heroicons:printer" class="h-4 w-4" aria-hidden="true" />
                {{ t('admin.invoices.print') }}
              </button>
              <button @click="showPreviewModal = false" class="text-gray-300 hover:text-white" :aria-label="t('close')">
                <Icon name="heroicons:x-mark" class="h-5 w-5" aria-hidden="true" />
              </button>
            </div>
          </div>
          <div v-if="previewInvoice" class="p-8 bg-white max-h-[70vh] overflow-y-auto" id="invoice-print-area">
            <!-- Invoice Header -->
            <div class="flex justify-between items-start mb-8">
              <div>
                <h1 class="text-2xl font-bold text-gray-900">{{ t('admin.invoices.invoice_heading') }}</h1>
                <p class="text-gray-500 mt-1">{{ previewInvoice.invoiceNumber }}</p>
                <p class="text-sm text-gray-500 mt-1">
                  {{ previewInvoice.type === 'proforma' ? t('admin.invoices.proforma_heading') : previewInvoice.type === 'credit_note' ? t('admin.invoices.credit_note_heading') : t('admin.invoices.tax_heading') }}
                </p>
              </div>
              <div class="text-right">
                <h2 class="text-xl font-bold text-orange-600">{{ $t('invoices.company_name') }}</h2>
                <p class="text-sm text-gray-500">{{ $t('invoices.company_tagline') }}</p>
              </div>
            </div>

            <!-- Dates & Status -->
            <div class="grid grid-cols-2 gap-8 mb-8">
              <div>
                <p class="text-sm font-medium text-gray-500 mb-1">{{ t('admin.invoices.invoice_date') }}</p>
                <p class="text-gray-900">{{ formatDate(previewInvoice.invoiceDate) }}</p>
              </div>
              <div>
                <p class="text-sm font-medium text-gray-500 mb-1">{{ t('admin.invoices.due_date') }}</p>
                <p class="text-gray-900">{{ formatDate(previewInvoice.dueDate) }}</p>
              </div>
            </div>

            <!-- Bill To -->
            <div class="mb-8">
              <p class="text-sm font-medium text-gray-500 mb-1">{{ t('admin.invoices.bill_to') }}</p>
              <p class="font-semibold text-gray-900">{{ previewInvoice.customerName }}</p>
              <p class="text-sm text-gray-600">{{ previewInvoice.customerEmail }}</p>
            </div>

            <!-- Amount Summary -->
            <div class="border-t border-b border-gray-200 py-6 mb-8">
              <div class="flex justify-between items-center">
                <span class="text-lg font-medium text-gray-700">{{ t('admin.invoices.total_amount') }}</span>
                <span class="text-3xl font-bold text-gray-900">{{ cur(previewInvoice.currency) }} {{ formatNumber(previewInvoice.totalAmount || 0) }}</span>
              </div>
              <div v-if="previewInvoice.paidAmount > 0" class="mt-2 flex justify-between items-center">
                <span class="text-emerald-600">{{ t('admin.invoices.paid_amount') }}</span>
                <span class="text-lg font-semibold text-emerald-600">{{ cur(previewInvoice.currency) }} {{ formatNumber(previewInvoice.paidAmount) }}</span>
              </div>
            </div>

            <!-- Notes -->
            <div v-if="previewInvoice.notes" class="text-sm text-gray-600">
              <p class="font-medium text-gray-700 mb-1">{{ t('admin.invoices.notes') }}</p>
              <p>{{ previewInvoice.notes }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- From Order Modal -->
    <div v-if="showFromOrderModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div class="bg-white rounded-xl shadow-xl w-full max-w-md p-6">
        <h3 class="text-lg font-semibold mb-4">{{ t('admin.invoices.from_order') }}</h3>
        <form @submit.prevent="createFromOrder">
          <label class="block text-sm font-medium text-gray-700">{{ t('admin.invoices.order') }}</label>
          <select v-model="fromOrderId" required class="mt-1 w-full rounded-lg border border-gray-300 px-3 py-2 text-sm">
            <option value="">{{ t('admin.invoices.select_order') }}</option>
            <option v-for="order in orders" :key="order.id" :value="order.id">#{{ order.orderNumber || order.id }}</option>
          </select>
          <p v-if="fromOrderError" class="mt-2 text-sm text-red-600">{{ fromOrderError }}</p>
          <div class="mt-4 flex justify-end gap-2">
            <button type="button" class="px-4 py-2 border rounded-lg text-sm" @click="showFromOrderModal = false">{{ t('admin.invoices.cancel') }}</button>
            <button type="submit" :disabled="fromOrderSaving" class="px-4 py-2 bg-orange-600 text-white rounded-lg text-sm disabled:opacity-50">{{ t('admin.invoices.create_invoice') }}</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { normalizeInvoice, normalizeInvoiceList, isoDatePrefix } from '~/utils/invoiceMapping'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t, te } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, cell, enumLabel, formatNumber, formatDate } = useDisplay()

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
const invoiceStats = ref<any>(null)
const showFromOrderModal = ref(false)
const fromOrderId = ref('')
const fromOrderError = ref('')
const fromOrderSaving = ref(false)

const form = reactive({
  invoiceNumber: '', type: 'invoice', orderId: '', currency: cur(null),
  totalAmount: 0, paidAmount: 0, invoiceDate: '', dueDate: '', status: 'draft', notes: ''
})

const statsCards = computed(() => {
  const s = invoiceStats.value
  if (s) {
    return [
      { label: t('admin.invoices.total'), value: s.total ?? 0, icon: 'heroicons:document-text', color: 'text-orange-600', bgColor: 'bg-orange-50' },
      { label: t('admin.invoices.draft'), value: s.draft ?? 0, icon: 'heroicons:pencil', color: 'text-gray-600', bgColor: 'bg-gray-50' },
      { label: t('admin.invoices.sent'), value: s.sent ?? 0, icon: 'heroicons:paper-airplane', color: 'text-yellow-600', bgColor: 'bg-yellow-50' },
      { label: t('admin.invoices.paid'), value: s.paid ?? 0, icon: 'heroicons:check-circle', color: 'text-emerald-600', bgColor: 'bg-emerald-50' },
    ]
  }
  const all = invoices.value.length
  const draft = invoices.value.filter(i => i.status === 'draft').length
  const sent = invoices.value.filter(i => i.status === 'sent').length
  const paid = invoices.value.filter(i => i.status === 'paid').length
  return [
    { label: t('admin.invoices.total'), value: all, icon: 'heroicons:document-text', color: 'text-orange-600', bgColor: 'bg-orange-50' },
    { label: t('admin.invoices.draft'), value: draft, icon: 'heroicons:pencil', color: 'text-gray-600', bgColor: 'bg-gray-50' },
    { label: t('admin.invoices.sent'), value: sent, icon: 'heroicons:paper-airplane', color: 'text-yellow-600', bgColor: 'bg-yellow-50' },
    { label: t('admin.invoices.paid'), value: paid, icon: 'heroicons:check-circle', color: 'text-emerald-600', bgColor: 'bg-emerald-50' }
  ]
})

const filteredInvoices = computed(() => invoices.value.filter(inv => {
  const matchesSearch = !searchQuery.value ||
    (inv.invoiceNumber || '').toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    (inv.customerName || '').toLowerCase().includes(searchQuery.value.toLowerCase())
  return matchesSearch && (statusFilter.value === 'all' || inv.status === statusFilter.value)
}))

const fetchInvoices = async () => {
  pending.value = true; error.value = ''
  try {
    const res = await api.get<any>('/admin/invoices', { page: page.value, limit: pageSize })
    // Normalize backend fields (invoiceNo→invoiceNumber, createdAt→invoiceDate,
    // derive paidAmount) so the UI can consume one consistent shape.
    invoices.value = normalizeInvoiceList(res.data || [])
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally { pending.value = false }
}

const fetchOrders = async () => {
  try { const res = await api.get<any>('/admin/orders?limit=100'); orders.value = res.data || [] }
  catch { orders.value = [] }
}

const fetchInvoiceStats = async () => {
  try { invoiceStats.value = await api.get<any>('/admin/invoices/stats') }
  catch { invoiceStats.value = null }
}

const openFromOrderModal = () => {
  fromOrderId.value = ''
  fromOrderError.value = ''
  showFromOrderModal.value = true
}

const createFromOrder = async () => {
  fromOrderSaving.value = true
  fromOrderError.value = ''
  try {
    await api.post('/admin/invoices/from-order', { orderId: fromOrderId.value })
    showFromOrderModal.value = false
    await fetchInvoices()
    await fetchInvoiceStats()
  } catch (err: any) {
    fromOrderError.value = err?.message || t('errors.api.save_failed')
  } finally {
    fromOrderSaving.value = false
  }
}

const deleteInvoice = async (invoice: any) => {
  if (!confirm(t('admin.confirm_delete'))) return
  try {
    await api.delete(`/admin/invoices/${invoice.id}`)
    await fetchInvoices()
    await fetchInvoiceStats()
  } catch (err: any) {
    notifyError(err, t('errors.api.delete_failed'))
  }
}

const exportInvoiceDoc = async (invoice: any) => {
  try {
    const { token } = useAuth()
    const config = useRuntimeConfig()
    const base = config.public.apiBase || '/api/v1'
    const resp = await fetch(`${base}/admin/invoices/${invoice.id}/export`, {
      headers: { Authorization: `Bearer ${token.value}` },
    })
    if (!resp.ok) throw new Error(t('errors.api.export_failed'))
    const blob = await resp.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${invoice.invoiceNumber || invoice.id}.docx`
    a.click()
    URL.revokeObjectURL(url)
  } catch (err: any) {
    notifyError(err, t('errors.api.export_failed'))
  }
}

const openCreateModal = () => {
  editingId.value = ''
  const today = new Date().toISOString().split('T')[0]
  const dueDate = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
  Object.assign(form, { invoiceNumber: `INV-${Date.now().toString().slice(-8)}`, type: 'invoice', orderId: '', currency: cur(null), totalAmount: 0, paidAmount: 0, invoiceDate: today, dueDate, status: 'draft', notes: '' })
  formError.value = ''; showModal.value = true
}

const openEditModal = (invoice: any) => {
  editingId.value = invoice.id
  Object.assign(form, { invoiceNumber: invoice.invoiceNumber || '', type: invoice.type || 'invoice', orderId: invoice.orderId || '', currency: cur(invoice.currency), totalAmount: invoice.totalAmount || 0, paidAmount: invoice.paidAmount || 0, invoiceDate: isoDatePrefix(invoice.invoiceDate), dueDate: isoDatePrefix(invoice.dueDate), status: invoice.status || 'draft', notes: invoice.notes || '' })
  formError.value = ''; showModal.value = true
}

const closeModal = () => { showModal.value = false; saving.value = false }

const saveInvoice = async () => {
  saving.value = true; formError.value = ''
  try {
    // Map UI-only field names (invoiceNumber → invoiceNo) and drop fields that
    // do not exist on the backend (paidAmount, invoiceDate). The backend assigns
    // an invoiceNo automatically when omitted, and createdAt is set on insert.
    const { invoiceNumber, paidAmount, invoiceDate, ...rest } = form
    const payload: Record<string, unknown> = { ...rest }
    if (invoiceNumber) payload.invoiceNo = invoiceNumber
    if (editingId.value) { await api.put(`/admin/invoices/${editingId.value}`, payload) }
    else { await api.post('/admin/invoices', payload) }
    closeModal(); await fetchInvoices()
  } catch (err: any) { formError.value = err?.message || t('errors.api.save_failed') }
  finally { saving.value = false }
}

const viewInvoice = (invoice: any) => { previewInvoice.value = normalizeInvoice(invoice); showPreviewModal.value = true }

const sendInvoice = async (invoice: any) => {
  try { await api.post(`/admin/invoices/${invoice.id}/send`, {}); await fetchInvoices() }
  catch (err: any) { notifyError(err, t('errors.api.send_failed')) }
}

const printInvoice = () => {
  window.print()
}

const exportInvoices = () => {
  const statusMap: Record<string, string> = {
    draft: t('enum.invoice_status.draft'), sent: t('enum.invoice_status.sent'),
    paid: t('enum.invoice_status.paid'), overdue: t('enum.invoice_status.overdue'),
    voided: t('enum.invoice_status.voided')
  }
  const headers = t('admin.invoices.csv_headers').split(',')
  const csv = [headers.join(','), ...filteredInvoices.value.map(inv => [inv.invoiceNumber, inv.orderNumber || inv.orderId || '', `"${inv.customerName || ''}"`, inv.totalAmount || 0, inv.paidAmount || 0, statusMap[inv.status] || inv.status || '', inv.invoiceDate || '', inv.dueDate || ''].join(','))].join('\n')
  const blob = new Blob([csv], { type: 'text/csv' }); const url = URL.createObjectURL(blob)
  const a = document.createElement('a'); a.href = url; a.download = `${t('admin.invoices.export_filename')}-${new Date().toISOString().split('T')[0]}.csv`; a.click()
}

const statusBadgeClass = (status: string) => {
  if (status === 'draft') return 'bg-gray-100 text-gray-800'; if (status === 'sent') return 'bg-yellow-100 text-yellow-800'
  if (status === 'paid') return 'bg-emerald-100 text-emerald-800'; if (status === 'overdue') return 'bg-red-100 text-red-800'
  return 'bg-gray-100 text-gray-800'
}
const formatStatus = (status: string) => {
  if (!status) return t('admin.invoices.status_unknown')
  const key = `admin.invoices.status_${String(status).toLowerCase()}`
  return te(key) ? t(key) : t('admin.invoices.status_unknown')
}
const isOverdue = (invoice: any) => { if (!invoice.dueDate || invoice.status === 'paid' || invoice.status === 'voided') return false; return new Date(invoice.dueDate) < new Date() }
const prevPage = () => { if (page.value > 1) { page.value -= 1; fetchInvoices() } }
const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) { page.value += 1; fetchInvoices() } }

onMounted(() => { fetchInvoices(); fetchOrders(); fetchInvoiceStats() })
</script>
