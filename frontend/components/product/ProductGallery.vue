<template>
  <div class="product-gallery">
    <!-- Main Image -->
    <div class="product-gallery__main">
      <button
        v-if="currentIndex > 0"
        class="product-gallery__nav product-gallery__nav--prev"
        @click="previousImage"
        :aria-label="t('common.a11y.previous_image')"
      >
        <Icon name="lucide:chevron-left" size="24" />
      </button>

      <div class="product-gallery__image-wrapper">
        <img
          :src="currentImage"
          :alt="`${alt} - Image ${currentIndex + 1}`"
          class="product-gallery__image"
          @click="openLightbox"
        />
      </div>

      <button
        v-if="currentIndex < images.length - 1"
        class="product-gallery__nav product-gallery__nav--next"
        @click="nextImage"
        :aria-label="t('common.a11y.next_image')"
      >
        <Icon name="lucide:chevron-right" size="24" />
      </button>

      <!-- Zoom indicator -->
      <button
        v-if="allowZoom"
        class="product-gallery__zoom"
        @click="openLightbox"
        :aria-label="t('common.a11y.zoom_image')"
      >
        <Icon name="lucide:zoom-in" size="18" />
      </button>
    </div>

    <!-- Thumbnails -->
    <div v-if="showThumbnails && images.length > 1" class="product-gallery__thumbnails">
      <button
        v-for="(image, index) in images"
        :key="index"
        class="product-gallery__thumbnail"
        :class="{ 'product-gallery__thumbnail--active': index === currentIndex }"
        @click="goToImage(index)"
        :aria-label="`View image ${index + 1}`"
        :aria-current="index === currentIndex ? 'true' : undefined"
      >
        <img :src="image" :alt="`${alt} - Thumbnail ${index + 1}`" loading="lazy" />
      </button>
    </div>

    <!-- Lightbox Modal -->
    <Teleport to="body">
      <div
        v-if="isLightboxOpen"
        class="product-gallery__lightbox"
        @click.self="closeLightbox"
        @keydown.esc="closeLightbox"
      >
        <button
          class="product-gallery__lightbox-close"
          @click="closeLightbox"
          :aria-label="t('common.a11y.close_lightbox')"
        >
          <Icon name="lucide:x" size="24" />
        </button>

        <button
          v-if="currentIndex > 0"
          class="product-gallery__lightbox-nav product-gallery__lightbox-nav--prev"
          @click="previousImage"
          :aria-label="t('common.a11y.previous_image')"
        >
          <Icon name="lucide:chevron-left" size="32" />
        </button>

        <div class="product-gallery__lightbox-content">
          <img
            :src="currentImage"
            :alt="`${alt} - Image ${currentIndex + 1}`"
            class="product-gallery__lightbox-image"
          />
        </div>

        <button
          v-if="currentIndex < images.length - 1"
          class="product-gallery__lightbox-nav product-gallery__lightbox-nav--next"
          @click="nextImage"
          :aria-label="t('common.a11y.next_image')"
        >
          <Icon name="lucide:chevron-right" size="32" />
        </button>

        <!-- Counter -->
        <div class="product-gallery__lightbox-counter">
          {{ currentIndex + 1 }} / {{ images.length }}
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
const { t } = useI18n()

interface Props {
  images: string[]
  alt: string
  initialIndex?: number
  showThumbnails?: boolean
  allowZoom?: boolean
  autoplay?: boolean
  autoplayInterval?: number
}

const props = withDefaults(defineProps<Props>(), {
  initialIndex: 0,
  showThumbnails: true,
  allowZoom: true,
  autoplay: false,
  autoplayInterval: 5000
})

const currentIndex = ref(props.initialIndex)
const isLightboxOpen = ref(false)
const autoplayTimer = ref<ReturnType<typeof setInterval> | null>(null)

const currentImage = computed(() => props.images[currentIndex.value] || props.images[0])

const nextImage = () => {
  currentIndex.value = (currentIndex.value + 1) % props.images.length
}

const previousImage = () => {
  currentIndex.value = currentIndex.value === 0 ? props.images.length - 1 : currentIndex.value - 1
}

const goToImage = (index: number) => {
  currentIndex.value = index
}

const openLightbox = () => {
  isLightboxOpen.value = true
  document.body.style.overflow = 'hidden'
}

const closeLightbox = () => {
  isLightboxOpen.value = false
  document.body.style.overflow = ''
}

// Keyboard navigation
const handleKeydown = (event: KeyboardEvent) => {
  if (!isLightboxOpen.value) return

  switch (event.key) {
    case 'ArrowLeft':
      previousImage()
      break
    case 'ArrowRight':
      nextImage()
      break
    case 'Escape':
      closeLightbox()
      break
  }
}

// Autoplay
const startAutoplay = () => {
  if (!props.autoplay) return
  stopAutoplay()
  autoplayTimer.value = setInterval(nextImage, props.autoplayInterval)
}

const stopAutoplay = () => {
  if (autoplayTimer.value) {
    clearInterval(autoplayTimer.value)
    autoplayTimer.value = null
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
  if (props.autoplay) {
    startAutoplay()
  }
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
  stopAutoplay()
})

// Watch for image changes to reset autoplay
watch(currentIndex, () => {
  if (props.autoplay) {
    startAutoplay()
  }
})

// Pause autoplay on hover
const pauseAutoplay = () => {
  if (props.autoplay) {
    stopAutoplay()
  }
}

const resumeAutoplay = () => {
  if (props.autoplay) {
    startAutoplay()
  }
}
</script>

<style scoped>
.product-gallery {
  --gallery-aspect-ratio: 4/3;
}

.product-gallery__main {
  position: relative;
  aspect-ratio: var(--gallery-aspect-ratio);
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.product-gallery__image-wrapper {
  width: 100%;
  height: 100%;
  cursor: zoom-in;
}

.product-gallery__image {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.product-gallery__nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background-color: white;
  border-radius: var(--radius-full);
  box-shadow: var(--shadow-md);
  color: var(--color-primary);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  z-index: 2;
}

.product-gallery__nav:hover {
  background-color: var(--color-primary);
  color: white;
}

.product-gallery__nav--prev {
  left: var(--spacing-md);
}

.product-gallery__nav--next {
  right: var(--spacing-md);
}

.product-gallery__zoom {
  position: absolute;
  bottom: var(--spacing-md);
  right: var(--spacing-md);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background-color: white;
  border-radius: var(--radius-full);
  box-shadow: var(--shadow-md);
  color: var(--color-primary);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  z-index: 2;
}

.product-gallery__zoom:hover {
  background-color: var(--color-primary);
  color: white;
}

.product-gallery__thumbnails {
  display: flex;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-md);
  overflow-x: auto;
  scrollbar-width: none;
}

.product-gallery__thumbnails::-webkit-scrollbar {
  display: none;
}

.product-gallery__thumbnail {
  flex-shrink: 0;
  width: 80px;
  height: 80px;
  padding: var(--spacing-xs);
  background-color: white;
  border: 2px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.product-gallery__thumbnail:hover {
  border-color: var(--color-accent);
}

.product-gallery__thumbnail--active {
  border-color: var(--color-highlight);
}

.product-gallery__thumbnail img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: var(--radius-sm);
}

/* Lightbox */
.product-gallery__lightbox {
  position: fixed;
  inset: 0;
  z-index: var(--z-modal);
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(0, 0, 0, 0.9);
  padding: var(--spacing-xl);
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.product-gallery__lightbox-close {
  position: absolute;
  top: var(--spacing-lg);
  right: var(--spacing-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  background-color: white;
  border-radius: var(--radius-full);
  color: var(--color-text);
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  z-index: 10;
}

.product-gallery__lightbox-close:hover {
  background-color: var(--color-error);
  color: white;
}

.product-gallery__lightbox-nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-full);
  color: white;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
  z-index: 5;
}

.product-gallery__lightbox-nav:hover {
  background-color: rgba(255, 255, 255, 0.2);
}

.product-gallery__lightbox-nav--prev {
  left: var(--spacing-lg);
}

.product-gallery__lightbox-nav--next {
  right: var(--spacing-lg);
}

.product-gallery__lightbox-content {
  max-width: 90vw;
  max-height: 90vh;
}

.product-gallery__lightbox-image {
  max-width: 100%;
  max-height: 85vh;
  object-fit: contain;
}

.product-gallery__lightbox-counter {
  position: absolute;
  bottom: var(--spacing-lg);
  left: 50%;
  transform: translateX(-50%);
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-full);
  color: white;
  font-size: var(--text-sm);
  font-weight: 500;
}

/* Aspect ratio variants */
.product-gallery--square .product-gallery__main {
  --gallery-aspect-ratio: 1/1;
}

.product-gallery--portrait .product-gallery__main {
  --gallery-aspect-ratio: 3/4;
}

.product-gallery--landscape .product-gallery__main {
  --gallery-aspect-ratio: 16/9;
}

/* Thumbnail sizes */
.product-gallery--thumbs-sm .product-gallery__thumbnail {
  width: 60px;
  height: 60px;
}

.product-gallery--thumbs-lg .product-gallery__thumbnail {
  width: 100px;
  height: 100px;
}

@media (max-width: 640px) {
  .product-gallery__nav {
    width: 32px;
    height: 32px;
  }

  .product-gallery__thumbnail {
    width: 60px;
    height: 60px;
  }

  .product-gallery__lightbox-nav {
    width: 44px;
    height: 44px;
  }

  .product-gallery__lightbox-nav--prev {
    left: var(--spacing-sm);
  }

  .product-gallery__lightbox-nav--next {
    right: var(--spacing-sm);
  }
}
</style>
