<template>
  <div class="bg-white shadow overflow-hidden sm:rounded-lg">
    <div class="px-4 py-5 sm:px-6 border-b border-gray-200 flex justify-between items-center">
      <div>
        <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.trades.title') }}</h3>
        <p class="mt-1 max-w-2xl text-sm text-gray-500">{{ t('customer.trades.subtitle') }}</p>
      </div>
      <div>
        <div class="flex items-center gap-2">
          <select v-model="createForm.incoterms" class="rounded-md border border-gray-300 px-2 py-2 text-sm">
            <option value="">{{ t('customer.trades.incoterms_label') }}</option>
            <option value="FOB">FOB</option>
            <option value="CIF">CIF</option>
            <option value="EXW">EXW</option>
            <option value="DAP">DAP</option>
            <option value="DDP">DDP</option>
          </select>
          <input v-model="createForm.currency" maxlength="3" class="w-20 rounded-md border border-gray-300 px-2 py-2 text-sm uppercase" :placeholder="t('customer.trades.currency_placeholder')" />
          <button @click="createTrade" :disabled="creating" class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-orange-600 hover:bg-orange-700 disabled:opacity-50">
            {{ creating ? t('customer.trades.starting') : t('customer.trades.start_new') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="pending" class="text-center py-10">
      <p class="text-gray-500">{{ t('customer.trades.loading') }}</p>
    </div>
    <div v-else-if="error" class="bg-red-50 p-4 rounded-md mx-4 my-4">
      <p class="text-red-700">{{ error }}</p>
    </div>
    <div v-else-if="trades.length === 0" class="text-center py-10">
      <Icon name="heroicons:inbox-in" class="mx-auto h-12 w-12 text-gray-400" />
      <h3 class="mt-2 text-sm font-medium text-gray-900">{{ t('customer.trades.no_trades') }}</h3>
      <p class="mt-1 text-sm text-gray-500">{{ t('customer.trades.no_trades_desc') }}</p>
      <div class="mt-6">
        <NuxtLink :to="localePath('/customer/inquiries')" class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-orange-600 hover:bg-orange-700">
          {{ t('customer.trades.view_inquiries') }}
        </NuxtLink>
      </div>
    </div>
    <div v-else class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('customer.trades.col_date') }}</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('customer.trades.col_incoterms') }}</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('customer.trades.col_status') }}</th>
            <th scope="col" class="relative px-6 py-3"><span class="sr-only">{{ t('customer.trades.open_dashboard') }}</span></th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="trade in trades" :key="trade.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ new Date(trade.createdAt || trade.created_at).toLocaleDateString() }}</td>
            <td class="px-6 py-4 text-sm text-gray-900">{{ trade.incoterms || trade.terms || t('common.display.tbd') }}</td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span :class="[
                trade.status === 'DRAFT' || trade.status === 'PENDING' ? 'bg-yellow-100 text-yellow-800' :
                trade.status === 'COMPLETED' ? 'bg-green-100 text-green-800' :
                trade.status === 'CANCELLED' ? 'bg-gray-100 text-gray-800' :
                'bg-orange-100 text-orange-800',
                'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium'
              ]">{{ enumLabel('trade_status', trade.status) }}</span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <NuxtLink :to="localePath(`/customer/trades/${trade.id}`)" class="text-orange-600 hover:text-orange-900">{{ t('customer.trades.open_dashboard') }}</NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="pagination && pagination.totalPages > 1" class="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
        <div class="flex-1 flex justify-between sm:hidden">
          <button @click="prevPage" :disabled="page <= 1" class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">{{ t('customer.trades.previous') }}</button>
          <button @click="nextPage" :disabled="page >= pagination.totalPages" class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">{{ t('customer.trades.next') }}</button>
        </div>
        <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
          <div>
            <p class="text-sm text-gray-700">
              {{ t('customer.trades.showing', { from: ((page - 1) * 20) + 1, to: Math.min(page * 20, pagination.total), total: pagination.total }) }}
            </p>
          </div>
          <div>
            <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px" aria-label="Pagination">
              <button @click="prevPage" :disabled="page <= 1" class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50">
                <Icon name="heroicons:chevron-left" class="h-5 w-5" />
              </button>
              <button @click="nextPage" :disabled="page >= pagination.totalPages" class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50">
                <Icon name="heroicons:chevron-right" class="h-5 w-5" />
              </button>
            </nav>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch, onMounted } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, incotermsOrDefault: defInc, enumLabel } = useDisplay()

const trades = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const creating = ref(false)
const createForm = reactive({ incoterms: '', currency: cur(null) })

const fetchTrades = async () => {
  pending.value = true; error.value = ''
  try {
    const res = await api.get<any>(`/user/trades?page=${page.value}&limit=20`)
    trades.value = res.data || []
    pagination.value = res.pagination || null
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally { pending.value = false }
}

const createTrade = async () => {
  creating.value = true
  try {
    await api.post('/user/trades', { incoterms: defInc(createForm.incoterms), currency: cur(createForm.currency).toUpperCase() })
    await fetchTrades()
  } catch (err: any) {
    error.value = err?.message || t('errors.api.trade_start_failed')
  } finally { creating.value = false }
}

const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) page.value++ }
const prevPage = () => { if (page.value > 1) page.value-- }

watch(page, fetchTrades)
onMounted(fetchTrades)
</script>