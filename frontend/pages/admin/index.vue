<template>
  <div>
    <PageHeader :title="t('admin.dashboard.title')" :description="t('admin.dashboard.subtitle')" />

    <!-- Loading: skeleton metric cards -->
    <div v-if="pending" class="dashboard-grid">
      <MetricCard v-for="i in 8" :key="i" :loading="true" />
    </div>

    <!-- Error -->
    <ErrorState v-else-if="error" :message="error" :action-text="t('errors.tryAgain')" @action="loadData" />

    <template v-else-if="stats">
      <!-- KPI Bento Grid — 8 semantically differentiated cards -->
      <div class="dashboard-grid">
        <MetricCard
          :title="t('admin.dashboard.total_users')"
          :value="stats.totalUsers"
          :trend="'up'"
          :trend-label="t('admin.dashboard.new_users_this_week', { count: stats.newUsersThisWeek })"
          icon="heroicons:users"
          icon-bg="info"
        />
        <MetricCard
          :title="t('admin.dashboard.total_orders')"
          :value="stats.totalOrders"
          :trend-label="`${stats.pendingOrders} ${t('admin.dashboard.pending')}`"
          icon="heroicons:shopping-bag"
          icon-bg="warning"
        />
        <MetricCard
          :title="t('admin.dashboard.total_inquiries')"
          :value="stats.totalInquiries"
          :trend-label="`${stats.pendingInquiries} ${t('admin.dashboard.pending')}`"
          icon="lucide:inbox"
          icon-bg="accent"
        />
        <MetricCard
          :title="t('admin.dashboard.total_revenue')"
          :value="`$${formatNumber(stats.totalSales)}`"
          icon="heroicons:currency-dollar"
          icon-bg="success"
        />
        <MetricCard
          :title="t('admin.dashboard.revenue_this_month')"
          :value="`$${formatNumber(stats.revenueThisMonth)}`"
          icon="heroicons:chart-bar"
          icon-bg="success"
        />
        <MetricCard
          :title="t('admin.dashboard.active_customers')"
          :value="stats.activeCustomers || 0"
          icon="heroicons:user-group"
          icon-bg="info"
        />
        <MetricCard
          :title="t('admin.dashboard.avg_order_value')"
          :value="`$${formatNumber(stats.avgOrderValue)}`"
          icon="heroicons:receipt-percent"
          icon-bg="default"
        />
        <MetricCard
          :title="t('admin.dashboard.conversion_rate')"
          :value="`${(stats.conversionRate || 0).toFixed(1)}%`"
          icon="heroicons:arrow-trending-up"
          icon-bg="accent"
        />
      </div>

      <!-- Secondary Widgets Row -->
      <div class="mt-6 grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Revenue Chart -->
        <div class="lg:col-span-2 card p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ t('admin.dashboard.revenue_chart') }}</h3>
          <div v-if="revenueByDay && revenueByDay.length" class="h-56">
            <Line :data="sparklineData" :options="sparklineOptions" />
          </div>
          <div v-else class="h-56 flex items-center justify-center text-sm text-gray-500">
            {{ t('admin.dashboard.no_activity') }}
          </div>
        </div>

        <!-- Recent Activity -->
        <div class="card flex flex-col">
          <div class="px-6 py-4 border-b border-gray-100">
            <h3 class="text-base font-semibold text-gray-900">{{ t('admin.dashboard.recent_activity') }}</h3>
          </div>
          <div class="flex-1 overflow-auto">
            <table class="min-w-full divide-y divide-gray-100">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-6 py-3 text-start text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('admin.dashboard.col_activity') }}</th>
                  <th class="px-6 py-3 text-start text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('admin.dashboard.col_time') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100">
                <tr v-for="(activity, idx) in stats.recentActivity" :key="idx" class="hover:bg-gray-50">
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    <span class="flex items-center gap-2">
                      <Icon v-if="activity.type === 'order'" name="heroicons:shopping-cart" class="h-4 w-4 text-gray-400" />
                      <Icon v-else-if="activity.type === 'inquiry'" name="lucide:inbox" class="h-4 w-4 text-gray-400" />
                      <Icon v-else-if="activity.type === 'user'" name="heroicons:user" class="h-4 w-4 text-gray-400" />
                      {{ activityMessage(activity) }}
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ formatRelativeTime(activity.createdAt) }}</td>
                </tr>
                <tr v-if="!stats.recentActivity?.length">
                  <td colspan="2" class="px-6 py-8 text-sm text-center text-gray-500">{{ t('admin.dashboard.no_activity') }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
      <!-- Reports Tabs -->
      <div class="mt-8 card">
        <div class="px-6 py-4 border-b border-gray-100 flex flex-wrap gap-2">
          <button
            v-for="tab in reportTabs"
            :key="tab.key"
            type="button"
            class="px-3 py-1.5 text-sm font-medium rounded-lg transition-colors"
            :class="activeReport === tab.key ? 'bg-orange-600 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'"
            @click="loadReport(tab.key)"
          >
            {{ tab.label }}
          </button>
        </div>
        <div class="p-6">
          <div v-if="reportPending" class="py-8 text-center text-sm text-gray-500">{{ t('admin.reports.loading') }}</div>
          <div v-else-if="reportError" class="py-4 text-sm text-red-600">{{ reportError }}</div>
          <div v-else-if="!reportData?.length" class="py-8 text-center text-sm text-gray-500">{{ t('admin.reports.no_data') }}</div>
          <div v-else class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200 text-sm">
              <thead class="bg-gray-50">
                <tr>
                  <th v-for="col in reportColumns" :key="col.key" class="px-4 py-2 text-left text-xs font-semibold text-gray-600 uppercase">{{ col.label }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100">
                <tr v-for="(row, idx) in reportData" :key="idx" class="hover:bg-gray-50">
                  <td v-for="col in reportColumns" :key="col.key" class="px-4 py-2 text-gray-700">{{ formatCell(row, col.key) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

    </template>
  </div>
</template>

<script setup lang="ts">
import { Line } from 'vue-chartjs'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const { t, te } = useI18n()
const api = useApi()
const { formatRelativeTime } = useRelativeTime()

const activityMessage = (activity: any) => {
  const key = activity?.messageKey || ''
  const vars = { ...(activity?.vars || {}) }
  if (key === 'admin.dashboard.activity_order' && !String(vars.user || '').trim()) {
    vars.user = t('admin.dashboard.unknown_customer')
  }
  if (key === 'admin.dashboard.activity_inquiry' && !String(vars.company || '').trim()) {
    vars.company = t('admin.dashboard.unknown_company')
  }
  if (key === 'admin.dashboard.activity_user' && !String(vars.user || '').trim()) {
    vars.user = t('admin.dashboard.unknown_user')
  }
  if (key && te(key)) return t(key, vars)
  if (activity?.message) return activity.message
  return key || '—'
}

const stats = ref<any>(null)
const pending = ref(true)
const error = ref('')
const activeReport = ref('sales-velocity')
const reportData = ref<any[]>([])
const reportPending = ref(false)
const reportError = ref('')

const reportTabs = computed(() => [
  { key: 'sales-velocity', label: t('admin.reports.sales_velocity'), endpoint: '/admin/reports/sales-velocity' },
  { key: 'rfm', label: t('admin.reports.rfm'), endpoint: '/admin/reports/rfm' },
  { key: 'churn', label: t('admin.reports.churn'), endpoint: '/admin/reports/churn' },
  { key: 'inventory-health', label: t('admin.reports.inventory_health'), endpoint: '/admin/reports/inventory-health' },
  { key: 'profit-loss', label: t('admin.reports.profit_loss'), endpoint: '/admin/reports/profit-loss' },
  { key: 'replenishment', label: t('admin.reports.replenishment'), endpoint: '/admin/reports/replenishment' },
  { key: 'inquiry-conversion', label: t('admin.reports.inquiry_conversion'), endpoint: '/admin/reports/conversion-trends' },
])

const reportColumnMap: Record<string, { key: string; label: string }[]> = {
  'sales-velocity': [
    { key: 'productName', label: t('admin.reports.col_product') },
    { key: 'productId', label: t('admin.reports.col_id') },
    { key: 'velocity', label: t('admin.reports.col_velocity') },
    { key: 'totalQuantity', label: t('admin.reports.col_qty') },
  ],
  rfm: [
    { key: 'customerName', label: t('admin.reports.col_customer') },
    { key: 'recency', label: t('admin.reports.col_recency') },
    { key: 'frequency', label: t('admin.reports.col_frequency') },
    { key: 'monetary', label: t('admin.reports.col_monetary') },
    { key: 'segment', label: t('admin.reports.col_segment') },
  ],
  churn: [
    { key: 'customerName', label: t('admin.reports.col_customer') },
    { key: 'email', label: t('admin.reports.col_email') },
    { key: 'daysInactive', label: t('admin.reports.col_days_inactive') },
    { key: 'lastOrderAt', label: t('admin.reports.col_recency') },
  ],
  'inventory-health': [
    { key: 'productName', label: t('admin.reports.col_product') },
    { key: 'stock', label: t('admin.reports.col_stock') },
    { key: 'daysOfSupply', label: t('admin.reports.col_velocity') },
    { key: 'healthStatus', label: t('admin.orders.col_status') },
  ],
  'profit-loss': [
    { key: 'period', label: t('admin.reports.col_month') },
    { key: 'revenue', label: t('admin.reports.col_revenue') },
    { key: 'cogs', label: t('admin.reports.col_cost') },
    { key: 'grossProfit', label: t('admin.reports.col_profit') },
  ],
  replenishment: [
    { key: 'productName', label: t('admin.reports.col_product') },
    { key: 'currentStock', label: t('admin.reports.col_stock') },
    { key: 'suggestedQty', label: t('admin.reports.col_reorder') },
    { key: 'avgDailySales', label: t('admin.reports.col_velocity') },
  ],
  'inquiry-conversion': [
    { key: 'month', label: t('admin.reports.col_month') },
    { key: 'inquiries', label: t('admin.reports.col_inquiries') },
    { key: 'converted', label: t('admin.reports.col_converted') },
    { key: 'conversionRate', label: t('admin.reports.col_rate') },
  ],
}

const reportColumns = computed(() => reportColumnMap[activeReport.value] || [])

const formatCell = (row: any, key: string) => {
  const val = row[key] ?? row[key.replace(/([A-Z])/g, '_$1').toLowerCase()]
  if (val == null) return '—'
  if (typeof val === 'number' && ['revenue', 'cogs', 'grossProfit', 'monetary'].includes(key)) return `$${formatNumber(val)}`
  if (key === 'conversionRate' && typeof val === 'number') return `${val.toFixed(1)}%`
  return String(val)
}

const loadReport = async (key?: string) => {
  if (key) activeReport.value = key
  const tab = reportTabs.value.find(t => t.key === activeReport.value)
  if (!tab) return
  reportPending.value = true
  reportError.value = ''
  try {
    const res = await api.get<any>(tab.endpoint)
    reportData.value = res.data || []
  } catch (err: any) {
    reportError.value = err?.message || t('admin.reports.error')
    reportData.value = []
  } finally {
    reportPending.value = false
  }
}

const { formatNumber, currencyOrDefault: cur } = useDisplay()
const { colors: chartColors, rgba } = useChartTheme()

const revenueByDay = computed(() => stats.value?.revenueByDay || [])

const sparklineData = computed(() => {
  const highlight = chartColors.value?.highlight ?? '#fd933d'
  return {
  labels: revenueByDay.value.map((d: any) => d.day || d.date || ''),
  datasets: [{
    label: t('admin.dashboard.revenue'),
    data: revenueByDay.value.map((d: any) => d.revenue || d.amount || 0),
    borderColor: highlight,
    backgroundColor: rgba(highlight, 0.08),
    fill: true,
    tension: 0.3,
    pointRadius: 2,
    pointHoverRadius: 4
  }]
}
})

const sparklineOptions = computed(() => {
  const muted = chartColors.value?.muted ?? '#7e7570'
  return {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
      tooltip: {
        callbacks: {
          label: (ctx: any) => `${cur()} ${formatNumber(ctx.parsed.y || 0)}`
        }
      }
    },
    scales: {
      x: {
        display: true,
        title: { display: true, text: t('admin.dashboard.date'), color: muted, font: { size: 12 } },
        ticks: { display: true, maxTicksLimit: 10, autoSkip: true, maxRotation: 45, color: muted, font: { size: 11 } },
        grid: { display: false }
      },
      y: {
        beginAtZero: true,
        ticks: { callback: (v: any) => cur() + ' ' + formatNumber(v) }
      }
    }
  }
})

const loadData = async () => {
  pending.value = true
  error.value = ''
  try {
    stats.value = await api.get<any>('/admin/dashboard/stats')
  } catch (err: any) {
    error.value = err?.message || t('admin.dashboard.error')
  } finally {
    pending.value = false
  }
}

onMounted(async () => {
  await loadData()
  await loadReport()
})
</script>

<style scoped>
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(1, 1fr);
  gap: var(--spacing-lg);
}

@media (min-width: 640px) {
  .dashboard-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .dashboard-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

.card {
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}
</style>
