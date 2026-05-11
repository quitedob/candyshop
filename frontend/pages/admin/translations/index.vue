<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: ['auth'] })

const { t } = useI18n()
const api = useApi()

const translations = ref<any[]>([])
const pagination = ref({ total: 0, page: 1, limit: 20, totalPages: 0 })
const loading = ref(false)
const error = ref('')

// Filters
const filterGroup = ref('')
const filterLocale = ref('')
const filterSearch = ref('')

// Editing
const editingId = ref<number | null>(null)
const editForm = ref({ value: '', group: '', isActive: true })

// Creating
const showCreateModal = ref(false)
const newForm = ref({ key: '', locale: 'en', value: '', group: 'errors' })

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
    error.value = err?.message || 'Failed to load translations'
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
    error.value = err?.message || 'Failed to update'
  }
}

async function deleteItem(id: number) {
  if (!confirm(t('admin.confirm_delete'))) return
  try {
    await api.deleteTranslation(String(id))
    await loadTranslations(pagination.value.page)
  } catch (err: any) {
    error.value = err?.message || 'Failed to delete'
  }
}

async function createItem() {
  try {
    await api.createTranslation(newForm.value)
    showCreateModal.value = false
    newForm.value = { key: '', locale: 'en', value: '', group: 'errors' }
    await loadTranslations(1)
  } catch (err: any) {
    error.value = err?.message || 'Failed to create'
  }
}

async function exportAll() {
  try {
    const data = await api.exportTranslations()
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'translations.json'
    a.click()
    URL.revokeObjectURL(url)
  } catch (err: any) {
    error.value = err?.message || 'Export failed'
  }
}

onMounted(() => loadTranslations())
</script>

<template>
  <div class="p-6 max-w-7xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
        {{ t('admin.translations.title') }}
      </h1>
      <div class="flex gap-3">
        <button
          class="px-4 py-2 text-sm bg-gray-100 dark:bg-gray-700 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600"
          @click="exportAll"
        >
          {{ t('admin.translations.export') }}
        </button>
        <button
          class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          @click="showCreateModal = true"
        >
          {{ t('admin.translations.add') }}
        </button>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex gap-3 mb-4 flex-wrap">
      <input
        v-model="filterSearch"
        type="text"
        :placeholder="t('admin.translations.search_placeholder')"
        class="px-3 py-2 border rounded-lg text-sm dark:bg-gray-800 dark:border-gray-600"
        @keyup.enter="loadTranslations(1)"
      />
      <select
        v-model="filterGroup"
        class="px-3 py-2 border rounded-lg text-sm dark:bg-gray-800 dark:border-gray-600"
        @change="loadTranslations(1)"
      >
        <option value="">{{ t('admin.translations.all_groups') }}</option>
        <option value="errors">errors</option>
        <option value="validation">validation</option>
        <option value="content">content</option>
      </select>
      <select
        v-model="filterLocale"
        class="px-3 py-2 border rounded-lg text-sm dark:bg-gray-800 dark:border-gray-600"
        @change="loadTranslations(1)"
      >
        <option value="">{{ t('admin.translations.all_locales') }}</option>
        <option value="zh">中文</option>
        <option value="en">English</option>
      </select>
      <button
        class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        @click="loadTranslations(1)"
      >
        {{ t('admin.translations.filter') }}
      </button>
    </div>

    <!-- Error -->
    <div v-if="error" class="mb-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 rounded-lg text-sm">
      {{ error }}
    </div>

    <!-- Table -->
    <div class="overflow-x-auto bg-white dark:bg-gray-800 rounded-lg shadow">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 dark:bg-gray-700 text-left">
          <tr>
            <th class="px-4 py-3 font-medium">Key</th>
            <th class="px-4 py-3 font-medium">Locale</th>
            <th class="px-4 py-3 font-medium">Group</th>
            <th class="px-4 py-3 font-medium">Value</th>
            <th class="px-4 py-3 font-medium">Active</th>
            <th class="px-4 py-3 font-medium text-right">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
          <tr v-if="loading">
            <td colspan="6" class="px-4 py-8 text-center text-gray-500">Loading...</td>
          </tr>
          <tr v-else-if="translations.length === 0">
            <td colspan="6" class="px-4 py-8 text-center text-gray-500">
              {{ t('admin.translations.no_data') }}
            </td>
          </tr>
          <tr
            v-for="item in translations"
            :key="item.id"
            class="hover:bg-gray-50 dark:hover:bg-gray-700/50"
          >
            <template v-if="editingId === item.id">
              <td class="px-4 py-3 font-mono text-xs">{{ item.key }}</td>
              <td class="px-4 py-3">{{ item.locale }}</td>
              <td class="px-4 py-3">
                <input v-model="editForm.group" class="w-24 px-2 py-1 border rounded text-xs dark:bg-gray-700 dark:border-gray-600" />
              </td>
              <td class="px-4 py-3">
                <textarea
                  v-model="editForm.value"
                  rows="2"
                  class="w-full px-2 py-1 border rounded text-xs dark:bg-gray-700 dark:border-gray-600"
                />
              </td>
              <td class="px-4 py-3">
                <input v-model="editForm.isActive" type="checkbox" />
              </td>
              <td class="px-4 py-3 text-right space-x-2">
                <button class="text-green-600 hover:text-green-800 text-xs font-medium" @click="saveEdit(item.id)">Save</button>
                <button class="text-gray-500 hover:text-gray-700 text-xs" @click="cancelEdit">Cancel</button>
              </td>
            </template>
            <template v-else>
              <td class="px-4 py-3 font-mono text-xs text-blue-600">{{ item.key }}</td>
              <td class="px-4 py-3">
                <span class="px-2 py-0.5 rounded text-xs bg-gray-100 dark:bg-gray-600">{{ item.locale }}</span>
              </td>
              <td class="px-4 py-3 text-xs">{{ item.group }}</td>
              <td class="px-4 py-3 text-xs max-w-md truncate">{{ item.value }}</td>
              <td class="px-4 py-3">
                <span :class="item.isActive ? 'text-green-600' : 'text-red-500'" class="text-xs">
                  {{ item.isActive ? 'Active' : 'Inactive' }}
                </span>
              </td>
              <td class="px-4 py-3 text-right space-x-2">
                <button class="text-blue-600 hover:text-blue-800 text-xs font-medium" @click="startEdit(item)">Edit</button>
                <button class="text-red-500 hover:text-red-700 text-xs" @click="deleteItem(item.id)">Delete</button>
              </td>
            </template>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="pagination.totalPages > 1" class="flex justify-center mt-4 gap-2">
      <button
        v-for="p in pagination.totalPages"
        :key="p"
        class="px-3 py-1 rounded text-sm"
        :class="p === pagination.page ? 'bg-blue-600 text-white' : 'bg-gray-100 dark:bg-gray-700 hover:bg-gray-200'"
        @click="loadTranslations(p)"
      >
        {{ p }}
      </button>
    </div>

    <!-- Create Modal -->
    <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showCreateModal = false">
      <div class="bg-white dark:bg-gray-800 rounded-lg p-6 w-full max-w-md shadow-xl">
        <h2 class="text-lg font-bold mb-4">{{ t('admin.translations.create') }}</h2>
        <div class="space-y-3">
          <div>
            <label class="text-sm font-medium">Key</label>
            <input v-model="newForm.key" class="w-full px-3 py-2 border rounded-lg text-sm dark:bg-gray-700 dark:border-gray-600" placeholder="e.g. errors.my_new_code" />
          </div>
          <div>
            <label class="text-sm font-medium">Locale</label>
            <select v-model="newForm.locale" class="w-full px-3 py-2 border rounded-lg text-sm dark:bg-gray-700 dark:border-gray-600">
              <option value="zh">中文</option>
              <option value="en">English</option>
            </select>
          </div>
          <div>
            <label class="text-sm font-medium">Group</label>
            <select v-model="newForm.group" class="w-full px-3 py-2 border rounded-lg text-sm dark:bg-gray-700 dark:border-gray-600">
              <option value="errors">errors</option>
              <option value="validation">validation</option>
              <option value="content">content</option>
            </select>
          </div>
          <div>
            <label class="text-sm font-medium">Value</label>
            <textarea v-model="newForm.value" rows="3" class="w-full px-3 py-2 border rounded-lg text-sm dark:bg-gray-700 dark:border-gray-600" />
          </div>
        </div>
        <div class="flex justify-end gap-3 mt-4">
          <button class="px-4 py-2 text-sm rounded-lg border hover:bg-gray-50 dark:hover:bg-gray-700" @click="showCreateModal = false">
            Cancel
          </button>
          <button class="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700" @click="createItem">
            Create
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
