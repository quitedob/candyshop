<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('customer.oemProjects.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('customer.oemProjects.subtitle') }}</p>
      </div>
      <div class="mt-4 sm:mt-0">
        <NuxtLink :to="localePath('/customer/oem-projects/new')" class="inline-flex items-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700">
          {{ t('customer.oemProjects.start_new_project') }}
        </NuxtLink>
      </div>
    </div>

    <div class="mt-8 overflow-hidden rounded-lg bg-white shadow ring-1 ring-black ring-opacity-5">
      <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
          <tr>
            <th class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{{ t('customer.oemProjects.col_project') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('customer.oemProjects.col_product') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('customer.oemProjects.col_status') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('customer.oemProjects.col_created') }}</th>
            <th class="relative py-3.5 pl-3 pr-4 sm:pr-6"><span class="sr-only">{{ t('customer.oemProjects.actions') }}</span></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <tr v-if="pending">
            <td colspan="5" class="py-5 text-center text-sm text-gray-500">{{ t('customer.oemProjects.loading') }}</td>
          </tr>
          <tr v-else-if="error">
            <td colspan="5" class="py-5 text-center text-sm text-red-600">{{ error }}</td>
          </tr>
          <tr v-else-if="projects.length === 0">
            <td colspan="5" class="py-10 text-center">
              <div class="text-gray-500">{{ t('customer.oemProjects.no_projects') }}</div>
              <NuxtLink :to="localePath('/customer/oem-projects/new')" class="mt-4 inline-flex items-center text-orange-600 hover:text-orange-900">
                {{ t('customer.oemProjects.start_first_project') }}
              </NuxtLink>
            </td>
          </tr>
          <tr v-else v-for="project in projects" :key="project.id">
            <td class="py-4 pl-4 pr-3 text-sm sm:pl-6">
              <div class="font-medium text-gray-900">{{ project.projectNumber || project.id.substring(0, 8) }}</div>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ project.productName || '-' }}</td>
            <td class="px-3 py-4 text-sm">
              <div class="flex items-center gap-1">
                <template v-for="(step, idx) in steps.slice(0, getStepIndex(project.status) + 1)" :key="idx">
                  <span :class="[stepClass(project.status, step.key), 'w-2 h-2 rounded-full']"></span>
                </template>
                <span class="ml-1 text-xs text-gray-500">{{ enumLabel('oem_status', project.status) }}</span>
              </div>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ formatDate(project.createdAt) }}</td>
            <td class="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
              <NuxtLink :to="localePath(`/customer/oem-projects/${project.id}`)" class="text-orange-600 hover:text-orange-900">{{ t('customer.oemProjects.view') }}</NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="pagination" class="mt-4 flex items-center justify-between rounded-lg border-t border-gray-200 bg-white px-4 py-3 shadow sm:px-6">
      <div class="text-sm text-gray-700">
        {{ t('customer.oemProjects.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
      </div>
      <div class="flex items-center gap-2">
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page <= 1" @click="prevPage">
          {{ t('customer.oemProjects.previous') }}
        </button>
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page >= pagination.totalPages" @click="nextPage">
          {{ t('customer.oemProjects.next') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

definePageMeta({ layout: 'customer', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const { enumLabel, formatDate } = useDisplay()
const localePath = useLocalePath()

const projects = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20

const steps = [
  { key: 'inquiry', label: t('customer.oemProjects.inquiry') }, { key: 'sampling', label: t('customer.oemProjects.sampling') },
  { key: 'formulation', label: t('customer.oemProjects.formulation') }, { key: 'quotation', label: t('customer.oemProjects.quotation') },
  { key: 'contract', label: t('customer.oemProjects.contract') }, { key: 'production', label: t('customer.oemProjects.production') },
  { key: 'delivery', label: t('customer.oemProjects.delivery') }, { key: 'completed', label: t('customer.oemProjects.completed') }
]

const fetchProjects = async () => {
  pending.value = true; error.value = ''
  try {
    const res = await api.get<any>(`/user/oem-projects?page=${page.value}&limit=${pageSize}`)
    projects.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally { pending.value = false }
}

const getStepIndex = (status: string) => { const idx = steps.findIndex(s => s.key === status); return idx >= 0 ? idx : 0 }
const stepClass = (status: string, stepKey: string) => {
  const currentIdx = getStepIndex(status); const stepIdx = steps.findIndex(s => s.key === stepKey)
  if (stepIdx < currentIdx) return 'bg-green-500'; if (stepIdx === currentIdx) return 'bg-orange-500'; return 'bg-gray-300'
}
const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) page.value += 1 }
const prevPage = () => { if (page.value > 1) page.value -= 1 }

watch(page, fetchProjects)
onMounted(fetchProjects)
</script>
