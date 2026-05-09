<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.orders.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.orders.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0">
        <button type="button" class="inline-flex items-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700" @click="openCreateModal">
          {{ t('admin.orders.new_order') }}
        </button>
      </div>
    </div>

    <div class="mt-8 overflow-hidden rounded-lg bg-white shadow ring-1 ring-black ring-opacity-5">
      <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
          <tr>
            <th class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{{ t('admin.orders.col_order') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.orders.col_customer') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.orders.col_amount') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.orders.col_status') }}</th>
            <th class="relative py-3.5 pl-3 pr-4 sm:pr-6"><span class="sr-only">{{ t('admin.orders.col_actions') }}</span></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <tr v-if="pending">
            <td colspan="5" class="py-5 text-center text-sm text-gray-500">{{ t('admin.orders.loading') }}</td>
          </tr>
          <tr v-else-if="error">
            <td colspan="5" class="py-5 text-center text-sm text-red-600">{{ error }}</td>
          </tr>
          <tr v-else-if="orders.length === 0">
            <td colspan="5" class="py-5 text-center text-sm text-gray-500">{{ t('admin.orders.no_data') }}</td>
          </tr>
          <tr v-else v-for="order in orders" :key="order.id">
            <td class="py-4 pl-4 pr-3 text-sm sm:pl-6">
              <div class="font-medium text-gray-900">{{ order.orderNumber || order.id }}</div>
              <div class="text-gray-500">{{ formatDate(order.createdAt) }}</div>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">
              <template v-if="order.user">
                <div class="font-medium text-gray-900">{{ order.user.firstName }} {{ order.user.lastName }}</div>
                <div class="text-xs text-gray-500">{{ order.user.email }}</div>
              </template>
              <template v-else>{{ order.userId }}</template>
            </td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
              {{ cur(order.currency) }} {{ formatNumber(order.totalAmount || 0) }}
            </td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
              <span class="inline-flex rounded-full px-2 text-xs font-semibold leading-5" :class="statusClass(order.status)">
                {{ enumLabel('order_status', order.status) }}
              </span>
            </td>
            <td class="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
              <NuxtLink :to="localePath(`/admin/orders/${order.id}`)" class="text-orange-600 hover:text-orange-900 mr-3">{{ t('admin.orders.view') }}</NuxtLink>
              <button type="button" class="text-orange-600 hover:text-orange-900" @click="openEditModal(order.id)">{{ t('admin.orders.edit') }}</button>
              <button type="button" class="ml-4 text-red-600 hover:text-red-900" @click="deleteOrder(order.id)">{{ t('admin.orders.delete') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="pagination" class="mt-4 flex items-center justify-between rounded-lg border-t border-gray-200 bg-white px-4 py-3 shadow sm:px-6">
      <div class="text-sm text-gray-700">
        {{ t('admin.orders.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
      </div>
      <div class="flex items-center gap-2">
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page <= 1" @click="prevPage">
          {{ t('admin.orders.previous') }}
        </button>
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page >= pagination.totalPages" @click="nextPage">
          {{ t('admin.orders.next') }}
        </button>
      </div>
    </div>

    <div v-if="showModal" class="fixed inset-0 z-10 overflow-y-auto" role="dialog" aria-modal="true">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <button type="button" class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity w-full border-0 cursor-pointer" @click="closeModal" :aria-label="t('common.close')"></button>
        <span class="hidden sm:inline-block sm:h-screen sm:align-middle" aria-hidden="true">&#8203;</span>
        <div class="inline-block w-full transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left align-bottom shadow-xl sm:my-8 sm:max-w-4xl sm:p-6 sm:align-middle">
          <h3 class="text-lg font-medium leading-6 text-gray-900">{{ editingId ? t('admin.orders.edit_order') : t('admin.orders.create_order') }}</h3>

          <form class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2" @submit.prevent="saveOrder">
            <div>
              <label for="order-userId" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.user_id') }}</label>
              <input id="order-userId" v-model="form.userId" name="userId" autocomplete="off" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="order-inquiryId" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.inquiry_id') }}</label>
              <input id="order-inquiryId" v-model="form.inquiryId" name="inquiryId" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="order-orderNumber" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.order_number') }}</label>
              <input id="order-orderNumber" v-model="form.orderNumber" name="orderNumber" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="order-currency" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.currency') }}</label>
              <input id="order-currency" v-model="form.currency" name="currency" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="order-status" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.status') }}</label>
              <select id="order-status" v-model="form.status" name="status" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option value="pending">{{ enumLabel('order_status', 'pending') }}</option>
                <option value="confirmed">{{ enumLabel('order_status', 'confirmed') }}</option>
                <option value="production">{{ enumLabel('order_status', 'production') }}</option>
                <option value="shipped">{{ enumLabel('order_status', 'shipped') }}</option>
                <option value="delivered">{{ enumLabel('order_status', 'delivered') }}</option>
                <option value="cancelled">{{ enumLabel('order_status', 'cancelled') }}</option>
              </select>
            </div>
            <div>
              <label for="order-paymentStatus" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.payment_status') }}</label>
              <select id="order-paymentStatus" v-model="form.paymentStatus" name="paymentStatus" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option value="unpaid">unpaid</option>
                <option value="partial">partial</option>
                <option value="paid">paid</option>
                <option value="refunded">refunded</option>
              </select>
            </div>
            <div>
              <label for="order-subtotal" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.subtotal') }}</label>
              <input id="order-subtotal" v-model.number="form.subtotal" name="subtotal" type="number" min="0" step="0.01" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="order-taxAmount" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.tax') }}</label>
              <input id="order-taxAmount" v-model.number="form.taxAmount" name="taxAmount" type="number" min="0" step="0.01" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="order-shippingAmount" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.shipping') }}</label>
              <input id="order-shippingAmount" v-model.number="form.shippingAmount" name="shippingAmount" type="number" min="0" step="0.01" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="order-totalAmount" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.total') }}</label>
              <input id="order-totalAmount" v-model.number="form.totalAmount" name="totalAmount" type="number" min="0" step="0.01" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div class="sm:col-span-2">
              <label for="order-trackingNumber" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.tracking_number') }}</label>
              <input id="order-trackingNumber" v-model="form.trackingNumber" name="trackingNumber" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>

            <!-- Structured Order Items -->
            <div class="sm:col-span-2">
              <div class="flex items-center justify-between mb-2">
                <span class="block text-sm font-medium text-gray-700">{{ t('admin.orders.items') }}</span>
                <button type="button" @click="addOrderItem" class="text-xs text-orange-600 hover:text-orange-800 font-medium">+ {{ t('admin.orders.add_item') }}</button>
              </div>
              <div class="space-y-2">
                <div v-for="(item, idx) in form.items" :key="idx" class="grid grid-cols-12 gap-2 items-end rounded border border-gray-200 p-2">
                  <div class="col-span-4">
                    <label :for="`order-item-productId-${idx}`" class="block text-xs text-gray-500">{{ t('admin.orders.item_product_id') }}</label>
                    <input :id="`order-item-productId-${idx}`" v-model="item.productId" :name="`items[${idx}].productId`" autocomplete="off" class="mt-0.5 w-full rounded border border-gray-300 px-2 py-1 text-xs" :placeholder="t('admin.orders.item_product_id')" />
                  </div>
                  <div class="col-span-2">
                    <label :for="`order-item-quantity-${idx}`" class="block text-xs text-gray-500">{{ t('admin.orders.item_qty') }}</label>
                    <input :id="`order-item-quantity-${idx}`" v-model.number="item.quantity" :name="`items[${idx}].quantity`" type="number" min="1" autocomplete="off" class="mt-0.5 w-full rounded border border-gray-300 px-2 py-1 text-xs" />
                  </div>
                  <div class="col-span-2">
                    <label :for="`order-item-unitPrice-${idx}`" class="block text-xs text-gray-500">{{ t('admin.orders.item_price') }}</label>
                    <input :id="`order-item-unitPrice-${idx}`" v-model.number="item.unitPrice" :name="`items[${idx}].unitPrice`" type="number" min="0" step="0.01" autocomplete="off" class="mt-0.5 w-full rounded border border-gray-300 px-2 py-1 text-xs" />
                  </div>
                  <div class="col-span-3">
                    <label :for="`order-item-spec-${idx}`" class="block text-xs text-gray-500">{{ t('admin.orders.item_spec') }}</label>
                    <input :id="`order-item-spec-${idx}`" v-model="item.specifications" :name="`items[${idx}].specifications`" autocomplete="off" class="mt-0.5 w-full rounded border border-gray-300 px-2 py-1 text-xs" />
                  </div>
                  <div class="col-span-1 flex justify-end">
                    <button v-if="form.items.length > 1" type="button" @click="form.items.splice(idx, 1)" class="text-red-500 hover:text-red-700 text-xs" :aria-label="t('common.close')">✕</button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Structured Shipping Address -->
            <div class="sm:col-span-2">
              <span class="block text-sm font-medium text-gray-700 mb-2">{{ t('admin.orders.shipping_address') }}</span>
              <div class="grid grid-cols-2 gap-2">
                <div class="col-span-2">
                  <label for="order-addr-street" class="block text-xs text-gray-500">{{ t('admin.orders.addr_street') }}</label>
                  <input id="order-addr-street" v-model="form.shippingAddress.street" name="shippingAddress.street" autocomplete="street-address" class="mt-0.5 w-full rounded border border-gray-300 px-2 py-1 text-sm" />
                </div>
                <div>
                  <label for="order-addr-city" class="block text-xs text-gray-500">{{ t('admin.orders.addr_city') }}</label>
                  <input id="order-addr-city" v-model="form.shippingAddress.city" name="shippingAddress.city" autocomplete="address-level2" class="mt-0.5 w-full rounded border border-gray-300 px-2 py-1 text-sm" />
                </div>
                <div>
                  <label for="order-addr-state" class="block text-xs text-gray-500">{{ t('admin.orders.addr_state') }}</label>
                  <input id="order-addr-state" v-model="form.shippingAddress.state" name="shippingAddress.state" autocomplete="address-level1" class="mt-0.5 w-full rounded border border-gray-300 px-2 py-1 text-sm" />
                </div>
                <div>
                  <label for="order-addr-zip" class="block text-xs text-gray-500">{{ t('admin.orders.addr_zip') }}</label>
                  <input id="order-addr-zip" v-model="form.shippingAddress.zipCode" name="shippingAddress.zipCode" autocomplete="postal-code" class="mt-0.5 w-full rounded border border-gray-300 px-2 py-1 text-sm" />
                </div>
                <div>
                  <label for="order-addr-country" class="block text-xs text-gray-500">{{ t('admin.orders.addr_country') }}</label>
                  <input id="order-addr-country" v-model="form.shippingAddress.country" name="shippingAddress.country" autocomplete="country-name" class="mt-0.5 w-full rounded border border-gray-300 px-2 py-1 text-sm" />
                </div>
              </div>
            </div>

            <div v-if="formError" class="sm:col-span-2 text-sm text-red-600">{{ formError }}</div>
            <div class="sm:col-span-2 mt-2 flex justify-end gap-3">
              <button type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700" @click="closeModal">
                {{ t('admin.orders.cancel') }}
              </button>
              <button type="submit" :disabled="saving" class="rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm text-white disabled:opacity-50">
                {{ saving ? t('admin.orders.saving') : (editingId ? t('admin.orders.update') : t('admin.orders.create')) }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <p v-if="actionMessage" class="mt-4 text-sm" :class="actionError ? 'text-red-600' : 'text-green-600'">
      {{ actionMessage }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch, onMounted } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const { t } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, cell, enumLabel, formatNumber, formatDate } = useDisplay()
const api = useApi()

const orders = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20

const showModal = ref(false)
const editingId = ref('')
const saving = ref(false)
const formError = ref('')
const actionMessage = ref('')
const actionError = ref(false)

type OrderItem = { productId: string; quantity: number; unitPrice: number; specifications: string }
type ShippingAddress = { street: string; city: string; state: string; zipCode: string; country: string }

const form = reactive({
  userId: '',
  inquiryId: '',
  orderNumber: '',
  status: 'pending',
  paymentStatus: 'unpaid',
  subtotal: 0,
  taxAmount: 0,
  shippingAmount: 0,
  totalAmount: 0,
  currency: cur(null),
  trackingNumber: '',
  items: [{ productId: '', quantity: 1, unitPrice: 0, specifications: '' }] as OrderItem[],
  shippingAddress: { street: '', city: '', state: '', zipCode: '', country: '' } as ShippingAddress
})

const addOrderItem = () => {
  form.items.push({ productId: '', quantity: 1, unitPrice: 0, specifications: '' })
}

const statusClass = (status: string) => {
  if (status === 'pending') return 'badge-warning'
  if (status === 'confirmed' || status === 'production') return 'badge-info'
  if (status === 'shipped') return 'badge-primary'
  if (status === 'delivered') return 'badge-success'
  return 'badge-default'
}

const fetchOrders = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await api.get<any>('/admin/orders', { page: page.value, limit: pageSize })
    orders.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const nextPage = () => {
  if (pagination.value && page.value < pagination.value.totalPages) {
    page.value += 1
  }
}

const prevPage = () => {
  if (page.value > 1) {
    page.value -= 1
  }
}

const resetForm = () => {
  form.userId = ''
  form.inquiryId = ''
  form.orderNumber = ''
  form.status = 'pending'
  form.paymentStatus = 'unpaid'
  form.subtotal = 0
  form.taxAmount = 0
  form.shippingAmount = 0
  form.totalAmount = 0
  form.currency = cur(null)
  form.trackingNumber = ''
  form.items = [{ productId: '', quantity: 1, unitPrice: 0, specifications: '' }]
  form.shippingAddress = { street: '', city: '', state: '', zipCode: '', country: '' }
}

const fillForm = (order: any) => {
  form.userId = order.userId || ''
  form.inquiryId = order.inquiryId || ''
  form.orderNumber = order.orderNumber || ''
  form.status = order.status || 'pending'
  form.paymentStatus = order.paymentStatus || 'unpaid'
  form.subtotal = order.subtotal || 0
  form.taxAmount = order.taxAmount || 0
  form.shippingAmount = order.shippingAmount || 0
  form.totalAmount = order.totalAmount || 0
  form.currency = cur(order.currency)
  form.trackingNumber = order.trackingNumber || ''
  form.items = Array.isArray(order.items) && order.items.length
    ? order.items.map((i: any) => ({ productId: i.productId || '', quantity: i.quantity || 1, unitPrice: i.unitPrice || 0, specifications: i.specifications || '' }))
    : [{ productId: '', quantity: 1, unitPrice: 0, specifications: '' }]
  const addr = order.shippingAddress || {}
  form.shippingAddress = { street: addr.street || '', city: addr.city || '', state: addr.state || '', zipCode: addr.zipCode || '', country: addr.country || '' }
}

const openCreateModal = () => {
  editingId.value = ''
  resetForm()
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false
  showModal.value = true
}

const openEditModal = async (orderID: string) => {
  editingId.value = orderID
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false
  try {
    const order = await api.get<any>(`/admin/orders/${orderID}`)
    fillForm(order)
    showModal.value = true
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('errors.api.load_failed')
  }
}

const closeModal = () => {
  showModal.value = false
  saving.value = false
  formError.value = ''
}

const buildPayload = () => {
  const items = form.items.filter(i => i.productId.trim()).map(i => ({
    productId: i.productId.trim(),
    quantity: Math.max(1, i.quantity),
    unitPrice: Math.max(0, i.unitPrice),
    specifications: i.specifications.trim()
  }))

  const payload: Record<string, any> = {
    userId: form.userId,
    orderNumber: form.orderNumber,
    status: form.status,
    paymentStatus: form.paymentStatus,
    subtotal: form.subtotal,
    taxAmount: form.taxAmount,
    shippingAmount: form.shippingAmount,
    totalAmount: form.totalAmount,
    currency: form.currency,
    trackingNumber: form.trackingNumber,
    items,
    shippingAddress: form.shippingAddress
  }

  if (form.inquiryId.trim()) {
    payload.inquiryId = form.inquiryId.trim()
  } else if (editingId.value) {
    payload.inquiryId = ''
  }

  return payload
}

const saveOrder = async () => {
  if (!form.userId.trim()) {
    formError.value = t('admin.orders.user_id_required')
    return
  }

  saving.value = true
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false

  try {
    const payload = buildPayload()
    if (editingId.value) {
      await api.put(`/admin/orders/${editingId.value}`, payload)
      actionMessage.value = t('admin.orders.updated_success')
    } else {
      await api.post('/admin/orders', payload)
      actionMessage.value = t('admin.orders.created_success')
    }
    closeModal()
    await fetchOrders()
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.save_failed')
  } finally {
    saving.value = false
  }
}

const deleteOrder = async (id: string) => {
  if (!confirm(t('admin.orders.confirm_delete'))) return

  actionMessage.value = ''
  actionError.value = false
  try {
    await api.del(`/admin/orders/${id}`)
    actionMessage.value = t('admin.orders.deleted_success')
    await fetchOrders()
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('errors.api.delete_failed')
  }
}

watch(page, fetchOrders)
onMounted(fetchOrders)
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-xl);
}

.page-title {
  font-size: var(--text-2xl);
  font-weight: 600;
  color: var(--color-primary);
}

.page-subtitle {
  margin-top: var(--spacing-xs);
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.table-container {
  overflow: hidden;
  border-radius: var(--radius-lg);
  background: white;
  box-shadow: var(--shadow-md);
  margin-top: var(--spacing-xl);
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  padding: var(--spacing-md);
  text-align: left;
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-primary);
  background: var(--color-bg-alt);
  border-bottom: 1px solid var(--color-border);
}

.data-table td {
  padding: var(--spacing-md);
  font-size: var(--text-sm);
  color: var(--color-text);
  border-bottom: 1px solid var(--color-border);
}

.data-table tbody tr:hover {
  background: var(--color-bg);
}

.data-table tbody tr:last-child td {
  border-bottom: none;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: var(--spacing-lg);
  padding: var(--spacing-md) var(--spacing-lg);
  background: white;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
}

.link {
  color: var(--color-highlight);
  font-weight: 500;
  transition: color var(--transition-fast);
}

.link:hover {
  color: var(--color-highlight-hover);
}

.link-danger {
  color: var(--color-error);
}

.link-danger:hover {
  color: #dc2626;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-lg);
  margin-top: var(--spacing-lg);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.form-group.col-span-2 {
  grid-column: span 2;
}

.form-label {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
}

.form-input {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-base);
  color: var(--color-text);
  background: var(--color-bg);
  border: 1.5px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(var(--color-highlight-rgb), 0.1);
}

.form-textarea {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-base);
  color: var(--color-text);
  background: var(--color-bg);
  border: 1.5px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  resize: vertical;
}

.form-textarea:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(var(--color-highlight-rgb), 0.1);
}

.form-section {
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border);
}

.form-section-title {
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: var(--spacing-sm);
}

.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl);
}

.modal-container {
  position: relative;
  width: 100%;
  max-width: 800px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
}

.modal-content {
  position: relative;
  background: white;
  border-radius: var(--radius-xl);
  padding: var(--spacing-xl);
  box-shadow: var(--shadow-xl);
  animation: modalIn 0.3s ease forwards;
}

@keyframes modalIn {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.modal-title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-primary);
}
</style>
