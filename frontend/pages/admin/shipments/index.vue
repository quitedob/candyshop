<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('admin.shipments.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ t('admin.shipments.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0 flex items-center gap-3">
        <button @click="exportShipments" class="inline-flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors">
          <Icon name="heroicons:arrow-down-tray" class="h-4 w-4" aria-hidden="true" />
          {{ t('admin.shipments.export') }}
        </button>
        <button @click="openCreateModal" class="inline-flex items-center gap-2 px-4 py-2 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 transition-colors">
          <Icon name="heroicons:plus" class="h-4 w-4" aria-hidden="true" />
          {{ t('admin.shipments.new_shipment') }}
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
            <label for="shipment-search" class="sr-only">{{ t('admin.shipments.search') }}</label>
            <Icon name="heroicons:magnifying-glass" class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" aria-hidden="true" />
            <input id="shipment-search" v-model="searchQuery" name="search" type="text" :placeholder="t('admin.shipments.search')" class="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
          </div>
        </div>
        <label for="shipment-filter-status" class="sr-only">{{ t('admin.shipments.filter_status') }}</label>
        <select id="shipment-filter-status" v-model="statusFilter" name="statusFilter" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500">
          <option value="all">{{ t('admin.shipments.filter_all') }}</option>
          <option value="pending">{{ t('admin.shipments.status_pending') }}</option>
          <option value="in_transit">{{ t('admin.shipments.status_in_transit') }}</option>
          <option value="delivered">{{ t('admin.shipments.status_delivered') }}</option>
          <option value="exception">{{ t('admin.shipments.status_exception') }}</option>
        </select>
        <label for="shipment-dateFrom" class="sr-only">{{ t('admin.shipments.date_from') }}</label>
        <input id="shipment-dateFrom" v-model="dateFrom" name="dateFrom" type="date" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
        <label for="shipment-dateTo" class="sr-only">{{ t('admin.shipments.date_to') }}</label>
        <input id="shipment-dateTo" v-model="dateTo" name="dateTo" type="date" class="px-4 py-2 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
      </div>
    </div>

    <!-- Shipments Table -->
    <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.shipments.col_tracking') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.shipments.col_order') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.shipments.col_carrier') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.shipments.col_destination') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.shipments.col_eta') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.shipments.col_status') }}</th>
              <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.shipments.col_actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white">
            <tr v-if="pending">
              <td colspan="7" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.shipments.loading') }}</td>
            </tr>
            <tr v-else-if="error">
              <td colspan="7" class="px-6 py-10 text-center text-sm text-red-600">{{ error }}</td>
            </tr>
            <tr v-else-if="filteredShipments.length === 0">
              <td colspan="7" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.shipments.no_data') }}</td>
            </tr>
            <tr v-else v-for="shipment in filteredShipments" :key="shipment.id" class="hover:bg-gray-50 transition-colors">
              <td class="px-6 py-4">
                <div>
                  <span class="font-mono font-medium text-gray-900">{{ shipment.trackingNumber }}</span>
                  <p class="text-xs text-gray-500 mt-0.5">{{ shipment.trackingProvider || 'Standard' }}</p>
                </div>
              </td>
              <td class="px-6 py-4 text-sm">
                <NuxtLink :to="localePath(`/admin/orders/${shipment.orderId}`)" class="text-orange-600 hover:text-orange-900">
                  #{{ shipment.orderNumber || shipment.orderId?.substring(0, 8) }}
                </NuxtLink>
              </td>
              <td class="px-6 py-4 text-sm text-gray-600">{{ shipment.carrier || '-' }}</td>
              <td class="px-6 py-4 text-sm text-gray-600">
                <div>{{ shipment.destination?.city }}, {{ shipment.destination?.country }}</div>
                <div class="text-xs text-gray-400">{{ shipment.destination?.recipientName }}</div>
              </td>
              <td class="px-6 py-4 text-sm text-gray-600">
                <div>{{ formatDate(shipment.estimatedDelivery) }}</div>
                <div v-if="shipment.actualDelivery" class="text-xs text-emerald-600">{{ t('admin.shipments.delivered_label') }}: {{ formatDate(shipment.actualDelivery) }}</div>
              </td>
              <td class="px-6 py-4">
                <span :class="statusBadgeClass(shipment.status)" class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold">
                  {{ formatStatus(shipment.status) }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm">
                <button @click="viewDetails(shipment)" class="text-orange-600 hover:text-orange-900 mr-3">{{ t('admin.shipments.view') }}</button>
                <button @click="openEditModal(shipment)" class="text-gray-600 hover:text-gray-900">{{ t('admin.shipments.edit') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="pagination" class="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
        <div class="text-sm text-gray-600">
          {{ t('admin.shipments.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
        </div>
        <div class="flex items-center gap-2">
          <button @click="prevPage" :disabled="page <= 1" class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed">
            {{ t('admin.shipments.previous') }}
          </button>
          <button @click="nextPage" :disabled="page >= pagination.totalPages" class="px-3 py-1.5 border border-gray-200 rounded-lg text-sm hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed">
            {{ t('admin.shipments.next') }}
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
            <h3 class="text-lg font-semibold text-white">{{ editingId ? t('admin.shipments.edit_shipment') : t('admin.shipments.create_shipment') }}</h3>
          </div>
          <form @submit.prevent="saveShipment" class="p-6 space-y-4">
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label for="shipment-orderId" class="block text-sm font-medium text-gray-700">{{ t('admin.shipments.order') }}</label>
                <select id="shipment-orderId" v-model="form.orderId" name="orderId" required class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500">
                  <option value="">{{ t('admin.shipments.select_order') }}</option>
                  <option v-for="order in orders" :key="order.id" :value="order.id">
                    #{{ order.orderNumber || String(order.id).substring(0, 8) }} - {{ order.user?.firstName }} {{ order.user?.lastName }}
                  </option>
                </select>
              </div>
              <div>
                <label for="shipment-carrier" class="block text-sm font-medium text-gray-700">{{ t('admin.shipments.carrier') }}</label>
                <select id="shipment-carrier" v-model="form.carrier" name="carrier" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500">
                  <option value="">{{ enumLabel('carrier', 'standard', 'standard') }}</option>
                  <option value="dhl">{{ enumLabel('carrier', 'dhl') }}</option>
                  <option value="fedex">{{ enumLabel('carrier', 'fedex') }}</option>
                  <option value="ups">{{ enumLabel('carrier', 'ups') }}</option>
                  <option value="usps">{{ enumLabel('carrier', 'usps') }}</option>
                  <option value="ems">{{ enumLabel('carrier', 'ems') }}</option>
                  <option value="sea">{{ enumLabel('carrier', 'sea') }}</option>
                  <option value="air">{{ enumLabel('carrier', 'air') }}</option>
                </select>
              </div>
              <div>
                <label for="shipment-trackingNumber" class="block text-sm font-medium text-gray-700">{{ t('admin.shipments.tracking_number') }}</label>
                <input id="shipment-trackingNumber" v-model="form.trackingNumber" name="trackingNumber" type="text" autocomplete="off" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
              </div>
              <div>
                <label for="shipment-status" class="block text-sm font-medium text-gray-700">{{ t('admin.shipments.status') }}</label>
                <select id="shipment-status" v-model="form.status" name="status" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500">
                  <option value="pending">{{ enumLabel('shipment_status', 'pending') }}</option>
                  <option value="in_transit">{{ enumLabel('shipment_status', 'in_transit') }}</option>
                  <option value="delivered">{{ enumLabel('shipment_status', 'delivered') }}</option>
                  <option value="exception">{{ enumLabel('shipment_status', 'exception') }}</option>
                </select>
              </div>
              <div>
                <label for="shipment-estimatedDelivery" class="block text-sm font-medium text-gray-700">{{ t('admin.shipments.estimated_delivery') }}</label>
                <input id="shipment-estimatedDelivery" v-model="form.estimatedDelivery" name="estimatedDelivery" type="date" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
              </div>
              <div>
                <label for="shipment-actualDelivery" class="block text-sm font-medium text-gray-700">{{ t('admin.shipments.actual_delivery') }}</label>
                <input id="shipment-actualDelivery" v-model="form.actualDelivery" name="actualDelivery" type="date" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
              </div>
            </div>
            <div>
              <label for="shipment-notes" class="block text-sm font-medium text-gray-700">{{ t('admin.shipments.notes') }}</label>
              <textarea id="shipment-notes" v-model="form.notes" name="notes" rows="2" class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500"></textarea>
            </div>
            <div v-if="formError" class="text-sm text-red-600">{{ formError }}</div>
            <div class="flex justify-end gap-3 pt-4 border-t border-gray-100">
              <button type="button" @click="closeModal" class="px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100">
                {{ t('admin.shipments.cancel') }}
              </button>
              <button type="submit" :disabled="saving" class="px-4 py-2 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 disabled:opacity-50 flex items-center gap-2">
                <Icon v-if="saving" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" aria-hidden="true" />
                {{ t('admin.shipments.save') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Details Modal -->
    <div v-if="showDetailsModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <button type="button" class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity w-full border-0 cursor-pointer" @click="showDetailsModal = false" :aria-label="t('close')"></button>
        <div class="inline-block w-full transform overflow-hidden rounded-xl bg-white text-left align-bottom shadow-xl sm:my-8 sm:max-w-3xl sm:align-middle">
          <div class="bg-gradient-to-r from-gray-700 to-gray-900 px-6 py-4 flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-white">{{ t('admin.shipments.shipment_details') }}</h3>
              <p class="text-sm text-gray-300 mt-0.5">{{ selectedShipment?.trackingNumber }}</p>
            </div>
            <button @click="showDetailsModal = false" class="text-gray-300 hover:text-white" :aria-label="t('close')">
              <Icon name="heroicons:x-mark" class="h-5 w-5" aria-hidden="true" />
            </button>
          </div>
          <div v-if="selectedShipment" class="p-6">
            <!-- Actions -->
            <div class="mb-6 flex flex-wrap gap-2">
              <button
                v-if="['pending', 'picked', 'packed'].includes(selectedShipment.rawStatus || selectedShipment.status)"
                type="button"
                class="px-3 py-1.5 bg-orange-600 text-white text-sm rounded-lg hover:bg-orange-700 disabled:opacity-50"
                :disabled="logisticsActing"
                @click="dispatchShipment"
              >
                {{ t('admin.shipments.dispatch') }}
              </button>
              <button
                v-if="['pending', 'in_transit', 'shipped'].includes(selectedShipment.status) || selectedShipment.rawStatus === 'shipped'"
                type="button"
                class="px-3 py-1.5 bg-emerald-600 text-white text-sm rounded-lg hover:bg-emerald-700 disabled:opacity-50"
                :disabled="logisticsActing"
                @click="confirmDelivery"
              >
                {{ t('admin.shipments.confirm_delivery') }}
              </button>
            </div>
            <p v-if="logisticsMessage" class="mb-4 text-sm" :class="logisticsError ? 'text-red-600' : 'text-green-600'">{{ logisticsMessage }}</p>

            <!-- Timeline -->
            <div class="mb-6">
              <h4 class="text-sm font-semibold text-gray-900 mb-4">{{ t('admin.shipments.tracking_history') }}</h4>
              <div v-if="timelineLoading" class="text-sm text-gray-500">{{ t('admin.shipments.loading_timeline') }}</div>
              <div v-else class="relative">
                <div class="absolute left-4 top-0 bottom-0 w-0.5 bg-gray-200" />
                <div class="space-y-6">
                  <div v-for="(event, idx) in timelineEvents" :key="idx" class="relative flex items-start gap-4 pl-10">
                    <div class="absolute left-2.5 w-3 h-3 rounded-full bg-orange-600 border-2 border-white" />
                    <div>
                      <p class="font-medium text-gray-900">{{ event.eventType || event.status || '—' }}</p>
                      <p class="text-sm text-gray-500">{{ event.location || event.description || '' }}</p>
                      <p class="text-xs text-gray-400 mt-1">{{ formatDate(event.eventTime || event.timestamp, { dateStyle: 'medium', timeStyle: 'short' }) }}</p>
                    </div>
                  </div>
                  <p v-if="!timelineEvents.length" class="text-sm text-gray-500 pl-10">{{ t('admin.shipments.no_data') }}</p>
                </div>
              </div>
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
const { enumLabel, formatNumber, formatDate } = useDisplay()
const localePath = useLocalePath()

const shipments = ref<any[]>([])
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
const showDetailsModal = ref(false)
const selectedShipment = ref<any>(null)
const timelineEvents = ref<any[]>([])
const timelineLoading = ref(false)
const logisticsActing = ref(false)
const logisticsMessage = ref('')
const logisticsError = ref(false)

const form = reactive({
  orderId: '', carrier: '', trackingNumber: '', status: 'pending',
  estimatedDelivery: '', actualDelivery: '', notes: ''
})

const statsCards = computed(() => {
  const all = shipments.value.length
  const pend = shipments.value.filter(s => s.status === 'pending').length
  const inTransit = shipments.value.filter(s => s.status === 'in_transit').length
  const delivered = shipments.value.filter(s => s.status === 'delivered').length
  return [
    { label: t('admin.shipments.total'), value: all, icon: 'heroicons:truck', color: 'text-orange-600', bgColor: 'bg-orange-50' },
    { label: t('admin.shipments.pending'), value: pend, icon: 'heroicons:clock', color: 'text-yellow-600', bgColor: 'bg-yellow-50' },
    { label: t('admin.shipments.in_transit'), value: inTransit, icon: 'heroicons:arrows-right-left', color: 'text-orange-600', bgColor: 'bg-amber-50' },
    { label: t('admin.shipments.delivered'), value: delivered, icon: 'heroicons:check-circle', color: 'text-emerald-600', bgColor: 'bg-emerald-50' }
  ]
})

const filteredShipments = computed(() => shipments.value.filter(s => {
  const matchesSearch = !searchQuery.value ||
    (s.trackingNumber || '').toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    (s.orderNumber || '').toLowerCase().includes(searchQuery.value.toLowerCase())
  const matchesStatus = statusFilter.value === 'all' || s.status === statusFilter.value
  return matchesSearch && matchesStatus
}))

const mapFulfillmentStatus = (status: string) => {
  if (status === 'shipped') return 'in_transit'
  if (status === 'cancelled') return 'exception'
  if (status === 'picked' || status === 'packed') return 'pending'
  return status
}

const apiStatusFromFilter = (filter: string) => {
  if (filter === 'in_transit') return 'shipped'
  if (filter === 'exception') return 'cancelled'
  return filter
}

const normalizeFulfillment = (row: any) => ({
  id: row.id,
  trackingNumber: row.trackingNumber || '',
  carrier: row.carrier || '',
  orderId: row.orderId,
  orderNumber: row.orderNumber,
  status: mapFulfillmentStatus(row.status || 'pending'),
  rawStatus: row.status,
  estimatedDelivery: row.shippedAt,
  actualDelivery: row.deliveredAt,
  notes: row.notes || '',
  destination: null,
})

const fetchShipments = async () => {
  pending.value = true; error.value = ''
  try {
    const params: Record<string, any> = { page: page.value, limit: pageSize }
    if (statusFilter.value !== 'all') {
      params.status = apiStatusFromFilter(statusFilter.value)
    }
    const res = await api.get<any>('/admin/fulfillments', params)
    shipments.value = (res.data || []).map(normalizeFulfillment)
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally { pending.value = false }
}

const fetchOrders = async () => {
  try {
    const res = await api.get<any>('/admin/orders?limit=100')
    orders.value = res.data || []
  } catch { orders.value = [] }
}

const openCreateModal = () => {
  editingId.value = ''
  Object.assign(form, { orderId: '', carrier: '', trackingNumber: '', status: 'pending', estimatedDelivery: '', actualDelivery: '', notes: '' })
  formError.value = ''; showModal.value = true
}

const openEditModal = (shipment: any) => {
  editingId.value = shipment.id
  Object.assign(form, {
    orderId: shipment.orderId || '', carrier: shipment.carrier || '',
    trackingNumber: shipment.trackingNumber || '', status: shipment.status || 'pending',
    estimatedDelivery: shipment.estimatedDelivery ? shipment.estimatedDelivery.split('T')[0] : '',
    actualDelivery: shipment.actualDelivery ? shipment.actualDelivery.split('T')[0] : '',
    notes: shipment.notes || ''
  })
  formError.value = ''; showModal.value = true
}

const closeModal = () => { showModal.value = false; saving.value = false }

const saveShipment = async () => {
  if (!editingId.value) {
    formError.value = t('admin.shipments.create_from_order')
    return
  }
  saving.value = true; formError.value = ''
  try {
    await api.put(`/admin/fulfillments/${editingId.value}/ship`, {
      trackingNumber: form.trackingNumber || `TRK-${editingId.value}`,
      carrier: form.carrier,
    })
    closeModal(); await fetchShipments()
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.save_failed')
  } finally { saving.value = false }
}

const viewDetails = async (shipment: any) => {
  selectedShipment.value = shipment
  showDetailsModal.value = true
  logisticsMessage.value = ''
  logisticsError.value = false
  timelineEvents.value = []
  timelineLoading.value = false
}

const dispatchShipment = async () => {
  if (!selectedShipment.value?.id) return
  logisticsActing.value = true
  logisticsMessage.value = ''
  logisticsError.value = false
  try {
    const tracking = selectedShipment.value.trackingNumber || `TRK-${selectedShipment.value.id}`
    await api.put(`/admin/fulfillments/${selectedShipment.value.id}/ship`, {
      trackingNumber: tracking,
      carrier: selectedShipment.value.carrier || 'standard',
    })
    logisticsMessage.value = t('admin.shipments.dispatch_success')
    await fetchShipments()
    selectedShipment.value = shipments.value.find(s => s.id === selectedShipment.value.id) || selectedShipment.value
  } catch (err: any) {
    logisticsError.value = true
    logisticsMessage.value = err?.message || t('errors.api.action_failed')
  } finally {
    logisticsActing.value = false
  }
}

const confirmDelivery = async () => {
  if (!selectedShipment.value?.id) return
  logisticsActing.value = true
  logisticsMessage.value = ''
  logisticsError.value = false
  try {
    await api.put(`/admin/fulfillments/${selectedShipment.value.id}/deliver`, {})
    logisticsMessage.value = t('admin.shipments.delivery_success')
    await fetchShipments()
    selectedShipment.value = shipments.value.find(s => s.id === selectedShipment.value.id) || selectedShipment.value
  } catch (err: any) {
    logisticsError.value = true
    logisticsMessage.value = err?.message || t('errors.api.action_failed')
  } finally {
    logisticsActing.value = false
  }
}

const exportShipments = () => {
  const shipmentStatusMap: Record<string, string> = {
    pending: t('enum.shipment_status.pending'), in_transit: t('enum.shipment_status.in_transit'),
    delivered: t('enum.shipment_status.delivered')
  }
  const headers = t('admin.shipments.csv_headers').split(',')
  const defaultCarrier = t('admin.shipments.carrier_default')
  const csv = [
    headers.join(','),
    ...filteredShipments.value.map(s => [
      s.trackingNumber, s.orderNumber || s.orderId, s.carrier || defaultCarrier,
      `"${s.destination?.city || ''}, ${s.destination?.country || ''}"`, shipmentStatusMap[s.status] || s.status || '',
      s.estimatedDelivery || '', s.actualDelivery || ''
    ].join(','))
  ].join('\n')
  const blob = new Blob([csv], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a'); a.href = url
  a.download = `${t('admin.shipments.export_filename')}-${new Date().toISOString().split('T')[0]}.csv`; a.click()
}

const statusBadgeClass = (status: string) => {
  if (status === 'pending') return 'bg-yellow-100 text-yellow-800'
  if (status === 'in_transit') return 'bg-amber-100 text-amber-800'
  if (status === 'delivered') return 'bg-emerald-100 text-emerald-800'
  if (status === 'exception') return 'bg-red-100 text-red-800'
  return 'bg-gray-100 text-gray-800'
}

const formatStatus = (status: string) => status ? enumLabel('shipment_status', status) : t('enum.order_status.unknown')
const prevPage = () => { if (page.value > 1) { page.value -= 1; fetchShipments() } }
const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) { page.value += 1; fetchShipments() } }

onMounted(() => { fetchShipments(); fetchOrders() })
</script>
