<template>
  <div class="auth-page">
    <div class="auth-page__container">
      <div class="auth-card" style="text-align: center; padding: 3rem 2rem;">
        <div class="auth-page__header auth-page__header--tight">
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
        </div>

        <Icon name="lucide:mail-check" size="48" style="color: var(--color-highlight); margin-bottom: 1.5rem;" aria-hidden="true" />

        <h2 style="font-size: 1.5rem; margin-bottom: 0.75rem;">{{ t('auth.check_email_title') }}</h2>

        <p style="color: var(--color-text-light); margin-bottom: 1.5rem; line-height: 1.6;">
          {{ t('auth.check_email_message', { email: email || '—' }) }}
        </p>

        <p style="font-size: 0.875rem; color: var(--color-text-light); margin-bottom: 1.5rem;">
          {{ t('auth.check_email_spam') }}
        </p>

        <button
          class="auth-btn"
          style="margin-bottom: 1rem;"
          :disabled="resending || cooldown > 0"
          @click="handleResend"
        >
          <span v-if="resending" class="auth-spinner" aria-hidden="true" />
          {{ resending ? t('auth.resending') : cooldown > 0 ? `${t('auth.resend_verification')} (${cooldown}s)` : t('auth.resend_verification') }}
        </button>

        <Transition name="auth-fade">
          <div v-if="resendSuccess" class="auth-alert auth-alert--success" role="status" style="margin-bottom: 1rem;">
            {{ t('auth.resend_success') }}
          </div>
        </Transition>

        <Transition name="auth-fade">
          <div v-if="resendError" class="auth-alert auth-alert--error" role="alert" style="margin-bottom: 1rem;">
            {{ resendError }}
          </div>
        </Transition>

        <NuxtLink :to="localePath('/auth/login')" class="auth-link-muted" style="display: inline-block; margin-top: 0.5rem;">
          {{ t('auth.go_to_login') }}
        </NuxtLink>
      </div>

      <p class="auth-page__footer">
        © {{ currentYear }} {{ t('auth.footer_rights') }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'

definePageMeta({
  layout: 'auth',
  middleware: ['guest']
})

const { t } = useI18n()
const localePath = useLocalePath()
const route = useRoute()
const { resendVerificationEmail } = useAuth()

const email = computed(() => (route.query.email as string) || '')
const resending = ref(false)
const resendSuccess = ref(false)
const resendError = ref('')
const currentYear = computed(() => new Date().getFullYear())
const cooldown = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

const handleResend = async () => {
  resending.value = true
  resendSuccess.value = false
  resendError.value = ''
  try {
    await resendVerificationEmail()
    resendSuccess.value = true
    cooldown.value = 60
    timer = setInterval(() => {
      cooldown.value--
      if (cooldown.value <= 0 && timer) {
        clearInterval(timer)
        timer = null
      }
    }, 1000)
  } catch (err: any) {
    resendError.value = err.message || t('auth.errors.resend_failed')
  } finally {
    resending.value = false
  }
}

onUnmounted(() => { if (timer) clearInterval(timer) })
</script>
