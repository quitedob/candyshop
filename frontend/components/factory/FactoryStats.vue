<template>
  <div class="factory-stats" :class="`factory-stats--${variant}`">
    <div
      v-for="stat in stats"
      :key="stat.id"
      class="stat"
      :class="{ 'stat--highlight': stat.highlight }"
    >
      <div class="stat__icon" v-if="stat.icon">
        <Icon :name="stat.icon" size="24" />
      </div>
      <div class="stat__content">
        <span class="stat__value">{{ stat.value }}</span>
        <span class="stat__label">{{ stat.label }}</span>
      </div>
      <div v-if="stat.description" class="stat__description">
        {{ stat.description }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Stat {
  id: string | number
  value: string
  label: string
  description?: string
  icon?: string
  highlight?: boolean
}

interface Props {
  stats: Stat[]
  variant?: 'default' | 'compact' | 'horizontal'
}

defineProps<Props>()
</script>

<style scoped>
.factory-stats {
  display: grid;
  gap: var(--spacing-md);
}

.factory-stats--default {
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
}

.factory-stats--compact {
  grid-template-columns: repeat(auto-fit, minmax(100px, 1fr));
}

.factory-stats--horizontal {
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-xl);
}

.stat {
  text-align: center;
  padding: var(--spacing-lg);
  background-color: white;
  border-radius: var(--radius-lg);
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.stat:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.stat--highlight {
  background: linear-gradient(135deg, var(--color-highlight) 0%, var(--color-accent) 100%);
  color: white;
}

.stat--highlight .stat__value,
.stat--highlight .stat__label {
  color: white;
}

.stat__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-full);
  color: var(--color-primary);
  margin-bottom: var(--spacing-sm);
}

.stat--highlight .stat__icon {
  background-color: rgba(255, 255, 255, 0.2);
  color: white;
}

.stat__value {
  display: block;
  font-size: var(--text-3xl);
  font-weight: 700;
  color: var(--color-highlight);
  line-height: 1;
  margin-bottom: var(--spacing-xs);
}

.stat__label {
  display: block;
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.stat__description {
  margin-top: var(--spacing-sm);
  font-size: var(--text-xs);
  color: var(--color-text-lighter);
}

/* Compact variant */
.factory-stats--compact .stat {
  padding: var(--spacing-md);
}

.factory-stats--compact .stat__value {
  font-size: var(--text-2xl);
}

.factory-stats--compact .stat__label {
  font-size: var(--text-xs);
}

.factory-stats--compact .stat__icon {
  width: 36px;
  height: 36px;
}

/* Horizontal variant */
.factory-stats--horizontal .stat {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: var(--spacing-md);
  text-align: left;
  padding: var(--spacing-md) var(--spacing-lg);
}

.factory-stats--horizontal .stat__icon {
  margin-bottom: 0;
}

.factory-stats--horizontal .stat__content {
  flex: 1;
}

.factory-stats--horizontal .stat__description {
  display: none;
}

@media (max-width: 640px) {
  .factory-stats--horizontal {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
