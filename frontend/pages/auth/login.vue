<template>
  <div class="login-page">
    <div class="login-page__container">
      <!-- Logo & Header -->
      <div class="login-page__header">
        <NuxtLink to="/" class="login-page__logo">
          <svg viewBox="0 0 180 40" fill="none" class="login-page__logo-svg">
            <circle cx="20" cy="20" r="16" fill="var(--color-highlight)" opacity="0.2"/>
            <path d="M14 20C14 16.6863 16.6863 14 20 14C23.3137 14 26 16.6863 26 20C26 23.3137 23.3137 26 20 26C16.6863 26 14 23.3137 14 20Z" stroke="var(--color-highlight)" stroke-width="2.5"/>
            <circle cx="17" cy="17" r="2" fill="var(--color-accent)"/>
            <circle cx="23" cy="23" r="2" fill="var(--color-primary)"/>
            <text x="45" y="27" font-family="Outfit, system-ui, sans-serif" font-size="18" font-weight="700" fill="var(--color-primary)">CandyPro</text>
            <text x="145" y="27" font-family="Outfit, system-ui, sans-serif" font-size="10" font-weight="600" fill="var(--color-accent)">OEM</text>
          </svg>
        </NuxtLink>
        <h1 class="login-page__title">{{ $t('auth.login_title') }}</h1>
        <p class="login-page__subtitle">
          {{ $t('auth.login_or') }}
          <NuxtLink to="/auth/register" class="login-page__link">
            {{ $t('auth.login_create_account') }}
          </NuxtLink>
        </p>
      </div>

      <!-- Login Form -->
      <div class="login-card">
        <form class="login-card__form" @submit.prevent="handleLogin">
          <!-- Email Field -->
          <div class="form-group">
            <label for="email" class="form-label">{{ $t('auth.email') }}</label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              autocomplete="email"
              required
              class="form-input"
              :class="{ 'form-input--error': errors.email }"
            />
            <span v-if="errors.email" class="form-error">{{ errors.email }}</span>
          </div>

          <!-- Password Field -->
          <div class="form-group">
            <label for="password" class="form-label">{{ $t('auth.password') }}</label>
            <input
              id="password"
              v-model="form.password"
              type="password"
              autocomplete="current-password"
              required
              class="form-input"
              :class="{ 'form-input--error': errors.password }"
            />
            <span v-if="errors.password" class="form-error">{{ errors.password }}</span>
          </div>

          <!-- Remember & Forgot -->
          <div class="login-card__options">
            <label class="login-card__remember">
              <input
                id="remember-me"
                v-model="form.remember"
                type="checkbox"
                class="login-card__checkbox"
              />
              <span class="login-card__checkbox-custom"></span>
              {{ $t('auth.remember_me') }}
            </label>
            <NuxtLink to="/auth/forgot-password" class="login-card__forgot">
              {{ $t('auth.forgot_password') }}
            </NuxtLink>
          </div>

          <!-- Submit Button -->
          <button type="submit" :disabled="loading" class="btn btn-highlight btn-lg login-card__submit">
            <span v-if="loading" class="login-card__spinner"></span>
            <span v-else>{{ $t('auth.sign_in') }}</span>
          </button>

          <!-- Error Message -->
          <Transition name="fade">
            <div v-if="error" class="login-card__error">
              <Icon name="lucide:alert-circle" size="18" />
              {{ error }}
            </div>
          </Transition>
        </form>
      </div>

      <!-- Footer -->
      <p class="login-page__footer">
        {{ new Date().getFullYear() }} CandyPro OEM. All rights reserved.
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
  // Reset errors
  errors.email = ''
  errors.password = ''
  error.value = ''

  // Basic validation
  if (!form.email) {
    errors.email = 'Email is required'
    return
  }
  if (!form.password) {
    errors.password = 'Password is required'
    return
  }

  loading.value = true

  try {
    const res = await login(form)
    const rawRedirect = typeof route.query.redirect === 'string' ? route.query.redirect : undefined
    // R4-15: Only allow relative paths to prevent open redirect
    const redirectUrl = rawRedirect && rawRedirect.startsWith('/') ? rawRedirect : null
    if (redirectUrl) {
      navigateTo(redirectUrl)
    } else if (res.user.role === 'admin' || res.user.role === 'superadmin') {
      navigateTo('/admin')
    } else {
      navigateTo('/customer/dashboard')
    }
  } catch (err) {
    error.value = err.message || 'Failed to login. Please check your credentials.'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: var(--spacing-xl);
}

.login-page__container {
  width: 100%;
  max-width: 420px;
  animation: fadeInUp 0.5s ease forwards;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Header */
.login-page__header {
  text-align: center;
  margin-bottom: var(--spacing-xl);
}

.login-page__logo {
  display: inline-block;
  margin-bottom: var(--spacing-lg);
}

.login-page__logo-svg {
  height: 48px;
  width: auto;
}

.login-page__title {
  font-family: var(--font-display);
  font-size: var(--text-3xl);
  font-weight: 700;
  color: var(--color-primary);
  margin-bottom: var(--spacing-sm);
}

.login-page__subtitle {
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.login-page__link {
  color: var(--color-highlight);
  font-weight: 500;
  transition: color var(--transition-fast);
}

.login-page__link:hover {
  color: var(--color-highlight-hover);
}

/* Login Card */
.login-card {
  background: white;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
  padding: var(--spacing-xl);
}

.login-card__form {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.login-card__options {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.login-card__remember {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: var(--text-sm);
  color: var(--color-text);
  cursor: pointer;
}

.login-card__checkbox {
  display: none;
}

.login-card__checkbox-custom {
  width: 18px;
  height: 18px;
  border: 2px solid var(--color-border);
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
  position: relative;
}

.login-card__checkbox:checked + .login-card__checkbox-custom {
  background: var(--color-highlight);
  border-color: var(--color-highlight);
}

.login-card__checkbox:checked + .login-card__checkbox-custom::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 5px;
  width: 4px;
  height: 8px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}

.login-card__forgot {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  transition: color var(--transition-fast);
}

.login-card__forgot:hover {
  color: var(--color-highlight);
}

.login-card__submit {
  width: 100%;
  margin-top: var(--spacing-sm);
}

.login-card__spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.login-card__error {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: var(--radius-md);
  color: var(--color-error);
  font-size: var(--text-sm);
}

/* Form Elements */
.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.form-label {
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text);
}

.form-input {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-base);
  color: var(--color-text);
  background: var(--color-bg);
  border: 1.5px solid var(--color-border);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(255, 107, 74, 0.1);
}

.form-input--error {
  border-color: var(--color-error);
}

.form-input--error:focus {
  box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
}

.form-error {
  font-size: var(--text-xs);
  color: var(--color-error);
}

/* Footer */
.login-page__footer {
  text-align: center;
  margin-top: var(--spacing-xl);
  font-size: var(--text-xs);
  color: var(--color-text-lighter);
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
