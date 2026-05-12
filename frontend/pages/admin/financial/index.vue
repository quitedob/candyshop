<template>
  <div>
    <!-- Page Header -->
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.financial.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ t('admin.financial.description') }}</p>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="mt-8 text-center py-10">
      <Icon name="heroicons:arrow-path" class="h-8 w-8 text-gray-400 animate-spin mx-auto" />
      <p class="mt-2 text-sm text-gray-500">{{ t('admin.financial.loading') }}</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="mt-8 bg-red-50 rounded-lg p-6 text-center">
      <Icon name="heroicons:exclamation-triangle" class="h-8 w-8 text-red-400 mx-auto" />
      <p class="mt-2 text-sm text-red-600">{{ t('admin.financial.error') }}</p>
      <button @click="loadData" class="mt-3 text-sm text-red-700 underline">{{ t('admin.dashboard.error') }}</button>
    </div>

    <div v-else class="mt-6 space-y-6">
      <!-- Summary Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div class="bg-white rounded-lg shadow p-5">
          <dt class="text-sm font-medium text-gray-500 truncate">{{ t('admin.financial.accounts_receivable') }}</dt>
          <dd class="mt-1 text-2xl font-bold text-gray-900">${{ formatNumber(overview?.accountsReceivable || 0) }}</dd>
        </div>
        <div class="bg-white rounded-lg shadow p-5">
          <dt class="text-sm font-medium text-gray-500 truncate">{{ t('admin.financial.overdue_amount') }}</dt>
          <dd class="mt-1 text-2xl font-bold text-red-600">${{ formatNumber(overview?.overdueAmount || 0) }}</dd>
        </div>
        <div class="bg-white rounded-lg shadow p-5">
          <dt class="text-sm font-medium text-gray-500 truncate">{{ t('admin.financial.revenue_this_month') }}</dt>
          <dd class="mt-1 text-2xl font-bold text-green-600">${{ formatNumber(overview?.revenueThisMonth || 0) }}</dd>
        </div>
        <div class="bg-white rounded-lg shadow p-5">
          <dt class="text-sm font-medium text-gray-500 truncate">{{ t('admin.financial.outstanding_invoices') }}</dt>
          <dd class="mt-1 text-2xl font-bold text-gray-900">{{ overview?.overdueCount || 0 }}</dd>
        </div>
      </div>

      <!-- Charts Row -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Payment Breakdown -->
        <div class="bg-white rounded-lg shadow p-6">
          <h3 class="text-base font-semibold text-gray-900 mb-4">{{ t('admin.financial.payment_breakdown') }}</h3>
          <div v-if="breakdown && breakdown.length" class="h-64">
            <Doughnut :data="paymentChartData" :options="paymentChartOptions" />
          </div>
          <div v-else class="h-64 flex items-center justify-center text-sm text-gray-500">{{ t('admin.financial.no_data') }}</div>
        </div>

        <!-- Revenue Trend placeholder -->
        <div class="bg-white rounded-lg shadow p-6">
          <h3 class="text-base font-semibold text-gray-900 mb-4">{{ t('admin.analytics.revenue_trends') }}</h3>
          <div v-if="revenueData && revenueData.length" class="h-64">
            <Line :data="revenueChartData" :options="lineChartOptions" />
          </div>
          <div v-else class="h-64 flex items-center justify-center text-sm text-gray-500">{{ t('admin.financial.no_data') }}</div>
        </div>
      </div>

      <!-- Outstanding Invoices Table -->
      <div class="bg-white rounded-lg shadow">
        <div class="px-6 py-4 border-b border-gray-200">
          <h3 class="text-base font-semibold text-gray-900">{{ t('admin.financial.outstanding_list') }}</h3>
        </div>
        <div v-if="invoicesLoading" class="p-6 text-center">
          <Icon name="heroicons:arrow-path" class="h-6 w-6 text-gray-400 animate-spin mx-auto" />
        </div>
        <div v-else-if="invoices && invoices.length" class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-300">
            <thead class="bg-gray-50">
              <tr>
                <th class="py-3.5 px-4 text-left text-sm font-semibold text-gray-900">{{ t('admin.financial.col_invoice_no') }}</th>
                <th class="py-3.5 px-4 text-left text-sm font-semibold text-gray-900">{{ t('admin.financial.col_amount') }}</th>
                <th class="py-3.5 px-4 text-left text-sm font-semibold text-gray-900">{{ t('admin.financial.col_due_date') }}</th>
                <th class="py-3.5 px-4 text-left text-sm font-semibold text-gray-900">{{ t('admin.financial.col_status') }}</th>
                <th class="py-3.5 px-4 text-left text-sm font-semibold text-gray-900">{{ t('admin.financial.col_order') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr v-for="inv in invoices" :key="inv.id" class="hover:bg-gray-50" :class="{ 'bg-red-50': isOverdue(inv) }">
                <td class="py-4 px-4 text-sm font-medium text-gray-900">{{ inv.invoice_no || inv.invoiceNo || '-' }}</td>
                <td class="py-4 px-4 text-sm text-gray-700">${{ formatNumber(inv.total_amount || inv.totalAmount || 0) }}</td>
                <td class="py-4 px-4 text-sm text-gray-700">{{ formatDate(inv.due_date || inv.dueDate) }}</td>
                <td class="py-4 px-4">
                  <span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium" :class="statusClass(inv.status)">
                    {{ inv.status }}
                  </span>
                </td>
                <td class="py-4 px-4 text-sm text-gray-500">{{ inv.order_id || inv.orderId || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="p-6 text-center text-sm text-gray-500">{{ t('admin.financial.no_invoices') }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Doughnut, Line } from 'vue-chartjs'

definePageMeta({ layout: 'admin', middleware: ['auth'] })
const { t } = useI18n()
const api = useApi()

const loading = ref(true)
const error = ref(false)
const overview = ref<any>(null)
const breakdown = ref<any[]>([])
const revenueData = ref<any[]>([])
const invoices = ref<any[]>([])
const invoicesLoading = ref(true)

const loadData = async () => {
  loading.value = true
  error.value = false
  try {
    const [overviewRes, breakdownRes, revenueRes] = await Promise.all([
      api.get<any>('/admin/financial/overview').catch(() => null),
      api.get<any>('/admin/financial/payment-breakdown').catch(() => ({ breakdown: [] })),
      api.get<any>('/admin/reports/revenue-trends?months=12').catch(() => ({ data: [] }))
    ])
    overview.value = overviewRes
    breakdown.value = breakdownRes?.breakdown || []
    revenueData.value = revenueRes?.data || []
  } catch (e) {
    error.value = true
  } finally {
    loading.value = false
  }

  invoicesLoading.value = true
  try {
    const invRes = await api.get<any>('/admin/financial/outstanding-invoices?page=1&limit=20')
    invoices.value = invRes?.data || invRes?.items || []
  } catch (e) {
    invoices.value = []
  } finally {
    invoicesLoading.value = false
  }
}

const paymentChartData = computed(() => {
  const colors = ['#10B981', '#F97316', '#F59E0B', '#EF4444', '#78716C']
  return {
    labels: breakdown.value.map(b => b.status || b.Status),
    datasets: [{
      data: breakdown.value.map(b => b.count || b.Count || 0),
      backgroundColor: colors.slice(0, breakdown.value.length)
    }]
  }
})

const paymentChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { position: 'bottom' as const } }
}

const revenueChartData = computed(() => ({
  labels: revenueData.value.map(d => d.month || d.Month),
  datasets: [{
    label: t('admin.financial.revenue'),
    data: revenueData.value.map(d => d.revenue || d.Revenue || 0),
    borderColor: '#10B981',
    backgroundColor: 'rgba(16, 185, 129, 0.1)',
    fill: true,
    tension: 0.3
  }]
}))

const lineChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: { y: { beginAtZero: true } }
}

const { formatNumber, formatDate } = useDisplay()
const isOverdue = (inv: any) => {
  const due = inv.due_date || inv.dueDate
  if (!due) return false
  return new Date(due) < new Date() && inv.status !== 'paid' && inv.status !== 'voided'
}
const statusClass = (status: string) => {
  const map: Record<string, string> = {
    paid: 'bg-green-100 text-green-800',
    sent: 'bg-orange-100 text-orange-800',
    draft: 'bg-gray-100 text-gray-800',
    overdue: 'bg-red-100 text-red-800',
    voided: 'bg-gray-100 text-gray-500'
  }
  return map[status] || 'bg-gray-100 text-gray-800'
}

onMounted(loadData)
</script>
