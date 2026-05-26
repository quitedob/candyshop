<template>
  <Teleport to="body">
    <!-- Backdrop -->
    <Transition name="modal-fade">
      <div
        v-if="open"
        class="admin-modal-backdrop"
        @click="$emit('close')"
      />
    </Transition>

    <!-- Panel -->
    <Transition name="modal-scale">
      <div
        v-if="open"
        class="admin-modal"
        :class="widthClass"
        role="dialog"
        aria-modal="true"
      >
        <!-- Header -->
        <div class="admin-modal__header">
          <h2 class="admin-modal__title">{{ title }}</h2>
          <button
            type="button"
            class="admin-modal__close"
            :aria-label="t('close')"
            @click="$emit('close')"
          >
            <Icon name="material-symbols:close" size="20" aria-hidden="true" />
          </button>
        </div>

        <!-- Body -->
        <div class="admin-modal__body">
          <slot />
        </div>

        <!-- Footer -->
        <div v-if="$slots.footer" class="admin-modal__footer">
          <slot name="footer" />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { watch, onBeforeUnmount } from 'vue'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  open: boolean
  title?: string
  width?: 'md' | 'lg' | 'xl'
}>(), {
  width: 'lg',
})

const emit = defineEmits<{
  close: []
}>()

const widthMap: Record<string, string> = {
  md: 'admin-modal--md',
  lg: 'admin-modal--lg',
  xl: 'admin-modal--xl',
}
const widthClass = widthMap[props.width] || widthMap.lg

function onEscape(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}

watch(() => props.open, (val) => {
  if (val) {
    document.body.classList.add('body-lock')
    document.addEventListener('keydown', onEscape)
  } else {
    document.body.classList.remove('body-lock')
    document.removeEventListener('keydown', onEscape)
  }
})

onBeforeUnmount(() => {
  document.body.classList.remove('body-lock')
  document.removeEventListener('keydown', onEscape)
})
</script>

<style scoped>
.admin-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: var(--z-modal-backdrop);
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(2px);
}

.admin-modal {
  position: fixed;
  inset: 0;
  z-index: var(--z-modal);
  display: flex;
  flex-direction: column;
  margin: auto;
  max-height: 90vh;
  background: var(--color-bg);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-2xl);
  overflow: hidden;
}

.admin-modal--md { width: calc(100% - 2rem); max-width: 480px; }
.admin-modal--lg { width: calc(100% - 2rem); max-width: 640px; }
.admin-modal--xl { width: calc(100% - 2rem); max-width: 800px; }

.admin-modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-light);
  flex-shrink: 0;
}

.admin-modal__title {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text);
  margin: 0;
}

.admin-modal__close {
  min-width: 44px;
  min-height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: var(--color-text-lighter);
  cursor: pointer;
  border-radius: var(--radius-md);
  transition: background-color var(--transition-fast), color var(--transition-fast);
}

.admin-modal__close:hover {
  background: var(--color-bg-alt);
  color: var(--color-text);
}

.admin-modal__body {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-lg);
}

.admin-modal__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--spacing-sm);
  padding: var(--spacing-md) var(--spacing-lg);
  border-top: 1px solid var(--color-border-light);
  flex-shrink: 0;
}

/* Transitions */
.modal-fade-enter-active { transition: opacity 0.2s ease; }
.modal-fade-leave-active { transition: opacity 0.15s ease; }
.modal-fade-enter-from,
.modal-fade-leave-to { opacity: 0; }

.modal-scale-enter-active { transition: opacity 0.2s ease, transform 0.2s ease; }
.modal-scale-leave-active { transition: opacity 0.15s ease, transform 0.15s ease; }
.modal-scale-enter-from,
.modal-scale-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

@media (prefers-reduced-motion: reduce) {
  .modal-fade-enter-active,
  .modal-fade-leave-active,
  .modal-scale-enter-active,
  .modal-scale-leave-active {
    transition: none;
  }
}

/* RTL */
[dir="rtl"] .admin-modal__header {
  flex-direction: row-reverse;
}

[dir="rtl"] .admin-modal__footer {
  flex-direction: row-reverse;
}
</style>
