<template>
  <div>
    <div class="mb-6">
      <NuxtLink to="/admin/oem-projects" class="flex items-center text-sm font-medium text-blue-600 hover:text-blue-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('admin.oemProjects.back') }}
      </NuxtLink>
    </div>

    <div v-if="pending" class="text-center py-10">
      <p class="text-gray-500">{{ t('admin.oemProjects.loading_details') }}</p>
    </div>

    <div v-else-if="error" class="bg-red-50 p-4 rounded-md">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <div v-else-if="project" class="space-y-6">
      <!-- Project Info Card -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 flex justify-between items-center bg-gray-50 border-b border-gray-200">
          <div>
            <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.oemProjects.project') }} #{{ project.projectNumber || project.id.substring(0, 8) }}</h3>
            <p class="mt-1 max-w-2xl text-sm text-gray-500">
              {{ t('admin.oemProjects.created_on') }} {{ project.createdAt ? new Date(project.createdAt).toLocaleString() : '-' }}
            </p>
          </div>
          <span :class="[statusBadgeClass(project.status), 'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5']">
            {{ project.status }}
          </span>
        </div>

        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.customer') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">
                <template v-if="project.user">
                  {{ project.user.firstName }} {{ project.user.lastName }}
                  <span class="text-gray-500">({{ project.user.email }})</span>
                </template>
                <template v-else>{{ project.userId || '-' }}</template>
              </dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.product_name') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.productName || '-' }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.target_market') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.targetMarket || '-' }}</dd>
            </div>
            <div v-if="project.assignedTo">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.assigned_to') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.assignedTo }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Step Tracker -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.oemProjects.step_tracker') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <div class="flex items-center justify-between">
            <template v-for="(step, idx) in steps" :key="step.key">
              <div class="flex flex-col items-center">
                <div :class="[getStepClass(step.key), 'w-8 h-8 rounded-full flex items-center justify-center text-white text-sm font-medium']">
                  {{ idx + 1 }}
                </div>
                <span class="mt-1 text-xs text-gray-500">{{ step.label }}</span>
              </div>
              <div v-if="idx < steps.length - 1" :class="['flex-1 h-1 mx-2', getStepLineClass(idx)]"></div>
            </template>
          </div>
        </div>
      </div>

      <!-- Requirements -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.oemProjects.requirements') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
            <div v-if="project.flavor">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.flavor') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.flavor }}</dd>
            </div>
            <div v-if="project.shape">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.shape') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.shape }}</dd>
            </div>
            <div v-if="project.packaging">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.packaging') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.packaging }}</dd>
            </div>
            <div v-if="project.targetMarket">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.target_market') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.targetMarket }}</dd>
            </div>
            <div v-if="project.certifications && project.certifications.length">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.certifications') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ Array.isArray(project.certifications) ? project.certifications.join(', ') : project.certifications }}</dd>
            </div>
            <div v-if="project.moq">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.moq') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.moq }}</dd>
            </div>
            <div v-if="project.requirements" class="sm:col-span-2">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.requirements_notes') }}</dt>
              <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ project.requirements }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Samples -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200 flex justify-between items-center">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.oemProjects.samples') }}</h3>
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oemProjects.sample_id') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oemProjects.sample_status') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oemProjects.requested_date') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('admin.oemProjects.shipped_date') }}</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-if="!project.samples || project.samples.length === 0">
                <td colspan="4" class="px-6 py-4 text-center text-sm text-gray-500">{{ t('admin.oemProjects.no_samples') }}</td>
              </tr>
              <tr v-else v-for="sample in project.samples" :key="sample.id">
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{ sample.id.substring(0, 8) }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">
                  <span :class="[sampleStatusClass(sample.status), 'inline-flex rounded-full px-2 text-xs font-semibold leading-5']">
                    {{ sample.status }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ sample.requestedAt ? new Date(sample.requestedAt).toLocaleDateString() : '-' }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ sample.shippedAt ? new Date(sample.shippedAt).toLocaleDateString() : '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Status Update -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.oemProjects.update_status') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <form @submit.prevent="updateStatus" class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.oemProjects.status') }}</label>
              <select v-model="statusInput" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option v-for="step in steps" :key="step.key" :value="step.key">{{ step.label }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.oemProjects.notes') }}</label>
              <textarea v-model="notesInput" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div v-if="statusMessage" class="text-sm" :class="statusError ? 'text-red-600' : 'text-green-600'">
              {{ statusMessage }}
            </div>
            <div class="flex justify-end">
              <button type="submit" :disabled="updatingStatus" class="inline-flex justify-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 disabled:opacity-50">
                {{ updatingStatus ? t('admin.oemProjects.updating') : t('admin.oemProjects.update') }}
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
const { token } = useAuth()
const { t } = useI18n()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'

const project = ref<any>(null)
const pending = ref(true)
const error = ref('')
const statusInput = ref('')
const notesInput = ref('')
const updatingStatus = ref(false)
const statusMessage = ref('')
const statusError = ref(false)

const steps = [
  { key: 'inquiry', label: 'Inquiry' },
  { key: 'sampling', label: 'Sampling' },
  { key: 'formulation', label: 'Formulation' },
  { key: 'quotation', label: 'Quotation' },
  { key: 'contract', label: 'Contract' },
  { key: 'production', label: 'Production' },
  { key: 'delivery', label: 'Delivery' },
  { key: 'completed', label: 'Completed' }
]

const fetchProject = async () => {
  pending.value = true
  error.value = ''
  try {
    project.value = await $fetch<any>(`${baseURL}/admin/oem-projects/${route.params.id}`, {
      headers: { Authorization: `Bearer ${token.value}` }
    })
    statusInput.value = project.value.status || 'inquiry'
    notesInput.value = project.value.adminNotes || ''
  } catch (err: any) {
    error.value = err?.data?.message || err.message || 'Failed to fetch project'
  } finally {
    pending.value = false
  }
}

const updateStatus = async () => {
  updatingStatus.value = true
  statusMessage.value = ''
  statusError.value = false

  try {
    await $fetch(`${baseURL}/admin/oem-projects/${route.params.id}/status`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${token.value}` },
      body: {
        status: statusInput.value,
        notes: notesInput.value
      }
    })
    project.value.status = statusInput.value
    statusMessage.value = t('admin.oemProjects.status_updated')
    setTimeout(() => { statusMessage.value = '' }, 3000)
  } catch (err: any) {
    statusError.value = true
    statusMessage.value = err?.data?.message || err.message || 'Failed to update status'
  } finally {
    updatingStatus.value = false
  }
}

const getStepIndex = (status: string) => {
  const idx = steps.findIndex(s => s.key === status)
  return idx >= 0 ? idx : 0
}

const getStepClass = (stepKey: string) => {
  const currentIdx = getStepIndex(project.value?.status || 'inquiry')
  const stepIdx = steps.findIndex(s => s.key === stepKey)
  if (stepIdx < currentIdx) return 'bg-green-500'
  if (stepIdx === currentIdx) return 'bg-blue-500'
  return 'bg-gray-300'
}

const getStepLineClass = (idx: number) => {
  const currentIdx = getStepIndex(project.value?.status || 'inquiry')
  if (idx < currentIdx) return 'bg-green-500'
  return 'bg-gray-300'
}

const statusBadgeClass = (status: string) => {
  const currentIdx = getStepIndex(status)
  const maxIdx = steps.length - 1
  if (currentIdx >= maxIdx) return 'bg-green-100 text-green-800'
  if (currentIdx >= 4) return 'bg-blue-100 text-blue-800'
  return 'bg-yellow-100 text-yellow-800'
}

const sampleStatusClass = (status: string) => {
  if (status === 'approved') return 'bg-green-100 text-green-800'
  if (status === 'rejected') return 'bg-red-100 text-red-800'
  if (status === 'shipped') return 'bg-blue-100 text-blue-800'
  return 'bg-gray-100 text-gray-800'
}

onMounted(fetchProject)
</script>
