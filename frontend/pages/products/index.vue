<template>
  <div class="products-page">
    <!-- Hero Section -->
    <section class="products-hero section">
      <div class="container">
        <h1 class="products-hero__title">{{ $t('nav.products') }}</h1>
        <p class="products-hero__description">{{ $t('product.browse_all') }}</p>
      </div>
    </section>

    <!-- Categories Grid -->
    <section class="products-categories section section-lg">
      <div class="container">
        <div v-if="categoriesStatus === 'pending'" class="categories-grid">
          <SkeletonLoader v-for="i in 6" :key="i" width="100%" height="300px" borderRadius="16px" />
        </div>
        <div v-else-if="categoriesStatus === 'error' || !categories.length">
          <ErrorState
            :title="t('errors.default')"
            :message="t('offline_message')"
            :actionText="t('errors.tryAgain')"
            @action="refreshCategories"
          />
        </div>
        <div v-else class="categories-grid">
          <NuxtLink
            v-for="category in categories"
            :key="category.slug"
            :to="localePath(`/products/${category.slug}`)"
            class="category-card"
          >
            <div class="category-card__image">
              <img :src="category.image" :alt="category.name" loading="lazy" />
              <div class="category-card__overlay">
                <span class="category-card__arrow">
                  <Icon name="lucide:arrow-right" size="24" />
                </span>
              </div>
            </div>
            <div class="category-card__content">
              <h2 class="category-card__title">{{ category.name }}</h2>
              <p class="category-card__count">{{ category.productCount }} {{ $t('product.products') }}</p>
              <p class="category-card__description">{{ category.description }}</p>
            </div>
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- All Products Grid -->
    <section class="products-all section section-lg">
      <div class="container">
        <h2 class="section-title">{{ $t('product.all_products') }}</h2>
        <div v-if="productsStatus === 'pending'" class="products-grid">
          <SkeletonLoader v-for="i in 8" :key="i" width="100%" height="300px" borderRadius="16px" />
        </div>
        <div v-else-if="productsError" class="text-error">{{ productsError }}</div>
        <div v-else-if="!products || products.length === 0" class="text-center text-light py-8">{{ $t('product.no_products') }}</div>
        <div v-else class="products-grid">
          <ProductCard v-for="product in products" :key="product.id" :product="product" />
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="products-cta section bg-alt">
      <div class="container">
        <div class="cta-content">
          <h2>{{ $t('product.need_custom') }}</h2>
          <p>{{ $t('product.oem_description') }}</p>
          <div class="cta-actions">
            <NuxtLink :to="orderingCtaPath" class="btn btn-highlight btn-lg">
              {{ orderingCtaLabel }}
            </NuxtLink>
            <NuxtLink :to="localePath('/oem-solutions')" class="btn btn-outline btn-lg">
              {{ $t('nav.oem') }}
            </NuxtLink>
            <NuxtLink :to="localePath('/contact')" class="btn btn-ghost btn-lg">
              {{ $t('contact.send_inquiry') }}
            </NuxtLink>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n, useLocalePath } from '#i18n'

const { t, locale } = useI18n()
const localePath = useLocalePath()
const { isAuthenticated, isAdmin, isPending } = useAuth()
const { getCategories, getProducts } = useApi()

const orderingCtaPath = computed(() => {
  if (isAuthenticated.value) {
    return localePath(isAdmin.value ? '/admin' : '/customer/products')
  }
  return localePath('/auth/register')
})

const orderingCtaLabel = computed(() => {
  if (isAuthenticated.value && isAdmin.value) {
    return t('nav.admin_panel')
  }
  if (isPending.value) {
    return t('product.pending_approval')
  }
  if (isAuthenticated.value) {
    return t('product.inquire_now')
  }
  return t('product.register_to_order')
})

const { data: categoriesData, status: categoriesStatus, refresh: refreshCategories } = await useAsyncData(
  () => `products-categories-${locale.value}`,
  async () => {
    return await getCategories()
  }
)

const { data: productsData, status: productsStatus, error: productsError, refresh: refreshProducts } = await useAsyncData(
  `all-products-${locale.value}`,
  async () => {
    const result = await getProducts({ limit: 12 })
    return result.data || []
  }
)

const products = computed(() => productsData.value || [])

const categories = computed(() => {
  return (categoriesData.value || []).map(category => ({
    slug: category.slug,
    name: category.name || t(`product.categories.${category.slug.replace(/-/g, '_')}`),
    count: category.productCount || 0,
    description: category.description || t(`product.category_descriptions.${category.slug.replace(/-/g, '_')}`),
    image: category.thumbnail || `/images/categories/${category.slug}.jpg`
  }))
})

// SEO
useSeo({
  title: t('seo.products_title'),
  description: t('seo.products_description'),
  ogType: 'website'
})
</script>

<style scoped>
.products-hero {
  padding-top: calc(var(--header-height) + var(--spacing-2xl));
  text-align: center;
  background: linear-gradient(135deg, #fff5f0 0%, #fef3e2 100%);
}

.products-hero__title {
  font-size: clamp(2rem, 4vw, 3rem);
  margin-bottom: var(--spacing-md);
  color: var(--color-primary);
}

.products-hero__description {
  font-size: var(--text-lg);
  color: var(--color-text-light);
  max-width: 600px;
  margin: 0 auto;
}

.categories-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-xl);
}

.products-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: var(--spacing-xl);
}

.category-card {
  display: flex;
  flex-direction: column;
  background-color: white;
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: var(--shadow-md);
  transition: all var(--transition-base);
  text-decoration: none;
  color: inherit;
}

.category-card:hover {
  transform: translateY(-8px);
  box-shadow: var(--shadow-xl);
}

.category-card__image {
  position: relative;
  aspect-ratio: 16/9;
  overflow: hidden;
  background-color: var(--color-bg-alt);
}

.category-card__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-slow);
}

.category-card:hover .category-card__image img {
  transform: scale(1.05);
}

.category-card__overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(0,0,0,0.6) 0%, transparent 50%);
  opacity: 0;
  transition: opacity var(--transition-base);
  display: flex;
  align-items: center;
  justify-content: center;
}

.category-card:hover .category-card__overlay {
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
  transform: translateY(20px);
  transition: transform var(--transition-base);
}

.category-card:hover .category-card__arrow {
  transform: translateY(0);
}

.category-card__content {
  padding: var(--spacing-lg);
  flex: 1;
}

.category-card__title {
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: var(--spacing-xs);
}

.category-card__count {
  font-size: var(--text-sm);
  color: var(--color-highlight);
  font-weight: 500;
  margin-bottom: var(--spacing-sm);
}

.category-card__description {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  line-height: 1.6;
}

.products-cta {
  text-align: center;
}

.cta-content {
  max-width: 700px;
  margin: 0 auto;
}

.cta-content h2 {
  margin-bottom: var(--spacing-md);
}

.cta-content p {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-xl);
}

.cta-actions {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: var(--spacing-md);
}
</style>
