<template>
  <div class="error-state">
    <div class="error-state__icon" aria-hidden="true">
      <Icon :name="icon" size="48" />
    </div>
    <h3 class="error-state__title">{{ title || t('errors.boundary') }}</h3>
    <p class="error-state__message">{{ message || t('errors.api.generic_failed') }}</p>
    <button v-if="actionText" type="button" class="btn btn-outline" @click="$emit('action')">
      {{ actionText || t('errors.tryAgain') }}
    </button>
  </div>
</template>

<script setup lang="ts">
const { t } = useI18n()

defineProps({
  title: {
    type: String,
    default: undefined
  },
  message: {
    type: String,
    default: undefined
  },
  icon: {
    type: String,
    default: 'lucide:alert-circle'
  },
  actionText: {
    type: String,
    default: undefined
  }
})

defineEmits(['action'])
</script>

<style scoped>
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-3xl);
  text-align: center;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-xl);
  margin: var(--spacing-xl) 0;
}

.error-state__icon {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-md);
  opacity: 0.5;
}

.error-state__title {
  font-size: var(--text-xl);
  margin-bottom: var(--spacing-sm);
  color: var(--color-text);
}

.error-state__message {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-lg);
  max-width: 400px;
}
</style>
