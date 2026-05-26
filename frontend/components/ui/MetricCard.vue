<template>
  <component
    :is="to ? resolveComponent('NuxtLink') : 'div'"
    :to="to"
    class="metric-card"
    :class="{ 'metric-card--clickable': !!to, 'curator-lift': !!to }"
  >
    <!-- Icon -->
    <div v-if="icon" class="metric-card__icon" :class="iconBgClass">
      <Icon :name="icon" size="24" aria-hidden="true" />
    </div>

    <!-- Skeleton loading -->
    <template v-if="loading">
      <div class="skeleton" style="height: 2rem; width: 60%;" />
      <div class="skeleton" style="height: 1rem; width: 40%; margin-top: 8px;" />
    </template>

    <!-- Content -->
    <template v-else>
      <div class="metric-card__value">{{ formattedValue }}</div>
      <div class="metric-card__title">{{ title }}</div>

      <!-- Trend -->
      <div v-if="trend && trend !== 'none'" class="metric-card__trend" :class="`metric-card__trend--${trend}`">
        <Icon
          :name="trend === 'up' ? 'material-symbols:trending-up' : 'material-symbols:trending-down'"
          size="14"
          aria-hidden="true"
        />
        <span>{{ trendLabel }}</span>
      </div>
    </template>
  </component>
</template>

<script setup lang="ts">
import { computed, resolveComponent } from 'vue'

const props = withDefaults(defineProps<{
  title?: string
  value?: string | number
  trend?: 'up' | 'down' | 'none'
  trendLabel?: string
  icon?: string
  iconBg?: 'default' | 'success' | 'warning' | 'error' | 'info' | 'accent'
  loading?: boolean
  to?: string
}>(), {
  trend: 'none',
  iconBg: 'default',
  loading: false,
})

const formattedValue = computed(() => {
  const v = props.value
  if (v === undefined || v === null) return '—'
  return String(v)
})

const iconBgClass = computed(() => `metric-card__icon--${props.iconBg}`)
</script>

<style scoped>
.metric-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.metric-card--clickable {
  cursor: pointer;
}

.metric-card__icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: auto;
  border: 1px solid var(--color-border-light);
}

.metric-card__icon--default {
  background: var(--color-bg-alt);
  color: var(--color-accent);
}

.metric-card__icon--success {
  background: rgba(var(--color-success-rgb), 0.1);
  color: var(--color-success);
}

.metric-card__icon--warning {
  background: rgba(var(--color-warning-rgb), 0.1);
  color: var(--color-warning);
}

.metric-card__icon--error {
  background: rgba(var(--color-error-rgb), 0.1);
  color: var(--color-error);
}

.metric-card__icon--info {
  background: rgba(var(--color-info-rgb), 0.1);
  color: var(--color-info);
}

.metric-card__icon--accent {
  background: var(--color-primary-container);
  color: var(--color-highlight);
}

.metric-card__value {
  font-family: var(--font-display);
  font-size: var(--text-4xl);
  font-weight: 700;
  color: var(--color-primary);
  line-height: 1.2;
}

.metric-card__title {
  font-size: var(--text-sm);
  color: var(--color-text-lighter);
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.metric-card__trend {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: var(--text-xs);
  font-weight: 600;
  margin-top: var(--spacing-xs);
}

.metric-card__trend--up {
  color: var(--color-success);
}

.metric-card__trend--down {
  color: var(--color-error);
}
</style>
