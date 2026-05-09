<template>
  <NuxtLink
    :to="localePath(`/products/${category.slug}`)"
    class="category-card"
    :class="`category-card--${variant}`"
  >
    <div class="category-card__image">
      <img
        :src="imgSrc"
        :alt="category.name"
        class="category-card__img"
        loading="lazy"
        @error="handleImageError"
      />
      <div class="category-card__overlay">
        <span class="category-card__count">{{ category.productCount }}+ {{ $t('nav.products') }}</span>
      </div>
      <div class="category-card__hover">
        <span class="category-card__arrow">
          <Icon name="lucide:arrow-right" size="24" />
        </span>
      </div>
    </div>
    <div class="category-card__content">
      <h3 class="category-card__title">{{ category.name }}</h3>
      <p v-if="category.description" class="category-card__description">{{ category.description }}</p>
      <span class="category-card__link">{{ $t('product.view_products') }}</span>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useLocalePath } from '#i18n'

interface Category {
  slug: string
  name: string
  description?: string
  image?: string
  productCount: number
  icon?: string
}

interface Props {
  category: Category
  variant?: 'default' | 'compact' | 'horizontal'
}

const props = defineProps<Props>()

const localePath = useLocalePath()

const FALLBACK = '/images/categories/gummy-candy.jpg'
const imgSrc = ref(props.category.image || FALLBACK)

watch(() => props.category.image, (val) => {
  imgSrc.value = val || FALLBACK
})

const handleImageError = () => {
  imgSrc.value = FALLBACK
}
</script>

<style scoped>
.category-card {
  display: block;
  background-color: white;
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: var(--shadow-md);
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
  text-decoration: none;
  color: inherit;
}

.category-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-xl);
}

.category-card__image {
  position: relative;
  aspect-ratio: 16/10;
  overflow: hidden;
  background-color: var(--color-bg-alt);
}

.category-card__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-slow);
}

.category-card:hover .category-card__img {
  transform: scale(1.05);
}

.category-card__overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: var(--spacing-md);
  background: linear-gradient(to top, rgba(0,0,0,0.6) 0%, transparent 100%);
}

.category-card__count {
  color: white;
  font-size: var(--text-sm);
  font-weight: 500;
}

.category-card__hover {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(26, 31, 58, 0.7);
  opacity: 0;
  transition: opacity var(--transition-base);
}

.category-card:hover .category-card__hover {
  opacity: 1;
}

.category-card__arrow {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  background-color: white;
  border-radius: var(--radius-full);
  color: var(--color-primary);
  transform: scale(0.8);
  transition: transform var(--transition-base);
}

.category-card:hover .category-card__arrow {
  transform: scale(1);
}

.category-card__content {
  padding: var(--spacing-lg);
  text-align: center;
}

.category-card__title {
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: var(--spacing-sm);
}

.category-card__description {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  line-height: 1.5;
  margin-bottom: var(--spacing-md);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.category-card__link {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-highlight);
  transition: gap var(--transition-fast);
}

.category-card:hover .category-card__link {
  gap: var(--spacing-sm);
}

/* Compact variant */
.category-card--compact .category-card__image {
  aspect-ratio: 1/1;
}

.category-card--compact .category-card__content {
  padding: var(--spacing-md);
}

.category-card--compact .category-card__title {
  font-size: var(--text-base);
}

.category-card--compact .category-card__description {
  display: none;
}

/* Horizontal variant */
.category-card--horizontal {
  display: grid;
  grid-template-columns: 200px 1fr;
  gap: 0;
}

.category-card--horizontal .category-card__image {
  aspect-ratio: 4/3;
}

.category-card--horizontal .category-card__content {
  text-align: left;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

@media (max-width: 640px) {
  .category-card--horizontal {
    grid-template-columns: 120px 1fr;
  }

  .category-card--horizontal .category-card__content {
    padding: var(--spacing-md);
  }
}
</style>
