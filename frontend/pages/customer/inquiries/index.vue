<template>
  <div class="bg-white shadow overflow-hidden sm:rounded-lg">
    <div class="px-4 py-5 sm:px-6 border-b border-gray-200 flex justify-between items-center">
      <div>
        <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.inquiries.title') }}</h3>
        <p class="mt-1 max-w-2xl text-sm text-gray-500">{{ t('customer.inquiries.subtitle') }}</p>
      </div>
      <div>
        <NuxtLink :to="localePath('/customer/inquiries/new')" class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-orange-600 hover:bg-orange-700">
          {{ t('customer.inquiries.new_inquiry') }}
        </NuxtLink>
      </div>
    </div>

    <!-- Inquiries List -->
    <div v-if="pending" class="text-center py-10">
      <p class="text-gray-500">{{ t('customer.inquiries.loading') }}</p>
    </div>
    
    <div v-else-if="error" class="bg-red-50 p-4 rounded-md mx-4 my-4">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <div v-else-if="inquiries.length === 0" class="text-center py-10">
      <Icon name="lucide:inbox" class="mx-auto h-12 w-12 text-gray-400" />
      <h3 class="mt-2 text-sm font-medium text-gray-900">{{ t('customer.inquiries.no_inquiries') }}</h3>
      <p class="mt-1 text-sm text-gray-500">{{ t('customer.inquiries.no_inquiries_desc') }}</p>
      <div class="mt-6">
        <NuxtLink to="/products" class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-orange-600 hover:bg-orange-700">
          {{ t('customer.inquiries.browse_products') }}
        </NuxtLink>
      </div>
    </div>
    
    <div v-else class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('customer.inquiries.col_date') }}</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('customer.inquiries.col_products') }}</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('customer.inquiries.col_status') }}</th>
            <th scope="col" class="relative px-6 py-3"><span class="sr-only">{{ t('customer.inquiries.view_details') }}</span></th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="inquiry in inquiries" :key="inquiry.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ formatDate(inquiry.createdAt) }}
            </td>
            <td class="px-6 py-4 text-sm text-gray-900 max-w-xs truncate">
              {{ formatProducts(inquiry.interestedProducts) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span :class="[
                inquiry.status === 'pending' ? 'bg-yellow-100 text-yellow-800' :
                inquiry.status === 'contacted' || inquiry.status === 'quoted' ? 'bg-blue-100 text-blue-800' :
                inquiry.status === 'negotiating' ? 'bg-indigo-100 text-indigo-800' :
                inquiry.status === 'won' || inquiry.status === 'converted' ? 'bg-green-100 text-green-800' :
                inquiry.status === 'lost' ? 'bg-red-100 text-red-800' : 'bg-gray-100 text-gray-800',
                'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium'
              ]">
                {{ enumLabel('inquiry_status', inquiry.status) }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
             <NuxtLink :to="localePath(`/customer/inquiries/${inquiry.id}`)" class="text-orange-600 hover:text-orange-900">{{ t('customer.inquiries.view_details') }}</NuxtLink>
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
              {{ t('customer.inquiries.showing', { from: ((page - 1) * 20) + 1, to: Math.min(page * 20, pagination.total), total: pagination.total }) }}
            </p>
          </div>
          <div>
            <nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px" :aria-label="t('pagination.nav_label')">
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
import { ref, watch, onMounted } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const { enumLabel, formatDate } = useDisplay()
const localePath = useLocalePath()

const inquiries = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)

const fetchInquiries = async () => {
  pending.value = true; error.value = ''
  try {
    const res = await api.get<any>(`/user/inquiries?page=${page.value}&limit=20`)
    inquiries.value = res.data || []
    pagination.value = res.pagination || null
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally { pending.value = false }
}

const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) page.value++ }
const prevPage = () => { if (page.value > 1) page.value-- }

const formatProducts = (products: unknown) => {
  if (Array.isArray(products)) return products.length > 0 ? products.join(', ') : t('customer.inquiries.general_inquiry')
  if (typeof products === 'string') return products.trim() !== '' ? products : t('customer.inquiries.general_inquiry')
  return t('customer.inquiries.general_inquiry')
}

watch(page, fetchInquiries)
onMounted(fetchInquiries)
</script>

