<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
      <div class="bg-white shadow sm:rounded-lg overflow-hidden">
        <div class="px-4 py-5 sm:p-6">
          <div class="text-center">
            <Icon v-if="verifying" name="heroicons:clock" class="mx-auto h-12 w-12 text-gray-400" />
            <Icon v-else-if="success" name="heroicons:check-circle" class="mx-auto h-12 w-12 text-green-500" />
            <Icon v-else name="heroicons:x-circle" class="mx-auto h-12 w-12 text-red-500" />

            <h3 class="mt-4 text-lg font-medium text-gray-900">
              {{ verifying ? t('auth.verifying_email') : success ? t('auth.email_verified') : t('auth.email_verification_failed') }}
            </h3>

            <p class="mt-2 text-sm text-gray-600">
              {{ success ? t('auth.email_verified_message') : errorMessage }}
            </p>

            <div class="mt-6">
              <NuxtLink v-if="success || (!verifying && !success)" to="/auth/login" class="inline-flex items-center justify-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700">
                {{ t('auth.go_to_login') }}
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

definePageMeta({
  layout: 'auth',
  middleware: ['guest']
})

const route = useRoute()
const { t } = useI18n()
const { verifyEmail } = useAuth()

const verifying = ref(true)
const success = ref(false)
const errorMessage = ref('')

onMounted(async () => {
  const token = route.query.token as string

  if (!token) {
    verifying.value = false
    success.value = false
    errorMessage.value = t('auth.missing_token')
    return
  }

  try {
    await verifyEmail(token)
    verifying.value = false
    success.value = true
  } catch (err: any) {
    verifying.value = false
    success.value = false
    errorMessage.value = err?.message || t('auth.verification_failed')
  }
})
</script>
