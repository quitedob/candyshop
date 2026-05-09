<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.certifications.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.certifications.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0">
        <button type="button" class="inline-flex items-center justify-center rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700" @click="openCreateModal">
          {{ t('admin.certifications.add_certification') }}
        </button>
      </div>
    </div>

    <div class="mt-8 overflow-hidden rounded-lg bg-white shadow ring-1 ring-black ring-opacity-5">
      <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
          <tr>
            <th class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900">{{ t('admin.certifications.col_certification') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.certifications.col_abbreviation') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.certifications.col_issuer') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.certifications.col_valid_until') }}</th>
            <th class="relative py-3.5 pl-3 pr-4 sm:pr-6"><span class="sr-only">Actions</span></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <tr v-if="pending">
            <td colspan="5" class="py-5 text-center text-sm text-gray-500">{{ t('admin.certifications.loading') }}</td>
          </tr>
          <tr v-else-if="error">
            <td colspan="5" class="py-5 text-center text-sm text-red-600">{{ error }}</td>
          </tr>
          <tr v-else-if="certifications.length === 0">
            <td colspan="5" class="py-5 text-center text-sm text-gray-500">{{ t('admin.certifications.no_data') }}</td>
          </tr>
          <tr v-else v-for="cert in certifications" :key="cert.id">
            <td class="py-4 pl-4 pr-3 text-sm">
              <div class="flex items-center">
                <div class="h-10 w-10 flex-shrink-0">
                  <img v-if="cert.badgeUrl" class="h-10 w-10 rounded-full object-cover" :src="cert.badgeUrl" alt="" />
                  <div v-else class="flex h-10 w-10 items-center justify-center rounded-full bg-gray-200 font-bold text-gray-500">
                    {{ cert.abbreviation }}
                  </div>
                </div>
                <div class="ml-4">
                  <div class="font-medium text-gray-900">{{ cert.name }}</div>
                </div>
              </div>
            </td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ cert.abbreviation }}</td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ cert.issuer }}</td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ cert.validUntil }}</td>
            <td class="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
              <button type="button" class="text-orange-600 hover:text-orange-900" @click="openEditModal(cert)">{{ t('admin.certifications.edit') }}</button>
              <button type="button" class="ml-4 text-red-600 hover:text-red-900" @click="deleteCertification(cert.id)">{{ t('admin.certifications.delete') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 z-10 overflow-y-auto" role="dialog" aria-modal="true">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <button type="button" class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity w-full border-0 cursor-pointer" @click="closeModal" :aria-label="t('common.close')"></button>
        <span class="hidden sm:inline-block sm:h-screen sm:align-middle" aria-hidden="true">&#8203;</span>
        <div class="inline-block w-full transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left align-bottom shadow-xl sm:my-8 sm:max-w-3xl sm:p-6 sm:align-middle">
          <h3 class="text-lg font-medium leading-6 text-gray-900">{{ editingId ? t('admin.certifications.edit_certification') : t('admin.certifications.create_certification') }}</h3>

          <form class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2" @submit.prevent="saveCertification">
            <div>
              <label for="cert-name" class="block text-sm font-medium text-gray-700">{{ t('admin.certifications.name') }}</label>
              <input id="cert-name" v-model="form.name" name="name" autocomplete="off" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="cert-abbreviation" class="block text-sm font-medium text-gray-700">{{ t('admin.certifications.abbreviation') }}</label>
              <input id="cert-abbreviation" v-model="form.abbreviation" name="abbreviation" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div class="sm:col-span-2">
              <label for="cert-description" class="block text-sm font-medium text-gray-700">{{ t('admin.certifications.description_label') }}</label>
              <textarea id="cert-description" v-model="form.description" name="description" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>
            <div>
              <label for="cert-issuer" class="block text-sm font-medium text-gray-700">{{ t('admin.certifications.issuer') }}</label>
              <input id="cert-issuer" v-model="form.issuer" name="issuer" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div>
              <label for="cert-validUntil" class="block text-sm font-medium text-gray-700">{{ t('admin.certifications.valid_until') }}</label>
              <input id="cert-validUntil" v-model="form.validUntil" name="validUntil" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div class="sm:col-span-2">
              <label for="cert-certificateUrl" class="block text-sm font-medium text-gray-700">{{ t('admin.certifications.certificate_url') }}</label>
              <input id="cert-certificateUrl" v-model="form.certificateUrl" name="certificateUrl" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>
            <div class="sm:col-span-2">
              <label for="cert-badgeUrl" class="block text-sm font-medium text-gray-700">{{ t('admin.certifications.badge_url') }}</label>
              <input id="cert-badgeUrl" v-model="form.badgeUrl" name="badgeUrl" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>

            <div v-if="formError" class="sm:col-span-2 text-sm text-red-600">{{ formError }}</div>
            <div class="sm:col-span-2 mt-2 flex justify-end gap-3">
              <button type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700" @click="closeModal">
                {{ t('admin.certifications.cancel') }}
              </button>
              <button type="submit" :disabled="saving" class="rounded-md border border-transparent bg-orange-600 px-4 py-2 text-sm text-white disabled:opacity-50">
                {{ saving ? t('admin.certifications.saving') : (editingId ? t('admin.certifications.update') : t('admin.certifications.create')) }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <p v-if="actionMessage" class="mt-4 text-sm" :class="actionError ? 'text-red-600' : 'text-green-600'">
      {{ actionMessage }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const { token } = useAuth()
const { t } = useI18n()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'

const certifications = ref<any[]>([])
const pending = ref(true)
const error = ref('')
const saving = ref(false)
const showModal = ref(false)
const editingId = ref('')
const formError = ref('')
const actionMessage = ref('')
const actionError = ref(false)

const form = reactive({
  name: '',
  abbreviation: '',
  description: '',
  issuer: '',
  validUntil: '',
  certificateUrl: '',
  badgeUrl: ''
})

const resetForm = () => {
  form.name = ''
  form.abbreviation = ''
  form.description = ''
  form.issuer = ''
  form.validUntil = ''
  form.certificateUrl = ''
  form.badgeUrl = ''
}

const fillFormFromCertification = (cert: any) => {
  form.name = cert.name || ''
  form.abbreviation = cert.abbreviation || ''
  form.description = cert.description || ''
  form.issuer = cert.issuer || ''
  form.validUntil = cert.validUntil || ''
  form.certificateUrl = cert.certificateUrl || ''
  form.badgeUrl = cert.badgeUrl || ''
}

const fetchCertifications = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await $fetch<any>(`${baseURL}/admin/certifications`, {
      headers: { Authorization: `Bearer ${token.value}` }
    })
    certifications.value = res.data || res || []
  } catch (err: any) {
    error.value = err?.data?.message || err.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const openCreateModal = () => {
  editingId.value = ''
  resetForm()
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false
  showModal.value = true
}

const openEditModal = (cert: any) => {
  editingId.value = cert.id
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false
  fillFormFromCertification(cert)
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  saving.value = false
  formError.value = ''
}

const saveCertification = async () => {
  if (!form.name.trim()) {
    formError.value = t('admin.certifications.name_required')
    return
  }

  saving.value = true
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false

  try {
    if (editingId.value) {
      await $fetch(`${baseURL}/admin/certifications/${editingId.value}`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${token.value}` },
        body: form
      })
      actionMessage.value = t('admin.certifications.updated_success')
    } else {
      await $fetch(`${baseURL}/admin/certifications`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token.value}` },
        body: form
      })
      actionMessage.value = t('admin.certifications.created_success')
    }
    closeModal()
    await fetchCertifications()
  } catch (err: any) {
    formError.value = err?.data?.message || t('errors.api.save_failed')
  } finally {
    saving.value = false
  }
}

const deleteCertification = async (id: string) => {
  if (!confirm(t('admin.certifications.confirm_delete'))) return

  actionMessage.value = ''
  actionError.value = false
  try {
    await $fetch(`${baseURL}/admin/certifications/${id}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token.value}` }
    })
    actionMessage.value = t('admin.certifications.deleted_success')
    await fetchCertifications()
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.data?.message || t('errors.api.delete_failed')
  }
}

onMounted(fetchCertifications)
</script>

<style scoped>
.admin-certifications-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: var(--spacing-md);
}
</style>
