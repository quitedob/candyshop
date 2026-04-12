<template>
  <div class="min-h-screen flex flex-col justify-center py-12 sm:px-6 lg:px-8">
    <div class="sm:mx-auto sm:w-full sm:max-w-md">
      <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
        Reset your password
      </h2>
      <p class="mt-2 text-center text-sm text-gray-600">
        Remembered your password?
        <NuxtLink to="/auth/login" class="font-medium text-blue-600 hover:text-blue-500">
          Sign in here
        </NuxtLink>
      </p>
    </div>

    <div class="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
      <div class="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
        <form class="space-y-6" @submit.prevent="handleReset">
          <div>
            <label for="email" class="block text-sm font-medium text-gray-700">
              Email address
            </label>
            <div class="mt-1">
              <input id="email" v-model="email" name="email" type="email" autocomplete="email" required class="appearance-none block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm" />
            </div>
            <p class="mt-2 text-sm text-gray-500">We'll send you a link to reset your password.</p>
          </div>

          <div>
            <button type="submit" :disabled="loading" class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50">
              <span v-if="loading">Sending...</span>
              <span v-else>Send Reset Link</span>
            </button>
          </div>

          <div v-if="success" class="mt-4 p-4 text-sm text-green-700 bg-green-100 rounded-md">
            If an account exists with this email, a reset link has been sent.
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

definePageMeta({
  layout: 'auth',
  middleware: ['auth']
})

const email = ref('')
const loading = ref(false)
const success = ref(false)

const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'

const handleReset = async () => {
  loading.value = true
  success.value = false

  try {
    await $fetch<any>(`${baseURL}/auth/forgot-password`, {
      method: 'POST',
      body: { email: email.value }
    })
    success.value = true
    email.value = ''
  } catch {
    success.value = true
  } finally {
    loading.value = false
  }
}
</script>
