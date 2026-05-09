<template>
  <div>
    <!-- Page Header -->
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.analytics.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ t('admin.analytics.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0 flex items-center gap-2">
        <button
          v-for="range in dateRanges"
          :key="range.value"
          @click="selectedRange = range.value"
          class="px-3 py-1.5 text-sm font-medium rounded-lg transition-colors"
          :class="selectedRange === range.value
            ? 'bg-orange-600 text-white'
            : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'"
        >
          {{ range.label }}
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="pending" class="mt-8 text-center py-10">
      <Icon name="heroicons:arrow-path" class="h-8 w-8 text-gray-400 animate-spin mx-auto" />
      <p class="mt-2 text-sm text-gray-500">{{ t('admin.analytics.loading') }}</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="mt-8 bg-red-50 p-4 rounded-md">
      <div class="flex">
        <div class="flex-shrink-0">
          <Icon name="heroicons:x-circle" class="h-5 w-5 text-red-400" />
        </div>
        <div class="ml-3">
          <h3 class="text-sm font-medium text-red-800">{{ t('admin.analytics.error') }}</h3>
          <div class="mt-2 text-sm text-red-700">
            <p>{{ error }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Charts Grid -->
    <template v-else>
      <div class="mt-8 grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Revenue Trends -->
        <div class="bg-white rounded-lg shadow p-6">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-medium text-gray-900">{{ t('admin.analytics.revenue_trends') }}</h2>
            <div class="h-10 w-10 rounded-lg bg-amber-50 flex items-center justify-center">
              <Icon name="heroicons:currency-dollar" class="h-5 w-5 text-orange-600" />
            </div>
          </div>
          <div class="h-72">
            <Line v-if="revenueChartData" :data="revenueChartData" :options="lineChartOptions" />
            <div v-else class="flex items-center justify-center h-full text-sm text-gray-500">
              {{ t('admin.analytics.no_data') }}
            </div>
          </div>
        </div>

        <!-- Order Trends -->
        <div class="bg-white rounded-lg shadow p-6">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-medium text-gray-900">{{ t('admin.analytics.order_trends') }}</h2>
            <div class="h-10 w-10 rounded-lg bg-emerald-50 flex items-center justify-center">
              <Icon name="heroicons:shopping-cart" class="h-5 w-5 text-emerald-600" />
            </div>
          </div>
          <div class="h-72">
            <Bar v-if="orderChartData" :data="orderChartData" :options="barChartOptions" />
            <div v-else class="flex items-center justify-center h-full text-sm text-gray-500">
              {{ t('admin.analytics.no_data') }}
            </div>
          </div>
        </div>

        <!-- Inquiry Conversion -->
        <div class="bg-white rounded-lg shadow p-6">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-medium text-gray-900">{{ t('admin.analytics.inquiry_conversion') }}</h2>
            <div class="h-10 w-10 rounded-lg bg-amber-50 flex items-center justify-center">
              <Icon name="heroicons:inbox-in" class="h-5 w-5 text-amber-600" />
            </div>
          </div>
          <div class="h-72 flex items-center justify-center">
            <Doughnut v-if="inquiryChartData" :data="inquiryChartData" :options="doughnutOptions" />
            <div v-else class="text-sm text-gray-500">
              {{ t('admin.analytics.no_data') }}
            </div>
          </div>
        </div>

        <!-- Top Products -->
        <div class="bg-white rounded-lg shadow p-6">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-medium text-gray-900">{{ t('admin.analytics.top_products') }}</h2>
            <div class="h-10 w-10 rounded-lg bg-orange-50 flex items-center justify-center">
              <Icon name="heroicons:trophy" class="h-5 w-5 text-orange-600" />
            </div>
          </div>
          <div class="overflow-hidden overflow-x-auto">
            <table v-if="topProducts.length > 0" class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-4 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">#</th>
                  <th class="px-4 py-3 text-left text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.analytics.col_product') }}</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.analytics.col_revenue') }}</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold text-gray-600 uppercase tracking-wider">{{ t('admin.analytics.col_quantity') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 bg-white">
                <tr v-for="(product, idx) in topProducts" :key="idx" class="hover:bg-gray-50 transition-colors">
                  <td class="px-4 py-3 text-sm font-medium text-gray-900">{{ idx + 1 }}</td>
                  <td class="px-4 py-3 text-sm text-gray-700">{{ product.name || product.productName || '-' }}</td>
                  <td class="px-4 py-3 text-sm text-gray-700 text-right font-medium">${{ (product.revenue || 0).toLocaleString() }}</td>
                  <td class="px-4 py-3 text-sm text-gray-700 text-right">{{ (product.quantity || product.totalQuantity || 0).toLocaleString() }}</td>
                </tr>
              </tbody>
            </table>
            <div v-else class="flex items-center justify-center h-40 text-sm text-gray-500">
              {{ t('admin.analytics.no_data') }}
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line, Bar, Doughnut } from 'vue-chartjs'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const { t } = useI18n()
const api = useApi()

const pending = ref(true)
const error = ref('')
const selectedRange = ref(12)

const revenueData = ref<any>(null)
const orderData = ref<any>(null)
const inquiryData = ref<any>(null)
const topProducts = ref<any[]>([])

const dateRanges = computed(() => [
  { label: t('admin.analytics.range_3m'), value: 3 },
  { label: t('admin.analytics.range_6m'), value: 6 },
  { label: t('admin.analytics.range_12m'), value: 12 }
])

// Chart options
const lineChartOptions = {
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
      ticks: {
        callback: (value: any) => '$' + value.toLocaleString()
      }
    }
  }
}

const barChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false }
  },
  scales: {
    y: {
      beginAtZero: true,
      ticks: {
        stepSize: 1
      }
    }
  }
}

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'bottom' as const,
      labels: {
        padding: 20,
        usePointStyle: true,
        pointStyleWidth: 10
      }
    }
  }
}

// Computed chart data
const revenueChartData = computed(() => {
  if (!revenueData.value?.data?.length) return null
  return {
    labels: revenueData.value.data.map((d: any) => d.month || d.label || ''),
    datasets: [{
      label: t('admin.analytics.revenue_label'),
      data: revenueData.value.data.map((d: any) => d.revenue || d.amount || 0),
      borderColor: '#4F46E5',
      backgroundColor: 'rgba(79, 70, 229, 0.1)',
      fill: true,
      tension: 0.3,
      pointBackgroundColor: '#4F46E5',
      pointBorderColor: '#fff',
      pointBorderWidth: 2,
      pointRadius: 4
    }]
  }
})

const orderChartData = computed(() => {
  if (!orderData.value?.data?.length) return null
  return {
    labels: orderData.value.data.map((d: any) => d.month || d.label || ''),
    datasets: [{
      label: t('admin.analytics.orders_label'),
      data: orderData.value.data.map((d: any) => d.count || d.orders || 0),
      backgroundColor: 'rgba(16, 185, 129, 0.7)',
      borderColor: '#10B981',
      borderWidth: 1,
      borderRadius: 4,
      barPercentage: 0.6
    }]
  }
})

const inquiryChartData = computed(() => {
  if (!inquiryData.value?.byStatus) return null
  const statusMap = inquiryData.value.byStatus
  const labels = Object.keys(statusMap)
  const values = Object.values(statusMap) as number[]

  const colorMap: Record<string, string> = {
    pending: '#F59E0B',
    reviewing: '#6366F1',
    quoted: '#3B82F6',
    accepted: '#10B981',
    rejected: '#EF4444',
    converted: '#059669',
    expired: '#9CA3AF'
  }

  return {
    labels: labels.map(l => l.charAt(0).toUpperCase() + l.slice(1)),
    datasets: [{
      data: values,
      backgroundColor: labels.map(l => colorMap[l] || '#6B7280'),
      borderWidth: 0,
      hoverOffset: 4
    }]
  }
})

const fetchData = async () => {
  pending.value = true
  error.value = ''

  try {
    const [rev, ord, inq, prod] = await Promise.allSettled([
      api.get<any>(`/admin/reports/revenue-trends?months=${selectedRange.value}`),
      api.get<any>(`/admin/reports/order-trends?months=${selectedRange.value}`),
      api.get<any>('/admin/reports/inquiries'),
      api.get<any>('/admin/reports/top-products?limit=10')
    ])

    if (rev.status === 'fulfilled') revenueData.value = rev.value
    if (ord.status === 'fulfilled') orderData.value = ord.value
    if (inq.status === 'fulfilled') inquiryData.value = inq.value
    if (prod.status === 'fulfilled') topProducts.value = prod.value?.data || prod.value || []

    const failed = [rev, ord, inq, prod].filter(r => r.status === 'rejected')
    if (failed.length === 4) {
      error.value = (failed[0] as PromiseRejectedResult).reason?.message || t('admin.analytics.error')
    }
  } catch (err: any) {
    error.value = err?.message || t('admin.analytics.error')
  } finally {
    pending.value = false
  }
}

watch(selectedRange, fetchData)
onMounted(fetchData)
</script>
