<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">{{ t('supplier.dashboard') }}</h1>
      <p class="text-gray-600 text-sm mt-1">{{ t('supplier.welcome') }}</p>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
      <div class="rounded-lg border border-gray-200 bg-white p-6">
        <p class="text-sm font-medium text-gray-500">{{ t('supplier.pending_pos') }}</p>
        <p class="mt-2 text-3xl font-semibold text-gray-900">{{ stats.pendingPOs }}</p>
      </div>
      <div class="rounded-lg border border-gray-200 bg-white p-6">
        <p class="text-sm font-medium text-gray-500">{{ t('supplier.active_pos') }}</p>
        <p class="mt-2 text-3xl font-semibold text-gray-900">{{ stats.activePOs }}</p>
      </div>
      <div class="rounded-lg border border-gray-200 bg-white p-6">
        <p class="text-sm font-medium text-gray-500">{{ t('supplier.completed_pos') }}</p>
        <p class="mt-2 text-3xl font-semibold text-gray-900">{{ stats.completedPOs }}</p>
      </div>
    </div>

    <!-- Recent POs -->
    <div class="rounded-lg border border-gray-200 bg-white">
      <div class="px-6 py-4 border-b border-gray-200">
        <h2 class="text-lg font-semibold text-gray-900">{{ t('supplier.recent_pos') }}</h2>
      </div>
      <div v-if="recentPOs.length === 0" class="px-6 py-8 text-center text-sm text-gray-400">
        {{ t('supplier.no_pos_yet') }}
      </div>
      <table v-else class="min-w-full divide-y divide-gray-200 text-sm">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('supplier.po_number') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('supplier.status') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('supplier.expected_date') }}</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">{{ t('supplier.total') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="po in recentPOs" :key="po.id">
            <td class="px-6 py-4 font-medium text-gray-900">{{ po.poNumber }}</td>
            <td class="px-6 py-4">
              <span :class="poStatusBadge(po.status)" class="inline-flex rounded-full px-2 py-0.5 text-xs font-semibold">
                {{ po.status }}
              </span>
            </td>
            <td class="px-6 py-4 text-gray-500">{{ po.expectedDate || '-' }}</td>
            <td class="px-6 py-4 text-right text-gray-900">{{ po.totalAmount || 0 }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'supplier' })

const { t } = useI18n()
const api = useApi()

const stats = reactive({ pendingPOs: 0, activePOs: 0, completedPOs: 0 })
const recentPOs = ref<any[]>([])

const poStatusBadge = (status: string) => {
  const map: Record<string, string> = {
    draft: 'bg-gray-100 text-gray-800',
    sent: 'bg-blue-100 text-blue-800',
    partial: 'bg-yellow-100 text-yellow-800',
    received: 'bg-green-100 text-green-800',
    cancelled: 'bg-red-100 text-red-800',
  }
  return map[status] || 'bg-gray-100 text-gray-800'
}

onMounted(async () => {
  try {
    const res = await api.get('/supplier/purchase-orders')
    const pos = res?.data || []
    recentPOs.value = pos.slice(0, 10)
    stats.pendingPOs = pos.filter((p: any) => p.status === 'draft' || p.status === 'sent').length
    stats.activePOs = pos.filter((p: any) => p.status === 'partial').length
    stats.completedPOs = pos.filter((p: any) => p.status === 'received').length
  } catch { /* silent */ }
})
</script>
