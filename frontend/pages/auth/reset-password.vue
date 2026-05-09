<template>
  <div class="auth-page">
    <div class="auth-page__container">
      <div class="auth-page__header">
        <NuxtLink :to="localePath('/')" class="auth-page__logo">
          <svg viewBox="0 0 180 40" fill="none" class="auth-page__logo-svg">
            <circle cx="20" cy="20" r="16" fill="var(--color-highlight)" opacity="0.2"/>
            <path d="M14 20C14 16.6863 16.6863 14 20 14C23.3137 14 26 16.6863 26 20C26 23.3137 23.3137 26 20 26C16.6863 26 14 23.3137 14 20Z" stroke="var(--color-highlight)" stroke-width="2.5"/>
            <circle cx="17" cy="17" r="2" fill="var(--color-accent)"/>
            <circle cx="23" cy="23" r="2" fill="var(--color-primary)"/>
            <text x="45" y="27" font-family="Georgia, serif" font-size="18" font-weight="500" fill="var(--color-primary)">CandyPro</text>
            <text x="145" y="27" font-family="system-ui, sans-serif" font-size="10" font-weight="600" fill="var(--color-accent)">OEM</text>
          </svg>
        </NuxtLink>
        <h1 class="auth-page__title">{{ t('auth_extra.set_new_password') }}</h1>
      </div>

      <div class="auth-card">
        <form v-if="!success" class="auth-form" @submit.prevent="handleReset">
          <div class="auth-field">
            <label for="password" class="auth-label">{{ t('auth_extra.new_password') }}</label>
            <input
              id="password"
              v-model="form.newPassword"
              type="password"
              required
              minlength="8"
              autocomplete="new-password"
              class="auth-input"
            />
          </div>

          <div class="auth-field">
            <label for="confirm" class="auth-label">{{ t('auth_extra.confirm_password') }}</label>
            <input
              id="confirm"
              v-model="form.confirmPassword"
              type="password"
              required
              minlength="8"
              autocomplete="new-password"
              class="auth-input"
            />
          </div>

          <button type="submit" :disabled="loading" class="auth-btn">
            <span v-if="loading" class="auth-spinner" aria-hidden="true" />
            <span v-if="loading" class="sr-only">{{ t('auth_extra.resetting') }}</span>
            <span v-else>{{ t('auth_extra.reset_password') }}</span>
          </button>

          <div v-if="errorMsg" class="auth-alert auth-alert--error" role="alert">
            {{ errorMsg }}
          </div>
        </form>

        <div v-else class="auth-form">
          <div class="auth-alert auth-alert--success" role="status">
            {{ t('auth_extra.reset_success') }}
          </div>
          <NuxtLink :to="localePath('/auth/login')" class="auth-btn auth-btn--secondary auth-btn--auto">
            {{ t('auth_extra.go_to_login') }}
          </NuxtLink>
        </div>
      </div>

      <p class="auth-page__footer">
        © {{ new Date().getFullYear() }} {{ t('auth.footer_rights') }}
      </p>
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
const localePath = useLocalePath()
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
