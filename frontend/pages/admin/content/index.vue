<template>
  <div>
    <PageHeader :title="t('admin.content.title')" :description="t('admin.content.description')">
      <template #actions>
        <button type="button" class="inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50" @click="openCreateModal">
          {{ t('admin.content.add_content') }}
        </button>
        <button type="button" class="inline-flex items-center gap-2 rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700" @click="openCreateDrawer">
          <Icon name="heroicons:sparkles" class="h-4 w-4" />
          {{ t('admin.content.ai_import') }}
        </button>
        <AiHelpHint topic="content_import" />
      </template>
    </PageHeader>

    <!-- Type Tabs -->
    <div class="mt-6 flex items-center gap-2">
      <button
        v-for="option in typeOptions"
        :key="option.value"
        type="button"
        class="rounded-md px-3 py-2 text-sm font-medium"
        :class="selectedType === option.value ? 'bg-orange-600 text-white' : 'bg-white text-gray-700 ring-1 ring-inset ring-gray-300 hover:bg-gray-50'"
        @click="switchType(option.value)"
      >
        {{ option.value === 'post' ? t('admin.content.tab_posts') : t('admin.content.tab_cases') }}
      </button>
    </div>

    <AdminTable
      class="mt-6"
      :columns="columns"
      :rows="contentList"
      :loading="pending"
      :error="!!error"
      :error-message="error"
      :empty-text="t('admin.content.no_data')"
      @retry="fetchContent"
    >
      <template #cell-type="{ row }">
        <span class="inline-flex rounded-full px-2 text-xs font-semibold leading-5" :class="row.type === 'post' ? 'bg-amber-100 text-amber-800' : 'bg-cyan-100 text-cyan-800'">
          {{ enumLabel('content_type', row.type) }}
        </span>
      </template>

      <template #cell-meta="{ row }">
        {{ row.type === 'post' ? (row.category || t('admin.content.category_general')) : (row.industry || t('admin.content.category_general')) }}
      </template>

      <template #cell-updatedAt="{ row }">
        {{ formatDate(row.updatedAt, { dateStyle: 'medium', timeStyle: 'short' }) }}
      </template>

      <template #cell-actions="{ row }">
        <button type="button" class="text-orange-600 hover:text-orange-900" @click="openEditModal(row)">{{ t('admin.content.edit') }}</button>
        <button type="button" class="ml-4 text-red-600 hover:text-red-900" @click="deleteContent(row)">{{ t('admin.content.delete') }}</button>
      </template>

      <template #bottom>
        <div v-if="pagination" class="flex items-center justify-between">
          <div class="text-sm text-gray-700">
            {{ t('admin.content.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}
          </div>
          <div class="flex items-center gap-2">
            <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page <= 1" @click="prevPage">
              {{ t('admin.content.previous') }}
            </button>
            <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page >= pagination.totalPages" @click="nextPage">
              {{ t('admin.content.next') }}
            </button>
          </div>
        </div>
      </template>
    </AdminTable>

    <!-- Editor Modal -->
    <AdminModal :open="showModal" :title="editingId ? t('admin.content.edit_content') : t('admin.content.create_content')" width="xl" @close="closeModal">
      <form @submit.prevent="saveContent" class="space-y-4">
        <ContentFormFields
          :form="form"
          :disable-type="Boolean(editingId)"
          :editing-id="editingId"
          :ai-translating="aiTranslating"
          :translation-locale="translationLocale"
          @ai-translate="aiTranslateContent"
          @update:translation-locale="translationLocale = $event"
        />

        <!-- AI Generate Bar -->
        <div class="flex items-center gap-3 rounded-lg bg-amber-50 border border-amber-200 p-3">
          <Icon name="heroicons:sparkles" class="h-5 w-5 text-amber-600 flex-shrink-0" aria-hidden="true" />
          <label for="content-aiTopic" class="sr-only">{{ t('admin.content.ai_topic_placeholder') }}</label>
          <input
            id="content-aiTopic"
            v-model="aiTopic"
            name="aiTopic"
            type="text"
            :placeholder="form.type === 'post' ? t('admin.content.ai_topic_placeholder') : t('admin.content.ai_case_topic_placeholder')"
            class="flex-1 rounded-md border border-amber-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500"
            @keydown.enter.prevent="generateWithAI"
          />
          <button
            type="button"
            :disabled="aiGenerating || !aiTopic.trim()"
            class="inline-flex items-center gap-1.5 rounded-md bg-orange-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50"
            @click="generateWithAI"
          >
            <Icon v-if="aiGenerating" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" aria-hidden="true" />
            <Icon v-else name="heroicons:sparkles" class="h-4 w-4" aria-hidden="true" />
            {{ aiGenerating ? t('admin.content.ai_generating') : t('admin.content.ai_generate') }}
          </button>
          <AiHelpHint topic="content_modal_ai" size="sm" />
        </div>

        <div v-if="formError" class="text-sm text-red-600">{{ formError }}</div>

        <div class="flex justify-end gap-3 pt-4 border-t border-gray-200">
          <button type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700" @click="closeModal">
            {{ t('admin.content.cancel') }}
          </button>
          <button type="submit" :disabled="saving" class="rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white disabled:opacity-50 hover:bg-orange-700">
            {{ saving ? t('admin.content.saving') : (editingId ? t('admin.content.update') : t('admin.content.create')) }}
          </button>
        </div>
      </form>
    </AdminModal>

    <!-- AI Generate Drawer：先中文草稿审阅，再翻译发布 -->
    <Drawer :open="showCreateDrawer" width="2xl" @close="showCreateDrawer = false">
      <template #title>
        <span class="inline-flex items-center">
          {{ t('admin.content.ai_create_drawer_title') }}
          <AiHelpHint topic="content_import" size="sm" />
        </span>
      </template>

      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6" style="min-height: 60vh;">
        <!-- LEFT：用户输入 -->
        <div class="lg:col-span-5 space-y-4">
          <div>
            <label for="ai-content-type" class="block text-sm font-medium text-gray-700">{{ t('admin.content.type') }}</label>
            <select id="ai-content-type" v-model="form.type" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :disabled="aiDraftReady">
              <option value="post">{{ enumLabel('content_type', 'post') }}</option>
              <option value="case">{{ enumLabel('content_type', 'case') }}</option>
            </select>
          </div>
          <div>
            <label for="ai-content-topic" class="block text-sm font-medium text-gray-700">{{ t('admin.content.ai_topic_label') }}</label>
            <textarea id="ai-content-topic" v-model="aiContentTopic" rows="6" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="form.type === 'post' ? t('admin.content.ai_topic_placeholder') : t('admin.content.ai_case_topic_placeholder')" :disabled="aiPublishing" />
            <p class="mt-1 text-xs text-gray-500">{{ t('admin.content.ai_brief_hint') }}</p>
          </div>
          <button type="button" :disabled="aiGeneratingContent || aiPublishing || !aiContentTopic.trim()" class="inline-flex items-center gap-2 rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50" @click="generateContentAI">
            <Icon name="heroicons:sparkles" class="h-4 w-4" />
            {{ aiGeneratingContent ? t('admin.content.ai_generating') : t('admin.content.ai_generate_zh_draft') }}
          </button>
          <p v-if="formError" class="text-sm text-red-600">{{ formError }}</p>
        </div>

        <!-- RIGHT：中文草稿审阅 -->
        <div class="lg:col-span-7 lg:border-l lg:pl-6 overflow-y-auto" style="max-height: 65vh;">
          <div v-if="!aiDraftReady && !aiGeneratingContent" class="flex flex-col items-center justify-center h-48 text-gray-400 text-sm">
            <Icon name="heroicons:sparkles" class="h-10 w-10 mb-2 text-gray-300" />
            {{ t('admin.content.ai_empty_preview') }}
          </div>

          <div v-if="aiGeneratingContent" class="flex flex-col items-center justify-center h-48 text-gray-500 text-sm gap-2">
            <svg class="animate-spin h-6 w-6 text-orange-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
            {{ t('admin.content.ai_generating') }}
          </div>

          <div v-if="aiDraftReady && !aiGeneratingContent" class="space-y-4">
            <div class="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
              <span class="inline-flex items-center">
                {{ t('admin.content.ai_review_banner') }}
                <AiHelpHint topic="content_inline_edit" size="sm" />
              </span>
            </div>

            <ContentFormFields
              :form="form"
              :disable-type="true"
              :editing-id="''"
              :ai-translating="false"
              :translation-locale="translationLocale"
              :hide-translations="true"
              :inline-ai-edit="true"
              @update:translation-locale="translationLocale = $event"
            />
          </div>
        </div>
      </div>

      <template #footer>
        <p v-if="formError" class="mb-3 text-sm text-red-600">{{ formError }}</p>
        <div class="flex justify-end gap-3">
          <button type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700" :disabled="aiPublishing" @click="showCreateDrawer = false">
            {{ t('admin.content.cancel') }}
          </button>
          <button v-if="aiDraftReady" type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="aiPublishing || aiGeneratingContent" @click="generateContentAI">
            {{ t('admin.content.ai_regenerate') }}
          </button>
          <button type="button" :disabled="aiPublishing || !aiDraftReady" class="inline-flex items-center gap-2 rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50" @click="publishAiContent">
            <Icon v-if="aiPublishing" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" />
            <Icon v-else name="heroicons:rocket-launch" class="h-4 w-4" />
            {{ aiPublishing ? t('admin.content.ai_publish_translating') : t('admin.content.ai_confirm_publish') }}
          </button>
        </div>
      </template>
    </Drawer>

    <p v-if="actionMessage" class="mt-4 text-sm" :class="actionError ? 'text-red-600' : 'text-green-600'">{{ actionMessage }}</p>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { marked } from 'marked'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { sanitize } = useSanitizer()
const { t, locale, defaultLocale } = useI18n()
const { enumLabel, formatDate } = useDisplay()

const typeOptions = [{ value: 'post' }, { value: 'case' }] as const
type ContentType = 'post' | 'case'

const columns = [
  { key: 'title', label: t('admin.content.col_title') },
  { key: 'type', label: t('admin.content.col_type') },
  { key: 'meta', label: t('admin.content.col_meta') },
  { key: 'updatedAt', label: t('admin.content.col_updated') },
  { key: 'actions', label: '' },
]

// List state
const selectedType = ref<ContentType>('post')
const contentList = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20

// Modal state
const showModal = ref(false)
const editingId = ref('')
const saving = ref(false)
const formError = ref('')
const actionMessage = ref('')
const actionError = ref(false)

// AI state (edit modal)
const aiTopic = ref('')
const aiGenerating = ref(false)
const aiTranslating = ref(false)
const translationLocale = ref(SOURCE_LOCALE)

// AI Drawer state（两阶段：中文草稿 → 翻译发布）
const showCreateDrawer = ref(false)
const aiContentTopic = ref('')
const aiGeneratingContent = ref(false)
const aiDraftReady = ref(false)
const aiPublishing = ref(false)
// ALL_LOCALES 由 Nuxt 自动导入
const contentLocales = ALL_LOCALES

const localeTabLabelMap: Record<string, string> = {
  en: 'admin.products.locale_tab_en', zh: 'admin.products.locale_tab_zh',
  ko: 'admin.products.locale_tab_ko', ar: 'admin.products.locale_tab_ar',
  ja: 'admin.products.locale_tab_ja', th: 'admin.products.locale_tab_th',
  vi: 'admin.products.locale_tab_vi', id: 'admin.products.locale_tab_id',
  ms: 'admin.products.locale_tab_ms',
}
const localeTabLabel = (loc: string) => t(localeTabLabelMap[loc] || loc)

function emptyContentTranslations() {
  const result: Record<string, {
    title: string; excerpt: string; content: string
    authorName: string; authorTitle: string; authorBio: string
    challenge: string; solution: string; result: string
  }> = {}
  for (const loc of contentLocales) {
    result[loc] = {
      title: '', excerpt: '', content: '', authorName: '', authorTitle: '', authorBio: '',
      challenge: '', solution: '', result: '',
    }
  }
  return result
}

const contentTranslationKeys = ['title', 'excerpt', 'content', 'authorName', 'authorTitle', 'authorBio', 'challenge', 'solution', 'result'] as const

/** 将主语言翻译同步到标量字段 */
const syncScalarsFromSourceLocale = () => {
  const src = form.translations[SOURCE_LOCALE]
  if (!src) return
  form.title = src.title?.trim?.() || form.title
  form.excerpt = src.excerpt?.trim?.() || form.excerpt
  form.content = src.content?.trim?.() || form.content
  form.authorName = src.authorName?.trim?.() || form.authorName
  form.authorTitle = src.authorTitle?.trim?.() || form.authorTitle
  form.authorBio = src.authorBio?.trim?.() || form.authorBio
  form.challenge = src.challenge?.trim?.() || form.challenge
  form.solution = src.solution?.trim?.() || form.solution
  form.result = src.result?.trim?.() || form.result
}

/** 旧数据回填主语言翻译 */
const backfillSourceLocaleFromScalars = () => {
  const src = form.translations[SOURCE_LOCALE]
  if (!src) return
  if (!src.title?.trim() && form.title) src.title = form.title
  if (!src.excerpt?.trim() && form.excerpt) src.excerpt = form.excerpt
  if (!src.content?.trim() && form.content) src.content = form.content
  if (!src.authorName?.trim() && form.authorName) src.authorName = form.authorName
  if (!src.authorTitle?.trim() && form.authorTitle) src.authorTitle = form.authorTitle
  if (!src.authorBio?.trim() && form.authorBio) src.authorBio = form.authorBio
  if (!src.challenge?.trim() && form.challenge) src.challenge = form.challenge
  if (!src.solution?.trim() && form.solution) src.solution = form.solution
  if (!src.result?.trim() && form.result) src.result = form.result
}

const applyContentTranslationFields = (loc: string, fields: Record<string, any>) => {
  if (!fields || typeof fields !== 'object' || !form.translations[loc]) return
  for (const key of contentTranslationKeys) {
    if (fields[key]) form.translations[loc][key] = fields[key]
  }
}

const form = reactive({
  type: 'post' as ContentType,
  title: '', slug: '', thumbnail: '', ogImage: '', category: '', readTime: 5,
  excerpt: '', tagsInput: '', content: '',
  authorName: '', authorAvatar: '', authorTitle: '', authorBio: '',
  client: '', industry: '', location: '', timeline: '',
  servicesInput: '', imagesInput: '', challenge: '', solution: '', result: '',
  translations: emptyContentTranslations()
})

const parseCSV = (v: string) => v.split(',').map(s => s.trim()).filter(Boolean)

const resetForm = () => {
  form.type = selectedType.value; form.title = ''; form.slug = ''; form.thumbnail = ''; form.ogImage = ''
  form.category = ''; form.readTime = 5; form.excerpt = ''; form.tagsInput = ''; form.content = ''
  form.authorName = ''; form.authorAvatar = ''; form.authorTitle = ''; form.authorBio = ''
  form.client = ''; form.industry = ''; form.location = ''; form.timeline = ''
  form.servicesInput = ''; form.imagesInput = ''; form.challenge = ''; form.solution = ''; form.result = ''
  form.translations = emptyContentTranslations()
}

const generateWithAI = async () => {
  if (!aiTopic.value.trim() || aiGenerating.value) return
  aiGenerating.value = true
  formError.value = ''
  try {
    const res = await api.adminAIGenerateContent({ topic: aiTopic.value, type: form.type, language: defaultLocale.value || 'zh', translate: false })
    const data = res.content

    if (res.parseError || typeof data === 'string') {
      const raw = typeof data === 'string' ? data : ''
      const lines = raw.split('\n')
      const titleLine = lines.find((l: string) => l.startsWith('# '))
      if (titleLine && !form.title) form.title = titleLine.replace(/^#\s+/, '')
      const tagsLine = lines.find((l: string) => l.startsWith('Tags:'))
      if (tagsLine && !form.tagsInput) form.tagsInput = tagsLine.replace(/^Tags:\s*/, '')
      const mdContent = lines.filter((l: string) => !l.startsWith('Tags:')).join('\n').trim()
      form.content = form.type === 'post' ? (marked.parse(mdContent) as string) : mdContent
      if (!form.excerpt) {
        const plain = form.content.replace(/<[^>]+>/g, '')
        form.excerpt = plain.substring(0, 200)
      }
    } else if (data && typeof data === 'object') {
      if (data.title && !form.title) form.title = data.title
      if (data.slug && !form.slug) form.slug = data.slug
      if (data.content && !form.content) form.content = data.content
      if (data.excerpt && !form.excerpt) form.excerpt = data.excerpt
      if (data.category && !form.category) form.category = data.category
      if (data.readTime && !form.readTime) form.readTime = data.readTime
      if (data.tags && !form.tagsInput) form.tagsInput = data.tags.join(', ')
      if (data.authorName && !form.authorName) form.authorName = data.authorName
      if (data.authorTitle && !form.authorTitle) form.authorTitle = data.authorTitle
      if (data.authorBio && !form.authorBio) form.authorBio = data.authorBio
      if (data.client && !form.client) form.client = data.client
      if (data.industry && !form.industry) form.industry = data.industry
      if (data.location && !form.location) form.location = data.location
      if (data.timeline && !form.timeline) form.timeline = data.timeline
      if (data.challenge && !form.challenge) form.challenge = data.challenge
      if (data.solution && !form.solution) form.solution = data.solution
      if (data.result && !form.result) form.result = data.result
      if (data.services && !form.servicesInput) form.servicesInput = data.services.join(', ')
      if (!form.readTime && form.content) {
        const wordCount = form.content.replace(/<[^>]+>/g, '').split(/\s+/).filter(Boolean).length
        form.readTime = Math.max(1, Math.ceil(wordCount / 200))
      }
    }

    backfillSourceLocaleFromScalars()
    const transMap = res.translations || {}
    if (transMap && typeof transMap === 'object') {
      for (const loc of contentLocales) {
        if (loc === SOURCE_LOCALE) continue
        applyContentTranslationFields(loc, transMap[loc])
      }
    }
  } catch (err: any) {
    formError.value = err?.data?.message || err.message || t('admin.content.ai_generate_failed')
  } finally { aiGenerating.value = false }
}

const fetchContent = async () => {
  pending.value = true; error.value = ''
  try {
    const res = await api.adminGetContent({ type: selectedType.value, page: page.value, limit: pageSize })
    contentList.value = (res.data || []).map((item: any) => ({ ...item, type: selectedType.value }))
    pagination.value = res.pagination
  } catch (err: any) { error.value = err?.message || t('admin.content.load_failed') }
  finally { pending.value = false }
}

const switchType = (type: ContentType) => { selectedType.value = type; page.value = 1 }
const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) page.value++ }
const prevPage = () => { if (page.value > 1) page.value-- }

const openCreateModal = () => {
  editingId.value = ''; formError.value = ''; actionMessage.value = ''; resetForm()
  aiTopic.value = ''; translationLocale.value = SOURCE_LOCALE; showModal.value = true
}

const openCreateDrawer = () => {
  editingId.value = ''; showCreateDrawer.value = true
  resetForm()
  form.type = selectedType.value
  aiContentTopic.value = ''
  aiGeneratingContent.value = false
  aiDraftReady.value = false
  aiPublishing.value = false
  formError.value = ''
  translationLocale.value = SOURCE_LOCALE
}

const fillFormFromGenerated = (data: any) => {
  form.slug = data.slug || ''
  form.category = data.category || ''; form.readTime = data.readTime || 5
  form.tagsInput = Array.isArray(data.tags) ? data.tags.join(', ') : ''
  form.authorAvatar = data.authorAvatar || ''
  form.client = data.client || ''; form.industry = data.industry || ''
  form.location = data.location || ''; form.timeline = data.timeline || ''
  form.servicesInput = Array.isArray(data.services) ? data.services.join(', ') : ''
  form.translations = emptyContentTranslations()
  applyContentTranslationFields(SOURCE_LOCALE, {
    title: data.title,
    excerpt: data.excerpt,
    content: data.content,
    authorName: data.authorName,
    authorTitle: data.authorTitle,
    authorBio: data.authorBio,
    challenge: data.challenge,
    solution: data.solution,
    result: data.result,
  })
  if (!data.readTime && data.content && form.type === 'post') {
    const wc = String(data.content).replace(/<[^>]+>/g, '').split(/\s+/).filter(Boolean).length
    form.readTime = Math.max(1, Math.ceil(wc / 200))
  }
  syncScalarsFromSourceLocale()
}

const buildSourceFieldsForTranslate = () => {
  syncScalarsFromSourceLocale()
  const fields: Record<string, string> = {
    title: form.title.trim(),
    excerpt: form.excerpt.trim(),
    content: sanitize(form.content),
  }
  if (form.type === 'post') {
    if (form.authorName.trim()) fields.authorName = form.authorName.trim()
    if (form.authorTitle.trim()) fields.authorTitle = form.authorTitle.trim()
    if (form.authorBio.trim()) fields.authorBio = form.authorBio.trim()
  } else {
    if (form.challenge.trim()) fields.challenge = sanitize(form.challenge)
    if (form.solution.trim()) fields.solution = sanitize(form.solution)
    if (form.result.trim()) fields.result = sanitize(form.result)
  }
  return fields
}

const applyTranslationsFromApi = (transMap: Record<string, Record<string, string>>) => {
  const targetLocales = contentLocales.filter(l => l !== SOURCE_LOCALE)
  for (const loc of targetLocales) {
    applyContentTranslationFields(loc, transMap[loc] || {})
  }
}

const generateContentAI = async () => {
  if (!aiContentTopic.value.trim() || aiGeneratingContent.value || aiPublishing.value) return
  aiGeneratingContent.value = true
  aiDraftReady.value = false
  formError.value = ''
  try {
    const res = await api.adminAIGenerateContent({
      topic: aiContentTopic.value,
      type: form.type,
      language: SOURCE_LOCALE,
      translate: false,
    })
    const data = res.content
    if (res.parseError || typeof data === 'string') {
      formError.value = t('admin.content.ai_generate_failed')
      return
    }
    if (data && typeof data === 'object') {
      fillFormFromGenerated(data)
    }
    aiDraftReady.value = true
  } catch (err: any) {
    formError.value = err?.message || t('admin.content.ai_generate_failed')
  } finally { aiGeneratingContent.value = false }
}

const publishAiContent = async () => {
  if (!aiDraftReady.value || aiPublishing.value) return
  if (!form.translations[SOURCE_LOCALE]?.title?.trim()) { formError.value = t('admin.content.title_required'); return }
  aiPublishing.value = true
  formError.value = ''
  try {
    const targetLocales = contentLocales.filter(l => l !== SOURCE_LOCALE)
    const transRes = await api.adminAITranslateContentFields({
      contentType: form.type,
      sourceLocale: SOURCE_LOCALE,
      targetLocales,
      fields: buildSourceFieldsForTranslate(),
    })
    if (transRes?.translations && typeof transRes.translations === 'object') {
      applyTranslationsFromApi(transRes.translations)
    }
    const payload = { ...buildPayload(), publishedAt: new Date().toISOString() }
    await api.adminCreateContent(payload)
    actionMessage.value = t('admin.content.ai_published_success')
    actionError.value = false
    showCreateDrawer.value = false
    aiDraftReady.value = false
    await fetchContent()
    refreshNuxtData('blog-posts-all')
  } catch (err: any) {
    formError.value = err?.message || t('admin.content.save_failed')
  } finally { aiPublishing.value = false }
}

const openEditModal = async (item: any) => {
  editingId.value = item.id; formError.value = ''; translationLocale.value = SOURCE_LOCALE
  try {
    const res = await api.adminGetContentById(item.id, item.type)
    const c = res.content || item
    form.type = item.type; form.title = c.title || ''; form.slug = c.slug || ''
    form.thumbnail = c.thumbnail || ''; form.ogImage = c.ogImage || ''; form.category = c.category || ''
    form.readTime = c.readTime || 5; form.excerpt = c.excerpt || ''
    form.tagsInput = Array.isArray(c.tags) ? c.tags.join(', ') : ''
    form.content = c.content || ''
    form.authorName = c.author?.name || ''; form.authorAvatar = c.author?.avatar || ''
    form.authorTitle = c.author?.title || ''; form.authorBio = c.author?.bio || ''
    form.client = c.client || ''; form.industry = c.industry || ''
    form.location = c.location || ''; form.timeline = c.timeline || ''
    form.servicesInput = Array.isArray(c.services) ? c.services.join(', ') : ''
    form.imagesInput = Array.isArray(c.images) ? c.images.join(', ') : ''
    form.challenge = c.challenge || ''; form.solution = c.solution || ''; form.result = c.result || ''
    if (c.translations && typeof c.translations === 'object') {
      form.translations = emptyContentTranslations()
      const src = c.translations
      for (const loc of contentLocales) {
        if (src[loc] && typeof src[loc] === 'object') {
          form.translations[loc].title = src[loc].title || ''
          form.translations[loc].excerpt = src[loc].excerpt || ''
          form.translations[loc].content = src[loc].content || ''
          form.translations[loc].authorName = src[loc].authorName || ''
          form.translations[loc].authorTitle = src[loc].authorTitle || ''
          form.translations[loc].authorBio = src[loc].authorBio || ''
          form.translations[loc].challenge = src[loc].challenge || ''
          form.translations[loc].solution = src[loc].solution || ''
          form.translations[loc].result = src[loc].result || ''
        }
      }
    } else { form.translations = emptyContentTranslations() }
    backfillSourceLocaleFromScalars()
    showModal.value = true
  } catch (err: any) {
    console.warn(`[content] Failed to load detail for content ${item.id}, falling back to list data`)
    actionError.value = true; actionMessage.value = err?.message || t('admin.content.load_failed')
    fillFormFromListItem(item); showModal.value = true
  }
}

function fillFormFromListItem(item: any) {
  form.type = item.type || selectedType.value; form.title = item.title || ''; form.slug = item.slug || ''
  form.thumbnail = item.thumbnail || ''; form.ogImage = item.ogImage || ''; form.category = item.category || ''
  form.readTime = item.readTime || 5; form.excerpt = item.excerpt || ''
  form.tagsInput = Array.isArray(item.tags) ? item.tags.join(', ') : ''
  form.content = item.content || ''
  form.authorName = item.author?.name || ''; form.authorAvatar = item.author?.avatar || ''
  form.authorTitle = item.author?.title || ''; form.authorBio = item.author?.bio || ''
  form.client = item.client || ''; form.industry = item.industry || ''
  form.location = item.location || ''; form.timeline = item.timeline || ''
  form.servicesInput = Array.isArray(item.services) ? item.services.join(', ') : ''
  form.imagesInput = Array.isArray(item.images) ? item.images.join(', ') : ''
  form.challenge = item.challenge || ''; form.solution = item.solution || ''; form.result = item.result || ''
  form.translations = emptyContentTranslations()
  backfillSourceLocaleFromScalars()
}

const closeModal = () => { showModal.value = false; saving.value = false; formError.value = '' }

const aiTranslateContent = async () => {
  if (!editingId.value || aiTranslating.value) return
  aiTranslating.value = true
  try {
    const sourceLocale = SOURCE_LOCALE
    const targetLocales = contentLocales.filter(l => l !== sourceLocale)
    const res = await api.adminAITranslateContent(editingId.value, {
      contentId: editingId.value, contentType: form.type, targetLocales,
    })
    if (res?.translations && typeof res.translations === 'object') {
      for (const loc of targetLocales) {
        applyContentTranslationFields(loc, res.translations[loc])
      }
    }
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.request_failed')
  } finally { aiTranslating.value = false }
}

const buildContentTranslationsPayload = () => {
  const result: Record<string, Record<string, string>> = {}
  for (const loc of contentLocales) {
    const locFields = form.translations[loc]
    if (!locFields) continue
    const entry: Record<string, string> = {}
    if (locFields.title?.trim()) entry.title = locFields.title.trim()
    if (locFields.excerpt?.trim()) entry.excerpt = locFields.excerpt.trim()
    if (locFields.content?.trim()) entry.content = locFields.content.trim()
    if (locFields.authorName?.trim()) entry.authorName = locFields.authorName.trim()
    if (locFields.authorTitle?.trim()) entry.authorTitle = locFields.authorTitle.trim()
    if (locFields.authorBio?.trim()) entry.authorBio = locFields.authorBio.trim()
    if (locFields.challenge?.trim()) entry.challenge = locFields.challenge.trim()
    if (locFields.solution?.trim()) entry.solution = locFields.solution.trim()
    if (locFields.result?.trim()) entry.result = locFields.result.trim()
    if (Object.keys(entry).length > 0) result[loc] = entry
  }
  return Object.keys(result).length > 0 ? result : null
}

const buildPayload = () => {
  syncScalarsFromSourceLocale()
  const translations = buildContentTranslationsPayload()
  const safeContent = sanitize(form.content)
  const safeChallenge = sanitize(form.challenge)
  const safeSolution = sanitize(form.solution)
  const safeResult = sanitize(form.result)
  if (form.type === 'post') {
    return { type: 'post', title: form.title, slug: form.slug, thumbnail: form.thumbnail, ogImage: form.ogImage, category: form.category, readTime: form.readTime, excerpt: form.excerpt, tags: parseCSV(form.tagsInput), content: safeContent, author: { name: form.authorName, avatar: form.authorAvatar, title: form.authorTitle, bio: form.authorBio }, ...(translations ? { translations } : {}) }
  }
  return { type: 'case', title: form.title, slug: form.slug, thumbnail: form.thumbnail, ogImage: form.ogImage, client: form.client, industry: form.industry, location: form.location, timeline: form.timeline, services: parseCSV(form.servicesInput), images: parseCSV(form.imagesInput), challenge: safeChallenge, solution: safeSolution, result: safeResult, content: safeContent, ...(translations ? { translations } : {}) }
}

const saveContent = async () => {
  if (!form.translations[SOURCE_LOCALE]?.title?.trim()) { formError.value = t('admin.content.title_required'); return }
  saving.value = true; formError.value = ''
  try {
    if (editingId.value) {
      await api.adminUpdateContent(editingId.value, buildPayload())
      actionMessage.value = t('admin.content.updated_success')
    } else {
      await api.adminCreateContent(buildPayload())
      actionMessage.value = t('admin.content.created_success')
    }
    showModal.value = false
    showCreateDrawer.value = false
    aiDraftReady.value = false
    await fetchContent()
    refreshNuxtData('blog-posts-all')
  } catch (err: any) { formError.value = err?.message || t('admin.content.save_failed') }
  finally { saving.value = false }
}

const deleteContent = async (item: any) => {
  if (!confirm(t('admin.content.confirm_delete'))) return
  try {
    await api.adminDeleteContent(item.id, item.type)
    actionMessage.value = t('admin.content.deleted_success'); await fetchContent()
    refreshNuxtData('blog-posts-all')
  } catch (err: any) { actionError.value = true; actionMessage.value = err?.message || t('admin.content.delete_failed') }
}

watch([selectedType, page], fetchContent, { immediate: true })
</script>
