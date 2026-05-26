<template>
  <div class="empty-state">
    <div class="empty-state__icon" aria-hidden="true">
      <Icon :name="icon" size="48" />
    </div>
    <h3 class="empty-state__title">{{ title }}</h3>
    <p v-if="description" class="empty-state__description">{{ description }}</p>
    <div v-if="actionText || $slots.actions" class="empty-state__actions">
      <slot name="actions">
        <NuxtLink v-if="actionText && actionTo" :to="actionTo" class="btn btn-highlight">
          {{ actionText }}
        </NuxtLink>
        <button v-else-if="actionText" type="button" class="btn btn-highlight" @click="$emit('action')">
          {{ actionText }}
        </button>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  icon?: string
  title?: string
  description?: string
  actionText?: string
  actionTo?: string
}>()

defineEmits<{
  action: []
}>()
</script>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-3xl) var(--spacing-lg);
  text-align: center;
  background: var(--color-bg);
  border: 2px dashed var(--color-border-light);
  border-radius: var(--radius-xl);
  min-height: 280px;
}

.empty-state__icon {
  color: var(--color-text-lighter);
  margin-bottom: var(--spacing-md);
  opacity: 0.4;
}

.empty-state__title {
  font-family: var(--font-display);
  font-size: var(--text-xl);
  color: var(--color-text);
  margin-bottom: var(--spacing-sm);
}

.empty-state__description {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-lg);
  max-width: 420px;
  font-size: var(--text-sm);
  line-height: 1.6;
}

.empty-state__actions {
  display: flex;
  gap: var(--spacing-sm);
}
</style>
