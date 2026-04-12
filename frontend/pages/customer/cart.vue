<template>
  <div class="min-h-screen bg-gradient-to-br from-amber-50 via-orange-50 to-yellow-50">
    <!-- Cart Header -->
    <div class="bg-white border-b border-orange-100 shadow-sm">
      <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div class="flex items-center justify-between">
          <div>
            <h1 class="text-3xl font-bold bg-gradient-to-r from-orange-600 to-amber-600 bg-clip-text text-transparent">
              {{ t('customer.cart.title') }}
            </h1>
            <p class="mt-1 text-sm text-gray-600">{{ t('customer.cart.subtitle') }}</p>
          </div>
          <NuxtLink to="/customer/products" class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-orange-100 text-orange-700 hover:bg-orange-200 transition-colors">
            <Icon name="heroicons:arrow-left" class="h-4 w-4" />
            {{ t('customer.cart.continue_shopping') }}
          </NuxtLink>
        </div>
      </div>
    </div>

    <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div v-if="pending" class="flex items-center justify-center py-20">
        <div class="text-center">
          <div class="animate-spin rounded-full h-12 w-12 border-4 border-orange-200 border-t-orange-600 mx-auto"></div>
          <p class="mt-4 text-gray-600">{{ t('customer.cart.loading') }}</p>
        </div>
      </div>

      <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-xl p-6 text-center">
        <Icon name="heroicons:exclamation-triangle" class="h-12 w-12 text-red-400 mx-auto mb-3" />
        <p class="text-red-700 font-medium">{{ error }}</p>
        <button @click="fetchCart" class="mt-4 px-4 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors">
          {{ t('customer.cart.retry') }}
        </button>
      </div>

      <div v-else-if="cartItems.length === 0" class="text-center py-20">
        <div class="w-32 h-32 mx-auto mb-6 rounded-full bg-gradient-to-br from-orange-100 to-amber-100 flex items-center justify-center">
          <Icon name="heroicons:shopping-bag" class="h-16 w-16 text-orange-400" />
        </div>
        <h2 class="text-2xl font-semibold text-gray-900 mb-2">{{ t('customer.cart.empty_title') }}</h2>
        <p class="text-gray-600 mb-6">{{ t('customer.cart.empty_desc') }}</p>
        <NuxtLink to="/customer/products" class="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-orange-500 to-amber-500 text-white font-medium rounded-xl hover:from-orange-600 hover:to-amber-600 transition-all shadow-lg shadow-orange-200">
          <Icon name="heroicons:sparkles" class="h-5 w-5" />
          {{ t('customer.cart.browse_products') }}
        </NuxtLink>
      </div>

      <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <!-- Cart Items -->
        <div class="lg:col-span-2 space-y-4">
          <div v-for="item in cartItems" :key="item.id" class="bg-white rounded-2xl shadow-sm border border-orange-100 overflow-hidden hover:shadow-md transition-shadow">
            <div class="p-5">
              <div class="flex items-start gap-4">
                <!-- Product Image -->
                <div class="flex-shrink-0 w-24 h-24 rounded-xl bg-gradient-to-br from-gray-50 to-gray-100 flex items-center justify-center overflow-hidden">
                  <img v-if="item.thumbnail" :src="item.thumbnail" :alt="item.name" class="w-full h-full object-cover" />
                  <Icon v-else name="heroicons:cube" class="h-10 w-10 text-gray-300" />
                </div>

                <!-- Product Details -->
                <div class="flex-1 min-w-0">
                  <div class="flex items-start justify-between gap-4">
                    <div>
                      <h3 class="text-lg font-semibold text-gray-900">{{ item.name }}</h3>
                      <p class="text-sm text-gray-500 mt-0.5">{{ item.category || 'Candy' }}</p>
                      <div class="flex items-center gap-2 mt-2">
                        <span v-if="item.halalCertified" class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 text-xs font-medium">
                          <Icon name="heroicons:check-badge" class="h-3 w-3" />
                          Halal
                        </span>
                        <span v-if="item.oemAvailable" class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-blue-50 text-blue-700 text-xs font-medium">
                          <Icon name="heroicons:sparkles" class="h-3 w-3" />
                          OEM
                        </span>
                      </div>
                    </div>
                    <button @click="removeItem(item.id)" class="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors">
                      <Icon name="heroicons:trash" class="h-5 w-5" />
                    </button>
                  </div>

                  <!-- Quantity & Price -->
                  <div class="flex items-end justify-between mt-4">
                    <div class="flex items-center gap-3">
                      <button @click="updateQuantity(item.id, item.quantity - 1)" :disabled="item.quantity <= 1" class="w-9 h-9 flex items-center justify-center rounded-lg border border-gray-200 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
                        <Icon name="heroicons:minus" class="h-4 w-4" />
                      </button>
                      <input v-model.number="item.quantity" type="number" min="1" :max="item.maxQuantity || 9999" @change="updateQuantity(item.id, item.quantity)" class="w-16 h-9 text-center border border-gray-200 rounded-lg font-medium focus:outline-none focus:ring-2 focus:ring-orange-500" />
                      <button @click="updateQuantity(item.id, item.quantity + 1)" class="w-9 h-9 flex items-center justify-center rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors">
                        <Icon name="heroicons:plus" class="h-4 w-4" />
                      </button>
                      <span class="text-sm text-gray-500">MOQ: {{ item.moq || 1 }}</span>
                    </div>
                    <div class="text-right">
                      <p class="text-lg font-bold text-orange-600">{{ item.currency || 'USD' }} {{ ((item.unitPrice || 0) * item.quantity).toLocaleString() }}</p>
                      <p class="text-xs text-gray-500">{{ item.currency || 'USD' }} {{ (item.unitPrice || 0).toLocaleString() }} / unit</p>
                    </div>
                  </div>

                  <!-- Specifications -->
                  <div v-if="item.specifications" class="mt-3 pt-3 border-t border-gray-100">
                    <p class="text-xs text-gray-500">{{ item.specifications }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Clear Cart -->
          <div class="flex justify-end">
            <button @click="clearCart" class="inline-flex items-center gap-2 px-4 py-2 text-sm text-red-600 hover:text-red-700 hover:bg-red-50 rounded-lg transition-colors">
              <Icon name="heroicons:trash" class="h-4 w-4" />
              {{ t('customer.cart.clear_cart') }}
            </button>
          </div>
        </div>

        <!-- Order Summary -->
        <div class="lg:col-span-1">
          <div class="bg-white rounded-2xl shadow-sm border border-orange-100 sticky top-8 overflow-hidden">
            <div class="bg-gradient-to-r from-orange-500 to-amber-500 px-6 py-4">
              <h2 class="text-lg font-semibold text-white">{{ t('customer.cart.order_summary') }}</h2>
            </div>
            <div class="p-6 space-y-4">
              <div class="flex justify-between text-sm">
                <span class="text-gray-600">{{ t('customer.cart.items_count', { count: cartItems.length }) }}</span>
                <span class="font-medium text-gray-900">{{ cartItems.reduce((sum, i) => sum + i.quantity, 0) }} units</span>
              </div>
              <div class="flex justify-between text-sm">
                <span class="text-gray-600">{{ t('customer.cart.subtotal') }}</span>
                <span class="font-medium text-gray-900">{{ summary.currency || 'USD' }} {{ (summary.subtotal || 0).toLocaleString() }}</span>
              </div>
              <div class="flex justify-between text-sm">
                <span class="text-gray-600">{{ t('customer.cart.estimated_tax') }}</span>
                <span class="font-medium text-gray-500 italic">{{ t('customer.cart.tbd') }}</span>
              </div>
              <div class="flex justify-between text-sm">
                <span class="text-gray-600">{{ t('customer.cart.estimated_shipping') }}</span>
                <span class="font-medium text-gray-500 italic">{{ t('customer.cart.tbd') }}</span>
              </div>
              <div class="pt-4 border-t border-gray-100">
                <div class="flex justify-between">
                  <span class="text-base font-semibold text-gray-900">{{ t('customer.cart.subtotal_label') }}</span>
                  <span class="text-xl font-bold text-orange-600">{{ summary.currency || 'USD' }} {{ (summary.subtotal || 0).toLocaleString() }}</span>
                </div>
                <p class="text-xs text-gray-400 mt-1">{{ t('customer.cart.tax_shipping_note') }}</p>
              </div>

              <!-- Shipping Address -->
              <div class="pt-4 border-t border-gray-100">
                <div class="flex items-center justify-between mb-2">
                  <h3 class="text-sm font-semibold text-gray-900">{{ t('customer.cart.shipping_to') }}</h3>
                  <button @click="showAddressForm = !showAddressForm" class="text-xs text-orange-600 hover:text-orange-700">
                    {{ showAddressForm ? t('customer.cart.hide') : t('customer.cart.change') }}
                  </button>
                </div>
                <div v-if="!showAddressForm && shippingAddress" class="text-sm text-gray-600">
                  <p>{{ shippingAddress.street }}</p>
                  <p>{{ shippingAddress.city }}, {{ shippingAddress.state }} {{ shippingAddress.zipCode }}</p>
                  <p>{{ shippingAddress.country }}</p>
                </div>
                <div v-else class="space-y-3">
                  <input v-model="tempAddress.country" type="text" :placeholder="t('customer.cart.country')" class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
                  <input v-model="tempAddress.city" type="text" :placeholder="t('customer.cart.city')" class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
                  <input v-model="tempAddress.state" type="text" :placeholder="t('customer.cart.state')" class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
                  <input v-model="tempAddress.street" type="text" :placeholder="t('customer.cart.street')" class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
                  <input v-model="tempAddress.zipCode" type="text" :placeholder="t('customer.cart.zip')" class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
                </div>
              </div>

              <!-- Action Buttons -->
              <div class="pt-4 space-y-3">
                <button @click="proceedToCheckout" :disabled="checkingOut" class="w-full py-3 bg-gradient-to-r from-orange-500 to-amber-500 text-white font-semibold rounded-xl hover:from-orange-600 hover:to-amber-600 transition-all shadow-lg shadow-orange-200 disabled:opacity-50 flex items-center justify-center gap-2">
                  <Icon v-if="checkingOut" name="heroicons:arrow-path" class="h-5 w-5 animate-spin" />
                  <Icon v-else name="heroicons:credit-card" class="h-5 w-5" />
                  {{ checkingOut ? t('customer.cart.processing') : t('customer.cart.proceed_checkout') }}
                </button>
                <NuxtLink to="/customer/orders/new" class="w-full py-3 bg-gray-100 text-gray-700 font-medium rounded-xl hover:bg-gray-200 transition-colors flex items-center justify-center gap-2">
                  <Icon name="heroicons:sparkles" class="h-5 w-5" />
                  {{ t('customer.cart.ai_assist') }}
                </NuxtLink>
              </div>

              <!-- Compliance Note -->
              <p class="text-xs text-gray-500 text-center pt-2">
                {{ t('customer.cart.compliance_note') }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Success Toast -->
    <Transition name="slide-up">
      <div v-if="showToast" class="fixed bottom-6 right-6 bg-emerald-600 text-white px-6 py-3 rounded-xl shadow-lg flex items-center gap-3 z-50">
        <Icon name="heroicons:check-circle" class="h-5 w-5" />
        {{ toastMessage }}
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'

definePageMeta({
  layout: 'customer',
  middleware: ['auth']
})

const { t, locale } = useI18n()
const api = useApi()
const tts = useTTS()

const cartItems = ref<any[]>([])
const pending = ref(true)
const error = ref('')
const checkingOut = ref(false)
const showAddressForm = ref(false)
const showToast = ref(false)
const toastMessage = ref('')

const shippingAddress = ref<any>(null)
const tempAddress = reactive({
  street: '',
  city: '',
  state: '',
  zipCode: '',
  country: ''
})

const summary = computed(() => {
  const subtotal = cartItems.value.reduce((sum, item) => sum + ((item.unitPrice || 0) * item.quantity), 0)
  // Tax and shipping are determined by admin/sales team after order review.
  // We show subtotal only; final amounts will be confirmed in the proforma invoice.
  return {
    subtotal,
    taxAmount: null as number | null,   // TBD by sales team
    shippingAmount: null as number | null, // TBD by incoterms
    totalAmount: subtotal,
    currency: cartItems.value[0]?.currency || 'USD'
  }
})

const fetchCart = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await api.getCart()
    cartItems.value = res.items || []
    shippingAddress.value = res.shippingAddress || null
    // Play TTS based on cart state
    const isZh = locale.value === 'zh'
    if (cartItems.value.length === 0) {
      tts.playCartEmpty(isZh)
    } else {
      tts.playCartWelcome(isZh)
    }
  } catch (err: any) {
    error.value = err?.message || t('customer.cart.fetch_error')
  } finally {
    pending.value = false
  }
}

const updateQuantity = async (itemId: string, quantity: number) => {
  if (quantity < 1) return
  const item = cartItems.value.find(i => i.id === itemId)
  if (!item) return
  const prevQuantity = item.quantity
  item.quantity = quantity
  try {
    await api.updateCartItem(itemId, { quantity })
    showNotification(t('customer.cart.updated'))
  } catch {
    item.quantity = prevQuantity // rollback on failure
    showNotification(t('customer.cart.update_error'), true)
  }
}

const removeItem = async (itemId: string) => {
  try {
    await api.removeCartItem(itemId)
    cartItems.value = cartItems.value.filter(i => i.id !== itemId)
    showNotification(t('customer.cart.item_removed'))
  } catch {
    showNotification(t('customer.cart.remove_error'), true)
  }
}

const clearCart = async () => {
  if (!confirm(t('customer.cart.confirm_clear'))) return
  try {
    await api.clearCart()
    cartItems.value = []
    showNotification(t('customer.cart.cleared'))
  } catch {
    showNotification(t('customer.cart.clear_error'), true)
  }
}

const proceedToCheckout = async () => {
  checkingOut.value = true
  try {
    const res = await api.checkoutCart({
      shippingAddress: showAddressForm.value ? tempAddress : shippingAddress.value
    })
    const orderId = res.id || res.orderId
    if (orderId) {
      await navigateTo(`/customer/orders/${orderId}`)
    } else {
      showNotification(t('customer.cart.checkout_success'))
    }
  } catch (err: any) {
    showNotification(err?.message || t('customer.cart.checkout_error'), true)
  } finally {
    checkingOut.value = false
  }
}

const showNotification = (message: string, isError = false) => {
  toastMessage.value = message
  showToast.value = true
  setTimeout(() => { showToast.value = false }, 3000)
}

onMounted(fetchCart)
</script>

<style scoped>
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}
.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px);
}
</style>
