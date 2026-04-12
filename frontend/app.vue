<template>
  <div>
    <!-- Skip to content for accessibility -->
    <a href="#main-content" class="skip-to-content">
      {{ $t('skip_to_content') }}
    </a>

    <!-- Header -->
    <Header v-if="showMarketingChrome" />

    <!-- Main content -->
    <main id="main-content">
      <NuxtErrorBoundary>
        <NuxtPage />
        <template #error="{ error, clearError }">
          <div class="error-boundary">
            <p>{{ error.message || 'Something went wrong' }}</p>
            <button class="btn btn-highlight btn-sm" @click="clearError">{{ $t('errors.retry') || 'Try again' }}</button>
          </div>
        </template>
      </NuxtErrorBoundary>
    </main>

    <!-- Footer -->
    <Footer v-if="showMarketingChrome" />

    <!-- Floating Inquiry Button -->
    <InquiryFloating v-if="showMarketingChrome" />

    <!-- Toast Notifications -->
    <Teleport to="body">
      <Transition name="toast">
        <div v-if="toast.state.visible" :class="['toast', `toast--${toast.state.type}`]" role="alert">
          {{ toast.state.message }}
        </div>
      </Transition>
    </Teleport>

    <!-- Page transition overlay (optional) -->
    <NuxtLoadingIndicator />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const route = useRoute()
const toast = useToast()
const authLegacyRoutes = ['/login', '/register', '/forgot-password', '/reset-password']

const showMarketingChrome = computed(() => {
  const path = route.path
  // Note: App routes and Auth logic regardless of locale prefix (e.g. /zh/auth/login)
  const isAppRoute = /\/(admin|customer|auth)(\/|$)/.test(path)
  const isLegacyAuth = authLegacyRoutes.some(p => path.endsWith(p))
  
  return !(isAppRoute || isLegacyAuth)
})

// Set page title template
useHead({
  titleTemplate: '%s | Candy Manufacturer - OEM & Private Label Solutions'
})
</script>

<style>
/* Skip to content link */
.skip-to-content {
  position: absolute;
  left: -9999px;
  top: auto;
  width: 1px;
  height: 1px;
  overflow: hidden;
  z-index: 10000;
  padding: 0.75rem 1.5rem;
  background: var(--color-primary, #1a1f3a);
  color: #fff;
  font-weight: 600;
  text-decoration: none;
  border-radius: 0 0 0.5rem 0;
}
.skip-to-content:focus {
  position: fixed;
  left: 0;
  top: 0;
  width: auto;
  height: auto;
  overflow: visible;
}

/* Error boundary */
.error-boundary {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  min-height: 40vh;
  padding: 2rem;
  text-align: center;
}

/* Toast notifications */
.toast {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  padding: 0.75rem 1.5rem;
  border-radius: 0.5rem;
  color: #fff;
  font-weight: 500;
  z-index: 10000;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}
.toast--success { background-color: #22c55e; }
.toast--error { background-color: #ef4444; }
.toast--warning { background-color: #f59e0b; }
.toast--info { background-color: #3b82f6; }
.toast-enter-active,
.toast-leave-active { transition: all 0.3s ease; }
.toast-enter-from,
.toast-leave-to { opacity: 0; transform: translateY(1rem); }

/* Nuxt loading indicator styles */
.nuxt-loading-indicator {
  position: fixed;
  top: 0;
  right: 0;
  left: 0;
  height: 3px;
  width: 0%;
  transition: width 0.2s, opacity 0.4s;
  opacity: 1;
  background: linear-gradient(
    to right,
    var(--color-highlight) 0%,
    var(--color-accent) 50%,
    var(--color-primary) 100%
  );
  z-index: 9999;
}

.nuxt-loading-indicator.nuxt-loading-indicator-enter {
  width: 0% !important;
}

.nuxt-loading-indicator.nuxt-loading-indicator-leave {
  opacity: 0;
}
</style>
