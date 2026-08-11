<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.trades.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.trades.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0">
        <button type="button" class="inline-flex items-center gap-2 rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50" @click="exportTradesXlsx">
          <Icon name="heroicons:arrow-down-tray" class="h-4 w-4" aria-hidden="true" />
          {{ t('admin.trades.export_xlsx') }}
        </button>
      </div>
    </div>

    <div class="mt-6 flex flex-wrap items-center gap-4">
      <div class="flex-1 min-w-[200px]">
        <input id="trades-search" v-model="searchInput" name="search" type="text" autocomplete="off" :placeholder="t('admin.trades.search_placeholder')" class="block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
      </div>
      <select id="trades-statusFilter" name="statusFilter" v-model="statusFilter" class="rounded-md border border-gray-300 px-3 py-2 text-sm">
        <option value="">{{ t('admin.trades.all_statuses') }}</option>
        <option value="DRAFT">{{ t('admin.trades.draft') }}</option>
        <option value="PENDING">{{ t('admin.trades.pending') }}</option>
        <option value="CONFIRMED">{{ t('admin.trades.confirmed') }}</option>
        <option value="PAID">{{ t('admin.trades.paid') }}</option>
        <option value="SHIPPED">{{ t('admin.trades.shipped') }}</option>
        <option value="COMPLETED">{{ t('admin.trades.completed') }}</option>
        <option value="CANCELLED">{{ t('admin.trades.cancelled') }}</option>
      </select>
    </div>

    <div class="mt-8 overflow-hidden rounded-lg bg-white shadow ring-1 ring-black ring-opacity-5">
      <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
          <tr>
            <th class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{{ t('admin.trades.col_transaction') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.trades.col_customer') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.trades.col_type') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.trades.col_amount') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.trades.col_status') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.trades.col_created') }}</th>
            <th class="relative py-3.5 pl-3 pr-4 text-sm font-semibold text-gray-900 sm:pr-6">{{ t('admin.trades.actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <tr v-if="pending">
            <td colspan="7" class="py-5 text-center text-sm text-gray-500">{{ t('admin.trades.loading') }}</td>
          </tr>
          <tr v-else-if="error">
            <td colspan="7" class="py-5 text-center text-sm text-red-600">{{ error }}</td>
          </tr>
          <tr v-else-if="trades.length === 0">
            <td colspan="7" class="py-5 text-center text-sm text-gray-500">{{ t('admin.trades.no_data') }}</td>
          </tr>
          <tr v-else v-for="trade in trades" :key="trade.id" class="hover:bg-gray-50 cursor-pointer" role="button" tabindex="0" :aria-label="t('admin.trades.view_detail')" @click="navigateToDetail(trade.id)" @keydown.enter.prevent="navigateToDetail(trade.id)" @keydown.space.prevent="navigateToDetail(trade.id)">
            <td class="py-4 pl-4 pr-3 text-sm sm:pl-6">
              <div class="font-medium text-gray-900">{{ trade.reference || String(trade.id).substring(0, 8) }}</div>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">
              {{ cell(trade.userId) }}
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ cell(trade.terms) }}</td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ cur(trade.currency) }} {{ formatNumber(trade.totalAmount || 0) }}</td>
            <td class="px-3 py-4 text-sm">
              <span :class="[statusBadgeClass(trade.status), 'inline-flex rounded-full px-2 text-xs font-semibold leading-5']">
                {{ enumLabel('trade_status', trade.status) }}
              </span>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ formatDate(trade.createdAt) }}</td>
            <td class="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
              <NuxtLink :to="localePath(`/admin/trades/${trade.id}`)" class="text-orange-600 hover:text-orange-900" @click.stop>{{ t('admin.trades.view') }}</NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="pagination" class="mt-4 flex items-center justify-between rounded-lg border-t border-gray-200 bg-white px-4 py-3 shadow sm:px-6">
      <div class="text-sm text-gray-700">
        {{ t('admin.trades.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
      </div>
      <div class="flex items-center gap-2">
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page <= 1" @click="prevPage">
          {{ t('admin.trades.previous') }}
        </button>
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page >= pagination.totalPages" @click="nextPage">
          {{ t('admin.trades.next') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const { t } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, cell, enumLabel, formatNumber, formatDate } = useDisplay()
const api = useApi()
const router = useRouter()

const trades = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20
const searchInput = ref('')
const statusFilter = ref('')

let searchTimeout: ReturnType<typeof setTimeout>

const fetchTrades = async () => {
  pending.value = true
  error.value = ''
  try {
    const params: Record<string, any> = { page: page.value, limit: pageSize }
    if (searchInput.value.trim()) params.search = searchInput.value.trim()
    if (statusFilter.value) params.status = statusFilter.value

    const res = await api.get<any>('/admin/trades', params)
    trades.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const navigateToDetail = (id: string) => {
  router.push(localePath(`/admin/trades/${id}`))
}

/** 导出贸易交易 XLSX */
const exportTradesXlsx = async () => {
  try {
    // useApi 走 HttpOnly cookie 凭证（credentials: 'include'）。
    // 不要用已废弃的 useAuth().token —— 恒为 null，会拼出 "Bearer null" 导致导出必然失败。
    const blob = await api.exportAdminXlsx('trades')
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `trades_${new Date().toISOString().split('T')[0]}.xlsx`
    a.click()
    URL.revokeObjectURL(url)
  } catch (err: any) {
    notifyError(err, t('errors.api.export_failed'))
  }
}

const statusBadgeClass = (status: string) => {
  // Backend trade statuses are UPPERCASE (DRAFT/PENDING/CONFIRMED/PAID/SHIPPED/COMPLETED/CANCELLED).
  // 兼容旧数据 / 非贸易状态字段被误传，统一 lowercase 比较（B-3）。
  const s = String(status || '').toLowerCase()
  if (s === 'draft') return 'bg-gray-100 text-gray-800'
  if (s === 'pending') return 'bg-yellow-100 text-yellow-800'
  if (s === 'confirmed' || s === 'in_progress') return 'bg-orange-100 text-orange-800'
  if (s === 'paid') return 'bg-blue-100 text-blue-800'
  if (s === 'shipped') return 'bg-indigo-100 text-indigo-800'
  if (s === 'completed') return 'bg-green-100 text-green-800'
  if (s === 'cancelled') return 'bg-red-100 text-red-800'
  return 'bg-gray-100 text-gray-800'
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

watch([page, searchInput, statusFilter], () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(fetchTrades, 300)
})

onMounted(fetchTrades)
</script>
