<template>
  <div>
    <PageHeader :title="t('admin.webhooks.title')" :description="t('admin.webhooks.description')">
      <template #actions>
        <button type="button" class="inline-flex items-center rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700" @click="openCreate">
          {{ t('admin.webhooks.create') }}
        </button>
      </template>
    </PageHeader>

    <div class="mb-4 border-b border-gray-200">
      <nav class="-mb-px flex gap-6">
        <button type="button" :class="tabClass('configs')" @click="activeTab = 'configs'">{{ t('admin.webhooks.tab_configs') }}</button>
        <button type="button" :class="tabClass('deliveries')" @click="activeTab = 'deliveries'; fetchDeliveries()">{{ t('admin.webhooks.tab_deliveries') }}</button>
      </nav>
    </div>

    <AdminTable
      v-if="activeTab === 'configs'"
      :columns="configColumns"
      :rows="webhooks"
      :loading="pending"
      :error="!!error"
      :error-message="error"
      :empty-text="t('admin.webhooks.no_data')"
      @retry="fetchWebhooks"
    >
      <template #cell-events="{ row }">
        <span class="text-xs text-gray-600">{{ (row.events || []).join(', ') }}</span>
      </template>
      <template #cell-status="{ row }">
        <StatusBadge :status="row.status" type="product" />
      </template>
      <template #cell-actions="{ row }">
        <button type="button" class="text-orange-600 hover:text-orange-900" @click="openEdit(row)">{{ t('admin.webhooks.edit') }}</button>
        <button type="button" class="ml-3 text-red-600 hover:text-red-900" @click="removeWebhook(row.id)">{{ t('admin.webhooks.delete') }}</button>
      </template>
    </AdminTable>

    <AdminTable
      v-else
      :columns="deliveryColumns"
      :rows="deliveries"
      :loading="deliveriesPending"
      :error="!!deliveriesError"
      :error-message="deliveriesError"
      :empty-text="t('admin.webhooks.no_deliveries')"
      @retry="fetchDeliveries"
    >
      <template #cell-success="{ row }">
        <span :class="row.success ? 'text-emerald-600' : 'text-red-600'">{{ row.success ? t('admin.webhooks.success') : t('admin.webhooks.failed') }}</span>
      </template>
    </AdminTable>

    <AdminModal :open="showModal" :title="editingId ? t('admin.webhooks.edit') : t('admin.webhooks.create')" width="lg" @close="closeModal">
      <form class="space-y-4" @submit.prevent="saveWebhook">
        <div>
          <label class="block text-sm font-medium text-gray-700">{{ t('admin.webhooks.name') }}</label>
          <input v-model="form.name" required class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700">{{ t('admin.webhooks.url') }}</label>
          <input v-model="form.url" type="url" required class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700">{{ t('admin.webhooks.events') }}</label>
          <div class="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-2">
            <label v-for="ev in eventOptions" :key="ev" class="inline-flex items-center gap-2 text-sm">
              <input v-model="form.events" type="checkbox" :value="ev" class="rounded border-gray-300" />
              {{ ev }}
            </label>
          </div>
        </div>
        <div v-if="editingId">
          <label class="block text-sm font-medium text-gray-700">{{ t('admin.webhooks.status') }}</label>
          <select v-model="form.status" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
            <option value="active">{{ t('admin.status_options.active') }}</option>
            <option value="inactive">{{ t('admin.status_options.inactive') }}</option>
          </select>
        </div>
        <p v-if="createdSecret" class="text-sm text-emerald-700 bg-emerald-50 p-3 rounded-md">{{ t('admin.webhooks.secret_created') }}: <code>{{ createdSecret }}</code></p>
        <p v-if="formError" class="text-sm text-red-600">{{ formError }}</p>
        <div class="flex justify-end gap-3 pt-2">
          <button type="button" class="rounded-md border border-gray-300 px-4 py-2 text-sm" @click="closeModal">{{ t('admin.webhooks.cancel') }}</button>
          <button type="submit" :disabled="saving" class="rounded-md bg-orange-600 px-4 py-2 text-sm text-white disabled:opacity-50">{{ t('admin.webhooks.save') }}</button>
        </div>
      </form>
    </AdminModal>
  </div>
</template>

<script setup lang="ts">
/** Webhook 配置 CRUD + 投递日志 — /admin/webhooks */
definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()

const eventOptions = [
  'order.created', 'order.confirmed', 'order.shipped', 'order.delivered',
  'payment.confirmed', 'return.created',
]

const activeTab = ref<'configs' | 'deliveries'>('configs')
const webhooks = ref<any[]>([])
const deliveries = ref<any[]>([])
const pending = ref(true)
const deliveriesPending = ref(false)
const error = ref('')
const deliveriesError = ref('')
const showModal = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const formError = ref('')
const createdSecret = ref('')

const form = reactive({ name: '', url: '', events: [] as string[], status: 'active' })

const configColumns = computed(() => [
  { key: 'name', label: t('admin.webhooks.col_name') },
  { key: 'url', label: t('admin.webhooks.col_url') },
  { key: 'events', label: t('admin.webhooks.col_events') },
  { key: 'status', label: t('admin.webhooks.col_status') },
  { key: 'actions', label: t('admin.webhooks.actions') },
])

const deliveryColumns = computed(() => [
  { key: 'eventType', label: t('admin.webhooks.col_event') },
  { key: 'statusCode', label: t('admin.webhooks.col_status_code') },
  { key: 'success', label: t('admin.webhooks.col_result') },
  { key: 'createdAt', label: t('admin.webhooks.col_time') },
])

const tabClass = (tab: string) => [
  'whitespace-nowrap border-b-2 py-3 text-sm font-medium',
  activeTab.value === tab ? 'border-orange-500 text-orange-600' : 'border-transparent text-gray-500 hover:text-gray-700',
]

const fetchWebhooks = async () => {
  pending.value = true
  error.value = ''
  try {
    webhooks.value = (await api.get<any[]>('/admin/webhooks')) || []
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const fetchDeliveries = async () => {
  deliveriesPending.value = true
  deliveriesError.value = ''
  try {
    const res = await api.get<any>('/admin/webhook-deliveries', { page: 1, limit: 50 })
    deliveries.value = res.data || []
  } catch (err: any) {
    deliveriesError.value = err?.message || t('errors.api.load_failed')
  } finally {
    deliveriesPending.value = false
  }
}

const openCreate = () => {
  editingId.value = null
  createdSecret.value = ''
  Object.assign(form, { name: '', url: '', events: ['order.created'], status: 'active' })
  formError.value = ''
  showModal.value = true
}

const openEdit = (row: any) => {
  editingId.value = row.id
  createdSecret.value = ''
  Object.assign(form, { name: row.name, url: row.url, events: [...(row.events || [])], status: row.status || 'active' })
  formError.value = ''
  showModal.value = true
}

const closeModal = () => { showModal.value = false; saving.value = false }

const saveWebhook = async () => {
  if (form.events.length === 0) {
    formError.value = t('admin.webhooks.events_required')
    return
  }
  saving.value = true
  formError.value = ''
  try {
    if (editingId.value) {
      await api.put(`/admin/webhooks/${editingId.value}`, { name: form.name, url: form.url, events: form.events, status: form.status })
    } else {
      const res = await api.post<any>('/admin/webhooks', { name: form.name, url: form.url, events: form.events })
      createdSecret.value = res.secret || ''
    }
    if (!createdSecret.value) closeModal()
    await fetchWebhooks()
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.save_failed')
  } finally {
    saving.value = false
  }
}

const removeWebhook = async (id: number) => {
  if (!confirm(t('admin.confirm_delete'))) return
  try {
    await api.delete(`/admin/webhooks/${id}`)
    await fetchWebhooks()
  } catch (err: any) {
    alert(err?.message || t('errors.api.delete_failed'))
  }
}

onMounted(fetchWebhooks)
</script>
