<template>
  <div>
    <div class="mb-6">
      <NuxtLink to="/admin/trades" class="flex items-center text-sm font-medium text-blue-600 hover:text-blue-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('admin.trades.back') }}
      </NuxtLink>
    </div>

    <div v-if="pending" class="text-center py-10 text-gray-500">{{ t('admin.trades.loading_details') }}</div>
    <div v-else-if="error" class="bg-red-50 p-4 rounded-md text-red-700">{{ error }}</div>

    <div v-else-if="trade" class="space-y-6">
      <!-- Trade Info -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 flex justify-between items-center bg-gray-50 border-b border-gray-200">
          <div>
            <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.trades.transaction') }} #{{ trade.reference || trade.id }}</h3>
            <p class="mt-1 text-sm text-gray-500">{{ t('admin.trades.created_on') }} {{ formatDate(trade.createdAt) }}</p>
          </div>
          <span :class="statusBadgeClass(trade.status)" class="inline-flex rounded-full px-3 py-1 text-sm font-semibold">{{ trade.status }}</span>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-4 sm:grid-cols-3 text-sm">
            <div>
              <dt class="text-gray-500">{{ t('admin.trades.customer') }}</dt>
              <dd class="mt-1 font-medium text-gray-900">
                <template v-if="trade.user">{{ trade.user.firstName }} {{ trade.user.lastName }} ({{ trade.user.email }})</template>
                <template v-else>{{ trade.userId }}</template>
              </dd>
            </div>
            <div>
              <dt class="text-gray-500">{{ t('admin.trades.incoterms') }}</dt>
              <dd class="mt-1 font-medium text-gray-900">{{ trade.terms || trade.incoterms || '-' }}</dd>
            </div>
            <div>
              <dt class="text-gray-500">{{ t('admin.trades.total_amount') }}</dt>
              <dd class="mt-1 font-medium text-gray-900">{{ trade.currency || 'USD' }} {{ (trade.totalAmount || 0).toLocaleString() }}</dd>
            </div>
            <div v-if="trade.orderId">
              <dt class="text-gray-500">{{ t('admin.trades.order_id') }}</dt>
              <dd class="mt-1">
                <NuxtLink :to="`/admin/orders/${trade.orderId}`" class="text-blue-600 hover:underline font-medium">{{ trade.orderId }}</NuxtLink>
              </dd>
            </div>
            <div v-if="trade.inquiryId">
              <dt class="text-gray-500">{{ t('admin.trades.inquiry_id') }}</dt>
              <dd class="mt-1">
                <NuxtLink :to="`/admin/inquiries/${trade.inquiryId}`" class="text-blue-600 hover:underline font-medium">{{ trade.inquiryId }}</NuxtLink>
              </dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Status Update -->
      <div class="bg-white shadow sm:rounded-lg">
        <div class="px-4 py-4 border-b border-gray-200 bg-gray-50">
          <h3 class="text-base font-medium text-gray-900">{{ t('admin.trades.update_status') }}</h3>
        </div>
        <div class="px-4 py-4 flex items-end gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('admin.trades.status') }}</label>
            <select v-model="statusInput" class="rounded-md border border-gray-300 px-3 py-2 text-sm">
              <option value="draft">draft</option>
              <option value="pending">pending</option>
              <option value="confirmed">confirmed</option>
              <option value="paid">paid</option>
              <option value="shipped">shipped</option>
              <option value="completed">completed</option>
              <option value="cancelled">cancelled</option>
            </select>
          </div>
          <button @click="updateStatus" :disabled="updatingStatus"
            class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50">
            {{ updatingStatus ? t('admin.trades.updating') : t('admin.trades.update') }}
          </button>
          <p v-if="statusMessage" class="text-sm" :class="statusError ? 'text-red-600' : 'text-green-600'">{{ statusMessage }}</p>
        </div>
      </div>

      <!-- Trade Documents (generic) -->
      <div class="bg-white shadow sm:rounded-lg">
        <div class="px-4 py-4 border-b border-gray-200 bg-gray-50 flex items-center justify-between">
          <h3 class="text-base font-medium text-gray-900">{{ t('admin.trades.documents') }}</h3>
        </div>
        <div class="p-4">
          <div v-if="!documents.length" class="text-gray-500 text-sm">{{ t('admin.trades.no_documents') }}</div>
          <table v-else class="min-w-full divide-y divide-gray-200 text-sm">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.trades.doc_type') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.trades.doc_name') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.trades.doc_status') }}</th>
                <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.trades.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr v-for="doc in documents" :key="doc.id">
                <td class="px-4 py-3 font-medium text-gray-900">{{ doc.type }}</td>
                <td class="px-4 py-3 text-gray-500">{{ doc.docNumber }}</td>
                <td class="px-4 py-3">
                  <span :class="docStatusBadge(doc.status)" class="px-2 py-0.5 text-xs font-semibold rounded-full">{{ doc.status }}</span>
                </td>
                <td class="px-4 py-3">
                  <button v-if="doc.status === 'DRAFT'" @click="confirmDoc(doc.id)"
                    class="text-green-600 hover:text-green-800 text-xs font-medium">{{ t('admin.trades.confirm') }}</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Rich Documents: PI / CI / B/L -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div v-for="rdoc in richDocCards" :key="rdoc.key" class="bg-white shadow sm:rounded-lg">
          <div class="px-4 py-3 border-b border-gray-200 bg-gray-50 flex items-center justify-between">
            <h4 class="text-sm font-semibold text-gray-900">{{ rdoc.label }}</h4>
            <span v-if="rdoc.data" :class="docStatusBadge(rdoc.data.status)" class="px-2 py-0.5 text-xs font-semibold rounded-full">{{ rdoc.data.status }}</span>
          </div>
          <div class="p-4">
            <div v-if="!rdoc.data" class="text-gray-400 text-xs mb-3">{{ t('admin.trades.not_created') }}</div>
            <div v-else class="text-xs text-gray-600 space-y-1 mb-3">
              <div><span class="font-medium">{{ rdoc.numberLabel }}:</span> {{ rdoc.data[rdoc.numberField] }}</div>
              <div v-if="rdoc.data.totalAmount"><span class="font-medium">Amount:</span> {{ rdoc.data.currency }} {{ rdoc.data.totalAmount?.toLocaleString() }}</div>
              <div v-if="rdoc.data.incoterms"><span class="font-medium">Incoterms:</span> {{ rdoc.data.incoterms }}</div>
            </div>
            <button @click="openRichDocModal(rdoc)" class="text-xs text-blue-600 hover:text-blue-800 font-medium">
              {{ rdoc.data ? t('admin.trades.edit') : t('admin.trades.create') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Rich Doc Edit Modal -->
      <div v-if="editingRichDoc" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-40">
        <div class="bg-white rounded-xl shadow-xl w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-semibold text-gray-900">{{ editingRichDoc.label }}</h3>
            <button @click="editingRichDoc = null" class="text-gray-400 hover:text-gray-600">
              <Icon name="heroicons:x-mark" class="h-5 w-5" />
            </button>
          </div>
          <form @submit.prevent="saveRichDoc" class="space-y-3">
            <template v-for="field in editingRichDoc.fields" :key="field.key">
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-0.5">{{ field.label }}</label>
                <input v-model="richDocForm[field.key]" :type="field.type || 'text'"
                  class="w-full border border-gray-300 rounded-md px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500" />
              </div>
            </template>
            <div class="flex gap-3 pt-2">
              <button type="submit" :disabled="savingRichDoc"
                class="px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50">
                {{ savingRichDoc ? t('admin.trades.saving') : t('admin.trades.save') }}
              </button>
              <button type="button" @click="editingRichDoc = null"
                class="px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-md hover:bg-gray-200">
                {{ t('admin.trades.cancel') }}
              </button>
            </div>
            <p v-if="richDocError" class="text-sm text-red-600">{{ richDocError }}</p>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive, computed } from 'vue'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const route = useRoute()
const api = useApi()
const { t } = useI18n()

const trade = ref<any>(null)
const documents = ref<any[]>([])
const pending = ref(true)
const error = ref('')
const statusInput = ref('')
const updatingStatus = ref(false)
const statusMessage = ref('')
const statusError = ref(false)

// Rich documents
const piData = ref<any>(null)
const ciData = ref<any>(null)
const blData = ref<any>(null)
const scData = ref<any>(null)
const plData = ref<any>(null)
const cooData = ref<any>(null)
const hcData = ref<any>(null)
const editingRichDoc = ref<any>(null)
const richDocForm = reactive<Record<string, any>>({})
const savingRichDoc = ref(false)
const richDocError = ref('')

const richDocCards = computed(() => [
  {
    key: 'pi', label: 'Proforma Invoice', data: piData.value,
    numberLabel: 'PI No.', numberField: 'piNumber',
    endpoint: 'proforma-invoice',
    fields: [
      { key: 'buyerName', label: 'Buyer Name' },
      { key: 'sellerName', label: 'Seller Name' },
      { key: 'incoterms', label: 'Incoterms' },
      { key: 'termsOfPayment', label: 'Payment Terms' },
      { key: 'totalAmount', label: 'Total Amount', type: 'number' },
      { key: 'currency', label: 'Currency' },
      { key: 'bankDetails', label: 'Bank Details' },
      { key: 'notes', label: 'Notes' },
      { key: 'status', label: 'Status' },
    ]
  },
  {
    key: 'ci', label: 'Commercial Invoice', data: ciData.value,
    numberLabel: 'CI No.', numberField: 'ciNumber',
    endpoint: 'commercial-invoice',
    fields: [
      { key: 'buyerName', label: 'Buyer Name' },
      { key: 'sellerName', label: 'Seller Name' },
      { key: 'incoterms', label: 'Incoterms' },
      { key: 'termsOfPayment', label: 'Payment Terms' },
      { key: 'totalAmount', label: 'Total Amount', type: 'number' },
      { key: 'currency', label: 'Currency' },
      { key: 'piNumber', label: 'PI Reference' },
      { key: 'bankDetails', label: 'Bank Details' },
      { key: 'status', label: 'Status' },
    ]
  },
  {
    key: 'bl', label: 'Bill of Lading', data: blData.value,
    numberLabel: 'B/L No.', numberField: 'blNumber',
    endpoint: 'bill-of-lading',
    fields: [
      { key: 'shipper', label: 'Shipper' },
      { key: 'consignee', label: 'Consignee' },
      { key: 'notifyParty', label: 'Notify Party' },
      { key: 'carrierName', label: 'Carrier' },
      { key: 'vesselVoyage', label: 'Vessel/Voyage' },
      { key: 'portOfLoading', label: 'Port of Loading' },
      { key: 'portOfDischarge', label: 'Port of Discharge' },
      { key: 'freightTerms', label: 'Freight Terms' },
      { key: 'goodsDescription', label: 'Goods Description' },
      { key: 'grossWeight', label: 'Gross Weight (kg)', type: 'number' },
      { key: 'measurement', label: 'Measurement (CBM)', type: 'number' },
      { key: 'numberOfPackages', label: 'No. of Packages', type: 'number' },
      { key: 'status', label: 'Status' },
    ]
  },
  {
    key: 'sc', label: 'Sales Contract', data: scData.value,
    numberLabel: 'Contract No.', numberField: 'contractNo',
    endpoint: 'sales-contract',
    fields: [
      { key: 'contractNo', label: 'Contract No.' },
      { key: 'buyerName', label: 'Buyer Name' },
      { key: 'sellerName', label: 'Seller Name' },
      { key: 'incoterms', label: 'Incoterms' },
      { key: 'paymentTerms', label: 'Payment Terms' },
      { key: 'totalAmount', label: 'Total Amount', type: 'number' },
      { key: 'currency', label: 'Currency' },
      { key: 'deliveryDate', label: 'Delivery Date', type: 'date' },
      { key: 'specialTerms', label: 'Special Terms' },
      { key: 'status', label: 'Status' },
    ]
  },
  {
    key: 'pl', label: 'Packing List', data: plData.value,
    numberLabel: 'PL No.', numberField: 'plNumber',
    endpoint: 'packing-list',
    fields: [
      { key: 'plNumber', label: 'PL No.' },
      { key: 'exporterName', label: 'Exporter Name' },
      { key: 'importerName', label: 'Importer Name' },
      { key: 'totalPackages', label: 'Total Packages', type: 'number' },
      { key: 'totalGrossWeight', label: 'Total Gross Weight (kg)', type: 'number' },
      { key: 'totalNetWeight', label: 'Total Net Weight (kg)', type: 'number' },
      { key: 'totalVolume', label: 'Total Volume (CBM)', type: 'number' },
      { key: 'shippingMark', label: 'Shipping Mark' },
      { key: 'notes', label: 'Notes' },
    ]
  },
  {
    key: 'coo', label: 'Certificate of Origin', data: cooData.value,
    numberLabel: 'Cert No.', numberField: 'certificateNo',
    endpoint: 'certificate-of-origin',
    fields: [
      { key: 'certificateNo', label: 'Certificate No.' },
      { key: 'exporterName', label: 'Exporter Name' },
      { key: 'importerName', label: 'Importer Name' },
      { key: 'countryOfOrigin', label: 'Country of Origin' },
      { key: 'destinationCountry', label: 'Destination Country' },
      { key: 'hsCode', label: 'HS Code' },
      { key: 'ftaType', label: 'FTA Type (e.g. RCEP, FORM E)' },
      { key: 'goodsDescription', label: 'Goods Description' },
      { key: 'grossWeight', label: 'Gross Weight (kg)', type: 'number' },
      { key: 'issuingAuthority', label: 'Issuing Authority' },
    ]
  },
  {
    key: 'hc', label: 'Health Certificate', data: hcData.value,
    numberLabel: 'Cert No.', numberField: 'certificateNo',
    endpoint: 'health-certificate',
    fields: [
      { key: 'certificateNo', label: 'Certificate No.' },
      { key: 'exporterName', label: 'Exporter Name' },
      { key: 'importerName', label: 'Importer Name' },
      { key: 'productName', label: 'Product Name' },
      { key: 'batchNumber', label: 'Batch Number' },
      { key: 'productionDate', label: 'Production Date', type: 'date' },
      { key: 'expiryDate', label: 'Expiry Date', type: 'date' },
      { key: 'issuingAuthority', label: 'Issuing Authority' },
      { key: 'inspectionResult', label: 'Inspection Result' },
      { key: 'notes', label: 'Notes' },
    ]
  },
])

const fetchTrade = async () => {
  pending.value = true
  error.value = ''
  try {
    trade.value = await api.get<any>(`/admin/trades/${route.params.id}`)
    statusInput.value = trade.value.status || 'DRAFT'
  } catch (err: any) {
    error.value = err?.message || 'Failed to fetch trade'
  } finally {
    pending.value = false
  }
}

const fetchDocuments = async () => {
  try {
    documents.value = await api.get<any[]>(`/admin/trades/${route.params.id}/documents`) || []
  } catch { documents.value = [] }
}

const fetchRichDocs = async () => {
  const tryFetch = async (path: string) => {
    try { return await api.get<any>(`/admin/trades/${route.params.id}/${path}`) } catch { return null }
  }
  const [pi, ci, bl, sc, pl, coo, hc] = await Promise.all([
    tryFetch('proforma-invoice'),
    tryFetch('commercial-invoice'),
    tryFetch('bill-of-lading'),
    tryFetch('sales-contract'),
    tryFetch('packing-list'),
    tryFetch('certificate-of-origin'),
    tryFetch('health-certificate'),
  ])
  piData.value = pi; ciData.value = ci; blData.value = bl
  scData.value = sc; plData.value = pl; cooData.value = coo; hcData.value = hc
}

const updateStatus = async () => {
  updatingStatus.value = true
  statusMessage.value = ''
  statusError.value = false
  try {
    await api.put(`/admin/trades/${route.params.id}/status`, { status: statusInput.value })
    trade.value.status = statusInput.value
    statusMessage.value = t('admin.trades.status_updated')
    setTimeout(() => { statusMessage.value = '' }, 3000)
  } catch (err: any) {
    statusError.value = true
    statusMessage.value = err?.message || 'Failed to update status'
  } finally {
    updatingStatus.value = false
  }
}

const confirmDoc = async (docId: string) => {
  try {
    await api.put(`/admin/trades/${route.params.id}/documents/${docId}`, { status: 'CONFIRMED' })
    await fetchDocuments()
  } catch (err: any) {
    alert(err?.message || 'Failed to confirm document')
  }
}

const openRichDocModal = (rdoc: any) => {
  editingRichDoc.value = rdoc
  richDocError.value = ''
  // Pre-fill form with existing data
  Object.keys(richDocForm).forEach(k => delete richDocForm[k])
  if (rdoc.data) {
    rdoc.fields.forEach((f: any) => { richDocForm[f.key] = rdoc.data[f.key] ?? '' })
  } else {
    rdoc.fields.forEach((f: any) => { richDocForm[f.key] = '' })
  }
}

const saveRichDoc = async () => {
  if (!editingRichDoc.value) return
  savingRichDoc.value = true
  richDocError.value = ''
  const { endpoint, data } = editingRichDoc.value
  try {
    const method = data ? 'put' : 'post'
    const result = await api[method]<any>(`/admin/trades/${route.params.id}/${endpoint}`, { ...richDocForm })
    // Update local data
    if (endpoint === 'proforma-invoice') piData.value = result
    else if (endpoint === 'commercial-invoice') ciData.value = result
    else if (endpoint === 'bill-of-lading') blData.value = result
    else if (endpoint === 'sales-contract') scData.value = result
    else if (endpoint === 'packing-list') plData.value = result
    else if (endpoint === 'certificate-of-origin') cooData.value = result
    else if (endpoint === 'health-certificate') hcData.value = result
    editingRichDoc.value = null
  } catch (err: any) {
    richDocError.value = err?.message || 'Failed to save'
  } finally {
    savingRichDoc.value = false
  }
}

const formatDate = (d: string) => d ? new Date(d).toLocaleString() : '-'

const statusBadgeClass = (status: string) => {
  const map: Record<string, string> = {
    DRAFT: 'bg-gray-100 text-gray-800', PENDING: 'bg-yellow-100 text-yellow-800',
    CONFIRMED: 'bg-blue-100 text-blue-800', PAID: 'bg-indigo-100 text-indigo-800',
    SHIPPED: 'bg-purple-100 text-purple-800', COMPLETED: 'bg-green-100 text-green-800',
    CANCELLED: 'bg-red-100 text-red-800',
  }
  return map[status] || 'bg-gray-100 text-gray-800'
}

const docStatusBadge = (status: string) => {
  if (status === 'CONFIRMED' || status === 'ISSUED' || status === 'PAID') return 'bg-green-100 text-green-800'
  if (status === 'DRAFT') return 'bg-gray-100 text-gray-800'
  return 'bg-yellow-100 text-yellow-800'
}

onMounted(() => {
  fetchTrade()
  fetchDocuments()
  fetchRichDocs()
})
</script>
