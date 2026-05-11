<template>
  <div class="auth-page">
    <div class="auth-page__container">
      <div class="auth-card">
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

        <Icon
          v-if="verifying"
          name="heroicons:clock"
          class="auth-verify-icon"
          aria-hidden="true"
        />
        <Icon
          v-else-if="success"
          name="heroicons:check-circle"
          class="auth-verify-icon auth-verify-icon--ok"
          aria-hidden="true"
        />
        <Icon
          v-else
          name="heroicons:x-circle"
          class="auth-verify-icon auth-verify-icon--fail"
          aria-hidden="true"
        />

        <h2 class="auth-verify-title">
          {{ verifying ? t('auth.verifying_email') : success ? t('auth.email_verified') : t('auth.email_verification_failed') }}
        </h2>

        <p class="auth-verify-text">
          {{
            verifying
              ? t('auth.verify_please_wait')
              : success
                ? t('auth.email_verified_message')
                : errorMessage
          }}
        </p>

        <div class="auth-verify-actions">
          <NuxtLink
            v-if="success || (!verifying && !success)"
            :to="localePath('/auth/login')"
            class="auth-btn auth-btn--auto"
          >
            {{ t('auth.go_to_login') }}
          </NuxtLink>
        </div>
      </div>

      <p class="auth-page__footer">
        © {{ currentYear }} {{ t('auth.footer_rights') }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

definePageMeta({
  layout: 'auth',
  middleware: ['guest']
})

const route = useRoute()
const { t } = useI18n()
const localePath = useLocalePath()
const { verifyEmail } = useAuth()

const currentYear = computed(() => new Date().getFullYear())
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
