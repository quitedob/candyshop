<template>
  <div class="space-y-6">
    <NuxtLink :to="localePath('/customer/invoices')" class="inline-flex items-center gap-1 text-sm text-orange-600 hover:text-orange-800">
      <Icon name="heroicons:arrow-left" class="h-4 w-4" />
      {{ $t('customer.invoices.back') }}
    </NuxtLink>

    <div v-if="pending" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-10 w-10 border-4 border-orange-200 border-t-orange-600"></div>
    </div>
    <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-xl p-6 text-red-700">{{ error }}</div>
    <div v-else-if="invoice" class="bg-white shadow rounded-lg p-6 space-y-4">
      <div class="flex items-start justify-between">
        <div>
          <h2 class="text-xl font-bold text-gray-900">{{ invoice.invoiceNumber || invoice.id }}</h2>
          <p class="text-sm text-gray-500 mt-1">{{ invoice.invoiceType || invoice.type }}</p>
        </div>
        <span :class="statusClass(invoice.status)" class="px-3 py-1 text-sm font-medium rounded-full">{{ invoice.status }}</span>
      </div>
      <div class="grid grid-cols-2 gap-4 text-sm">
        <div>
          <p class="text-gray-500">{{ $t('customer.invoices.order_id') }}</p>
          <p class="font-medium">{{ invoice.orderId || '-' }}</p>
        </div>
        <div>
          <p class="text-gray-500">{{ $t('customer.invoices.amount') }}</p>
          <p class="font-medium text-lg text-orange-600">{{ invoice.currency }} {{ formatNumber(invoice.totalAmount || 0) }}</p>
        </div>
        <div>
          <p class="text-gray-500">{{ $t('customer.invoices.due_date') }}</p>
          <p class="font-medium">{{ formatDate(invoice.dueDate) }}</p>
        </div>
        <div>
          <p class="text-gray-500">{{ $t('customer.invoices.issued_date') }}</p>
          <p class="font-medium">{{ formatDate(invoice.createdAt) }}</p>
        </div>
      </div>
      <div v-if="invoice.notes" class="pt-4 border-t border-gray-100">
        <p class="text-sm text-gray-500">{{ $t('customer.invoices.notes') }}</p>
        <p class="text-sm mt-1">{{ invoice.notes }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'customer', middleware: ['auth'] })

const { t } = useI18n()
const { formatNumber, formatDate } = useDisplay()
const localePath = useLocalePath()
const route = useRoute()
const api = useApi()
const invoice = ref<any>(null)
const pending = ref(true)
const error = ref('')

const fetchInvoice = async () => {
  try {
    invoice.value = await api.get(`/user/invoices/${route.params.id}`)
  } catch (e: any) {
    error.value = e?.message || t('customer.invoices.not_found')
  } finally {
    pending.value = false
  }
}


const statusClass = (status: string) => {
  const map: Record<string, string> = {
    paid: 'bg-green-100 text-green-800',
    pending: 'bg-yellow-100 text-yellow-800',
    overdue: 'bg-red-100 text-red-800',
  }
  return map[status?.toLowerCase()] || 'bg-gray-100 text-gray-800'
}

onMounted(fetchInvoice)
</script>
