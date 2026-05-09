<template>
  <div class="bg-white shadow overflow-hidden sm:rounded-lg">
    <div class="px-4 py-5 sm:px-6 border-b border-gray-200 flex justify-between items-center">
      <div>
        <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.inquiries_new.title') }}</h3>
        <p class="mt-1 max-w-2xl text-sm text-gray-500">{{ t('customer.inquiries_new.subtitle') }}</p>
      </div>
      <NuxtLink :to="localePath('/customer/inquiries')" class="text-sm font-medium text-gray-500 hover:text-gray-700">
        {{ t('customer.inquiries_new.cancel') }}
      </NuxtLink>
    </div>

    <div class="px-4 py-5 sm:p-6">
      <form @submit.prevent="handleSubmit" class="space-y-6">
        <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-2">
          <div class="sm:col-span-2">
            <label for="products" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries_new.interested_products') }}</label>
            <div class="mt-1">
              <input type="text" id="products" v-model="form.interestedProducts"
                :placeholder="t('customer.inquiries_new.interested_products_placeholder')" required
                class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-orange-500 focus:border-orange-500 sm:text-sm" />
            </div>
            <p class="mt-2 text-sm text-gray-500">{{ t('customer.inquiries_new.interested_products_hint') }}</p>
          </div>

          <div>
            <label for="quantity" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries_new.estimated_quantity') }}</label>
            <div class="mt-1">
              <input type="text" id="quantity" v-model="form.estimatedQuantity"
                :placeholder="t('customer.inquiries_new.estimated_quantity_placeholder')" required
                class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-orange-500 focus:border-orange-500 sm:text-sm" />
            </div>
          </div>

          <div>
            <label for="delivery" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries_new.expected_delivery') }}</label>
            <div class="mt-1">
              <input type="text" id="delivery" v-model="form.expectedDelivery"
                :placeholder="t('customer.inquiries_new.expected_delivery_placeholder')"
                class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-orange-500 focus:border-orange-500 sm:text-sm" />
            </div>
          </div>

          <div class="sm:col-span-2">
            <label class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries_new.oem_question') }}</label>
            <div class="mt-2 space-y-4 sm:flex sm:items-center sm:space-y-0 sm:space-x-10">
              <div class="flex items-center">
                <input id="oem-yes" name="oem" type="radio" :value="true" v-model="form.oemNeeded"
                  class="focus:ring-orange-500 h-4 w-4 text-orange-600 border-gray-300">
                <label for="oem-yes" class="ml-3 block text-sm font-medium text-gray-700">{{ t('customer.inquiries_new.yes') }}</label>
              </div>
              <div class="flex items-center">
                <input id="oem-no" name="oem" type="radio" :value="false" v-model="form.oemNeeded"
                  class="focus:ring-orange-500 h-4 w-4 text-orange-600 border-gray-300">
                <label for="oem-no" class="ml-3 block text-sm font-medium text-gray-700">{{ t('customer.inquiries_new.no') }}</label>
              </div>
            </div>
          </div>

          <div class="sm:col-span-2">
            <label for="message" class="block text-sm font-medium text-gray-700">{{ t('customer.inquiries_new.additional_details') }}</label>
            <div class="mt-1">
              <textarea id="message" v-model="form.message" rows="4"
                :placeholder="t('customer.inquiries_new.additional_details_placeholder')"
                class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-orange-500 focus:border-orange-500 sm:text-sm"></textarea>
            </div>
          </div>
        </div>

        <div v-if="error" class="text-sm text-red-600 flex items-center">
          <Icon name="heroicons:exclamation-circle" class="h-5 w-5 mr-1" />
          {{ error }}
        </div>

        <div v-if="success" class="p-4 bg-green-50 rounded-md flex items-center text-sm text-green-700">
          <Icon name="heroicons:check-circle" class="h-5 w-5 mr-2 text-green-500" />
          {{ t('customer.inquiries_new.success') }}
        </div>

        <div class="pt-5 border-t border-gray-200">
          <div class="flex justify-end">
            <NuxtLink :to="localePath('/customer/inquiries')" class="bg-white py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50">
              {{ t('customer.inquiries_new.cancel') }}
            </NuxtLink>
            <button type="submit" :disabled="loading"
              class="ml-3 inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-orange-600 hover:bg-orange-700 disabled:opacity-50">
              {{ loading ? t('customer.inquiries_new.submitting') : t('customer.inquiries_new.submit') }}
            </button>
          </div>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const { user } = useAuth()
const { t } = useI18n()
const localePath = useLocalePath()
const route = useRoute()
const { submitCustomerInquiry } = useApi()

// Pre-fill product name from query params
const prefillProducts = computed(() => {
  const name = route.query.name as string
  const products = route.query.products as string
  return name || products || ''
})

const form = reactive({
  companyName: user.value?.company || '',
  contactPerson: `${user.value?.firstName || ''} ${user.value?.lastName || ''}`.trim(),
  email: user.value?.email || '',
  interestedProducts: prefillProducts.value,
  estimatedQuantity: '',
  oemNeeded: false,
  expectedDelivery: '',
  message: ''
})

const loading = ref(false)
const error = ref('')
const success = ref(false)

const handleSubmit = async () => {
  loading.value = true; error.value = ''; success.value = false
  try {
    const interestedProducts = form.interestedProducts.split(',').map(item => item.trim()).filter(Boolean)
    await submitCustomerInquiry({
      companyName: form.companyName, contactPerson: form.contactPerson, email: form.email,
      estimatedQuantity: form.estimatedQuantity, oemNeeded: form.oemNeeded,
      expectedDelivery: form.expectedDelivery, message: form.message, interestedProducts
    })
    success.value = true
    setTimeout(() => navigateTo(localePath('/customer/inquiries')), 2000)
  } catch (err: any) {
    error.value = err?.message || t('customer.inquiries_new.error')
  } finally { loading.value = false }
}
</script>
