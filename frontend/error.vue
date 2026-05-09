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
</script>

<style scoped>
.error-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
}

.error-content {
  text-align: center;
  max-width: 480px;
}

.error-code {
  font-size: 6rem;
  font-weight: 800;
  color: #cbd5e1;
  line-height: 1;
  margin: 0 0 1rem;
}

.error-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #1e293b;
  margin: 0 0 0.75rem;
}

.error-message {
  color: #64748b;
  margin: 0 0 2rem;
  line-height: 1.6;
}

.error-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

.btn-home,
.btn-retry {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  font-weight: 600;
  text-decoration: none;
  transition: opacity 0.2s;
}

.btn-home {
  background: #ea580c;
  color: white;
}

.btn-home:hover {
  opacity: 0.92;
}

.btn-retry {
  background: white;
  color: #1e293b;
  border: 1px solid #e2e8f0;
  cursor: pointer;
}

.btn-retry:hover {
  background: #f8fafc;
}
</style>
