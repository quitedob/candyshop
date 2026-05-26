<template>
  <div>
    <div class="mb-6">
      <NuxtLink :to="localePath('/admin/companies')" class="flex items-center text-sm font-medium text-orange-600 hover:text-orange-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('admin.companies.back') }}
      </NuxtLink>
    </div>

    <div v-if="pending" class="text-center py-10">
      <p class="text-gray-500">{{ t('admin.companies.loading_details') }}</p>
    </div>

    <div v-else-if="error" class="bg-red-50 p-4 rounded-md">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <div v-else-if="company" class="space-y-6">
      <!-- Company Info Card -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 flex justify-between items-center bg-gray-50 border-b border-gray-200">
          <div>
            <h3 class="text-lg leading-6 font-medium text-gray-900">{{ company.name }}</h3>
            <p class="mt-1 max-w-2xl text-sm text-gray-500">{{ t('admin.companies.company_details') }}</p>
          </div>
          <span :class="[statusBadgeClass(company.kybStatus), 'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5']">
            {{ enumLabel('kyb_status', company.kybStatus) }}
          </span>
        </div>

        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.companies.name') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ company.name }}</dd>
            </div>
            <div v-if="company.taxId">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.companies.tax_id') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ company.taxId }}</dd>
            </div>
            <div v-if="company.registrationNumber">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.companies.registration_number') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ company.registrationNumber }}</dd>
            </div>
            <div v-if="company.address">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.companies.address') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ formatAddress(company.address) }}</dd>
            </div>
            <div v-if="company.phone">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.companies.phone') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ company.phone }}</dd>
            </div>
            <div v-if="company.website">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.companies.website') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ company.website }}</dd>
            </div>
            <div v-if="company.businessLicenseUrl">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.companies.business_license') }}</dt>
              <dd class="mt-1 text-sm">
                <a :href="company.businessLicenseUrl" target="_blank" rel="noopener noreferrer" class="text-orange-600 hover:text-orange-900">{{ t('admin.companies.view_document') }}</a>
              </dd>
            </div>
            <div v-if="company.kybVerifiedAt">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.companies.verified_at') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ formatDate(company.kybVerifiedAt, { dateStyle: 'medium', timeStyle: 'short' }) }}</dd>
            </div>
            <div v-if="company.kybReviewedBy">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.companies.reviewed_by') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ company.kybReviewedBy }}</dd>
            </div>
            <div class="sm:col-span-2" v-if="company.kybNotes">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.companies.kyb_notes') }}</dt>
              <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ company.kybNotes }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Credit & Pricing -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.companies.credit_pricing') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-3">
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.companies.credit_limit') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ company.creditLimit ? `${cur(company.currency)} ${formatNumber(company.creditLimit)}` : cell(null) }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.companies.payment_terms') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ cell(company.paymentTerms) }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.companies.price_list') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ cell(company.priceListName) }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Linked Users -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.companies.linked_users') }}</h3>
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.companies.user_name') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.companies.user_email') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.companies.user_role') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.companies.user_status') }}</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-if="!company.users || company.users.length === 0">
                <td colspan="4" class="px-6 py-4 text-center text-sm text-gray-500">{{ t('admin.companies.no_users') }}</td>
              </tr>
              <tr v-else v-for="user in company.users" :key="user.id">
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{ user.firstName }} {{ user.lastName }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ user.email }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ cell(user.role?.name || user.role) }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ enumLabel('company_status', user.status) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Status Management -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.companies.status_management') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <form @submit.prevent="updateStatus" class="space-y-4">
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label for="company-kybStatus" class="block text-sm font-medium text-gray-700">{{ t('admin.companies.kyb_status') }}</label>
                <select id="company-kybStatus" name="kybStatus" v-model="statusInput" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                  <option value="pending">{{ t('admin.companies.pending') }}</option>
                  <option value="verified">{{ t('admin.companies.verified') }}</option>
                  <option value="rejected">{{ t('admin.companies.rejected') }}</option>
                </select>
              </div>
            </div>
            <div>
              <label for="company-notes" class="block text-sm font-medium text-gray-700">{{ t('admin.companies.kyb_notes') }}</label>
              <textarea id="company-notes" v-model="notesInput" name="notes" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div v-if="statusMessage" class="text-sm" :class="statusError ? 'text-red-600' : 'text-green-600'">
              {{ statusMessage }}
            </div>
            <div class="flex justify-end">
              <button type="submit" :disabled="updatingStatus" class="inline-flex justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700 disabled:opacity-50">
                {{ updatingStatus ? t('admin.companies.updating') : t('admin.companies.update') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const route = useRoute()
const { adminGetCompany, adminVerifyCompany } = useApi()
const { t } = useI18n()
const localePath = useLocalePath()
const { currencyOrDefault: cur, cell, enumLabel, formatNumber, formatDate } = useDisplay()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'

const company = ref<any>(null)
const pending = ref(true)
const error = ref('')
const statusInput = ref('pending')
const notesInput = ref('')
const updatingStatus = ref(false)
const statusMessage = ref('')
const statusError = ref(false)

const fetchCompany = async () => {
  pending.value = true
  error.value = ''
  try {
    company.value = await adminGetCompany(String(route.params.id))
    statusInput.value = company.value.kybStatus || 'pending'
    notesInput.value = company.value.kybNotes || ''
  } catch (err: any) {
    error.value = err?.data?.message || err.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const updateStatus = async () => {
  updatingStatus.value = true
  statusMessage.value = ''
  statusError.value = false

  try {
    await adminVerifyCompany(String(route.params.id), statusInput.value)
    statusMessage.value = t('admin.companies.status_updated')
    await fetchCompany()
  } catch (err: any) {
    statusError.value = true
    statusMessage.value = err?.data?.message || err.message || t('errors.api.status_failed')
  } finally {
    updatingStatus.value = false
  }
}

const statusBadgeClass = (status: string) => {
  if (status === 'verified') return 'bg-green-100 text-green-800'
  if (status === 'rejected') return 'bg-red-100 text-red-800'
  return 'bg-yellow-100 text-yellow-800'
}

const formatAddress = (address: any) => {
  if (!address) return '-'
  if (typeof address === 'string') return address
  const parts = [address.street, address.city, address.state, address.zipCode, address.country].filter(Boolean)
  return parts.join(', ')
}

onMounted(fetchCompany)
</script>
