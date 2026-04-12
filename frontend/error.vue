<template>
  <div class="error-page">
    <div class="error-content">
      <h1 class="error-code">{{ error?.statusCode || 404 }}</h1>
      <h2 class="error-title">{{ errorTitle }}</h2>
      <p class="error-message">{{ errorMessage }}</p>
      <div class="error-actions">
        <NuxtLink to="/" class="btn-home">
          {{ $t('errors.goHome') }}
        </NuxtLink>
        <button @click="handleError" class="btn-retry">
          {{ $t('errors.tryAgain') }}
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

const errorTitle = computed(() => {
  switch (props.error?.statusCode) {
    case 404:
      return 'Page Not Found'
    case 500:
      return 'Server Error'
    case 403:
      return 'Access Denied'
    default:
      return 'Something went wrong'
  }
})

const errorMessage = computed(() => {
  switch (props.error?.statusCode) {
    case 404:
      return 'The page you are looking for does not exist or has been moved.'
    case 500:
      return 'We encountered an unexpected error. Please try again later.'
    case 403:
      return 'You do not have permission to access this page.'
    default:
      return props.error?.message || 'An unexpected error occurred.'
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 2rem;
}

.error-content {
  text-align: center;
  color: white;
}

.error-code {
  font-size: 8rem;
  font-weight: 700;
  margin: 0;
  line-height: 1;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.2);
}

.error-title {
  font-size: 2rem;
  margin: 1rem 0;
}

.error-message {
  font-size: 1.125rem;
  opacity: 0.9;
  max-width: 400px;
  margin: 0 auto 2rem;
}

.error-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
}

.btn-home,
.btn-retry {
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-home {
  background: white;
  color: #667eea;
  text-decoration: none;
}

.btn-home:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.btn-retry {
  background: transparent;
  border: 2px solid white;
  color: white;
}

.btn-retry:hover {
  background: rgba(255, 255, 255, 0.1);
}
</style>
