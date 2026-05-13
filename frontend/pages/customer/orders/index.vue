<template>
  <div class="bg-white shadow overflow-hidden sm:rounded-lg">
    <div class="px-4 py-5 sm:px-6 border-b border-gray-200 flex items-center justify-between gap-4">
      <div>
        <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.orders.title') }}</h3>
        <p class="mt-1 max-w-2xl text-sm text-gray-500">{{ t('customer.orders.subtitle') }}</p>
      </div>
      <NuxtLink :to="localePath('/customer/orders/new')" class="inline-flex items-center rounded-md bg-orange-600 px-3 py-2 text-sm font-medium text-white hover:bg-orange-700">
        <Icon name="heroicons:sparkles" class="mr-1.5 h-4 w-4" />
        {{ t('customer.orders.place_new') }}
      </NuxtLink>
    </div>

    <!-- Orders List -->
    <div v-if="pending" class="text-center py-10">
      <p class="text-gray-500">{{ t('customer.orders.loading') }}</p>
    </div>
    
    <div v-else-if="error" class="bg-red-50 p-4 rounded-md mx-4 my-4">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <div v-else-if="orders.length === 0" class="text-center py-10">
      <Icon name="heroicons:inbox" class="mx-auto h-12 w-12 text-gray-400" />
      <h3 class="mt-2 text-sm font-medium text-gray-900">{{ t('customer.orders.no_orders') }}</h3>
      <p class="mt-1 text-sm text-gray-500">{{ t('customer.orders.no_orders_desc') }}</p>
      <div class="mt-6">
        <NuxtLink :to="localePath('/customer/orders/new')" class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-orange-600 hover:bg-orange-700">
          {{ t('customer.orders.start_order') }}
        </NuxtLink>
      </div>
    </div>
    
    <div v-else class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('customer.orders.col_order_id') }}</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('customer.orders.col_date') }}</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('customer.orders.col_amount') }}</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('customer.orders.col_status') }}</th>
            <th scope="col" class="relative px-6 py-3"><span class="sr-only">{{ t('customer.orders.view_details') }}</span></th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="order in orders" :key="order.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-orange-600">
              #{{ order.orderNumber || order.id.substring(0,8) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ formatDate(order.createdAt) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ cur(order.currency) }} {{ order.totalAmount ? formatNumber(order.totalAmount) : t('display.zero') }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span :class="[statusBadge(order.status), 'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium']">
                {{ enumLabel('order_status', order.status) }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
             <NuxtLink :to="localePath(`/customer/orders/${order.id}`)" class="text-orange-600 hover:text-orange-900">{{ t('customer.orders.view_details') }}</NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>
      
      <!-- Pagination -->
      <div v-if="pagination && pagination.totalPages > 1" class="bg-white px-4 py-3 flex items-center justify-between border-t border-gray-200 sm:px-6">
        <div class="flex-1 flex justify-between sm:hidden">
          <button @click="prevPage" :disabled="page <= 1" class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">{{ t('common.previous') }}</button>
          <button @click="nextPage" :disabled="page >= pagination.totalPages" class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50">{{ t('common.next') }}</button>
        </div>
        <div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
          <div>
            <p class="text-sm text-gray-700">
              {{ t('customer.orders.showing', { from: ((page - 1) * 20) + 1, to: Math.min(page * 20, pagination.total), total: pagination.total }) }}
            </p>
          </div>
          <div>
            <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px" :aria-label="t('pagination.nav_label')">
              <button @click="prevPage" :disabled="page <= 1" class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50">
                <span class="sr-only">{{ t('common.previous') }}</span>
                <Icon name="heroicons:chevron-left" class="h-5 w-5" />
              </button>
              <button @click="nextPage" :disabled="page >= pagination.totalPages" class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50">
                <span class="sr-only">{{ t('common.next') }}</span>
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
import { ref, watch, onMounted } from 'vue'

definePageMeta({
  layout: 'customer',
  middleware: ['auth']
})

const { t } = useI18n()
const { formatNumber, formatDate, currencyOrDefault: cur, enumLabel } = useDisplay()
const localePath = useLocalePath()
const api = useApi()

const orders = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)

const fetchOrders = async () => {
  pending.value = true
  error.value = ''
  
  try {
    const res = await api.getOrders({ page: page.value, limit: 20 })
    orders.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const statusBadge = (status: string) => {
  if (status === 'pending_confirmation') return 'bg-orange-100 text-orange-800'
  if (status === 'pending' || status === 'processing' || status === 'production') return 'bg-yellow-100 text-yellow-800'
  if (status === 'confirmed') return 'bg-amber-100 text-amber-800'
  if (status === 'shipped') return 'bg-orange-100 text-orange-800'
  if (status === 'delivered') return 'bg-green-100 text-green-800'
  if (status === 'cancelled') return 'bg-red-100 text-red-800'
  return 'bg-gray-100 text-gray-800'
}

const nextPage = () => {
  if (page.value < pagination.value.totalPages) page.value++
}

const prevPage = () => {
  if (page.value > 1) page.value--
}

watch(page, fetchOrders)
onMounted(fetchOrders)
</script>

