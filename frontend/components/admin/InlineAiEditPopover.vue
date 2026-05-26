<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="inline-ai-popover"
      :style="{ top: `${position.top}px`, left: `${position.left}px` }"
      role="dialog"
      :aria-label="t('admin.content.inline_ai_title')"
      @mousedown.stop
    >
      <div class="inline-ai-popover__header">
        <Icon name="heroicons:sparkles" class="h-4 w-4 text-orange-500" />
        <span class="inline-ai-popover__title">{{ t('admin.content.inline_ai_title') }}</span>
        <AiHelpHint topic="content_inline_edit" size="sm" />
        <span class="inline-ai-popover__kbd">Ctrl+K</span>
        <button type="button" class="inline-ai-popover__close" :aria-label="t('admin.content.cancel')" @click="$emit('close')">
          <Icon name="heroicons:x-mark" class="h-4 w-4" />
        </button>
      </div>

      <!-- 输入指令 -->
      <div v-if="phase === 'input'" class="inline-ai-popover__body">
        <p v-if="selectedPreview" class="inline-ai-popover__selection">
          <span class="inline-ai-popover__selection-label">{{ t('admin.content.inline_ai_selection') }}</span>
          {{ selectedPreview }}
        </p>
        <textarea
          ref="inputRef"
          :value="instruction"
          rows="2"
          class="inline-ai-popover__input"
          :placeholder="t('admin.content.inline_ai_placeholder')"
          @input="$emit('update:instruction', ($event.target as HTMLTextAreaElement).value)"
          @keydown.enter.meta.prevent="$emit('submit')"
          @keydown.enter.ctrl.prevent="$emit('submit')"
          @keydown.escape.prevent="$emit('close')"
        />
        <div class="inline-ai-popover__actions">
          <button type="button" class="inline-ai-popover__btn inline-ai-popover__btn--primary" :disabled="!instruction.trim()" @click="$emit('submit')">
            {{ t('admin.content.inline_ai_generate') }}
          </button>
        </div>
      </div>

      <!-- 加载中 -->
      <div v-else-if="phase === 'loading'" class="inline-ai-popover__body inline-ai-popover__loading">
        <Icon name="heroicons:arrow-path" class="h-5 w-5 animate-spin text-orange-500" />
        <span>{{ t('admin.content.inline_ai_loading') }}</span>
      </div>

      <!-- Diff 预览 + Accept/Reject -->
      <div v-else-if="phase === 'preview'" class="inline-ai-popover__body">
        <div class="inline-ai-popover__diff">
          <div class="inline-ai-popover__diff-old">
            <span class="inline-ai-popover__diff-label">{{ t('admin.content.inline_ai_before') }}</span>
            <p>{{ originalPreview }}</p>
          </div>
          <div class="inline-ai-popover__diff-new">
            <span class="inline-ai-popover__diff-label">{{ t('admin.content.inline_ai_after') }}</span>
            <p>{{ replacementPreview }}</p>
          </div>
        </div>
        <div class="inline-ai-popover__actions">
          <button type="button" class="inline-ai-popover__btn inline-ai-popover__btn--ghost" @click="$emit('reject')">
            {{ t('admin.content.inline_ai_reject') }}
          </button>
          <button type="button" class="inline-ai-popover__btn inline-ai-popover__btn--primary" @click="$emit('accept')">
            {{ t('admin.content.inline_ai_accept') }}
          </button>
        </div>
      </div>

      <p v-if="error" class="inline-ai-popover__error">{{ error }}</p>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'

const props = defineProps<{
  open: boolean
  phase: 'input' | 'loading' | 'preview'
  position: { top: number; left: number }
  instruction: string
  selectedPreview?: string
  originalPreview?: string
  replacementPreview?: string
  error?: string
}>()

defineEmits<{
  'update:instruction': [value: string]
  submit: []
  accept: []
  reject: []
  close: []
}>()

const { t } = useI18n()
const inputRef = ref<HTMLTextAreaElement>()

watch(() => props.open, (v) => {
  if (v && props.phase === 'input') {
    nextTick(() => inputRef.value?.focus())
  }
})
</script>

<style scoped>
.inline-ai-popover {
  position: fixed;
  z-index: 10050;
  width: min(380px, calc(100vw - 24px));
  border-radius: 10px;
  border: 1px solid #e5e7eb;
  background: #fff;
  box-shadow: 0 12px 40px rgba(15, 23, 42, 0.18);
  font-size: 13px;
}
.inline-ai-popover__header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 10px;
  border-bottom: 1px solid #f3f4f6;
  background: #fafafa;
  border-radius: 10px 10px 0 0;
}
.inline-ai-popover__title {
  flex: 1;
  font-weight: 600;
  color: #111827;
}
.inline-ai-popover__kbd {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
  background: #e5e7eb;
  color: #4b5563;
}
.inline-ai-popover__close {
  border: none;
  background: transparent;
  color: #6b7280;
  cursor: pointer;
  padding: 2px;
  border-radius: 4px;
}
.inline-ai-popover__close:hover { background: #e5e7eb; }
.inline-ai-popover__body { padding: 10px; }
.inline-ai-popover__selection {
  margin: 0 0 8px;
  padding: 6px 8px;
  border-radius: 6px;
  background: #fff7ed;
  border: 1px solid #fed7aa;
  color: #9a3412;
  font-size: 11px;
  line-height: 1.4;
  max-height: 48px;
  overflow: hidden;
}
.inline-ai-popover__selection-label {
  font-weight: 600;
  margin-right: 4px;
}
.inline-ai-popover__input {
  width: 100%;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  padding: 8px;
  font-size: 13px;
  resize: vertical;
  min-height: 56px;
}
.inline-ai-popover__input:focus {
  outline: none;
  border-color: #ea580c;
  box-shadow: 0 0 0 2px rgba(234, 88, 12, 0.15);
}
.inline-ai-popover__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 8px;
}
.inline-ai-popover__btn {
  border: none;
  border-radius: 6px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
}
.inline-ai-popover__btn--primary {
  background: #ea580c;
  color: #fff;
}
.inline-ai-popover__btn--primary:hover:not(:disabled) { background: #c2410c; }
.inline-ai-popover__btn--primary:disabled { opacity: 0.5; cursor: default; }
.inline-ai-popover__btn--ghost {
  background: #f3f4f6;
  color: #374151;
}
.inline-ai-popover__btn--ghost:hover { background: #e5e7eb; }
.inline-ai-popover__loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #6b7280;
  padding: 16px 10px;
}
.inline-ai-popover__diff {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 200px;
  overflow-y: auto;
}
.inline-ai-popover__diff-old,
.inline-ai-popover__diff-new {
  padding: 8px;
  border-radius: 6px;
  font-size: 12px;
  line-height: 1.5;
}
.inline-ai-popover__diff-old {
  background: #fef2f2;
  border: 1px solid #fecaca;
}
.inline-ai-popover__diff-old p {
  margin: 4px 0 0;
  text-decoration: line-through;
  color: #991b1b;
  white-space: pre-wrap;
  word-break: break-word;
}
.inline-ai-popover__diff-new {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
}
.inline-ai-popover__diff-new p {
  margin: 4px 0 0;
  color: #166534;
  white-space: pre-wrap;
  word-break: break-word;
}
.inline-ai-popover__diff-label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.inline-ai-popover__error {
  margin: 0;
  padding: 6px 10px 10px;
  font-size: 11px;
  color: #dc2626;
}
</style>
