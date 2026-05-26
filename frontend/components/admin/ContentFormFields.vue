<template>
  <div class="space-y-4">
    <!-- 类型与 Slug（不可翻译） -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <div>
        <label for="content-type" class="admin-form-label">{{ t('admin.content.type') }}</label>
        <select id="content-type" v-model="form.type" name="type" class="admin-form-control" :disabled="disableType">
          <option value="post">{{ enumLabel('content_type', 'post') }}</option>
          <option value="case">{{ enumLabel('content_type', 'case') }}</option>
        </select>
      </div>
      <div>
        <label for="content-slug" class="admin-form-label">{{ t('admin.content.content_slug') }}</label>
        <input id="content-slug" v-model="form.slug" name="slug" type="text" autocomplete="off" class="admin-form-control" />
      </div>
    </div>

    <!-- 多语言内容（唯一入口） -->
    <div class="admin-form-section">
      <div class="flex items-center justify-between mb-2">
        <div>
          <h4 class="admin-form-section-title">{{ t('admin.content.translations_label') }}</h4>
          <p v-if="!hideTranslations" class="mt-1 text-xs text-gray-500">{{ t('admin.content.translations_hint') }}</p>
        </div>
        <div v-if="editingId && !hideTranslations" class="inline-flex items-center">
          <button type="button" :disabled="aiTranslating" class="admin-btn-ghost" @click="$emit('aiTranslate')">
            <Icon name="heroicons:sparkles" class="h-3.5 w-3.5" />
            {{ aiTranslating ? t('admin.content.ai_translating') : t('admin.content.ai_translate') }}
          </button>
          <AiHelpHint topic="content_translate" size="sm" />
        </div>
      </div>
      <div v-if="!hideTranslations" class="flex gap-2 mb-3 flex-wrap">
        <button
          v-for="loc in translationLocales"
          :key="loc"
          type="button"
          :class="['admin-locale-tab', activeLocale === loc ? 'admin-locale-tab--active' : '']"
          @click="$emit('update:translationLocale', loc)"
        >
          {{ localeTabLabel(loc) }}
        </button>
      </div>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div class="sm:col-span-2">
          <InlineAiField
            v-if="inlineAiEdit"
            :model-value="activeEntry.title"
            :enabled="inlineAiEdit"
            field-type="title"
            :label="`${t('admin.content.content_title')} (${activeLocale})`"
            input-id="content-title"
            required
            @update:model-value="activeEntry.title = $event"
          />
          <template v-else>
            <label :for="`trans-title-${activeLocale}`" class="admin-form-label">{{ t('admin.content.content_title') }} ({{ activeLocale }})</label>
            <input
              :id="`trans-title-${activeLocale}`"
              v-model="activeEntry.title"
              :required="activeLocale === SOURCE_LOCALE"
              autocomplete="off"
              class="admin-form-control"
            />
          </template>
        </div>
        <div v-if="form.type === 'post'" class="sm:col-span-2">
          <InlineAiField
            v-if="inlineAiEdit"
            :model-value="activeEntry.excerpt"
            :enabled="inlineAiEdit"
            field-type="plain"
            :label="`${t('admin.content.post_excerpt')} (${activeLocale})`"
            input-id="content-excerpt"
            multiline
            :rows="2"
            @update:model-value="activeEntry.excerpt = $event"
          />
          <template v-else>
            <label :for="`trans-excerpt-${activeLocale}`" class="admin-form-label">{{ t('admin.content.post_excerpt') }} ({{ activeLocale }})</label>
            <textarea :id="`trans-excerpt-${activeLocale}`" v-model="activeEntry.excerpt" rows="2" class="admin-form-control"></textarea>
          </template>
        </div>
        <div v-if="form.type === 'post'" class="sm:col-span-2">
          <label class="admin-form-label mb-1">{{ t('admin.content.post_content') }} ({{ activeLocale }})</label>
          <RichTextEditor
            v-if="activeLocale === SOURCE_LOCALE"
            v-model="activeEntry.content"
            :placeholder="t('admin.content.editor_placeholder')"
            :inline-ai-edit="inlineAiEdit"
            :inline-ai-language="SOURCE_LOCALE"
          />
          <textarea
            v-else
            v-model="activeEntry.content"
            rows="6"
            class="admin-form-control"
            :placeholder="t('admin.content.trans_content_placeholder')"
          />
        </div>
        <div v-if="form.type === 'case'" class="sm:col-span-2">
          <label :for="`trans-content-${activeLocale}`" class="admin-form-label">{{ t('admin.content.case_content') }} ({{ activeLocale }})</label>
          <textarea :id="`trans-content-${activeLocale}`" v-model="activeEntry.content" rows="6" class="admin-form-control" :placeholder="t('admin.content.editor_placeholder')"></textarea>
        </div>
        <template v-if="form.type === 'post'">
          <div>
            <label :for="`trans-authorName-${activeLocale}`" class="admin-form-label">{{ t('admin.content.author_name') }} ({{ activeLocale }})</label>
            <input :id="`trans-authorName-${activeLocale}`" v-model="activeEntry.authorName" autocomplete="off" class="admin-form-control" />
          </div>
          <div>
            <label :for="`trans-authorTitle-${activeLocale}`" class="admin-form-label">{{ t('admin.content.author_title') }} ({{ activeLocale }})</label>
            <input :id="`trans-authorTitle-${activeLocale}`" v-model="activeEntry.authorTitle" autocomplete="off" class="admin-form-control" />
          </div>
          <div class="sm:col-span-2">
            <label :for="`trans-authorBio-${activeLocale}`" class="admin-form-label">{{ t('admin.content.author_bio') }} ({{ activeLocale }})</label>
            <input :id="`trans-authorBio-${activeLocale}`" v-model="activeEntry.authorBio" autocomplete="off" class="admin-form-control" />
          </div>
        </template>
        <template v-if="form.type === 'case'">
          <div>
            <label :for="`trans-challenge-${activeLocale}`" class="admin-form-label">{{ t('admin.content.case_challenge') }} ({{ activeLocale }})</label>
            <textarea :id="`trans-challenge-${activeLocale}`" v-model="activeEntry.challenge" rows="3" class="admin-form-control"></textarea>
          </div>
          <div>
            <label :for="`trans-solution-${activeLocale}`" class="admin-form-label">{{ t('admin.content.case_solution') }} ({{ activeLocale }})</label>
            <textarea :id="`trans-solution-${activeLocale}`" v-model="activeEntry.solution" rows="3" class="admin-form-control"></textarea>
          </div>
          <div>
            <label :for="`trans-result-${activeLocale}`" class="admin-form-label">{{ t('admin.content.case_result') }} ({{ activeLocale }})</label>
            <textarea :id="`trans-result-${activeLocale}`" v-model="activeEntry.result" rows="3" class="admin-form-control"></textarea>
          </div>
        </template>
      </div>
    </div>

    <!-- 文章元数据（不可翻译） -->
    <div v-if="form.type === 'post'" class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <div>
        <label for="content-category" class="admin-form-label">{{ t('admin.content.post_category') }}</label>
        <select id="content-category" v-model="form.category" name="category" class="admin-form-control">
          <option value="">{{ t('admin.content.category_general') }}</option>
          <option v-for="slug in blogCategorySlugs" :key="slug" :value="slug">{{ t(`blog.categories.${slug}`) }}</option>
        </select>
      </div>
      <div>
        <label for="content-thumbnail" class="admin-form-label">{{ t('admin.content.content_thumbnail') }}</label>
        <input id="content-thumbnail" v-model="form.thumbnail" name="thumbnail" type="text" autocomplete="off" class="admin-form-control" :placeholder="t('admin.content.url_or_upload')" />
      </div>
      <div>
        <label for="content-ogImage" class="admin-form-label">{{ t('admin.content.og_image') }}</label>
        <input id="content-ogImage" v-model="form.ogImage" name="ogImage" type="text" autocomplete="off" class="admin-form-control" :placeholder="t('admin.content.og_image_hint')" />
      </div>
      <div>
        <label for="content-readTime" class="admin-form-label">{{ t('admin.content.post_read_time') }}</label>
        <input id="content-readTime" v-model.number="form.readTime" name="readTime" type="number" min="1" autocomplete="off" class="admin-form-control" />
      </div>
      <div>
        <label for="content-tagsInput" class="admin-form-label">{{ t('admin.content.post_tags') }}</label>
        <input id="content-tagsInput" v-model="form.tagsInput" name="tagsInput" type="text" autocomplete="off" class="admin-form-control" />
      </div>
      <div>
        <label for="content-authorAvatar" class="admin-form-label">{{ t('admin.content.author_avatar') }}</label>
        <input id="content-authorAvatar" v-model="form.authorAvatar" name="authorAvatar" type="text" autocomplete="off" class="admin-form-control" :placeholder="t('admin.content.url_or_upload')" />
      </div>
    </div>

    <!-- 案例元数据（不可翻译） -->
    <div v-if="form.type === 'case'" class="grid grid-cols-1 gap-4 sm:grid-cols-4">
      <div><label for="content-client" class="admin-form-label">{{ t('admin.content.case_client') }}</label><input id="content-client" v-model="form.client" name="client" type="text" autocomplete="off" class="admin-form-control" /></div>
      <div><label for="content-industry" class="admin-form-label">{{ t('admin.content.case_industry') }}</label><input id="content-industry" v-model="form.industry" name="industry" type="text" autocomplete="off" class="admin-form-control" /></div>
      <div><label for="content-location" class="admin-form-label">{{ t('admin.content.case_location') }}</label><input id="content-location" v-model="form.location" name="location" type="text" autocomplete="off" class="admin-form-control" /></div>
      <div><label for="content-timeline" class="admin-form-label">{{ t('admin.content.case_timeline') }}</label><input id="content-timeline" v-model="form.timeline" name="timeline" type="text" autocomplete="off" class="admin-form-control" /></div>
    </div>
    <div v-if="form.type === 'case'" class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <div>
        <label for="content-case-thumbnail" class="admin-form-label">{{ t('admin.content.content_thumbnail') }}</label>
        <input id="content-case-thumbnail" v-model="form.thumbnail" name="thumbnail" type="text" autocomplete="off" class="admin-form-control" :placeholder="t('admin.content.url_or_upload')" />
      </div>
      <div>
        <label for="content-case-ogImage" class="admin-form-label">{{ t('admin.content.og_image') }}</label>
        <input id="content-case-ogImage" v-model="form.ogImage" name="ogImage" type="text" autocomplete="off" class="admin-form-control" :placeholder="t('admin.content.og_image_hint')" />
      </div>
    </div>
    <div v-if="form.type === 'case'">
      <label for="content-imagesInput" class="admin-form-label">{{ t('admin.content.case_images') }}</label>
      <input id="content-imagesInput" v-model="form.imagesInput" name="imagesInput" type="text" autocomplete="off" class="admin-form-control" :placeholder="t('admin.content.url_or_upload')" />
    </div>
    <div v-if="form.type === 'case'">
      <label for="content-servicesInput" class="admin-form-label">{{ t('admin.content.case_services') }}</label>
      <input id="content-servicesInput" v-model="form.servicesInput" name="servicesInput" type="text" autocomplete="off" class="admin-form-control" />
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  form: any
  disableType: boolean
  editingId?: string
  aiTranslating?: boolean
  translationLocale?: string
  hideTranslations?: boolean
  inlineAiEdit?: boolean
}>()

defineEmits<{
  aiTranslate: []
  'update:translationLocale': [locale: string]
}>()

const { t } = useI18n()
const { enumLabel } = useDisplay()

const blogCategorySlugs = ['compliance', 'product_knowledge', 'packaging', 'market_insights'] as const
const translationLocales = ALL_LOCALES

const activeLocale = computed(() => {
  if (props.hideTranslations) return SOURCE_LOCALE
  return props.translationLocale || SOURCE_LOCALE
})

const activeEntry = computed(() => {
  if (!props.form.translations[activeLocale.value]) {
    props.form.translations[activeLocale.value] = {
      title: '', excerpt: '', content: '', authorName: '', authorTitle: '', authorBio: '',
      challenge: '', solution: '', result: '',
    }
  }
  return props.form.translations[activeLocale.value]
})

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
