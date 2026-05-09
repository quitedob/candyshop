<template>
  <div>
    <div class="mb-6">
      <NuxtLink :to="localePath('/customer/orders')" class="flex items-center text-sm font-medium text-highlight hover:text-highlight-hover">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('customer.orders.back') }}
      </NuxtLink>
    </div>
    <div v-if="pending" class="text-center py-10">
      <p class="text-light">{{ t('customer.orders.loading_details') }}</p>
    </div>
    <div v-else-if="error" class="bg-error-10 p-4 rounded-md">
      <p class="text-error">{{ error }}</p>
    </div>
    <div v-else-if="order" class="bg-white shadow-lg rounded-xl overflow-hidden">
      <div class="px-6 py-5 flex justify-between items-center bg-bg-alt border-b border-border">
        <div>
          <h3 class="text-lg leading-6 font-semibold text-primary">{{ t('customer.orders.detail_title', { id: order.orderNumber || order.id.substring(0, 8) }) }}</h3>
          <p class="mt-1 max-w-2xl text-sm text-light">{{ t('customer.orders.placed_on') }} {{ new Date(order.createdAt).toLocaleDateString() }}</p>
        </div>
        <span :class="[statusBadge(order.status), 'badge']">{{ enumLabel('order_status', order.status) }}</span>
      </div>
      <div v-if="order.status === 'pending_confirmation'" class="mx-6 mt-4 rounded-lg border border-warning bg-warning-10 px-4 py-3">
        <div class="flex items-start justify-between gap-4">
          <div>
            <p class="text-sm font-medium text-warning">{{ t('customer.orders.ai_draft_notice') }}</p>
            <p class="mt-1 text-sm text-warning">{{ t('customer.orders.ai_draft_review') }}</p>
            <div v-if="!order.complianceOfficialEvidence" class="mt-3 rounded-md border border-error bg-error-10 px-3 py-2">
              <p class="text-sm font-medium text-error">{{ t('customer.orders.compliance_required') }}</p>
              <p class="mt-1 text-sm text-error">{{ t('customer.orders.compliance_no_evidence') }}</p>
              <label class="mt-2 inline-flex items-start gap-2 text-sm text-error">
                <input v-model="complianceAck" type="checkbox" class="mt-0.5 rounded border-error text-error focus:ring-error" />
                <span>{{ t('customer.orders.compliance_ack') }}</span>
              </label>
            </div>
          </div>
          <button :disabled="confirming || (!order.complianceOfficialEvidence && !complianceAck)" @click="confirmOrder" class="btn btn-warning">
            <Icon name="heroicons:check-circle" class="mr-1.5 h-4 w-4" />
            {{ confirming ? t('customer.orders.confirming') : t('customer.orders.confirm_order') }}
          </button>
        </div>
      </div>
      <div v-if="order.status === 'pending'" class="mx-6 mt-4 rounded-lg border border-blue-200 bg-orange-50 px-4 py-3">
        <div class="flex items-start gap-3">
          <Icon name="heroicons:clock" class="h-5 w-5 text-blue-500 flex-shrink-0 mt-0.5" />
          <div>
            <p class="text-sm font-medium text-orange-800">{{ t('customer.orders.pending_request_title') }}</p>
            <p class="mt-1 text-sm text-orange-700">{{ t('customer.orders.pending_request_body') }}</p>
          </div>
        </div>
      </div>
      <div class="px-6 py-6">
        <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium text-light">{{ t('customer.orders.reference_number') }}</dt>
            <dd class="mt-1 text-sm text-primary font-mono">{{ order.id }}</dd>
          </div>
          <div class="sm:col-span-1">
            <dt class="text-sm font-medium text-light">{{ t('customer.orders.total_amount') }}</dt>
            <dd class="mt-1 text-sm font-bold text-highlight">{{ cur(order.currency) }} {{ order.totalAmount != null ? order.totalAmount.toLocaleString() : t('common.display.zero') }}</dd>
          </div>
          <div class="sm:col-span-1 border-t border-border pt-4">
            <dt class="text-sm font-medium text-light">{{ t('customer.orders.shipping_address') }}</dt>
            <dd class="mt-1 text-sm text-primary whitespace-pre-line">{{ formatAddress(order.shippingAddress) }}</dd>
          </div>
          <div class="sm:col-span-1 border-t border-border pt-4">
            <dt class="text-sm font-medium text-light">{{ t('customer.orders.payment_status') }}</dt>
            <dd class="mt-1 text-sm text-primary">{{ enumLabel('payment_status', order.paymentStatus, 'unpaid') }}</dd>
          </div>
          <div class="sm:col-span-2 border-t border-border pt-4">
            <dt class="text-sm font-medium text-light">{{ t('customer.orders.tracking_number') }}</dt>
            <dd class="mt-1 text-sm text-primary">{{ order.trackingNumber || t('customer.orders.not_available') }}</dd>
          </div>
          <div class="sm:col-span-1 border-t border-border pt-4">
            <dt class="text-sm font-medium text-light">{{ t('customer.orders.confirmed_at') }}</dt>
            <dd class="mt-1 text-sm text-primary">{{ formatDate(order.confirmedAt) }}</dd>
          </div>
          <div class="sm:col-span-1 border-t border-border pt-4">
            <dt class="text-sm font-medium text-light">{{ t('customer.orders.shipped_at') }}</dt>
            <dd class="mt-1 text-sm text-primary">{{ formatDate(order.shippedAt) }}</dd>
          </div>
          <div class="sm:col-span-1 border-t border-border pt-4">
            <dt class="text-sm font-medium text-light">{{ t('customer.orders.estimated_completion') }}</dt>
            <dd class="mt-1 text-sm text-primary">{{ formatDate(order.estimatedCompletion) }}</dd>
          </div>
          <div class="sm:col-span-1 border-t border-border pt-4">
            <dt class="text-sm font-medium text-light">{{ t('customer.orders.delivered_at') }}</dt>
            <dd class="mt-1 text-sm text-primary">{{ formatDate(order.deliveredAt) }}</dd>
          </div>
        </dl>
      </div>
      <div class="px-6 py-6 border-t border-border bg-bg">
        <h4 class="text-md font-semibold text-primary mb-4">{{ t('customer.orders.order_summary') }}</h4>
        <div v-if="order.items && order.items.length > 0" class="overflow-x-auto bg-white border border-border rounded-lg shadow-sm">
          <table class="min-w-full divide-y divide-border">
            <thead class="bg-bg-alt">
              <tr>
                <th class="px-4 py-3 text-left text-xs font-semibold text-primary uppercase tracking-wider">{{ t('customer.orders.col_product_id') }}</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-primary uppercase tracking-wider">{{ t('customer.orders.col_quantity') }}</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-primary uppercase tracking-wider">{{ t('customer.orders.col_unit_price') }}</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-primary uppercase tracking-wider">{{ t('customer.orders.col_subtotal') }}</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-primary uppercase tracking-wider">{{ t('customer.orders.col_specifications') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border bg-white">
              <tr v-for="(item, idx) in order.items" :key="`${item.productId}-${idx}`">
                <td class="px-4 py-3 text-sm text-primary font-mono">{{ item.productId }}</td>
                <td class="px-4 py-3 text-sm text-primary">{{ item.quantity }}</td>
                <td class="px-4 py-3 text-sm text-primary">{{ cur(order.currency) }} {{ Number(item.unitPrice || 0).toLocaleString() }}</td>
                <td class="px-4 py-3 text-sm text-primary font-semibold">{{ cur(order.currency) }} {{ Number((item.quantity || 0) * (item.unitPrice || 0)).toLocaleString() }}</td>
                <td class="px-4 py-3 text-sm text-light">{{ item.specifications || t('customer.orders.spec_na') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="text-center py-8 text-light bg-white border border-border rounded-lg shadow-sm">{{ t('customer.orders.no_items') }}</div>
      </div>

      <!-- Cancel Order -->
      <div v-if="order.status === 'pending'" class="px-6 py-4 border-t border-border bg-red-50">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-red-800">{{ t('customer.orders.cancel_order') }}</p>
            <p class="text-xs text-red-600 mt-0.5">{{ t('customer.orders.cancel_order_desc') }}</p>
          </div>
          <button @click="cancelOrder" :disabled="cancelling" class="px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-lg hover:bg-red-700 disabled:opacity-50 transition-colors">
            {{ cancelling ? t('customer.orders.cancelling') : t('customer.orders.cancel_order') }}
          </button>
        </div>
        <p v-if="cancelError" class="mt-2 text-sm text-red-700">{{ cancelError }}</p>
      </div>

      <!-- Payment Proof Upload -->
      <div v-if="order.paymentStatus !== 'paid' && order.paymentStatus !== 'refunded'" class="px-6 py-6 border-t border-border">
        <h4 class="text-md font-semibold text-primary mb-4">{{ t('customer.orders.upload_payment_proof') }}</h4>
        <form @submit.prevent="uploadPaymentProof" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-primary mb-1">{{ t('customer.orders.payment_method') }}</label>
            <select v-model="paymentForm.method" class="form-input">
              <option value="bank_transfer">{{ t('customer.orders.payment_method_bank') }}</option>
              <option value="swift">{{ t('customer.orders.payment_method_swift') }}</option>
              <option value="other">{{ t('customer.orders.payment_method_other') }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-primary mb-1">{{ t('customer.orders.payment_proof') }} *</label>
            <input ref="fileInput" type="file" accept="image/*,.pdf" @change="handleFileChange" class="form-input file-input" />
          </div>
          <div v-if="uploadMessage" class="text-sm" :class="uploadError ? 'text-error' : 'text-success'">
            {{ uploadMessage }}
          </div>
          <div>
            <button type="submit" :disabled="uploadingPayment" class="btn btn-highlight">
              {{ uploadingPayment ? t('customer.orders.uploading') : t('customer.orders.upload') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const route = useRoute()
const { t } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, enumLabel } = useDisplay()
const api = useApi()
const id = route.params.id as string
const order = ref<any>(null)
const pending = ref(true)
const error = ref('')
const confirming = ref(false)
const complianceAck = ref(false)
const cancelling = ref(false)
const cancelError = ref('')
const uploadingPayment = ref(false)
const uploadMessage = ref('')
const uploadError = ref(false)
const selectedFile = ref<File | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

const paymentForm = reactive({
  method: 'bank_transfer'
})

onMounted(async () => { await fetchOrder() })

const fetchOrder = async () => {
  pending.value = true; error.value = ''
  try { order.value = await api.getOrder(id); complianceAck.value = false }
  catch (err: any) { error.value = err?.message || t('errors.api.load_failed') }
  finally { pending.value = false }
}

const confirmOrder = async () => {
  if (!order.value?.id) return
  if (!order.value.complianceOfficialEvidence && !complianceAck.value) { error.value = t('customer.orders.error_compliance_ack'); return }
  confirming.value = true; error.value = ''
  try { await api.confirmOrder(order.value.id, order.value.complianceOfficialEvidence ? true : complianceAck.value); await fetchOrder() }
  catch (err: any) { error.value = err?.message || t('errors.api.confirm_failed') }
  finally { confirming.value = false }
}

const cancelOrder = async () => {
  if (!order.value?.id) return
  if (!confirm(t('customer.orders.cancel_confirm'))) return
  cancelling.value = true; cancelError.value = ''
  try {
    await api.post(`/user/orders/${order.value.id}/cancel`, {})
    await fetchOrder()
  } catch (err: any) {
    cancelError.value = err?.message || t('errors.api.cancel_failed')
  } finally {
    cancelling.value = false
  }
}

const formatDate = (value: string | null | undefined) => { if (!value) return t('customer.orders.date_na'); const date = new Date(value); if (Number.isNaN(date.getTime())) return t('customer.orders.date_na'); return date.toLocaleString() }
const formatAddress = (address: any) => { if (!address || typeof address !== 'object') return t('customer.orders.date_na'); const fields = [address.street, address.city, address.state, address.zipCode, address.country].filter((v) => typeof v === 'string' && v.trim() !== ''); if (fields.length === 0) return t('customer.orders.date_na'); return fields.join(', ') }
const statusBadge = (status: string) => {
  if (status === 'pending_confirmation') return 'badge-warning'
  if (status === 'pending' || status === 'processing' || status === 'production') return 'badge-warning'
  if (status === 'confirmed') return 'badge-primary'
  if (status === 'shipped') return 'badge-info'
  if (status === 'delivered') return 'badge-success'
  if (status === 'cancelled') return 'badge-error'
  return 'badge-default'
}

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    selectedFile.value = target.files[0]
  }
}

const uploadPaymentProof = async () => {
  if (!selectedFile.value) {
    uploadError.value = true
    uploadMessage.value = t('customer.orders.error_select_file')
    return
  }
  uploadingPayment.value = true; uploadMessage.value = ''; uploadError.value = false
  try {
    const formData = new FormData()
    formData.append('method', paymentForm.method)
    formData.append('proof', selectedFile.value)
    await api.customerUploadPaymentProof(id, formData)
    uploadMessage.value = t('customer.orders.payment_uploaded')
    uploadError.value = false
    selectedFile.value = null
    if (fileInput.value) fileInput.value.value = ''
    await fetchOrder()
  } catch (err: any) {
    uploadError.value = true
    uploadMessage.value = err?.data?.message || err?.message || t('errors.api.upload_failed')
  } finally { uploadingPayment.value = false }
}
</script>

<style scoped>
.form-input {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-base);
  color: var(--color-text);
  background: var(--color-bg);
  border: 1.5px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(var(--color-highlight-rgb), 0.1);
}

.file-input {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-sm);
  color: var(--color-text-light);
  background: var(--color-bg);
  border: 1.5px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.file-input::file-selector-button {
  margin-right: var(--spacing-md);
  padding: var(--spacing-xs) var(--spacing-md);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-highlight);
  background: var(--color-bg-alt);
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.file-input::file-selector-button:hover {
  background: var(--color-bg);
}
</style>
