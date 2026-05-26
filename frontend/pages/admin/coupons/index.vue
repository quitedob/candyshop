<template>
  <div>
    <PageHeader :title="t('admin.coupons.title')" :description="t('admin.coupons.description')">
      <template #actions>
        <button type="button" class="btn btn-highlight" @click="openCreate">{{ t('admin.coupons.create') }}</button>
      </template>
    </PageHeader>
    <AdminTable class="mt-4" :columns="columns" :rows="coupons" :loading="pending" :empty-text="t('admin.coupons.no_data')" @retry="load">
      <template #cell-code="{ row }"><span class="font-mono">{{ row.code }}</span></template>
      <template #cell-type="{ row }">{{ formatType(row.type) }}</template>
      <template #cell-value="{ row }">{{ formatValue(row) }}</template>
      <template #cell-status="{ row }"><StatusBadge :status="row.status" type="product" /></template>
      <template #cell-actions="{ row }">
        <button type="button" class="text-red-600 text-sm" @click="remove(row.id)">{{ t('common.delete') }}</button>
      </template>
    </AdminTable>
    <AdminModal :open="showModal" :title="t('admin.coupons.create')" width="md" @close="showModal = false">
      <form class="space-y-3" @submit.prevent="save">
        <div>
          <label class="admin-form-label">{{ t('admin.coupons.code') }}</label>
          <input v-model="form.code" required class="form-input" />
        </div>
        <div>
          <label class="admin-form-label">{{ t('admin.coupons.type') }}</label>
          <select v-model="form.type" class="form-input">
            <option value="percentage">{{ t('admin.coupons.type_percentage') }}</option>
            <option value="fixed">{{ t('admin.coupons.type_fixed') }}</option>
          </select>
        </div>
        <div>
          <label class="admin-form-label">{{ t('admin.coupons.value') }}</label>
          <input v-model.number="form.value" type="number" min="0.01" step="0.01" required class="form-input" />
        </div>
        <div>
          <label class="admin-form-label">{{ t('admin.coupons.starts_at') }}</label>
          <input v-model="form.startsAt" type="datetime-local" required class="form-input" />
        </div>
        <div>
          <label class="admin-form-label">{{ t('admin.coupons.expires_at') }}</label>
          <input v-model="form.expiresAt" type="datetime-local" required class="form-input" />
        </div>
        <div>
          <label class="admin-form-label">{{ t('admin.coupons.min_order') }}</label>
          <input v-model.number="form.minOrderAmount" type="number" min="0" step="0.01" class="form-input" />
        </div>
        <p v-if="formError" class="text-sm text-red-600">{{ formError }}</p>
        <div class="flex justify-end gap-2 pt-2">
          <button type="button" class="btn btn-secondary" @click="showModal = false">{{ t('admin.coupons.cancel') }}</button>
          <button type="submit" class="btn btn-highlight" :disabled="saving">{{ saving ? t('common.saving') : t('common.save') }}</button>
        </div>
      </form>
    </AdminModal>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: ['auth'] })
const { t } = useI18n()
const api = useApi()
const columns = computed(() => [
  { key: 'code', label: t('admin.coupons.col_code') },
  { key: 'type', label: t('admin.coupons.col_type') },
  { key: 'value', label: t('admin.coupons.col_value') },
  { key: 'status', label: t('admin.coupons.col_status') },
  { key: 'actions', label: '' },
])
const coupons = ref<any[]>([])
const pending = ref(true)
const showModal = ref(false)
const saving = ref(false)
const formError = ref('')
const form = reactive({
  code: '',
  type: 'percentage',
  value: 10,
  startsAt: '',
  expiresAt: '',
  minOrderAmount: 0,
})

function defaultStartsAt() {
  const d = new Date()
  d.setMinutes(0, 0, 0)
  return toDatetimeLocal(d)
}

function defaultExpiresAt() {
  const d = new Date()
  d.setFullYear(d.getFullYear() + 1)
  d.setMinutes(0, 0, 0)
  return toDatetimeLocal(d)
}

function toDatetimeLocal(d: Date) {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function toRFC3339(local: string) {
  return new Date(local).toISOString()
}

const formatType = (type: string) => {
  if (type === 'percentage') return t('admin.coupons.type_percentage')
  if (type === 'fixed') return t('admin.coupons.type_fixed')
  return type
}

const formatValue = (row: any) => {
  if (row.type === 'percentage') return `${row.value}%`
  return String(row.value)
}

const load = async () => {
  pending.value = true
  try {
    const res = await api.get<any>('/admin/coupons')
    coupons.value = res.data || res || []
  } finally { pending.value = false }
}

const openCreate = () => {
  form.code = ''
  form.type = 'percentage'
  form.value = 10
  form.startsAt = defaultStartsAt()
  form.expiresAt = defaultExpiresAt()
  form.minOrderAmount = 0
  formError.value = ''
  showModal.value = true
}

const save = async () => {
  if (form.value <= 0) {
    formError.value = t('errors.validation.invalid_request')
    return
  }
  saving.value = true
  formError.value = ''
  try {
    await api.post('/admin/coupons', {
      code: form.code.trim(),
      type: form.type,
      value: form.value,
      startsAt: toRFC3339(form.startsAt),
      expiresAt: toRFC3339(form.expiresAt),
      minOrderAmount: form.minOrderAmount || 0,
    })
    showModal.value = false
    await load()
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.request_failed')
  } finally {
    saving.value = false
  }
}

const remove = async (id: string) => {
  if (!confirm(t('admin.confirm_delete'))) return
  await api.del(`/admin/coupons/${id}`)
  await load()
}

onMounted(load)
</script>
