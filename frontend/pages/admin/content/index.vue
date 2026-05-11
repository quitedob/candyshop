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
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-gray-500/75 p-4" role="dialog" aria-modal="true" @click.self="closeModal">
      <div class="my-8 w-full max-w-6xl rounded-lg bg-white shadow-xl">
        <!-- Modal Header -->
        <div class="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h3 class="text-lg font-semibold text-gray-900">{{ editingId ? t('admin.content.edit_content') : t('admin.content.create_content') }}</h3>
          <button type="button" class="text-gray-400 hover:text-gray-600" @click="closeModal" :aria-label="t('common.close')">
            <Icon name="heroicons:x-mark" class="h-6 w-6" aria-hidden="true" />
          </button>
        </div>

        <form @submit.prevent="saveContent">
          <div class="px-6 py-4 space-y-4">
            <!-- Meta fields row -->
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div>
                <label for="content-type" class="block text-sm font-medium text-gray-700">{{ t('admin.content.type') }}</label>
                <select id="content-type" v-model="form.type" name="type" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :disabled="Boolean(editingId)">
                  <option value="post">{{ enumLabel('content_type', 'post') }}</option>
                  <option value="case">{{ enumLabel('content_type', 'case') }}</option>
                </select>
              </div>
              <div>
                <label for="content-title" class="block text-sm font-medium text-gray-700">{{ t('admin.content.content_title') }}</label>
                <input id="content-title" v-model="form.title" name="title" type="text" autocomplete="off" required class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="content-slug" class="block text-sm font-medium text-gray-700">{{ t('admin.content.content_slug') }}</label>
                <input id="content-slug" v-model="form.slug" name="slug" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
            </div>

            <!-- Post-specific meta -->
            <div v-if="form.type === 'post'" class="grid grid-cols-1 gap-4 sm:grid-cols-4">
              <div>
                <label for="content-category" class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_category') }}</label>
                <input id="content-category" v-model="form.category" name="category" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="content-thumbnail" class="block text-sm font-medium text-gray-700">{{ t('admin.content.content_thumbnail') }}</label>
                <input id="content-thumbnail" v-model="form.thumbnail" name="thumbnail" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="t('admin.content.url_or_upload')" />
              </div>
              <div>
                <label for="content-readTime" class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_read_time') }}</label>
                <input id="content-readTime" v-model.number="form.readTime" name="readTime" type="number" min="1" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="content-tagsInput" class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_tags') }}</label>
                <input id="content-tagsInput" v-model="form.tagsInput" name="tagsInput" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
            </div>

            <!-- Author fields (post only) -->
            <div v-if="form.type === 'post'" class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label for="content-authorName" class="block text-sm font-medium text-gray-700">{{ t('admin.content.author_name') }}</label>
                <input id="content-authorName" v-model="form.authorName" name="authorName" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="content-authorTitle" class="block text-sm font-medium text-gray-700">{{ t('admin.content.author_title') }}</label>
                <input id="content-authorTitle" v-model="form.authorTitle" name="authorTitle" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label for="content-authorAvatar" class="block text-sm font-medium text-gray-700">{{ t('admin.content.author_avatar') }}</label>
                <input id="content-authorAvatar" v-model="form.authorAvatar" name="authorAvatar" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="t('admin.content.url_or_upload')" />
              </div>
              <div>
                <label for="content-authorBio" class="block text-sm font-medium text-gray-700">{{ t('admin.content.author_bio') }}</label>
                <input id="content-authorBio" v-model="form.authorBio" name="authorBio" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
            </div>

            <!-- Case-specific meta -->
            <div v-if="form.type === 'case'" class="grid grid-cols-1 gap-4 sm:grid-cols-4">
              <div><label for="content-client" class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_client') }}</label><input id="content-client" v-model="form.client" name="client" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" /></div>
              <div><label for="content-industry" class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_industry') }}</label><input id="content-industry" v-model="form.industry" name="industry" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" /></div>
              <div><label for="content-location" class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_location') }}</label><input id="content-location" v-model="form.location" name="location" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" /></div>
              <div><label for="content-timeline" class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_timeline') }}</label><input id="content-timeline" v-model="form.timeline" name="timeline" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" /></div>
            </div>

            <!-- Case images (comma-separated URLs) -->
            <div v-if="form.type === 'case'">
              <label for="content-imagesInput" class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_images') }}</label>
              <input id="content-imagesInput" v-model="form.imagesInput" name="imagesInput" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="t('admin.content.url_or_upload')" />
            </div>

            <!-- Excerpt (post only) -->
            <div v-if="form.type === 'post'">
              <label for="content-excerpt" class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_excerpt') }}</label>
              <textarea id="content-excerpt" v-model="form.excerpt" name="excerpt" rows="2" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
            </div>

            <!-- AI Generate Bar -->
            <div v-if="form.type === 'post'" class="flex items-center gap-3 rounded-lg bg-amber-50 border border-amber-200 p-3">
              <Icon name="heroicons:sparkles" class="h-5 w-5 text-amber-600 flex-shrink-0" />
              <label for="content-aiTopic" class="sr-only">{{ t('admin.content.ai_topic_placeholder') }}</label>
              <input id="content-aiTopic" v-model="aiTopic" name="aiTopic" type="text" :placeholder="t('admin.content.ai_topic_placeholder')" class="flex-1 rounded-md border border-amber-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500" @keydown.enter.prevent="generateWithAI" />
              <button type="button" :disabled="aiGenerating || !aiTopic.trim()" class="inline-flex items-center gap-1.5 rounded-md bg-orange-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50" @click="generateWithAI">
                <Icon v-if="aiGenerating" name="heroicons:arrow-path" class="h-4 w-4 animate-spin" />
                <Icon v-else name="heroicons:sparkles" class="h-4 w-4" />
                {{ aiGenerating ? t('admin.content.ai_generating') : t('admin.content.ai_generate') }}
              </button>
            </div>

            <!-- Rich Text Editor (blog posts) -->
            <div v-if="form.type === 'post'">
              <label class="block text-sm font-medium text-gray-700 mb-1">{{ t('admin.content.post_content') }}</label>
              <RichTextEditor v-model="form.content" :placeholder="t('admin.content.editor_placeholder')" />
            </div>

            <!-- Plain textarea for case study content -->
            <div v-if="form.type === 'case'">
              <label for="content-body" class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_content') }}</label>
              <textarea id="content-body" v-model="form.content" name="content" rows="8" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :placeholder="t('admin.content.editor_placeholder')"></textarea>
            </div>

            <!-- Case-specific fields below editor -->
            <template v-if="form.type === 'case'">
              <div><label for="content-servicesInput" class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_services') }}</label><input id="content-servicesInput" v-model="form.servicesInput" name="servicesInput" type="text" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" /></div>
              <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
                <div><label for="content-challenge" class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_challenge') }}</label><textarea id="content-challenge" v-model="form.challenge" name="challenge" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea></div>
                <div><label for="content-solution" class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_solution') }}</label><textarea id="content-solution" v-model="form.solution" name="solution" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea></div>
                <div><label for="content-result" class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_result') }}</label><textarea id="content-result" v-model="form.result" name="result" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea></div>
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
import { reactive, ref, watch, onMounted } from 'vue'
import { marked } from 'marked'

definePageMeta({ layout: 'admin', middleware: ['auth'] })

const { token } = useAuth()
const { t, locale } = useI18n()
const { enumLabel, formatDate } = useDisplay()
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

// AI state
const aiTopic = ref('')
const aiGenerating = ref(false)

const form = reactive({
  type: 'post' as ContentType,
  title: '', slug: '', thumbnail: '', category: '', readTime: 5,
  excerpt: '', tagsInput: '', content: '',
  authorName: '', authorAvatar: '', authorTitle: '', authorBio: '',
  client: '', industry: '', location: '', timeline: '',
  servicesInput: '', imagesInput: '', challenge: '', solution: '', result: ''
})

const parseCSV = (v: string) => v.split(',').map(s => s.trim()).filter(Boolean)

const resetForm = () => {
  form.type = selectedType.value; form.title = ''; form.slug = ''; form.thumbnail = ''
  form.category = ''; form.readTime = 5; form.excerpt = ''; form.tagsInput = ''; form.content = ''
  form.authorName = ''; form.authorAvatar = ''; form.authorTitle = ''; form.authorBio = ''
  form.client = ''; form.industry = ''; form.location = ''; form.timeline = ''
  form.servicesInput = ''; form.imagesInput = ''; form.challenge = ''; form.solution = ''; form.result = ''
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
    const lines = raw.split('\n')
    const titleLine = lines.find((l: string) => l.startsWith('# '))
    if (titleLine && !form.title) form.title = titleLine.replace(/^#\s+/, '')
    // Extract tags line
    const tagsLine = lines.find((l: string) => l.startsWith('Tags:'))
    if (tagsLine && !form.tagsInput) form.tagsInput = tagsLine.replace(/^Tags:\s*/, '')
    // Remove tags line, convert Markdown to HTML for the rich editor
    const mdContent = lines.filter((l: string) => !l.startsWith('Tags:')).join('\n').trim()
    form.content = marked.parse(mdContent) as string
    // Extract first paragraph as excerpt
    const contentLines = form.content.replace(/<[^>]+>/g, '').split('\n').filter((l: string) => l.trim())
    if (contentLines.length && !form.excerpt) form.excerpt = contentLines[0].substring(0, 200)
  } catch (err: any) {
    formError.value = err?.data?.message || err.message || t('admin.content.ai_generate_failed')
  } finally { aiGenerating.value = false }
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
  } catch (err: any) { error.value = err?.data?.message || err.message || t('admin.content.load_failed') }
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
    form.authorName = c.author?.name || ''; form.authorAvatar = c.author?.avatar || ''
    form.authorTitle = c.author?.title || ''; form.authorBio = c.author?.bio || ''
    form.client = c.client || ''; form.industry = c.industry || ''
    form.location = c.location || ''; form.timeline = c.timeline || ''
    form.servicesInput = Array.isArray(c.services) ? c.services.join(', ') : ''
    form.imagesInput = Array.isArray(c.images) ? c.images.join(', ') : ''
    form.challenge = c.challenge || ''; form.solution = c.solution || ''; form.result = c.result || ''
    showModal.value = true
  } catch (err: any) { actionError.value = true; actionMessage.value = err?.data?.message || 'Load failed' }
}

const closeModal = () => { showModal.value = false; saving.value = false; formError.value = '' }

const buildPayload = () => {
  if (form.type === 'post') {
    return { type: 'post', title: form.title, slug: form.slug, thumbnail: form.thumbnail, category: form.category, readTime: form.readTime, excerpt: form.excerpt, tags: parseCSV(form.tagsInput), content: form.content, author: { name: form.authorName, avatar: form.authorAvatar, title: form.authorTitle, bio: form.authorBio } }
  }
  return { type: 'case', title: form.title, slug: form.slug, thumbnail: form.thumbnail, client: form.client, industry: form.industry, location: form.location, timeline: form.timeline, services: parseCSV(form.servicesInput), images: parseCSV(form.imagesInput), challenge: form.challenge, solution: form.solution, result: form.result, content: form.content }
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
    // Invalidate public blog cache so changes appear immediately
    refreshNuxtData('blog-posts-all')
  } catch (err: any) { formError.value = err?.data?.message || err.message || t('admin.content.save_failed') }
  finally { saving.value = false }
}

const deleteContent = async (item: any) => {
  if (!confirm(t('admin.content.confirm_delete'))) return
  try {
    await $fetch(`${baseURL}/admin/content/${item.id}?type=${item.type}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token.value}` } })
    actionMessage.value = t('admin.content.deleted_success'); await fetchContent()
    refreshNuxtData('blog-posts-all')
  } catch (err: any) { actionError.value = true; actionMessage.value = err?.data?.message || t('admin.content.delete_failed') }
}

watch([selectedType, page], fetchContent)
onMounted(fetchContent)
</script>
