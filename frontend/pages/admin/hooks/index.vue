<template>
  <div>
    <PageHeader :title="t('admin.hooks.title')" :description="t('admin.hooks.description')">
      <template #actions>
        <button type="button" class="inline-flex items-center rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700" @click="openCreate">
          {{ t('admin.hooks.create') }}
        </button>
      </template>
    </PageHeader>

    <div class="mb-4 border-b border-gray-200">
      <nav class="-mb-px flex gap-6">
        <button type="button" :class="tabClass('hooks')" @click="activeTab = 'hooks'">{{ t('admin.hooks.tab_hooks') }}</button>
        <button type="button" :class="tabClass('events')" @click="activeTab = 'events'; fetchEvents()">{{ t('admin.hooks.tab_events') }}</button>
        <button type="button" :class="tabClass('executions')" @click="activeTab = 'executions'; fetchExecutions()">{{ t('admin.hooks.tab_executions') }}</button>
      </nav>
    </div>

    <AdminTable
      v-if="activeTab === 'hooks'"
      :columns="hookColumns"
      :rows="hooks"
      :loading="pending"
      :error="!!error"
      :error-message="error"
      :empty-text="t('admin.hooks.no_data')"
      @retry="fetchHooks"
    >
      <template #cell-status="{ row }">
        <StatusBadge :status="row.status" type="product" />
      </template>
      <template #cell-actions="{ row }">
        <button type="button" class="text-orange-600 hover:text-orange-900" @click="openEdit(row)">{{ t('admin.hooks.edit') }}</button>
        <button type="button" class="ml-3 text-red-600 hover:text-red-900" @click="removeHook(row.id)">{{ t('admin.hooks.delete') }}</button>
      </template>
    </AdminTable>

    <AdminTable
      v-else-if="activeTab === 'events'"
      :columns="eventColumns"
      :rows="events"
      :loading="eventsPending"
      :error="!!eventsError"
      :error-message="eventsError"
      :empty-text="t('admin.hooks.no_events')"
      @retry="fetchEvents"
    />

    <div v-else class="space-y-4">
      <div class="flex items-end gap-3">
        <div>
          <label class="block text-xs font-medium text-gray-600">{{ t('admin.hooks.hook_id') }}</label>
          <input v-model="executionHookId" type="number" min="1" class="mt-1 rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <button type="button" class="rounded-md bg-gray-100 px-4 py-2 text-sm hover:bg-gray-200" @click="fetchExecutions">{{ t('admin.hooks.load') }}</button>
      </div>
      <AdminTable
        :columns="executionColumns"
        :rows="executions"
        :loading="executionsPending"
        :error="!!executionsError"
        :error-message="executionsError"
        :empty-text="t('admin.hooks.no_executions')"
        @retry="fetchExecutions"
      />
    </div>

    <AdminModal :open="showModal" :title="editingId ? t('admin.hooks.edit') : t('admin.hooks.create')" width="lg" @close="closeModal">
      <form class="space-y-4" @submit.prevent="saveHook">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.hooks.name') }}</label>
            <input v-model="form.name" required class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.hooks.event_name') }}</label>
            <input v-model="form.eventName" required class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.hooks.type') }}</label>
            <select v-model="form.type" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
              <option value="webhook">webhook</option>
              <option value="plugin">plugin</option>
              <option value="internal">internal</option>
            </select>
          </div>
          <div v-if="editingId">
            <label class="block text-sm font-medium text-gray-700">{{ t('admin.hooks.status') }}</label>
            <select v-model="form.status" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
              <option value="active">{{ t('admin.status_options.active') }}</option>
              <option value="inactive">{{ t('admin.status_options.inactive') }}</option>
            </select>
          </div>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700">{{ t('admin.hooks.config_json') }}</label>
          <textarea v-model="form.configJson" rows="4" class="mt-1 w-full rounded-md border border-gray-300 px-3 py-2 font-mono text-xs" />
        </div>
        <p v-if="formError" class="text-sm text-red-600">{{ formError }}</p>
        <div class="flex justify-end gap-3 pt-2">
          <button type="button" class="rounded-md border border-gray-300 px-4 py-2 text-sm" @click="closeModal">{{ t('admin.hooks.cancel') }}</button>
          <button type="submit" :disabled="saving" class="rounded-md bg-orange-600 px-4 py-2 text-sm text-white disabled:opacity-50">{{ t('admin.hooks.save') }}</button>
        </div>
      </form>
    </AdminModal>
  </div>
</template>

<script setup lang="ts">
/** 插件 Hook CRUD + 事件/执行日志 — /admin/hooks */
definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()

const activeTab = ref<'hooks' | 'events' | 'executions'>('hooks')
const hooks = ref<any[]>([])
const events = ref<any[]>([])
const executions = ref<any[]>([])
const pending = ref(true)
const eventsPending = ref(false)
const executionsPending = ref(false)
const error = ref('')
const eventsError = ref('')
const executionsError = ref('')
const showModal = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const formError = ref('')
const executionHookId = ref('')

const form = reactive({ name: '', eventName: '', type: 'webhook', status: 'active', configJson: '{}' })

const hookColumns = computed(() => [
  { key: 'name', label: t('admin.hooks.col_name') },
  { key: 'eventName', label: t('admin.hooks.col_event') },
  { key: 'type', label: t('admin.hooks.col_type') },
  { key: 'status', label: t('admin.hooks.col_status') },
  { key: 'actions', label: t('admin.hooks.actions') },
])

const eventColumns = computed(() => [
  { key: 'name', label: t('admin.hooks.col_event') },
  { key: 'payload', label: t('admin.hooks.col_payload') },
  { key: 'createdAt', label: t('admin.hooks.col_time') },
])

const executionColumns = computed(() => [
  { key: 'hookId', label: t('admin.hooks.hook_id') },
  { key: 'status', label: t('admin.hooks.col_status') },
  { key: 'error', label: t('admin.hooks.col_error') },
  { key: 'createdAt', label: t('admin.hooks.col_time') },
])

const tabClass = (tab: string) => [
  'whitespace-nowrap border-b-2 py-3 text-sm font-medium',
  activeTab.value === tab ? 'border-orange-500 text-orange-600' : 'border-transparent text-gray-500 hover:text-gray-700',
]

const fetchHooks = async () => {
  pending.value = true
  error.value = ''
  try {
    hooks.value = (await api.get<any[]>('/admin/hooks')) || []
  } catch (err: any) {
    error.value = err?.message || t('errors.api.load_failed')
  } finally {
    pending.value = false
  }
}

const fetchEvents = async () => {
  eventsPending.value = true
  eventsError.value = ''
  try {
    const res = await api.get<any>('/admin/events', { page: 1, limit: 50 })
    events.value = (res.data || []).map((e: any) => ({
      ...e,
      payload: typeof e.payload === 'object' ? JSON.stringify(e.payload).slice(0, 80) : String(e.payload || '').slice(0, 80),
    }))
  } catch (err: any) {
    eventsError.value = err?.message || t('errors.api.load_failed')
  } finally {
    eventsPending.value = false
  }
}

const fetchExecutions = async () => {
  if (!executionHookId.value) return
  executionsPending.value = true
  executionsError.value = ''
  try {
    const res = await api.get<any>('/admin/hooks/executions', { hookId: executionHookId.value, page: 1, limit: 50 })
    executions.value = res.data || []
  } catch (err: any) {
    executionsError.value = err?.message || t('errors.api.load_failed')
  } finally {
    executionsPending.value = false
  }
}

const openCreate = () => {
  editingId.value = null
  Object.assign(form, { name: '', eventName: '', type: 'webhook', status: 'active', configJson: '{}' })
  formError.value = ''
  showModal.value = true
}

const openEdit = (row: any) => {
  editingId.value = row.id
  Object.assign(form, {
    name: row.name, eventName: row.eventName, type: row.type,
    status: row.status || 'active', configJson: JSON.stringify(row.config || {}, null, 2),
  })
  formError.value = ''
  showModal.value = true
}

const closeModal = () => { showModal.value = false; saving.value = false }

const saveHook = async () => {
  saving.value = true
  formError.value = ''
  let config: Record<string, unknown> = {}
  try {
    config = JSON.parse(form.configJson || '{}')
  } catch {
    formError.value = t('admin.hooks.invalid_json')
    saving.value = false
    return
  }
  try {
    const payload = { name: form.name, eventName: form.eventName, type: form.type, config }
    if (editingId.value) {
      await api.put(`/admin/hooks/${editingId.value}`, { ...payload, status: form.status })
    } else {
      await api.post('/admin/hooks', payload)
    }
    closeModal()
    await fetchHooks()
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.save_failed')
  } finally {
    saving.value = false
  }
}

const removeHook = async (id: number) => {
  if (!confirm(t('admin.confirm_delete'))) return
  try {
    await api.delete(`/admin/hooks/${id}`)
    await fetchHooks()
  } catch (err: any) {
    alert(err?.message || t('errors.api.delete_failed'))
  }
}

onMounted(fetchHooks)
</script>
