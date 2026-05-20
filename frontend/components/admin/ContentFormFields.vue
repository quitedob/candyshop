<template>
  <div class="space-y-4">
    <!-- Meta fields row -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
      <div>
        <label for="content-type" class="block text-sm font-medium text-gray-700">{{ t('admin.content.type') }}</label>
        <select id="content-type" v-model="form.type" name="type" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" :disabled="disableType">
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

    <!-- Translations -->
    <div class="mt-2 border-t pt-3">
      <div class="flex items-center justify-between mb-2">
        <h4 class="text-sm font-semibold text-gray-900">{{ t('admin.content.translations_label') }}</h4>
        <button
          v-if="editingId"
          type="button"
          :disabled="aiTranslating"
          class="inline-flex items-center gap-1.5 rounded-lg border border-orange-200 bg-orange-50 px-3 py-1 text-xs font-medium text-orange-700 hover:bg-orange-100 disabled:opacity-50 transition-colors"
          @click="$emit('aiTranslate')"
        >
          <Icon name="heroicons:sparkles" class="h-3.5 w-3.5" />
          {{ aiTranslating ? t('admin.content.ai_translating') : t('admin.content.ai_translate') }}
        </button>
      </div>
      <div class="flex gap-2 mb-3 flex-wrap">
        <button
          v-for="loc in translationLocales"
          :key="loc"
          type="button"
          :class="[
            'rounded-lg px-3 py-1 text-xs font-medium transition-colors',
            translationLocale === loc ? 'bg-orange-500 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
          ]"
          @click="$emit('update:translationLocale', loc)"
        >
          {{ localeTabLabel(loc) }}
        </button>
      </div>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div>
          <label :for="`trans-title-${translationLocale}`" class="block text-sm font-medium text-gray-700">{{ t('admin.content.content_title') }} ({{ translationLocale }})</label>
          <input :id="`trans-title-${translationLocale}`" v-model="form.translations[translationLocale].title" :name="`transTitle-${translationLocale}`" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label :for="`trans-excerpt-${translationLocale}`" class="block text-sm font-medium text-gray-700">{{ t('admin.content.post_excerpt') }} ({{ translationLocale }})</label>
          <input :id="`trans-excerpt-${translationLocale}`" v-model="form.translations[translationLocale].excerpt" :name="`transExcerpt-${translationLocale}`" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div class="sm:col-span-2">
          <label :for="`trans-content-${translationLocale}`" class="block text-sm font-medium text-gray-700">{{ form.type === 'post' ? t('admin.content.post_content') : t('admin.content.case_content') }} ({{ translationLocale }})</label>
          <textarea :id="`trans-content-${translationLocale}`" v-model="form.translations[translationLocale].content" :name="`transContent-${translationLocale}`" rows="3" autocomplete="off" class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm"></textarea>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  form: any
  disableType: boolean
  editingId?: string
  aiTranslating?: boolean
  translationLocale?: string
}>()

defineEmits<{
  aiTranslate: []
  'update:translationLocale': [locale: string]
}>()

const { t } = useI18n()
const { enumLabel } = useDisplay()

const translationLocales = ['en', 'zh', 'ko', 'ar', 'ja', 'th', 'vi', 'id', 'ms']

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
</script>
