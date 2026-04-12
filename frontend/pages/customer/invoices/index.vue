<template>
  <div class="space-y-6">
    <div v-if="pending" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-10 w-10 border-4 border-orange-200 border-t-orange-600"></div>
    </div>
    <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-xl p-6 text-center text-red-700">{{ error }}</div>
    <div v-else-if="!invoices.length" class="text-center py-16 text-gray-500">
      <Icon name="heroicons:document-text" class="h-12 w-12 mx-auto mb-3 text-gray-300" />
      <p>{{ $t('customer.invoices.no_invoices') }}</p>
    </div>
    <div v-else class="bg-white shadow rounded-lg overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ $t('customer.invoices.col_number') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ $t('customer.invoices.col_type') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ $t('customer.invoices.col_amount') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ $t('customer.invoices.col_status') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ $t('customer.invoices.col_date') }}</th>
            <th class="px-6 py-3"></th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="inv in invoices" :key="inv.id">
            <td class="px-6 py-4 text-sm font-medium text-gray-900">{{ inv.invoiceNumber || inv.id }}</td>
            <td class="px-6 py-4 text-sm text-gray-500">{{ inv.invoiceType || inv.type }}</td>
            <td class="px-6 py-4 text-sm text-gray-900">{{ inv.currency }} {{ (inv.totalAmount || 0).toLocaleString() }}</td>
            <td class="px-6 py-4">
              <span :class="statusClass(inv.status)" class="px-2 py-1 text-xs font-medium rounded-full">{{ inv.status }}</span>
            </td>
            <td class="px-6 py-4 text-sm text-gray-500">{{ formatDate(inv.createdAt) }}</td>
            <td class="px-6 py-4 text-right">
              <NuxtLink :to="`/customer/invoices/${inv.id}`" class="text-orange-600 hover:text-orange-800 text-sm font-medium">
                {{ $t('customer.invoices.view') }}
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

const api = useApi()
const invoices = ref<any[]>([])
const pending = ref(true)
const error = ref('')

const fetchInvoices = async () => {
  try {
    const res = await api.get('/user/invoices')
    invoices.value = res.data || []
  } catch (e: any) {
    error.value = e?.message || 'Failed to load invoices'
  } finally {
    pending.value = false
  }
}

const formatDate = (d: string) => d ? new Date(d).toLocaleDateString() : '-'

const statusClass = (status: string) => {
  const map: Record<string, string> = {
    paid: 'bg-green-100 text-green-800',
    pending: 'bg-yellow-100 text-yellow-800',
    overdue: 'bg-red-100 text-red-800',
  }
  return map[status?.toLowerCase()] || 'bg-gray-100 text-gray-800'
}

onMounted(fetchInvoices)
</script>
