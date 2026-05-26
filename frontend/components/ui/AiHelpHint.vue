<template>
  <span ref="anchorRef" class="ai-help-hint">
    <button
      type="button"
      class="ai-help-hint__btn"
      :class="size === 'sm' ? 'ai-help-hint__btn--sm' : ''"
      :aria-expanded="open"
      :aria-label="t('admin.ai_help.button_label')"
      @click.stop="toggle"
    >!</button>

    <Teleport to="body">
      <div
        v-if="open"
        class="ai-help-hint__popover"
        :style="{ top: `${pos.top}px`, left: `${pos.left}px` }"
        role="dialog"
        :aria-label="title"
        @click.stop
      >
        <div class="ai-help-hint__popover-header">
          <span class="ai-help-hint__popover-title">{{ title }}</span>
          <button type="button" class="ai-help-hint__close" :aria-label="t('admin.ai_help.got_it')" @click="open = false">
            <Icon name="heroicons:x-mark" class="h-4 w-4" />
          </button>
        </div>
        <div class="ai-help-hint__body">{{ body }}</div>
        <button type="button" class="ai-help-hint__got-it" @click="open = false">{{ t('admin.ai_help.got_it') }}</button>
      </div>
    </Teleport>
  </span>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'

const props = withDefaults(defineProps<{
  /** i18n key suffix under admin.ai_help.* */
  topic: string
  size?: 'sm' | 'md'
}>(), {
  size: 'md',
})

const { t } = useI18n()
const open = ref(false)
const anchorRef = ref<HTMLElement>()
const pos = ref({ top: 0, left: 0 })

const title = computed(() => t(`admin.ai_help.${props.topic}.title`))
const body = computed(() => t(`admin.ai_help.${props.topic}.body`))

const POPOVER_W = 320

const updatePosition = () => {
  const el = anchorRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  const pad = 12
  let left = rect.left + rect.width / 2 - POPOVER_W / 2
  left = Math.max(pad, Math.min(left, window.innerWidth - POPOVER_W - pad))
  const top = rect.bottom + 8
  pos.value = { top: Math.min(top, window.innerHeight - 280), left }
}

const toggle = () => {
  if (!open.value) updatePosition()
  open.value = !open.value
}

const onDocClick = () => {
  if (open.value) open.value = false
}

onMounted(() => document.addEventListener('click', onDocClick))
onUnmounted(() => document.removeEventListener('click', onDocClick))
</script>

<style scoped>
.ai-help-hint {
  display: inline-flex;
  vertical-align: middle;
  margin-left: 4px;
}
.ai-help-hint__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 9999px;
  border: 1.5px solid #fb923c;
  background: #fff7ed;
  color: #ea580c;
  font-size: 12px;
  font-weight: 800;
  line-height: 1;
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.15s, transform 0.1s;
}
.ai-help-hint__btn--sm {
  width: 18px;
  height: 18px;
  font-size: 11px;
}
.ai-help-hint__btn:hover {
  background: #ffedd5;
  transform: scale(1.05);
}
.ai-help-hint__popover {
  position: fixed;
  z-index: 10060;
  width: min(320px, calc(100vw - 24px));
  border-radius: 10px;
  border: 1px solid #e5e7eb;
  background: #fff;
  box-shadow: 0 12px 40px rgba(15, 23, 42, 0.16);
  padding: 12px;
}
.ai-help-hint__popover-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 8px;
}
.ai-help-hint__popover-title {
  font-size: 14px;
  font-weight: 600;
  color: #111827;
  line-height: 1.35;
}
.ai-help-hint__close {
  border: none;
  background: transparent;
  color: #9ca3af;
  cursor: pointer;
  padding: 2px;
  border-radius: 4px;
  flex-shrink: 0;
}
.ai-help-hint__close:hover { color: #4b5563; background: #f3f4f6; }
.ai-help-hint__body {
  font-size: 12px;
  line-height: 1.65;
  color: #4b5563;
  white-space: pre-line;
  max-height: 240px;
  overflow-y: auto;
}
.ai-help-hint__got-it {
  margin-top: 10px;
  width: 100%;
  border: none;
  border-radius: 6px;
  background: #ea580c;
  color: #fff;
  font-size: 12px;
  font-weight: 500;
  padding: 7px 12px;
  cursor: pointer;
}
.ai-help-hint__got-it:hover { background: #c2410c; }
</style>
