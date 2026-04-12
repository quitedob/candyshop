<template>
  <div class="space-y-6">
    <p class="text-sm text-gray-500">{{ $t('customer.shipments.subtitle') }}</p>

    <div v-if="pending" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-10 w-10 border-4 border-orange-200 border-t-orange-600"></div>
    </div>

    <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-xl p-6 text-center">
      <Icon name="heroicons:exclamation-triangle" class="h-8 w-8 text-red-400 mx-auto mb-2" />
      <p class="text-red-700 text-sm">{{ error }}</p>
      <button @click="load" class="mt-3 px-4 py-2 bg-red-100 text-red-700 rounded-lg text-sm hover:bg-red-200">
        {{ $t('customer.shipments.retry') }}
      </button>
    </div>

    <div v-else-if="!trades.length" class="text-center py-16 text-gray-500">
      <Icon name="heroicons:truck" class="h-12 w-12 mx-auto mb-3 text-gray-300" />
      <p>{{ $t('customer.shipments.no_trades') }}</p>
      <NuxtLink to="/customer/trades" class="mt-4 inline-block text-orange-600 hover:underline text-sm">
        {{ $t('customer.shipments.view_trades') }}
      </NuxtLink>
    </div>

    <div v-else class="space-y-4">
      <div v-for="trade in trades" :key="trade.id" class="bg-white shadow rounded-lg overflow-hidden">
        <div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between">
          <div>
            <h3 class="font-semibold text-gray-900">{{ $t('customer.shipments.trade_label') }} #{{ trade.id }}</h3>
            <p class="text-sm text-gray-500">{{ trade.status }} · {{ trade.incoterms || trade.terms }}</p>
          </div>
          <NuxtLink :to="`/customer/trades/${trade.id}`" class="text-sm text-orange-600 hover:text-orange-800">
            {{ $t('customer.shipments.view_trade') }}
          </NuxtLink>
        </div>
        <div class="p-6">
          <div v-if="shipmentMap[trade.id]?.length" class="space-y-3">
            <div v-for="s in shipmentMap[trade.id]" :key="s.id" class="flex items-start gap-4 text-sm">
              <div class="flex-shrink-0 w-2 h-2 mt-1.5 rounded-full bg-orange-400"></div>
              <div>
                <p class="font-medium text-gray-900">{{ s.carrier }} · {{ s.trackingNumber }}</p>
                <p class="text-gray-500">{{ s.status }} — {{ formatDate(s.updatedAt) }}</p>
                <p v-if="s.estimatedDelivery" class="text-gray-400 text-xs">ETA: {{ formatDate(s.estimatedDelivery) }}</p>
              </div>
            </div>
          </div>
          <p v-else class="text-sm text-gray-400">{{ $t('customer.shipments.no_shipments') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'customer', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const trades = ref<any[]>([])
const shipmentMap = ref<Record<string, any[]>>({})
const pending = ref(true)
const error = ref('')

const load = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await api.get<any>('/user/trades?page=1&limit=20')
    trades.value = res.data || []
    await Promise.all(
      trades.value.map(async (trade: any) => {
        try {
          const s = await api.get<any>(`/user/trades/${trade.id}/shipments`)
          shipmentMap.value[trade.id] = s.data || []
        } catch {
          shipmentMap.value[trade.id] = []
        }
      })
    )
  } catch (err: any) {
    error.value = err?.message || t('customer.shipments.load_error')
  } finally {
    pending.value = false
  }
}

const formatDate = (d: string) => d ? new Date(d).toLocaleDateString() : '-'

onMounted(load)
</script>
