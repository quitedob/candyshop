<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.oemProjects.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.oemProjects.description') }}</p>
      </div>
    </div>

    <div class="mt-6 flex flex-wrap items-center gap-4">
      <div class="flex-1 min-w-[200px]">
        <input id="oem-search" v-model="searchInput" name="search" type="text" autocomplete="off" :placeholder="t('admin.oemProjects.search_placeholder')" class="block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
      </div>
      <select id="oem-statusFilter" name="statusFilter" v-model="statusFilter" class="rounded-md border border-gray-300 px-3 py-2 text-sm">
        <option value="">{{ t('admin.oemProjects.all_statuses') }}</option>
        <option value="inquiry">{{ t('admin.oemProjects.inquiry') }}</option>
        <option value="sampling">{{ t('admin.oemProjects.sampling') }}</option>
        <option value="formulation">{{ t('admin.oemProjects.formulation') }}</option>
        <option value="quotation">{{ t('admin.oemProjects.quotation') }}</option>
        <option value="contract">{{ t('admin.oemProjects.contract') }}</option>
        <option value="production">{{ t('admin.oemProjects.production') }}</option>
        <option value="delivery">{{ t('admin.oemProjects.delivery') }}</option>
        <option value="completed">{{ t('admin.oemProjects.completed') }}</option>
      </select>
    </div>

    <div class="mt-8 overflow-hidden rounded-lg bg-white shadow ring-1 ring-black ring-opacity-5">
      <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
          <tr>
            <th class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{{ t('admin.oemProjects.col_project') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.oemProjects.col_customer') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.oemProjects.col_product') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.oemProjects.col_status') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.oemProjects.col_assigned') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.oemProjects.col_created') }}</th>
            <th class="relative py-3.5 pl-3 pr-4 sm:pr-6"><span class="sr-only">{{ t('admin.oemProjects.actions') }}</span></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <tr v-if="pending">
            <td colspan="7" class="py-5 text-center text-sm text-gray-500">{{ t('admin.oemProjects.loading') }}</td>
          </tr>
          <tr v-else-if="error">
            <td colspan="7" class="py-5 text-center text-sm text-red-600">{{ error }}</td>
          </tr>
          <tr v-else-if="projects.length === 0">
            <td colspan="7" class="py-5 text-center text-sm text-gray-500">{{ t('admin.oemProjects.no_data') }}</td>
          </tr>
          <tr v-else v-for="project in projects" :key="project.id" class="hover:bg-gray-50 cursor-pointer" role="button" tabindex="0" :aria-label="t('admin.oemProjects.view_detail')" @click="navigateToDetail(project.id)" @keydown.enter.prevent="navigateToDetail(project.id)" @keydown.space.prevent="navigateToDetail(project.id)">
            <td class="py-4 pl-4 pr-3 text-sm sm:pl-6">
              <div class="font-medium text-gray-900">{{ project.projectNumber || project.id.substring(0, 8) }}</div>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">
              <template v-if="project.user">
                <div class="font-medium text-gray-900">{{ project.user.firstName }} {{ project.user.lastName }}</div>
                <div class="text-xs text-gray-500">{{ project.user.email }}</div>
              </template>
              <template v-else>{{ project.userId || '-' }}</template>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ project.productName || '-' }}</td>
            <td class="px-3 py-4 text-sm">
              <div class="flex items-center gap-1">
                <template v-for="(step, idx) in steps.slice(0, getStepIndex(project.status) + 1)" :key="idx">
                  <span :class="[stepClass(project.status, step.key), 'w-2 h-2 rounded-full']"></span>
                </template>
                <span class="ml-1 text-xs text-gray-500">{{ project.status }}</span>
              </div>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ project.assignedTo || '-' }}</td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ formatDate(project.createdAt) }}</td>
            <td class="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
              <NuxtLink :to="localePath(`/admin/oem-projects/${project.id}`)" class="text-orange-600 hover:text-orange-900" @click.stop>{{ t('admin.oemProjects.view') }}</NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="pagination" class="mt-4 flex items-center justify-between rounded-lg border-t border-gray-200 bg-white px-4 py-3 shadow sm:px-6">
      <div class="text-sm text-gray-700">
        {{ t('admin.oemProjects.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
      </div>
      <div class="flex items-center gap-2">
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page <= 1" @click="prevPage">
          {{ t('admin.oemProjects.previous') }}
        </button>
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page >= pagination.totalPages" @click="nextPage">
          {{ t('admin.oemProjects.next') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const { formatDate } = useDisplay()
const localePath = useLocalePath()
const router = useRouter()

const projects = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20
const searchInput = ref('')
const statusFilter = ref('')

const steps = [
  { key: 'inquiry', label: t('admin.oemProjects.inquiry') }, { key: 'sampling', label: t('admin.oemProjects.sampling') },
  { key: 'formulation', label: t('admin.oemProjects.formulation') }, { key: 'quotation', label: t('admin.oemProjects.quotation') },
  { key: 'contract', label: t('admin.oemProjects.contract') }, { key: 'production', label: t('admin.oemProjects.production') },
  { key: 'delivery', label: t('admin.oemProjects.delivery') }, { key: 'completed', label: t('admin.oemProjects.completed') }
]

let searchTimeout: ReturnType<typeof setTimeout>

const fetchProjects = async () => {
  pending.value = true; error.value = ''
  try {
    const params: Record<string, any> = { page: page.value, limit: pageSize }
    if (searchInput.value.trim()) params.search = searchInput.value.trim()
    if (statusFilter.value) params.status = statusFilter.value
    const res = await api.get<any>('/admin/oem-projects', params)
    projects.value = res.data || []
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally { pending.value = false }
}

const navigateToDetail = (id: string) => router.push(localePath(`/admin/oem-projects/${id}`))

const getStepIndex = (status: string) => { const idx = steps.findIndex(s => s.key === status); return idx >= 0 ? idx : 0 }
const stepClass = (status: string, stepKey: string) => {
  const currentIdx = getStepIndex(status); const stepIdx = steps.findIndex(s => s.key === stepKey)
  if (stepIdx < currentIdx) return 'bg-green-500'; if (stepIdx === currentIdx) return 'bg-orange-500'; return 'bg-gray-300'
}
const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) page.value += 1 }
const prevPage = () => { if (page.value > 1) page.value -= 1 }

watch([page, searchInput, statusFilter], () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(fetchProjects, 300)
})

onMounted(fetchProjects)
</script>
