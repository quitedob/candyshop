<template>
  <component
    :is="tag"
    :to="to"
    :href="href"
    :target="target"
    :type="tag === 'button' ? nativeType : undefined"
    :disabled="disabled || loading"
    :class="buttonClasses"
    @click="handleClick"
  >
    <!-- Loading spinner -->
    <span v-if="loading" class="btn__spinner">
      <Icon name="lucide:loader-2" size="16" />
    </span>

    <!-- Icon before -->
    <Icon v-if="icon && iconPosition === 'start'" :name="icon" :size="iconSize" />

    <!-- Slot content -->
    <span v-if="$slots.default" :class="{'btn__text': true, 'btn__text--hidden': loading}">
      <slot />
    </span>

    <!-- Badge -->
    <span v-if="badge" class="btn__badge">{{ badge }}</span>

    <!-- Icon after -->
    <Icon v-if="icon && iconPosition === 'end'" :name="icon" :size="iconSize" />
  </component>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'primary' | 'secondary' | 'accent' | 'highlight' | 'outline' | 'ghost' | 'link'
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  tag?: 'button' | 'a' | 'nuxt-link'
  to?: string
  href?: string
  target?: '_blank' | '_self'
  nativeType?: 'button' | 'submit' | 'reset'
  disabled?: boolean
  loading?: boolean
  block?: boolean
  rounded?: boolean
  icon?: string
  iconPosition?: 'start' | 'end'
  iconSize?: number
  badge?: string | number
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  tag: 'button',
  nativeType: 'button',
  disabled: false,
  loading: false,
  block: false,
  rounded: false,
  iconPosition: 'start',
  iconSize: 18
})

const emit = defineEmits<{
  click: [event: Event]
}>()

const handleClick = (event: Event) => {
  if (!props.disabled && !props.loading) {
    emit('click', event)
  }
}

const buttonClasses = computed(() => {
  return [
    'btn',
    `btn--${props.variant}`,
    `btn--${props.size}`,
    {
      'btn--block': props.block,
      'btn--rounded': props.rounded,
      'btn--disabled': props.disabled,
      'btn--loading': props.loading,
      'btn--has-icon': props.icon,
      'btn--icon-only': props.icon && !$slots.default
    }
  ]
})
</script>

<style scoped>
.btn {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  font-family: var(--font-body);
  font-weight: 600;
  line-height: 1.5;
  text-align: center;
  white-space: nowrap;
  user-select: none;
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
  cursor: pointer;
  border: 2px solid transparent;
  border-radius: var(--radius-md);
}

.btn:focus-visible {
  outline: 2px solid var(--color-highlight);
  outline-offset: 2px;
}

.btn--disabled {
  opacity: 0.6;
  cursor: not-allowed;
  pointer-events: none;
}

.btn--loading {
  pointer-events: none;
}

/* Spinner */
.btn__spinner {
  display: flex;
  align-items: center;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.btn__text--hidden {
  visibility: hidden;
}

/* Badge */
.btn__badge {
  position: absolute;
  top: -6px;
  right: -6px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  font-size: 10px;
  font-weight: 700;
  line-height: 18px;
  text-align: center;
  background-color: var(--color-highlight);
  color: white;
  border-radius: var(--radius-full);
  box-shadow: var(--shadow-sm);
}

/* Icon only */
.btn--icon-only {
  padding: 0;
  gap: 0;
}

/* Block */
.btn--block {
  display: flex;
  width: 100%;
}

/* Rounded */
.btn--rounded {
  border-radius: var(--radius-full);
}

/* Sizes */
.btn--xs {
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-xs);
  gap: var(--spacing-xs);
}

.btn--sm {
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-sm);
  gap: var(--spacing-xs);
}

.btn--md {
  padding: var(--spacing-sm) var(--spacing-lg);
  font-size: var(--text-base);
}

.btn--lg {
  padding: var(--spacing-md) var(--spacing-xl);
  font-size: var(--text-lg);
  gap: var(--spacing-sm);
}

.btn--xl {
  padding: var(--spacing-lg) var(--spacing-2xl);
  font-size: var(--text-xl);
  gap: var(--spacing-md);
}

.btn--icon-only.btn--xs {
  width: 44px;
  height: 44px;
}

.btn--icon-only.btn--sm {
  width: 44px;
  height: 44px;
}

.btn--icon-only.btn--md {
  width: 40px;
  height: 40px;
}

.btn--icon-only.btn--lg {
  width: 48px;
  height: 48px;
}

.btn--icon-only.btn--xl {
  width: 56px;
  height: 56px;
}

/* Variants */
.btn--primary {
  background-color: var(--color-primary);
  border-color: var(--color-primary);
  color: var(--color-text-on-primary);
}

.btn--primary:hover:not(.btn--disabled) {
  background-color: var(--color-primary-light);
  border-color: var(--color-primary-light);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.btn--secondary {
  background-color: var(--color-text);
  border-color: var(--color-text);
  color: var(--color-text-on-primary);
}

.btn--secondary:hover:not(.btn--disabled) {
  background-color: var(--color-primary);
  border-color: var(--color-primary);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.btn--accent {
  background-color: var(--color-accent);
  border-color: var(--color-accent);
  color: var(--color-text-on-primary);
}

.btn--accent:hover:not(.btn--disabled) {
  background-color: var(--color-accent-dark);
  border-color: var(--color-accent-dark);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.btn--highlight {
  background-color: var(--color-highlight);
  border-color: var(--color-highlight);
  color: var(--color-text-on-primary);
}

.btn--highlight:hover:not(.btn--disabled) {
  background-color: var(--color-highlight-hover);
  border-color: var(--color-highlight-hover);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.btn--outline {
  background-color: transparent;
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.btn--outline:hover:not(.btn--disabled) {
  background-color: var(--color-primary);
  color: var(--color-text-on-primary);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.btn--ghost {
  background-color: transparent;
  border-color: transparent;
  color: var(--color-text);
}

.btn--ghost:hover:not(.btn--disabled) {
  background-color: var(--color-bg-alt);
}

.btn--link {
  background-color: transparent;
  border-color: transparent;
  color: var(--color-highlight);
  padding: 0;
  border-radius: 0;
}

.btn--link:hover:not(.btn--disabled) {
  color: var(--color-highlight-hover);
  text-decoration: underline;
}

/* Dark theme support */
@media (prefers-color-scheme: dark) {
  .btn--outline {
    border-color: var(--color-text-on-primary);
    color: var(--color-text-on-primary);
  }

  .btn--ghost {
    color: var(--color-text-on-primary);
  }
}
</style>
