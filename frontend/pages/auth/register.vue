<template>
  <div class="min-h-screen flex flex-col justify-center py-12 sm:px-6 lg:px-8">
    <div class="sm:mx-auto sm:w-full sm:max-w-md">
      <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
        {{ $t('auth.register_title') }}
      </h2>
      <p class="mt-2 text-center text-sm text-gray-600">
        {{ $t('auth.register_already') }}
        <NuxtLink to="/auth/login" class="font-medium text-blue-600 hover:text-blue-500">
          {{ $t('auth.register_sign_in') }}
        </NuxtLink>
      </p>
    </div>

    <div class="mt-8 sm:mx-auto sm:w-full sm:max-w-lg">
      <div class="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
        <form class="space-y-6" @submit.prevent="handleRegister">
          <div class="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-2">
            <div>
              <label for="first-name" class="block text-sm font-medium text-gray-700">{{ $t('auth.first_name') }}</label>
              <div class="mt-1">
                <input id="first-name" v-model="form.firstName" type="text" required class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
              </div>
            </div>

            <div>
              <label for="last-name" class="block text-sm font-medium text-gray-700">{{ $t('auth.last_name') }}</label>
              <div class="mt-1">
                <input id="last-name" v-model="form.lastName" type="text" required class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
              </div>
            </div>
          </div>

          <div>
            <label for="company-name" class="block text-sm font-medium text-gray-700">{{ $t('auth.company_name') }}</label>
            <div class="mt-1">
              <input id="company-name" v-model="form.companyName" type="text" required class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
            </div>
          </div>

          <div>
            <label for="email" class="block text-sm font-medium text-gray-700">{{ $t('auth.business_email') }}</label>
            <div class="mt-1">
              <input id="email" v-model="form.email" type="email" required class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
            </div>
          </div>

          <div>
            <label for="password" class="block text-sm font-medium text-gray-700">{{ $t('auth.password') }}</label>
            <div class="mt-1">
              <input id="password" v-model="form.password" type="password" required minlength="8" class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
            </div>
            <p class="mt-2 text-xs text-gray-500">{{ $t('auth.password_min_length') }}</p>
          </div>

          <div>
            <button type="submit" :disabled="loading" class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50">
              <span v-if="loading">{{ $t('auth.registering') }}</span>
              <span v-else>{{ $t('auth.create_account') }}</span>
            </button>
          </div>

          <div v-if="error" class="mt-4 text-sm text-red-600 text-center">
            {{ error }}
          </div>
          <div v-if="success" class="mt-4 text-sm text-green-600 text-center">
            {{ $t('auth.register_success') }}
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'

definePageMeta({
  layout: 'auth',
  middleware: ['guest']
})

const { register } = useAuth()

const form = reactive({
  firstName: '',
  lastName: '',
  companyName: '',
  email: '',
  password: ''
})

const loading = ref(false)
const error = ref('')
const success = ref(false)

const handleRegister = async () => {
  loading.value = true
  error.value = ''
  success.value = false

  try {
    await register(form)
    success.value = true
    // Do not auto-redirect — user must verify email first
  } catch (err: any) {
    error.value = err.message || 'Failed to register account'
  } finally {
    loading.value = false
  }
}
</script>
