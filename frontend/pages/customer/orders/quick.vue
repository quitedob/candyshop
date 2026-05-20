<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('customer.quick_order.title') }}</h1>
        <p class="text-gray-600 text-sm mt-1">{{ t('customer.quick_order.description') }}</p>
      </div>
      <NuxtLink :to="localePath('/customer/orders')" class="inline-flex items-center text-sm font-medium text-orange-600 hover:text-orange-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('customer.common.back_to_orders') }}
      </NuxtLink>
    </div>

    <!-- Tab navigation -->
    <div class="flex gap-1 rounded-lg bg-gray-100 p-1 w-fit">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        :class="['rounded-md px-4 py-2 text-sm font-medium transition-colors', activeTab === tab.key ? 'bg-white text-orange-600 shadow-sm' : 'text-gray-600 hover:text-gray-900']"
        @click="activeTab = tab.key"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Manual Entry Tab -->
    <div v-if="activeTab === 'manual'" class="rounded-lg border border-gray-200 bg-white p-6">
      <h3 class="text-sm font-semibold text-gray-900 mb-4">{{ t('customer.quick_order.manual_entry') }}</h3>
      <div class="space-y-3">
        <div
          v-for="(row, idx) in manualRows"
          :key="idx"
          class="flex gap-3 items-start"
        >
          <input
            v-model="row.productId"
            type="text"
            :placeholder="t('customer.quick_order.sku_placeholder')"
            class="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-orange-400 focus:outline-none focus:ring-1 focus:ring-orange-100"
          />
          <input
            v-model.number="row.quantity"
            type="number"
            min="1"
            :placeholder="t('customer.quick_order.qty_placeholder')"
            class="w-24 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-orange-400 focus:outline-none focus:ring-1 focus:ring-orange-100"
          />
          <button
            class="inline-flex items-center rounded-lg p-2 text-gray-400 hover:text-red-500 transition-colors"
            :aria-label="t('customer.quick_order.remove_row')"
            @click="manualRows.splice(idx, 1)"
          >
            <Icon name="heroicons:x-mark" class="h-4 w-4" />
          </button>
        </div>
        <button class="inline-flex items-center gap-1 text-sm font-medium text-orange-600 hover:text-orange-500" @click="addManualRow">
          <Icon name="heroicons:plus" class="h-4 w-4" />
          {{ t('customer.quick_order.add_row') }}
        </button>
      </div>
      <button
        :disabled="!validManualRows.length || submitting"
        class="mt-6 inline-flex items-center gap-2 rounded-lg bg-orange-500 px-5 py-2.5 text-sm font-medium text-white hover:bg-orange-600 disabled:opacity-50 transition-colors"
        @click="submitManualOrder"
      >
        <Icon name="heroicons:shopping-cart" class="h-4 w-4" />
        {{ submitting ? t('customer.quick_order.creating') : t('customer.quick_order.create_order') }}
      </button>
      <p v-if="submitError" class="mt-2 text-sm text-red-600">{{ submitError }}</p>
      <p v-if="submitSuccess" class="mt-2 text-sm text-green-700">{{ submitSuccess }}</p>
    </div>

    <!-- CSV Upload Tab -->
    <div v-if="activeTab === 'csv'" class="rounded-lg border border-gray-200 bg-white p-6">
      <h3 class="text-sm font-semibold text-gray-900 mb-4">{{ t('customer.quick_order.csv_upload') }}</h3>
      <p class="text-xs text-gray-500 mb-4">{{ t('customer.quick_order.csv_format_hint') }}</p>
      <div class="flex items-center gap-4">
        <label class="inline-flex items-center gap-2 rounded-lg border border-gray-300 bg-white px-4 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-50 cursor-pointer transition-colors">
          <Icon name="heroicons:document-arrow-up" class="h-4 w-4 text-orange-500" />
          {{ csvFileName || t('customer.quick_order.choose_file') }}
          <input type="file" accept=".csv" class="hidden" @change="onCsvFileChange" />
        </label>
        <button
          :disabled="!csvFile || submitting"
          class="inline-flex items-center gap-2 rounded-lg bg-orange-500 px-5 py-2.5 text-sm font-medium text-white hover:bg-orange-600 disabled:opacity-50 transition-colors"
          @click="submitCsvOrder"
        >
          <Icon name="heroicons:arrow-up-tray" class="h-4 w-4" />
          {{ submitting ? t('customer.quick_order.uploading') : t('customer.quick_order.upload_csv') }}
        </button>
      </div>
      <p v-if="csvPreview.length" class="mt-4 text-xs text-gray-600">
        {{ t('customer.quick_order.rows_found', { count: csvPreview.length }) }}
      </p>
    </div>

    <!-- Requisition Lists Tab -->
    <div v-if="activeTab === 'requisition'" class="rounded-lg border border-gray-200 bg-white p-6">
      <h3 class="text-sm font-semibold text-gray-900 mb-4">{{ t('customer.quick_order.requisition_lists') }}</h3>
      <div v-if="requisitionLists.length === 0" class="text-sm text-gray-400 text-center py-6">
        {{ t('customer.quick_order.no_requisition_lists') }}
      </div>
      <div v-for="list in requisitionLists" :key="list.id" class="flex items-center justify-between border-b border-gray-100 py-3 last:border-0">
        <div>
          <p class="text-sm font-medium text-gray-900">{{ list.name }}</p>
          <p v-if="list.notes" class="text-xs text-gray-500">{{ list.notes }}</p>
        </div>
        <button
          class="inline-flex items-center gap-1 rounded-lg bg-orange-50 px-3 py-1.5 text-xs font-medium text-orange-700 hover:bg-orange-100 transition-colors"
          @click="convertRequisitionToOrder(list.id)"
        >
          <Icon name="heroicons:arrow-right-circle" class="h-3.5 w-3.5" />
          {{ t('customer.quick_order.create_from_list') }}
        </button>
      </div>
    </div>

    <!-- Reorder from History Tab -->
    <div v-if="activeTab === 'reorder'" class="rounded-lg border border-gray-200 bg-white p-6">
      <h3 class="text-sm font-semibold text-gray-900 mb-4">{{ t('customer.quick_order.reorder_from_history') }}</h3>
      <div v-if="orderHistory.length === 0" class="text-sm text-gray-400 text-center py-6">
        {{ t('customer.quick_order.no_order_history') }}
      </div>
      <div v-for="order in orderHistory" :key="order.id" class="flex items-center justify-between border-b border-gray-100 py-3 last:border-0">
        <div>
          <p class="text-sm font-medium text-gray-900">#{{ order.orderNumber }}</p>
          <p class="text-xs text-gray-500">
            {{ formatDate(order.createdAt) }} &middot;
            {{ order.items?.length || 0 }} items &middot;
            {{ formatCurrency(order.totalAmount) }}
          </p>
        </div>
        <button
          class="inline-flex items-center gap-1 rounded-lg bg-orange-50 px-3 py-1.5 text-xs font-medium text-orange-700 hover:bg-orange-100 transition-colors"
          @click="reorderFromHistory(order.id)"
        >
          <Icon name="heroicons:arrow-path" class="h-3.5 w-3.5" />
          {{ t('customer.quick_order.reorder') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'customer', middleware: ['auth'] })
const { t } = useI18n()
const localePath = useLocalePath()
const api = useApi()
const router = useRouter()

interface ManualRow {
  productId: string
  quantity: number | null
}

const activeTab = ref('manual')
const submitting = ref(false)
const submitError = ref('')
const submitSuccess = ref('')
const manualRows = ref<ManualRow[]>([{ productId: '', quantity: null }])
const csvFile = ref<File | null>(null)
const csvFileName = ref('')
const csvPreview = ref<any[]>([])
const requisitionLists = ref<any[]>([])
const orderHistory = ref<any[]>([])

const tabs = [
  { key: 'manual', label: t('customer.quick_order.tab_manual') },
  { key: 'csv', label: t('customer.quick_order.tab_csv') },
  { key: 'requisition', label: t('customer.quick_order.tab_requisition') },
  { key: 'reorder', label: t('customer.quick_order.tab_reorder') },
]

const validManualRows = computed(() =>
  manualRows.value.filter((r) => r.productId.trim() && (r.quantity ?? 0) > 0)
)

function addManualRow() {
  manualRows.value.push({ productId: '', quantity: null })
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString()
}

function formatCurrency(amount: number): string {
  if (amount == null) return '—'
  return `$${amount.toFixed(2)}`
}

async function submitManualOrder() {
  const items = validManualRows.value.map((r) => ({
    productId: r.productId.trim(),
    quantity: r.quantity!,
  }))
  if (!items.length) return

  submitting.value = true
  submitError.value = ''
  submitSuccess.value = ''

  try {
    const order = await api.post('/user/orders', { items })
    submitSuccess.value = t('customer.quick_order.order_created')
    setTimeout(() => router.push({ path: localePath(`/customer/orders/${order.id}`) }), 1500)
  } catch (err: any) {
    submitError.value = err?.message || t('errors.unknown')
  } finally {
    submitting.value = false
  }
}

function onCsvFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  csvFile.value = file
  csvFileName.value = file.name
  // Preview first few rows
  const reader = new FileReader()
  reader.onload = (ev) => {
    const text = ev.target?.result as string
    const lines = text.split('\n').filter((l) => l.trim())
    csvPreview.value = lines.slice(1, 6).map((l) => {
      const [productId, quantity] = l.split(',')
      return { productId: productId?.trim(), quantity: parseInt(quantity?.trim()) || 0 }
    })
  }
  reader.readAsText(file)
}

async function submitCsvOrder() {
  if (!csvFile.value) return

  submitting.value = true
  submitError.value = ''
  submitSuccess.value = ''

  try {
    const formData = new FormData()
    formData.append('file', csvFile.value)
    const order = await api.post('/user/orders/bulk', formData)
    submitSuccess.value = t('customer.quick_order.order_created')
    setTimeout(() => router.push({ path: localePath(`/customer/orders/${order.id}`) }), 1500)
  } catch (err: any) {
    submitError.value = err?.message || t('errors.unknown')
  } finally {
    submitting.value = false
  }
}

async function loadRequisitionLists() {
  try {
    requisitionLists.value = await api.get('/user/requisition-lists')
  } catch {
    // silent
  }
}

async function loadOrderHistory() {
  try {
    const res = await api.get('/user/orders', { limit: 10 })
    orderHistory.value = res?.data || []
  } catch {
    // silent
  }
}

async function convertRequisitionToOrder(listId: string) {
  submitting.value = true
  submitError.value = ''
  try {
    const order = await api.post(`/user/requisition-lists/${listId}/convert-to-order`)
    router.push({ path: localePath(`/customer/orders/${order.id}`) })
  } catch (err: any) {
    submitError.value = err?.message || t('errors.unknown')
  } finally {
    submitting.value = false
  }
}

async function reorderFromHistory(orderId: string) {
  submitting.value = true
  submitError.value = ''
  try {
    const order = await api.post(`/user/orders/${orderId}/reorder`)
    router.push({ path: localePath(`/customer/orders/${order.id}`) })
  } catch (err: any) {
    submitError.value = err?.message || t('errors.unknown')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadRequisitionLists()
  loadOrderHistory()
})
</script>
