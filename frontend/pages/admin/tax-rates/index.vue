<template>
  <div>
    <PageHeader :title="t('admin.taxRates.title')" :description="t('admin.taxRates.description')">
      <template #actions>
        <button type="button" class="btn btn-highlight inline-flex items-center gap-2" @click="openCreate">
          <Icon name="material-symbols:add" size="18" aria-hidden="true" />
          {{ t('admin.taxRates.create') }}
        </button>
      </template>
    </PageHeader>

    <div class="tax-rates-toolbar">
      <div class="tax-rates-toolbar__field tax-rates-toolbar__field--grow">
        <label class="admin-form-label admin-form-label--xs" for="tax-rates-search">{{ t('admin.products.search') }}</label>
        <input
          id="tax-rates-search"
          v-model="searchQuery"
          type="search"
          autocomplete="off"
          class="admin-form-control admin-form-control--flat"
          :placeholder="t('admin.taxRates.col_country')"
        />
      </div>
      <div class="tax-rates-toolbar__field">
        <label class="admin-form-label admin-form-label--xs" for="tax-rates-status">{{ t('admin.taxRates.col_status') }}</label>
        <select id="tax-rates-status" v-model="statusFilter" class="admin-form-control admin-form-control--flat">
          <option value="">{{ t('admin.products.status_all') }}</option>
          <option value="active">{{ t('admin.status_options.active') }}</option>
          <option value="inactive">{{ t('admin.status_options.inactive') }}</option>
        </select>
      </div>
      <div class="tax-rates-toolbar__stats">
        <span class="tax-rates-stat">
          <span class="tax-rates-stat__value">{{ allRates.length }}</span>
          <span class="tax-rates-stat__label">{{ t('admin.taxRates.col_country') }}</span>
        </span>
        <span class="tax-rates-stat tax-rates-stat--active">
          <span class="tax-rates-stat__value">{{ activeCount }}</span>
          <span class="tax-rates-stat__label">{{ t('admin.status_options.active') }}</span>
        </span>
      </div>
    </div>

    <AdminTable
      class="tax-rates-table"
      :columns="columns"
      :rows="filteredRates"
      :loading="pending"
      :error="!!error"
      :error-message="error"
      :empty-text="t('admin.taxRates.no_data')"
      @retry="fetchRates"
    >
      <template #empty-actions>
        <button type="button" class="btn btn-highlight" @click="openCreate">{{ t('admin.taxRates.create') }}</button>
      </template>
      <template #cell-country="{ row }">
        <span class="tax-rate-code">{{ row.country || '—' }}</span>
      </template>
      <template #cell-region="{ row }">
        <span v-if="row.region" class="tax-rate-region">{{ row.region }}</span>
        <span v-else class="tax-rate-muted">—</span>
      </template>
      <template #cell-name="{ row }">
        <span class="tax-rate-name">{{ row.name || '—' }}</span>
      </template>
      <template #cell-rate="{ row }">
        <span class="tax-rate-value">{{ formatPercent(row.rate) }}</span>
      </template>
      <template #cell-isActive="{ row }">
        <StatusBadge :status="row.isActive ? 'active' : 'inactive'" type="product" />
      </template>
      <template #cell-actions="{ row }">
        <div class="tax-rate-actions">
          <button type="button" class="admin-btn-ghost" @click="openEdit(row)">{{ t('admin.taxRates.edit') }}</button>
          <button type="button" class="admin-btn-danger-ghost" @click="removeRate(row.id)">{{ t('admin.taxRates.delete') }}</button>
        </div>
      </template>
    </AdminTable>

    <AdminModal
      :open="showModal"
      :title="editingId ? t('admin.taxRates.edit') : t('admin.taxRates.create')"
      width="lg"
      @close="closeModal"
    >
      <form id="tax-rate-form" class="tax-rate-form" @submit.prevent="saveRate">
        <div class="tax-rate-form__grid">
          <div>
            <label class="admin-form-label" for="tax-country">{{ t('admin.taxRates.country') }}</label>
            <input
              id="tax-country"
              v-model="form.country"
              required
              maxlength="2"
              class="admin-form-control uppercase"
              placeholder="US"
            />
          </div>
          <div>
            <label class="admin-form-label" for="tax-region">{{ t('admin.taxRates.region') }}</label>
            <input
              id="tax-region"
              v-model="form.region"
              class="admin-form-control uppercase"
              placeholder="NY"
            />
          </div>
          <div>
            <label class="admin-form-label" for="tax-rate">{{ t('admin.taxRates.rate') }}</label>
            <input
              id="tax-rate"
              v-model.number="form.rate"
              type="number"
              step="0.0001"
              min="0"
              max="1"
              required
              class="admin-form-control"
            />
            <p class="admin-form-hint mt-1">{{ t('admin.taxRates.rate_hint') }}</p>
            <p v-if="form.rate > 0" class="tax-rate-preview">{{ formatPercent(form.rate) }}</p>
          </div>
          <div>
            <label class="admin-form-label" for="tax-name">{{ t('admin.taxRates.name') }}</label>
            <input id="tax-name" v-model="form.name" class="admin-form-control" />
          </div>
        </div>
        <label class="tax-rate-check">
          <input v-model="form.isActive" type="checkbox" class="admin-form-checkbox" />
          <span class="admin-form-check-label">{{ t('admin.taxRates.active') }}</span>
        </label>
        <p v-if="formError" class="admin-form-msg--error text-sm mt-3">{{ formError }}</p>
      </form>

      <template #footer>
        <button type="button" class="admin-btn-outline" @click="closeModal">{{ t('admin.taxRates.cancel') }}</button>
        <button type="submit" form="tax-rate-form" :disabled="saving" class="admin-btn-primary">
          {{ t('admin.taxRates.save') }}
        </button>
      </template>
    </AdminModal>
  </div>
</template>

<script setup lang="ts">
/** 税率 CRUD — /admin/tax-rates */
definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()

const columns = computed(() => [
  { key: 'country', label: t('admin.taxRates.col_country'), width: '100px' },
  { key: 'region', label: t('admin.taxRates.col_region'), width: '100px' },
  { key: 'name', label: t('admin.taxRates.col_name') },
  { key: 'rate', label: t('admin.taxRates.col_rate'), width: '120px' },
  { key: 'isActive', label: t('admin.taxRates.col_status'), width: '120px' },
  { key: 'actions', label: t('admin.taxRates.actions'), width: '160px' },
])

const allRates = ref<any[]>([])
const pending = ref(true)
const error = ref('')
const showModal = ref(false)
const editingId = ref<string | null>(null)
const saving = ref(false)
const formError = ref('')
const searchQuery = ref('')
const statusFilter = ref('')

const form = reactive({
  country: '', region: '', rate: 0, name: '', isActive: true,
})

const activeCount = computed(() => allRates.value.filter(r => r.isActive).length)

const filteredRates = computed(() => {
  let rows = allRates.value
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    rows = rows.filter(r =>
      (r.country || '').toLowerCase().includes(q)
      || (r.region || '').toLowerCase().includes(q)
      || (r.name || '').toLowerCase().includes(q),
    )
  }
  if (statusFilter.value === 'active') rows = rows.filter(r => r.isActive)
  if (statusFilter.value === 'inactive') rows = rows.filter(r => !r.isActive)
  return rows
})

const formatPercent = (rate: number) => `${((rate || 0) * 100).toFixed(2)}%`

const fetchRates = async () => {
  pending.value = true
  error.value = ''
  try {
    allRates.value = await api.get<any[]>('/admin/tax-rates') || []
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const openCreate = () => {
  editingId.value = null
  Object.assign(form, { country: '', region: '', rate: 0, name: '', isActive: true })
  formError.value = ''
  showModal.value = true
}

const openEdit = (row: any) => {
  editingId.value = String(row.id)
  Object.assign(form, {
    country: row.country || '',
    region: row.region || '',
    rate: row.rate ?? 0,
    name: row.name || '',
    isActive: row.isActive ?? true,
  })
  formError.value = ''
  showModal.value = true
}

const closeModal = () => { showModal.value = false; saving.value = false }

const saveRate = async () => {
  saving.value = true
  formError.value = ''
  try {
    const payload = {
      country: form.country.trim().toUpperCase(),
      region: form.region.trim().toUpperCase(),
      rate: form.rate,
      name: form.name.trim(),
      isActive: form.isActive,
    }
    if (editingId.value) {
      await api.put(`/admin/tax-rates/${editingId.value}`, payload)
    } else {
      await api.post('/admin/tax-rates', payload)
    }
    closeModal()
    await fetchRates()
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.save_failed')
  } finally {
    saving.value = false
  }
}

const removeRate = async (id: string) => {
  if (!confirm(t('admin.confirm_delete'))) return
  try {
    await api.delete(`/admin/tax-rates/${id}`)
    await fetchRates()
  } catch (err: any) {
    alert(err?.message || t('errors.api.delete_failed'))
  }
}

onMounted(fetchRates)
</script>

<style scoped>
.tax-rates-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
  padding: var(--spacing-md) var(--spacing-lg);
  background: var(--color-bg);
  border: 1px solid var(--color-border-light);
  border-radius: var(--radius-lg);
}

.tax-rates-toolbar__field {
  min-width: 140px;
}

.tax-rates-toolbar__field--grow {
  flex: 1;
  min-width: 200px;
}

.tax-rates-toolbar__stats {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-inline-start: auto;
}

.tax-rates-stat {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  min-width: 4.5rem;
  padding: var(--spacing-xs) var(--spacing-md);
  border-radius: var(--radius-md);
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border-light);
}

.tax-rates-stat--active {
  background: rgba(var(--color-success-rgb), 0.08);
  border-color: rgba(var(--color-success-rgb), 0.2);
}

.tax-rates-stat__value {
  font-size: var(--text-lg);
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--color-text);
  line-height: 1.2;
}

.tax-rates-stat--active .tax-rates-stat__value {
  color: var(--color-success);
}

.tax-rates-stat__label {
  font-size: var(--text-xs);
  color: var(--color-text-lighter);
  margin-top: 2px;
}

.tax-rates-table {
  margin-top: 0;
}

.tax-rate-code {
  display: inline-flex;
  align-items: center;
  padding: 0.125rem 0.5rem;
  border-radius: var(--radius-sm);
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border-light);
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: var(--text-xs);
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--color-text);
}

.tax-rate-region {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-text-light);
}

.tax-rate-name {
  font-weight: 500;
  color: var(--color-text);
}

.tax-rate-value {
  font-variant-numeric: tabular-nums;
  font-weight: 700;
  font-size: var(--text-sm);
  color: var(--color-highlight-hover);
}

.tax-rate-muted {
  color: var(--color-text-lighter);
}

.tax-rate-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--spacing-xs);
  flex-wrap: wrap;
}

.tax-rate-form__grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-md);
}

@media (min-width: 640px) {
  .tax-rate-form__grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.tax-rate-check {
  display: inline-flex;
  align-items: center;
  margin-top: var(--spacing-md);
  cursor: pointer;
}

.tax-rate-preview {
  margin-top: var(--spacing-xs);
  font-size: var(--text-sm);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: var(--color-highlight);
}

@media (max-width: 640px) {
  .tax-rates-toolbar__stats {
    width: 100%;
    margin-inline-start: 0;
    justify-content: flex-start;
  }
}
</style>
