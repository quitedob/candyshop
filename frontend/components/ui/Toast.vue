<template>
  <Transition name="toast">
    <div v-if="visible" class="toast" :class="`toast-${type}`" role="alert">
      <div class="toast-icon" aria-hidden="true">
        <Icon :name="iconName" />
      </div>
      <div class="toast-content">
        <p class="toast-message">{{ message }}</p>
      </div>
      <button type="button" @click="close" class="toast-close" :aria-label="$t('common.close')">
        <Icon name="mdi:close" aria-hidden="true" />
      </button>
    </div>
  </Transition>
</template>

<script setup lang="ts">
const props = defineProps<{
  message: string
  type?: 'success' | 'error' | 'warning' | 'info'
  duration?: number
}>()

const visible = ref(true)

const iconName = computed(() => {
  switch (props.type) {
    case 'success': return 'mdi:check-circle'
    case 'error': return 'mdi:alert-circle'
    case 'warning': return 'mdi:alert'
    case 'info': return 'mdi:information'
    default: return 'mdi:information'
  }
})

const close = () => {
  visible.value = false
}

onMounted(() => {
  if (props.duration) {
    setTimeout(close, props.duration)
  }
})
</script>

<style scoped>
.toast {
  position: fixed;
  top: 1rem;
  right: 1rem;
  padding: 1rem 1.5rem;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  gap: 0.75rem;
  box-shadow: var(--shadow-lg);
  z-index: 9999;
  color: var(--color-text-on-primary);
}

.toast-success { background: var(--color-success); }
.toast-error { background: var(--color-error); }
.toast-warning { background: var(--color-warning); }
.toast-info { background: var(--color-info); }

.toast-icon {
  font-size: 1.25rem;
}

.toast-message {
  margin: 0;
  font-weight: 500;
}

.toast-close {
  background: none;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 0;
  opacity: 0.7;
  transition: opacity var(--transition-fast);
  min-width: 44px;
  min-height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.toast-close:hover {
  opacity: 1;
}

.toast-enter-active,
.toast-leave-active {
  transition: opacity var(--transition-base), transform var(--transition-base);
}

@media (prefers-reduced-motion: reduce) {
  .toast-enter-active,
  .toast-leave-active {
    transition: opacity 0.01ms ease;
  }
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
</style>
