<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.fulfillments.title') }}</h1>
        <p class="text-sm text-gray-500 mt-1">{{ t('admin.fulfillments.subtitle') }}</p>
      </div>
      <div class="flex items-center gap-2">
        <select v-model="statusFilter" class="border border-gray-200 rounded-lg px-3 py-2 text-sm">
          <option value="">{{ t('admin.fulfillments.filter_all') }}</option>
          <option value="pending">{{ t('admin.fulfillments.status_pending') }}</option>
          <option value="picking">{{ t('admin.fulfillments.status_picking') }}</option>
          <option value="shipped">{{ t('admin.fulfillments.status_shipped') }}</option>
          <option value="delivered">{{ t('admin.fulfillments.status_delivered') }}</option>
          <option value="cancelled">{{ t('admin.fulfillments.status_cancelled') }}</option>
        </select>
        <button class="px-3 py-2 border border-gray-200 rounded-lg text-sm hover:bg-gray-50" @click="fetchRows">
          {{ t('admin.fulfillments.refresh') }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-sm text-gray-500">{{ t('admin.fulfillments.loading') }}</div>
    <div v-else-if="error" class="text-sm text-red-600">{{ error }}</div>
    <div v-else class="bg-white shadow rounded-lg overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200 text-sm">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-4 py-2 text-left text-xs font-semibold text-gray-500 uppercase">{{ t('admin.fulfillments.col_id') }}</th>
            <th class="px-4 py-2 text-left text-xs font-semibold text-gray-500 uppercase">{{ t('admin.fulfillments.col_order') }}</th>
            <th class="px-4 py-2 text-left text-xs font-semibold text-gray-500 uppercase">{{ t('admin.fulfillments.col_status') }}</th>
            <th class="px-4 py-2 text-left text-xs font-semibold text-gray-500 uppercase">{{ t('admin.fulfillments.col_warehouse') }}</th>
            <th class="px-4 py-2 text-left text-xs font-semibold text-gray-500 uppercase">{{ t('admin.fulfillments.col_tracking') }}</th>
            <th class="px-4 py-2 text-left text-xs font-semibold text-gray-500 uppercase">{{ t('admin.fulfillments.col_created') }}</th>
            <th class="px-4 py-2 text-left text-xs font-semibold text-gray-500 uppercase">{{ t('admin.fulfillments.col_actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr v-if="rows.length === 0">
            <td colspan="7" class="px-4 py-6 text-center text-sm text-gray-500">{{ t('admin.fulfillments.empty') }}</td>
          </tr>
          <tr v-for="f in rows" :key="f.id" class="hover:bg-gray-50">
            <td class="px-4 py-2 font-mono text-xs text-gray-700">{{ String(f.id).substring(0, 12) }}</td>
            <td class="px-4 py-2">
              <NuxtLink :to="localePath(`/admin/orders/${f.orderId}`)" class="text-orange-600 hover:text-orange-900">
                {{ f.orderNumber || String(f.orderId).substring(0, 8) }}
              </NuxtLink>
            </td>
            <td class="px-4 py-2">
              <span :class="statusBadgeClass(f.status)" class="inline-flex rounded-full px-2 text-xs font-semibold">
                {{ f.status }}
              </span>
            </td>
            <td class="px-4 py-2 text-gray-500">{{ f.warehouseId || '—' }}</td>
            <td class="px-4 py-2 text-gray-500">{{ f.trackingNumber || '—' }}</td>
            <td class="px-4 py-2 text-gray-500">{{ formatDate(f.createdAt) }}</td>
            <td class="px-4 py-2 text-sm">
              <button v-if="f.status === 'pending' || f.status === 'picking'"
                      class="text-orange-600 hover:text-orange-900 mr-2"
                      :disabled="busyId === f.id"
                      @click="ship(f)">
                {{ t('admin.fulfillments.ship') }}
              </button>
              <button v-if="f.status === 'shipped'"
                      class="text-emerald-600 hover:text-emerald-900 mr-2"
                      :disabled="busyId === f.id"
                      @click="deliver(f)">
                {{ t('admin.fulfillments.deliver') }}
              </button>
              <button v-if="f.status === 'pending' || f.status === 'picking'"
                      class="text-red-600 hover:text-red-900"
                      :disabled="busyId === f.id"
                      @click="cancel(f)">
                {{ t('admin.fulfillments.cancel') }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="pagination" class="px-4 py-3 flex items-center justify-between border-t border-gray-100 bg-gray-50">
        <p class="text-xs text-gray-500">
          {{ t('admin.fulfillments.pagination', { total: pagination.total, page: pagination.page, pages: pagination.totalPages }) }}
        </p>
        <div class="flex items-center gap-1">
          <button class="px-2 py-1 border border-gray-200 rounded text-xs"
                  :disabled="page <= 1"
                  @click="prevPage">
            {{ t('admin.fulfillments.prev') }}
          </button>
          <button class="px-2 py-1 border border-gray-200 rounded text-xs"
                  :disabled="!pagination || page >= pagination.totalPages"
                  @click="nextPage">
            {{ t('admin.fulfillments.next') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * 管理员履约总览页（M-15）
 *
 * 直接消费后端 GET /admin/fulfillments，提供独立的履约视图，
 * 与现有的 /admin/shipments 分工：
 *   - shipments：以"在途运单"视角聚焦轨迹/承运商，状态做了 PENDING→IN_TRANSIT 等映射；
 *   - fulfillments：以"仓库出库工单"视角，保留后端原生状态值，便于对账和按订单回查。
 */
import { ref, watch, onMounted } from 'vue'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const localePath = useLocalePath()
const { formatDate } = useDisplay()

const rows = ref<any[]>([])
const pagination = ref<any>(null)
const loading = ref(false)
const error = ref('')
const page = ref(1)
const pageSize = 20
const statusFilter = ref('')
const busyId = ref('')

const fetchRows = async () => {
  loading.value = true
  error.value = ''
  try {
    const params: Record<string, unknown> = { page: page.value, limit: pageSize }
    if (statusFilter.value) params.status = statusFilter.value
    const res = await api.get<any>('/admin/fulfillments', params)
    rows.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    loading.value = false
  }
}

const statusBadgeClass = (status: string) => {
  switch (status) {
    case 'pending':
      return 'bg-yellow-100 text-yellow-800'
    case 'picking':
      return 'bg-blue-100 text-blue-800'
    case 'shipped':
      return 'bg-orange-100 text-orange-800'
    case 'delivered':
      return 'bg-emerald-100 text-emerald-800'
    case 'cancelled':
      return 'bg-red-100 text-red-800'
    default:
      return 'bg-gray-100 text-gray-800'
  }
}

const ship = async (f: any) => {
  const trackingNumber = window.prompt(t('admin.fulfillments.tracking_prompt'))
  if (!trackingNumber || !trackingNumber.trim()) return
  const carrier = window.prompt(t('admin.fulfillments.carrier_prompt')) || ''
  busyId.value = f.id
  try {
    await api.put(`/admin/fulfillments/${f.id}/ship`, {
      trackingNumber: trackingNumber.trim(),
      carrier: carrier.trim(),
    })
    await fetchRows()
  } catch (err: any) {
    error.value = err?.message || t('errors.api.save_failed')
  } finally {
    busyId.value = ''
  }
}

const deliver = async (f: any) => {
  if (!window.confirm(t('admin.fulfillments.deliver_confirm'))) return
  busyId.value = f.id
  try {
    await api.put(`/admin/fulfillments/${f.id}/deliver`, {})
    await fetchRows()
  } catch (err: any) {
    error.value = err?.message || t('errors.api.save_failed')
  } finally {
    busyId.value = ''
  }
}

const cancel = async (f: any) => {
  if (!window.confirm(t('admin.fulfillments.cancel_confirm'))) return
  busyId.value = f.id
  try {
    await api.put(`/admin/fulfillments/${f.id}/cancel`, {})
    await fetchRows()
  } catch (err: any) {
    error.value = err?.message || t('errors.api.save_failed')
  } finally {
    busyId.value = ''
  }
}

const prevPage = () => { if (page.value > 1) { page.value--; fetchRows() } }
const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) { page.value++; fetchRows() } }

watch(statusFilter, () => { page.value = 1; fetchRows() })

onMounted(fetchRows)
</script>
