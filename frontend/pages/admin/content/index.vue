<template>
  <div>
    <!-- Header -->
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.content.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.content.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0 flex items-center gap-3">
        <button type="button" class="inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50" @click="openCreateModal">
          {{ t('admin.content.add_content') }}
        </button>
        <button type="button" class="inline-flex items-center gap-2 rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700" @click="openCreateDrawer">
          <Icon name="heroicons:sparkles" class="h-4 w-4" />
          {{ t('admin.content.ai_import') }}
        </button>
      </div>
    </div>

    <!-- Type Tabs -->
    <div class="mt-6 flex items-center gap-2">
      <button v-for="option in typeOptions" :key="option.value" type="button" class="rounded-md px-3 py-2 text-sm font-medium" :class="selectedType === option.value ? 'bg-orange-600 text-white' : 'bg-white text-gray-700 ring-1 ring-inset ring-gray-300 hover:bg-gray-50'" @click="switchType(option.value)">
        {{ option.value === 'post' ? t('admin.content.tab_posts') : t('admin.content.tab_cases') }}
      </button>
    </div>

    <!-- Table -->
    <div class="mt-6 overflow-hidden rounded-lg bg-white shadow ring-1 ring-black ring-opacity-5">
      <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
          <tr>
            <th class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">{{ t('admin.content.col_title') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.content.col_type') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.content.col_meta') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ t('admin.content.col_updated') }}</th>
            <th class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">{{ $t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <tr v-if="pending"><td colspan="5" class="py-5 text-center text-sm text-gray-500">{{ t('admin.content.loading') }}</td></tr>
          <tr v-else-if="error"><td colspan="5" class="py-5 text-center text-sm text-red-600">{{ error }}</td></tr>
          <tr v-else-if="contentList.length === 0"><td colspan="5" class="py-5 text-center text-sm text-gray-500">{{ t('admin.content.no_data') }}</td></tr>
          <tr v-else v-for="item in contentList" :key="item.id">
            <td class="py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">{{ item.title }}</td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
              <span class="inline-flex rounded-full px-2 text-xs font-semibold leading-5" :class="item.type === 'post' ? 'bg-amber-100 text-amber-800' : 'bg-cyan-100 text-cyan-800'">{{ enumLabel('content_type', item.type) }}</span>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ item.type === 'post' ? (item.category || t('admin.content.category_general')) : (item.industry || t('admin.content.category_general')) }}</td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ formatDate(item.updatedAt, { dateStyle: 'medium', timeStyle: 'short' }) }}</td>
            <td class="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
              <button type="button" class="text-orange-600 hover:text-orange-900" @click="openEditModal(item)">{{ t('admin.content.edit') }}</button>
              <button type="button" class="ml-4 text-red-600 hover:text-red-900" @click="deleteContent(item)">{{ t('admin.content.delete') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="pagination" class="mt-4 flex items-center justify-between rounded-lg border-t border-gray-200 bg-white px-4 py-3 shadow sm:px-6">
      <div class="text-sm text-gray-700">{{ t('admin.content.showing', { from: ((page - 1) * pageSize) + 1, to: Math.min(page * pageSize, pagination.total), total: pagination.total }) }}</div>
      <div class="flex items-center gap-2">
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page <= 1" @click="prevPage">{{ t('admin.content.previous') }}</button>
        <button type="button" class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-700 disabled:opacity-50" :disabled="page >= pagination.totalPages" @click="nextPage">{{ t('admin.content.next') }}</button>
      </div>
    </div>

    <!-- Editor Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto p-4">
      <button type="button" class="fixed inset-0 w-full h-full bg-gray-500/75 border-0 cursor-pointer" @click="closeModal" :aria-label="t('close')" />
      <div class="relative my-8 w-full max-w-6xl rounded-lg bg-white shadow-xl" role="dialog" aria-modal="true">
        <!-- Modal Header -->
        <div class="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h3 class="text-lg font-semibold text-gray-900">{{ editingId ? t('admin.content.edit_content') : t('admin.content.create_content') }}</h3>
          <button type="button" class="text-gray-400 hover:text-gray-600" @click="closeModal" :aria-label="t('close')">
            <Icon name="heroicons:x-mark" class="h-6 w-6" aria-hidden="true" />
          </button>
        </div>

        <form @submit.prevent="saveContent">
          <div class="px-6 py-4 space-y-4">
            <ContentFormFields
              :form="form"
              :disable-type="Boolean(editingId)"
              :editing-id="editingId"
              :ai-translating="aiTranslating"
              :translation-locale="translationLocale"
              @ai-translate="aiTranslateContent"
              @update:translation-locale="translationLocale = $event"
            />

            <!-- AI Generate Bar (edit modal only) -->
            <div class="flex items-center gap-3 rounded-lg bg-amber-50 border border-amber-200 p-3">
              <Icon name="heroicons:sparkles" class="h-5 w-5 text-amber-600 flex-shrink-0" aria-hidden="true" />
              <label for="content-aiTopic" class="sr-only">{{ t('admin.content.ai_topic_placeholder') }}</label>
              <input id="content-aiTopic" v-model="aiTopic" name="aiTopic" type="text" :placeholder="form.type === 'post' ? t('admin.content.ai_topic_placeholder') : t('admin.content.ai_case_topic_placeholder')" class="flex-1 rounded-md border border-amber-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" @keydown.enter.prevent="generateWithAI" />
              <button type="button" :disabled="aiGenerating || !aiTopic.trim()" class="inline-flex items-center gap-1.5 rounded-md bg-orange-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50" @click="generateWithAI">
                <Icon v-if="aiGenerating" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" aria-hidden="true" />
                <Icon v-else name="heroicons:sparkles" class="h-4 w-4" aria-hidden="true" />
                {{ aiGenerating ? t('admin.content.ai_generating') : t('admin.content.ai_generate') }}
              </button>
            </div>

            <div v-if="formError" class="text-sm text-red-600">{{ formError }}</div>
          </div>

          <!-- Modal Footer -->
          <div class="flex justify-end gap-3 border-t border-gray-200 px-6 py-4">
            <button type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700" @click="closeModal">{{ t('admin.content.cancel') }}</button>
            <button type="submit" :disabled="saving" class="rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white disabled:opacity-50 hover:bg-orange-700">
              {{ saving ? t('admin.content.saving') : (editingId ? t('admin.content.update') : t('admin.content.create')) }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- AI Generate Drawer -->
    <Drawer :open="showCreateDrawer" width="2xl" @close="showCreateDrawer = false">
      <template #title>{{ t('admin.content.ai_create_drawer_title') }}</template>

      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6" style="min-height: 60vh;">
        <!-- LEFT -->
        <div class="lg:col-span-5 space-y-4">
          <div>
            <label for="ai-content-type" class="block text-sm font-medium text-gray-700">{{ t('admin.content.type') }}</label>
            <select id="ai-content-type" v-model="form.type" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
              <option value="post">{{ enumLabel('content_type', 'post') }}</option>
              <option value="case">{{ enumLabel('content_type', 'case') }}</option>
            </select>
          </div>
          <div>
            <label for="ai-content-topic" class="block text-sm font-medium text-gray-700">{{ t('admin.content.ai_topic_label') }}</label>
            <textarea id="ai-content-topic" v-model="aiContentTopic" rows="6" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="form.type === 'post' ? t('admin.content.ai_topic_placeholder') : t('admin.content.ai_case_topic_placeholder')"></textarea>
          </div>
          <div>
            <label for="ai-content-lang" class="block text-sm font-medium text-gray-700">{{ t('admin.content.ai_language_label') }}</label>
            <select id="ai-content-lang" v-model="aiContentLanguage" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm">
              <option v-for="loc in contentLocales" :key="loc" :value="loc">{{ localeTabLabel(loc) }}</option>
            </select>
          </div>
          <button type="button" :disabled="aiGeneratingContent || !aiContentTopic.trim()" class="inline-flex items-center gap-2 rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50" @click="generateContentAI">
            <Icon name="heroicons:sparkles" class="h-4 w-4" />
            {{ aiGeneratingContent ? t('admin.content.ai_generating') : t('admin.content.ai_generate') }}
          </button>
          <p v-if="formError" class="text-sm text-red-600">{{ formError }}</p>
        </div>

        <!-- RIGHT: Existing form template auto-filled by AI -->
        <div class="lg:col-span-7 lg:border-l lg:pl-6 overflow-y-auto" style="max-height: 65vh;">
          <div v-if="!aiHasContent && !aiGeneratingContent" class="flex flex-col items-center justify-center h-48 text-gray-400 text-sm">
            <Icon name="heroicons:sparkles" class="h-10 w-10 mb-2 text-gray-300" />
            {{ t('admin.content.ai_empty_preview') }}
          </div>

          <div v-if="aiGeneratingContent" class="flex flex-col items-center justify-center h-48 text-gray-500 text-sm gap-2">
            <svg class="animate-spin h-6 w-6 text-orange-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
            {{ t('admin.content.ai_generating') }}
          </div>

          <ContentFormFields v-if="aiHasContent" :form="form" :disable-type="false" :editing-id="''" :ai-translating="false" :translation-locale="translationLocale" @update:translation-locale="translationLocale = $event" />
        </div>
      </div>

      <template #footer>
        <p v-if="formError" class="mb-3 text-sm text-red-600">{{ formError }}</p>
        <div class="flex justify-end gap-3">
          <button type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700" @click="showCreateDrawer = false">{{ t('admin.content.cancel') }}</button>
          <button type="button" :disabled="saving || !aiHasContent" class="inline-flex items-center gap-2 rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50" @click="saveContent()">
            <Icon name="heroicons:sparkles" class="h-4 w-4" />
            {{ saving ? t('admin.content.saving') : t('admin.content.create') }}
          </button>
        </div>
      </template>
    </Drawer>

    <p v-if="actionMessage" class="mt-4 text-sm" :class="actionError ? 'text-red-600' : 'text-green-600'">{{ actionMessage }}</p>
  </div>
</template>


<script setup lang="ts">
import { reactive, ref, watch, onMounted } from 'vue'
import { marked } from 'marked'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const api = useApi()
const { t, locale } = useI18n()
const { enumLabel, formatDate } = useDisplay()

const typeOptions = [{ value: 'post' }, { value: 'case' }] as const
type ContentType = 'post' | 'case'

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

// AI state (edit modal - legacy)
const aiTopic = ref('')
const aiGenerating = ref(false)
const aiTranslating = ref(false)
const translationLocale = ref('en')

// AI Drawer state
const showCreateDrawer = ref(false)
const aiContentTopic = ref('')
const aiContentLanguage = ref('zh')
const aiGeneratingContent = ref(false)
const aiHasContent = ref(false)
const contentLocales = ['en', 'zh', 'ko', 'ar', 'ja', 'th', 'vi', 'id', 'ms']

const localeTabLabelMap: Record<string, string> = {
  en: 'admin.products.locale_tab_en',
  zh: 'admin.products.locale_tab_zh',
  ko: 'admin.products.locale_tab_ko',
  ar: 'admin.products.locale_tab_ar',
  ja: 'admin.products.locale_tab_ja',
  th: 'admin.products.locale_tab_th',
  vi: 'admin.products.locale_tab_vi',
  id: 'admin.products.locale_tab_id',
  ms: 'admin.products.locale_tab_ms',
}
const localeTabLabel = (loc: string) => t(localeTabLabelMap[loc] || loc)

function emptyContentTranslations() {
  const result: Record<string, { title: string; excerpt: string; content: string }> = {}
  for (const loc of contentLocales) {
    result[loc] = { title: '', excerpt: '', content: '' }
  }
  return result
}

const form = reactive({
  type: 'post' as ContentType,
  title: '', slug: '', thumbnail: '', category: '', readTime: 5,
  excerpt: '', tagsInput: '', content: '',
  authorName: '', authorAvatar: '', authorTitle: '', authorBio: '',
  client: '', industry: '', location: '', timeline: '',
  servicesInput: '', imagesInput: '', challenge: '', solution: '', result: '',
  translations: emptyContentTranslations()
})

const parseCSV = (v: string) => v.split(',').map(s => s.trim()).filter(Boolean)

const resetForm = () => {
  form.type = selectedType.value; form.title = ''; form.slug = ''; form.thumbnail = ''
  form.category = ''; form.readTime = 5; form.excerpt = ''; form.tagsInput = ''; form.content = ''
  form.authorName = ''; form.authorAvatar = ''; form.authorTitle = ''; form.authorBio = ''
  form.client = ''; form.industry = ''; form.location = ''; form.timeline = ''
  form.servicesInput = ''; form.imagesInput = ''; form.challenge = ''; form.solution = ''; form.result = ''
  form.translations = emptyContentTranslations()
}

// --- AI Generate ---
const generateWithAI = async () => {
  if (!aiTopic.value.trim() || aiGenerating.value) return
  aiGenerating.value = true
  formError.value = ''
  try {
    const res = await api.adminAIGenerateContent({ topic: aiTopic.value, type: form.type, language: locale.value })
    const data = res.content

    // Fallback: if AI returned raw text (parse error on backend), try old markdown parsing
    if (res.parseError || typeof data === 'string') {
      const raw = typeof data === 'string' ? data : ''
      const lines = raw.split('\n')
      const titleLine = lines.find((l: string) => l.startsWith('# '))
      if (titleLine && !form.title) form.title = titleLine.replace(/^#\s+/, '')
      const tagsLine = lines.find((l: string) => l.startsWith('Tags:'))
      if (tagsLine && !form.tagsInput) form.tagsInput = tagsLine.replace(/^Tags:\s*/, '')
      const mdContent = lines.filter((l: string) => !l.startsWith('Tags:')).join('\n').trim()
      if (form.type === 'post') {
        form.content = marked.parse(mdContent) as string
      } else {
        form.content = mdContent
      }
      if (!form.excerpt) {
        const plain = form.content.replace(/<[^>]+>/g, '')
        form.excerpt = plain.substring(0, 200)
      }
    } else if (data && typeof data === 'object') {
      // Structured response: fill all fields, respecting existing manual edits
      if (data.title && !form.title) form.title = data.title
      if (data.slug && !form.slug) form.slug = data.slug
      if (data.content && !form.content) form.content = data.content
      if (data.excerpt && !form.excerpt) form.excerpt = data.excerpt
      if (data.category && !form.category) form.category = data.category
      if (data.readTime && !form.readTime) form.readTime = data.readTime
      if (data.tags && !form.tagsInput) form.tagsInput = data.tags.join(', ')
      // Author fields (post only)
      if (data.authorName && !form.authorName) form.authorName = data.authorName
      if (data.authorTitle && !form.authorTitle) form.authorTitle = data.authorTitle
      if (data.authorBio && !form.authorBio) form.authorBio = data.authorBio
      // Case-specific fields
      if (data.client && !form.client) form.client = data.client
      if (data.industry && !form.industry) form.industry = data.industry
      if (data.location && !form.location) form.location = data.location
      if (data.timeline && !form.timeline) form.timeline = data.timeline
      if (data.challenge && !form.challenge) form.challenge = data.challenge
      if (data.solution && !form.solution) form.solution = data.solution
      if (data.result && !form.result) form.result = data.result
      if (data.services && !form.servicesInput) form.servicesInput = data.services.join(', ')

      // Auto-calculate readTime from content length if AI didn't provide it
      if (!form.readTime && form.content) {
        const wordCount = form.content.replace(/<[^>]+>/g, '').split(/\s+/).filter(Boolean).length
        form.readTime = Math.max(1, Math.ceil(wordCount / 200))
      }
    }
  } catch (err: any) {
    formError.value = err?.data?.message || err.message || t('admin.content.ai_generate_failed')
  } finally { aiGenerating.value = false }
}

// --- CRUD ---
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
  aiTopic.value = ''; translationLocale.value = 'en'; showModal.value = true
}

const openCreateDrawer = () => {
  showCreateDrawer.value = true
  resetForm()
  form.type = selectedType.value
  aiContentTopic.value = ''
  aiContentLanguage.value = 'zh'
  aiGeneratingContent.value = false
  aiHasContent.value = false
  formError.value = ''
  translationLocale.value = 'en'
}

const generateContentAI = async () => {
  if (!aiContentTopic.value.trim() || aiGeneratingContent.value) return
  aiGeneratingContent.value = true
  formError.value = ''
  try {
    const res = await api.adminAIGenerateContent({
      topic: aiContentTopic.value,
      type: form.type,
      language: aiContentLanguage.value,
    })
    const data = res.content
    if (res.parseError || typeof data === 'string') {
      formError.value = t('admin.content.ai_generate_failed')
      return
    }
    if (data && typeof data === 'object') {
      form.title = data.title || ''
      form.slug = data.slug || ''
      form.content = data.content || ''
      form.excerpt = data.excerpt || ''
      form.category = data.category || ''
      form.readTime = data.readTime || 5
      form.tagsInput = Array.isArray(data.tags) ? data.tags.join(', ') : ''
      form.authorName = data.authorName || ''
      form.authorTitle = data.authorTitle || ''
      form.authorBio = data.authorBio || ''
      form.client = data.client || ''
      form.industry = data.industry || ''
      form.location = data.location || ''
      form.timeline = data.timeline || ''
      form.challenge = data.challenge || ''
      form.solution = data.solution || ''
      form.result = data.result || ''
      form.servicesInput = Array.isArray(data.services) ? data.services.join(', ') : ''
      if (!data.readTime && data.content && form.type === 'post') {
        const wc = data.content.replace(/<[^>]+>/g, '').split(/\s+/).filter(Boolean).length
        form.readTime = Math.max(1, Math.ceil(wc / 200))
      }
      // Fill AI translations
      const transMap = res.translations || {}
      form.translations = emptyContentTranslations()
      // Populate source locale from generated content
      const sourceLoc = aiContentLanguage.value
      if (sourceLoc && form.translations[sourceLoc]) {
        form.translations[sourceLoc].title = form.title
        form.translations[sourceLoc].excerpt = form.excerpt
        form.translations[sourceLoc].content = form.content
      }
      // Populate translations from backend
      for (const loc of contentLocales) {
        if (transMap[loc] && typeof transMap[loc] === 'object') {
          form.translations[loc].title = transMap[loc].title || form.translations[loc].title
          form.translations[loc].excerpt = transMap[loc].excerpt || form.translations[loc].excerpt
          form.translations[loc].content = transMap[loc].content || form.translations[loc].content
        }
      }
    }
    aiHasContent.value = true
  } catch (err: any) {
    formError.value = err?.message || t('admin.content.ai_generate_failed')
  } finally {
    aiGeneratingContent.value = false
  }
}

const openEditModal = async (item: any) => {
  editingId.value = item.id; formError.value = ''; translationLocale.value = 'en'
  try {
    const res = await api.adminGetContentById(item.id, item.type)
    const c = res.content || item
    form.type = item.type; form.title = c.title || ''; form.slug = c.slug || ''
    form.thumbnail = c.thumbnail || ''; form.category = c.category || ''
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
        }
      }
    } else {
      form.translations = emptyContentTranslations()
    }
    showModal.value = true
  } catch (err: any) { actionError.value = true; actionMessage.value = err?.message || t('admin.content.load_failed') }
}

const closeModal = () => { showModal.value = false; saving.value = false; formError.value = '' }

const aiTranslateContent = async () => {
  if (!editingId.value || aiTranslating.value) return
  aiTranslating.value = true
  try {
    const targetLocales = contentLocales.filter(l => l !== 'zh')
    const res = await api.post<any>(`/admin/content/${editingId.value}/ai-translate`, {
      contentId: editingId.value,
      contentType: form.type,
      targetLocales,
    })
    if (res?.translations && typeof res.translations === 'object') {
      for (const loc of targetLocales) {
        const fields = res.translations[loc]
        if (fields && typeof fields === 'object') {
          if (fields.title) form.translations[loc].title = fields.title
          if (fields.excerpt) form.translations[loc].excerpt = fields.excerpt
          if (fields.content) form.translations[loc].content = fields.content
        }
      }
    }
  } catch (err: any) {
    formError.value = err?.message || t('errors.api.request_failed')
  } finally {
    aiTranslating.value = false
  }
}

const buildContentTranslationsPayload = () => {
  const result: Record<string, Record<string, string>> = {}
  for (const loc of contentLocales) {
    const t = form.translations[loc]
    if (!t) continue
    const entry: Record<string, string> = {}
    if (t.title?.trim()) entry.title = t.title.trim()
    if (t.excerpt?.trim()) entry.excerpt = t.excerpt.trim()
    if (t.content?.trim()) entry.content = t.content.trim()
    if (Object.keys(entry).length > 0) result[loc] = entry
  }
  return Object.keys(result).length > 0 ? result : null
}

const buildPayload = () => {
  const translations = buildContentTranslationsPayload()
  if (form.type === 'post') {
    return { type: 'post', title: form.title, slug: form.slug, thumbnail: form.thumbnail, category: form.category, readTime: form.readTime, excerpt: form.excerpt, tags: parseCSV(form.tagsInput), content: form.content, author: { name: form.authorName, avatar: form.authorAvatar, title: form.authorTitle, bio: form.authorBio }, ...(translations ? { translations } : {}) }
  }
  return { type: 'case', title: form.title, slug: form.slug, thumbnail: form.thumbnail, client: form.client, industry: form.industry, location: form.location, timeline: form.timeline, services: parseCSV(form.servicesInput), images: parseCSV(form.imagesInput), challenge: form.challenge, solution: form.solution, result: form.result, content: form.content, ...(translations ? { translations } : {}) }
}

const saveContent = async () => {
  if (!form.title.trim()) { formError.value = t('admin.content.title_required'); return }
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
    aiHasContent.value = false
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

watch([selectedType, page], fetchContent)
onMounted(fetchContent)
</script>
