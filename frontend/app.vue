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
        <NuxtLayout>
          <NuxtPage />
        </NuxtLayout>
        <template #error="{ error, clearError }">
          <div class="error-boundary" role="alert">
            <p>{{ error.message || $t('errors.boundary') }}</p>
            <button class="btn btn-highlight btn-sm" @click="clearError">{{ $t('errors.retry') }}</button>
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
const { t } = useI18n()
const titleTemplate = computed(() => t('seo.title_template'))
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
  titleTemplate
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
  background: var(--color-primary);
  color: var(--color-text-on-primary);
  font-weight: 600;
  text-decoration: none;
  border-radius: 0 0 var(--radius-md) 0;
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

/* Toast：右下角为主，预留刘海/手势条；窄屏限制最大宽度避免溢出 */
.toast {
  position: fixed;
  bottom: max(1rem, calc(env(safe-area-inset-bottom, 0px) + 0.5rem));
  right: max(1rem, calc(env(safe-area-inset-right, 0px) + 0.25rem));
  left: auto;
  max-width: min(28rem, calc(100vw - 2rem - env(safe-area-inset-left, 0px) - env(safe-area-inset-right, 0px)));
  padding: 0.75rem 1.5rem;
  border-radius: var(--radius-md);
  color: var(--color-text-on-primary);
  font-weight: 500;
  z-index: 10000;
  box-shadow: var(--shadow-lg);
  box-sizing: border-box;
  overscroll-behavior: contain;
}
.toast--success { background-color: var(--color-success); }
.toast--error { background-color: var(--color-error); }
.toast--warning { background-color: var(--color-warning); }
.toast--info { background-color: var(--color-info); }
.toast-enter-active,
.toast-leave-active { transition: opacity 0.3s ease, transform 0.3s ease; }
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
  transform-origin: left;
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
