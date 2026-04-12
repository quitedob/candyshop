<template>
  <div class="min-h-screen flex flex-col justify-center py-12 sm:px-6 lg:px-8">
    <div class="sm:mx-auto sm:w-full sm:max-w-md">
      <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
        {{ t('auth_extra.set_new_password') }}
      </h2>
    </div>

    <div class="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
      <div class="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
        <form v-if="!success" class="space-y-6" @submit.prevent="handleReset">
          <div>
            <label for="password" class="block text-sm font-medium text-gray-700">
              {{ t('auth_extra.new_password') }}
            </label>
            <div class="mt-1">
              <input id="password" v-model="form.newPassword" type="password" required minlength="8" class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
            </div>
          </div>

          <div>
            <label for="confirm" class="block text-sm font-medium text-gray-700">
              {{ t('auth_extra.confirm_password') }}
            </label>
            <div class="mt-1">
              <input id="confirm" v-model="form.confirmPassword" type="password" required minlength="8" class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
            </div>
          </div>

          <div>
            <button type="submit" :disabled="loading" class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50">
              <span v-if="loading">{{ t('auth_extra.resetting') }}</span>
              <span v-else>{{ t('auth_extra.reset_password') }}</span>
            </button>
          </div>

          <div v-if="errorMsg" class="mt-4 p-4 text-sm text-red-700 bg-red-100 rounded-md">
            {{ errorMsg }}
          </div>
        </form>

        <div v-else class="text-center space-y-4">
          <div class="p-4 text-sm text-green-700 bg-green-100 rounded-md">
            {{ t('auth_extra.reset_success') }}
          </div>
          <NuxtLink to="/auth/login" class="inline-flex justify-center w-full py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700">
            {{ t('auth_extra.go_to_login') }}
          </NuxtLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

definePageMeta({
  layout: 'auth',
  middleware: ['auth']
})

const { t } = useI18n()
const route = useRoute()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'

const loading = ref(false)
const success = ref(false)
const errorMsg = ref('')

const form = reactive({
  token: '',
  newPassword: '',
  confirmPassword: ''
})

onMounted(() => {
  const tokenParam = route.query.token
  if (tokenParam && typeof tokenParam === 'string') {
    form.token = tokenParam
  } else {
    errorMsg.value = t('auth_extra.invalid_token')
  }
})

const handleReset = async () => {
  if (form.newPassword !== form.confirmPassword) {
    errorMsg.value = t('auth_extra.passwords_mismatch')
    return
  }
  if (!form.token) {
    errorMsg.value = t('auth_extra.missing_token')
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    await $fetch<any>(`${baseURL}/auth/reset-password`, {
      method: 'POST',
      body: {
        token: form.token,
        newPassword: form.newPassword
      }
    })
    success.value = true
  } catch (err: any) {
    errorMsg.value = err.data?.message || t('auth_extra.reset_failed')
  } finally {
    loading.value = false
  }
}
</script>
