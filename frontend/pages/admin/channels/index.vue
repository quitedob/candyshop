<template>
  <div>
    <PageHeader :title="t('admin.channels.title')" :description="t('admin.channels.description')">
      <template #actions>
        <button type="button" class="inline-flex items-center rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700" @click="openCreate">
          {{ t('admin.channels.create') }}
        </button>
      </template>
    </PageHeader>

    <AdminTable
      :columns="columns"
      :rows="channels"
      :loading="pending"
      :error="!!error"
      :error-message="error"
      :empty-text="t('admin.channels.no_data')"
      @retry="fetchChannels"
    >
      <template #cell-code="{ row }">
        <span class="font-mono text-sm">{{ row.code }}</span>
      </template>
      <template #cell-status="{ row }">
        <StatusBadge :status="row.status" type="product" />
      </template>
      <template #cell-actions="{ row }">
        <button type="button" class="text-orange-600 hover:text-orange-900" @click="openEdit(row)">{{ t('admin.channels.edit') }}</button>
        <button type="button" class="ml-3 text-red-600 hover:text-red-900" @click="removeChannel(row.id)">{{ t('admin.channels.delete') }}</button>
      </template>
    </AdminTable>

    <AdminModal :open="showModal" :title="editingId ? t('admin.channels.edit') : t('admin.channels.create')" width="lg" @close="closeModal">
      <form class="space-y-4" @submit.prevent="saveChannel">
        <div v-if="!editingId" class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.channels.code') }}</label>
            <input v-model="form.code" required class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
          </div>
        </div>
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.channels.name') }}</label>
            <input v-model="form.name" required class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.channels.type') }}</label>
            <select v-model="form.type" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
              <option value="marketplace">{{ t('admin.channels.type_marketplace') }}</option>
              <option value="webstore">{{ t('admin.channels.type_webstore') }}</option>
              <option value="distributor">{{ t('admin.channels.type_distributor') }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.channels.currency') }}</label>
            <input v-model="form.currency" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm uppercase" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.channels.price_multiplier') }}</label>
            <input v-model.number="form.priceMultiplier" type="number" step="0.01" min="0" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.channels.fulfillment') }}</label>
            <select v-model="form.fulfillmentMode" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
              <option value="self">{{ t('admin.channels.fulfillment_self') }}</option>
              <option value="3pl">{{ t('admin.channels.fulfillment_3pl') }}</option>
            </select>
          </div>
          <div v-if="editingId">
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.channels.status') }}</label>
            <select v-model="form.status" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
              <option value="active">{{ t('admin.status_options.active') }}</option>
              <option value="inactive">{{ t('admin.status_options.inactive') }}</option>
            </select>
          </div>
        </div>
        <p v-if="formError" class="text-sm text-red-600">{{ formError }}</p>
        <div class="flex justify-end gap-3 pt-2">
          <button type="button" class="rounded-md border border-gray-300 px-4 py-2 text-sm" @click="closeModal">{{ t('admin.channels.cancel') }}</button>
          <button type="submit" :disabled="saving" class="rounded-md bg-orange-600 px-4 py-2 text-sm text-white disabled:opacity-50">{{ t('admin.channels.save') }}</button>
        </div>
      </form>
    </AdminModal>
  </div>
</template>

<script setup lang="ts">
/** 销售渠道 CRUD — /admin/channels */
definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()

const columns = computed(() => [
  { key: 'code', label: t('admin.channels.col_code') },
  { key: 'name', label: t('admin.channels.col_name') },
  { key: 'type', label: t('admin.channels.col_type') },
  { key: 'currency', label: t('admin.channels.col_currency') },
  { key: 'status', label: t('admin.channels.col_status') },
  { key: 'actions', label: t('admin.channels.actions') },
])

const channels = ref<any[]>([])
const pending = ref(true)
const error = ref('')
const showModal = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const formError = ref('')

const form = reactive({
  code: '', name: '', type: 'marketplace', currency: 'USD',
  priceMultiplier: 1, fulfillmentMode: 'self', status: 'active',
})

const fetchChannels = async () => {
  pending.value = true
  error.value = ''
  try {
    channels.value = (await api.get<any[]>('/admin/channels')) || []
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const openCreate = () => {
  editingId.value = null
  Object.assign(form, { code: '', name: '', type: 'marketplace', currency: 'USD', priceMultiplier: 1, fulfillmentMode: 'self', status: 'active' })
  formError.value = ''
  showModal.value = true
}

const openEdit = (row: any) => {
  editingId.value = row.id
  Object.assign(form, {
    code: row.code, name: row.name, type: row.type || 'marketplace',
    currency: row.currency || 'USD', priceMultiplier: row.priceMultiplier ?? 1,
    fulfillmentMode: row.fulfillmentMode || 'self', status: row.status || 'active',
  })
  formError.value = ''
  showModal.value = true
}

const closeModal = () => { showModal.value = false; saving.value = false }

const saveChannel = async () => {
  saving.value = true
  formError.value = ''
  try {
    if (editingId.value) {
      await api.put(`/admin/channels/${editingId.value}`, {
        name: form.name, type: form.type, currency: form.currency,
        priceMultiplier: form.priceMultiplier, fulfillmentMode: form.fulfillmentMode, status: form.status,
      })
    } else {
      await api.post('/admin/channels', { ...form })
    }
    closeModal()
    await fetchChannels()
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.save_failed')
  } finally {
    saving.value = false
  }
}

const removeChannel = async (id: number) => {
  if (!confirm(t('admin.confirm_delete'))) return
  try {
    await api.del(`/admin/channels/${id}`)
    await fetchChannels()
  } catch (err: any) {
    notifyError(err, t('errors.api.delete_failed'))
  }
}

onMounted(fetchChannels)
</script>
