<template>
  <div>
    <div class="mb-6">
      <NuxtLink :to="localePath('/customer/inquiries')" class="flex items-center text-sm font-medium text-orange-600 hover:text-orange-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('customer.inquiries.back') }}
      </NuxtLink>
    </div>

    <div v-if="pending" class="text-center py-10">
      <p class="text-gray-500">{{ t('customer.inquiries.loading_details') }}</p>
    </div>

    <div v-else-if="error" class="bg-red-50 p-4 rounded-md">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <div v-else-if="inquiry" class="bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="px-4 py-5 sm:px-6 flex justify-between items-center bg-gray-50 border-b border-gray-200">
        <div>
          <h3 class="text-lg leading-6 font-medium text-gray-900">Inquiry #{{ inquiry.id.substring(0, 8) }}</h3>
          <p class="mt-1 max-w-2xl text-sm text-gray-500">
            {{ t('customer.inquiries.submitted_on') }} {{ new Date(inquiry.createdAt).toLocaleDateString() }}
          </p>
        </div>
        <span :class="[statusClass(inquiry.status), 'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5']">
          {{ inquiry.status }}
        </span>
      </div>

      <div class="px-4 py-5 sm:p-6">
        <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
          <div>
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.company') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ inquiry.companyName || t('customer.common.na') }}</dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.contact_person') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ inquiry.contactPerson || t('customer.common.na') }}</dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.email') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ inquiry.email || t('customer.common.na') }}</dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.whatsapp') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ inquiry.whatsapp || t('customer.common.na') }}</dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.target_country') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ inquiry.targetCountry || t('customer.common.na') }}</dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.estimated_quantity') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ inquiry.estimatedQuantity || t('customer.common.na') }}</dd>
          </div>
          <div class="sm:col-span-2">
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.interested_products') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ formatProducts(inquiry.interestedProducts) }}</dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.packaging_requirements') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ inquiry.packagingRequirements || t('customer.common.na') }}</dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.flavor_requirements') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ inquiry.flavorRequirements || t('customer.common.na') }}</dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.oem_needed') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ inquiry.oemNeeded ? t('customer.common.yes') : t('customer.common.no') }}</dd>
          </div>
          <div>
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.expected_delivery') }}</dt>
            <dd class="mt-1 text-sm text-gray-900">{{ inquiry.expectedDelivery || t('customer.common.na') }}</dd>
          </div>
          <div class="sm:col-span-2">
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.message') }}</dt>
            <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ inquiry.message || t('customer.inquiries.no_message') }}</dd>
          </div>
          <div class="sm:col-span-2" v-if="inquiry.customerNotes">
            <dt class="text-sm font-medium text-gray-500">{{ t('customer.inquiries.business_notes') }}</dt>
            <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ inquiry.customerNotes }}</dd>
          </div>
        </dl>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const route = useRoute()
const api = useApi()
const { t } = useI18n()
const localePath = useLocalePath()

const inquiry = ref<any>(null)
const pending = ref(true)
const error = ref('')

const fetchInquiry = async () => {
  pending.value = true; error.value = ''
  try { inquiry.value = await api.get<any>(`/user/inquiries/${route.params.id}`) }
  catch (err: any) { error.value = err?.message || t('errors.api.load_failed') }
  finally { pending.value = false }
}

const formatProducts = (products: unknown) => {
  if (Array.isArray(products)) return products.length > 0 ? products.join(', ') : t('customer.inquiries.general_inquiry')
  if (typeof products === 'string') return products.trim() !== '' ? products : t('customer.inquiries.general_inquiry')
  return t('customer.inquiries.general_inquiry')
}

const statusClass = (status: string) => {
  if (status === 'pending') return 'bg-yellow-100 text-yellow-800'
  if (status === 'quoted' || status === 'contacted') return 'bg-orange-100 text-orange-800'
  if (status === 'negotiating') return 'bg-amber-100 text-amber-800'
  if (status === 'won') return 'bg-green-100 text-green-800'
  if (status === 'lost' || status === 'closed') return 'bg-gray-100 text-gray-800'
  return 'bg-gray-100 text-gray-800'
}

onMounted(fetchInquiry)
</script>
