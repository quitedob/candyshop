<template>
  <div>
    <div class="sm:flex sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900">{{ t('admin.content.title') }}</h1>
        <p class="mt-2 text-sm text-gray-700">{{ t('admin.content.description') }}</p>
      </div>
      <div class="mt-4 sm:mt-0">
        <button type="button" class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700" @click="openCreateModal">
          {{ t('admin.content.add_content') }}
        </button>
      </div>
    </div>

    <div class="mt-6 flex items-center gap-2">
      <button
        v-for="option in typeOptions"
        :key="option.value"
        type="button"
        class="rounded-md px-3 py-2 text-sm font-medium"
        :class="selectedType === option.value ? 'bg-blue-600 text-white' : 'bg-white text-gray-700 ring-1 ring-inset ring-gray-300 hover:bg-gray-50'"
        @click="switchType(option.value)"
      >
        {{ option.value === 'post' ? t('admin.content.tab_posts') : t('admin.content.tab_cases') }}
      </button>
    </div>

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
          <tr v-if="pending">
            <td colspan="5" class="py-5 text-center text-sm text-gray-500">{{ t('admin.content.loading') }}</td>
          </tr>
          <tr v-else-if="error">
            <td colspan="5" class="py-5 text-center text-sm text-red-600">{{ error }}</td>
          </tr>
          <tr v-else-if="contentList.length === 0">
            <td colspan="5" class="py-5 text-center text-sm text-gray-500">{{ t('admin.content.no_data') }}</td>
          </tr>
          <tr v-else v-for="item in contentList" :key="item.id">
            <td class="py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">{{ item.title }}</td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
              <span class="inline-flex rounded-full px-2 text-xs font-semibold leading-5" :class="item.type === 'post' ? 'bg-purple-100 text-purple-800' : 'bg-cyan-100 text-cyan-800'">
                {{ item.type }}
              </span>
            </td>
            <td class="px-3 py-4 text-sm text-gray-500">
              {{ item.type === 'post' ? (item.category || 'General') : (item.industry || 'General') }}
            </td>
            <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
              {{ item.updatedAt ? new Date(item.updatedAt).toLocaleString() : '-' }}
            </td>
            <td class="whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
              <button type="button" class="text-blue-600 hover:text-blue-900" @click="openEditModal(item)">{{ t('admin.content.edit') }}</button>
              <button type="button" class="ml-4 text-red-600 hover:text-red-900" @click="deleteContent(item)">{{ t('admin.content.delete') }}</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="pagination" class="mt-4 flex items-center justify-between rounded-lg border-t border-gray-200 bg-white px-4 py-3 shadow sm:px-6">
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

    <div v-if="showModal" class="fixed inset-0 z-10 overflow-y-auto" role="dialog" aria-modal="true">
      <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="closeModal"></div>
        <span class="hidden sm:inline-block sm:h-screen sm:align-middle" aria-hidden="true">&#8203;</span>
        <div class="inline-block w-full transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left align-bottom shadow-xl transition-all sm:my-8 sm:max-w-3xl sm:p-6 sm:align-middle">
          <h3 class="text-lg font-medium leading-6 text-gray-900">{{ editingId ? t('admin.content.edit_content') : t('admin.content.create_content') }}</h3>

          <form class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2" @submit.prevent="saveContent">
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
            <div>
              <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.content_thumbnail') }}</label>
              <input v-model="form.thumbnail" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
            </div>

            <template v-if="form.type === 'post'">
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_category') }}</label>
                <input v-model="form.category" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_read_time') }}</label>
                <input v-model.number="form.readTime" type="number" min="1" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div class="sm:col-span-2">
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_excerpt') }}</label>
                <textarea v-model="form.excerpt" rows="2" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
              </div>
              <div class="sm:col-span-2">
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_tags') }}</label>
                <input v-model="form.tagsInput" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div class="sm:col-span-2">
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_content') }}</label>
                <textarea v-model="form.content" rows="6" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
              </div>
            </template>

            <template v-else>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_client') }}</label>
                <input v-model="form.client" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_industry') }}</label>
                <input v-model="form.industry" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_location') }}</label>
                <input v-model="form.location" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_timeline') }}</label>
                <input v-model="form.timeline" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div class="sm:col-span-2">
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_services') }}</label>
                <input v-model="form.servicesInput" type="text" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
              </div>
              <div class="sm:col-span-2">
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_challenge') }}</label>
                <textarea v-model="form.challenge" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
              </div>
              <div class="sm:col-span-2">
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_solution') }}</label>
                <textarea v-model="form.solution" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
              </div>
              <div class="sm:col-span-2">
                <label class="block text-sm font-medium text-gray-700">{{ t('admin.content.case_result') }}</label>
                <textarea v-model="form.result" rows="3" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
              </div>
            </template>

            <div v-if="formError" class="sm:col-span-2 text-sm text-red-600">{{ formError }}</div>
            <div class="sm:col-span-2 mt-2 flex justify-end gap-3">
              <button type="button" class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700" @click="closeModal">{{ t('admin.content.cancel') }}</button>
              <button type="submit" :disabled="saving" class="rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white disabled:opacity-50">
                {{ saving ? t('admin.content.saving') : (editingId ? t('admin.content.update') : t('admin.content.create')) }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <p v-if="actionMessage" class="mt-4 text-sm" :class="actionError ? 'text-red-600' : 'text-green-600'">
      {{ actionMessage }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch, onMounted } from 'vue'

definePageMeta({
  layout: 'admin',
  middleware: ['auth']
})

const { token } = useAuth()
const { t } = useI18n()
const config = useRuntimeConfig()
const baseURL = config.public.apiBase || '/api/v1'

const typeOptions = [
  { value: 'post', label: 'Posts' },
  { value: 'case', label: 'Case Studies' }
] as const
type ContentType = typeof typeOptions[number]['value']

const selectedType = ref<ContentType>('post')
const contentList = ref<any[]>([])
const pagination = ref<any>(null)
const pending = ref(true)
const error = ref('')
const page = ref(1)
const pageSize = 20

const showModal = ref(false)
const editingId = ref('')
const saving = ref(false)
const formError = ref('')
const actionMessage = ref('')
const actionError = ref(false)

const form = reactive({
  type: 'post' as ContentType,
  title: '',
  slug: '',
  thumbnail: '',
  category: '',
  readTime: 5,
  excerpt: '',
  tagsInput: '',
  content: '',
  client: '',
  industry: '',
  location: '',
  timeline: '',
  servicesInput: '',
  challenge: '',
  solution: '',
  result: ''
})

const resetForm = () => {
  form.type = selectedType.value
  form.title = ''
  form.slug = ''
  form.thumbnail = ''
  form.category = ''
  form.readTime = 5
  form.excerpt = ''
  form.tagsInput = ''
  form.content = ''
  form.client = ''
  form.industry = ''
  form.location = ''
  form.timeline = ''
  form.servicesInput = ''
  form.challenge = ''
  form.solution = ''
  form.result = ''
}

const parseCSV = (value: string) => value.split(',').map(v => v.trim()).filter(Boolean)

const fetchContent = async () => {
  pending.value = true
  error.value = ''
  try {
    const res = await $fetch<any>(`${baseURL}/admin/content?type=${selectedType.value}&page=${page.value}&limit=${pageSize}`, {
      headers: { Authorization: `Bearer ${token.value}` }
    })
    contentList.value = (res.data || []).map((item: any) => ({
      ...item,
      type: selectedType.value
    }))
    pagination.value = res.pagination
  } catch (err: any) {
    error.value = err?.data?.message || err.message || 'Failed to fetch content'
  } finally {
    pending.value = false
  }
}

const switchType = (type: ContentType) => {
  selectedType.value = type
  page.value = 1
}

const nextPage = () => {
  if (pagination.value && page.value < pagination.value.totalPages) {
    page.value += 1
  }
}

const prevPage = () => {
  if (page.value > 1) {
    page.value -= 1
  }
}

const openCreateModal = () => {
  editingId.value = ''
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false
  resetForm()
  showModal.value = true
}

const openEditModal = async (item: any) => {
  editingId.value = item.id
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false
  try {
    const res = await $fetch<any>(`${baseURL}/admin/content/${item.id}?type=${item.type}`, {
      headers: { Authorization: `Bearer ${token.value}` }
    })
    const content = res.content || item
    form.type = item.type
    form.title = content.title || ''
    form.slug = content.slug || ''
    form.thumbnail = content.thumbnail || ''
    form.category = content.category || ''
    form.readTime = content.readTime || 5
    form.excerpt = content.excerpt || ''
    form.tagsInput = Array.isArray(content.tags) ? content.tags.join(', ') : ''
    form.content = content.content || ''
    form.client = content.client || ''
    form.industry = content.industry || ''
    form.location = content.location || ''
    form.timeline = content.timeline || ''
    form.servicesInput = Array.isArray(content.services) ? content.services.join(', ') : ''
    form.challenge = content.challenge || ''
    form.solution = content.solution || ''
    form.result = content.result || ''
    showModal.value = true
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.data?.message || err.message || 'Failed to load content details'
  }
}

const closeModal = () => {
  showModal.value = false
  saving.value = false
  formError.value = ''
}

const buildPayload = () => {
  if (form.type === 'post') {
    return {
      type: 'post',
      title: form.title,
      slug: form.slug,
      thumbnail: form.thumbnail,
      category: form.category,
      readTime: form.readTime,
      excerpt: form.excerpt,
      tags: parseCSV(form.tagsInput),
      content: form.content
    }
  }

  return {
    type: 'case',
    title: form.title,
    slug: form.slug,
    thumbnail: form.thumbnail,
    client: form.client,
    industry: form.industry,
    location: form.location,
    timeline: form.timeline,
    services: parseCSV(form.servicesInput),
    challenge: form.challenge,
    solution: form.solution,
    result: form.result
  }
}

const saveContent = async () => {
  if (!form.title.trim()) {
    formError.value = t('admin.content.title_required')
    return
  }

  saving.value = true
  formError.value = ''
  actionMessage.value = ''
  actionError.value = false

  const payload = buildPayload()
  try {
    if (editingId.value) {
      await $fetch(`${baseURL}/admin/content/${editingId.value}`, {
        method: 'PUT',
        headers: { Authorization: `Bearer ${token.value}` },
        body: payload
      })
      actionMessage.value = t('admin.content.updated_success')
    } else {
      await $fetch(`${baseURL}/admin/content`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token.value}` },
        body: payload
      })
      actionMessage.value = t('admin.content.created_success')
    }

    closeModal()
    await fetchContent()
  } catch (err: any) {
    formError.value = err?.data?.message || err.message || 'Failed to save content'
  } finally {
    saving.value = false
  }
}

const deleteContent = async (item: any) => {
  if (!confirm(t('admin.content.confirm_delete'))) return

  actionMessage.value = ''
  actionError.value = false
  try {
    await $fetch(`${baseURL}/admin/content/${item.id}?type=${item.type}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token.value}` }
    })
    actionMessage.value = t('admin.content.deleted_success')
    await fetchContent()
  } catch (err: any) {
    actionError.value = true
    actionMessage.value = err?.data?.message || err.message || 'Failed to delete content'
  }
}

watch([selectedType, page], fetchContent)
onMounted(fetchContent)
</script>
