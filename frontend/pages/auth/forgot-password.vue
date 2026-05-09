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
        <h1 class="auth-page__title">{{ t('auth.forgot_title') }}</h1>
        <p class="auth-page__subtitle">
          {{ t('auth.forgot_subtitle') }}
          <NuxtLink :to="localePath('/auth/login')" class="auth-page__link">
            {{ t('auth.forgot_sign_in') }}
          </NuxtLink>
        </p>
      </div>

      <div class="auth-card">
        <form class="auth-form" @submit.prevent="handleReset">
          <div class="auth-field">
            <label for="email" class="auth-label">{{ t('auth.email') }}</label>
            <input
              id="email"
              v-model="email"
              name="email"
              type="email"
              autocomplete="email"
              required
              class="auth-input"
            />
            <p class="auth-help">{{ t('auth.forgot_email_hint') }}</p>
          </div>

          <button type="submit" :disabled="loading" class="auth-btn">
            <span v-if="loading" class="auth-spinner" aria-hidden="true" />
            <span v-if="loading" class="sr-only">{{ t('auth.forgot_sending') }}</span>
            <span v-else>{{ t('auth.forgot_send') }}</span>
          </button>

          <div v-if="success" class="auth-alert auth-alert--success" role="status">
            {{ t('auth.forgot_success') }}
          </div>
        </form>
      </div>

      <p class="auth-page__footer">
        © {{ new Date().getFullYear() }} {{ t('auth.footer_rights') }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

definePageMeta({
  layout: 'auth',
  middleware: ['auth']
})

const { t } = useI18n()
const localePath = useLocalePath()

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
    email.value = ''
  } finally {
    loading.value = false
  }
}
</script>
