<template>
  <div class="blog-page">
    <!-- Breadcrumb -->
    <div class="container">
      <Breadcrumb :items="[{ label: t('blog.title') }]" />
    </div>

    <!-- Hero -->
    <section class="blog-hero section">
      <div class="container">
        <div class="blog-hero__inner">
          <h1 class="blog-hero__title">{{ t('blog.title') }}</h1>
          <p class="blog-hero__subtitle">
            {{ t('blog_extra.subscribe_desc') }}
          </p>
        </div>
      </div>
    </section>

    <!-- Categories -->
    <section class="categories section-sm">
      <div class="container">
        <div class="categories__inner">
          <button
            class="category-tab"
            :class="{ 'category-tab--active': !activeCategory }"
            @click="setCategory('')"
          >
            {{ t('blog_extra.all_posts') }}
          </button>
          <button
            v-for="cat in categories"
            :key="cat.id"
            class="category-tab"
            :class="{ 'category-tab--active': activeCategory === cat.id }"
            @click="setCategory(cat.id)"
          >
            {{ cat.name }}
          </button>
        </div>
      </div>
    </section>

    <!-- Featured Post -->
    <article
      v-if="featuredPost && !activeCategory"
      class="featured-post section"
    >
      <div class="container">
        <NuxtLink
          :to="localePath(`/blog/${featuredPost.slug}`)"
          class="featured-post__inner"
        >
          <div class="featured-post__image">
            <img :src="featuredPost.thumbnail || '/images/blog/blog-placeholder.jpg'" :alt="tField(featuredPost, 'title')" />
            <span class="featured-post__badge">{{ t('blog_extra.featured') }}</span>
          </div>
          <div class="featured-post__content">
            <span class="featured-post__category">{{ getCategoryName(featuredPost.category) }}</span>
            <h2 class="featured-post__title">{{ tField(featuredPost, 'title') }}</h2>
            <p class="featured-post__excerpt">{{ tField(featuredPost, 'excerpt') }}</p>
            <div class="featured-post__meta">
              <span class="featured-post__author">
                <Icon name="lucide:user" size="14" />
                {{ featuredPost.author?.name || 'CandyPro' }}
              </span>
              <span class="featured-post__date">
                <Icon name="lucide:calendar" size="14" />
                {{ fmtDate(featuredPost.publishedAt) }}
              </span>
              <span class="featured-post__read-time">
                <Icon name="lucide:clock" size="14" />
                {{ t('blog_extra.min_read', { count: featuredPost.readTime }) }}
              </span>
            </div>
          </div>
        </NuxtLink>
      </div>
    </article>

    <!-- Posts Grid -->
    <section class="posts section">
      <div class="container">
        <div v-if="pending" class="posts__empty">
          <p>
            {{ t('blog_extra.loading_posts') }}
          </p>
        </div>

        <ErrorState
          v-else-if="postsError"
          :title="t('errors.default')"
          :message="t('offline_message')"
          :actionText="t('errors.tryAgain')"
          @action="refreshPosts"
        />

        <div v-else-if="filteredPosts.length === 0" class="posts__empty">
          <p>{{ t('blog_extra.no_posts') }}</p>
        </div>

        <div v-else class="posts__grid">
          <article
            v-for="post in pagedPosts"
            :key="post.id"
            class="post-card"
          >
            <NuxtLink :to="localePath(`/blog/${post.slug}`)" class="post-card__link">
              <div class="post-card__image">
                <img :src="post.thumbnail || '/images/blog/blog-placeholder.jpg'" :alt="tField(post, 'title')" />
                <span class="post-card__category">{{ getCategoryName(post.category) }}</span>
              </div>
              <div class="post-card__content">
                <h3 class="post-card__title">{{ tField(post, 'title') }}</h3>
                <p class="post-card__excerpt">{{ tField(post, 'excerpt') }}</p>
                <div class="post-card__meta">
                  <span>{{ fmtDate(post.publishedAt) }}</span>
                  <span>{{ t('blog_extra.min_read', { count: post.readTime }) }}</span>
                </div>
              </div>
            </NuxtLink>
          </article>
        </div>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="posts__pagination">
          <button
            class="pagination__btn"
            :disabled="currentPage === 1"
            @click="goToPage(currentPage - 1)"
          >
            {{ t('pagination.prev') }}
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
            {{ t('pagination.next') }}
          </button>
        </div>
      </div>
    </section>

    <!-- Contact CTA -->
    <section class="newsletter section bg-alt">
      <div class="container container-narrow">
        <div class="newsletter__inner">
          <h2>{{ t('blog_extra.cta_questions') }}</h2>
          <p>{{ t('blog_extra.subscribe_desc') }}</p>
          <NuxtLink :to="localePath('/contact')" class="btn btn-primary">
            {{ t('form.submit') }}
          </NuxtLink>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n, useLocalePath } from '#i18n'
import { useTranslation } from '~/composables/useTranslation'

const route = useRoute()
const { t } = useI18n()
const localePath = useLocalePath()
const { tField } = useTranslation()
const { getPosts } = useApi()

const activeCategory = ref(typeof route.query.category === 'string' ? route.query.category : '')
const currentPage = ref(1)
const limit = 9

// The backend clamps the page size to 50 (pagination.ParsePagination max), so
// requesting limit: 200 silently truncated the category list. Page through the
// full set at the real page size to surface the actual total.
const ALL_POSTS_PAGE_SIZE = 50

// Fetch all posts once for building the category filter list
const { data: allPostsData } = await useAsyncData('blog-categories', async () => {
  const first = await getPosts({ page: 1, limit: ALL_POSTS_PAGE_SIZE })
  const totalPages = first.pagination?.totalPages ?? 1
  if (totalPages <= 1) return first

  const rest = await Promise.all(
    Array.from({ length: totalPages - 1 }, (_, index) =>
      getPosts({ page: index + 2, limit: ALL_POSTS_PAGE_SIZE })
    )
  )
  return {
    data: first.data.concat(...rest.map((response) => response.data)),
    pagination: first.pagination
  }
})

const allPosts = computed(() => {
  const posts = (allPostsData.value?.data || []) as any[]
  return [...posts].sort((a, b) => {
    return new Date(b.publishedAt).getTime() - new Date(a.publishedAt).getTime()
  })
})

// Server-side paginated posts for the grid
const { data: postsResponse, pending, error: postsError, refresh: refreshPosts } = await useAsyncData(
  () => `blog-posts-${activeCategory.value}-${currentPage.value}`,
  async () => {
    const params: any = { page: currentPage.value, limit }
    if (activeCategory.value) params.category = activeCategory.value
    return await getPosts(params)
  },
  { watch: [activeCategory, currentPage] }
)

const mapCategoryName = (categoryId: string) => {
  const mapping: Record<string, string> = {
    compliance: t('blog.categories.compliance'),
    product_knowledge: t('blog.categories.product_knowledge'),
    packaging: t('blog.categories.packaging'),
    market_insights: t('blog.categories.market_insights')
  }
  return mapping[categoryId] || categoryId
}

const categories = computed(() => {
  const unique = Array.from(new Set(allPosts.value.map((post) => post.category).filter(Boolean)))
  return unique.map((id) => ({ id, name: mapCategoryName(id) }))
})

const featuredPost = computed(() => {
  if (activeCategory.value || currentPage.value !== 1 || allPosts.value.length === 0) {
    return null
  }
  return allPosts.value[0]
})

const pagedPosts = computed(() => {
  const posts = (postsResponse.value?.data || []) as any[]
  // Filter out featured post from first page results when no category filter
  if (!activeCategory.value && currentPage.value === 1 && featuredPost.value && posts.length > 0) {
    return posts.filter(p => p.slug !== featuredPost.value.slug)
  }
  return posts
})

const totalPages = computed(() => postsResponse.value?.pagination?.totalPages || 0)

const filteredPosts = computed(() => (postsResponse.value?.data || []) as any[])

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

const getCategoryName = (categoryId: string) => {
  return mapCategoryName(categoryId)
}

const { formatDate } = useDisplay()
const fmtDate = (dateString: string) => formatDate(dateString, { month: 'short', day: 'numeric', year: 'numeric' })

const setCategory = (categoryId: string) => {
  activeCategory.value = categoryId
  navigateTo({ query: categoryId ? { category: categoryId } : {} })
}

const goToPage = (page: number) => {
  if (page < 1 || page > totalPages.value) return
  currentPage.value = page
}

watch(activeCategory, () => {
  currentPage.value = 1
})

// SEO
usePageOgImage({
  title: `${t('blog.title')} | ${t('seo.default_title')}`,
  description: t('blog_extra.seo_description'),
  ogType: 'website'
})
</script>

<style scoped>
.blog-hero__inner {
  text-align: center;
  max-width: 700px;
  margin: 0 auto;
}

.blog-hero__title {
  font-size: clamp(2rem, 4vw, 3rem);
  margin-bottom: var(--spacing-md);
}

.blog-hero__subtitle {
  font-size: var(--text-lg);
  color: var(--color-text-light);
}

/* Categories */
.categories__inner {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

.category-tab {
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-sm);
  font-weight: 500;
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  cursor: pointer;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.category-tab:hover {
  border-color: var(--color-accent);
}

.category-tab--active {
  background-color: var(--color-primary);
  border-color: var(--color-primary);
  color: var(--color-text-on-primary);
}

/* Featured Post */
.featured-post__inner {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-2xl);
  text-decoration: none;
  color: inherit;
}

@media (min-width: 1024px) {
  .featured-post__inner {
    grid-template-columns: 1.5fr 1fr;
  }
}

.featured-post__image {
  position: relative;
  aspect-ratio: 16/9;
  border-radius: var(--radius-xl);
  overflow: hidden;
}

.featured-post__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-slow);
}

.featured-post__inner:hover .featured-post__image img {
  transform: scale(1.03);
}

.featured-post__badge {
  position: absolute;
  top: var(--spacing-md);
  left: var(--spacing-md);
  padding: var(--spacing-xs) var(--spacing-md);
  background-color: var(--color-highlight);
  color: var(--color-text-on-primary);
  font-size: var(--text-xs);
  font-weight: 600;
  border-radius: var(--radius-sm);
}

.featured-post__content {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.featured-post__category {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-xs);
  font-weight: 600;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-sm);
  color: var(--color-accent);
  margin-bottom: var(--spacing-md);
}

.featured-post__title {
  font-size: var(--text-3xl);
  margin-bottom: var(--spacing-md);
}

.featured-post__excerpt {
  font-size: var(--text-lg);
  color: var(--color-text-light);
  line-height: 1.6;
  margin-bottom: var(--spacing-xl);
}

.featured-post__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-lg);
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.featured-post__meta span {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

/* Posts Grid */
.posts__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--spacing-xl);
  margin-bottom: var(--spacing-3xl);
}

.post-card__link {
  display: block;
  text-decoration: none;
  color: inherit;
}

.post-card__image {
  position: relative;
  aspect-ratio: 16/10;
  border-radius: var(--radius-lg);
  overflow: hidden;
  margin-bottom: var(--spacing-md);
  background-color: var(--color-bg-alt);
}

.post-card__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-base);
}

.post-card:hover .post-card__image img {
  transform: scale(1.05);
}

.post-card__category {
  position: absolute;
  top: var(--spacing-sm);
  left: var(--spacing-sm);
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-xs);
  font-weight: 600;
  background-color: var(--color-bg);
  border-radius: var(--radius-sm);
  color: var(--color-primary);
}

.post-card__title {
  font-size: var(--text-xl);
  margin-bottom: var(--spacing-sm);
}

.post-card__excerpt {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  line-height: 1.6;
  margin-bottom: var(--spacing-md);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
}

.post-card__meta {
  display: flex;
  justify-content: space-between;
  font-size: var(--text-xs);
  color: var(--color-text-light);
}

/* Pagination */
.posts__pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-md);
}

.pagination__btn {
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
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
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
}

.pagination__page--active {
  background-color: var(--color-primary);
  border-color: var(--color-primary);
  color: var(--color-text-on-primary);
}

/* Contact CTA */
.newsletter__inner {
  text-align: center;
}

.newsletter__inner h2 {
  margin-bottom: var(--spacing-md);
}

.newsletter__inner p {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-xl);
}

.posts__empty {
  text-align: center;
  padding: var(--spacing-5xl) var(--spacing-lg);
  color: var(--color-text-light);
}
</style>
