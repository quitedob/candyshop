<template>
  <Teleport to="body">
    <Transition name="cookie-consent">
      <div v-if="visible" role="dialog" aria-labelledby="cookie-consent-title" aria-modal="true" class="cookie-consent">
        <div class="cookie-consent__card">
          <div class="cookie-consent__body">
            <h3 id="cookie-consent-title" class="cookie-consent__title">{{ t('cookie_consent.title') }}</h3>
            <p class="cookie-consent__text">{{ t('cookie_consent.message') }}</p>
            <div class="cookie-consent__actions">
              <button
                class="cookie-consent__btn cookie-consent__btn--essential"
                @click="acceptEssential"
              >
                {{ t('cookie_consent.essential_only') }}
              </button>
              <button
                class="cookie-consent__btn cookie-consent__btn--accept"
                @click="acceptAll"
              >
                {{ t('cookie_consent.accept_all') }}
              </button>
            </div>
            <p class="cookie-consent__link">
              <NuxtLink :to="localePath('/legal/cookies')">
                {{ t('cookie_consent.learn_more') }}
              </NuxtLink>
            </p>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n, useLocalePath } from '#i18n'

const { t } = useI18n()
const localePath = useLocalePath()

const COOKIE_CONSENT_KEY = 'candypro_cookie_consent'

const visible = ref(false)

function getStoredConsent(): string | null {
  if (import.meta.server) return null
  return localStorage.getItem(COOKIE_CONSENT_KEY)
}

function setConsent(level: string) {
  if (import.meta.server) return
  localStorage.setItem(COOKIE_CONSENT_KEY, JSON.stringify({
    level,
    timestamp: Date.now(),
  }))
  visible.value = false
}

function acceptAll() {
  setConsent('all')
}

function acceptEssential() {
  setConsent('essential')
}

onMounted(() => {
  const stored = getStoredConsent()
  if (!stored) {
    visible.value = true
  }
})
</script>

<style scoped>
.cookie-consent {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 9999;
  padding: var(--spacing-md);
  background: linear-gradient(to top, rgba(0, 0, 0, 0.6), transparent);
}

.cookie-consent__card {
  max-width: 720px;
  margin: 0 auto var(--spacing-lg);
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.18);
  overflow: hidden;
}

.cookie-consent__body {
  padding: var(--spacing-xl);
}

.cookie-consent__title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: var(--spacing-sm);
}

.cookie-consent__text {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  line-height: 1.6;
  margin-bottom: var(--spacing-lg);
}

.cookie-consent__actions {
  display: flex;
  gap: var(--spacing-md);
  flex-wrap: wrap;
}

.cookie-consent__btn {
  padding: var(--spacing-sm) var(--spacing-xl);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  font-weight: 600;
  cursor: pointer;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  border: none;
}

.cookie-consent__btn--essential {
  background-color: var(--color-bg-alt);
  color: var(--color-text);
  border: 1px solid var(--color-border);
}

.cookie-consent__btn--essential:hover {
  background-color: var(--color-border);
}

.cookie-consent__btn--accept {
  background-color: var(--color-highlight);
  color: var(--color-text-on-primary);
}

.cookie-consent__btn--accept:hover {
  background-color: var(--color-highlight-hover);
  transform: translateY(-1px);
}

.cookie-consent__link {
  margin-top: var(--spacing-md);
  font-size: var(--text-xs);
}

.cookie-consent__link a {
  color: var(--color-highlight);
  text-decoration: underline;
}

.cookie-consent__link a:hover {
  color: var(--color-highlight-hover);
}

/* Transition */
.cookie-consent-enter-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}
.cookie-consent-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.cookie-consent-enter-from {
  opacity: 0;
  transform: translateY(20px);
}
.cookie-consent-leave-to {
  opacity: 0;
  transform: translateY(20px);
}

@media (prefers-reduced-motion: reduce) {
  .cookie-consent-enter-active,
  .cookie-consent-leave-active {
    transition: opacity 0.1s ease;
  }
}

@media (max-width: 640px) {
  .cookie-consent__actions {
    flex-direction: column;
  }
  .cookie-consent__btn {
    width: 100%;
    text-align: center;
  }
}
</style>
