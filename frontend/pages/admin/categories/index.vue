<template>
  <div>
    <PageHeader :title="t('admin.categories.title')" :description="t('admin.categories.description')">
      <template #actions>
        <button type="button" class="inline-flex items-center rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700" @click="openCreate">
          {{ t('admin.categories.add_category') }}
        </button>
      </template>
    </PageHeader>

    <AdminTable
      class="mt-6"
      :columns="columns"
      :rows="categories"
      :loading="pending"
      :error="!!error"
      :error-message="error"
      :empty-text="t('admin.categories.no_data')"
      @retry="fetchCategories"
    >
      <template #cell-name="{ row }">
        <div class="font-medium text-gray-900">{{ tField(row, 'name') || row.name }}</div>
        <div v-if="row.alias || tField(row, 'alias')" class="text-xs text-gray-500">{{ tField(row, 'alias') || row.alias }}</div>
      </template>
      <template #cell-productCount="{ row }">{{ row.productCount ?? 0 }}</template>
      <template #cell-actions="{ row }">
        <button type="button" class="text-orange-600 hover:text-orange-900" @click="openEdit(row)">{{ t('admin.categories.edit') }}</button>
        <button type="button" class="ml-4 text-red-600 hover:text-red-900" @click="removeCategory(row.slug)">{{ t('admin.categories.delete') }}</button>
      </template>
    </AdminTable>

    <AdminModal :open="showModal" :title="editingSlug ? t('admin.categories.edit_category') : t('admin.categories.create_category')" width="lg" @close="closeModal">
      <form @submit.prevent="saveCategory">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="admin-form-label">{{ t('admin.categories.slug') }}</label>
            <input v-model="form.slug" :disabled="Boolean(editingSlug)" required class="admin-form-control" />
          </div>
          <div>
            <label class="admin-form-label">{{ t('admin.categories.icon') }}</label>
            <input v-model="form.icon" class="admin-form-control" />
          </div>
          <div class="sm:col-span-2">
            <label class="admin-form-label">{{ t('admin.categories.thumbnail') }}</label>
            <input v-model="form.thumbnail" class="admin-form-control" />
          </div>
        </div>

        <div class="admin-form-section mt-4">
          <h4 class="admin-form-section-title">{{ t('admin.categories.translations_label') }}</h4>
          <p class="mt-1 text-xs text-gray-500">{{ t('admin.categories.translations_hint') }}</p>
          <div class="flex gap-2 my-3 flex-wrap">
            <button
              v-for="loc in contentLocales"
              :key="loc"
              type="button"
              :class="['admin-locale-tab', translationLocale === loc ? 'admin-locale-tab--active' : '']"
              @click="translationLocale = loc"
            >
              {{ localeTabLabel(loc) }}
            </button>
          </div>
          <div class="grid grid-cols-1 gap-4">
            <div>
              <label class="admin-form-label">{{ t('admin.categories.field_name') }} ({{ translationLocale }})</label>
              <input v-model="form.translations[translationLocale].name" :required="translationLocale === SOURCE_LOCALE" class="admin-form-control" />
            </div>
            <div>
              <label class="admin-form-label">{{ t('admin.categories.field_alias') }} ({{ translationLocale }})</label>
              <input v-model="form.translations[translationLocale].alias" class="admin-form-control" />
            </div>
            <div>
              <label class="admin-form-label">{{ t('admin.categories.field_description') }} ({{ translationLocale }})</label>
              <textarea v-model="form.translations[translationLocale].description" rows="3" class="admin-form-control"></textarea>
            </div>
          </div>
        </div>

        <div v-if="formError" class="mt-4 text-sm text-red-600">{{ formError }}</div>
        <div class="mt-4 flex justify-end gap-3">
          <button type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700" @click="closeModal">{{ t('admin.categories.cancel') }}</button>
          <button type="submit" :disabled="saving" class="rounded-md bg-orange-600 px-4 py-2 text-sm text-white disabled:opacity-50">{{ saving ? t('admin.categories.saving') : (editingSlug ? t('admin.categories.update') : t('admin.categories.create')) }}</button>
        </div>
      </form>
    </AdminModal>

    <p v-if="actionMessage" class="mt-4 text-sm" :class="actionError ? 'text-red-600' : 'text-green-600'">{{ actionMessage }}</p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t } = useI18n()
const { tField } = useTranslation()

const contentLocales = ALL_LOCALES
const translationLocale = ref(SOURCE_LOCALE)

const columns = [
  { key: 'name', label: t('admin.categories.col_name') },
  { key: 'slug', label: t('admin.categories.col_slug') },
  { key: 'productCount', label: t('admin.categories.col_products') },
  { key: 'actions', label: '' },
]

const localeTabLabelMap: Record<string, string> = {
  en: 'admin.products.locale_tab_en', zh: 'admin.products.locale_tab_zh',
  ko: 'admin.products.locale_tab_ko', ar: 'admin.products.locale_tab_ar',
  ja: 'admin.products.locale_tab_ja', th: 'admin.products.locale_tab_th',
  vi: 'admin.products.locale_tab_vi', id: 'admin.products.locale_tab_id',
  ms: 'admin.products.locale_tab_ms',
}
const localeTabLabel = (loc: string) => t(localeTabLabelMap[loc] || loc)

function emptyCategoryTranslations() {
  const result: Record<string, { name: string; alias: string; description: string }> = {}
  for (const loc of contentLocales) {
    result[loc] = { name: '', alias: '', description: '' }
  }
  return result
}

const categories = ref<any[]>([])
const pending = ref(true)
const error = ref('')
const showModal = ref(false)
const editingSlug = ref('')
const saving = ref(false)
const formError = ref('')
const actionMessage = ref('')
const actionError = ref(false)

const form = reactive({
  slug: '',
  icon: '',
  thumbnail: '',
  translations: emptyCategoryTranslations(),
})

const resetForm = () => {
  form.slug = ''
  form.icon = ''
  form.thumbnail = ''
  form.translations = emptyCategoryTranslations()
}

const syncScalarsFromSource = () => {
  const src = form.translations[SOURCE_LOCALE]
  if (!src) return
  return {
    name: src.name?.trim() || '',
    alias: src.alias?.trim() || '',
    description: src.description?.trim() || '',
  }
}

const buildTranslationsPayload = () => {
  const result: Record<string, Record<string, string>> = {}
  for (const loc of contentLocales) {
    const tr = form.translations[loc]
    if (!tr) continue
    const entry: Record<string, string> = {}
    if (tr.name?.trim()) entry.name = tr.name.trim()
    if (tr.alias?.trim()) entry.alias = tr.alias.trim()
    if (tr.description?.trim()) entry.description = tr.description.trim()
    if (Object.keys(entry).length > 0) result[loc] = entry
  }
  return Object.keys(result).length > 0 ? result : null
}

const fetchCategories = async () => {
  pending.value = true
  error.value = ''
  try {
    categories.value = await api.adminGetCategories()
  } catch (err: any) {
    error.value = err?.message || t('admin.categories.load_failed')
  } finally {
    pending.value = false
  }
}

const openCreate = () => {
  editingSlug.value = ''
  formError.value = ''
  translationLocale.value = SOURCE_LOCALE
  resetForm()
  showModal.value = true
}

const openEdit = (row: any) => {
  editingSlug.value = row.slug
  formError.value = ''
  translationLocale.value = SOURCE_LOCALE
  resetForm()
  form.slug = row.slug || ''
  form.icon = row.icon || ''
  form.thumbnail = row.thumbnail || ''
  form.translations = emptyCategoryTranslations()
  if (row.translations && typeof row.translations === 'object') {
    for (const loc of contentLocales) {
      if (row.translations[loc]) {
        form.translations[loc].name = row.translations[loc].name || ''
        form.translations[loc].alias = row.translations[loc].alias || ''
        form.translations[loc].description = row.translations[loc].description || ''
      }
    }
  }
  const src = form.translations[SOURCE_LOCALE]
  if (!src.name?.trim() && row.name) src.name = row.name
  if (!src.alias?.trim() && row.alias) src.alias = row.alias
  if (!src.description?.trim() && row.description) src.description = row.description
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  saving.value = false
  formError.value = ''
}

const saveCategory = async () => {
  const scalars = syncScalarsFromSource()
  if (!scalars?.name) {
    formError.value = t('admin.categories.name_required')
    return
  }
  if (!form.slug.trim()) {
    formError.value = t('admin.categories.slug_required')
    return
  }
  saving.value = true
  formError.value = ''
  const payload = {
    slug: form.slug.trim(),
    name: scalars.name,
    alias: scalars.alias,
    description: scalars.description,
    icon: form.icon.trim(),
    thumbnail: form.thumbnail.trim(),
    translations: buildTranslationsPayload(),
  }
  try {
    if (editingSlug.value) {
      await api.adminUpdateCategory(editingSlug.value, payload)
      actionMessage.value = t('admin.categories.updated_success')
    } else {
      await api.adminCreateCategory(payload)
      actionMessage.value = t('admin.categories.created_success')
    }
    actionError.value = false
    showModal.value = false
    await fetchCategories()
  } catch (err: any) {
    formError.value = err?.message || t('admin.categories.save_failed')
  } finally {
    saving.value = false
  }
}

const removeCategory = async (slug: string) => {
  if (!confirm(t('admin.categories.confirm_delete'))) return
  try {
    await api.adminDeleteCategory(slug)
    actionMessage.value = t('admin.categories.deleted_success')
    actionError.value = false
    await fetchCategories()
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.message || t('admin.categories.save_failed')
  }
}

onMounted(fetchCategories)
</script>
