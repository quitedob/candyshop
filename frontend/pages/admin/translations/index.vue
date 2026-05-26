<template>
  <div>
    <PageHeader :title="t('admin.translations.title')">
      <template #actions>
        <button
          type="button"
          class="px-4 py-2 text-sm bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
          @click="exportAll"
        >
          {{ t('admin.translations.export') }}
        </button>
        <button
          type="button"
          class="px-4 py-2 text-sm bg-orange-600 text-white rounded-lg hover:bg-orange-700"
          @click="showCreateModal = true"
        >
          {{ t('admin.translations.add') }}
        </button>
      </template>
    </PageHeader>

    <!-- Filters -->
    <div class="flex gap-3 mb-4 flex-wrap">
      <input
        v-model="filterSearch"
        type="text"
        :placeholder="t('admin.translations.search_placeholder')"
        class="px-3 py-2 border rounded-lg text-sm"
        @keyup.enter="loadTranslations(1)"
      />
      <select
        v-model="filterGroup"
        class="px-3 py-2 border rounded-lg text-sm"
        @change="loadTranslations(1)"
      >
        <option value="">{{ t('admin.translations.all_groups') }}</option>
        <option value="errors">errors</option>
        <option value="validation">validation</option>
        <option value="content">content</option>
      </select>
      <select
        v-model="filterLocale"
        class="px-3 py-2 border rounded-lg text-sm"
        @change="loadTranslations(1)"
      >
        <option value="">{{ t('admin.translations.all_locales') }}</option>
        <option v-for="loc in locales" :key="loc.code" :value="loc.code">{{ loc.name || loc.code }}</option>
      </select>
      <button
        type="button"
        class="px-4 py-2 text-sm bg-orange-600 text-white rounded-lg hover:bg-orange-700"
        @click="loadTranslations(1)"
      >
        {{ t('admin.translations.filter') }}
      </button>
    </div>

    <!-- Error banner -->
    <div v-if="error" class="mb-4 p-3 bg-red-50 text-red-700 rounded-lg text-sm">
      {{ error }}
    </div>

    <AdminTable
      :columns="columns"
      :rows="translations"
      :loading="loading"
      :empty-text="t('admin.translations.no_data')"
    >
      <template #cell-key="{ row }">
        <span class="font-mono text-xs text-orange-600">{{ row.key }}</span>
      </template>

      <template #cell-locale="{ row }">
        <span class="px-2 py-0.5 rounded text-xs bg-gray-100">{{ row.locale }}</span>
      </template>

      <template #cell-isActive="{ row }">
        <span :class="row.isActive ? 'text-green-600' : 'text-red-500'" class="text-xs">
          {{ row.isActive ? t('admin.translations.active') : t('admin.translations.inactive') }}
        </span>
      </template>

      <template #cell-actions="{ row }">
        <button
          v-if="editingId !== row.id"
          type="button"
          class="text-orange-600 hover:text-orange-800 text-xs font-medium mr-3"
          @click="startEdit(row)"
        >
          {{ t('admin.translations.edit') }}
        </button>
        <button
          v-else
          type="button"
          class="text-green-600 hover:text-green-800 text-xs font-medium mr-3"
          @click="saveEdit(row.id)"
        >
          {{ t('admin.translations.save') }}
        </button>
        <button
          v-if="editingId === row.id"
          type="button"
          class="text-gray-500 hover:text-gray-700 text-xs mr-3"
          @click="cancelEdit"
        >
          {{ t('admin.translations.cancel') }}
        </button>
        <button
          type="button"
          class="text-red-500 hover:text-red-700 text-xs"
          @click="deleteItem(row.id)"
        >
          {{ t('admin.translations.delete') }}
        </button>
      </template>

      <template #cell-value="{ row }">
        <textarea
          v-if="editingId === row.id"
          v-model="editForm.value"
          rows="2"
          class="w-full px-2 py-1 border rounded text-xs"
        />
        <span v-else class="text-xs max-w-md truncate block">{{ row.value }}</span>
      </template>

      <template #cell-group="{ row }">
        <input
          v-if="editingId === row.id"
          v-model="editForm.group"
          class="w-24 px-2 py-1 border rounded text-xs"
        />
        <span v-else class="text-xs">{{ row.group }}</span>
      </template>

      <template #bottom>
        <div v-if="pagination.totalPages > 1" class="flex justify-center gap-2">
          <button
            v-for="p in pagination.totalPages"
            :key="p"
            type="button"
            class="px-3 py-1 rounded text-sm"
            :class="p === pagination.page ? 'bg-orange-600 text-white' : 'bg-gray-100 hover:bg-gray-200'"
            @click="loadTranslations(p)"
          >
            {{ p }}
          </button>
        </div>
      </template>
    </AdminTable>

    <!-- Create Modal -->
    <AdminModal :open="showCreateModal" :title="t('admin.translations.create')" width="md" @close="showCreateModal = false">
      <div class="space-y-3">
        <div>
          <label class="text-sm font-medium">{{ t('admin.translations.label_key') }}</label>
          <input v-model="newForm.key" class="w-full px-3 py-2 border rounded-lg text-sm" :placeholder="t('admin.translations.placeholder_key')" />
        </div>
        <div>
          <label class="text-sm font-medium">{{ t('admin.translations.label_locale') }}</label>
          <select v-model="newForm.locale" class="w-full px-3 py-2 border rounded-lg text-sm">
            <option v-for="loc in locales" :key="loc.code" :value="loc.code">{{ loc.name || loc.code }}</option>
          </select>
        </div>
        <div>
          <label class="text-sm font-medium">{{ t('admin.translations.label_group') }}</label>
          <select v-model="newForm.group" class="w-full px-3 py-2 border rounded-lg text-sm">
            <option value="errors">errors</option>
            <option value="validation">validation</option>
            <option value="content">content</option>
          </select>
        </div>
        <div>
          <label class="text-sm font-medium">{{ t('admin.translations.label_value') }}</label>
          <textarea v-model="newForm.value" rows="3" class="w-full px-3 py-2 border rounded-lg text-sm" />
        </div>
      </div>
      <div class="flex justify-end gap-3 mt-4">
        <button type="button" class="px-4 py-2 text-sm rounded-lg border hover:bg-gray-50" @click="showCreateModal = false">
          {{ t('admin.translations.cancel') }}
        </button>
        <button type="button" class="px-4 py-2 text-sm bg-orange-600 text-white rounded-lg hover:bg-orange-700" @click="createItem">
          {{ t('admin.translations.create') }}
        </button>
      </div>
    </AdminModal>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: ['auth'] })

const { t, locales } = useI18n()
const api = useApi()

const columns = [
  { key: 'key', label: t('admin.translations.col_key') },
  { key: 'locale', label: t('admin.translations.col_locale') },
  { key: 'group', label: t('admin.translations.col_group') },
  { key: 'value', label: t('admin.translations.col_value') },
  { key: 'isActive', label: t('admin.translations.col_active') },
  { key: 'actions', label: t('admin.translations.col_actions') },
]

const translations = ref<any[]>([])
const pagination = ref({ total: 0, page: 1, limit: 20, totalPages: 0 })
const loading = ref(false)
const error = ref('')

const filterGroup = ref('')
const filterLocale = ref('')
const filterSearch = ref('')

const editingId = ref<number | null>(null)
const editForm = ref({ value: '', group: '', isActive: true })

const showCreateModal = ref(false)
const newForm = ref({ key: '', locale: 'zh', value: '', group: 'errors' })

async function loadTranslations(page = 1) {
  loading.value = true
  error.value = ''
  try {
    const params: Record<string, any> = { page, limit: pagination.value.limit }
    if (filterGroup.value) params.group = filterGroup.value
    if (filterLocale.value) params.locale = filterLocale.value
    if (filterSearch.value) params.search = filterSearch.value

    const res = await api.getTranslations(params)
    translations.value = res.data || []
    pagination.value = res.pagination || pagination.value
  } catch (err: any) {
    error.value = err?.message || t('admin.translations.load_failed')
  } finally {
    loading.value = false
  }
}

function startEdit(item: any) {
  editingId.value = item.id
  editForm.value = { value: item.value, group: item.group, isActive: item.isActive }
}

function cancelEdit() {
  editingId.value = null
}

async function saveEdit(id: number) {
  try {
    await api.updateTranslation(String(id), editForm.value)
    editingId.value = null
    await loadTranslations(pagination.value.page)
  } catch (err: any) {
    error.value = err?.message || t('admin.translations.update_failed')
  }
}

async function deleteItem(id: number) {
  if (!confirm(t('admin.confirm_delete'))) return
  try {
    await api.deleteTranslation(String(id))
    await loadTranslations(pagination.value.page)
  } catch (err: any) {
    error.value = err?.message || t('admin.translations.delete_failed')
  }
}

async function createItem() {
  try {
    await api.createTranslation(newForm.value)
    showCreateModal.value = false
    newForm.value = { key: '', locale: 'zh', value: '', group: 'errors' }
    await loadTranslations(1)
  } catch (err: any) {
    error.value = err?.message || t('admin.translations.create_failed')
  }
}

async function exportAll() {
  try {
    const data = await api.exportTranslations()
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `translations_${new Date().toISOString().split('T')[0]}.json`
    a.click()
    URL.revokeObjectURL(url)
  } catch (err: any) {
    error.value = err?.message || t('admin.translations.export_failed')
  }
}

onMounted(() => loadTranslations())
</script>
