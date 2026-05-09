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
        <h1 class="auth-page__title">{{ $t('auth.login_title') }}</h1>
        <p class="auth-page__subtitle">
          {{ $t('auth.login_or') }}
          <NuxtLink :to="localePath('/auth/register')" class="auth-page__link">
            {{ $t('auth.login_create_account') }}
          </NuxtLink>
        </p>
      </div>

      <div class="auth-card">
        <form class="auth-form" @submit.prevent="handleLogin">
          <div class="auth-field">
            <label for="email" class="auth-label">{{ $t('auth.email') }}</label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              autocomplete="email"
              required
              class="auth-input"
              :class="{ 'auth-input--error': errors.email }"
            />
            <p v-if="errors.email" class="auth-field-error">{{ errors.email }}</p>
          </div>

          <div class="auth-field">
            <label for="password" class="auth-label">{{ $t('auth.password') }}</label>
            <input
              id="password"
              v-model="form.password"
              type="password"
              autocomplete="current-password"
              required
              class="auth-input"
              :class="{ 'auth-input--error': errors.password }"
            />
            <p v-if="errors.password" class="auth-field-error">{{ errors.password }}</p>
          </div>

          <div class="auth-actions-row">
            <label class="auth-checkbox-label">
              <input id="remember-me" v-model="form.remember" type="checkbox" class="auth-checkbox" />
              <span class="auth-checkbox-ui" aria-hidden="true" />
              {{ $t('auth.remember_me') }}
            </label>
            <NuxtLink :to="localePath('/auth/forgot-password')" class="auth-link-muted">
              {{ $t('auth.forgot_password') }}
            </NuxtLink>
          </div>

          <button type="submit" :disabled="loading" class="auth-btn">
            <span v-if="loading" class="auth-spinner" aria-hidden="true" />
            <span v-if="loading" class="sr-only">{{ $t('auth.signing_in') }}</span>
            <span v-else>{{ $t('auth.sign_in') }}</span>
          </button>

          <Transition name="auth-fade">
            <div v-if="error" class="auth-alert auth-alert--error" role="alert">
              <Icon name="lucide:alert-circle" size="18" aria-hidden="true" />
              <span>{{ error }}</span>
            </div>
          </Transition>
        </form>
      </div>

      <p class="auth-page__footer">
        © {{ new Date().getFullYear() }} {{ $t('auth.footer_rights') }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'

definePageMeta({
  layout: 'auth',
  middleware: ['auth']
})

const { login } = useAuth()
const route = useRoute()
const localePath = useLocalePath()
const { t } = useI18n()

const form = reactive({
  email: '',
  password: '',
  remember: false
})

const loading = ref(false)
const error = ref('')
const errors = reactive({
  email: '',
  password: ''
})

const handleLogin = async () => {
  errors.email = ''
  errors.password = ''
  error.value = ''

  if (!form.email) {
    errors.email = t('auth.validation_email_required')
    return
  }
  if (!form.password) {
    errors.password = t('auth.validation_password_required')
    return
  }

  loading.value = true

  try {
    const res = await login(form)
    // Redirect unverified users to check-email page
    if (res.user.emailVerified === false) {
      navigateTo(localePath('/auth/check-email') + '?email=' + encodeURIComponent(form.email))
      return
    }
    const rawRedirect = typeof route.query.redirect === 'string' ? route.query.redirect : undefined
    const redirectUrl = rawRedirect && rawRedirect.startsWith('/') ? rawRedirect : null
    if (redirectUrl) {
      navigateTo(redirectUrl)
    } else if (res.user.role === 'admin' || res.user.role === 'superadmin') {
      navigateTo(localePath('/admin'))
    } else {
      navigateTo(localePath('/customer/dashboard'))
    }
  } catch (err) {
    error.value = err.message || t('auth.errors.login_network')
  } finally {
    loading.value = false
  }
}
</script>
