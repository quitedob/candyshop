<template>
  <div>
    <div class="mb-6">
      <NuxtLink :to="localePath('/admin/inquiries')" class="flex items-center text-sm font-medium text-orange-600 hover:text-orange-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('admin.inquiryDetail.back') }}
      </NuxtLink>
    </div>

    <div v-if="pending" class="text-center py-10">
      <p class="text-gray-500">{{ t('admin.inquiryDetail.loading_details') }}</p>
    </div>

    <div v-else-if="error" class="bg-red-50 p-4 rounded-md">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <div v-else-if="inquiry" class="space-y-6">
      <!-- Inquiry Info Card -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 flex justify-between items-center bg-gray-50 border-b border-gray-200">
          <div>
            <h3 class="text-lg leading-6 font-medium text-gray-900">Inquiry #{{ inquiry.id.substring(0, 8) }}</h3>
            <p class="mt-1 max-w-2xl text-sm text-gray-500">
              {{ t('admin.inquiryDetail.submitted_on') }} {{ inquiry.createdAt ? new Date(inquiry.createdAt).toLocaleString() : cell(null) }}
            </p>
          </div>
          <div class="flex items-center gap-3">
            <span :class="[statusClass(inquiry.status), 'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5']">
              {{ enumLabel('inquiry_status', inquiry.status) }}
            </span>
            <span v-if="inquiry.priority" :class="[priorityClass(inquiry.priority), 'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5']">
              {{ enumLabel('inquiry_priority', inquiry.priority) }}
            </span>
          </div>
        </div>

        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
            <div v-if="inquiry.userId">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.user') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.userId }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.company_name') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ cell(inquiry.companyName) }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.contact_person') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ cell(inquiry.contactPerson) }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.email') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ cell(inquiry.email) }}</dd>
            </div>
            <div v-if="inquiry.whatsapp">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.whatsapp') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.whatsapp }}</dd>
            </div>
            <div v-if="inquiry.targetCountry">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.target_country') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.targetCountry }}</dd>
            </div>
            <div v-if="inquiry.estimatedQuantity">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.estimated_quantity') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.estimatedQuantity }}</dd>
            </div>
            <div v-if="inquiry.expectedDelivery">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.expected_delivery') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.expectedDelivery }}</dd>
            </div>
            <div v-if="inquiry.assignedTo">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.assigned_to') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.assignedTo }}</dd>
            </div>
            <div v-if="inquiry.quotedAmount" class="sm:col-span-2">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.quoted_amount') }}</dt>
              <dd class="mt-1 text-sm font-medium text-gray-900">{{ cur(inquiry.currency) }} {{ (inquiry.quotedAmount || 0).toLocaleString() }}</dd>
            </div>
            <div class="sm:col-span-2" v-if="inquiry.interestedProducts && inquiry.interestedProducts.length">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.interested_products') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ Array.isArray(inquiry.interestedProducts) ? inquiry.interestedProducts.join(', ') : inquiry.interestedProducts }}</dd>
            </div>
            <div v-if="inquiry.packagingRequirements">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.packaging_requirements') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.packagingRequirements }}</dd>
            </div>
            <div v-if="inquiry.flavorRequirements">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.flavor_requirements') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ inquiry.flavorRequirements }}</dd>
            </div>
            <div v-if="inquiry.oemNeeded">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.oem_needed') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ t('admin.inquiryDetail.yes') }}</dd>
            </div>
            <div class="sm:col-span-2" v-if="inquiry.message">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.message') }}</dt>
              <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ inquiry.message }}</dd>
            </div>
            <div class="sm:col-span-2" v-if="inquiry.customerNotes">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.inquiryDetail.business_notes') }}</dt>
              <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ inquiry.customerNotes }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- AI Compliance Check Results -->
      <div v-if="inquiry.aiComplianceCheck" class="bg-amber-50 shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-amber-100 border-b border-amber-200">
          <h3 class="text-lg leading-6 font-medium text-amber-900">{{ t('admin.inquiryDetail.ai_compliance_check') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-4 sm:grid-cols-2">
            <div>
              <dt class="text-sm font-medium text-amber-700">{{ t('admin.inquiryDetail.compliance_status') }}</dt>
              <dd class="mt-1">
                <span :class="[inquiry.aiComplianceCheck.passed ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800', 'inline-flex rounded-full px-2 text-xs font-semibold leading-5']">
                  {{ inquiry.aiComplianceCheck.passed ? t('admin.inquiryDetail.passed') : t('admin.inquiryDetail.failed') }}
                </span>
              </dd>
            </div>
            <div v-if="inquiry.aiComplianceCheck.warnings && inquiry.aiComplianceCheck.warnings.length">
              <dt class="text-sm font-medium text-amber-700">{{ t('admin.inquiryDetail.warnings') }}</dt>
              <dd class="mt-1 text-sm text-amber-900">
                <ul class="list-disc list-inside">
                  <li v-for="(warning, idx) in inquiry.aiComplianceCheck.warnings" :key="idx">{{ warning }}</li>
                </ul>
              </dd>
            </div>
            <div v-if="inquiry.aiComplianceCheck.summary" class="sm:col-span-2">
              <dt class="text-sm font-medium text-amber-700">{{ t('admin.inquiryDetail.summary') }}</dt>
              <dd class="mt-1 text-sm text-amber-900">{{ inquiry.aiComplianceCheck.summary }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Quotation Section -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.inquiryDetail.quotation') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <form @submit.prevent="submitQuote" class="space-y-4">
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.quoted_amount') }}</label>
                <input v-model.number="quoteForm.quotedAmount" type="number" min="0" step="0.01" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.valid_until') }}</label>
                <input v-model="quoteForm.validUntil" type="datetime-local" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.notes') }}</label>
              <textarea v-model="quoteForm.customerNotes" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div v-if="quoteMessage" class="text-sm" :class="quoteError ? 'text-red-600' : 'text-green-600'">
              {{ quoteMessage }}
            </div>
            <div class="flex justify-end">
              <button type="submit" :disabled="submittingQuote" class="inline-flex justify-center rounded-md border border-transparent bg-emerald-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-emerald-700 disabled:opacity-50">
                {{ submittingQuote ? t('admin.inquiryDetail.submitting') : t('admin.inquiryDetail.submit_quote') }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Convert to Order -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.inquiryDetail.convert_to_order') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <p class="text-sm text-gray-600 mb-4">{{ t('admin.inquiryDetail.convert_description') }}</p>
          <form @submit.prevent="convertToOrder" class="space-y-4">
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.shipping_street') }}</label>
                <input v-model="convertForm.street" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.shipping_city') }}</label>
                <input v-model="convertForm.city" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.shipping_state') }}</label>
                <input v-model="convertForm.state" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.shipping_zip') }}</label>
                <input v-model="convertForm.zipCode" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.inquiryDetail.shipping_country') }}</label>
                <input v-model="convertForm.country" type="text" :placeholder="inquiry.targetCountry || ''" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
            </div>
            <div v-if="convertMessage" class="text-sm" :class="convertError ? 'text-red-600' : 'text-green-600'">
              {{ convertMessage }}
            </div>
            <div class="flex justify-end">
              <button type="submit" :disabled="converting" class="inline-flex justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700 disabled:opacity-50">
                {{ converting ? t('admin.inquiryDetail.converting') : t('admin.inquiryDetail.convert_button') }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- Activity Log -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.inquiryDetail.activity_log') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <div v-if="!inquiry.statusHistory || inquiry.statusHistory.length === 0" class="text-center text-gray-500">
            {{ t('admin.inquiryDetail.no_activity') }}
          </div>
          <ul v-else class="border-l-2 border-gray-200 ml-3 space-y-4">
            <li v-for="(entry, idx) in inquiry.statusHistory" :key="idx" class="ml-4">
              <div class="flex items-start">
                <div class="flex-shrink-0 w-2 h-2 mt-2 rounded-full bg-gray-400"></div>
                <div class="ml-3">
                  <p class="text-sm text-gray-900">{{ entry.status }} {{ entry.note ? `- ${entry.note}` : '' }}</p>
                  <p class="text-xs text-gray-500">{{ entry.timestamp ? new Date(entry.timestamp).toLocaleString() : '-' }}</p>
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
import { onMounted, ref, reactive } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const route = useRoute()
const { token } = useAuth()
const { t } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, cell, enumLabel } = useDisplay()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'

const inquiry = ref<any>(null)
const pending = ref(true)
const error = ref('')

const submittingQuote = ref(false)
const quoteMessage = ref('')
const quoteError = ref(false)

const converting = ref(false)
const convertMessage = ref('')
const convertError = ref(false)

const quoteForm = reactive({
  quotedAmount: 0,
  validUntil: '',
  customerNotes: ''
})

const convertForm = reactive({
  street: '',
  city: '',
  state: '',
  zipCode: '',
  country: ''
})

const fetchInquiry = async () => {
  pending.value = true
  error.value = ''
  try {
    inquiry.value = await $fetch<any>(`${baseURL}/admin/inquiries/${route.params.id}`, {
      headers: { Authorization: `Bearer ${token.value}` }
    })
    quoteForm.quotedAmount = inquiry.value.quotedAmount || 0
    quoteForm.validUntil = inquiry.value.validUntil ? new Date(inquiry.value.validUntil).toISOString().slice(0, 16) : ''
    quoteForm.customerNotes = inquiry.value.customerNotes || ''
  } catch (err: any) {
    error.value = err?.data?.message || err.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const submitQuote = async () => {
  submittingQuote.value = true
  quoteMessage.value = ''
  quoteError.value = false

  try {
    await $fetch(`${baseURL}/admin/inquiries/${route.params.id}/quote`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token.value}` },
      body: {
        quotedAmount: quoteForm.quotedAmount,
        validUntil: quoteForm.validUntil ? new Date(quoteForm.validUntil).toISOString() : '',
        customerNotes: quoteForm.customerNotes
      }
    })
    quoteMessage.value = t('admin.inquiryDetail.quote_submitted')
    await fetchInquiry()
  } catch (err: any) {
    quoteError.value = true
    quoteMessage.value = err?.data?.message || err.message || t('errors.api.quote_failed')
  } finally {
    submittingQuote.value = false
  }
}

const convertToOrder = async () => {
  converting.value = true
  convertMessage.value = ''
  convertError.value = false

  try {
    const result = await $fetch<any>(`${baseURL}/admin/inquiries/${route.params.id}/convert-to-order`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token.value}` },
      body: {
        shippingAddress: {
          street: convertForm.street,
          city: convertForm.city,
          state: convertForm.state,
          zipCode: convertForm.zipCode,
          country: convertForm.country
        }
      }
    })
    convertMessage.value = t('admin.inquiryDetail.converted_success') + (result.orderId ? ` Order ID: ${result.orderId}` : '')
    await fetchInquiry()
  } catch (err: any) {
    convertError.value = true
    convertMessage.value = err?.data?.message || err.message || t('errors.api.inquiry_convert_failed')
  } finally {
    converting.value = false
  }
}

const statusClass = (status: string) => {
  if (status === 'pending') return 'bg-yellow-100 text-yellow-800'
  if (status === 'contacted' || status === 'quoted') return 'bg-orange-100 text-orange-800'
  if (status === 'negotiating') return 'bg-amber-100 text-amber-800'
  if (status === 'won') return 'bg-green-100 text-green-800'
  if (status === 'lost' || status === 'closed') return 'bg-gray-100 text-gray-800'
  return 'bg-gray-100 text-gray-800'
}

const priorityClass = (priority: string) => {
  if (priority === 'high' || priority === 'urgent') return 'bg-red-100 text-red-800'
  if (priority === 'normal') return 'bg-orange-100 text-orange-800'
  return 'bg-gray-100 text-gray-800'
}

onMounted(fetchInquiry)
</script>
