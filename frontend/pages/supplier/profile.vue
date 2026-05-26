<template>
  <div class="space-y-4">
    <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">{{ t('supplier.nav_profile') }}</h1>
    <form class="max-w-lg space-y-4 rounded-lg border border-gray-200 bg-white p-6 dark:border-gray-700 dark:bg-gray-900" @submit.prevent="save">
      <div>
        <label for="supplier-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('supplier.company_name') }}</label>
        <input id="supplier-name" v-model="form.name" name="name" autocomplete="organization" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm dark:bg-gray-800 dark:border-gray-600" />
      </div>
      <div>
        <label for="supplier-email" class="block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('supplier.contact_email') }}</label>
        <input id="supplier-email" v-model="form.email" name="email" type="email" autocomplete="email" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm dark:bg-gray-800 dark:border-gray-600" />
      </div>
      <p v-if="message" class="text-sm text-green-600">{{ message }}</p>
      <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
      <button type="submit" class="rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700">{{ t('common.save') }}</button>
    </form>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'supplier', middleware: ['auth'] })
const { t } = useI18n()
const api = useApi()
const form = reactive({ name: '', email: '' })
const message = ref('')
const error = ref('')

onMounted(async () => {
  try {
    const profile = await api.get<any>('/supplier/profile')
    form.name = profile?.name || ''
    form.email = profile?.email || ''
  } catch { /* silent */ }
})

async function save() {
  message.value = ''
  error.value = ''
  try {
    await api.put('/supplier/profile', { name: form.name, email: form.email })
    message.value = t('supplier.profile_saved')
  } catch (e: any) {
    error.value = e?.message || t('errors.api.request_failed')
  }
}
</script>
