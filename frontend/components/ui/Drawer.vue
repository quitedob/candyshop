<template>
  <Teleport to="body">
    <!-- Backdrop -->
    <Transition name="drawer-fade">
      <div
        v-if="open"
        class="fixed inset-0 z-40 bg-gray-500/75 transition-opacity"
        @click="$emit('close')"
      />
    </Transition>

    <!-- Drawer panel -->
    <Transition name="drawer-slide">
      <div
        v-if="open"
        class="fixed inset-y-0 right-0 z-50 flex w-full sm"
        :class="widthClass"
        role="dialog"
        aria-modal="true"
        aria-labelledby="drawer-title"
      >
        <div class="flex h-full w-full flex-col bg-white shadow-xl">
          <!-- Header -->
          <div class="flex items-center justify-between border-b px-6 py-4">
            <h2 id="drawer-title" class="text-lg font-semibold text-gray-900">
              <slot name="title" />
            </h2>
            <button
              type="button"
              class="rounded-md text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-orange-500"
              :aria-label="$t('close')"
              @click="$emit('close')"
            >
              <Icon name="heroicons:x-mark" class="h-6 w-6" />
            </button>
          </div>

          <!-- Body -->
          <div class="flex-1 overflow-y-auto p-6">
            <slot />
          </div>

          <!-- Footer -->
          <div v-if="$slots.footer" class="border-t px-6 py-4">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { watch, onBeforeUnmount } from 'vue'

const props = withDefaults(defineProps<{
  open: boolean
  width?: 'md' | 'lg' | 'xl' | '2xl' | '3xl'
}>(), {
  width: '3xl',
})

const emit = defineEmits<{
  close: []
}>()

const widthMap: Record<string, string> = {
  md: 'sm:max-w-md',
  lg: 'sm:max-w-lg',
  xl: 'sm:max-w-xl',
  '2xl': 'sm:max-w-2xl',
  '3xl': 'sm:max-w-3xl',
}
const widthClass = widthMap[props.width] || widthMap['3xl']

function onEscape(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    emit('close')
  }
}

watch(() => props.open, (val) => {
  if (val) {
    document.body.style.overflow = 'hidden'
    document.addEventListener('keydown', onEscape)
  } else {
    document.body.style.overflow = ''
    document.removeEventListener('keydown', onEscape)
  }
})

onBeforeUnmount(() => {
  document.body.style.overflow = ''
  document.removeEventListener('keydown', onEscape)
})
</script>

<style scoped>
.drawer-slide-enter-active {
  transition: transform 0.3s ease;
}
.drawer-slide-leave-active {
  transition: transform 0.2s ease;
}
.drawer-slide-enter-from,
.drawer-slide-leave-to {
  transform: translateX(100%);
}

.drawer-fade-enter-active {
  transition: opacity 0.3s ease;
}
.drawer-fade-leave-active {
  transition: opacity 0.2s ease;
}
.drawer-fade-enter-from,
.drawer-fade-leave-to {
  opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
  .drawer-slide-enter-active,
  .drawer-slide-leave-active,
  .drawer-fade-enter-active,
  .drawer-fade-leave-active {
    transition: none;
  }
}
</style>
