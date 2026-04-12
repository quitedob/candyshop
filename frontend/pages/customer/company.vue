<template>
  <div class="max-w-2xl space-y-6">
    <div v-if="pending" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-10 w-10 border-4 border-orange-200 border-t-orange-600"></div>
    </div>
    <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-xl p-6 text-red-700">{{ error }}</div>
    <div v-else-if="!company" class="text-center py-16 text-gray-500">
      <Icon name="heroicons:building-office" class="h-12 w-12 mx-auto mb-3 text-gray-300" />
      <p>{{ $t('customer.company.no_company') }}</p>
    </div>
    <div v-else class="bg-white shadow rounded-lg p-6 space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold text-gray-900">{{ company.name }}</h2>
        <button v-if="!editing" @click="startEdit" class="text-sm text-orange-600 hover:text-orange-800 font-medium">
          {{ $t('customer.company.edit') }}
        </button>
      </div>

      <div v-if="!editing" class="grid grid-cols-2 gap-4 text-sm">
        <div>
          <p class="text-gray-500">{{ $t('customer.company.phone') }}</p>
          <p class="font-medium">{{ company.phone || '-' }}</p>
        </div>
        <div>
          <p class="text-gray-500">{{ $t('customer.company.website') }}</p>
          <p class="font-medium">{{ company.website || '-' }}</p>
        </div>
        <div class="col-span-2">
          <p class="text-gray-500">{{ $t('customer.company.address') }}</p>
          <p class="font-medium">
            {{ [company.address?.street, company.address?.city, company.address?.state, company.address?.country].filter(Boolean).join(', ') || '-' }}
          </p>
        </div>
      </div>

      <form v-else @submit.prevent="saveCompany" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ $t('customer.company.phone') }}</label>
          <input v-model="form.phone" type="text" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ $t('customer.company.website') }}</label>
          <input v-model="form.website" type="text" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">{{ $t('customer.company.street') }}</label>
          <input v-model="form.address.street" type="text" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ $t('customer.company.city') }}</label>
            <input v-model="form.address.city" type="text" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ $t('customer.company.country') }}</label>
            <input v-model="form.address.country" type="text" class="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" />
          </div>
        </div>
        <div class="flex gap-3 pt-2">
          <button type="submit" :disabled="saving" class="px-4 py-2 bg-orange-500 text-white text-sm font-medium rounded-lg hover:bg-orange-600 disabled:opacity-50">
            {{ saving ? $t('customer.company.saving') : $t('customer.company.save') }}
          </button>
          <button type="button" @click="editing = false" class="px-4 py-2 bg-gray-100 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-200">
            {{ $t('customer.company.cancel') }}
          </button>
        </div>
        <p v-if="saveError" class="text-sm text-red-600">{{ saveError }}</p>
        <p v-if="saveSuccess" class="text-sm text-green-600">{{ $t('customer.company.saved') }}</p>
      </form>
      <!-- KYB Document Upload -->
      <div v-if="!editing" class="border-t border-gray-200 pt-4">
        <h3 class="text-sm font-semibold text-gray-900 mb-3">{{ $t('customer.company.kyb_documents') }}</h3>
        <div v-if="company.businessLicense" class="flex items-center gap-3 mb-3 p-3 bg-green-50 rounded-lg border border-green-200">
          <Icon name="heroicons:document-check" class="h-5 w-5 text-green-600" />
          <div>
            <p class="text-sm font-medium text-green-800">{{ $t('customer.company.business_license_uploaded') }}</p>
            <a :href="company.businessLicense" target="_blank" class="text-xs text-green-600 hover:underline">{{ $t('customer.company.view_document') }}</a>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <input ref="kybFileInput" type="file" accept=".pdf,image/jpeg,image/png" class="hidden" @change="handleKYBFile" />
          <button @click="kybFileInput?.click()" :disabled="uploadingKYB" class="px-4 py-2 border border-gray-300 text-sm font-medium rounded-lg hover:bg-gray-50 disabled:opacity-50 flex items-center gap-2">
            <Icon name="heroicons:arrow-up-tray" class="h-4 w-4" />
            {{ uploadingKYB ? $t('customer.company.uploading') : $t('customer.company.upload_business_license') }}
          </button>
          <span v-if="kybFileName" class="text-sm text-gray-600">{{ kybFileName }}</span>
        </div>
        <p v-if="kybError" class="mt-2 text-sm text-red-600">{{ kybError }}</p>
        <p v-if="kybSuccess" class="mt-2 text-sm text-green-600">{{ $t('customer.company.kyb_uploaded') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'customer', middleware: ['auth'] })

const api = useApi()
const company = ref<any>(null)
const pending = ref(true)
const error = ref('')
const editing = ref(false)
const saving = ref(false)
const saveError = ref('')
const saveSuccess = ref(false)

const form = reactive({
  phone: '',
  website: '',
  address: { street: '', city: '', state: '', country: '' }
})

const load = async () => {
  try {
    company.value = await api.get('/user/company')
  } catch (e: any) {
    error.value = e?.message || 'Failed to load company'
  } finally {
    pending.value = false
  }
}

const startEdit = () => {
  form.phone = company.value?.phone || ''
  form.website = company.value?.website || ''
  form.address = { ...company.value?.address } || { street: '', city: '', state: '', country: '' }
  editing.value = true
  saveError.value = ''
  saveSuccess.value = false
}

const saveCompany = async () => {
  saving.value = true
  saveError.value = ''
  try {
    company.value = await api.put('/user/company', form)
    editing.value = false
    saveSuccess.value = true
    setTimeout(() => { saveSuccess.value = false }, 3000)
  } catch (e: any) {
    saveError.value = e?.message || 'Failed to save'
  } finally {
    saving.value = false
  }
}

// KYB document upload
const kybFileInput = ref<HTMLInputElement | null>(null)
const kybFileName = ref('')
const uploadingKYB = ref(false)
const kybError = ref('')
const kybSuccess = ref(false)

const handleKYBFile = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  kybFileName.value = file.name
  uploadingKYB.value = true
  kybError.value = ''
  kybSuccess.value = false
  try {
    const formData = new FormData()
    formData.append('file', file)
    await api.post('/user/company/kyb-document', formData)
    kybSuccess.value = true
    await load()
    setTimeout(() => { kybSuccess.value = false }, 4000)
  } catch (e: any) {
    kybError.value = e?.message || 'Failed to upload document'
  } finally {
    uploadingKYB.value = false
    if (kybFileInput.value) kybFileInput.value.value = ''
  }
}

onMounted(load)
</script>
