<template>
  <div>
    <PageHeader :title="t('admin.shippingRates.title')" :description="t('admin.shippingRates.description')">
      <template #actions>
        <button type="button" class="btn btn-highlight inline-flex items-center gap-2" @click="openCreate">
          <Icon name="material-symbols:add" size="18" aria-hidden="true" />
          {{ t('admin.shippingRates.create') }}
        </button>
      </template>
    </PageHeader>

    <div class="shipping-rates-toolbar">
      <div class="shipping-rates-toolbar__field shipping-rates-toolbar__field--grow">
        <label class="admin-form-label admin-form-label--xs" for="shipping-rates-search">{{ t('admin.products.search') }}</label>
        <input
          id="shipping-rates-search"
          v-model="searchQuery"
          type="search"
          autocomplete="off"
          class="admin-form-control admin-form-control--flat"
          :placeholder="t('admin.shippingRates.col_destination')"
        />
      </div>
      <div class="shipping-rates-toolbar__field">
        <label class="admin-form-label admin-form-label--xs" for="shipping-rates-status">{{ t('admin.shippingRates.col_status') }}</label>
        <select id="shipping-rates-status" v-model="statusFilter" class="admin-form-control admin-form-control--flat">
          <option value="">{{ t('admin.products.status_all') }}</option>
          <option value="active">{{ t('admin.status_options.active') }}</option>
          <option value="inactive">{{ t('admin.status_options.inactive') }}</option>
        </select>
      </div>
      <div class="shipping-rates-toolbar__stats">
        <span class="shipping-rates-stat">
          <span class="shipping-rates-stat__value">{{ allRates.length }}</span>
          <span class="shipping-rates-stat__label">{{ t('admin.shippingRates.col_destination') }}</span>
        </span>
        <span class="shipping-rates-stat shipping-rates-stat--active">
          <span class="shipping-rates-stat__value">{{ activeCount }}</span>
          <span class="shipping-rates-stat__label">{{ t('admin.status_options.active') }}</span>
        </span>
      </div>
    </div>

    <AdminTable
      class="shipping-rates-table"
      :columns="columns"
      :rows="filteredRates"
      :loading="pending"
      :error="!!error"
      :error-message="error"
      :empty-text="t('admin.shippingRates.no_data')"
      @retry="fetchRates"
    >
      <template #empty-actions>
        <button type="button" class="btn btn-highlight" @click="openCreate">{{ t('admin.shippingRates.create') }}</button>
      </template>
      <template #cell-destination="{ row }">
        <div class="shipping-destination">
          <span class="shipping-destination__name">{{ row.destination || '—' }}</span>
          <span class="shipping-destination__weight">{{ formatWeightRange(row) }}</span>
        </div>
      </template>
      <template #cell-carrier="{ row }">
        <span v-if="row.carrier" class="shipping-carrier">{{ row.carrier }}</span>
        <span v-else class="shipping-muted">—</span>
      </template>
      <template #cell-baseCost="{ row }">
        <span class="shipping-cost">{{ cur(row.currency || 'USD') }} {{ formatNumber(row.baseCost || 0) }}</span>
        <span v-if="row.costPerKg" class="shipping-cost-per-kg">
          + {{ cur(row.currency || 'USD') }} {{ formatNumber(row.costPerKg || 0) }}/kg
        </span>
      </template>
      <template #cell-estimatedDays="{ row }">
        <span class="shipping-days">{{ row.estimatedDays ?? '—' }}</span>
      </template>
      <template #cell-isActive="{ row }">
        <StatusBadge :status="row.isActive ? 'active' : 'inactive'" type="product" />
      </template>
      <template #cell-actions="{ row }">
        <div class="shipping-rate-actions">
          <button type="button" class="admin-btn-ghost" @click="openEdit(row)">{{ t('admin.shippingRates.edit') }}</button>
          <button type="button" class="admin-btn-danger-ghost" @click="removeRate(row.id)">{{ t('admin.shippingRates.delete') }}</button>
        </div>
      </template>
    </AdminTable>

    <AdminModal
      :open="showModal"
      :title="editingId ? t('admin.shippingRates.edit') : t('admin.shippingRates.create')"
      width="lg"
      @close="closeModal"
    >
      <form id="shipping-rate-form" class="shipping-rate-form" @submit.prevent="saveRate">
        <div class="shipping-rate-form__grid">
          <div class="shipping-rate-form__full">
            <label class="admin-form-label" for="shipping-destination">{{ t('admin.shippingRates.destination') }}</label>
            <input id="shipping-destination" v-model="form.destination" required class="admin-form-control" />
          </div>
          <div>
            <label class="admin-form-label" for="shipping-carrier">{{ t('admin.shippingRates.carrier') }}</label>
            <input id="shipping-carrier" v-model="form.carrier" class="admin-form-control" />
          </div>
          <div>
            <label class="admin-form-label" for="shipping-currency">{{ t('admin.shippingRates.currency') }}</label>
            <input id="shipping-currency" v-model="form.currency" class="admin-form-control uppercase" maxlength="3" />
          </div>
          <div>
            <label class="admin-form-label" for="shipping-min-weight">{{ t('admin.shippingRates.min_weight') }}</label>
            <input id="shipping-min-weight" v-model.number="form.minWeightKg" type="number" step="0.1" min="0" class="admin-form-control" />
          </div>
          <div>
            <label class="admin-form-label" for="shipping-max-weight">{{ t('admin.shippingRates.max_weight') }}</label>
            <input id="shipping-max-weight" v-model.number="form.maxWeightKg" type="number" step="0.1" min="0" class="admin-form-control" />
          </div>
          <div>
            <label class="admin-form-label" for="shipping-base-cost">{{ t('admin.shippingRates.base_cost') }}</label>
            <input id="shipping-base-cost" v-model.number="form.baseCost" type="number" step="0.01" min="0" class="admin-form-control" />
          </div>
          <div>
            <label class="admin-form-label" for="shipping-cost-per-kg">{{ t('admin.shippingRates.cost_per_kg') }}</label>
            <input id="shipping-cost-per-kg" v-model.number="form.costPerKg" type="number" step="0.01" min="0" class="admin-form-control" />
          </div>
          <div>
            <label class="admin-form-label" for="shipping-days">{{ t('admin.shippingRates.estimated_days') }}</label>
            <input id="shipping-days" v-model.number="form.estimatedDays" type="number" min="1" class="admin-form-control" />
            <p class="admin-form-hint mt-1">{{ t('admin.shippingRates.estimated_days_hint') }}</p>
          </div>
        </div>
        <label class="shipping-rate-check">
          <input v-model="form.isActive" type="checkbox" class="admin-form-checkbox" />
          <span class="admin-form-check-label">{{ t('admin.shippingRates.active') }}</span>
        </label>
        <p v-if="formError" class="admin-form-msg--error text-sm mt-3">{{ formError }}</p>
      </form>

      <template #footer>
        <button type="button" class="admin-btn-outline" @click="closeModal">{{ t('admin.shippingRates.cancel') }}</button>
        <button type="submit" form="shipping-rate-form" :disabled="saving" class="admin-btn-primary">
          {{ t('admin.shippingRates.save') }}
        </button>
      </template>
    </AdminModal>
  </div>
</template>

<script setup lang="ts">
/** 运费费率 CRUD — /admin/shipping-rates */
definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const { formatNumber, currencyOrDefault: cur } = useDisplay()

const columns = computed(() => [
  { key: 'destination', label: t('admin.shippingRates.col_destination') },
  { key: 'carrier', label: t('admin.shippingRates.col_carrier'), width: '140px' },
  { key: 'baseCost', label: t('admin.shippingRates.col_base_cost'), width: '160px' },
  { key: 'estimatedDays', label: t('admin.shippingRates.col_days'), width: '100px' },
  { key: 'isActive', label: t('admin.shippingRates.col_status'), width: '120px' },
  { key: 'actions', label: t('admin.shippingRates.actions'), width: '160px' },
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

// R2 B-2: form fields use camelCase to match backend DTO and model output.
// Previously the form was snake_case (matching the old request DTO) while the
// list response was camelCase, forcing every cell template to handle both.
// Now everything is camelCase end-to-end.
const form = reactive({
  destination: '', carrier: '', minWeightKg: 0, maxWeightKg: 0,
  baseCost: 0, costPerKg: 0, currency: 'USD', estimatedDays: 7, isActive: true,
})

const activeCount = computed(() => allRates.value.filter(r => r.isActive).length)

const filteredRates = computed(() => {
  let rows = allRates.value
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    rows = rows.filter(r =>
      (r.destination || '').toLowerCase().includes(q)
      || (r.carrier || '').toLowerCase().includes(q),
    )
  }
  if (statusFilter.value === 'active') rows = rows.filter(r => r.isActive)
  if (statusFilter.value === 'inactive') rows = rows.filter(r => !r.isActive)
  return rows
})

const formatWeightRange = (row: any) => {
  const min = row.minWeightKg ?? 0
  const max = row.maxWeightKg ?? 0
  if (max > 0) return `${min}–${max} kg`
  if (min > 0) return `≥ ${min} kg`
  return ''
}

const fetchRates = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await api.get<any>('/admin/shipping-rates')
    allRates.value = Array.isArray(res) ? res : (res?.data || [])
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const openCreate = () => {
  editingId.value = null
  Object.assign(form, {
    destination: '', carrier: '', minWeightKg: 0, maxWeightKg: 100,
    baseCost: 0, costPerKg: 0, currency: 'USD', estimatedDays: 7, isActive: true,
  })
  formError.value = ''
  showModal.value = true
}

const openEdit = (row: any) => {
  editingId.value = String(row.id)
  Object.assign(form, {
    destination: row.destination, carrier: row.carrier || '',
    minWeightKg: row.minWeightKg ?? 0,
    maxWeightKg: row.maxWeightKg ?? 0,
    baseCost: row.baseCost ?? 0,
    costPerKg: row.costPerKg ?? 0,
    currency: row.currency || 'USD',
    estimatedDays: row.estimatedDays ?? 7,
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
      ...form,
      currency: (form.currency || 'USD').trim().toUpperCase(),
      destination: form.destination.trim(),
      carrier: form.carrier.trim(),
    }
    if (editingId.value) {
      await api.put(`/admin/shipping-rates/${editingId.value}`, payload)
    } else {
      await api.post('/admin/shipping-rates', payload)
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
    await api.del(`/admin/shipping-rates/${id}`)
    await fetchRates()
  } catch (err: any) {
    notifyError(err, t('errors.api.delete_failed'))
  }
}

onMounted(fetchRates)
</script>

<style scoped>
.shipping-rates-toolbar {
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

.shipping-rates-toolbar__field {
  min-width: 140px;
}

.shipping-rates-toolbar__field--grow {
  flex: 1;
  min-width: 200px;
}

.shipping-rates-toolbar__stats {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-inline-start: auto;
}

.shipping-rates-stat {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  min-width: 4.5rem;
  padding: var(--spacing-xs) var(--spacing-md);
  border-radius: var(--radius-md);
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border-light);
}

.shipping-rates-stat--active {
  background: rgba(var(--color-success-rgb), 0.08);
  border-color: rgba(var(--color-success-rgb), 0.2);
}

.shipping-rates-stat__value {
  font-size: var(--text-lg);
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--color-text);
  line-height: 1.2;
}

.shipping-rates-stat--active .shipping-rates-stat__value {
  color: var(--color-success);
}

.shipping-rates-stat__label {
  font-size: var(--text-xs);
  color: var(--color-text-lighter);
  margin-top: 2px;
}

.shipping-rates-table {
  margin-top: 0;
}

.shipping-destination {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.shipping-destination__name {
  font-weight: 600;
  color: var(--color-text);
}

.shipping-destination__weight {
  font-size: var(--text-xs);
  color: var(--color-text-lighter);
  font-variant-numeric: tabular-nums;
}

.shipping-carrier {
  display: inline-flex;
  padding: 0.125rem 0.5rem;
  border-radius: var(--radius-sm);
  background: var(--color-bg-alt);
  border: 1px solid var(--color-border-light);
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-text-light);
}

.shipping-cost {
  display: block;
  font-variant-numeric: tabular-nums;
  font-weight: 700;
  color: var(--color-highlight-hover);
}

.shipping-cost-per-kg {
  display: block;
  margin-top: 2px;
  font-size: var(--text-xs);
  color: var(--color-text-lighter);
  font-variant-numeric: tabular-nums;
}

.shipping-days {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 2rem;
  padding: 0.125rem 0.5rem;
  border-radius: var(--radius-sm);
  background: rgba(var(--color-info-rgb), 0.1);
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  font-size: var(--text-sm);
  color: var(--color-info-hover);
}

.shipping-muted {
  color: var(--color-text-lighter);
}

.shipping-rate-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--spacing-xs);
  flex-wrap: wrap;
}

.shipping-rate-form__grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-md);
}

.shipping-rate-form__full {
  grid-column: 1 / -1;
}

@media (min-width: 640px) {
  .shipping-rate-form__grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.shipping-rate-check {
  display: inline-flex;
  align-items: center;
  margin-top: var(--spacing-md);
  cursor: pointer;
}

@media (max-width: 640px) {
  .shipping-rates-toolbar__stats {
    width: 100%;
    margin-inline-start: 0;
    justify-content: flex-start;
  }
}
</style>
