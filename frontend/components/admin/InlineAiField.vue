<template>
  <div class="inline-ai-field">
    <label v-if="label" :for="inputId" class="admin-form-label">{{ label }}</label>
    <component
      :is="multiline ? 'textarea' : 'input'"
      :id="inputId"
      ref="inputRef"
      :value="modelValue"
      :rows="multiline ? rows : undefined"
      :type="multiline ? undefined : 'text'"
      autocomplete="off"
      class="admin-form-control"
      v-bind="$attrs"
      @input="onInput"
      @keydown="onKeydown"
      @mouseup="captureSelection"
      @keyup="captureSelection"
    />
    <InlineAiEditPopover
      v-if="enabled"
      :open="popover.open"
      :phase="popover.phase"
      :position="popover.position"
      :instruction="popover.instruction"
      :selected-preview="truncate(popover.selectedText)"
      :original-preview="truncate(stripHtml(popover.selectedText))"
      :replacement-preview="truncate(stripHtml(popover.replacement))"
      :error="popover.error"
      @update:instruction="popover.instruction = $event"
      @submit="submitEdit"
      @accept="acceptEdit"
      @reject="rejectEdit"
      @close="closePopover"
    />
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { InlineAiFieldType } from '~/composables/useInlineAiEdit'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{
  modelValue: string
  enabled?: boolean
  fieldType?: InlineAiFieldType
  label?: string
  inputId?: string
  multiline?: boolean
  rows?: number
  language?: string
}>(), {
  enabled: false,
  fieldType: 'plain',
  multiline: false,
  rows: 2,
  language: 'zh',
})

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const { t } = useI18n()
const { requestInlineEdit, clampPopoverPosition } = useInlineAiEdit(props.language)
const inputRef = ref<HTMLInputElement | HTMLTextAreaElement>()
const selectionRange = ref({ start: 0, end: 0 })

const popover = reactive({
  open: false,
  phase: 'input' as 'input' | 'loading' | 'preview',
  instruction: '',
  position: { top: 0, left: 0 },
  selectedText: '',
  contextBefore: '',
  contextAfter: '',
  replacement: '',
  error: '',
})

const stripHtml = (s: string) => s.replace(/<[^>]+>/g, '').trim()
const truncate = (s: string, max = 120) => (s.length > max ? `${s.slice(0, max)}…` : s)

const onInput = (e: Event) => {
  emit('update:modelValue', (e.target as HTMLInputElement).value)
}

const captureSelection = () => {
  if (!props.enabled) return
  const el = inputRef.value
  if (!el) return
  const start = el.selectionStart ?? 0
  const end = el.selectionEnd ?? 0
  if (start !== end) selectionRange.value = { start, end }
}

const openPopover = () => {
  const el = inputRef.value
  if (!el) return
  const { start, end } = selectionRange.value
  if (start === end) return
  const value = el.value
  const selected = value.slice(start, end)
  if (selected.trim().length < 1) return

  const rect = el.getBoundingClientRect()
  const lineHeight = 22
  popover.selectedText = selected
  popover.contextBefore = value.slice(Math.max(0, start - 300), start)
  popover.contextAfter = value.slice(end, Math.min(value.length, end + 300))
  popover.instruction = ''
  popover.replacement = ''
  popover.error = ''
  popover.phase = 'input'
  popover.position = clampPopoverPosition(rect.top + lineHeight, rect.left)
  popover.open = true
}

const onKeydown = (e: KeyboardEvent) => {
  if (!props.enabled) return
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    captureSelection()
    openPopover()
  }
}

const closePopover = () => {
  popover.open = false
  popover.phase = 'input'
  popover.error = ''
}

const submitEdit = async () => {
  if (!popover.instruction.trim()) return
  popover.phase = 'loading'
  popover.error = ''
  try {
    const res = await requestInlineEdit({
      instruction: popover.instruction.trim(),
      selectedText: popover.selectedText,
      contextBefore: popover.contextBefore,
      contextAfter: popover.contextAfter,
      fieldType: props.fieldType,
      language: props.language,
    })
    popover.replacement = res.replacement
    popover.phase = 'preview'
  } catch {
    popover.error = t('admin.content.inline_ai_failed')
    popover.phase = 'input'
  }
}

const acceptEdit = () => {
  const el = inputRef.value
  if (!el) return
  const { start, end } = selectionRange.value
  const value = el.value
  const next = value.slice(0, start) + popover.replacement + value.slice(end)
  emit('update:modelValue', next)
  closePopover()
}

const rejectEdit = () => {
  popover.phase = 'input'
  popover.replacement = ''
}
</script>
