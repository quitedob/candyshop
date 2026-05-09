<template>
  <div>
    <div class="mb-6">
      <NuxtLink :to="localePath('/customer/oem-projects')" class="flex items-center text-sm font-medium text-orange-600 hover:text-orange-500">
        <Icon name="heroicons:arrow-left" class="mr-1 h-4 w-4" />
        {{ t('customer.oemProjects.back') }}
      </NuxtLink>
    </div>

    <div v-if="pending" class="text-center py-10">
      <p class="text-gray-500">{{ t('customer.oemProjects.loading_details') }}</p>
    </div>

    <div v-else-if="error" class="bg-red-50 p-4 rounded-md">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <div v-else-if="project" class="space-y-6">
      <!-- Project Info Card -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 flex justify-between items-center bg-gray-50 border-b border-gray-200">
          <div>
            <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.oemProjects.project') }} #{{ project.projectNumber || project.id.substring(0, 8) }}</h3>
            <p class="mt-1 max-w-2xl text-sm text-gray-500">
              {{ t('customer.oemProjects.created_on') }} {{ formatDate(project.createdAt, { dateStyle: 'medium', timeStyle: 'short' }) }}
            </p>
          </div>
          <span :class="[statusBadgeClass(project.status), 'inline-flex rounded-full px-3 py-1 text-sm font-semibold leading-5']">
            {{ project.status }}
          </span>
        </div>

        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
            <div>
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.oemProjects.product_name') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.productName || '-' }}</dd>
            </div>
            <div v-if="project.targetMarket">
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.oemProjects.target_market') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.targetMarket }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Step Tracker -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.oemProjects.progress') }}</h3>
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
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.oemProjects.your_requirements') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <dl class="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
            <div v-if="project.flavor">
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.oemProjects.flavor') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.flavor }}</dd>
            </div>
            <div v-if="project.shape">
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.oemProjects.shape') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.shape }}</dd>
            </div>
            <div v-if="project.packaging">
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.oemProjects.packaging') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.packaging }}</dd>
            </div>
            <div v-if="project.targetMarket">
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.oemProjects.target_market') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.targetMarket }}</dd>
            </div>
            <div v-if="project.certifications && project.certifications.length">
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.oemProjects.certifications') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ Array.isArray(project.certifications) ? project.certifications.join(', ') : project.certifications }}</dd>
            </div>
            <div v-if="project.moq">
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.oemProjects.moq') }}</dt>
              <dd class="mt-1 text-sm text-gray-900">{{ project.moq }}</dd>
            </div>
            <div v-if="project.requirements" class="sm:col-span-2">
              <dt class="text-sm font-medium text-gray-500">{{ t('customer.oemProjects.requirements') }}</dt>
              <dd class="mt-1 text-sm text-gray-900 whitespace-pre-line">{{ project.requirements }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Samples -->
      <div class="bg-white shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-gray-50 border-b border-gray-200 flex justify-between items-center">
          <h3 class="text-lg leading-6 font-medium text-gray-900">{{ t('customer.oemProjects.samples') }}</h3>
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200">
            <thead class="bg-gray-50">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('customer.oemProjects.sample_id') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('customer.oemProjects.status') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('customer.oemProjects.requested_date') }}</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">{{ t('customer.oemProjects.shipped_date') }}</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
              <tr v-if="!project.samples || project.samples.length === 0">
                <td colspan="4" class="px-6 py-4 text-center text-sm text-gray-500">{{ t('customer.oemProjects.no_samples') }}</td>
              </tr>
              <tr v-else v-for="sample in project.samples" :key="sample.id">
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{ sample.id.substring(0, 8) }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">
                  <span :class="[sampleStatusClass(sample.status), 'inline-flex rounded-full px-2 text-xs font-semibold leading-5']">
                    {{ sample.status }}
                  </span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ formatDate(sample.requestedAt) }}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ formatDate(sample.shippedAt) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Notes/Feedback -->
      <div v-if="project.adminNotes" class="bg-orange-50 shadow overflow-hidden sm:rounded-lg">
        <div class="px-4 py-5 sm:px-6 bg-orange-100 border-b border-blue-200">
          <h3 class="text-lg leading-6 font-medium text-blue-900">{{ t('customer.oemProjects.notes_from_team') }}</h3>
        </div>
        <div class="px-4 py-5 sm:p-6">
          <p class="text-sm text-blue-900 whitespace-pre-line">{{ project.adminNotes }}</p>
        </div>
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
const { formatDate } = useDisplay()
const localePath = useLocalePath()

const project = ref<any>(null)
const pending = ref(true)
const error = ref('')

const steps = [
  { key: 'inquiry', label: t('customer.oemProjects.inquiry') }, { key: 'sampling', label: t('customer.oemProjects.sampling') },
  { key: 'formulation', label: t('customer.oemProjects.formulation') }, { key: 'quotation', label: t('customer.oemProjects.quotation') },
  { key: 'contract', label: t('customer.oemProjects.contract') }, { key: 'production', label: t('customer.oemProjects.production') },
  { key: 'delivery', label: t('customer.oemProjects.delivery') }, { key: 'completed', label: t('customer.oemProjects.completed') }
]

const fetchProject = async () => {
  pending.value = true; error.value = ''
  try { project.value = await api.get<any>(`/user/oem-projects/${route.params.id}`) }
  catch (err: any) { error.value = err?.message || t('errors.api.load_failed') }
  finally { pending.value = false }
}

const getStepIndex = (status: string) => { const idx = steps.findIndex(s => s.key === status); return idx >= 0 ? idx : 0 }
const getStepClass = (stepKey: string) => {
  const currentIdx = getStepIndex(project.value?.status || 'inquiry')
  const stepIdx = steps.findIndex(s => s.key === stepKey)
  if (stepIdx < currentIdx) return 'bg-green-500'; if (stepIdx === currentIdx) return 'bg-orange-500'; return 'bg-gray-300'
}
const getStepLineClass = (idx: number) => { const currentIdx = getStepIndex(project.value?.status || 'inquiry'); return idx < currentIdx ? 'bg-green-500' : 'bg-gray-300' }
const statusBadgeClass = (status: string) => {
  const currentIdx = getStepIndex(status); const maxIdx = steps.length - 1
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
