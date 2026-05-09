<template>
  <div>
    <!-- Header -->
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.content.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.content.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0">
        <button type="button" class="inline-flex items-center rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-orange-700" @click="openCreateModal">
          {{ t('admin.content.add_content') }}
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
            <th class="relative py-3.5 pl-3 pr-4 sm:pr-6"><span class="sr-only">Actions</span></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 bg-white">
          <tr v-if="pending"><td colspan="5" class="py-5 text-center text-sm text-gray-500">{{ t('admin.content.loading') }}</td></tr>
          <tr v-else-if="error"><td colspan="5" class="py-5 text-center text-sm text-red-600">{{ error }}</td></tr>
          <tr v-else-if="contentList.length === 0"><td colspan="5" class="py-5 text-center text-sm text-gray-500">{{ t('admin.content.no_data') }}</td></tr>
          <tr v-else v-for="item in contentList" :key="item.id">
            <td class="py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">{{ item.title }}</td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
              <span class="inline-flex rounded-full px-2 text-xs font-semibold leading-5" :class="item.type === 'post' ? 'bg-amber-100 text-amber-800' : 'bg-cyan-100 text-cyan-800'">{{ item.type }}</span>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">{{ item.type === 'post' ? (item.category || t('admin.content.category_general')) : (item.industry || t('admin.content.category_general')) }}</td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ item.updatedAt ? new Date(item.updatedAt).toLocaleString() : '-' }}</td>
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
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-gray-500/75 p-4" @click.self="closeModal">
      <div class="my-8 w-full max-w-6xl rounded-lg bg-white shadow-xl">
        <!-- Modal Header -->
        <div class="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h3 class="text-lg font-semibold text-gray-900">{{ editingId ? t('admin.content.edit_content') : t('admin.content.create_content') }}</h3>
          <button type="button" class="text-gray-400 hover:text-gray-600" @click="closeModal">
            <Icon name="heroicons:x-mark" class="h-6 w-6" />
          </button>
        </div>

        <form @submit.prevent="saveContent">
          <div class="px-6 py-4 space-y-4">
            <!-- Meta fields row -->
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.type') }}</label>
                <select v-model="form.type" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :disabled="Boolean(editingId)">
                  <option value="post">Post</option>
                  <option value="case">Case</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.content_title') }}</label>
                <input v-model="form.title" type="text" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.content_slug') }}</label>
                <input v-model="form.slug" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
            </div>

            <!-- Post-specific meta -->
            <div v-if="form.type === 'post'" class="grid grid-cols-1 gap-4 sm:grid-cols-4">
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_category') }}</label>
                <input v-model="form.category" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.content_thumbnail') }}</label>
                <input v-model="form.thumbnail" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" placeholder="URL or upload" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_read_time') }}</label>
                <input v-model.number="form.readTime" type="number" min="1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_tags') }}</label>
                <input v-model="form.tagsInput" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
            </div>

            <!-- Case-specific meta -->
            <div v-if="form.type === 'case'" class="grid grid-cols-1 gap-4 sm:grid-cols-4">
              <div><label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_client') }}</label><input v-model="form.client" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" /></div>
              <div><label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_industry') }}</label><input v-model="form.industry" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" /></div>
              <div><label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_location') }}</label><input v-model="form.location" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" /></div>
              <div><label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_timeline') }}</label><input v-model="form.timeline" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" /></div>
            </div>

            <!-- Excerpt (post only) -->
            <div v-if="form.type === 'post'">
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_excerpt') }}</label>
              <textarea v-model="form.excerpt" rows="2" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>

            <!-- AI Generate Bar -->
            <div v-if="form.type === 'post'" class="flex items-center gap-3 rounded-lg bg-amber-50 border border-amber-200 p-3">
              <Icon name="heroicons:sparkles" class="h-5 w-5 text-amber-600 flex-shrink-0" />
              <input v-model="aiTopic" type="text" :placeholder="t('admin.content.ai_topic_placeholder')" class="flex-1 rounded-md border border-amber-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" @keydown.enter.prevent="generateWithAI" />
              <button type="button" :disabled="aiGenerating || !aiTopic.trim()" class="inline-flex items-center gap-1.5 rounded-md bg-orange-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50" @click="generateWithAI">
                <Icon v-if="aiGenerating" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" />
                <Icon v-else name="heroicons:sparkles" class="h-4 w-4" />
                {{ aiGenerating ? t('admin.content.ai_generating') : t('admin.content.ai_generate') }}
              </button>
            </div>

            <!-- Markdown Editor + Preview -->
            <div class="editor-container">
              <!-- Toolbar -->
              <div class="flex items-center gap-1 border border-gray-300 rounded-t-md bg-gray-50 px-2 py-1.5">
                <button type="button" class="toolbar-btn" title="Bold" @click="insertMd('**', '**')"><Icon name="heroicons:bold" class="h-4 w-4" /></button>
                <button type="button" class="toolbar-btn" title="Italic" @click="insertMd('*', '*')"><Icon name="heroicons:italic" class="h-4 w-4" /></button>
                <span class="mx-1 h-4 w-px bg-gray-300"></span>
                <button type="button" class="toolbar-btn" title="H2" @click="insertMd('\n## ', '\n')">H2</button>
                <button type="button" class="toolbar-btn" title="H3" @click="insertMd('\n### ', '\n')">H3</button>
                <span class="mx-1 h-4 w-px bg-gray-300"></span>
                <button type="button" class="toolbar-btn" title="Bullet List" @click="insertMd('\n- ', '\n')"><Icon name="heroicons:list-bullet" class="h-4 w-4" /></button>
                <button type="button" class="toolbar-btn" title="Link" @click="insertMd('[', '](url)')"><Icon name="heroicons:link" class="h-4 w-4" /></button>
                <button type="button" class="toolbar-btn" title="Image" @click="showImageInsert = !showImageInsert"><Icon name="heroicons:photo" class="h-4 w-4" /></button>
                <span class="mx-1 h-4 w-px bg-gray-300"></span>
                <button type="button" class="toolbar-btn" :disabled="convertingToMd" :title="t('admin.content.convert_to_md')" @click="convertToMarkdown">
                  <Icon v-if="convertingToMd" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" />
                  <Icon v-else name="heroicons:sparkles" class="h-4 w-4" />
                  {{ t('admin.content.convert_to_md') }}
                </button>
                <button type="button" class="toolbar-btn" :class="showPreview ? 'bg-orange-100 text-orange-700' : ''" @click="showPreview = !showPreview">
                  <Icon name="heroicons:eye" class="h-4 w-4" />
                  {{ t('admin.content.preview') }}
                </button>
              </div>

              <!-- Image Insert Bar -->
              <div v-if="showImageInsert" class="flex items-center gap-2 border-x border-gray-300 bg-gray-50 px-3 py-2">
                <input v-model="imageUrl" type="text" placeholder="Image URL or upload..." class="flex-1 rounded border border-gray-300 px-2 py-1 text-sm" />
                <label class="inline-flex items-center gap-1 cursor-pointer rounded bg-gray-200 px-2 py-1 text-xs font-medium hover:bg-gray-300">
                  <Icon name="heroicons:arrow-up-tray" class="h-3.5 w-3.5" />
                  Upload
                  <input type="file" accept="image/*" class="hidden" @change="handleImageUpload" />
                </label>
                <button type="button" class="rounded bg-orange-600 px-2 py-1 text-xs font-medium text-white hover:bg-orange-700" :disabled="!imageUrl" @click="insertImage">Insert</button>
                <span v-if="imageUploading" class="text-xs text-gray-500">Uploading...</span>
              </div>

              <!-- Editor + Preview Split -->
              <div class="flex" :class="showPreview ? 'editor-split' : ''">
                <textarea ref="editorRef" v-model="form.content" :rows="form.type === 'post' ? 18 : 8" class="editor-textarea" :class="showPreview ? 'w-1/2 rounded-bl-md border-r-0' : 'w-full rounded-b-md'" :placeholder="form.type === 'post' ? 'Write your article in Markdown...' : ''"></textarea>
                <div v-if="showPreview" class="w-1/2 overflow-y-auto border border-gray-300 rounded-br-md bg-white p-4 prose prose-sm max-w-none" style="max-height: 500px" v-html="renderedContent"></div>
              </div>
            </div>

            <!-- Case-specific fields below editor -->
            <template v-if="form.type === 'case'">
              <div><label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_services') }}</label><input v-model="form.servicesInput" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" /></div>
              <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
                <div><label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_challenge') }}</label><textarea v-model="form.challenge" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea></div>
                <div><label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_solution') }}</label><textarea v-model="form.solution" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea></div>
                <div><label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_result') }}</label><textarea v-model="form.result" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea></div>
              </div>
            </template>

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

    <p v-if="actionMessage" class="mt-4 text-sm" :class="actionError ? 'text-red-600' : 'text-green-600'">{{ actionMessage }}</p>
  </div>
</template>


<script setup lang="ts">
import { reactive, ref, computed, watch, onMounted, nextTick } from 'vue'
import { marked } from 'marked'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const { token } = useAuth()
const { t, locale } = useI18n()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'

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

// Editor state
const showPreview = ref(true)
const showImageInsert = ref(false)
const imageUrl = ref('')
const imageUploading = ref(false)
const editorRef = ref<HTMLTextAreaElement | null>(null)

// AI state
const aiTopic = ref('')
const aiGenerating = ref(false)
const convertingToMd = ref(false)

const form = reactive({
  type: 'post' as ContentType,
  title: '', slug: '', thumbnail: '', category: '', readTime: 5,
  excerpt: '', tagsInput: '', content: '',
  client: '', industry: '', location: '', timeline: '',
  servicesInput: '', challenge: '', solution: '', result: ''
})

// Markdown preview
const renderedContent = computed(() => {
  if (!form.content) return '<p style="color:#9ca3af">Preview will appear here...</p>'
  try { return marked.parse(form.content) } catch { return form.content }
})

const parseCSV = (v: string) => v.split(',').map(s => s.trim()).filter(Boolean)

const resetForm = () => {
  form.type = selectedType.value; form.title = ''; form.slug = ''; form.thumbnail = ''
  form.category = ''; form.readTime = 5; form.excerpt = ''; form.tagsInput = ''; form.content = ''
  form.client = ''; form.industry = ''; form.location = ''; form.timeline = ''
  form.servicesInput = ''; form.challenge = ''; form.solution = ''; form.result = ''
}

// --- Toolbar helpers ---
const insertMd = (before: string, after: string) => {
  const el = editorRef.value
  if (!el) return
  const start = el.selectionStart, end = el.selectionEnd
  const selected = form.content.substring(start, end) || 'text'
  form.content = form.content.substring(0, start) + before + selected + after + form.content.substring(end)
  nextTick(() => { el.focus(); el.setSelectionRange(start + before.length, start + before.length + selected.length) })
}

const insertImage = () => {
  if (!imageUrl.value) return
  const md = `\n![image](${imageUrl.value})\n`
  const el = editorRef.value
  if (el) {
    const pos = el.selectionStart
    form.content = form.content.substring(0, pos) + md + form.content.substring(pos)
  } else {
    form.content += md
  }
  imageUrl.value = ''
  showImageInsert.value = false
}

const handleImageUpload = async (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  imageUploading.value = true
  try {
    const fd = new FormData()
    fd.append('file', file)
    fd.append('folder', 'content')
    const res = await $fetch<any>(`${baseURL}/admin/upload/image`, {
      method: 'POST', headers: { Authorization: `Bearer ${token.value}` }, body: fd
    })
    imageUrl.value = res.url || res.filename || ''
  } catch { imageUrl.value = '' }
  finally { imageUploading.value = false; (e.target as HTMLInputElement).value = '' }
}

// --- AI Generate ---
const generateWithAI = async () => {
  if (!aiTopic.value.trim() || aiGenerating.value) return
  aiGenerating.value = true
  formError.value = ''
  try {
    const res = await $fetch<any>(`${baseURL}/admin/content/ai-generate`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token.value}` },
      body: { topic: aiTopic.value, type: form.type, language: locale.value }
    })
    const raw = res.content || ''
    // Parse AI output: extract title, excerpt, tags, content
    const lines = raw.split('\n')
    let titleLine = lines.find((l: string) => l.startsWith('# '))
    if (titleLine && !form.title) form.title = titleLine.replace(/^#\s+/, '')
    // Extract tags
    const tagsLine = lines.find((l: string) => l.startsWith('Tags:'))
    if (tagsLine && !form.tagsInput) form.tagsInput = tagsLine.replace(/^Tags:\s*/, '')
    // Set content (remove the Tags line)
    form.content = lines.filter((l: string) => !l.startsWith('Tags:')).join('\n').trim()
    // Extract first paragraph as excerpt
    const contentLines = form.content.split('\n').filter((l: string) => l.trim() && !l.startsWith('#'))
    if (contentLines.length && !form.excerpt) form.excerpt = contentLines[0].substring(0, 200)
  } catch (err: any) {
    formError.value = err?.data?.message || err.message || 'AI generation failed'
  } finally { aiGenerating.value = false }
}

// Convert plain text content to Markdown via AI
const convertToMarkdown = async () => {
  if (!form.content.trim() || convertingToMd.value) return
  convertingToMd.value = true
  formError.value = ''
  try {
    const res = await $fetch<any>(`${baseURL}/admin/content/ai-generate`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token.value}` },
      body: {
        topic: `Convert the following plain text into well-formatted Markdown. Keep all original content, just add proper Markdown formatting (headings, bold, lists, paragraphs). Do NOT add new content. Output ONLY the formatted Markdown:\n\n${form.content}`,
        type: form.type,
        language: locale.value
      }
    })
    if (res.content) {
      form.content = res.content.replace(/^Tags:.*$/m, '').trim()
    }
  } catch (err: any) {
    formError.value = err?.data?.message || err.message || 'Conversion failed'
  } finally { convertingToMd.value = false }
}

// --- CRUD ---
const fetchContent = async () => {
  pending.value = true; error.value = ''
  try {
    const res = await $fetch<any>(`${baseURL}/admin/content?type=${selectedType.value}&page=${page.value}&limit=${pageSize}`, {
      headers: { Authorization: `Bearer ${token.value}` }
    })
    contentList.value = (res.data || []).map((item: any) => ({ ...item, type: selectedType.value }))
    pagination.value = res.pagination
  } catch (err: any) { error.value = err?.data?.message || err.message || 'Load failed' }
  finally { pending.value = false }
}

const switchType = (type: ContentType) => { selectedType.value = type; page.value = 1 }
const nextPage = () => { if (pagination.value && page.value < pagination.value.totalPages) page.value++ }
const prevPage = () => { if (page.value > 1) page.value-- }

const openCreateModal = () => {
  editingId.value = ''; formError.value = ''; actionMessage.value = ''; resetForm()
  aiTopic.value = ''; showModal.value = true
}

const openEditModal = async (item: any) => {
  editingId.value = item.id; formError.value = ''
  try {
    const res = await $fetch<any>(`${baseURL}/admin/content/${item.id}?type=${item.type}`, {
      headers: { Authorization: `Bearer ${token.value}` }
    })
    const c = res.content || item
    form.type = item.type; form.title = c.title || ''; form.slug = c.slug || ''
    form.thumbnail = c.thumbnail || ''; form.category = c.category || ''
    form.readTime = c.readTime || 5; form.excerpt = c.excerpt || ''
    form.tagsInput = Array.isArray(c.tags) ? c.tags.join(', ') : ''
    form.content = c.content || ''
    form.client = c.client || ''; form.industry = c.industry || ''
    form.location = c.location || ''; form.timeline = c.timeline || ''
    form.servicesInput = Array.isArray(c.services) ? c.services.join(', ') : ''
    form.challenge = c.challenge || ''; form.solution = c.solution || ''; form.result = c.result || ''
    showModal.value = true
  } catch (err: any) { actionError.value = true; actionMessage.value = err?.data?.message || 'Load failed' }
}

const closeModal = () => { showModal.value = false; saving.value = false; formError.value = '' }

const buildPayload = () => {
  if (form.type === 'post') {
    return { type: 'post', title: form.title, slug: form.slug, thumbnail: form.thumbnail, category: form.category, readTime: form.readTime, excerpt: form.excerpt, tags: parseCSV(form.tagsInput), content: form.content }
  }
  return { type: 'case', title: form.title, slug: form.slug, thumbnail: form.thumbnail, client: form.client, industry: form.industry, location: form.location, timeline: form.timeline, services: parseCSV(form.servicesInput), challenge: form.challenge, solution: form.solution, result: form.result, content: form.content }
}

const saveContent = async () => {
  if (!form.title.trim()) { formError.value = t('admin.content.title_required'); return }
  saving.value = true; formError.value = ''
  try {
    if (editingId.value) {
      await $fetch(`${baseURL}/admin/content/${editingId.value}`, { method: 'PUT', headers: { Authorization: `Bearer ${token.value}` }, body: buildPayload() })
      actionMessage.value = t('admin.content.updated_success')
    } else {
      await $fetch(`${baseURL}/admin/content`, { method: 'POST', headers: { Authorization: `Bearer ${token.value}` }, body: buildPayload() })
      actionMessage.value = t('admin.content.created_success')
    }
    closeModal(); await fetchContent()
  } catch (err: any) { formError.value = err?.data?.message || err.message || 'Save failed' }
  finally { saving.value = false }
}

const deleteContent = async (item: any) => {
  if (!confirm(t('admin.content.confirm_delete'))) return
  try {
    await $fetch(`${baseURL}/admin/content/${item.id}?type=${item.type}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token.value}` } })
    actionMessage.value = t('admin.content.deleted_success'); await fetchContent()
  } catch (err: any) { actionError.value = true; actionMessage.value = err?.data?.message || 'Delete failed' }
}

watch([selectedType, page], fetchContent)
onMounted(fetchContent)
</script>


<style scoped>
.toolbar-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 28px;
  padding: 0 4px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  color: #374151;
  cursor: pointer;
  transition: all 0.15s;
}
.toolbar-btn:hover { background: #e5e7eb; }

.editor-textarea {
  display: block;
  border: 1px solid #d1d5db;
  border-top: none;
  padding: 12px;
  font-family: 'SF Mono', 'Roboto Mono', monospace;
  font-size: 13px;
  line-height: 1.6;
  resize: vertical;
  min-height: 300px;
  max-height: 600px;
  outline: none;
}
.editor-textarea:focus { box-shadow: inset 0 0 0 2px #f97316; }

.editor-split .editor-textarea { border-right: none; border-bottom-right-radius: 0; }

/* Prose styles for preview */
.prose :deep(h1) { font-size: 1.5rem; font-weight: 700; margin: 0.5em 0; }
.prose :deep(h2) { font-size: 1.25rem; font-weight: 600; margin: 0.75em 0 0.5em; border-bottom: 1px solid #e5e7eb; padding-bottom: 0.25em; }
.prose :deep(h3) { font-size: 1.1rem; font-weight: 600; margin: 0.5em 0; }
.prose :deep(p) { margin: 0.5em 0; line-height: 1.7; }
.prose :deep(ul), .prose :deep(ol) { padding-left: 1.5em; margin: 0.5em 0; }
.prose :deep(li) { margin: 0.25em 0; }
.prose :deep(ul) { list-style: disc; }
.prose :deep(ol) { list-style: decimal; }
.prose :deep(strong) { font-weight: 600; }
.prose :deep(a) { color: #ea580c; text-decoration: underline; }
.prose :deep(img) { max-width: 100%; border-radius: 8px; margin: 1em 0; }
.prose :deep(blockquote) { border-left: 3px solid #f97316; padding-left: 1em; color: #6b7280; margin: 1em 0; }
.prose :deep(code) { background: #f3f4f6; padding: 2px 4px; border-radius: 3px; font-size: 0.9em; }
.prose :deep(pre) { background: #1f2937; color: #e5e7eb; padding: 1em; border-radius: 6px; overflow-x: auto; margin: 1em 0; }
</style>
