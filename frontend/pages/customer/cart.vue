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
          <NuxtLink :to="localePath('/customer/products')" class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-orange-100 text-orange-700 hover:bg-orange-200 transition-colors">
            <Icon name="heroicons:arrow-left" class="h-4 w-4" aria-hidden="true" />
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
        <Icon name="heroicons:exclamation-triangle" class="h-12 w-12 text-red-400 mx-auto mb-3" aria-hidden="true" />
        <p class="text-red-700 font-medium">{{ error }}</p>
        <button @click="fetchCart" class="mt-4 px-4 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors">
          {{ t('customer.cart.retry') }}
        </button>
      </div>

      <div v-else-if="cartItems.length === 0" class="text-center py-20">
        <div class="w-32 h-32 mx-auto mb-6 rounded-full bg-gradient-to-br from-orange-100 to-amber-100 flex items-center justify-center">
          <Icon name="heroicons:shopping-bag" class="h-16 w-16 text-orange-400" aria-hidden="true" />
        </div>
        <h2 class="text-2xl font-semibold text-gray-900 mb-2">{{ t('customer.cart.empty_title') }}</h2>
        <p class="text-gray-600 mb-6">{{ t('customer.cart.empty_desc') }}</p>
        <NuxtLink :to="localePath('/customer/products')" class="inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-orange-500 to-amber-500 text-white font-medium rounded-xl hover:from-orange-600 hover:to-amber-600 shadow-lg shadow-orange-200">
          <Icon name="heroicons:sparkles" class="h-5 w-5" aria-hidden="true" />
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
                  <img v-if="item.thumbnail" :src="item.thumbnail" :alt="itemDisplayName(item)" class="w-full h-full object-cover" />
                  <Icon v-else name="heroicons:cube" class="h-10 w-10 text-gray-300" aria-hidden="true" />
                </div>

                <!-- Product Details -->
                <div class="flex-1 min-w-0">
                  <div class="flex items-start justify-between gap-4">
                    <div>
                      <h3 class="text-lg font-semibold text-gray-900">{{ itemDisplayName(item) }}</h3>
                      <p class="text-sm text-gray-500 mt-0.5">{{ tField(item, 'category') || t('customer.cart.default_category') }}</p>
                      <div class="flex items-center gap-2 mt-2">
                        <span v-if="item.halalCertified" class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 text-xs font-medium">
                          <Icon name="heroicons:check-badge" class="h-3 w-3" aria-hidden="true" />
                          {{ t('customer.cart.badge_halal') }}
                        </span>
                        <span v-if="item.oemAvailable" class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-orange-50 text-orange-700 text-xs font-medium">
                          <Icon name="heroicons:sparkles" class="h-3 w-3" aria-hidden="true" />
                          {{ t('customer.cart.badge_oem') }}
                        </span>
                      </div>
                    </div>
                    <button @click="removeItem(item.id)" class="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors" :aria-label="t('customer.cart.remove_item')">
                      <Icon name="heroicons:trash" class="h-5 w-5" aria-hidden="true" />
                    </button>
                  </div>

                  <!-- Quantity & Price -->
                  <div class="flex items-end justify-between mt-4">
                    <div class="flex items-center gap-3">
                      <button @click="changeQuantity(item, item.quantity - 1)" :disabled="item.quantity <= 1" class="w-9 h-9 flex items-center justify-center rounded-lg border border-gray-200 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors" :aria-label="t('customer.cart.decrease_quantity')">
                        <Icon name="heroicons:minus" class="h-4 w-4" aria-hidden="true" />
                      </button>
                      <input v-model.number="item.quantity" type="number" min="1" :max="quantityMax(item)" @change="changeQuantity(item, item.quantity)" class="w-16 h-9 text-center border border-gray-200 rounded-lg font-medium focus:outline-none focus:ring-2 focus:ring-orange-500" />
                      <button @click="changeQuantity(item, item.quantity + 1)" :disabled="item.quantity >= quantityMax(item)" class="w-9 h-9 flex items-center justify-center rounded-lg border border-gray-200 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors" :aria-label="t('customer.cart.increase_quantity')">
                        <Icon name="heroicons:plus" class="h-4 w-4" aria-hidden="true" />
                      </button>
                      <span class="text-sm text-gray-500">{{ $t('product.moq_prefix') }}{{ item.moq || 1 }}</span>
                    </div>
                    <div class="text-right">
                      <p class="text-lg font-bold text-orange-600">{{ cur(item.currency) }} {{ formatNumber((item.unitPrice || 0) * item.quantity) }}</p>
                      <p class="text-xs text-gray-500">{{ t('customer.cart.unit_price_per', { currency: cur(item.currency), price: formatNumber(item.unitPrice || 0) }) }}</p>
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
              <Icon name="heroicons:trash" class="h-4 w-4" aria-hidden="true" />
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
                <span class="font-medium text-gray-900">{{ t('customer.cart.units_count', { count: cartItems.reduce((sum, i) => sum + i.quantity, 0) }) }}</span>
              </div>
              <div class="flex justify-between text-sm">
                <span class="text-gray-600">{{ t('customer.cart.subtotal') }}</span>
                <span class="font-medium text-gray-900">{{ cur(summary.currency) }} {{ formatNumber(summary.subtotal || 0) }}</span>
              </div>
              <div v-if="appliedCoupon" class="flex justify-between text-sm text-emerald-700">
                <span>{{ t('customer.cart.coupon_applied', { code: appliedCoupon.code }) }}</span>
                <span>-{{ cur(summary.currency) }} {{ formatNumber(appliedCoupon.amount || 0) }}</span>
              </div>
              <div class="pt-2 border-t border-dashed border-gray-100">
                <label class="block text-xs font-medium text-gray-600 mb-1">{{ t('customer.cart.coupon_code') }}</label>
                <div class="flex gap-2">
                  <input v-model="couponCode" type="text" :placeholder="t('customer.cart.coupon_placeholder')" class="flex-1 px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500 uppercase" />
                  <button type="button" :disabled="couponApplying || !couponCode.trim()" class="px-3 py-2 text-sm font-medium rounded-lg bg-orange-100 text-orange-700 hover:bg-orange-200 disabled:opacity-50" @click="applyCoupon">{{ t('customer.cart.apply_coupon') }}</button>
                </div>
                <button v-if="appliedCoupon" type="button" class="mt-2 text-xs text-gray-500 hover:text-red-600" @click="removeCoupon">{{ t('customer.cart.remove_coupon') }}</button>
                <p class="text-xs text-gray-400 mt-1">{{ t('customer.cart.coupon_validate_hint') }}</p>
              </div>
              <div class="flex justify-between text-sm">
                <span class="text-gray-600">{{ t('customer.cart.estimated_tax_ref') }}</span>
                <span v-if="summary.taxStatus === FEE_COMPUTED" class="font-medium text-gray-900">
                  {{ cur(summary.currency) }} {{ formatNumber(summary.taxAmount ?? 0) }}
                  <span v-if="summary.taxAmount === 0" class="text-gray-400 text-xs">{{ t('customer.cart.fees_computed_zero_hint') }}</span>
                </span>
                <span v-else class="font-medium text-gray-400 italic">{{ feePendingLabel(summary.taxStatus) }}</span>
              </div>
              <div class="flex justify-between text-sm">
                <span class="text-gray-600">{{ t('customer.cart.estimated_shipping_ref') }}</span>
                <span v-if="summary.shippingStatus === FEE_COMPUTED" class="font-medium text-gray-900">
                  {{ cur(summary.currency) }} {{ formatNumber(summary.shippingAmount ?? 0) }}
                  <span v-if="summary.shippingAmount === 0" class="text-gray-400 text-xs">{{ t('customer.cart.fees_computed_zero_hint') }}</span>
                </span>
                <span v-else class="font-medium text-gray-400 italic">{{ feePendingLabel(summary.shippingStatus) }}</span>
              </div>
              <div class="pt-4 border-t border-gray-100">
                <div class="flex justify-between">
                  <span class="text-base font-semibold text-gray-900">{{ t(summary.totalLabelKey) }}</span>
                  <span class="text-xl font-bold text-orange-600">{{ cur(summary.currency) }} {{ formatNumber(summary.totalAmount) }}</span>
                </div>
                <p class="text-xs text-gray-400 mt-1">{{ t('customer.cart.checkout_currency_note', { currency: cur(summary.currency) }) }}</p>
                <p class="text-xs text-gray-400 mt-0.5">{{ t(summary.feesDisclaimerKey) }}</p>
              </div>

              <!-- Shipping Address -->
              <div class="pt-4 border-t border-gray-100">
                <div class="flex items-center justify-between mb-2">
                  <h3 class="text-sm font-semibold text-gray-900">{{ t('customer.cart.shipping_to') }}</h3>
                  <button @click="toggleAddressForm" class="text-xs text-orange-600 hover:text-orange-700">
                    {{ showAddressForm ? t('customer.cart.hide') : t('customer.cart.change') }}
                  </button>
                </div>
                <div v-if="!showAddressForm && shippingAddress" class="text-sm text-gray-600">
                  <p>{{ shippingAddress.street }}</p>
                  <p>{{ shippingAddress.city }}, {{ shippingAddress.state }} {{ shippingAddress.zipCode }}</p>
                  <p>{{ shippingAddress.country }}</p>
                </div>
                <div v-else class="space-y-3">
                  <p v-if="!shippingAddress" class="text-xs text-amber-700 bg-amber-50 border border-amber-100 rounded-lg px-3 py-2">
                    {{ t('customer.cart.address_required_hint') }}
                  </p>
                  <select v-model="tempAddress.country" class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500">
                    <option value="">{{ t('customer.cart.country') }}</option>
                    <option v-for="opt in countryOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                  </select>
                  <input v-model="tempAddress.city" type="text" :placeholder="t('customer.cart.city')" class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
                  <input v-model="tempAddress.state" type="text" :placeholder="t('customer.cart.state')" class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
                  <input v-model="tempAddress.street" type="text" :placeholder="t('customer.cart.street')" class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
                  <input v-model="tempAddress.zipCode" type="text" :placeholder="t('customer.cart.zip')" class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500" />
                </div>
              </div>

              <!-- Incoterms：影响参考运费估算 -->
              <div class="pt-4 border-t border-gray-100">
                <label class="block text-sm font-semibold text-gray-900 mb-1">{{ t('customer.cart.incoterms_label') }}</label>
                <select v-model="selectedIncoterms" class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-orange-500">
                  <option v-for="opt in incotermsOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                </select>
                <p class="text-xs text-gray-400 mt-1">{{ t('customer.cart.incoterms_hint') }}</p>
                <p v-if="incotermsHint(selectedIncoterms)" class="text-xs text-amber-700 mt-0.5">{{ incotermsHint(selectedIncoterms) }}</p>
              </div>

              <!-- Action Buttons -->
              <div class="pt-4 space-y-3">
                <button @click="proceedToCheckout" :disabled="checkingOut" class="w-full py-3 bg-gradient-to-r from-orange-500 to-amber-500 text-white font-semibold rounded-xl hover:from-orange-600 hover:to-amber-600 shadow-lg shadow-orange-200 disabled:opacity-50 flex items-center justify-center gap-2">
                  <Icon v-if="checkingOut" name="heroicons:arrow-path" class="h-5 w-5 animate-spin" aria-hidden="true" />
                  <Icon v-else name="heroicons:credit-card" class="h-5 w-5" aria-hidden="true" />
                  {{ checkingOut ? t('customer.cart.processing') : t('customer.cart.proceed_checkout') }}
                </button>
                <NuxtLink :to="localePath('/customer/orders/new')" class="w-full py-3 bg-gray-100 text-gray-700 font-medium rounded-xl hover:bg-gray-200 transition-colors flex items-center justify-center gap-2">
                  <Icon name="heroicons:sparkles" class="h-5 w-5" aria-hidden="true" />
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

    <!-- 结账后合规/价格/库存警告确认（与后端 checkout 响应字段对齐） -->
    <Teleport to="body">
      <Transition name="slide-up">
        <div
          v-if="showCheckoutReviewModal"
          class="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4"
          role="dialog"
          aria-modal="true"
          :aria-label="t('customer.cart.checkout_review_title')"
        >
          <div class="bg-white rounded-2xl shadow-xl max-w-lg w-full max-h-[85vh] overflow-hidden flex flex-col">
            <div class="px-6 py-4 border-b border-amber-100 bg-amber-50">
              <h2 class="text-lg font-semibold text-amber-900">{{ t('customer.cart.checkout_review_title') }}</h2>
              <p class="text-sm text-amber-800 mt-1">{{ t('customer.cart.checkout_review_subtitle') }}</p>
            </div>
            <div class="px-6 py-4 overflow-y-auto space-y-4 text-sm">
              <div v-if="checkoutReviewCountry">
                <p class="font-medium text-gray-900">{{ t('customer.cart.checkout_destination') }}: {{ checkoutReviewCountry }}</p>
              </div>
              <div v-if="checkoutComplianceWarnings.length">
                <p class="font-medium text-gray-800 mb-2">{{ t('customer.cart.checkout_compliance_section') }}</p>
                <ul class="list-disc pl-5 space-y-1 text-amber-900">
                  <li v-for="(w, i) in checkoutComplianceWarnings" :key="'c'+i">{{ w }}</li>
                </ul>
              </div>
              <div v-if="checkoutInventoryWarnings.length">
                <p class="font-medium text-gray-800 mb-2">{{ t('customer.cart.checkout_inventory_section') }}</p>
                <ul class="list-disc pl-5 space-y-1 text-gray-800">
                  <li v-for="(w, i) in checkoutInventoryWarnings" :key="'i'+i">{{ w }}</li>
                </ul>
              </div>
              <div v-if="checkoutPricing" class="rounded-lg border border-blue-100 bg-blue-50/80 p-3 space-y-2">
                <p class="font-medium text-blue-900">{{ t('customer.cart.checkout_pricing_section') }}</p>
                <div class="grid grid-cols-2 gap-x-4 gap-y-1 text-gray-800">
                  <span>{{ t('customer.cart.checkout_pricing_subtotal') }}</span>
                  <span class="text-right font-medium">{{ cur(checkoutPricing.currency) }} {{ formatNumber(checkoutPricing.subtotal ?? 0) }}</span>
                  <span>{{ t('customer.cart.checkout_pricing_tax') }}</span>
                  <span class="text-right font-medium">{{ cur(checkoutPricing.currency) }} {{ formatNumber(checkoutPricing.taxAmount ?? 0) }}</span>
                  <span>{{ t('customer.cart.checkout_pricing_shipping') }}</span>
                  <span class="text-right font-medium">{{ cur(checkoutPricing.currency) }} {{ formatNumber(checkoutPricing.shippingAmount ?? 0) }}</span>
                  <span class="font-semibold">{{ t('customer.cart.checkout_pricing_total') }}</span>
                  <span class="text-right font-semibold text-orange-700">{{ cur(checkoutPricing.currency) }} {{ formatNumber(checkoutPricing.totalAmount ?? 0) }}</span>
                </div>
                <p class="text-xs text-blue-800/90">{{ checkoutPricingNoticeText }}</p>
              </div>
            </div>
            <div class="px-6 py-4 border-t border-gray-100 flex gap-3 justify-end bg-gray-50">
              <button
                type="button"
                class="px-4 py-2 rounded-lg border border-gray-300 text-gray-700 hover:bg-gray-100"
                @click="cancelCheckoutReview"
              >
                {{ t('customer.cart.checkout_review_stay') }}
              </button>
              <button
                type="button"
                class="px-4 py-2 rounded-lg bg-gradient-to-r from-orange-500 to-amber-500 text-white font-medium hover:from-orange-600 hover:to-amber-600"
                @click="confirmCheckoutReviewNavigate"
              >
                {{ t('customer.cart.checkout_review_continue') }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Toast：成功 / 失败样式区分 -->
    <Transition name="slide-up">
      <div
        v-if="showToast"
        :class="toastIsError ? 'bg-red-600' : 'bg-emerald-600'"
        class="fixed bottom-6 right-6 text-white px-6 py-3 rounded-xl shadow-lg flex items-center gap-3 z-50"
      >
        <Icon :name="toastIsError ? 'heroicons:exclamation-circle' : 'heroicons:check-circle'" class="h-5 w-5" aria-hidden="true" />
        {{ toastMessage }}
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useTranslation } from '~/composables/useTranslation'

definePageMeta({
  layout: 'customer',
  middleware: ['auth']
})

const { t, locale } = useI18n()
const { tField } = useTranslation()
const localePath = useLocalePath()
const { currencyOrDefault: cur, formatNumber } = useDisplay()
const api = useApi()
const tts = useTTS()
const { countryOptions } = useCountryOptions()
const { incotermsOptions, defaultIncoterms, incotermsHint } = useIncotermsOptions()
const selectedIncoterms = ref(defaultIncoterms)

const cartItems = ref<any[]>([])
/** 与后端 fee estimate status 对齐 */
const FEE_COMPUTED = 'computed'
const FEE_PENDING_DESTINATION = 'pending_destination'
const FEE_PENDING_RATES = 'pending_rates'
const FEE_PENDING_WEIGHT = 'pending_weight'

const cartSummary = ref<{
  taxAmount?: number | null
  shippingAmount?: number | null
  taxStatus?: string
  shippingStatus?: string
  pricingScope?: string
  subtotal?: number
} | null>(null)
const pending = ref(true)
const error = ref('')
const checkingOut = ref(false)
const showAddressForm = ref(false)
const showToast = ref(false)
const toastMessage = ref('')
const toastIsError = ref(false)
const showCheckoutReviewModal = ref(false)
const checkoutPendingOrderId = ref<string | null>(null)
const checkoutComplianceWarnings = ref<string[]>([])
const checkoutInventoryWarnings = ref<string[]>([])
const checkoutReviewCountry = ref('')
const checkoutPricing = ref<{
  subtotal?: number
  taxAmount?: number
  shippingAmount?: number
  totalAmount?: number
  currency?: string
  notice?: string
} | null>(null)
const couponCode = ref('')
const couponApplying = ref(false)
const appliedCoupon = ref<{ code: string; amount: number } | null>(null)

const shippingAddress = ref<any>(null)
const tempAddress = reactive({
  street: '',
  city: '',
  state: '',
  zipCode: '',
  country: ''
})

const summary = computed(() => {
  const subtotal = cartSummary.value?.subtotal ?? cartItems.value.reduce((sum, item) => sum + ((item.unitPrice || 0) * item.quantity), 0)
  const taxStatus = cartSummary.value?.taxStatus ?? FEE_PENDING_DESTINATION
  const shippingStatus = cartSummary.value?.shippingStatus ?? FEE_PENDING_DESTINATION
  const taxIncluded = taxStatus === FEE_COMPUTED
  const shipIncluded = shippingStatus === FEE_COMPUTED
  const taxAmount = taxIncluded ? (cartSummary.value?.taxAmount ?? 0) : null
  const shippingAmount = shipIncluded ? (cartSummary.value?.shippingAmount ?? 0) : null
  const couponDiscount = appliedCoupon.value?.amount ?? 0
  let totalAmount = subtotal - couponDiscount
  if (taxIncluded) totalAmount += taxAmount ?? 0
  if (shipIncluded) totalAmount += shippingAmount ?? 0

  let totalLabelKey = 'customer.cart.total_merchandise_only'
  if (taxIncluded && shipIncluded) {
    totalLabelKey = 'customer.cart.total_with_estimates'
  } else if (taxIncluded || shipIncluded) {
    totalLabelKey = 'customer.cart.total_partial_estimates'
  }

  let feesDisclaimerKey = 'customer.cart.fees_pending_disclaimer'
  if (taxIncluded && shipIncluded) {
    feesDisclaimerKey = 'customer.cart.fees_reference_disclaimer'
  } else if (taxIncluded || shipIncluded) {
    feesDisclaimerKey = 'customer.cart.fees_partial_disclaimer'
  }

  return {
    subtotal,
    taxAmount,
    shippingAmount,
    taxStatus,
    shippingStatus,
    totalAmount,
    totalLabelKey,
    feesDisclaimerKey,
    currency: cur(cartItems.value[0]?.currency)
  }
})

/** 未算出费用时的行内提示文案 */
const feePendingLabel = (status?: string) => {
  switch (status) {
    case FEE_PENDING_DESTINATION:
      return t('customer.cart.fees_pending_destination')
    case FEE_PENDING_WEIGHT:
      return t('customer.cart.fees_pending_weight')
    case FEE_PENDING_RATES:
    default:
      return t('customer.cart.fees_pending_rates')
  }
}

/** 结账弹窗 pricing.notice → i18n */
const checkoutPricingNoticeText = computed(() => {
  const notice = checkoutPricing.value?.notice
  if (notice === 'reference_until_pi') return t('customer.cart.checkout_pricing_notice_reference')
  if (notice === 'partial_until_pi') return t('customer.cart.checkout_pricing_notice_partial')
  return t('customer.cart.checkout_pricing_notice_pending')
})

/** 产品展示名：优先 translations.name，其次 name / productName */
const itemDisplayName = (item: any) => {
  const translated = tField(item, 'name')
  if (translated) return translated
  return item?.productName || item?.name || ''
}

/** 数量输入上限：确保当前数量不会触发 HTML5 invalid */
const quantityMax = (item: any) => {
  const cap = item?.maxQuantity
  const qty = item?.quantity ?? 1
  if (typeof cap === 'number' && cap > 0) {
    return Math.max(qty, cap)
  }
  return Math.max(qty, 999999)
}

/** 解析结账用收货地址：无已存地址或正在编辑时使用 tempAddress */
const resolveCheckoutAddress = () => {
  if (showAddressForm.value || !shippingAddress.value) {
    return { ...tempAddress }
  }
  return shippingAddress.value
}

/** 校验结账地址必填字段（与后端 cart checkout 对齐） */
const validateCheckoutAddress = (addr: typeof tempAddress) => {
  if (!addr.country?.trim()) return t('customer.cart.country')
  if (!addr.street?.trim()) return t('customer.cart.street')
  if (!addr.city?.trim()) return t('customer.cart.city')
  return ''
}

/** 当前用于税费/运费估算的目的地 */
const cartDestination = computed(() => {
  return resolveCheckoutAddress()?.country?.trim() || ''
})

const parseCartSummary = (res: Record<string, any>) => {
  const s = res?.summary
  const taxEst = res?.taxEstimate
  const shipEst = res?.shippingEstimate
  const taxStatus = (s?.taxStatus ?? taxEst?.status ?? FEE_PENDING_DESTINATION) as string
  const shippingStatus = (s?.shippingStatus ?? shipEst?.status ?? FEE_PENDING_DESTINATION) as string
  return {
    subtotal: typeof s?.subtotal === 'number' ? s.subtotal : undefined,
    taxStatus,
    shippingStatus,
    pricingScope: s?.pricingScope ?? 'cart_reference',
    taxAmount: taxStatus === FEE_COMPUTED ? (s?.taxAmount ?? taxEst?.amount ?? 0) : null,
    shippingAmount: shippingStatus === FEE_COMPUTED ? (s?.shippingAmount ?? shipEst?.cost ?? 0) : null
  }
}

const fetchCart = async () => {
  pending.value = true
  error.value = ''
  try {
    const params: Record<string, string> = {}
    const destination = cartDestination.value
    if (destination) {
      params.destination = destination
      const addr = resolveCheckoutAddress()
      if (addr?.state?.trim()) params.region = addr.state.trim()
    if (selectedIncoterms.value) params.incoterms = selectedIncoterms.value
    }
    const res = await api.getCart(Object.keys(params).length ? params : undefined)
    cartItems.value = res.items || []
    shippingAddress.value = res.shippingAddress || shippingAddress.value || null
    cartSummary.value = parseCartSummary(res)
    if (cartItems.value.length === 0) {
      tts.playCartEmpty()
    } else {
      tts.playCartWelcome()
    }
  } catch (err: any) {
    error.value = err?.message || t('customer.cart.fetch_error')
  } finally {
    pending.value = false
  }
}

const quantityTimers = new Map<string | number, ReturnType<typeof setTimeout>>()

/** 本地更新数量并 debounce 同步到服务端 */
const changeQuantity = (item: any, quantity: number) => {
  const max = quantityMax(item)
  const next = Math.min(Math.max(1, quantity), max)
  const prevQuantity = item.quantity
  item.quantity = next
  const itemId = item.id
  const existing = quantityTimers.get(itemId)
  if (existing) clearTimeout(existing)
  quantityTimers.set(itemId, setTimeout(async () => {
    quantityTimers.delete(itemId)
    try {
      await api.updateCartItem(String(itemId), { quantity: next })
      await fetchCart()
      showNotification(t('customer.cart.updated'))
    } catch {
      item.quantity = prevQuantity
      showNotification(t('customer.cart.update_error'), true)
    }
  }, 300))
}

const removeItem = async (itemId: string) => {
  try {
    await api.removeCartItem(itemId)
    await fetchCart()
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

/** 解析 checkout 响应中的 pricing 块 */
function extractCheckoutPricing(res: Record<string, any>) {
  const p = res?.pricing
  if (!p || typeof p !== 'object') return null
  return {
    subtotal: p.subtotal,
    taxAmount: p.taxAmount,
    shippingAmount: p.shippingAmount,
    totalAmount: p.totalAmount,
    currency: res?.currency || 'USD',
    notice: p.notice as string | undefined
  }
}

/** 解析后端 /user/cart/checkout 返回的合规与库存/价格警告 */
function extractCheckoutReview(res: Record<string, any>) {
  const compliance = res?.compliance
  const inventory = res?.inventory
  const country = typeof compliance?.country === 'string' ? compliance.country : ''
  const cw = Array.isArray(compliance?.warnings) ? compliance.warnings.filter(Boolean) : []
  const iw = Array.isArray(inventory?.warnings) ? inventory.warnings.filter(Boolean) : []
  return { country, complianceWarnings: cw as string[], inventoryWarnings: iw as string[] }
}

const proceedToCheckout = async () => {
  const addr = resolveCheckoutAddress()
  const missingField = validateCheckoutAddress(addr)
  if (missingField) {
    showNotification(t('customer.cart.address_required_hint'), true)
    return
  }
  checkingOut.value = true
  try {
    const payload: { shippingAddress: typeof tempAddress; couponCode?: string; incoterms?: string } = {
      shippingAddress: addr,
      incoterms: selectedIncoterms.value
    }
    if (couponCode.value.trim()) {
      payload.couponCode = couponCode.value.trim()
    }
    const res = await api.checkoutCart(payload)
    const orderId = (res?.id || res?.orderId) as string | undefined

    const { country, complianceWarnings, inventoryWarnings } = extractCheckoutReview(res)
    const pricing = extractCheckoutPricing(res)
    const pricingNeedsReview = pricing?.notice && pricing.notice !== 'reference_until_pi'
    const hasReview =
      complianceWarnings.length > 0 ||
      inventoryWarnings.length > 0 ||
      !!pricingNeedsReview

    if (orderId && hasReview) {
      checkoutPendingOrderId.value = orderId
      checkoutReviewCountry.value = country
      checkoutComplianceWarnings.value = complianceWarnings
      checkoutInventoryWarnings.value = inventoryWarnings
      checkoutPricing.value = pricing
      showCheckoutReviewModal.value = true
    } else if (orderId) {
      await navigateTo(localePath(`/customer/orders/${orderId}`))
    } else {
      showNotification(t('customer.cart.checkout_success'))
    }
  } catch (err: any) {
    showNotification(err?.message || t('customer.cart.checkout_error'), true)
  } finally {
    checkingOut.value = false
  }
}

/** 用户确认已阅读警告后跳转订单详情 */
const confirmCheckoutReviewNavigate = async () => {
  const id = checkoutPendingOrderId.value
  showCheckoutReviewModal.value = false
  checkoutPendingOrderId.value = null
  checkoutComplianceWarnings.value = []
  checkoutInventoryWarnings.value = []
  checkoutReviewCountry.value = ''
  checkoutPricing.value = null
  if (id) {
    await navigateTo(localePath(`/customer/orders/${id}`))
  }
}

/** 留在购物车（订单已创建，仅关闭弹窗并刷新） */
const cancelCheckoutReview = async () => {
  showCheckoutReviewModal.value = false
  checkoutPendingOrderId.value = null
  checkoutComplianceWarnings.value = []
  checkoutInventoryWarnings.value = []
  checkoutReviewCountry.value = ''
  checkoutPricing.value = null
  await fetchCart()
  showNotification(t('customer.cart.checkout_review_order_created'))
}

const showNotification = (message: string, isError = false) => {
  toastMessage.value = message
  toastIsError.value = isError
  showToast.value = true
  setTimeout(() => { showToast.value = false }, 3000)
}

/** 切换地址表单时预填当前地址 */
const toggleAddressForm = () => {
  if (!showAddressForm.value && shippingAddress.value) {
    Object.assign(tempAddress, {
      street: shippingAddress.value.street || '',
      city: shippingAddress.value.city || '',
      state: shippingAddress.value.state || '',
      zipCode: shippingAddress.value.zipCode || '',
      country: shippingAddress.value.country || ''
    })
  }
  showAddressForm.value = !showAddressForm.value
}

/** 验证并预填优惠券折扣 */
const applyCoupon = async () => {
  const code = couponCode.value.trim()
  if (!code || couponApplying.value) return
  couponApplying.value = true
  try {
    const subtotal = cartItems.value.reduce((sum, item) => sum + ((item.unitPrice || 0) * item.quantity), 0)
    const discount = await api.validateCartCoupon({ code, subtotal })
    appliedCoupon.value = {
      code,
      amount: discount?.amount || discount?.discountAmount || 0
    }
    showNotification(t('customer.cart.coupon_valid'))
  } catch (err: any) {
    appliedCoupon.value = null
    showNotification(err?.message || t('customer.cart.coupon_apply_failed'), true)
  } finally {
    couponApplying.value = false
  }
}

const removeCoupon = () => {
  couponCode.value = ''
  appliedCoupon.value = null
  showNotification(t('customer.cart.coupon_removed'))
}

onMounted(fetchCart)

// 目的地变更后 debounce 重新拉取税费/运费估算
let addressFetchTimer: ReturnType<typeof setTimeout> | null = null
watch(cartDestination, (dest) => {
  if (!dest || pending.value || cartItems.value.length === 0) return
  if (addressFetchTimer) clearTimeout(addressFetchTimer)
  addressFetchTimer = setTimeout(() => { fetchCart() }, 500)
})

watch(selectedIncoterms, () => {
  if (pending.value || cartItems.value.length === 0) return
  if (addressFetchTimer) clearTimeout(addressFetchTimer)
  addressFetchTimer = setTimeout(() => { fetchCart() }, 400)
})
</script>

<style scoped>
.slide-up-enter-active,
.slide-up-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}
.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px);
}
</style>
