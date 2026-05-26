<template>
  <span class="status-badge" :class="badgeClass">
    <slot>{{ displayLabel }}</slot>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const statusColorMap: Record<string, Record<string, string>> = {
  order: {
    pending: 'badge-warning',
    confirmed: 'badge-info',
    processing: 'badge-production',
    production: 'badge-production',
    allocated: 'badge-info',
    fulfilled: 'badge-success',
    shipped: 'badge-success',
    delivered: 'badge-success',
    cancelled: 'badge-error',
    returned: 'badge-error',
    refunded: 'badge-error',
    pending_confirmation: 'badge-warning',
  },
  inquiry: {
    new: 'badge-info',
    assigned: 'badge-production',
    quoted: 'badge-accent',
    responded: 'badge-success',
    closed: 'badge-default',
    archived: 'badge-default',
  },
  product: {
    active: 'badge-success',
    inactive: 'badge-default',
    draft: 'badge-warning',
    discontinued: 'badge-error',
  },
  user: {
    customer: 'badge-info',
    admin: 'badge-accent',
    superadmin: 'badge-highlight',
    active: 'badge-success',
    inactive: 'badge-default',
    pending: 'badge-warning',
    suspended: 'badge-error',
  },
  trade: {
    draft: 'badge-default',
    submitted: 'badge-info',
    in_review: 'badge-warning',
    approved: 'badge-success',
    rejected: 'badge-error',
  },
  payment: {
    pending: 'badge-warning',
    paid: 'badge-success',
    failed: 'badge-error',
    refunded: 'badge-info',
  },
}

const props = withDefaults(defineProps<{
  status?: string
  type?: 'order' | 'inquiry' | 'product' | 'user' | 'trade' | 'payment'
  label?: string
}>(), {
  type: 'order',
})

const { enumLabel } = useDisplay()

const enumGroupByType: Record<string, string> = {
  order: 'order_status',
  inquiry: 'inquiry_status',
  product: 'product_status',
  user: 'user_status',
  trade: 'trade_status',
  payment: 'payment_status',
}

const displayLabel = computed(() => {
  if (props.label) return props.label
  if (!props.status) return '—'
  const group = enumGroupByType[props.type] || 'order_status'
  return enumLabel(group, props.status)
})

const badgeClass = computed(() => {
  if (!props.status) return 'badge-default'
  const map = statusColorMap[props.type]
  if (!map) return 'badge-default'
  return map[props.status] || 'badge-default'
})
</script>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 10px;
  font-size: var(--text-xs);
  font-weight: 600;
  border-radius: var(--radius-sm);
  text-transform: none;
  white-space: nowrap;
  line-height: 1.5;
}

/* Color classes mapped to CSS variables */
.status-badge.badge-warning {
  background: rgba(var(--color-warning-rgb), 0.12);
  color: var(--color-warning-hover);
}
.status-badge.badge-info {
  background: rgba(var(--color-info-rgb), 0.12);
  color: var(--color-info-hover);
}
.status-badge.badge-success {
  background: rgba(var(--color-success-rgb), 0.12);
  color: var(--color-success);
}
.status-badge.badge-error {
  background: rgba(var(--color-error-rgb), 0.12);
  color: var(--color-error);
}
.status-badge.badge-production {
  background: rgba(var(--color-production-rgb), 0.12);
  color: var(--color-accent-dark);
}
.status-badge.badge-accent {
  background: var(--color-highlight-light);
  color: var(--color-accent-dark);
}
.status-badge.badge-highlight {
  background: var(--color-highlight-light);
  color: var(--color-highlight-hover);
}
.status-badge.badge-default {
  background: var(--color-bg-alt);
  color: var(--color-text-lighter);
}
</style>
