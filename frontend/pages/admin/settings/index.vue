<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('admin.settings.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ t('admin.settings.description') }}</p>
      </div>
    </div>

    <!-- Tabs -->
    <div class="mt-6 border-b border-gray-200">
      <nav class="-mb-px flex space-x-8">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          @click="switchTab(tab.key)"
          :class="[
            activeTab === tab.key
              ? 'border-orange-500 text-orange-600'
              : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
            'whitespace-nowrap border-b-2 py-4 px-1 text-sm font-medium'
          ]"
        >
          {{ tab.label }}
        </button>
      </nav>
    </div>

    <!-- Settings Form -->
    <div class="mt-6 space-y-4">
      <div v-if="loading" class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <p class="text-sm text-gray-500">{{ t('admin.settings.loading') }}</p>
      </div>
      <div v-else-if="loadError" class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <p class="text-sm text-red-600">{{ loadError }}</p>
      </div>
      <div v-else-if="settings.length === 0" class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <p class="text-sm text-gray-500">{{ t('admin.settings.no_settings') }}</p>
      </div>
      <div
        v-else
        v-for="setting in settings"
        :key="setting.key"
        class="bg-white rounded-xl shadow-sm border border-gray-200 p-5"
      >
        <div class="flex items-start gap-4">
          <div class="flex-1 min-w-0">
            <label :for="`setting-${setting.key}`" class="block text-sm font-medium text-gray-900">
              {{ setting.label || setting.key }}
            </label>
            <p v-if="setting.description" class="mt-0.5 text-xs text-gray-500">{{ setting.description }}</p>
            <div class="mt-2">
              <input
                :id="`setting-${setting.key}`"
                v-model="setting.editValue"
                type="text"
                class="block w-full max-w-lg rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500 focus:border-orange-500"
                :disabled="savingKey === setting.key"
              />
            </div>
          </div>
          <div class="flex-shrink-0 pt-6">
            <button
              type="button"
              @click="saveSetting(setting)"
              :disabled="savingKey === setting.key || setting.editValue === setting.value"
              class="inline-flex items-center gap-1.5 px-3 py-2 bg-orange-600 text-white rounded-lg text-sm font-medium hover:bg-orange-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <Icon v-if="savingKey === setting.key" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" />
              <Icon v-else name="heroicons:check" class="h-4 w-4" />
              {{ savingKey === setting.key ? t('admin.settings.saving') : t('admin.settings.save') }}
            </button>
          </div>
        </div>
        <div v-if="saveErrors[setting.key]" class="mt-2 text-xs text-red-600">{{ saveErrors[setting.key] }}</div>
        <div v-if="savedKeys[setting.key]" class="mt-2 text-xs text-green-600">{{ t('admin.settings.saved') }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()

const activeTab = ref('general')
const tabs = [
  { key: 'general', label: t('admin.settings.tab_general') },
  { key: 'email', label: t('admin.settings.tab_email') },
  { key: 'trade', label: t('admin.settings.tab_trade') },
  { key: 'notifications', label: t('admin.settings.tab_notifications') }
]

interface Setting {
  key: string
  value: string
  editValue: string
  label?: string
  description?: string
  category?: string
}

const settings = ref<Setting[]>([])
const loading = ref(false)
const loadError = ref('')
const savingKey = ref<string | null>(null)
const saveErrors = reactive<Record<string, string>>({})
const savedKeys = reactive<Record<string, boolean>>({})

const fetchSettings = async () => {
  loading.value = true
  loadError.value = ''
  try {
    const res = await api.get<any>('/admin/settings', { category: activeTab.value })
    const raw = res.data || res || []
    settings.value = (Array.isArray(raw) ? raw : []).map((s: any) => ({
      key: s.key,
      value: s.value ?? '',
      editValue: s.value ?? '',
      label: s.label || '',
      description: s.description || '',
      category: s.category || ''
    }))
  } catch (err: any) {
    loadError.value = err?.message || t('errors.api.load_failed')
  } finally {
    loading.value = false
  }
}

const switchTab = (key: string) => {
  activeTab.value = key
  // Clear save confirmations
  Object.keys(savedKeys).forEach(k => delete savedKeys[k])
  Object.keys(saveErrors).forEach(k => delete saveErrors[k])
  fetchSettings()
}

const saveSetting = async (setting: Setting) => {
  savingKey.value = setting.key
  saveErrors[setting.key] = ''
  savedKeys[setting.key] = false
  try {
    await api.put(`/admin/settings/${setting.key}`, {
      value: setting.editValue,
      category: activeTab.value
    })
    setting.value = setting.editValue
    savedKeys[setting.key] = true
    setTimeout(() => { savedKeys[setting.key] = false }, 3000)
  } catch (err: any) {
    saveErrors[setting.key] = err?.message || t('errors.api.save_failed')
  } finally {
    savingKey.value = null
  }
}

onMounted(fetchSettings)
</script>
