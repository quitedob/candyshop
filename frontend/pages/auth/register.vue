<template>
  <div class="auth-page">
    <div class="auth-page__container auth-page__container--wide">
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
        <h1 class="auth-page__title">{{ $t('auth.register_title') }}</h1>
        <p class="auth-page__subtitle">
          {{ $t('auth.register_already') }}
          <NuxtLink :to="localePath('/auth/login')" class="auth-page__link">
            {{ $t('auth.register_sign_in') }}
          </NuxtLink>
        </p>
      </div>

      <div class="auth-card">
        <form class="auth-form" @submit.prevent="handleRegister">
          <div class="auth-row-2">
            <div class="auth-field">
              <label for="first-name" class="auth-label">{{ $t('auth.first_name') }}</label>
              <input id="first-name" v-model="form.firstName" type="text" required autocomplete="given-name" class="auth-input" />
            </div>
            <div class="auth-field">
              <label for="last-name" class="auth-label">{{ $t('auth.last_name') }}</label>
              <input id="last-name" v-model="form.lastName" type="text" required autocomplete="family-name" class="auth-input" />
            </div>
          </div>

          <div class="auth-field">
            <label for="company-name" class="auth-label">{{ $t('auth.company_name') }}</label>
            <input id="company-name" v-model="form.companyName" type="text" required autocomplete="organization" class="auth-input" />
          </div>

          <div class="auth-field">
            <label for="email" class="auth-label">{{ $t('auth.business_email') }}</label>
            <input id="email" v-model="form.email" type="email" required autocomplete="email" class="auth-input" />
          </div>

          <div class="auth-field">
            <label for="password" class="auth-label">{{ $t('auth.password') }}</label>
            <input id="password" v-model="form.password" type="password" required minlength="8" autocomplete="new-password" class="auth-input" />
            <p class="auth-help">{{ $t('auth.password_min_length') }}</p>
          </div>

          <button type="submit" :disabled="loading" class="auth-btn">
            <span v-if="loading" class="auth-spinner" aria-hidden="true" />
            <span v-if="loading" class="sr-only">{{ $t('auth.registering') }}</span>
            <span v-else>{{ $t('auth.create_account') }}</span>
          </button>

          <div v-if="error" class="auth-alert auth-alert--error" role="alert">
            {{ error }}
          </div>
          <div v-if="success" class="auth-alert auth-alert--success" role="status">
            {{ $t('auth.register_success') }}
          </div>
        </form>
      </div>

      <p class="auth-page__footer">
        © {{ new Date().getFullYear() }} {{ $t('auth.footer_rights') }}
      </p>
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
const localePath = useLocalePath()
const { t } = useI18n()

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
    await navigateTo(localePath('/auth/check-email') + '?email=' + encodeURIComponent(form.email))
  } catch (err: any) {
    error.value = err.message || t('auth.errors.register_failed')
  } finally {
    loading.value = false
  }
}
</script>
