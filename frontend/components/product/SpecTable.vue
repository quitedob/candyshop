<template>
  <div class="spec-table" :class="`spec-table--${variant}`">
    <table>
      <tbody>
        <tr v-for="row in rows" :key="row.label">
          <th class="spec-table__label">{{ row.label }}</th>
          <td class="spec-table__value">
            <template v-if="row.type === 'badge'">
              <span
                v-for="(value, index) in row.values"
                :key="index"
                class="spec-table__badge"
                :class="`spec-table__badge--${value.variant || 'default'}`"
              >
                <Icon v-if="value.icon" :name="value.icon" size="14" />
                {{ value.label }}
              </span>
            </template>
            <template v-else-if="row.type === 'link'">
              <a :href="row.value" target="_blank" rel="noopener noreferrer" class="spec-table__link">
                {{ row.displayValue || row.value }}
                <Icon name="lucide:external-link" size="14" />
              </a>
            </template>
            <template v-else-if="row.type === 'icon'">
              <span class="spec-table__icon">
                <Icon :name="row.icon || 'lucide:info'" size="16" />
                {{ row.value }}
              </span>
            </template>
            <template v-else>
              {{ row.value }}
            </template>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
interface SpecBadge {
  label: string
  icon?: string
  variant?: 'default' | 'success' | 'warning' | 'error' | 'info'
}

interface SpecRow {
  label: string
  value?: string | string[]
  displayValue?: string
  icon?: string
  type?: 'text' | 'badge' | 'link' | 'icon'
  values?: SpecBadge[]
}

interface Props {
  rows: SpecRow[]
  variant?: 'default' | 'compact' | 'bordered'
}

defineProps<Props>()
</script>

<style scoped>
.spec-table {
  width: 100%;
  border-collapse: collapse;
}

.spec-table table {
  width: 100%;
}

.spec-table tr {
  border-bottom: 1px solid var(--color-border-light);
}

.spec-table tr:last-child {
  border-bottom: none;
}

.spec-table__label,
.spec-table__value {
  padding: var(--spacing-md);
  text-align: left;
}

.spec-table__label {
  width: 40%;
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-text-light);
  background-color: var(--color-bg-alt);
}

.spec-table__value {
  font-size: var(--text-sm);
  color: var(--color-text);
}

/* Bordered variant */
.spec-table--bordered tr {
  border: 1px solid var(--color-border);
}

/* Compact variant */
.spec-table--compact .spec-table__label,
.spec-table--compact .spec-table__value {
  padding: var(--spacing-sm) var(--spacing-md);
}

.spec-table--compact .spec-table__label {
  font-size: var(--text-xs);
}

.spec-table--compact .spec-table__value {
  font-size: var(--text-xs);
}

/* Badge */
.spec-table__badge {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-xs);
  font-weight: 500;
  border-radius: var(--radius-sm);
  margin-right: var(--spacing-xs);
  margin-bottom: var(--spacing-xs);
}

.spec-table__badge--default {
  background-color: var(--color-bg-alt);
  color: var(--color-text);
}

.spec-table__badge--success {
  background-color: #dcfce7;
  color: #166534;
}

.spec-table__badge--warning {
  background-color: #fef3c7;
  color: #92400e;
}

.spec-table__badge--error {
  background-color: #fef2f2;
  color: #dc2626;
}

.spec-table__badge--info {
  background-color: #dbeafe;
  color: #1e40af;
}

/* Link */
.spec-table__link {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--color-highlight);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.spec-table__link:hover {
  color: var(--color-highlight-hover);
  text-decoration: underline;
}

/* Icon */
.spec-table__icon {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
}

/* Responsive */
@media (max-width: 640px) {
  .spec-table__label,
  .spec-table__value {
    display: block;
    width: 100%;
    padding: var(--spacing-sm);
  }

  .spec-table__label {
    background-color: transparent;
    color: var(--color-text-light);
    font-size: var(--text-xs);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .spec-table__value {
    padding-top: var(--spacing-xs);
    padding-bottom: var(--spacing-md);
  }
}

/* Grid variant for grouped specifications */
.spec-table--group {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: var(--spacing-md);
}

.spec-table--group tr {
  display: block;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.spec-table--group .spec-table__label,
.spec-table--group .spec-table__value {
  display: block;
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
}

.spec-table--group .spec-table__label {
  background-color: var(--color-bg-alt);
  border-bottom: 1px solid var(--color-border-light);
  border-radius: var(--radius-md) var(--radius-md) 0 0;
}

.spec-table--group .spec-table__value {
  border-radius: 0 0 var(--radius-md) var(--radius-md);
}
</style>
