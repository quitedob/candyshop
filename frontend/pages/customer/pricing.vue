<template>
  <div class="space-y-6">
    <div v-if="pending" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-10 w-10 border-4 border-orange-200 border-t-orange-600"></div>
    </div>
    <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-xl p-6 text-red-700">{{ error }}</div>
    <div v-else-if="!priceList" class="text-center py-16 text-gray-500">
      <Icon name="heroicons:tag" class="h-12 w-12 mx-auto mb-3 text-gray-300" />
      <p>{{ $t('customer.pricing.no_price_list') }}</p>
      <p class="text-sm mt-2">{{ $t('customer.pricing.contact_sales') }}</p>
    </div>
    <div v-else class="space-y-4">
      <div class="bg-white shadow rounded-lg p-6">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h2 class="text-lg font-semibold text-gray-900">{{ priceList.name }}</h2>
            <p class="text-sm text-gray-500">{{ priceList.description }}</p>
          </div>
          <span class="px-3 py-1 bg-green-100 text-green-800 text-xs font-medium rounded-full">{{ $t('customer.pricing.active') }}</span>
        </div>
        <div v-if="priceList.prices?.length" class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 text-sm">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ $t('customer.pricing.col_product') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ $t('customer.pricing.col_min_qty') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ $t('customer.pricing.col_unit_price') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ $t('customer.pricing.col_currency') }}</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-for="p in priceList.prices" :key="p.id">
                <td class="px-4 py-3 font-medium text-gray-900">
                  {{ productNames[p.productId] || p.productId }}
                </td>
                <td class="px-4 py-3 text-gray-500">{{ p.minQuantity || 1 }}</td>
                <td class="px-4 py-3 text-orange-600 font-semibold">{{ p.unitPrice }}</td>
                <td class="px-4 py-3 text-gray-500">{{ cur(p.currency) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-else class="text-sm text-gray-400 mt-2">{{ $t('customer.pricing.no_items') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'customer', middleware: ['auth'] })

const { t } = useI18n()
const { currencyOrDefault: cur } = useDisplay()
const api = useApi()
const priceList = ref<any>(null)
const pending = ref(true)
const error = ref('')
const productNames = ref<Record<string, string>>({})

const load = async () => {
  try {
    const res = await api.get('/user/price-list')
    priceList.value = res.priceList || null
    if (priceList.value?.prices?.length) {
      const ids: string[] = [...new Set(priceList.value.prices.map((p: any) => p.productId))]
      await Promise.all(ids.map(async (id: string) => {
        try {
          const product = await api.getProduct(id)
          if (product?.name) productNames.value[id] = product.name
        } catch { /* keep UUID as fallback */ }
      }))
    }
  } catch (e: any) {
    error.value = e?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

onMounted(load)
</script>
