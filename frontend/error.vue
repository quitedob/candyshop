<template>
  <div class="error-page">
    <div class="error-content">
      <h1 class="error-code">{{ error?.statusCode || 404 }}</h1>
      <h2 class="error-title">{{ errorTitle }}</h2>
      <p class="error-message">{{ errorMessage }}</p>
      <div class="error-actions">
        <NuxtLink to="/" class="btn-home">
          {{ t('errors.goHome') }}
        </NuxtLink>
        <button @click="handleError" class="btn-retry">
          {{ t('errors.tryAgain') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { NuxtError } from '#app'

const props = defineProps<{
  error: NuxtError
}>()

const { t } = useI18n()

const errorTitle = computed(() => {
  switch (props.error?.statusCode) {
    case 404:
      return t('errors.404.title')
    case 500:
      return t('errors.500.title')
    case 403:
      return t('errors.403.title')
    default:
      return t('errors.boundary')
  }
})

const errorMessage = computed(() => {
  switch (props.error?.statusCode) {
    case 404:
      return t('errors.404.description')
    case 500:
      return t('errors.500.description')
    case 403:
      return t('errors.403.description')
    default:
      return props.error?.message ? props.error.message : t('errors.default')
  }
})

const handleError = () => clearError({ redirect: '/' })

useSeo({ noindex: true })
</script>

<style scoped>
.error-page {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl);
  background: var(--color-bg-alt);
  overscroll-behavior: contain;
  box-sizing: border-box;
}

.error-content {
  text-align: center;
  max-width: 480px;
}

.error-code {
  font-size: var(--text-6xl);
  font-weight: 800;
  color: var(--color-border-dark);
  line-height: 1;
  margin: 0 0 var(--spacing-md);
}

.error-title {
  font-size: var(--text-2xl);
  font-weight: 700;
  color: var(--color-primary);
  margin: 0 0 var(--spacing-sm);
}

.error-message {
  color: var(--color-text-light);
  margin: 0 0 var(--spacing-xl);
  line-height: 1.6;
  overflow-wrap: break-word;
}

.error-actions {
  display: flex;
  gap: var(--spacing-md);
  justify-content: center;
  flex-wrap: wrap;
}

.btn-home,
.btn-retry {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.75rem 1.5rem;
  border-radius: var(--radius-md);
  font-weight: 600;
  text-decoration: none;
  transition: opacity var(--transition-fast), background-color var(--transition-base);
  min-height: 44px;
}

.btn-home {
  background: var(--color-highlight);
  color: var(--color-text-on-primary);
  border: none;
}

.btn-home:hover {
  background: var(--color-highlight-hover);
}

.btn-retry {
  background: var(--color-bg);
  color: var(--color-primary);
  border: 1px solid var(--color-border);
  cursor: pointer;
}

.btn-retry:hover {
  background: var(--color-bg-alt);
}
</style>
