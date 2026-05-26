<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-primary">{{ t('customer.returns.title') }}</h1>
      <p class="mt-1 text-sm text-light">{{ t('customer.returns.subtitle') }}</p>
    </div>

    <div v-if="pending" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-10 w-10 border-4 border-orange-200 border-t-orange-600" />
    </div>
    <div v-else-if="error" class="bg-error-10 border border-error rounded-xl p-6 text-center text-error">{{ error }}</div>
    <div v-else-if="!returns.length" class="text-center py-16 text-light">
      <Icon name="heroicons:arrow-uturn-left" class="h-12 w-12 mx-auto mb-3 text-gray-300" aria-hidden="true" />
      <p>{{ t('customer.returns.no_returns') }}</p>
      <NuxtLink :to="localePath('/customer/orders')" class="mt-3 inline-block text-sm text-highlight hover:underline">
        {{ t('customer.returns.view_orders') }}
      </NuxtLink>
    </div>
    <div v-else class="bg-white shadow rounded-lg overflow-hidden">
      <table class="min-w-full divide-y divide-border">
        <thead class="bg-bg-alt">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-light uppercase">{{ t('customer.returns.col_id') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-light uppercase">{{ t('customer.returns.col_order') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-light uppercase">{{ t('customer.returns.col_reason') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-light uppercase">{{ t('customer.returns.col_status') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-light uppercase">{{ t('customer.returns.col_date') }}</th>
            <th class="px-6 py-3" />
          </tr>
        </thead>
        <tbody class="divide-y divide-border">
          <tr v-for="ret in returns" :key="ret.id">
            <td class="px-6 py-4 text-sm font-mono text-primary">{{ ret.id }}</td>
            <td class="px-6 py-4 text-sm text-primary">{{ ret.orderId }}</td>
            <td class="px-6 py-4 text-sm text-light">{{ ret.reason }}</td>
            <td class="px-6 py-4">
              <span :class="statusClass(ret.status)" class="px-2 py-1 text-xs font-medium rounded-full">{{ statusLabel(ret.status) }}</span>
            </td>
            <td class="px-6 py-4 text-sm text-light">{{ formatDate(ret.createdAt) }}</td>
            <td class="px-6 py-4 text-right">
              <NuxtLink :to="localePath(`/customer/returns/${ret.id}`)" class="text-highlight hover:text-highlight-hover text-sm font-medium">
                {{ t('customer.returns.view') }}
              </NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'customer', middleware: ['auth'] })

const { t } = useI18n()
const { formatDate } = useDisplay()
const localePath = useLocalePath()
const api = useApi()

const returns = ref<any[]>([])
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
    const res = await api.get<any>('/user/returns')
    returns.value = Array.isArray(res) ? res : (res.data || [])
  } catch (e: any) {
    error.value = e?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
})
</script>
