<template>
  <div>
    <div class="mb-6">
      <NuxtLink :to="localePath('/customer/returns')" class="flex items-center text-sm font-medium text-highlight hover:text-highlight-hover">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" aria-hidden="true" />
        {{ t('customer.returns.back') }}
      </NuxtLink>
    </div>

    <div v-if="pending" class="text-center py-10 text-light">{{ t('customer.returns.loading') }}</div>
    <div v-else-if="error" class="bg-error-10 p-4 rounded-md text-error">{{ error }}</div>
    <div v-else-if="ret" class="bg-white shadow-lg rounded-xl overflow-hidden">
      <div class="px-6 py-5 flex justify-between items-center bg-bg-alt border-b border-border">
        <div>
          <h3 class="text-lg font-semibold text-primary">{{ t('customer.returns.detail_title', { id: ret.id }) }}</h3>
          <p class="mt-1 text-sm text-light">{{ formatDate(ret.createdAt) }}</p>
        </div>
        <span :class="statusClass(ret.status)" class="px-3 py-1 text-xs font-semibold rounded-full">{{ statusLabel(ret.status) }}</span>
      </div>

      <div class="px-6 py-6">
        <dl class="grid grid-cols-1 gap-x-4 gap-y-4 sm:grid-cols-2">
          <div>
            <dt class="text-sm font-medium text-light">{{ t('customer.returns.col_order') }}</dt>
            <dd class="mt-1 text-sm text-primary">
              <NuxtLink :to="localePath(`/customer/orders/${ret.orderId}`)" class="text-highlight hover:underline">{{ ret.orderId }}</NuxtLink>
            </dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-light">{{ t('customer.returns.col_reason') }}</dt>
            <dd class="mt-1 text-sm text-primary">{{ ret.reason }}</dd>
          </div>
          <div v-if="ret.notes" class="sm:col-span-2">
            <dt class="text-sm font-medium text-light">{{ t('customer.returns.notes') }}</dt>
            <dd class="mt-1 text-sm text-primary whitespace-pre-line">{{ ret.notes }}</dd>
          </div>
          <div v-if="ret.rejectReason" class="sm:col-span-2">
            <dt class="text-sm font-medium text-error">{{ t('customer.returns.reject_reason') }}</dt>
            <dd class="mt-1 text-sm text-error">{{ ret.rejectReason }}</dd>
          </div>
        </dl>
      </div>

      <div class="px-6 py-6 border-t border-border bg-bg">
        <h4 class="text-md font-semibold text-primary mb-4">{{ t('customer.returns.items_title') }}</h4>
        <div v-if="items.length" class="overflow-x-auto bg-white border border-border rounded-lg">
          <table class="min-w-full divide-y divide-border">
            <thead class="bg-bg-alt">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-semibold text-primary uppercase">{{ t('customer.returns.col_product') }}</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-primary uppercase">{{ t('customer.returns.col_qty') }}</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-primary uppercase">{{ t('customer.returns.col_reason_code') }}</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-primary uppercase">{{ t('customer.returns.col_refund') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border">
              <tr v-for="item in items" :key="item.id">
                <td class="px-4 py-3 text-sm font-mono text-primary">{{ item.productId }}</td>
                <td class="px-4 py-3 text-sm text-primary">{{ item.quantity }}</td>
                <td class="px-4 py-3 text-sm text-light">{{ item.reasonCode }}</td>
                <td class="px-4 py-3 text-sm text-primary">{{ item.refundAmount ? formatNumber(item.refundAmount) : '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-else class="text-sm text-light">{{ t('customer.returns.no_items') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'customer', middleware: ['auth'] })

const route = useRoute()
const { t } = useI18n()
const { formatDate, formatNumber } = useDisplay()
const localePath = useLocalePath()
const api = useApi()

const id = route.params.id as string
const ret = ref<any>(null)
const items = ref<any[]>([])
const pending = ref(true)
const error = ref('')

const statusClass = (status: string) => {
  const map: Record<string, string> = {
    pending: 'bg-yellow-100 text-yellow-800',
    approved: 'bg-blue-100 text-blue-800',
    received: 'bg-indigo-100 text-indigo-800',
    refunded: 'bg-green-100 text-green-800',
    rejected: 'bg-red-100 text-red-800',
  }
  return map[status] || 'bg-gray-100 text-gray-800'
}

const statusLabel = (status: string) => t(`customer.returns.status_${status}`, status)

onMounted(async () => {
  try {
    const res = await api.get<any>(`/user/returns/${id}`)
    ret.value = res.return || res
    items.value = res.items || []
  } catch (e: any) {
    error.value = e?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
})
</script>
