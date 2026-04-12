<template>
  <div class="bg-white shadow overflow-hidden sm:rounded-lg">
    <div class="px-4 py-5 sm:px-6 border-b border-gray-200">
      <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.quotes.title') }}</h3>
      <p class="mt-1 text-sm text-gray-500">{{ t('customer.quotes.subtitle') }}</p>
    </div>

    <div v-if="pending" class="text-center py-10 text-gray-500">{{ t('customer.quotes.loading') }}</div>
    <div v-else-if="error" class="mx-4 my-4 rounded-md bg-red-50 p-4 text-sm text-red-700">{{ error }}</div>
    <div v-else-if="quotes.length === 0" class="text-center py-10 text-gray-500">
      {{ t('customer.quotes.no_quotes') }}
    </div>

    <div v-else class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">{{ t('customer.quotes.col_inquiry') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">{{ t('customer.quotes.col_amount') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">{{ t('customer.quotes.col_valid_until') }}</th>
            <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">{{ t('customer.quotes.col_status') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <tr v-for="quote in quotes" :key="quote.id">
            <td class="px-6 py-4 text-sm text-gray-900">
              <NuxtLink :to="`/customer/inquiries/${quote.inquiryId || quote.id}`" class="text-blue-600 hover:text-blue-900">{{ quote.id }}</NuxtLink>
            </td>
            <td class="px-6 py-4 text-sm text-gray-700">{{ formatCurrency(quote.quotedAmount) }}</td>
            <td class="px-6 py-4 text-sm text-gray-700">{{ formatDate(quote.validUntil) }}</td>
            <td class="px-6 py-4 text-sm">
              <span class="inline-flex rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-800">{{ quote.status }}</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()

const quotes = ref<any[]>([])
const pending = ref(true)
const error = ref('')

const fetchQuotes = async () => {
  pending.value = true; error.value = ''
  try {
    const res = await api.get<any>('/user/quotes?page=1&limit=50')
    quotes.value = res.data || []
  } catch (err: any) {
    error.value = err?.message || 'Failed to load quotes'
  } finally { pending.value = false }
}

const formatCurrency = (amount: number) => `USD ${Number(amount || 0).toLocaleString()}`
const formatDate = (value: string | null | undefined) => {
  if (!value) return t('customer.common.na')
  const dt = new Date(value)
  if (Number.isNaN(dt.getTime())) return t('customer.common.na')
  return dt.toLocaleDateString()
}

onMounted(fetchQuotes)
</script>