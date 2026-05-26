<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('admin.trade_shipments.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ t('admin.trade_shipments.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0">
        <button @click="openCreateModal" class="inline-flex items-center gap-2 px-4 py-2 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 transition-colors">
          <Icon name="heroicons:plus" class="h-4 w-4" aria-hidden="true" />
          {{ t('admin.trade_shipments.new_shipment') }}
        </button>
      </div>
    </div>

    <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 p-4">
      <select v-model="statusFilter" class="px-4 py-2 border border-gray-200 rounded-lg text-sm" @change="page = 1; fetchShipments()">
        <option value="all">{{ t('admin.shipments.filter_all') }}</option>
        <option value="pending">{{ t('admin.shipments.status_pending') }}</option>
        <option value="in_transit">{{ t('admin.shipments.status_in_transit') }}</option>
        <option value="delivered">{{ t('admin.shipments.status_delivered') }}</option>
        <option value="exception">{{ t('admin.shipments.status_exception') }}</option>
      </select>
    </div>

    <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase">{{ t('admin.trade_shipments.col_bol') }}</th>
            <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase">{{ t('admin.trade_shipments.col_trade') }}</th>
            <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase">{{ t('admin.shipments.col_carrier') }}</th>
            <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase">{{ t('admin.shipments.col_status') }}</th>
            <th class="px-6 py-3 text-left text-xs font-semibold text-gray-600 uppercase">{{ t('admin.shipments.col_actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-if="pending"><td colspan="5" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.shipments.loading') }}</td></tr>
          <tr v-else-if="error"><td colspan="5" class="px-6 py-10 text-center text-sm text-red-600">{{ error }}</td></tr>
          <tr v-else-if="shipments.length === 0"><td colspan="5" class="px-6 py-10 text-center text-sm text-gray-500">{{ t('admin.shipments.no_data') }}</td></tr>
          <tr v-else v-for="shipment in shipments" :key="shipment.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 font-mono text-sm">{{ shipment.billOfLadingNo || '—' }}</td>
            <td class="px-6 py-4 text-sm">#{{ shipment.transactionId }}</td>
            <td class="px-6 py-4 text-sm">{{ shipment.carrierName || '—' }}</td>
            <td class="px-6 py-4"><span :class="statusBadgeClass(shipment.uiStatus)" class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold">{{ formatStatus(shipment.uiStatus) }}</span></td>
            <td class="px-6 py-4 text-sm">
              <button class="text-orange-600 hover:text-orange-900 mr-3" @click="openEditModal(shipment)">{{ t('admin.shipments.edit') }}</button>
              <button v-if="shipment.rawStatus === 'PENDING'" class="text-emerald-600 hover:text-emerald-900" @click="dispatchShipment(shipment)">{{ t('admin.shipments.dispatch') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-screen items-center justify-center p-4">
        <button type="button" class="fixed inset-0 bg-gray-500/75 border-0" @click="closeModal" :aria-label="t('close')" />
        <form class="relative w-full max-w-lg rounded-xl bg-white p-6 shadow-xl space-y-4" @submit.prevent="saveShipment">
          <h3 class="text-lg font-semibold">{{ editingId ? t('admin.shipments.edit_shipment') : t('admin.shipments.create_shipment') }}</h3>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.trade_shipments.col_trade') }}</label>
            <input v-model.number="form.transactionId" type="number" required class="mt-1 w-full rounded-lg border px-3 py-2 text-sm" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.trade_shipments.col_bol') }}</label>
            <input v-model="form.billOfLadingNo" class="mt-1 w-full rounded-lg border px-3 py-2 text-sm" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.shipments.col_carrier') }}</label>
            <input v-model="form.carrierName" class="mt-1 w-full rounded-lg border px-3 py-2 text-sm" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.shipments.status') }}</label>
            <select v-model="form.status" class="mt-1 w-full rounded-lg border px-3 py-2 text-sm">
              <option value="PENDING">PENDING</option>
              <option value="DISPATCHED">DISPATCHED</option>
              <option value="IN_TRANSIT">IN_TRANSIT</option>
              <option value="DELIVERED">DELIVERED</option>
              <option value="EXCEPTION">EXCEPTION</option>
            </select>
          </div>
          <p v-if="formError" class="text-sm text-red-600">{{ formError }}</p>
          <div class="flex justify-end gap-3">
            <button type="button" class="rounded-lg border px-4 py-2 text-sm" @click="closeModal">{{ t('admin.shipments.cancel') }}</button>
            <button type="submit" class="rounded-lg bg-orange-600 px-4 py-2 text-sm text-white" :disabled="saving">{{ t('admin.shipments.save') }}</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const { enumLabel } = useDisplay()

const shipments = ref<any[]>([])
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20
const statusFilter = ref('all')
const showModal = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const formError = ref('')

const form = reactive({
  transactionId: 0,
  billOfLadingNo: '',
  carrierName: '',
  status: 'PENDING',
})

const mapTradeStatus = (status: string) => {
  const s = (status || 'PENDING').toUpperCase()
  if (s === 'PENDING') return 'pending'
  if (s === 'DELIVERED') return 'delivered'
  if (s === 'EXCEPTION') return 'exception'
  return 'in_transit'
}

const apiStatusFromFilter = (filter: string) => {
  if (filter === 'pending') return 'PENDING'
  if (filter === 'delivered') return 'DELIVERED'
  if (filter === 'exception') return 'EXCEPTION'
  if (filter === 'in_transit') return 'IN_TRANSIT'
  return filter.toUpperCase()
}

const normalizeTradeShipment = (row: any) => ({
  ...row,
  rawStatus: (row.status || 'PENDING').toUpperCase(),
  uiStatus: mapTradeStatus(row.status),
})

const fetchShipments = async () => {
  pending.value = true
  error.value = ''
  try {
    const params: Record<string, any> = { page: page.value, limit: pageSize }
    if (statusFilter.value !== 'all') params.status = apiStatusFromFilter(statusFilter.value)
    const res = await api.get<any>('/admin/shipments', params)
    shipments.value = (res.data || []).map(normalizeTradeShipment)
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const openCreateModal = () => {
  editingId.value = null
  Object.assign(form, { transactionId: 0, billOfLadingNo: '', carrierName: '', status: 'PENDING' })
  formError.value = ''
  showModal.value = true
}

const openEditModal = (shipment: any) => {
  editingId.value = shipment.id
  Object.assign(form, {
    transactionId: shipment.transactionId,
    billOfLadingNo: shipment.billOfLadingNo || '',
    carrierName: shipment.carrierName || '',
    status: shipment.rawStatus || 'PENDING',
  })
  formError.value = ''
  showModal.value = true
}

const closeModal = () => { showModal.value = false; saving.value = false }

const saveShipment = async () => {
  saving.value = true
  formError.value = ''
  try {
    const payload = { ...form }
    if (editingId.value) {
      await api.put(`/admin/shipments/${editingId.value}`, payload)
    } else {
      await api.post('/admin/shipments', payload)
    }
    closeModal()
    await fetchShipments()
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.save_failed')
  } finally {
    saving.value = false
  }
}

const dispatchShipment = async (shipment: any) => {
  try {
    await api.post(`/admin/shipments/${shipment.id}/dispatch`, {})
    await fetchShipments()
  } catch (err: any) {
    error.value = err?.message || t('errors.api.action_failed')
  }
}

const statusBadgeClass = (status: string) => {
  if (status === 'pending') return 'bg-yellow-100 text-yellow-800'
  if (status === 'in_transit') return 'bg-amber-100 text-amber-800'
  if (status === 'delivered') return 'bg-emerald-100 text-emerald-800'
  if (status === 'exception') return 'bg-red-100 text-red-800'
  return 'bg-gray-100 text-gray-800'
}

const formatStatus = (status: string) => enumLabel('shipment_status', status)

onMounted(fetchShipments)
</script>
