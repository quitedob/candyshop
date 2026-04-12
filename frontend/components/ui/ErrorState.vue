<template>
  <div class="error-state">
    <div class="error-state__icon">
      <Icon :name="icon" size="48" />
    </div>
    <h3 class="error-state__title">{{ title }}</h3>
    <p class="error-state__message">{{ message }}</p>
    <button v-if="actionText" class="btn btn-outline" @click="$emit('action')">
      {{ actionText }}
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps({
  title: {
    type: String,
    default: 'Error Loading Data'
  },
  message: {
    type: String,
    default: 'Something went wrong while fetching data. Please check your connection and try again.'
  },
  icon: {
    type: String,
    default: 'lucide:alert-circle'
  },
  actionText: {
    type: String,
    default: 'Try Again'
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
