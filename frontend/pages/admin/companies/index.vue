<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.companies.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.companies.description') }}</p>
      </div>
    </div>

    <div class="mt-6 flex flex-wrap items-center gap-4">
      <div class="flex-1 min-w-[200px]">
        <input v-model="searchInput" type="text" :placeholder="t('admin.companies.search_placeholder')" class="block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
      </div>
      <select v-model="statusFilter" class="rounded-md border border-gray-300 px-3 py-2 text-sm">
        <option value="">{{ t('admin.companies.all_statuses') }}</option>
        <option value="pending">{{ t('admin.companies.pending') }}</option>
        <option value="verified">{{ t('admin.companies.verified') }}</option>
        <option value="rejected">{{ t('admin.companies.rejected') }}</option>
      </select>
    </div>

    <div class="mt-8 overflow-hidden rounded-lg bg-white shadow ring-1 ring-black ring-opacity-5">
      <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
          <tr>
            <th class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{{ t('admin.companies.col_name') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.companies.col_tax_id') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.companies.col_status') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.companies.col_credit_limit') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.companies.col_payment_terms') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.companies.col_price_list') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.companies.col_verified_at') }}</th>
            <th class="relative py-3.5 pl-3 pr-4 sm:pr-6"><span class="sr-only">{{ t('admin.companies.actions') }}</span></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <tr v-if="pending">
            <td colspan="8" class="py-5 text-center text-sm text-gray-500">{{ t('admin.companies.loading') }}</td>
          </tr>
          <tr v-else-if="error">
            <td colspan="8" class="py-5 text-center text-sm text-red-600">{{ error }}</td>
          </tr>
          <tr v-else-if="companies.length === 0">
            <td colspan="8" class="py-5 text-center text-sm text-gray-500">{{ t('admin.companies.no_data') }}</td>
          </tr>
          <tr v-else v-for="company in companies" :key="company.id">
            <td class="py-4 pl-4 pr-3 text-sm sm:pl-6">
              <div class="font-medium text-gray-900">{{ company.name }}</div>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ company.taxId || '-' }}</td>
            <td class="px-3 py-4 text-sm">
              <span :class="[statusBadgeClass(company.status), 'inline-flex rounded-full px-2 text-xs font-semibold leading-5']">
                {{ company.status || 'pending' }}
              </span>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ company.creditLimit ? `${company.currency || 'USD'} ${company.creditLimit.toLocaleString()}` : '-' }}</td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ company.paymentTerms || '-' }}</td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ company.priceListName || '-' }}</td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ company.verifiedAt ? new Date(company.verifiedAt).toLocaleDateString() : '-' }}</td>
            <td class="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
              <NuxtLink :to="`/admin/companies/${company.id}`" class="text-blue-600 hover:text-blue-900 mr-3">{{ t('admin.companies.view') }}</NuxtLink>
              <button v-if="company.status === 'pending'" type="button" class="text-green-600 hover:text-green-900 mr-3" @click="verifyCompany(company.id, 'verified')">{{ t('admin.companies.verify') }}</button>
              <button v-if="company.status === 'pending'" type="button" class="text-red-600 hover:text-red-900" @click="verifyCompany(company.id, 'rejected')">{{ t('admin.companies.reject') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="pagination" class="mt-4 flex items-center justify-between rounded-lg border-t border-gray-200 bg-white px-4 py-3 shadow sm:px-6">
      <div class="text-sm text-gray-700">
        {{ t('admin.companies.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
      </div>
      <div class="flex items-center gap-2">
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page <= 1" @click="prevPage">
          {{ t('admin.companies.previous') }}
        </button>
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page >= pagination.totalPages" @click="nextPage">
          {{ t('admin.companies.next') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const { token } = useAuth()
const { t } = useI18n()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'

const companies = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20
const searchInput = ref('')
const statusFilter = ref('')

let searchTimeout: ReturnType<typeof setTimeout>

const fetchCompanies = async () => {
  pending.value = true
  error.value = ''
  try {
    const params: Record<string, any> = { page: page.value, limit: pageSize }
    if (searchInput.value.trim()) params.search = searchInput.value.trim()
    if (statusFilter.value) params.status = statusFilter.value

    const res = await $fetch<any>(`${baseURL}/admin/companies`, {
      headers: { Authorization: `Bearer ${token.value}` },
      params
    })
    companies.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.data?.message || err.message || 'Failed to fetch companies'
  } finally {
    pending.value = false
  }
}

const verifyCompany = async (id: string, status: string) => {
  try {
    await $fetch(`${baseURL}/admin/companies/${id}/verify`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${token.value}` },
      body: { status }
    })
    await fetchCompanies()
  } catch (err: any) {
    alert(err?.data?.message || err.message || 'Failed to verify company')
  }
}

const statusBadgeClass = (status: string) => {
  if (status === 'verified') return 'bg-green-100 text-green-800'
  if (status === 'rejected') return 'bg-red-100 text-red-800'
  return 'bg-yellow-100 text-yellow-800'
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

watch([page, searchInput, statusFilter], () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(fetchCompanies, 300)
})

onMounted(fetchCompanies)
</script>
