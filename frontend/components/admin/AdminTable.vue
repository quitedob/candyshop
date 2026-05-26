<template>
  <div class="admin-table-wrapper">
    <!-- Header bar with actions -->
    <div v-if="$slots['header-actions']" class="admin-table__header actions">
      <slot name="header-actions" />
    </div>

    <!-- Loading: skeleton rows -->
    <div v-if="loading" class="admin-table__table" role="status" aria-label="Loading">
      <div class="admin-table__skeleton-header">
        <div v-for="col in columns" :key="col.key" class="skeleton" style="height: 1rem; width: auto;" />
      </div>
      <div v-for="i in 5" :key="i" class="admin-table__skeleton-row">
        <div v-for="col in columns" :key="col.key" class="skeleton" style="height: 1rem;" />
      </div>
    </div>

    <!-- Error state -->
    <ErrorState
      v-else-if="error"
      :title="errorTitle"
      :message="errorMessage"
      @action="$emit('retry')"
    />

    <!-- Empty state -->
    <EmptyState
      v-else-if="!rows || rows.length === 0"
      :icon="emptyIcon"
      :title="emptyText"
      :description="emptyDescription"
    >
      <template v-if="$slots['empty-actions']" #actions>
        <slot name="empty-actions" />
      </template>
    </EmptyState>

    <!-- Table -->
    <div v-else class="admin-table__scroll">
      <table class="admin-table__table">
        <thead class="admin-table__thead">
          <tr>
            <th
              v-for="col in columns"
              :key="col.key"
              class="admin-table__th"
              :style="col.width ? { width: col.width } : {}"
              :class="{ 'admin-table__th--sortable': col.sortable }"
            >
              {{ col.label }}
            </th>
            <th v-if="$slots['row-actions']" class="admin-table__th admin-table__th--actions" />
          </tr>
        </thead>
        <tbody class="admin-table__tbody">
          <tr
            v-for="(row, idx) in rows"
            :key="row.id || idx"
            class="admin-table__tr"
          >
            <td
              v-for="col in columns"
              :key="col.key"
              class="admin-table__td"
            >
              <slot :name="`cell-${col.key}`" :row="row" :value="row[col.key]">
                {{ row[col.key] }}
              </slot>
            </td>
            <td v-if="$slots['row-actions']" class="admin-table__td admin-table__td--actions">
              <slot name="row-actions" :row="row" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Bottom bar -->
    <div v-if="$slots['bottom']" class="admin-table__bottom">
      <slot name="bottom" />
    </div>
  </div>
</template>

<script setup lang="ts">
interface Column {
  key: string
  label: string
  sortable?: boolean
  width?: string
}

withDefaults(defineProps<{
  columns: Column[]
  rows?: Record<string, any>[]
  loading?: boolean
  error?: boolean
  errorTitle?: string
  errorMessage?: string
  emptyText?: string
  emptyDescription?: string
  emptyIcon?: string
}>(), {
  loading: false,
  error: false,
  emptyIcon: 'material-symbols:inbox',
  emptyText: 'No data found',
})

defineEmits<{
  retry: []
}>()
</script>

<style scoped>
.admin-table-wrapper {
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.admin-table__header {
  padding: var(--spacing-md) var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-light);
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  background: var(--color-bg-alt);
}

.admin-table__scroll {
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

.admin-table__table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--text-sm);
}

.admin-table__thead {
  background: var(--color-bg-alt);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 10;
}

.admin-table__th {
  padding: var(--spacing-sm) var(--spacing-lg);
  text-align: left;
  font-size: var(--text-xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-lighter);
  white-space: nowrap;
}

.admin-table__th--sortable {
  cursor: pointer;
  user-select: none;
}

.admin-table__th--sortable:hover {
  color: var(--color-text);
}

.admin-table__th--actions {
  width: 60px;
  text-align: right;
}

.admin-table__tbody {
  background: var(--color-bg);
}

.admin-table__tr {
  border-bottom: 1px solid var(--color-border-light);
  transition: background-color var(--transition-fast);
}

.admin-table__tr:hover {
  background: var(--color-bg-alt);
}

.admin-table__td {
  padding: var(--spacing-sm) var(--spacing-lg);
  color: var(--color-text);
  vertical-align: middle;
}

.admin-table__td--actions {
  text-align: right;
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.admin-table__tr:hover .admin-table__td--actions {
  opacity: 1;
}

/* Skeleton */
.admin-table__skeleton-header {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(80px, 1fr));
  gap: var(--spacing-lg);
  padding: var(--spacing-sm) var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-light);
  background: var(--color-bg-alt);
}

.admin-table__skeleton-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(80px, 1fr));
  gap: var(--spacing-lg);
  padding: var(--spacing-sm) var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-light);
}

.admin-table__bottom {
  padding: var(--spacing-md) var(--spacing-lg);
  border-top: 1px solid var(--color-border-light);
  background: var(--color-bg-alt);
}

/* RTL */
[dir="rtl"] .admin-table__th {
  text-align: right;
}

[dir="rtl"] .admin-table__th--actions,
[dir="rtl"] .admin-table__td--actions {
  text-align: left;
}
</style>
