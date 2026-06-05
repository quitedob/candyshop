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
            <span :class="[order.complianceOfficialEvidence ? 'bg-emerald-100 text-emerald-800' : 'bg-amber-100 text-amber-800', 'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5']">
              {{ order.complianceOfficialEvidence ? t('admin.orders.compliance_verified') : t('admin.orders.compliance_manual') }}
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
            <div v-if="order.cogs != null && order.cogs !== undefined">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.orders.cogs') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ formatNumber(order.cogs || 0) }}</dd>
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
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{ tField(item, 'productName') || item.productId }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ item.quantity }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ cur(order.currency) }} {{ formatNumber(item.unitPrice || 0) }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ cur(order.currency) }} {{ formatNumber((item.quantity || 0) * (item.unitPrice || 0)) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-if="order.inventoryWarnings?.length" class="px-6 py-4 bg-amber-50 border-t border-amber-100">
          <p class="text-sm font-medium text-amber-900 mb-2">{{ t('admin.orders.inventory_warnings') }}</p>
          <ul class="list-disc list-inside text-sm text-amber-800 space-y-1">
            <li v-for="(w, wi) in order.inventoryWarnings" :key="wi">{{ w }}</li>
          </ul>
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
          <div v-if="showPaymentRequiredBanner" class="mb-4 p-3 bg-amber-50 border border-amber-200 rounded-md text-sm text-amber-900 flex items-start justify-between gap-2">
            <span>{{ t('admin.payment_policy.requires_payment_before_execution') }}</span>
          </div>
          <form @submit.prevent="updateStatus" class="flex flex-wrap items-end gap-4">
            <div>
              <label for="order-status" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.status') }}</label>
              <select id="order-status" name="status" v-model="statusInput" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option
                  v-for="opt in statusOptions"
                  :key="opt.value"
                  :value="opt.value"
                  :disabled="opt.disabled"
                  :title="opt.paymentBlocked ? t('admin.payment_policy.requires_payment_before_execution') : undefined"
                >
                  {{ enumLabel('order_status', opt.value) }}{{ opt.paymentBlocked ? ' ⛔' : '' }}
                </option>
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
          <div v-if="statusMessage" class="mt-3 text-sm flex items-start justify-between gap-2" :class="statusError ? 'text-red-600' : 'text-green-600'">
            <span>{{ statusMessage }}</span>
            <button v-if="statusError" type="button" class="text-xs underline shrink-0" @click="statusMessage = ''">{{ t('admin.orders.dismiss') }}</button>
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
                    <!-- M-16: also allow capture from authorized state (Stripe manual-capture flow) -->
                    <button v-if="payment.status === 'authorized'" type="button" class="text-emerald-600 hover:text-emerald-900 mr-3" @click="confirmPayment(payment.id)">
                      {{ t('admin.orders.confirm') }}
                    </button>
                    <button v-if="payment.status === 'confirmed' || payment.status === 'partial_refund'" type="button" class="text-orange-600 hover:text-orange-900" @click="refundPayment(payment.id)">
                      {{ t('admin.orders.refund') }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Fulfillments — M-15 -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200 flex items-center justify-between">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.orders.fulfillments_title') }}</h3>
          <button type="button"
                  class="text-sm text-orange-600 hover:text-orange-900"
                  :disabled="fulfillmentSubmitting"
                  @click="openFulfillmentDialog">
            {{ t('admin.orders.fulfillments_create') }}
          </button>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <p v-if="!fulfillments.length" class="text-sm text-gray-500">{{ t('admin.orders.fulfillments_empty') }}</p>
          <table v-else class="min-w-full divide-y divide-gray-200 text-sm">
            <thead>
              <tr>
                <th class="px-3 py-2 text-left text-xs font-semibold text-gray-500">{{ t('admin.orders.fulfillment_id') }}</th>
                <th class="px-3 py-2 text-left text-xs font-semibold text-gray-500">{{ t('admin.orders.fulfillment_status') }}</th>
                <th class="px-3 py-2 text-left text-xs font-semibold text-gray-500">{{ t('admin.orders.fulfillment_warehouse') }}</th>
                <th class="px-3 py-2 text-left text-xs font-semibold text-gray-500">{{ t('admin.orders.fulfillment_tracking') }}</th>
                <th class="px-3 py-2 text-left text-xs font-semibold text-gray-500">{{ t('admin.orders.fulfillment_actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="f in fulfillments" :key="f.id">
                <td class="px-3 py-2 text-gray-700">{{ String(f.id).substring(0, 8) }}</td>
                <td class="px-3 py-2">
                  <span class="inline-flex rounded-full px-2 text-xs font-semibold leading-5 bg-gray-100 text-gray-800">{{ f.status }}</span>
                </td>
                <td class="px-3 py-2 text-gray-500">{{ cell(f.warehouseId) }}</td>
                <td class="px-3 py-2 text-gray-500">{{ cell(f.trackingNumber) }}</td>
                <td class="px-3 py-2 text-sm">
                  <button v-if="f.status === 'pending' || f.status === 'picking'"
                          type="button"
                          class="text-orange-600 hover:text-orange-900 mr-2"
                          @click="openShipDialog(f)">
                    {{ t('admin.orders.fulfillment_ship') }}
                  </button>
                  <button v-if="f.status === 'shipped'"
                          type="button"
                          class="text-emerald-600 hover:text-emerald-900 mr-2"
                          @click="markFulfillmentDelivered(f)">
                    {{ t('admin.orders.fulfillment_deliver') }}
                  </button>
                  <button v-if="f.status === 'pending' || f.status === 'picking'"
                          type="button"
                          class="text-red-600 hover:text-red-900"
                          @click="cancelFulfillment(f)">
                    {{ t('admin.orders.fulfillment_cancel') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          <p v-if="fulfillmentMessage" class="mt-3 text-sm text-emerald-600">{{ fulfillmentMessage }}</p>
          <p v-if="fulfillmentError" class="mt-3 text-sm text-red-600">{{ fulfillmentError }}</p>
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
                <span
                  v-if="msg.senderType === 'admin'"
                  class="block text-right text-xs opacity-70"
                  :class="{ 'text-emerald-200': msg.readAt }"
                  :title="msg.readAt ? formatDate(msg.readAt) : ''"
                  aria-hidden="true"
                >{{ msg.readAt ? '✓✓' : '✓' }}</span>
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
          <div v-if="!activityEntries.length" class="text-center text-gray-500">
            {{ t('admin.orders.no_activity') }}
          </div>
          <ul v-else class="border-l-2 border-gray-200 ml-3 space-y-4">
            <li v-for="(entry, idx) in activityEntries" :key="idx" class="ml-4">
              <div class="flex items-start">
                <div class="flex-shrink-0 w-2 h-2 mt-2 rounded-full" :class="activityDotClass(entry.type)"></div>
                <div class="ml-3">
                  <p class="text-sm text-gray-900">{{ activityLabel(entry) }}</p>
                  <p v-if="entry.detail && entry.type === 'status'" class="text-xs text-gray-600">{{ entry.detail }}</p>
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
import { useTranslation } from '~/composables/useTranslation'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const route = useRoute()
const api = useApi()
const { t, te } = useI18n()
const { tField } = useTranslation()
const localePath = useLocalePath()
const { currencyOrDefault: cur, cell, enumLabel, formatNumber, formatDate } = useDisplay()
const { statusSelectOptions, isPaidOrPartial, paymentInsufficientForExecution } = useOrderStatusTransitions()

const order = ref<any>(null)
const pending = ref(true)
const error = ref('')
const statusInput = ref('')
const trackingInput = ref('')
const updatingStatus = ref(false)
const statusMessage = ref('')
const statusError = ref(false)
const payments = ref<any[]>([])
const fulfillments = ref<any[]>([])
const fulfillmentMessage = ref('')
const fulfillmentError = ref('')
const fulfillmentSubmitting = ref(false)
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
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  }
}

const markMessagesRead = async () => {
  if (!route.params.id) return
  try {
    await api.post(`/admin/orders/${route.params.id}/messages/read`, {})
  } catch {
    // A reconnect REST fetch also marks messages read.
  }
}

const appendMessage = (msg: any) => {
  if (!msg?.id || messages.value.some((existing) => existing.id === msg.id)) return
  messages.value.push(msg)
  nextTick(() => { if (msgListRef.value) msgListRef.value.scrollTop = msgListRef.value.scrollHeight })
}

const applyReadReceipt = (payload: any) => {
  if (!payload?.readerType || !payload?.readAt) return
  for (const msg of messages.value) {
    if (msg.senderType !== payload.readerType && !msg.readAt) msg.readAt = payload.readAt
  }
}

const messageStream = useOrderMessageStream(
  `/admin/orders/${route.params.id}/messages/ws`,
  (event) => {
    if (event.type === 'order_message') {
      appendMessage(event.payload)
      if (event.payload?.senderType === 'customer') void markMessagesRead()
    } else if (event.type === 'order_messages_read') {
      applyReadReceipt(event.payload)
    }
  },
  fetchMessages,
)

const sendMessage = async () => {
  if (!route.params.id || !newMessage.value.trim()) return
  sending.value = true
  try {
    const msg = await api.post<any>(`/admin/orders/${route.params.id}/messages`, { message: newMessage.value.trim() })
    appendMessage(msg)
    newMessage.value = ''
    nextTick(() => { if (msgListRef.value) msgListRef.value.scrollTop = msgListRef.value.scrollHeight })
  } catch (err: any) {
    notifyError(err, t('admin.orders.message_send_error'))
  } finally { sending.value = false }
}

onMounted(() => { fetchOrder(); fetchPayments(); fetchFulfillments(); fetchMessages(); messageStream.connect() })
onUnmounted(messageStream.stop)

const paymentPolicyWarning = computed(() => {
  const key = `admin.payment_policy.${statusInput.value}`
  return te(key) ? t(key) : ''
})

const statusOptions = computed(() => {
  if (!order.value) return []
  return statusSelectOptions(order.value.status || 'pending', order.value.paymentStatus || 'unpaid')
})

const showPaymentRequiredBanner = computed(() => {
  if (!order.value) return false
  return !isPaidOrPartial(order.value.paymentStatus || 'unpaid')
})

const activityEntries = computed(() => {
  const hist = order.value?.activityHistory
  if (Array.isArray(hist) && hist.length) return hist
  const legacy = order.value?.statusHistory || []
  return legacy.map((e: any) => ({ type: 'status', label: e.status, detail: e.note, timestamp: e.timestamp }))
})

const activityLabel = (entry: any) => {
  if (entry.type === 'status') return enumLabel('order_status', entry.label || entry.status)
  const key = `admin.orders.activity_${entry.type}`
  const base = te(key) ? t(key) : entry.label
  if (entry.amount != null && entry.amount > 0) {
    return `${base} — ${cur(order.value?.currency)} ${formatNumber(entry.amount)}`
  }
  return base
}

const activityDotClass = (type: string) => {
  if (type?.startsWith('payment')) return 'bg-green-500'
  if (type === 'invoice_auto_created') return 'bg-blue-500'
  return 'bg-gray-400'
}

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

// M-15: load and act on fulfillments. The backend exposes
//   GET    /admin/orders/:id/fulfillments
//   POST   /admin/orders/:id/fulfillments
//   PUT    /admin/fulfillments/:id/{ship,deliver,cancel}
// but had no admin UI surface; this section adds the missing controls.
const fetchFulfillments = async () => {
  fulfillmentError.value = ''
  try {
    const res = await api.get<any[]>(`/admin/orders/${route.params.id}/fulfillments`)
    fulfillments.value = Array.isArray(res) ? res : []
  } catch (err: any) {
    fulfillmentError.value = err?.message || t('errors.api.load_failed')
  }
}

const openFulfillmentDialog = async () => {
  if (!order.value || !Array.isArray(order.value.items) || order.value.items.length === 0) {
    fulfillmentError.value = t('admin.orders.fulfillments_no_items')
    return
  }
  // Default to creating a fulfillment that ships every line in full.
  const items = order.value.items.map((it: any, idx: number) => ({
    orderItemIdx: idx,
    productId: String(it.productId),
    quantity: Math.max(1, Number(it.quantity || 0) - Number(it.fulfilledQuantity || 0)),
  })).filter((it: any) => it.quantity > 0)
  if (items.length === 0) {
    fulfillmentError.value = t('admin.orders.fulfillments_already_fulfilled')
    return
  }
  fulfillmentSubmitting.value = true
  fulfillmentMessage.value = ''
  fulfillmentError.value = ''
  try {
    await api.post(`/admin/orders/${route.params.id}/fulfillments`, {
      warehouseId: order.value.warehouseId || '',
      notes: '',
      items,
    })
    fulfillmentMessage.value = t('admin.orders.fulfillments_created')
    await fetchFulfillments()
  } catch (err: any) {
    fulfillmentError.value = err?.message || t('errors.api.save_failed')
  } finally {
    fulfillmentSubmitting.value = false
  }
}

const openShipDialog = async (f: any) => {
  const trackingNumber = window.prompt(t('admin.orders.fulfillment_tracking_prompt'))
  if (!trackingNumber || !trackingNumber.trim()) return
  const carrier = window.prompt(t('admin.orders.fulfillment_carrier_prompt')) || ''
  fulfillmentSubmitting.value = true
  fulfillmentMessage.value = ''
  fulfillmentError.value = ''
  try {
    await api.put(`/admin/fulfillments/${f.id}/ship`, {
      trackingNumber: trackingNumber.trim(),
      carrier: carrier.trim(),
    })
    fulfillmentMessage.value = t('admin.orders.fulfillment_ship_success')
    await fetchFulfillments()
    await fetchOrder()
  } catch (err: any) {
    fulfillmentError.value = err?.message || t('errors.api.save_failed')
  } finally {
    fulfillmentSubmitting.value = false
  }
}

const markFulfillmentDelivered = async (f: any) => {
  if (!window.confirm(t('admin.orders.fulfillment_deliver_confirm'))) return
  fulfillmentSubmitting.value = true
  try {
    await api.put(`/admin/fulfillments/${f.id}/deliver`, {})
    fulfillmentMessage.value = t('admin.orders.fulfillment_deliver_success')
    await fetchFulfillments()
    await fetchOrder()
  } catch (err: any) {
    fulfillmentError.value = err?.message || t('errors.api.save_failed')
  } finally {
    fulfillmentSubmitting.value = false
  }
}

const cancelFulfillment = async (f: any) => {
  if (!window.confirm(t('admin.orders.fulfillment_cancel_confirm'))) return
  fulfillmentSubmitting.value = true
  try {
    await api.put(`/admin/fulfillments/${f.id}/cancel`, {})
    fulfillmentMessage.value = t('admin.orders.fulfillment_cancel_success')
    await fetchFulfillments()
  } catch (err: any) {
    fulfillmentError.value = err?.message || t('errors.api.save_failed')
  } finally {
    fulfillmentSubmitting.value = false
  }
}

const updateStatus = async () => {
  const status = String(statusInput.value || '').toLowerCase().trim()
  if (paymentInsufficientForExecution(status, order.value?.paymentStatus || 'unpaid')) {
    statusError.value = true
    statusMessage.value = t('admin.payment_policy.requires_payment_before_execution')
    return
  }
  updatingStatus.value = true; statusMessage.value = ''; statusError.value = false
  try {
    const payload: Record<string, any> = { status }
    if (trackingInput.value.trim()) payload.trackingNumber = trackingInput.value.trim()
    await api.put(`/admin/orders/${route.params.id}/status`, payload)
    await fetchOrder()
    statusMessage.value = t('admin.orders.status_updated')
    setTimeout(() => { if (!statusError.value) statusMessage.value = '' }, 3000)
  } catch (err: any) {
    statusError.value = true
    statusMessage.value = err?.message || t('errors.api.status_failed')
  } finally { updatingStatus.value = false }
}

const confirmPayment = async (paymentId: string) => {
  try {
    await api.put(`/admin/orders/${route.params.id}/payments/${paymentId}/confirm`, {})
    await fetchPayments()
    await fetchOrder()
  } catch (err: any) { notifyError(err, t('errors.api.payment_confirm_failed')) }
}

const refundPayment = async (paymentId: string) => {
  if (!confirm(t('admin.orders.confirm_refund'))) return
  try {
    await api.put(`/admin/orders/${route.params.id}/payments/${paymentId}/refund`, {})
    await fetchPayments()
  } catch (err: any) { notifyError(err, t('errors.api.refund_failed')) }
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
  // H-9: cover all 13 order statuses defined on the backend (see models/order/order.go).
  // Previously only 6 statuses had explicit colours and the rest displayed as gray badges.
  switch (status) {
    case 'pending_confirmation':
      return 'bg-orange-100 text-orange-800'
    case 'pending_approval':
      return 'bg-purple-100 text-purple-800'
    case 'pending':
      return 'bg-yellow-100 text-yellow-800'
    case 'confirmed':
      return 'bg-blue-100 text-blue-800'
    case 'production':
      return 'bg-orange-100 text-orange-800'
    case 'partially_shipped':
      return 'bg-amber-100 text-amber-800'
    case 'shipped':
      return 'bg-amber-100 text-amber-800'
    case 'partially_delivered':
      return 'bg-emerald-100 text-emerald-800'
    case 'delivered':
      return 'bg-green-100 text-green-800'
    case 'partially_returned':
      return 'bg-rose-100 text-rose-800'
    case 'returned':
      return 'bg-rose-200 text-rose-900'
    case 'cancelled':
      return 'bg-red-100 text-red-800'
    case 'expired':
      return 'bg-gray-200 text-gray-700'
    default:
      return 'bg-gray-100 text-gray-800'
  }
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
