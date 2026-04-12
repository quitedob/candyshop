<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.trades.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.trades.description') }}</p>
      </div>
    </div>

    <div class="mt-6 flex flex-wrap items-center gap-4">
      <div class="flex-1 min-w-[200px]">
        <input v-model="searchInput" type="text" :placeholder="t('admin.trades.search_placeholder')" class="block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
      </div>
      <select v-model="statusFilter" class="rounded-md border border-gray-300 px-3 py-2 text-sm">
        <option value="">{{ t('admin.trades.all_statuses') }}</option>
        <option value="draft">{{ t('admin.trades.draft') }}</option>
        <option value="pending">{{ t('admin.trades.pending') }}</option>
        <option value="confirmed">{{ t('admin.trades.confirmed') }}</option>
        <option value="in_progress">{{ t('admin.trades.in_progress') }}</option>
        <option value="completed">{{ t('admin.trades.completed') }}</option>
        <option value="cancelled">{{ t('admin.trades.cancelled') }}</option>
      </select>
      <select v-model="typeFilter" class="rounded-md border border-gray-300 px-3 py-2 text-sm">
        <option value="">{{ t('admin.trades.all_types') }}</option>
        <option value="import">{{ t('admin.trades.import') }}</option>
        <option value="export">{{ t('admin.trades.export') }}</option>
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
            <th class="relative py-3.5 pl-3 pr-4 sm:pr-6"><span class="sr-only">{{ t('admin.trades.actions') }}</span></th>
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
          <tr v-else v-for="trade in trades" :key="trade.id" class="hover:bg-gray-50 cursor-pointer" @click="navigateToDetail(trade.id)">
            <td class="py-4 pl-4 pr-3 text-sm sm:pl-6">
              <div class="font-medium text-gray-900">{{ trade.transactionNumber || trade.id.substring(0, 8) }}</div>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">
              <template v-if="trade.user">
                <div class="font-medium text-gray-900">{{ trade.user.firstName }} {{ trade.user.lastName }}</div>
                <div class="text-xs text-gray-500">{{ trade.user.email }}</div>
              </template>
              <template v-else>{{ trade.userId || '-' }}</template>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ trade.tradeType || '-' }}</td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ trade.currency || 'USD' }} {{ (trade.totalAmount || 0).toLocaleString() }}</td>
            <td class="px-3 py-4 text-sm">
              <span :class="[statusBadgeClass(trade.status), 'inline-flex rounded-full px-2 text-xs font-semibold leading-5']">
                {{ trade.status }}
              </span>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ trade.createdAt ? new Date(trade.createdAt).toLocaleDateString() : '-' }}</td>
            <td class="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
              <NuxtLink :to="`/admin/trades/${trade.id}`" class="text-blue-600 hover:text-blue-900" @click.stop>{{ t('admin.trades.view') }}</NuxtLink>
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
const typeFilter = ref('')

let searchTimeout: ReturnType<typeof setTimeout>

const fetchTrades = async () => {
  pending.value = true
  error.value = ''
  try {
    const params: Record<string, any> = { page: page.value, limit: pageSize }
    if (searchInput.value.trim()) params.search = searchInput.value.trim()
    if (statusFilter.value) params.status = statusFilter.value
    if (typeFilter.value) params.type = typeFilter.value

    const res = await api.get<any>('/admin/trades', params)
    trades.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || 'Failed to fetch trades'
  } finally {
    pending.value = false
  }
}

const navigateToDetail = (id: string) => {
  router.push(`/admin/trades/${id}`)
}

const statusBadgeClass = (status: string) => {
  if (status === 'draft') return 'bg-gray-100 text-gray-800'
  if (status === 'pending') return 'bg-yellow-100 text-yellow-800'
  if (status === 'confirmed' || status === 'in_progress') return 'bg-blue-100 text-blue-800'
  if (status === 'completed') return 'bg-green-100 text-green-800'
  if (status === 'cancelled') return 'bg-red-100 text-red-800'
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

watch([page, searchInput, statusFilter, typeFilter], () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(fetchTrades, 300)
})

onMounted(fetchTrades)
</script>
