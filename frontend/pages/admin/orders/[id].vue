<template>
  <div>
    <div class="mb-6">
      <NuxtLink :to="localePath('/admin/orders')" class="flex items-center text-sm font-medium text-orange-600 hover:text-orange-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('admin.orders.back') }}
      </NuxtLink>
    </div>

    <div v-if="pending" class="text-center py-10">
      <p class="text-gray-500">{{ t('admin.orders.loading_details') }}</p>
    </div>

    <div v-else-if="error" class="bg-red-50 p-4 rounded-md">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <div v-else-if="order" class="space-y-6">
      <!-- Order Info Card -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 flex justify-between items-center bg-gray-50 border-b border-gray-200">
          <div>
            <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.order_prefix') }}{{ order.orderNumber || String(order.id).substring(0, 8) }}</h3>
            <p class="mt-1 max-w-2xl text-sm text-gray-500">
              {{ t('admin.orders.placed_on') }} {{ formatDate(order.createdAt) }}
            </p>
          </div>
          <div class="flex items-center gap-3">
            <span :class="[statusClass(order.status), 'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5']">
              {{ enumLabel('order_status', order.status) }}
            </span>
            <span :class="[paymentStatusClass(order.paymentStatus), 'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5']">
              {{ t('admin.orders.payment') }}: {{ enumLabel('payment_status', order.paymentStatus, 'unpaid') }}
            </span>
          </div>
        </div>

        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.orders.customer') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">
                <template v-if="order.user">
                  {{ order.user.firstName }} {{ order.user.lastName }}
                  <span class="text-gray-500">({{ order.user.email }})</span>
                </template>
                <template v-else>{{ cell(order.userId) }}</template>
              </dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.orders.currency') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ cur(order.currency) }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.orders.subtotal') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ formatNumber(order.subtotal || 0) }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.orders.tax') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ formatNumber(order.taxAmount || 0) }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.orders.shipping') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ formatNumber(order.shippingAmount || 0) }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.orders.total') }}</dt>
              <dd class="mt-1 text-sm font-medium text-gray-900">{{ formatNumber(order.totalAmount || 0) }}</dd>
            </div>
            <div v-if="order.trackingNumber">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.orders.tracking_number') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ order.trackingNumber }}</dd>
            </div>
            <div v-if="order.inquiryId">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.orders.inquiry') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ order.inquiryId }}</dd>
            </div>
            <div v-if="order.shippingAddress" class="sm:col-span-2">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.orders.shipping_address') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">
                {{ formatAddress(order.shippingAddress) }}
              </dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Order Items -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.orders.items') }}</h3>
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('admin.orders.col_product') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('admin.orders.col_quantity') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('admin.orders.col_unit_price') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{{ t('admin.orders.col_subtotal') }}</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-if="!order.items || order.items.length === 0">
                <td colspan="4" class="px-6 py-4 text-center text-sm text-gray-500">{{ t('admin.orders.no_items') }}</td>
              </tr>
              <tr v-else v-for="(item, idx) in order.items" :key="idx">
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{ item.productName || item.productId }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ item.quantity }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ cur(order.currency) }} {{ formatNumber(item.unitPrice || 0) }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ cur(order.currency) }} {{ formatNumber((item.quantity || 0) * (item.unitPrice || 0)) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Create Trade Button -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200 flex items-center justify-between">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.orders.trade_section') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6 flex items-center gap-4">
          <button @click="createTrade" :disabled="creatingTrade"
            class="inline-flex items-center gap-2 px-4 py-2 bg-orange-600 text-white text-sm font-medium rounded-md hover:bg-orange-700 disabled:opacity-50 transition-colors">
            <Icon v-if="creatingTrade" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" />
            <Icon v-else name="heroicons:arrow-path-rounded-square" class="h-4 w-4" />
            {{ creatingTrade ? t('admin.orders.creating_trade') : t('admin.orders.create_trade') }}
          </button>
          <p v-if="tradeMessage" class="text-sm" :class="tradeError ? 'text-red-600' : 'text-green-600'">{{ tradeMessage }}</p>
          <NuxtLink v-if="createdTradeId" :to="localePath(`/admin/trades/${createdTradeId}`)"
            class="text-sm text-orange-600 hover:text-amber-800 font-medium">
            {{ t('admin.orders.view_trade') }} →
          </NuxtLink>
        </div>
      </div>

      <!-- Status Update Section -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.orders.update_status') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <form @submit.prevent="updateStatus" class="flex flex-wrap items-end gap-4">
            <div>
              <label for="order-status" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.status') }}</label>
              <select id="order-status" name="status" v-model="statusInput" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option value="pending">{{ enumLabel('order_status', 'pending') }}</option>
                <option value="confirmed">{{ enumLabel('order_status', 'confirmed') }}</option>
                <option value="production">{{ enumLabel('order_status', 'production') }}</option>
                <option value="shipped">{{ enumLabel('order_status', 'shipped') }}</option>
                <option value="delivered">{{ enumLabel('order_status', 'delivered') }}</option>
                <option value="cancelled">{{ enumLabel('order_status', 'cancelled') }}</option>
              </select>
            </div>
            <div>
              <label for="order-trackingNumber" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.tracking_number') }}</label>
              <input id="order-trackingNumber" v-model="trackingInput" name="trackingNumber" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <button type="submit" :disabled="updatingStatus" class="inline-flex justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700 disabled:opacity-50">
              {{ updatingStatus ? t('admin.orders.updating') : t('admin.orders.update') }}
            </button>
          </form>
          <div v-if="statusMessage" class="mt-3 text-sm" :class="statusError ? 'text-red-600' : 'text-green-600'">
            {{ statusMessage }}
          </div>
          <div v-if="paymentPolicyWarning" class="mt-3 p-3 bg-yellow-50 rounded-md text-sm text-yellow-800">
            {{ paymentPolicyWarning }}
          </div>
        </div>
      </div>

      <!-- Payment Management -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.orders.payment_management') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <div v-if="loadingPayments" class="text-center py-4">
            <p class="text-gray-500">{{ t('admin.orders.loading_payments') }}</p>
          </div>
          <div v-else-if="paymentError" class="text-red-600 text-sm">{{ paymentError }}</div>
          <div v-else>
            <div v-if="payments.length === 0" class="text-center py-4 text-gray-500">
              {{ t('admin.orders.no_payments') }}
            </div>
            <table v-else class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.col_id') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.orders.amount') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.orders.method') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.orders.status') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.orders.date') }}</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.orders.actions') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
                <tr v-for="payment in payments" :key="payment.id">
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ String(payment.id).substring(0, 8) }}</td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{ cur(order.currency) }} {{ formatNumber(payment.amount || 0) }}</td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ cell(payment.method) }}</td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <span :class="[paymentStatusBadge(payment.status), 'inline-flex rounded-full px-2 text-xs font-semibold leading-5']">
                      {{ enumLabel('payment_status', payment.status) }}
                    </span>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ formatDate(payment.createdAt) }}</td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <button v-if="payment.status === 'pending'" type="button" class="text-green-600 hover:text-green-900 mr-3" @click="confirmPayment(payment.id)">
                      {{ t('admin.orders.confirm') }}
                    </button>
                    <button v-if="payment.status === 'confirmed'" type="button" class="text-orange-600 hover:text-orange-900" @click="refundPayment(payment.id)">
                      {{ t('admin.orders.refund') }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Order Messages -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.orders.messages_title') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <div class="bg-gray-50 rounded-lg p-4 max-h-80 overflow-y-auto space-y-3 mb-4" ref="msgListRef">
            <div v-if="messages.length === 0" class="text-center text-sm text-gray-500 py-4">
              {{ t('admin.orders.no_messages') }}
            </div>
            <div v-for="msg in messages" :key="msg.id" :class="['flex', msg.senderType === 'admin' ? 'justify-end' : 'justify-start']">
              <div :class="['max-w-xs lg:max-w-md rounded-lg px-4 py-2', msg.senderType === 'admin' ? 'bg-orange-600 text-white' : 'bg-white border border-gray-200 text-gray-900']">
                <div class="text-xs font-semibold mb-1">
                  {{ msg.senderType === 'admin' ? t('admin.orders.sender_admin') : t('admin.orders.sender_customer') }}
                </div>
                <p class="text-sm whitespace-pre-wrap">{{ msg.message }}</p>
                <p class="text-xs mt-1 opacity-70">{{ formatDate(msg.createdAt) }}</p>
              </div>
            </div>
          </div>
          <div class="flex gap-2">
            <input v-model="newMessage" @keyup.enter="sendMessage" :placeholder="t('admin.orders.message_placeholder')"
              class="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm" :disabled="sending" />
            <button @click="sendMessage" :disabled="sending || !newMessage.trim()"
              class="inline-flex items-center gap-1 rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50">
              {{ sending ? '...' : t('admin.orders.send_message') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Timeline / Activity Log -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.orders.activity_log') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <div v-if="!order.statusHistory || order.statusHistory.length === 0" class="text-center text-gray-500">
            {{ t('admin.orders.no_activity') }}
          </div>
          <ul v-else class="border-l-2 border-gray-200 ml-3 space-y-4">
            <li v-for="(entry, idx) in order.statusHistory" :key="idx" class="ml-4">
              <div class="flex items-start">
                <div class="flex-shrink-0 w-2 h-2 mt-2 rounded-full bg-gray-400"></div>
                <div class="ml-3">
                  <p class="text-sm text-gray-900">{{ entry.status }} {{ entry.note ? `- ${entry.note}` : '' }}</p>
                  <p class="text-xs text-gray-500">{{ formatDate(entry.timestamp) }}</p>
                </div>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, nextTick } from 'vue'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const route = useRoute()
const api = useApi()
const { t, te } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, cell, enumLabel, formatNumber, formatDate } = useDisplay()

const order = ref<any>(null)
const pending = ref(true)
const error = ref('')
const statusInput = ref('')
const trackingInput = ref('')
const updatingStatus = ref(false)
const statusMessage = ref('')
const statusError = ref(false)
const payments = ref<any[]>([])
const loadingPayments = ref(false)
const paymentError = ref('')
const creatingTrade = ref(false)
const tradeMessage = ref('')
const tradeError = ref(false)
const createdTradeId = ref<number | null>(null)
const messages = ref<any[]>([])
const newMessage = ref('')
const sending = ref(false)
const msgListRef = ref<HTMLElement | null>(null)

const fetchMessages = async () => {
  if (!route.params.id) return
  try {
    const res = await api.get<any>(`/admin/orders/${route.params.id}/messages`)
    messages.value = res.data || []
  } catch (_) {}
}

const sendMessage = async () => {
  if (!route.params.id || !newMessage.value.trim()) return
  sending.value = true
  try {
    const msg = await api.post<any>(`/admin/orders/${route.params.id}/messages`, { message: newMessage.value.trim() })
    messages.value.push(msg)
    newMessage.value = ''
    nextTick(() => { if (msgListRef.value) msgListRef.value.scrollTop = msgListRef.value.scrollHeight })
  } catch (err: any) {
    alert(err?.message || t('admin.orders.message_send_error'))
  } finally { sending.value = false }
}

let msgInterval: ReturnType<typeof setInterval> | null = null
onMounted(() => { fetchOrder(); fetchPayments(); fetchMessages(); msgInterval = setInterval(fetchMessages, 30000) })
onUnmounted(() => { if (msgInterval) clearInterval(msgInterval) })

const paymentPolicyWarning = computed(() => {
  const key = `admin.payment_policy.${statusInput.value}`
  return te(key) ? t(key) : ''
})

const fetchOrder = async () => {
  pending.value = true; error.value = ''
  try {
    order.value = await api.get<any>(`/admin/orders/${route.params.id}`)
    statusInput.value = order.value.status || 'pending'
    trackingInput.value = order.value.trackingNumber || ''
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally { pending.value = false }
}

const fetchPayments = async () => {
  loadingPayments.value = true; paymentError.value = ''
  try {
    payments.value = await api.get<any[]>(`/admin/orders/${route.params.id}/payments`) || []
  } catch (err: any) {
    paymentError.value = err?.message || t('errors.api.fetch_payments_failed')
  } finally { loadingPayments.value = false }
}

const updateStatus = async () => {
  updatingStatus.value = true; statusMessage.value = ''; statusError.value = false
  try {
    const payload: Record<string, any> = { status: statusInput.value }
    if (trackingInput.value.trim()) payload.trackingNumber = trackingInput.value.trim()
    await api.put(`/admin/orders/${route.params.id}`, payload)
    order.value.status = statusInput.value
    if (trackingInput.value.trim()) order.value.trackingNumber = trackingInput.value.trim()
    statusMessage.value = t('admin.orders.status_updated')
    setTimeout(() => { statusMessage.value = '' }, 3000)
  } catch (err: any) {
    statusError.value = true
    statusMessage.value = err?.message || t('errors.api.status_failed')
  } finally { updatingStatus.value = false }
}

const confirmPayment = async (paymentId: string) => {
  try {
    await api.put(`/admin/orders/${route.params.id}/payments/${paymentId}/confirm`, {})
    await fetchPayments()
  } catch (err: any) { alert(err?.message || t('errors.api.payment_confirm_failed')) }
}

const refundPayment = async (paymentId: string) => {
  if (!confirm(t('admin.orders.confirm_refund'))) return
  try {
    await api.put(`/admin/orders/${route.params.id}/payments/${paymentId}/refund`, {})
    await fetchPayments()
  } catch (err: any) { alert(err?.message || t('errors.api.refund_failed')) }
}

const createTrade = async () => {
  creatingTrade.value = true; tradeMessage.value = ''; tradeError.value = false
  try {
    const res = await api.post<any>(`/admin/orders/${route.params.id}/create-trade`, {
      terms: t('defaults.incoterms'),
      currency: cur(order.value?.currency)
    })
    createdTradeId.value = res.trade?.id || null
    tradeMessage.value = t('admin.orders.trade_created')
  } catch (err: any) {
    tradeError.value = true
    tradeMessage.value = err?.message || t('errors.api.trade_create_failed')
  } finally { creatingTrade.value = false }
}

const statusClass = (status: string) => {
  if (status === 'pending') return 'bg-yellow-100 text-yellow-800'
  if (status === 'confirmed' || status === 'production') return 'bg-orange-100 text-orange-800'
  if (status === 'shipped') return 'bg-amber-100 text-amber-800'
  if (status === 'delivered') return 'bg-green-100 text-green-800'
  if (status === 'cancelled') return 'bg-red-100 text-red-800'
  return 'bg-gray-100 text-gray-800'
}

const paymentStatusClass = (status: string) => {
  if (status === 'paid') return 'bg-green-100 text-green-800'
  if (status === 'partial') return 'bg-yellow-100 text-yellow-800'
  if (status === 'refunded') return 'bg-orange-100 text-orange-800'
  return 'bg-gray-100 text-gray-800'
}

const paymentStatusBadge = (status: string) => {
  if (status === 'pending') return 'bg-yellow-100 text-yellow-800'
  if (status === 'confirmed' || status === 'paid') return 'bg-green-100 text-green-800'
  if (status === 'refunded') return 'bg-orange-100 text-orange-800'
  return 'bg-gray-100 text-gray-800'
}

const formatAddress = (address: any) => {
  if (!address) return '-'
  return [address.street, address.city, address.state, address.zipCode, address.country].filter(Boolean).join(', ')
}

</script>
