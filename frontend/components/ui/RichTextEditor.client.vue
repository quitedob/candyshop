<template>
  <div class="rte-wrapper" :class="{ 'rte--focused': focused }">
    <!-- Toolbar -->
    <div v-if="editor" class="rte-toolbar" role="toolbar" :aria-label="t('common.rte.text_formatting')">
      <div class="rte-toolbar__group">
        <button type="button" class="rte-btn" :class="{ 'rte-btn--active': editor.isActive('bold') }" :aria-label="t('common.rte.bold')" :title="t('common.rte.bold')" @click="editor.chain().focus().toggleBold().run()">
          <Icon name="heroicons:bold" class="rte-btn__icon" aria-hidden="true" />
        </button>
        <button type="button" class="rte-btn" :class="{ 'rte-btn--active': editor.isActive('italic') }" :aria-label="t('common.rte.italic')" :title="t('common.rte.italic')" @click="editor.chain().focus().toggleItalic().run()">
          <Icon name="heroicons:italic" class="rte-btn__icon" aria-hidden="true" />
        </button>
        <button type="button" class="rte-btn" :class="{ 'rte-btn--active': editor.isActive('underline') }" :aria-label="t('common.rte.underline')" :title="t('common.rte.underline')" @click="editor.chain().focus().toggleUnderline().run()">
          <Icon name="heroicons:underline" class="rte-btn__icon" aria-hidden="true" />
        </button>
        <button type="button" class="rte-btn" :class="{ 'rte-btn--active': editor.isActive('strike') }" :aria-label="t('common.rte.strikethrough')" :title="t('common.rte.strikethrough')" @click="editor.chain().focus().toggleStrike().run()">
          <Icon name="heroicons:strikethrough" class="rte-btn__icon" aria-hidden="true" />
        </button>
      </div>

      <span class="rte-toolbar__divider" />

      <div class="rte-toolbar__group">
        <button type="button" class="rte-btn" :class="{ 'rte-btn--active': editor.isActive('heading', { level: 1 }) }" :aria-label="t('common.rte.heading1')" :title="t('common.rte.heading1')" @click="editor.chain().focus().toggleHeading({ level: 1 }).run()">H1</button>
        <button type="button" class="rte-btn" :class="{ 'rte-btn--active': editor.isActive('heading', { level: 2 }) }" :aria-label="t('common.rte.heading2')" :title="t('common.rte.heading2')" @click="editor.chain().focus().toggleHeading({ level: 2 }).run()">H2</button>
        <button type="button" class="rte-btn" :class="{ 'rte-btn--active': editor.isActive('heading', { level: 3 }) }" :aria-label="t('common.rte.heading3')" :title="t('common.rte.heading3')" @click="editor.chain().focus().toggleHeading({ level: 3 }).run()">H3</button>
        <button type="button" class="rte-btn" :aria-label="t('common.rte.paragraph')" :title="t('common.rte.paragraph')" @click="editor.chain().focus().setParagraph().run()">P</button>
      </div>

      <span class="rte-toolbar__divider" />

      <div class="rte-toolbar__group">
        <button type="button" class="rte-btn" :class="{ 'rte-btn--active': editor.isActive('bulletList') }" :aria-label="t('common.rte.bullet_list')" :title="t('common.rte.bullet_list')" @click="editor.chain().focus().toggleBulletList().run()">
          <Icon name="heroicons:list-bullet" class="rte-btn__icon" aria-hidden="true" />
        </button>
        <button type="button" class="rte-btn" :class="{ 'rte-btn--active': editor.isActive('orderedList') }" :aria-label="t('common.rte.ordered_list')" :title="t('common.rte.ordered_list')" @click="editor.chain().focus().toggleOrderedList().run()">
          <Icon name="heroicons:queue-list" class="rte-btn__icon" aria-hidden="true" />
        </button>
        <button type="button" class="rte-btn" :class="{ 'rte-btn--active': editor.isActive('blockquote') }" :aria-label="t('common.rte.blockquote')" :title="t('common.rte.blockquote')" @click="editor.chain().focus().toggleBlockquote().run()">
          <Icon name="heroicons:chat-bubble-bottom-center-text" class="rte-btn__icon" aria-hidden="true" />
        </button>
      </div>

      <span class="rte-toolbar__divider" />

      <div class="rte-toolbar__group">
        <button type="button" class="rte-btn" :class="{ 'rte-btn--active': editor.isActive({ textAlign: 'left' }) }" :aria-label="t('common.rte.align_left')" :title="t('common.rte.align_left')" @click="editor.chain().focus().setTextAlign('left').run()">
          <Icon name="heroicons:bars-3-bottom-left" class="rte-btn__icon" aria-hidden="true" />
        </button>
        <button type="button" class="rte-btn" :class="{ 'rte-btn--active': editor.isActive({ textAlign: 'center' }) }" :aria-label="t('common.rte.align_center')" :title="t('common.rte.align_center')" @click="editor.chain().focus().setTextAlign('center').run()">
          <Icon name="heroicons:bars-3" class="rte-btn__icon" aria-hidden="true" />
        </button>
        <button type="button" class="rte-btn" :class="{ 'rte-btn--active': editor.isActive({ textAlign: 'right' }) }" :aria-label="t('common.rte.align_right')" :title="t('common.rte.align_right')" @click="editor.chain().focus().setTextAlign('right').run()">
          <Icon name="heroicons:bars-3-bottom-right" class="rte-btn__icon" aria-hidden="true" />
        </button>
      </div>

      <span class="rte-toolbar__divider" />

      <div class="rte-toolbar__group">
        <button type="button" class="rte-btn" :aria-label="t('common.rte.insert_link')" :title="t('common.rte.insert_link')" @click="openLinkDialog">
          <Icon name="heroicons:link" class="rte-btn__icon" aria-hidden="true" />
        </button>
        <button type="button" class="rte-btn" :aria-label="t('common.rte.insert_image')" :title="t('common.rte.insert_image')" @click="triggerImageUpload">
          <Icon name="heroicons:photo" class="rte-btn__icon" aria-hidden="true" />
        </button>
        <label class="sr-only" for="rte-image-upload">{{ t('common.rte.upload_image') }}</label>
        <input id="rte-image-upload" ref="imageInputRef" type="file" accept="image/*" class="rte-file-input" @change="handleImageFile" />
      </div>

      <span class="rte-toolbar__divider" />

      <div class="rte-toolbar__group">
        <button type="button" class="rte-btn" :aria-label="t('common.rte.undo')" :title="t('common.rte.undo')" :disabled="!editor.can().undo()" @click="editor.chain().focus().undo().run()">
          <Icon name="heroicons:arrow-uturn-left" class="rte-btn__icon" aria-hidden="true" />
        </button>
        <button type="button" class="rte-btn" :aria-label="t('common.rte.redo')" :title="t('common.rte.redo')" :disabled="!editor.can().redo()" @click="editor.chain().focus().redo().run()">
          <Icon name="heroicons:arrow-uturn-right" class="rte-btn__icon" aria-hidden="true" />
        </button>
      </div>
    </div>

    <!-- Link Dialog -->
    <div v-if="linkDialogOpen" class="rte-link-dialog">
      <label for="rte-link-input" class="sr-only">{{ t('common.rte.link_url') }}</label>
      <input id="rte-link-input" ref="linkInputRef" v-model="linkUrl" type="url" :placeholder="t('common.rte.link_placeholder')" class="rte-link-dialog__input" @keydown.enter="setLink" @keydown.escape="linkDialogOpen = false" />
      <button type="button" class="rte-link-dialog__btn" @click="setLink">{{ t('common.rte.set') }}</button>
      <button type="button" class="rte-link-dialog__btn rte-link-dialog__btn--cancel" @click="removeLink">{{ t('common.rte.remove') }}</button>
    </div>

    <!-- Image Upload Status -->
    <div v-if="imageUploading" class="rte-image-status">
      <Icon name="heroicons:arrow-path" class="w-4 h-4 animate-spin" aria-hidden="true" /> {{ t('common.rte.uploading_image') }}
    </div>

    <!-- Editor Content Area -->
    <TiptapEditorContent :editor="editor" class="rte-content" @focus="focused = true" @blur="focused = false" @paste="handlePaste" @drop="handleDrop" />
  </div>
</template>

<script setup lang="ts">
import { useEditor, EditorContent as TiptapEditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import Link from '@tiptap/extension-link'
import ImageExt from '@tiptap/extension-image'
import TextAlign from '@tiptap/extension-text-align'
import Placeholder from '@tiptap/extension-placeholder'
import { useI18n } from '#i18n'

const props = withDefaults(defineProps<{
  modelValue?: string
  placeholder?: string
}>(), {
  modelValue: '',
  placeholder: undefined
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const focused = ref(false)
const linkDialogOpen = ref(false)
const linkUrl = ref('')
const { t } = useI18n()
const defaultPlaceholder = t('common.rte.placeholder')
const imageUploading = ref(false)
const imageInputRef = ref<HTMLInputElement>()
const linkInputRef = ref<HTMLInputElement>()

const editor = useEditor({
  content: props.modelValue,
  extensions: [
    StarterKit,
    Underline,
    Link.configure({ openOnClick: false }),
    ImageExt,
    TextAlign.configure({ types: ['heading', 'paragraph'] }),
    Placeholder.configure({ placeholder: props.placeholder || defaultPlaceholder })
  ],
  onUpdate: ({ editor }) => {
    emit('update:modelValue', editor.getHTML())
  }
})

watch(() => props.modelValue, (val) => {
  if (editor.value && val !== editor.value.getHTML()) {
    editor.value.commands.setContent(val, false)
  }
})

const openLinkDialog = () => {
  linkUrl.value = editor.value?.getAttributes('link').href || ''
  linkDialogOpen.value = true
  nextTick(() => linkInputRef.value?.focus())
}

const setLink = () => {
  if (linkUrl.value) {
    editor.value?.chain().focus().extendMarkRange('link').setLink({ href: linkUrl.value }).run()
  }
  linkDialogOpen.value = false
}

const removeLink = () => {
  editor.value?.chain().focus().extendMarkRange('link').unsetLink().run()
  linkDialogOpen.value = false
}

const triggerImageUpload = () => {
  imageInputRef.value?.click()
}

const handleImageFile = async (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  imageUploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    const { fetchApi } = useApi()
    const res = await fetchApi<{ url: string }>('/upload', { method: 'POST', body: formData })
    if (res.url) {
      editor.value?.chain().focus().setImage({ src: res.url }).run()
    }
  } catch { /* silently fail */ }
  finally {
    imageUploading.value = false
    if (imageInputRef.value) imageInputRef.value.value = ''
  }
}

const handlePaste = (e: ClipboardEvent) => {
  const items = e.clipboardData?.items
  if (!items) return
  for (const item of items) {
    if (item.type.startsWith('image/')) {
      e.preventDefault()
    }
  }
}

const handleDrop = (e: DragEvent) => {
  if (e.dataTransfer?.files.length) {
    e.preventDefault()
  }
}

onBeforeUnmount(() => {
  editor.value?.destroy()
})
</script>

<style scoped>
.rte-wrapper {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  overflow: hidden;
  transition: border-color 0.15s, box-shadow 0.15s;
  background: var(--color-bg);
}
.rte--focused {
  border-color: var(--color-highlight);
  box-shadow: 0 0 0 3px rgba(var(--color-highlight-rgb), 0.1);
}

.rte-toolbar {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 6px 8px;
  background: var(--color-bg-alt);
  border-bottom: 1px solid var(--color-border);
  flex-wrap: wrap;
}
.rte-toolbar__group {
  display: flex;
  align-items: center;
  gap: 1px;
}
.rte-toolbar__divider {
  width: 1px;
  height: 20px;
  background: var(--color-border-dark);
  margin: 0 4px;
}

.rte-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  height: 32px;
  padding: 0 6px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text);
  cursor: pointer;
  font-size: var(--text-xs);
  font-weight: 600;
  transition: background-color 0.15s;
}
.rte-btn:hover { background: var(--color-bg-dark); }
.rte-btn--active { background: var(--color-highlight); color: var(--color-text-on-primary); }
.rte-btn--active:hover { background: var(--color-highlight-hover); }
.rte-btn:disabled { opacity: 0.35; cursor: default; }
.rte-btn__icon { width: 16px; height: 16px; }

.rte-file-input { display: none; }

.rte-link-dialog {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 8px;
  background: var(--color-accent-light);
  border-bottom: 1px solid var(--color-border-dark);
}
.rte-link-dialog__input {
  flex: 1;
  padding: 4px 8px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-size: var(--text-xs);
  background: var(--color-bg);
  color: var(--color-text);
}
.rte-link-dialog__input:focus-visible {
  outline: none;
  border-color: var(--color-highlight);
}
.rte-link-dialog__btn {
  padding: 4px 10px;
  border: none;
  border-radius: var(--radius-sm);
  background: var(--color-highlight);
  color: var(--color-text-on-primary);
  font-size: var(--text-xs);
  font-weight: 500;
  cursor: pointer;
  min-height: 32px;
}
.rte-link-dialog__btn--cancel {
  background: var(--color-text-lighter);
}

.rte-image-status {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--color-bg-alt);
  border-bottom: 1px solid var(--color-border);
  font-size: var(--text-xs);
  color: var(--color-text-light);
}

.rte-content {
  min-height: 300px;
  max-height: 600px;
  overflow-y: auto;
}

:deep(.rte-editor) {
  padding: 16px;
  min-height: 300px;
  font-size: 14px;
  line-height: 1.75;
  color: var(--color-text);
}
:deep(.rte-editor:focus-visible) {
  outline: 2px solid var(--color-highlight);
  outline-offset: -2px;
}

:deep(.rte-editor h1) {
  font-size: var(--text-3xl);
  font-weight: 700;
  margin: 1em 0 0.5em;
}
:deep(.rte-editor h2) {
  font-size: var(--text-2xl);
  font-weight: 600;
  margin: 0.75em 0 0.5em;
  border-bottom: 1px solid var(--color-border);
  padding-bottom: 0.25em;
}
:deep(.rte-editor h3) {
  font-size: var(--text-xl);
  font-weight: 600;
  margin: 0.5em 0;
}

:deep(.rte-editor p) {
  margin: 0.5em 0;
}

:deep(.rte-editor ul),
:deep(.rte-editor ol) {
  padding-left: 1.5em;
  margin: 0.5em 0;
}
:deep(.rte-editor ul) { list-style: disc; }
:deep(.rte-editor ol) { list-style: decimal; }

:deep(.rte-editor blockquote) {
  border-left: 3px solid var(--color-border-dark);
  padding-left: 1em;
  margin: 0.5em 0;
  color: var(--color-text-light);
}

:deep(.rte-editor code) {
  background: var(--color-bg-dark);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-size: 0.875em;
}

:deep(.rte-editor a) {
  color: var(--color-highlight);
  text-decoration: underline;
}

:deep(.rte-editor img) {
  max-width: 100%;
  height: auto;
  border-radius: var(--radius-sm);
}

:deep(.ProseMirror p.is-editor-empty:first-child::before) {
  color: var(--color-text-lighter);
  content: attr(data-placeholder);
  float: left;
  height: 0;
  pointer-events: none;
}
</style>
