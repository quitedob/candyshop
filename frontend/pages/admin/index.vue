<template>
  <div>
    <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.dashboard.title') }}</h1>
    
    <div class="mt-4">
      <div v-if="pending" class="text-center py-10">
        <p class="text-gray-500">{{ t('admin.dashboard.loading') }}</p>
      </div>
      
      <div v-else-if="error" class="bg-red-50 p-4 rounded-md">
        <div class="flex">
          <div class="flex-shrink-0">
            <Icon name="heroicons:x-circle" class="h-5 w-5 text-red-400" />
          </div>
          <div class="ml-3">
            <h3 class="text-sm font-medium text-red-800">{{ t('admin.dashboard.error') }}</h3>
            <div class="mt-2 text-sm text-red-700">
              <p>{{ error }}</p>
            </div>
          </div>
        </div>
      </div>
      
      <template v-else-if="stats">
        <!-- Stats Widgets -->
        <dl class="mt-5 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
          <div class="relative bg-white pt-5 px-4 pb-12 sm:pt-6 sm:px-6 shadow rounded-lg overflow-hidden">
            <dt>
              <div class="absolute bg-blue-500 rounded-md p-3">
                <Icon name="heroicons:users" class="h-6 w-6 text-white" />
              </div>
              <p class="ml-16 text-sm font-medium text-gray-500 truncate">{{ t('admin.dashboard.total_users') }}</p>
            </dt>
            <dd class="ml-16 pb-6 flex items-baseline sm:pb-7">
              <p class="text-2xl font-semibold text-gray-900">{{ stats.totalUsers }}</p>
              <p class="ml-2 flex items-baseline text-sm font-semibold text-green-600">
                <Icon name="heroicons:arrow-up" class="self-center flex-shrink-0 h-5 w-5 text-green-500" />
                <span class="sr-only">Increased by</span>
                {{ stats.newUsersThisWeek }} {{ t('admin.dashboard.this_week') }}
              </p>
            </dd>
          </div>

          <div class="relative bg-white pt-5 px-4 pb-12 sm:pt-6 sm:px-6 shadow rounded-lg overflow-hidden">
            <dt>
              <div class="absolute bg-yellow-500 rounded-md p-3">
                <Icon name="heroicons:shopping-bag" class="h-6 w-6 text-white" />
              </div>
              <p class="ml-16 text-sm font-medium text-gray-500 truncate">{{ t('admin.dashboard.total_orders') }}</p>
            </dt>
            <dd class="ml-16 pb-6 flex items-baseline sm:pb-7">
              <p class="text-2xl font-semibold text-gray-900">{{ stats.totalOrders }}</p>
              <p class="ml-2 flex items-baseline text-sm font-semibold text-yellow-600">
                <span class="sr-only">Pending</span>
                {{ stats.pendingOrders }} {{ t('admin.dashboard.pending') }}
              </p>
            </dd>
          </div>

          <div class="relative bg-white pt-5 px-4 pb-12 sm:pt-6 sm:px-6 shadow rounded-lg overflow-hidden">
            <dt>
              <div class="absolute bg-indigo-500 rounded-md p-3">
                <Icon name="heroicons:inbox-in" class="h-6 w-6 text-white" />
              </div>
              <p class="ml-16 text-sm font-medium text-gray-500 truncate">{{ t('admin.dashboard.total_inquiries') }}</p>
            </dt>
            <dd class="ml-16 pb-6 flex items-baseline sm:pb-7">
              <p class="text-2xl font-semibold text-gray-900">{{ stats.totalInquiries }}</p>
              <p class="ml-2 flex items-baseline text-sm font-semibold text-indigo-600">
                <span class="sr-only">Pending</span>
                {{ stats.pendingInquiries }} {{ t('admin.dashboard.pending') }}
              </p>
            </dd>
          </div>

          <div class="relative bg-white pt-5 px-4 pb-12 sm:pt-6 sm:px-6 shadow rounded-lg overflow-hidden">
            <dt>
              <div class="absolute bg-green-500 rounded-md p-3">
                <Icon name="heroicons:currency-dollar" class="h-6 w-6 text-white" />
              </div>
              <p class="ml-16 text-sm font-medium text-gray-500 truncate">{{ t('admin.dashboard.total_revenue') }}</p>
            </dt>
            <dd class="ml-16 pb-6 flex items-baseline sm:pb-7">
              <p class="text-2xl font-semibold text-gray-900">${{ stats.totalSales?.toLocaleString() }}</p>
            </dd>
          </div>
        </dl>
        
        <!-- Recent Activity -->
        <div class="mt-8">
          <h2 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.dashboard.recent_activity') }}</h2>
          <div class="mt-4 shadow overflow-hidden border-b border-gray-200 sm:rounded-lg">
            <table class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('admin.dashboard.col_activity') }}</th>
                  <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('admin.dashboard.col_time') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
                <tr v-for="(activity, idx) in stats.recentActivity" :key="idx">
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    <span class="flex items-center">
                      <Icon v-if="activity.type === 'order'" name="heroicons:shopping-cart" class="mr-2 h-4 w-4 text-gray-400" />
                      <Icon v-else-if="activity.type === 'inquiry'" name="heroicons:inbox-in" class="mr-2 h-4 w-4 text-gray-400" />
                      <Icon v-else-if="activity.type === 'user'" name="heroicons:user" class="mr-2 h-4 w-4 text-gray-400" />
                      {{ activity.message }}
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {{ activity.time }}
                  </td>
                </tr>
                <tr v-if="!stats.recentActivity?.length">
                  <td colspan="2" class="px-6 py-4 whitespace-nowrap text-sm text-center text-gray-500">{{ t('admin.dashboard.no_activity') }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const { t } = useI18n()
const api = useApi()

const stats = ref<any>(null)
const pending = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    stats.value = await api.get<any>('/admin/dashboard/stats')
  } catch (err: any) {
    error.value = err?.message || 'Failed to load dashboard data'
  } finally {
    pending.value = false
  }
})
</script>
