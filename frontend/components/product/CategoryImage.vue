<template>
  <img
    v-if="imageSource"
    ref="imageElementRef"
    :src="imageSource"
    :alt="alt"
    loading="lazy"
    @error="handleImageError"
  />
  <div v-else class="category-image-placeholder" aria-hidden="true">
    <Icon name="lucide:package" size="48" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'

const props = defineProps<{
  slug: string
  src?: string
  alt: string
}>()

// Only these category photographs are bundled in public/images/categories.
// Custom categories must not produce guessed URLs for files that do not exist.
const BUNDLED_CATEGORY_IMAGE_SLUGS = new Set([
  'aerated-candy',
  'compound-chocolate',
  'gummy-candy',
  'hard-candy',
  'licorice',
  'sour-candies',
  'toffee-candy',
])

const bundledImageSource = computed(() => BUNDLED_CATEGORY_IMAGE_SLUGS.has(props.slug)
  ? `/images/categories/${props.slug}.jpg`
  : '')
const imageSource = ref('')
const imageElementRef = ref<HTMLImageElement | null>(null)

watch([() => props.src, bundledImageSource], () => {
  imageSource.value = props.src || bundledImageSource.value
}, { immediate: true })

const handleImageError = () => {
  imageSource.value = imageSource.value === bundledImageSource.value ? '' : bundledImageSource.value
}

onMounted(() => {
  // An SSR image can fail before hydration attaches its error handler.
  if (imageElementRef.value?.complete && imageElementRef.value.naturalWidth === 0) handleImageError()
})
</script>

<style scoped>
.category-image-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: var(--color-text-light);
  background: var(--color-bg-alt);
}
</style>
