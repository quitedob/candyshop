<template>
  <div>
    <div class="mb-6">
      <NuxtLink :to="localePath('/admin/oem-projects')" class="flex items-center text-sm font-medium text-orange-600 hover:text-orange-500">
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
            <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('admin.oemProjects.project') }} #{{ project.projectNumber || String(project.id).substring(0, 8) }}</h3>
            <p class="mt-1 max-w-2xl text-sm text-gray-500">
              {{ t('admin.oemProjects.created_on') }} {{ formatDate(project.createdAt, { dateStyle: 'medium', timeStyle: 'short' }) }}
            </p>
          </div>
          <span :class="[statusBadgeClass(project.status), 'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5']">
            {{ enumLabel('oem_status', project.status) }}
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
              <dd class="mt-1 text-sm text-gray-900">{{ tField(project, 'productName') || '-' }}</dd>
            </div>
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.target_market') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.requirements?.targetMarket || '-' }}</dd>
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
            <div v-if="project.requirements?.flavor">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.flavor') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.requirements.flavor }}</dd>
            </div>
            <div v-if="project.requirements?.shape">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.shape') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.requirements.shape }}</dd>
            </div>
            <div v-if="project.requirements?.packaging">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.packaging') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.requirements.packaging }}</dd>
            </div>
            <div v-if="project.requirements?.targetMarket">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.target_market') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.requirements.targetMarket }}</dd>
            </div>
            <div v-if="project.requirements?.certifications?.length">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.certifications') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.requirements.certifications.join(', ') }}</dd>
            </div>
            <div v-if="project.requirements?.moq">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.moq') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.requirements.moq }}</dd>
            </div>
            <div v-if="project.notes" class="sm:col-span-2">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.requirements_notes') }}</dt>
              <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ project.notes }}</dd>
            </div>
            <div v-if="project.adminNotes" class="sm:col-span-2">
              <dt class="text-sm font-medium text-gray-500">{{ t('admin.oemProjects.notes') }}</dt>
              <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ project.adminNotes }}</dd>
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
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{ String(sample.id).substring(0, 8) }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">
                  <span :class="[sampleStatusClass(sample.status), 'inline-flex rounded-full px-2 text-xs font-semibold leading-5']">
                    {{ enumLabel('sample_status', sample.status) }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ formatDate(sample.sentAt) }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ formatDate(sample.receivedAt) }}</td>
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
              <label for="oem-status" class="block text-sm font-medium text-gray-700">{{ t('admin.oemProjects.status') }}</label>
              <select id="oem-status" name="status" v-model="statusInput" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
                <option v-for="step in steps" :key="step.key" :value="step.key">{{ step.label }}</option>
                <option value="cancelled">{{ t('admin.oemProjects.cancelled') }}</option>
              </select>
            </div>
            <div>
              <label for="oem-notes" class="block text-sm font-medium text-gray-700">{{ t('admin.oemProjects.notes') }}</label>
              <textarea id="oem-notes" v-model="notesInput" name="notes" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div v-if="statusMessage" class="text-sm" :class="statusError ? 'text-red-600' : 'text-green-600'">
              {{ statusMessage }}
            </div>
            <div class="flex justify-end">
              <button type="submit" :disabled="updatingStatus" class="inline-flex justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700 disabled:opacity-50">
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
import { useTranslation } from '~/composables/useTranslation'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const route = useRoute()
const { adminGetOemProject, adminUpdateOemStatus, adminUpdateOemProject } = useApi()
const { t } = useI18n()
const { tField } = useTranslation()
const { enumLabel, formatDate } = useDisplay()
const localePath = useLocalePath()

const project = ref<any>(null)
const pending = ref(true)
const error = ref('')
const statusInput = ref('')
const notesInput = ref('')
const updatingStatus = ref(false)
const statusMessage = ref('')
const statusError = ref(false)

const steps = [
  { key: 'inquiry', label: t('admin.oemProjects.inquiry') },
  { key: 'sampling', label: t('admin.oemProjects.sampling') },
  { key: 'formulation', label: t('admin.oemProjects.formulation') },
  { key: 'quotation', label: t('admin.oemProjects.quotation') },
  { key: 'contract', label: t('admin.oemProjects.contract') },
  { key: 'production', label: t('admin.oemProjects.production') },
  { key: 'delivery', label: t('admin.oemProjects.delivery') },
  { key: 'completed', label: t('admin.oemProjects.completed') }
]

const fetchProject = async () => {
  pending.value = true
  error.value = ''
  try {
    project.value = await adminGetOemProject(String(route.params.id))
    statusInput.value = project.value.status || 'inquiry'
    notesInput.value = project.value.adminNotes || ''
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
    await adminUpdateOemStatus(String(route.params.id), {
      status: statusInput.value,
    })
    if (notesInput.value !== (project.value.adminNotes || '')) {
      await adminUpdateOemProject(String(route.params.id), { adminNotes: notesInput.value })
      project.value.adminNotes = notesInput.value
    }
    project.value.status = statusInput.value
    statusMessage.value = t('admin.oemProjects.status_updated')
    setTimeout(() => { statusMessage.value = '' }, 3000)
  } catch (err: any) {
    statusError.value = true
    statusMessage.value = err?.data?.message || err.message || t('errors.api.status_failed')
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
  if (stepIdx === currentIdx) return 'bg-orange-500'
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
  if (currentIdx >= 4) return 'bg-orange-100 text-orange-800'
  return 'bg-yellow-100 text-yellow-800'
}

const sampleStatusClass = (status: string) => {
  if (status === 'approved') return 'bg-green-100 text-green-800'
  if (status === 'rejected') return 'bg-red-100 text-red-800'
  if (status === 'shipped') return 'bg-orange-100 text-orange-800'
  return 'bg-gray-100 text-gray-800'
}

onMounted(fetchProject)
</script>
