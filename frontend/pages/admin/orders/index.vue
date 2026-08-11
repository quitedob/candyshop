<template>
  <div>
    <PageHeader :title="t('admin.orders.title')" :description="t('admin.orders.description')">
      <template #actions>
        <button type="button" class="inline-flex items-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700" @click="openCreateModal">
          {{ t('admin.orders.new_order') }}
        </button>
      </template>
    </PageHeader>

    <!-- Filters -->
    <div class="mt-6 grid grid-cols-2 sm:grid-cols-4 gap-3">
      <div>
        <label for="filter-status" class="block text-xs font-medium text-gray-500">{{ t('admin.orders.filter_status') }}</label>
        <select id="filter-status" v-model="filterStatus" name="filterStatus" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm">
          <option value="">{{ t('admin.orders.filter_all') }}</option>
          <option value="pending">{{ enumLabel('order_status', 'pending') }}</option>
          <option value="pending_approval">{{ enumLabel('order_status', 'pending_approval') }}</option>
          <option value="pending_confirmation">{{ enumLabel('order_status', 'pending_confirmation') }}</option>
          <option value="confirmed">{{ enumLabel('order_status', 'confirmed') }}</option>
          <option value="production">{{ enumLabel('order_status', 'production') }}</option>
          <option value="shipped">{{ enumLabel('order_status', 'shipped') }}</option>
          <option value="delivered">{{ enumLabel('order_status', 'delivered') }}</option>
          <option value="cancelled">{{ enumLabel('order_status', 'cancelled') }}</option>
        </select>
      </div>
      <div>
        <label for="filter-userId" class="block text-xs font-medium text-gray-500">{{ t('admin.orders.filter_user') }}</label>
        <input id="filter-userId" v-model="filterUserId" name="filterUserId" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm" :placeholder="t('admin.orders.filter_user_placeholder')" />
      </div>
      <div>
        <label for="filter-dateFrom" class="block text-xs font-medium text-gray-500">{{ t('admin.orders.filter_from') }}</label>
        <input id="filter-dateFrom" v-model="filterDateFrom" name="filterDateFrom" type="date" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm" />
      </div>
      <div>
        <label for="filter-dateTo" class="block text-xs font-medium text-gray-500">{{ t('admin.orders.filter_to') }}</label>
        <input id="filter-dateTo" v-model="filterDateTo" name="filterDateTo" type="date" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm" />
      </div>
    </div>

    <AdminTable
      class="mt-4"
      :columns="columns"
      :rows="orders"
      :loading="pending"
      :error="!!error"
      :error-message="error"
      :empty-text="t('admin.orders.no_data')"
      @retry="fetchOrders"
    >
      <template #cell-orderNumber="{ row }">
        <div class="font-medium text-gray-900">{{ row.orderNumber || row.id }}</div>
        <div class="text-gray-500">{{ formatDate(row.createdAt) }}</div>
      </template>

      <template #cell-customer="{ row }">
        <template v-if="row.user">
          <div class="font-medium text-gray-900">{{ row.user.firstName }} {{ row.user.lastName }}</div>
          <div class="text-xs text-gray-500">{{ row.user.email }}</div>
        </template>
        <template v-else>{{ row.userId }}</template>
      </template>

      <template #cell-amount="{ row }">
        {{ cur(row.currency) }} {{ formatNumber(row.totalAmount || 0) }}
      </template>

      <template #cell-status="{ row }">
        <StatusBadge :status="row.status" type="order" />
      </template>

      <template #cell-actions="{ row }">
        <NuxtLink :to="localePath(`/admin/orders/${row.id}`)" class="text-orange-600 hover:text-orange-900 mr-3">{{ t('admin.orders.view') }}</NuxtLink>
        <button type="button" class="text-blue-600 hover:text-blue-900 mr-3" @click="openFulfillmentModal(row)">{{ t('admin.orders.fulfillment') }}</button>
        <button type="button" class="text-orange-600 hover:text-orange-900 mr-3" @click="openEditModal(row.id)">{{ t('admin.orders.edit') }}</button>
        <button type="button" class="text-red-600 hover:text-red-900" @click="deleteOrder(row.id)">{{ t('admin.orders.delete') }}</button>
      </template>

      <template #bottom>
        <div v-if="pagination" class="flex items-center justify-between">
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
      </template>
    </AdminTable>

    <!-- Editor Modal -->
    <AdminModal :open="showModal" :title="editingId ? t('admin.orders.edit_order') : t('admin.orders.create_order')" width="xl" @close="closeModal">
      <form class="grid grid-cols-1 gap-4 sm:grid-cols-2" @submit.prevent="saveOrder">
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
          <div v-if="showEditPaymentBanner" class="mt-1 mb-2 p-2 bg-amber-50 border border-amber-200 rounded-md text-xs text-amber-900">
            {{ t('admin.payment_policy.requires_payment_before_execution') }}
          </div>
          <select id="order-status" v-model="form.status" name="status" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
            <option
              v-for="opt in editStatusOptions"
              :key="opt.value"
              :value="opt.value"
              :disabled="opt.disabled"
              :title="opt.paymentBlocked ? t('admin.payment_policy.requires_payment_before_execution') : undefined"
            >
              {{ enumLabel('order_status', opt.value) }}{{ opt.paymentBlocked ? ' ⛔' : '' }}
            </option>
          </select>
          <p v-if="editStatusError" class="mt-1 text-sm text-red-600">{{ editStatusError }}</p>
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

        <!-- Order Items -->
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
                <button v-if="form.items.length > 1" type="button" @click="form.items.splice(idx, 1)" class="text-red-500 hover:text-red-700 text-xs" :aria-label="t('close')">✕</button>
              </div>
            </div>
          </div>
        </div>

        <!-- Shipping Address -->
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
    </AdminModal>

    <!-- Fulfillment Modal -->
    <AdminModal :open="showFulfillmentModal" :title="t('admin.orders.fulfillment_title')" width="xl" @close="showFulfillmentModal = false">
      <div v-if="fulfillmentOrder" class="space-y-4">
        <p class="text-sm text-gray-600">{{ fulfillmentOrder.orderNumber || fulfillmentOrder.id }}</p>
        <div v-if="fulfillments.length" class="space-y-2">
          <div v-for="f in fulfillments" :key="f.id" class="border rounded-lg p-3 flex items-center justify-between">
            <div>
              <p class="text-sm font-mono">{{ f.id }}</p>
              <p class="text-xs text-gray-500">{{ t('admin.orders.fulfillment_status') }}: {{ f.status }}</p>
            </div>
            <div class="flex gap-2">
              <button v-if="f.status === 'pending'" type="button" class="text-xs text-blue-600" @click="shipFulfillment(f.id)">{{ t('admin.orders.fulfillment_ship') }}</button>
              <button v-if="f.status === 'shipped'" type="button" class="text-xs text-green-600" @click="deliverFulfillment(f.id)">{{ t('admin.orders.fulfillment_deliver') }}</button>
              <button v-if="['pending', 'shipped'].includes(f.status)" type="button" class="text-xs text-red-600" @click="cancelFulfillment(f.id)">{{ t('admin.orders.fulfillment_cancel') }}</button>
            </div>
          </div>
        </div>
        <p v-else class="text-sm text-gray-500">{{ t('admin.orders.fulfillment_no_data') }}</p>
        <form class="border-t pt-4 space-y-3" @submit.prevent="createFulfillment">
          <div v-if="enableMultiWarehouse">
            <label for="fulfillment-warehouse" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.fulfillment_warehouse') }}</label>
            <input id="fulfillment-warehouse" v-model="fulfillmentForm.warehouseId" required class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <div>
            <label for="fulfillment-notes" class="block text-sm font-medium text-gray-700">{{ t('admin.orders.fulfillment_notes') }}</label>
            <input id="fulfillment-notes" v-model="fulfillmentForm.notes" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <button type="submit" :disabled="fulfillmentSaving" class="rounded-md bg-orange-600 px-4 py-2 text-sm text-white disabled:opacity-50">{{ t('admin.orders.fulfillment_create') }}</button>
        </form>
      </div>
    </AdminModal>

    <p v-if="actionMessage" class="mt-4 text-sm" :class="actionError ? 'text-red-600' : 'text-green-600'">
      {{ actionMessage }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch, onMounted, computed } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const { t } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, enumLabel, formatNumber, formatDate } = useDisplay()
const api = useApi()
const { enableMultiWarehouse } = useFeatureFlags()
const { statusSelectOptions, paymentInsufficientForExecution, isPaidOrPartial } = useOrderStatusTransitions()

const columns = [
  { key: 'orderNumber', label: t('admin.orders.col_order') },
  { key: 'customer', label: t('admin.orders.col_customer') },
  { key: 'amount', label: t('admin.orders.col_amount') },
  { key: 'status', label: t('admin.orders.col_status') },
  { key: 'actions', label: '' },
]

const filterStatus = ref('')
const filterUserId = ref('')
const filterDateFrom = ref('')
const filterDateTo = ref('')

const orders = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20

const showModal = ref(false)
const editingId = ref('')
const editOriginalStatus = ref('pending')
const saving = ref(false)
const formError = ref('')
const editStatusError = ref('')
const actionMessage = ref('')
const actionError = ref(false)

const editStatusOptions = computed(() =>
  statusSelectOptions(editOriginalStatus.value, form.paymentStatus || 'unpaid')
)

const showEditPaymentBanner = computed(() =>
  showModal.value && !isPaidOrPartial(form.paymentStatus || 'unpaid')
)

const showFulfillmentModal = ref(false)
const fulfillmentOrder = ref<any>(null)
const fulfillments = ref<any[]>([])
const fulfillmentSaving = ref(false)
const fulfillmentForm = reactive({ warehouseId: '', notes: '' })

const openFulfillmentModal = async (order: any) => {
  fulfillmentOrder.value = order
  fulfillmentForm.warehouseId = ''
  fulfillmentForm.notes = ''
  showFulfillmentModal.value = true
  try {
    fulfillments.value = await api.get<any[]>(`/admin/orders/${order.id}/fulfillments`) || []
  } catch { fulfillments.value = [] }
}

const createFulfillment = async () => {
  if (!fulfillmentOrder.value) return
  fulfillmentSaving.value = true
  try {
    const items = (fulfillmentOrder.value.items || []).map((item: any, idx: number) => ({
      orderItemIdx: idx,
      productId: item.productId,
      quantity: item.quantity || 1,
    }))
    await api.post(`/admin/orders/${fulfillmentOrder.value.id}/fulfillments`, {
      ...(enableMultiWarehouse ? { warehouseId: fulfillmentForm.warehouseId } : {}),
      notes: fulfillmentForm.notes,
      items,
    })
    actionMessage.value = t('admin.orders.fulfillment_created')
    fulfillments.value = await api.get<any[]>(`/admin/orders/${fulfillmentOrder.value.id}/fulfillments`) || []
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('errors.api.save_failed')
  } finally {
    fulfillmentSaving.value = false
  }
}

const shipFulfillment = async (id: string) => {
  const tracking = prompt(t('admin.orders.fulfillment_tracking'))
  if (!tracking) return
  try {
    await api.put(`/admin/fulfillments/${id}/ship`, { trackingNumber: tracking, carrier: 'standard' })
    actionMessage.value = t('admin.orders.fulfillment_shipped')
    if (fulfillmentOrder.value) fulfillments.value = await api.get<any[]>(`/admin/orders/${fulfillmentOrder.value.id}/fulfillments`) || []
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('errors.api.action_failed')
  }
}

const deliverFulfillment = async (id: string) => {
  try {
    await api.put(`/admin/fulfillments/${id}/deliver`, {})
    actionMessage.value = t('admin.orders.fulfillment_delivered')
    if (fulfillmentOrder.value) fulfillments.value = await api.get<any[]>(`/admin/orders/${fulfillmentOrder.value.id}/fulfillments`) || []
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('errors.api.action_failed')
  }
}

const cancelFulfillment = async (id: string) => {
  try {
    await api.put(`/admin/fulfillments/${id}/cancel`, {})
    if (fulfillmentOrder.value) fulfillments.value = await api.get<any[]>(`/admin/orders/${fulfillmentOrder.value.id}/fulfillments`) || []
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('errors.api.action_failed')
  }
}

type OrderItem = { productId: string; quantity: number; unitPrice: number; specifications: string; fulfilledQuantity?: number; shippedQuantity?: number }
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

const fetchOrders = async () => {
  pending.value = true
  error.value = ''
  try {
    const params: Record<string, any> = { page: page.value, limit: pageSize }
    if (filterStatus.value) params.status = filterStatus.value
    if (filterUserId.value) params.userId = filterUserId.value
    if (filterDateFrom.value) params.dateFrom = filterDateFrom.value
    if (filterDateTo.value) params.dateTo = filterDateTo.value
    const res = await api.get<any>('/admin/orders', params)
    orders.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const nextPage = () => {
  if (pagination.value && page.value < pagination.value.totalPages) page.value += 1
}

const prevPage = () => {
  if (page.value > 1) page.value -= 1
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
    ? order.items.map((i: any) => ({ productId: i.productId || '', quantity: i.quantity || 1, unitPrice: i.unitPrice || 0, specifications: i.specifications || '', fulfilledQuantity: i.fulfilledQuantity || 0, shippedQuantity: i.shippedQuantity || 0 }))
    : [{ productId: '', quantity: 1, unitPrice: 0, specifications: '' }]
  const addr = order.shippingAddress || {}
  form.shippingAddress = { street: addr.street || '', city: addr.city || '', state: addr.state || '', zipCode: addr.zipCode || '', country: addr.country || '' }
}

const openCreateModal = () => {
  editingId.value = ''
  editOriginalStatus.value = 'pending'
  resetForm()
  formError.value = ''
  editStatusError.value = ''
  actionMessage.value = ''
  actionError.value = false
  showModal.value = true
}

const openEditModal = async (orderID: string) => {
  editingId.value = orderID
  formError.value = ''
  editStatusError.value = ''
  actionMessage.value = ''
  actionError.value = false
  try {
    const order = await api.get<any>(`/admin/orders/${orderID}`)
    editOriginalStatus.value = order.status || 'pending'
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
  editStatusError.value = ''
}

const buildPayload = () => {
  const items = form.items.filter(i => i.productId.trim()).map(i => ({
    productId: i.productId.trim(),
    quantity: Math.max(1, i.quantity),
    unitPrice: Math.max(0, i.unitPrice),
    specifications: i.specifications.trim(),
    // Preserve fulfillment progress so the backend doesn't zero-out
    // fulfilledQuantity/shippedQuantity on edit (which previously made
    // re-ship/fulfillment bookkeeping inconsistent and failed with 422).
    fulfilledQuantity: i.fulfilledQuantity || 0,
    shippedQuantity: i.shippedQuantity || 0,
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

  if (paymentInsufficientForExecution(form.status, form.paymentStatus || 'unpaid')) {
    const msg = t('admin.payment_policy.requires_payment_before_execution')
    formError.value = msg
    editStatusError.value = msg
    return
  }

  saving.value = true
  formError.value = ''
  editStatusError.value = ''
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
    if (paymentInsufficientForExecution(form.status, form.paymentStatus || 'unpaid')) {
      editStatusError.value = formError.value
    }
  } finally {
    saving.value = false
  }
}

watch(() => form.status, () => {
  if (editStatusError.value && !paymentInsufficientForExecution(form.status, form.paymentStatus || 'unpaid')) {
    editStatusError.value = ''
  }
})

watch(() => form.paymentStatus, () => {
  if (editStatusError.value && !paymentInsufficientForExecution(form.status, form.paymentStatus || 'unpaid')) {
    editStatusError.value = ''
  }
})

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
watch([filterStatus, filterUserId, filterDateFrom, filterDateTo], () => {
  page.value = 1
  fetchOrders()
})
onMounted(fetchOrders)
</script>
