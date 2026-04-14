<template>
  <div>
    <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.dashboard.title') }}</h1>

    <div class="mt-4">
      <!-- Loading -->
      <div v-if="pending" class="text-center py-10">
        <Icon name="heroicons:arrow-path" class="h-8 w-8 text-gray-400 animate-spin mx-auto" />
        <p class="mt-2 text-sm text-gray-500">{{ t('admin.dashboard.loading') }}</p>
      </div>

      <!-- Error -->
      <div v-else-if="error" class="bg-red-50 rounded-lg p-6 text-center">
        <Icon name="heroicons:exclamation-triangle" class="h-8 w-8 text-red-400 mx-auto" />
        <p class="mt-2 text-sm text-red-600">{{ t('admin.dashboard.error') }}</p>
        <button @click="loadData" class="mt-3 text-sm text-red-700 underline">{{ t('admin.dashboard.error') }}</button>
      </div>

      <template v-else-if="stats">
        <!-- Stat Cards — 3 rows of 3 on lg, 2 cols on sm -->
        <div class="mt-5 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
          <!-- Total Users -->
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
                {{ stats.newUsersThisWeek }} {{ t('admin.dashboard.this_week') }}
              </p>
            </dd>
          </div>

          <!-- Total Orders -->
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
                {{ stats.pendingOrders }} {{ t('admin.dashboard.pending') }}
              </p>
            </dd>
          </div>

          <!-- Total Inquiries -->
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
                {{ stats.pendingInquiries }} {{ t('admin.dashboard.pending') }}
              </p>
            </dd>
          </div>

          <!-- Total Revenue -->
          <div class="relative bg-white pt-5 px-4 pb-12 sm:pt-6 sm:px-6 shadow rounded-lg overflow-hidden">
            <dt>
              <div class="absolute bg-green-500 rounded-md p-3">
                <Icon name="heroicons:currency-dollar" class="h-6 w-6 text-white" />
              </div>
              <p class="ml-16 text-sm font-medium text-gray-500 truncate">{{ t('admin.dashboard.total_revenue') }}</p>
            </dt>
            <dd class="ml-16 pb-6 flex items-baseline sm:pb-7">
              <p class="text-2xl font-semibold text-gray-900">${{ formatNumber(stats.totalSales) }}</p>
            </dd>
          </div>

          <!-- Revenue This Month -->
          <div class="relative bg-white pt-5 px-4 pb-12 sm:pt-6 sm:px-6 shadow rounded-lg overflow-hidden">
            <dt>
              <div class="absolute bg-emerald-500 rounded-md p-3">
                <Icon name="heroicons:chart-bar" class="h-6 w-6 text-white" />
              </div>
              <p class="ml-16 text-sm font-medium text-gray-500 truncate">{{ t('admin.dashboard.revenue_this_month') }}</p>
            </dt>
            <dd class="ml-16 pb-6 flex items-baseline sm:pb-7">
              <p class="text-2xl font-semibold text-emerald-600">${{ formatNumber(stats.revenueThisMonth) }}</p>
            </dd>
          </div>

          <!-- Active Customers -->
          <div class="relative bg-white pt-5 px-4 pb-12 sm:pt-6 sm:px-6 shadow rounded-lg overflow-hidden">
            <dt>
              <div class="absolute bg-violet-500 rounded-md p-3">
                <Icon name="heroicons:user-group" class="h-6 w-6 text-white" />
              </div>
              <p class="ml-16 text-sm font-medium text-gray-500 truncate">{{ t('admin.dashboard.active_customers') }}</p>
            </dt>
            <dd class="ml-16 pb-6 flex items-baseline sm:pb-7">
              <p class="text-2xl font-semibold text-gray-900">{{ stats.activeCustomers || 0 }}</p>
            </dd>
          </div>

          <!-- Avg Order Value -->
          <div class="relative bg-white pt-5 px-4 pb-12 sm:pt-6 sm:px-6 shadow rounded-lg overflow-hidden">
            <dt>
              <div class="absolute bg-cyan-500 rounded-md p-3">
                <Icon name="heroicons:receipt-percent" class="h-6 w-6 text-white" />
              </div>
              <p class="ml-16 text-sm font-medium text-gray-500 truncate">{{ t('admin.dashboard.avg_order_value') }}</p>
            </dt>
            <dd class="ml-16 pb-6 flex items-baseline sm:pb-7">
              <p class="text-2xl font-semibold text-gray-900">${{ formatNumber(stats.avgOrderValue) }}</p>
            </dd>
          </div>

          <!-- Conversion Rate -->
          <div class="relative bg-white pt-5 px-4 pb-12 sm:pt-6 sm:px-6 shadow rounded-lg overflow-hidden">
            <dt>
              <div class="absolute bg-orange-500 rounded-md p-3">
                <Icon name="heroicons:arrow-trending-up" class="h-6 w-6 text-white" />
              </div>
              <p class="ml-16 text-sm font-medium text-gray-500 truncate">{{ t('admin.dashboard.conversion_rate') }}</p>
            </dt>
            <dd class="ml-16 pb-6 flex items-baseline sm:pb-7">
              <p class="text-2xl font-semibold text-gray-900">{{ (stats.conversionRate || 0).toFixed(1) }}%</p>
            </dd>
          </div>
        </div>

        <!-- Revenue Sparkline + Recent Activity -->
        <div class="mt-6 grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Revenue 30-day chart -->
          <div class="bg-white rounded-lg shadow p-6">
            <h3 class="text-base font-semibold text-gray-900 mb-4">{{ t('admin.dashboard.revenue_chart') }}</h3>
            <div v-if="revenueByDay && revenueByDay.length" class="h-56">
              <Line :data="sparklineData" :options="sparklineOptions" />
            </div>
            <div v-else class="h-56 flex items-center justify-center text-sm text-gray-500">
              {{ t('admin.dashboard.no_activity') }}
            </div>
          </div>

          <!-- Recent Activity -->
          <div class="bg-white rounded-lg shadow">
            <div class="px-6 py-4 border-b border-gray-200">
              <h3 class="text-base font-semibold text-gray-900">{{ t('admin.dashboard.recent_activity') }}</h3>
            </div>
            <div class="overflow-hidden">
              <table class="min-w-full divide-y divide-gray-200">
                <thead class="bg-gray-50">
                  <tr>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('admin.dashboard.col_activity') }}</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('admin.dashboard.col_time') }}</th>
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
                    <td colspan="2" class="px-6 py-4 text-sm text-center text-gray-500">{{ t('admin.dashboard.no_activity') }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Line } from 'vue-chartjs'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const { t } = useI18n()
const api = useApi()

const stats = ref<any>(null)
const pending = ref(true)
const error = ref('')

const revenueByDay = computed(() => stats.value?.revenueByDay || [])

const sparklineData = computed(() => ({
  labels: revenueByDay.value.map((d: any) => d.day || d.date || ''),
  datasets: [{
    label: 'Revenue',
    data: revenueByDay.value.map((d: any) => d.revenue || d.amount || 0),
    borderColor: '#10B981',
    backgroundColor: 'rgba(16, 185, 129, 0.1)',
    fill: true,
    tension: 0.3,
    pointRadius: 2,
    pointHoverRadius: 4
  }]
}))

const sparklineOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (ctx: any) => `$${(ctx.parsed.y || 0).toLocaleString()}`
      }
    }
  },
  scales: {
    y: {
      beginAtZero: true,
      ticks: { callback: (v: any) => '$' + v.toLocaleString() }
    }
  }
}

const formatNumber = (num: number) => (num || 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

const loadData = async () => {
  pending.value = true
  error.value = ''
  try {
    stats.value = await api.get<any>('/admin/dashboard/stats')
  } catch (err: any) {
    error.value = err?.message || 'Failed to load dashboard data'
  } finally {
    pending.value = false
  }
}

onMounted(loadData)
</script>
