<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">{{ t('supplier.nav_pos') }}</h1>
    <div v-if="pending" class="text-sm text-gray-500">{{ t('common.loading') }}</div>
    <div v-else-if="!pos.length" class="text-sm text-gray-400">{{ t('supplier.no_pos_yet') }}</div>
    <table v-else class="min-w-full divide-y divide-gray-200 text-sm dark:divide-gray-700">
      <thead class="bg-gray-50 dark:bg-gray-800">
        <tr>
          <th class="px-4 py-2 text-left">{{ t('supplier.po_number') }}</th>
          <th class="px-4 py-2 text-left">{{ t('supplier.status') }}</th>
          <th class="px-4 py-2 text-right">{{ t('supplier.total') }}</th>
          <th class="px-4 py-2 text-right">{{ t('supplier.actions') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="po in pos" :key="po.id" class="border-t border-gray-100 dark:border-gray-700">
          <td class="px-4 py-2">{{ po.poNumber }}</td>
          <td class="px-4 py-2">{{ po.status }}</td>
          <td class="px-4 py-2 text-right">{{ po.totalAmount }}</td>
          <td class="px-4 py-2 text-right">
            <button
              v-if="po.status === 'sent'"
              type="button"
              class="mr-2 rounded border border-gray-300 px-2 py-1 text-xs hover:bg-gray-50 dark:border-gray-600 dark:hover:bg-gray-800"
              :aria-label="t('supplier.accept_po')"
              @click="updateStatus(po.id, 'partially_received')"
            >
              {{ t('supplier.accept_po') }}
            </button>
            <button
              v-if="po.status === 'sent' || po.status === 'partially_received'"
              type="button"
              class="rounded border border-orange-300 px-2 py-1 text-xs text-orange-700 hover:bg-orange-50 dark:border-orange-700 dark:text-orange-300"
              :aria-label="t('supplier.mark_shipped')"
              @click="updateStatus(po.id, 'received')"
            >
              {{ t('supplier.mark_received') }}
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'supplier', middleware: ['auth'] })
const { t } = useI18n()
const api = useApi()
const pos = ref<any[]>([])
const pending = ref(true)

async function load() {
  pending.value = true
  try {
    const res = await api.get<any>('/supplier/purchase-orders')
    pos.value = res?.data || res || []
  } finally {
    pending.value = false
  }
}

async function updateStatus(id: string, status: string) {
  await api.put(`/supplier/purchase-orders/${id}/status`, { status })
  await load()
}

onMounted(load)
</script>
