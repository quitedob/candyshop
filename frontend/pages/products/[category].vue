<template>
  <div class="category-page">
    <!-- Breadcrumb -->
    <div class="container">
      <Breadcrumb :items="breadcrumbItems" />
    </div>

    <!-- Category Header -->
    <section class="category-header section-sm">
      <div class="container">
        <div v-if="categoryStatus === 'pending'" class="category-header__inner">
          <SkeletonLoader width="100%" height="200px" borderRadius="16px" />
        </div>
        <div v-else-if="categoryStatus === 'error' || !categoryData">
          <ErrorState
            :title="t('errors.default')"
            :message="t('offline_message')"
            :actionText="t('errors.tryAgain')"
            @action="refreshCategory"
          />
        </div>
        <div v-else class="category-header__inner">
          <div class="category-header__content">
            <span class="category-header__badge">{{ category.name }}</span>
            <h1 class="category-header__title">{{ category.name }}</h1>
            <p class="category-header__description">{{ category.description }}</p>

            <div class="category-header__meta">
              <span v-if="category.productCount" class="category-header__count">
                {{ category.productCount }}+ {{ $t('nav.products') }}
              </span>
              <span v-if="category.minMOQ" class="category-header__moq">
                {{ $t('oem.moq') }}: {{ category.minMOQ }}
              </span>
              <span v-if="category.leadTime" class="category-header__lead">
                {{ $t('oem.lead_time') }}: {{ category.leadTime }}
              </span>
            </div>
          </div>

          <div v-if="category.image" class="category-header__image">
            <img :src="category.image" :alt="category.name" />
          </div>
        </div>
      </div>
    </section>

    <!-- Filter Tags -->
    <section class="filters section-sm">
      <div class="container">
        <div class="filters__inner">
          <div class="filters__label">{{ $t('filters.title') }}:</div>

          <div class="filters__tags">
            <button
              class="filter-tag"
              :class="{ 'filter-tag--active': !activeFilters.length }"
              @click="clearFilters"
            >
              {{ $t('filters.all') }}
            </button>

            <button
              v-for="tag in filterTags"
              :key="tag.id"
              class="filter-tag"
              :class="{ 'filter-tag--active': activeFilters.includes(tag.id) }"
              @click="toggleFilter(tag.id)"
            >
              <Icon v-if="tag.icon" :name="tag.icon" size="14" />
              {{ tag.label }}
            </button>
          </div>

          <button
            class="filters__toggle"
            :class="{ 'filters__toggle--active': showMobileFilters }"
            @click="showMobileFilters = !showMobileFilters"
          >
            <Icon name="lucide:sliders-horizontal" size="18" />
            <span class="hide-mobile">{{ $t('filters.title') }}</span>
          </button>
        </div>
      </div>
    </section>

    <!-- Products Grid -->
    <section class="products section">
      <div class="container">
        <div v-if="productsStatus === 'pending'" class="products__grid">
          <SkeletonLoader v-for="i in 8" :key="i" width="100%" height="320px" borderRadius="16px" />
        </div>
        <div v-else-if="productsStatus === 'error'">
          <ErrorState
            :title="t('errors.default')"
            :message="t('offline_message')"
            :actionText="t('errors.tryAgain')"
            @action="refreshProducts"
          />
        </div>
        <div v-else-if="products.length === 0" class="products__empty">
          <div class="products__empty-icon">
            <Icon name="lucide:search" size="48" />
          </div>
          <h3>{{ $t('product.no_products') || 'No products found' }}</h3>
          <p>{{ $t('product.try_different_filters') || 'Try adjusting your filters' }}</p>
          <button class="btn btn-outline" @click="clearFilters">
            {{ $t('filters.clear') || 'Clear Filters' }}
          </button>
        </div>

        <div v-else class="products__grid">
          <ProductCard
            v-for="product in products"
            :key="product.id"
            :product="product"
            @inquire="handleInquire"
            @sample="handleSample"
          />
        </div>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="products__pagination">
          <button
            class="pagination__btn"
            :disabled="currentPage === 1"
            @click="goToPage(currentPage - 1)"
          >
            <Icon name="lucide:chevron-left" size="18" />
            {{ $t('pagination.prev') }}
          </button>

          <div class="pagination__pages">
            <button
              v-for="page in visiblePages"
              :key="page"
              class="pagination__page"
              :class="{ 'pagination__page--active': page === currentPage }"
              @click="goToPage(page)"
            >
              {{ page }}
            </button>
          </div>

          <button
            class="pagination__btn"
            :disabled="currentPage === totalPages"
            @click="goToPage(currentPage + 1)"
          >
            {{ $t('pagination.next') }}
            <Icon name="lucide:chevron-right" size="18" />
          </button>
        </div>

        <div v-if="total > 0" class="products__info">
          {{ $t('pagination.showing').replace('{from}', String((currentPage - 1) * limit + 1))
            .replace('{to}', String(Math.min(currentPage * limit, total)))
            .replace('{total}', String(total)) }}
        </div>
      </div>
    </section>

    <!-- Factory Capability -->
    <section class="capability section bg-alt">
      <div class="container">
        <div class="section-header">
          <h2>{{ $t('factory.capacity') }}</h2>
        </div>

        <div class="capability__grid">
          <div class="capability-card">
            <Icon name="lucide:box" size="32" />
            <h4>{{ $t('oem.moq') }}</h4>
            <p>{{ category.minMOQ || '1,000 pcs' }}</p>
          </div>
          <div class="capability-card">
            <Icon name="lucide:clock" size="32" />
            <h4>{{ $t('oem.lead_time') }}</h4>
            <p>{{ category.leadTime || '2-4 weeks' }}</p>
          </div>
          <div class="capability-card">
            <Icon name="lucide:package" size="32" />
            <h4>{{ $t('factory.daily_output') }}</h4>
            <p>{{ category.dailyCapacity || '10-20 tons' }}</p>
          </div>
          <div class="capability-card">
            <Icon name="lucide:settings" size="32" />
            <h4>{{ $t('oem.custom_packaging') }}</h4>
            <p>{{ category.packagingOptions || 'Bag, Jar, Gift Box' }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Related Categories -->
    <section v-if="relatedCategories.length" class="related section">
      <div class="container">
        <div class="section-header">
          <h2>{{ $t('product.categories_title') }}</h2>
        </div>

        <div class="related__grid">
          <CategoryCard
            v-for="cat in relatedCategories"
            :key="cat.slug"
            :category="cat"
            variant="compact"
          />
        </div>
      </div>
    </section>

    <!-- FAQ Section -->
    <section class="faq section bg-alt">
      <div class="container container-narrow">
        <div class="section-header">
          <h2>{{ $t('home.faq.title') }}</h2>
        </div>

        <div class="faq__list">
          <div
            v-for="(faq, index) in categoryFAQs"
            :key="index"
            class="faq-item"
            :class="{ 'faq-item--open': openFAQ === index }"
          >
            <button class="faq-item__question" @click="toggleFAQ(index)">
              <span>{{ faq.question }}</span>
              <Icon name="lucide:chevron-down" size="18" />
            </button>
            <div v-if="openFAQ === index" class="faq-item__answer">
              {{ faq.answer }}
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="cta section">
      <div class="container">
        <div class="cta__inner">
          <h2>{{ $t('product.inquire_now') }}</h2>
          <p>{{ $t('form.subtitle') }}</p>
          <div class="cta__actions">
            <a :href="whatsappUrl" target="_blank" rel="noopener noreferrer" class="btn btn-highlight">
              <WhatsAppIcon size="20" />
              {{ $t('whatsapp.us') }}
            </a>
            <NuxtLink :to="localePath('/contact')" class="btn btn-outline">
              {{ $t('form.submit') }}
            </NuxtLink>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n, useLocalePath } from '#i18n'

const { t, locale } = useI18n()
const localePath = useLocalePath()
const route = useRoute()
const router = useRouter()
const config = useRuntimeConfig()
const { getProducts, getCategory } = useApi()
const { isAuthenticated } = useAuth()

// State
const currentPage = ref(1)
const limit = 12
const activeFilters = ref<string[]>([])
const showMobileFilters = ref(false)
const openFAQ = ref(0)

// Category slug
const categorySlug = computed(() => route.params.category as string)

// Fetch category and products
const { data: categoryData, status: categoryStatus, refresh: refreshCategory } = await useAsyncData(
  `category-${categorySlug.value}-${locale.value}`,
  async () => {
    return await getCategory(categorySlug.value)
  }
)

const category = computed(() => {
  const data = categoryData.value
  if (!data) {
    return {
      name: t(`product.categories.${categorySlug.value.replace(/-/g, '_')}`) || categorySlug.value,
      description: '',
      productCount: 0,
      image: `/images/categories/${categorySlug.value}.jpg`
    }
  }
  return {
    ...data,
    image: data.thumbnail || `/images/categories/${categorySlug.value}.jpg`
  }
})

// Fetch products
const { data: productsData, status: productsStatus, refresh: refreshProducts } = await useAsyncData(
  `products-${categorySlug.value}-${locale.value}-${currentPage.value}`,
  async () => {
    return await getProducts({
      category: categorySlug.value,
      page: currentPage.value,
      limit
    })
  }
)

const products = computed(() => {
  const all = productsData.value?.data || []
  if (!activeFilters.value.length) return all
  return all.filter(p => {
    return activeFilters.value.every(filter => {
      if (filter === 'halal') return p.halalCertified
      if (filter === 'sugar-free') {
        const text = `${p.name} ${p.summary || ''} ${p.description || ''}`.toLowerCase()
        return text.includes('sugar-free') || text.includes('sugar free') || text.includes('无糖')
      }
      if (filter === '4d') {
        const text = `${p.name} ${p.summary || ''}`.toLowerCase()
        return text.includes('4d') || text.includes('4-d')
      }
      if (filter === 'filled') {
        const text = `${p.name} ${p.summary || ''} ${p.description || ''}`.toLowerCase()
        return text.includes('filled') || text.includes('filling') || text.includes('夹心')
      }
      if (filter === 'vitamin') {
        const text = `${p.name} ${p.summary || ''} ${p.description || ''}`.toLowerCase()
        return text.includes('vitamin') || text.includes('维生素')
      }
      return true
    })
  })
})
const total = computed(() => productsData.value?.pagination?.total || 0)
const totalPages = computed(() => productsData.value?.pagination?.totalPages || 1)

// Breadcrumb
const breadcrumbItems = computed(() => [
  { label: t('nav.products'), to: '/products' },
  { label: category.value.name }
])

// Filter tags
const filterTags = [
  { id: 'sugar-free', label: t('filters.sugar_free'), icon: 'lucide:leaf' },
  { id: 'halal', label: t('filters.halal'), icon: 'lucide:star' },
  { id: 'filled', label: t('filters.filled'), icon: 'lucide:circle' },
  { id: '4d', label: t('filters.4d'), icon: 'lucide:box' },
  { id: 'vitamin', label: t('filters.vitamin'), icon: 'lucide:pill' }
]

// Related categories
const relatedCategories = computed(() => {
  const allCategories = [
    { slug: 'gummy-candy', name: t('product.categories.gummy_candy'), productCount: 120 },
    { slug: 'hard-candy', name: t('product.categories.hard_candy'), productCount: 85 },
    { slug: 'aerated-candy', name: t('product.categories.aerated_candy'), productCount: 45 },
    { slug: 'toffee-candy', name: t('product.categories.toffee_candy'), productCount: 35 },
    { slug: 'compound-chocolate', name: t('product.categories.compound_chocolate'), productCount: 60 },
    { slug: 'licorice', name: t('product.categories.licorice'), productCount: 25 },
    { slug: 'sour-candies', name: t('product.categories.sour_candies'), productCount: 40 }
  ]
  return allCategories.filter(c => c.slug !== categorySlug.value)
})

// Category FAQs
const categoryFAQs = [
  {
    question: t('oem.moq') + '?',
    answer: category.value.minMOQ || 'Our minimum order quantity starts from 1,000 pieces for standard items.'
  },
  {
    question: t('oem.lead_time') + '?',
    answer: category.value.leadTime || 'Standard orders take 2-3 weeks, custom OEM orders take 4-6 weeks.'
  },
  {
    question: t('factory.certifications') + '?',
    answer: 'We are HACCP, ISO22000, BRC Grade A, and Halal certified.'
  }
]

// Pagination
const visiblePages = computed(() => {
  const pages: number[] = []
  const showPages = 5
  let start = Math.max(1, currentPage.value - Math.floor(showPages / 2))
  let end = Math.min(totalPages.value, start + showPages - 1)

  if (end - start < showPages - 1) {
    start = Math.max(1, end - showPages + 1)
  }

  for (let i = start; i <= end; i++) {
    pages.push(i)
  }

  return pages
})

const goToPage = (page: number) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
  refreshProducts()
}

const toggleFilter = (filterId: string) => {
  const index = activeFilters.value.indexOf(filterId)
  if (index === -1) {
    activeFilters.value.push(filterId)
  } else {
    activeFilters.value.splice(index, 1)
  }
}

const clearFilters = () => {
  activeFilters.value = []
}

const toggleFAQ = (index: number) => {
  openFAQ.value = openFAQ.value === index ? -1 : index
}

const handleInquire = (product: any) => {
  if (!isAuthenticated.value) {
    router.push({ path: localePath('/auth/login'), query: { redirect: route.fullPath } })
    return
  }
  router.push({
    path: localePath('/contact'),
    query: { product: product.name }
  })
}

const handleSample = (product: any) => {
  if (!isAuthenticated.value) {
    router.push({ path: localePath('/auth/login'), query: { redirect: route.fullPath } })
    return
  }
  router.push({
    path: localePath('/contact'),
    query: { product: product.name, sample: 'true' }
  })
}

// WhatsApp URL
const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  const message = encodeURIComponent(`Hi, I'm interested in your ${category.value.name} products. Can you provide more information?`)
  return `https://wa.me/${number}?text=${message}`
})

// SEO
useSeo({
  title: `${category.value.name} | ${t('seo.default_title')}`,
  description: category.value.description || t('seo.default_description'),
  ogType: 'website',
  schema: useBreadcrumbSchema([
    { name: t('nav.home'), item: config.public.siteUrl },
    { name: t('nav.products'), item: `${config.public.siteUrl}/products` },
    { name: category.value.name, item: `${config.public.siteUrl}/products/${categorySlug.value}` }
  ])
})
</script>

<style scoped>
.category-header__inner {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-3xl);
}

@media (min-width: 1024px) {
  .category-header__inner {
    grid-template-columns: 1fr 1fr;
    align-items: center;
  }
}

.category-header__badge {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-md);
  background-color: var(--color-accent);
  color: white;
  font-size: var(--text-sm);
  font-weight: 600;
  border-radius: var(--radius-full);
  margin-bottom: var(--spacing-md);
}

.category-header__title {
  font-size: clamp(2rem, 4vw, 3rem);
  margin-bottom: var(--spacing-md);
}

.category-header__description {
  font-size: var(--text-lg);
  color: var(--color-text-light);
  max-width: 600px;
}

.category-header__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-md);
  margin-top: var(--spacing-xl);
}

.category-header__count,
.category-header__moq,
.category-header__lead {
  display: inline-flex;
  align-items: center;
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  font-weight: 500;
}

.category-header__image img {
  width: 100%;
  border-radius: var(--radius-2xl);
  box-shadow: var(--shadow-xl);
}

/* Filters */
.filters__inner {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex-wrap: wrap;
}

.filters__label {
  font-weight: 600;
  color: var(--color-text);
}

.filters__tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
  flex: 1;
}

.filter-tag {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs) var(--spacing-md);
  font-size: var(--text-sm);
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.filter-tag:hover {
  border-color: var(--color-accent);
}

.filter-tag--active {
  background-color: var(--color-primary);
  border-color: var(--color-primary);
  color: white;
}

.filters__toggle {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
}

/* Products */
.products__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--spacing-xl);
  margin-bottom: var(--spacing-3xl);
}

.products__empty {
  text-align: center;
  padding: var(--spacing-5xl) var(--spacing-lg);
}

.products__empty-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  margin-bottom: var(--spacing-lg);
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-full);
  color: var(--color-text-light);
}

.products__empty h3 {
  margin-bottom: var(--spacing-sm);
}

.products__empty p {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-lg);
}

.products__pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
}

.pagination__btn {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.pagination__btn:hover:not(:disabled) {
  border-color: var(--color-highlight);
}

.pagination__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pagination__pages {
  display: flex;
  gap: var(--spacing-xs);
}

.pagination__page {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.pagination__page:hover {
  border-color: var(--color-accent);
}

.pagination__page--active {
  background-color: var(--color-primary);
  border-color: var(--color-primary);
  color: white;
}

.products__info {
  text-align: center;
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

/* Capability */
.capability__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-lg);
}

.capability-card {
  text-align: center;
  padding: var(--spacing-xl);
  background-color: white;
  border-radius: var(--radius-lg);
}

.capability-card svg {
  color: var(--color-accent);
  margin-bottom: var(--spacing-md);
}

.capability-card h4 {
  font-size: var(--text-lg);
  margin-bottom: var(--spacing-xs);
}

.capability-card p {
  color: var(--color-text-light);
}

/* Related */
.related__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: var(--spacing-lg);
}

/* FAQ */
.faq__list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.faq-item {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.faq-item__question {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: var(--spacing-lg);
  font-weight: 500;
  background: none;
  border: none;
  cursor: pointer;
}

.faq-item__answer {
  padding: 0 var(--spacing-lg) var(--spacing-lg);
  color: var(--color-text-light);
}

/* CTA */
.cta__inner {
  text-align: center;
  max-width: 600px;
  margin: 0 auto;
}

.cta__inner h2 {
  margin-bottom: var(--spacing-md);
}

.cta__inner p {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-xl);
}

.cta__actions {
  display: flex;
  justify-content: center;
  gap: var(--spacing-md);
  flex-wrap: wrap;
}
</style>
