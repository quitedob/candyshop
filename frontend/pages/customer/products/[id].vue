<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <NuxtLink :to="localePath('/customer/products')" class="inline-flex items-center text-sm text-orange-600 hover:text-orange-500 mb-6">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('customer.products.back_to_catalog') }}
      </NuxtLink>

      <div v-if="pending" class="flex items-center justify-center py-20">
        <div class="animate-spin rounded-full h-12 w-12 border-4 border-blue-200 border-t-blue-600"></div>
      </div>

      <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-xl p-6 text-center">
        <p class="text-red-700">{{ error }}</p>
        <NuxtLink :to="localePath('/customer/products')" class="mt-4 inline-block px-4 py-2 bg-red-100 text-red-700 rounded-lg">
          {{ t('customer.products.back_to_catalog') }}
        </NuxtLink>
      </div>

      <div v-else-if="product" class="grid grid-cols-1 lg:grid-cols-2 gap-12">
        <!-- Product Images -->
        <div class="space-y-4">
          <div class="relative aspect-square rounded-2xl overflow-hidden bg-gradient-to-br from-orange-50 to-amber-50">
            <img v-if="selectedImage" :src="selectedImage" :alt="product.name" class="w-full h-full object-cover" />
            <div v-else class="w-full h-full flex items-center justify-center">
              <Icon name="heroicons:cube" class="h-24 w-24 text-blue-200" />
            </div>
            <!-- Badges -->
            <div class="absolute top-4 left-4 flex flex-col gap-2">
              <span v-if="product.halalCertified" class="px-3 py-1.5 bg-emerald-400 text-emerald-900 text-sm font-bold rounded-full">
                {{ t('customer.products.halal_certified') }}
              </span>
              <span v-if="product.featured" class="px-3 py-1.5 bg-amber-400 text-amber-900 text-sm font-bold rounded-full">
                Featured
              </span>
            </div>
          </div>

          <!-- Thumbnail Gallery -->
          <div v-if="allImages.length > 1" class="grid grid-cols-6 gap-2">
            <button v-for="(img, idx) in allImages" :key="idx" @click="selectedImage = img" class="aspect-square rounded-lg overflow-hidden border-2 transition-colors" :class="selectedImage === img ? 'border-orange-500' : 'border-transparent hover:border-gray-200'">
              <img :src="img" class="w-full h-full object-cover" />
            </button>
          </div>
        </div>

        <!-- Product Info -->
        <div>
          <div class="text-sm text-orange-600 font-medium mb-1">{{ product.category }}</div>
          <h1 class="text-3xl font-bold text-gray-900 mb-3">{{ product.name }}</h1>
          <p class="text-gray-600 mb-6">{{ product.summary || product.description }}</p>

          <!-- Price & MOQ -->
          <div class="bg-white rounded-xl border border-orange-100 p-5 mb-6">
            <div class="flex items-end justify-between mb-4">
              <div>
                <p class="text-sm text-gray-500">{{ t('customer.products.unit_price') }}</p>
                <div class="flex items-center gap-2">
                  <p class="text-3xl font-bold text-orange-600">{{ cur(product.currency) }} {{ formatNumber(displayPrice) }}</p>
                  <span v-if="contractPrice !== null" class="px-2 py-0.5 bg-emerald-100 text-emerald-700 text-xs font-semibold rounded-full">{{ t('customer.products.contract_price') }}</span>
                  <span v-else-if="loadingPrice" class="text-xs text-gray-400">{{ t('customer.products.loading_price') }}</span>
                </div>
              </div>
              <div class="text-right">
                <p class="text-sm text-gray-500">{{ t('customer.products.moq') }}</p>
                <p class="text-xl font-semibold text-gray-900">{{ t('customer.products.moq_units', { count: product.moq || 1 }) }}</p>
              </div>
            </div>

            <!-- SKU Variant Selector -->
            <div v-if="variants.length > 0" class="mb-4 space-y-3">
              <div v-if="flavorOptions.length > 0">
                <p class="text-sm font-medium text-gray-700 mb-2">{{ t('customer.products.select_flavor') }}</p>
                <div class="flex flex-wrap gap-2">
                  <button v-for="flavor in flavorOptions" :key="flavor"
                    @click="selectedFlavor = flavor"
                    :class="['px-3 py-1.5 text-sm rounded-full border transition-colors', selectedFlavor === flavor ? 'bg-orange-600 text-white border-orange-600' : 'bg-white text-gray-700 border-gray-200 hover:border-orange-300']">
                    {{ flavor }}
                  </button>
                </div>
              </div>
              <div v-if="weightOptions.length > 0">
                <p class="text-sm font-medium text-gray-700 mb-2">{{ t('customer.products.select_weight') }}</p>
                <div class="flex flex-wrap gap-2">
                  <button v-for="weight in weightOptions" :key="weight"
                    @click="selectedWeight = weight"
                    :class="['px-3 py-1.5 text-sm rounded-full border transition-colors', selectedWeight === weight ? 'bg-orange-600 text-white border-orange-600' : 'bg-white text-gray-700 border-gray-200 hover:border-orange-300']">
                    {{ weight }}
                  </button>
                </div>
              </div>
              <div v-if="selectedVariant" class="text-xs text-gray-500 bg-gray-50 rounded-lg px-3 py-2">
                {{ t('customer.products.sku_label') }} {{ selectedVariant.sku }}
                <span v-if="selectedVariant.priceModifier !== 0" class="ml-2 font-medium" :class="selectedVariant.priceModifier > 0 ? 'text-red-600' : 'text-green-600'">
                  {{ selectedVariant.priceModifier > 0 ? '+' : '' }}{{ selectedVariant.priceModifier }}
                </span>
              </div>
            </div>

            <!-- Destination Country -->
            <div class="flex items-center gap-4 mb-4">
              <label for="dest-country" class="text-sm font-medium text-gray-700 whitespace-nowrap">{{ t('customer.products.destination_country') }}</label>
              <input
                id="dest-country"
                v-model="shippingCountry"
                type="text"
                :placeholder="t('customer.products.destination_country_placeholder')"
                class="flex-1 h-10 px-3 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
              />
            </div>

            <!-- Quantity Selector -->
            <div class="flex items-center gap-4 mb-4">
              <label class="text-sm font-medium text-gray-700">{{ t('customer.products.quantity') }}</label>
              <div class="flex items-center gap-2">
                <button @click="quantity = Math.max(product.moq || 1, quantity - 1)" class="w-10 h-10 flex items-center justify-center rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors">
                  <Icon name="heroicons:minus" class="h-4 w-4" />
                </button>
                <input v-model.number="quantity" type="number" :min="product.moq || 1" class="w-24 h-10 text-center border border-gray-200 rounded-lg font-medium focus:outline-none focus:ring-2 focus:ring-orange-500" />
                <button @click="quantity += 1" class="w-10 h-10 flex items-center justify-center rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors">
                  <Icon name="heroicons:plus" class="h-4 w-4" />
                </button>
              </div>
            </div>

            <!-- Total -->
            <div class="flex items-center justify-between pt-4 border-t border-gray-100">
              <span class="text-gray-600">{{ t('customer.products.estimated_total') }}</span>
              <span class="text-2xl font-bold text-gray-900">{{ cur(product.currency) }} {{ formatNumber(displayPrice * quantity) }}</span>
            </div>
          </div>

          <!-- OEM Available -->
          <div v-if="product.oemAvailable" class="bg-gradient-to-r from-orange-50 to-amber-50 rounded-xl border border-orange-100 p-5 mb-6">
            <div class="flex items-start gap-3">
              <Icon name="heroicons:sparkles" class="h-6 w-6 text-orange-600 flex-shrink-0 mt-0.5" />
              <div>
                <h3 class="font-semibold text-gray-900">{{ t('customer.products.oem_available') }}</h3>
                <p class="text-sm text-gray-600 mt-1">{{ t('customer.products.oem_private_label_hint') }}</p>
                <NuxtLink :to="localePath('/customer/oem-projects/new')" class="inline-flex items-center gap-1 mt-2 text-sm text-orange-600 hover:text-orange-700 font-medium">
                  {{ t('customer.products.start_oem') }}
                  <Icon name="heroicons:arrow-right" class="h-4 w-4" />
                </NuxtLink>
              </div>
            </div>
          </div>

          <!-- Actions: Inquire + Place Order -->
          <div class="space-y-3">
            <button @click="submitOrderRequest" :disabled="submitting" class="w-full py-3 bg-gradient-to-r from-orange-500 to-amber-600 text-white font-semibold rounded-xl hover:from-orange-600 hover:to-amber-700 shadow-lg shadow-orange-200 disabled:opacity-50 flex items-center justify-center gap-2">
              <Icon v-if="submitting" name="heroicons:arrow-path" class="h-5 w-5 animate-spin" />
              <Icon v-else name="heroicons:paper-airplane" class="h-5 w-5" />
              {{ submitting ? t('customer.products.submitting') : t('customer.products.submit_request') }}
            </button>
            <div class="flex gap-3">
              <NuxtLink :to="`${localePath('/customer/inquiries/new')}?product=${product.id}&name=${encodeURIComponent(product.name)}`" class="flex-1 py-3 border-2 border-orange-500 text-orange-600 font-semibold rounded-xl hover:bg-orange-50 transition-colors flex items-center justify-center gap-2">
                <Icon name="heroicons:chat-bubble-left" class="h-5 w-5" />
                {{ t('customer.products.inquire') }}
              </NuxtLink>
              <NuxtLink v-if="product.oemAvailable" :to="localePath('/customer/oem-projects/new')" class="px-6 py-3 border-2 border-orange-500 text-orange-600 font-semibold rounded-xl hover:bg-orange-50 transition-colors flex items-center justify-center gap-2">
                <Icon name="heroicons:sparkles" class="h-5 w-5" />
                {{ t('customer.products.oem_short') }}
              </NuxtLink>
            </div>
          </div>

          <!-- Success: We will contact you -->
          <Transition name="slide-up">
            <div v-if="showSuccess" class="mt-4 p-4 bg-emerald-50 border border-emerald-200 rounded-xl">
              <div class="flex items-start gap-3">
                <Icon name="heroicons:check-circle" class="h-6 w-6 text-emerald-600 flex-shrink-0 mt-0.5" />
                <div>
                  <p class="font-semibold text-emerald-800">{{ t('customer.products.order_submitted') }}</p>
                  <p class="text-sm text-emerald-700 mt-1">{{ t('customer.products.will_contact') }}</p>
                  <NuxtLink :to="localePath('/customer/orders')" class="inline-flex items-center gap-1 mt-2 text-sm font-medium text-emerald-700 hover:text-emerald-800">
                    {{ t('customer.nav.orders') }} →
                  </NuxtLink>
                </div>
              </div>
            </div>
          </Transition>

          <!-- Product Details -->
          <div class="mt-8 space-y-6">
            <div class="border-t border-gray-200 pt-6">
              <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ t('customer.products.product_details') }}</h3>
              <div class="grid grid-cols-2 gap-4">
                <div v-if="product.leadTime">
                  <p class="text-sm text-gray-500">{{ t('customer.products.lead_time') }}</p>
                  <p class="font-medium text-gray-900">{{ product.leadTime }}</p>
                </div>
                <div v-if="product.shelfLife">
                  <p class="text-sm text-gray-500">{{ t('customer.products.shelf_life') }}</p>
                  <p class="font-medium text-gray-900">{{ product.shelfLife }}</p>
                </div>
                <div v-if="product.storage">
                  <p class="text-sm text-gray-500">{{ t('customer.products.storage') }}</p>
                  <p class="font-medium text-gray-900">{{ product.storage }}</p>
                </div>
              </div>
            </div>

            <!-- Flavors -->
            <div v-if="product.flavors?.length">
              <h4 class="text-sm font-medium text-gray-700 mb-2">{{ t('customer.products.flavors') }}</h4>
              <div class="flex flex-wrap gap-2">
                <span v-for="flavor in product.flavors" :key="flavor" class="px-3 py-1.5 bg-gray-100 text-gray-700 rounded-full text-sm">
                  {{ flavor }}
                </span>
              </div>
            </div>

            <!-- Shapes -->
            <div v-if="product.shapes?.length">
              <h4 class="text-sm font-medium text-gray-700 mb-2">{{ t('customer.products.shapes') }}</h4>
              <div class="flex flex-wrap gap-2">
                <span v-for="shape in product.shapes" :key="shape" class="px-3 py-1.5 bg-orange-50 text-orange-700 rounded-full text-sm">
                  {{ shape }}
                </span>
              </div>
            </div>

            <!-- Certifications -->
            <div v-if="product.certifications?.length">
              <h4 class="text-sm font-medium text-gray-700 mb-2">{{ t('customer.products.certifications') }}</h4>
              <div class="flex flex-wrap gap-2">
                <span v-for="cert in product.certifications" :key="cert" class="px-3 py-1.5 bg-emerald-50 text-emerald-700 rounded-full text-sm">
                  {{ cert }}
                </span>
              </div>
            </div>

            <!-- Ingredients & Allergens -->
            <div v-if="product.ingredients">
              <h4 class="text-sm font-medium text-gray-700 mb-2">{{ t('customer.products.ingredients') }}</h4>
              <p class="text-sm text-gray-600">{{ product.ingredients }}</p>
            </div>
            <div v-if="product.allergens">
              <h4 class="text-sm font-medium text-gray-700 mb-2">{{ t('customer.products.allergens') }}</h4>
              <p class="text-sm text-amber-600">{{ product.allergens }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

definePageMeta({
  layout: 'customer',
  middleware: ['auth']
})

const { t } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, formatNumber } = useDisplay()
const api = useApi()
const route = useRoute()

const product = ref<any>(null)
const pending = ref(true)
const error = ref('')
const quantity = ref(1)
const selectedImage = ref('')
const contractPrice = ref<number | null>(null)
const loadingPrice = ref(false)
const shippingCountry = ref('')
const submitting = ref(false)
const showSuccess = ref(false)

// SKU variants
const variants = ref<any[]>([])
const selectedFlavor = ref('')
const selectedWeight = ref('')

const flavorOptions = computed(() => [...new Set(variants.value.filter(v => v.isActive && v.flavor).map(v => v.flavor))])
const weightOptions = computed(() => [...new Set(variants.value.filter(v => v.isActive && (!selectedFlavor.value || v.flavor === selectedFlavor.value) && v.weight).map(v => v.weight))])
const selectedVariant = computed(() => {
  if (!variants.value.length) return null
  return variants.value.find(v =>
    v.isActive &&
    (!selectedFlavor.value || v.flavor === selectedFlavor.value) &&
    (!selectedWeight.value || v.weight === selectedWeight.value)
  ) || null
})

const allImages = computed(() => {
  if (!product.value) return []
  const images = []
  if (product.value.thumbnail) images.push(product.value.thumbnail)
  if (product.value.images?.length) images.push(...product.value.images)
  return images
})

// Displayed unit price: contract price takes priority over public price
const displayPrice = computed(() => contractPrice.value ?? product.value?.basePrice ?? 0)

const fetchProduct = async () => {
  pending.value = true
  error.value = ''
  try {
    product.value = await api.getProduct(route.params.id as string)
    selectedImage.value = product.value.thumbnail || ''
    quantity.value = product.value.moq || 1
    // Fetch contract price and variants in background
    fetchContractPrice()
    fetchVariants()
  } catch (err: any) {
    error.value = err?.message || t('errors.api.product_load_failed')
  } finally {
    pending.value = false
  }
}

const fetchVariants = async () => {
  if (!product.value?.id) return
  try {
    const res = await api.get<any[]>(`/public/products/${product.value.id}/variants`)
    variants.value = res || []
    if (variants.value.length > 0) {
      selectedFlavor.value = variants.value[0]?.flavor || ''
      selectedWeight.value = variants.value[0]?.weight || ''
    }
  } catch {
    variants.value = []
  }
}

const fetchContractPrice = async () => {
  if (!product.value?.id) return
  loadingPrice.value = true
  try {
    const res = await api.get<any>(`/user/products/${product.value.id}/price?quantity=${quantity.value}`)
    if (res?.unitPrice != null) contractPrice.value = res.unitPrice
  } catch {
    // No contract price — fall back to public price silently
  } finally {
    loadingPrice.value = false
  }
}

// Re-fetch price when quantity changes (tier pricing)
watch(quantity, () => { fetchContractPrice() })

const submitOrderRequest = async () => {
  submitting.value = true
  try {
    // Create order directly with the selected product
    await api.createOrder({
      items: [{
        productId: product.value.id,
        quantity: quantity.value,
        unitPrice: displayPrice.value,
        specifications: selectedVariant.value?.sku || ''
      }],
      shippingAddress: { country: shippingCountry.value || '' } as any
    })
    showSuccess.value = true
  } catch (err: any) {
    // Fallback: add to cart then navigate to cart
    try {
      await api.addToCart(product.value.id, {
        quantity: quantity.value,
        unitPrice: displayPrice.value,
        specifications: selectedVariant.value?.sku || ''
      })
      await navigateTo(localePath('/customer/cart'))
    } catch (cartErr: any) {
      alert(cartErr?.message || t('errors.api.cart_submit_failed'))
    }
  } finally {
    submitting.value = false
  }
}

onMounted(fetchProduct)
</script>

<style scoped>
.slide-up-enter-active, .slide-up-leave-active { transition: opacity 0.3s ease, transform 0.3s ease; }
.slide-up-enter-from, .slide-up-leave-to { opacity: 0; transform: translateY(10px); }
</style>
