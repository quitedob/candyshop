<template>
  <div class="blog-post-page">
    <!-- Breadcrumb -->
    <div class="container">
      <Breadcrumb :items="breadcrumbItems" />
    </div>

    <!-- Article -->
    <article class="article section">
      <div class="container">
        <div class="article__inner">
          <!-- Header -->
          <header class="article__header">
            <span class="article__category">{{ getCategoryName(post.category) }}</span>
            <h1 class="article__title">{{ tField(post, 'title') }}</h1>
            <p class="article__excerpt">{{ tField(post, 'excerpt') }}</p>

            <div class="article__meta">
              <div class="article__author">
                <div class="article__author-avatar">
                  <img v-if="translatedAuthor.avatar" :src="translatedAuthor.avatar" :alt="translatedAuthor.name" />
                  <Icon v-else name="lucide:user" size="24" />
                </div>
                <div>
                  <span class="article__author-name">{{ translatedAuthor.name }}</span>
                  <span v-if="translatedAuthor.title" class="article__author-title">{{ translatedAuthor.title }}</span>
                </div>
              </div>
              <div class="article__dates">
                <span class="article__published">
                  <Icon name="lucide:calendar" size="14" />
                  {{ fmtDate(post.publishedAt) }}
                </span>
                <span class="article__read-time">
                  <Icon name="lucide:clock" size="14" />
                  {{ t('blog_extra.min_read', { count: post.readTime }) }}
                </span>
              </div>
            </div>
          </header>

          <!-- Cover Image -->
          <div v-if="post.thumbnail" class="article__cover">
            <img :src="post.thumbnail" :alt="tField(post, 'title')" />
          </div>

          <!-- Content -->
          <div class="article__content">
            <div class="prose" v-html="renderedContent"></div>
          </div>

          <!-- Tags -->
          <div v-if="post.tags" class="article__tags">
            <span v-for="tag in post.tags" :key="tag" class="article__tag">
              #{{ tag }}
            </span>
          </div>
        </div>
      </div>
    </article>

    <!-- Share -->
    <section class="share section">
      <div class="container container-narrow">
        <div class="share__inner">
          <h3>{{ t('blog.share') }}</h3>
          <div class="share__buttons">
            <button
              @click="shareOnSocial('twitter', tField(post, 'title'), fullUrl)"
              class="share__button share__button--twitter"
            >
              <Icon name="lucide:twitter" size="20" />
              {{ t('blog_extra.share_twitter') }}
            </button>
            <button
              @click="shareOnSocial('linkedin', tField(post, 'title'), fullUrl)"
              class="share__button share__button--linkedin"
            >
              <Icon name="lucide:linkedin" size="20" />
              {{ t('blog_extra.share_linkedin') }}
            </button>
            <button
              @click="shareOnSocial('facebook', tField(post, 'title'), fullUrl)"
              class="share__button share__button--facebook"
            >
              <Icon name="lucide:facebook" size="20" />
              {{ t('blog_extra.share_facebook') }}
            </button>
            <button
              @click="copyLink"
              class="share__button share__button--copy"
            >
              <Icon name="lucide:link" size="20" />
              {{ copied ? t('blog_extra.share_copied') : t('blog_extra.share_copy_link') }}
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- Author Bio -->
    <section v-if="translatedAuthor.bio" class="author-bio section bg-alt">
      <div class="container container-narrow">
        <div class="author-bio__inner">
          <div class="author-bio__avatar">
            <img v-if="translatedAuthor.avatar" :src="translatedAuthor.avatar" :alt="translatedAuthor.name" />
            <Icon v-else name="lucide:user" size="48" />
          </div>
          <div class="author-bio__content">
            <h4>{{ translatedAuthor.name }}</h4>
            <p class="author-bio__title">{{ translatedAuthor.title }}</p>
            <p class="author-bio__bio">{{ translatedAuthor.bio }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Related Posts -->
    <section v-if="relatedPosts.length" class="related section">
      <div class="container">
        <div class="section-header">
          <h2>{{ t('blog.related_articles') }}</h2>
        </div>

        <div class="related__grid">
          <NuxtLink
            v-for="related in relatedPosts"
            :key="related.id"
            :to="localePath(`/blog/${related.slug}`)"
            class="related-post"
          >
            <img :src="related.thumbnail || '/images/blog-placeholder.jpg'" :alt="tField(related, 'title')" />
            <div class="related-post__content">
              <span class="related-post__category">{{ getCategoryName(related.category) }}</span>
              <h4>{{ tField(related, 'title') }}</h4>
            </div>
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="cta section bg-alt">
      <div class="container">
        <div class="cta__inner">
          <h2>{{ t('blog_extra.cta_heading') }}</h2>
          <p>{{ t('blog_extra.cta_questions') }}</p>
          <div class="cta__actions">
            <a v-if="whatsappUrl" :href="whatsappUrl" target="_blank" rel="noopener noreferrer" class="btn btn-highlight">
              <WhatsAppIcon size="20" />
              {{ t('whatsapp.us') }}
            </a>
            <NuxtLink :to="localePath('/contact')" class="btn btn-outline">
              {{ t('form.submit') }}
            </NuxtLink>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n, useLocalePath } from '#i18n'
import { useTranslation } from '~/composables/useTranslation'

const { t, locale } = useI18n()
const localePath = useLocalePath()
const { tField } = useTranslation()
const route = useRoute()
const config = useRuntimeConfig()
const { getPost, getRelatedPosts } = useApi()
const { sanitize } = useSanitizer()

const copied = ref(false)
const slug = computed(() => route.params.slug as string)

const { data: postData, error: postError } = await useAsyncData(
  () => `blog-post-${slug.value}`,
  async () => await getPost(slug.value),
  { watch: [slug] }
)

if (postError.value) {
  throw createError({ statusCode: 404, statusMessage: 'Post Not Found' })
}

const { data: relatedData } = await useAsyncData(
  () => `blog-related-${slug.value}`,
  async () => await getRelatedPosts(slug.value, 3),
  { watch: [slug] }
)

const renderedContent = computed(() => {
  const raw = tField(post.value, 'content') || ''
  if (!raw) return ''
  return sanitize(raw)
})

const post = computed(() => {
  if (!postData.value) {
    return {
      id: '',
      slug: slug.value,
      title: '',
      excerpt: '',
      category: '',
      author: { name: 'CandyPro' },
      publishedAt: new Date().toISOString(),
      readTime: 0,
      thumbnail: '',
      tags: [],
      content: ''
    }
  }
  return postData.value
})

const translatedAuthor = computed(() => {
  const p = post.value as any
  const author = p?.author || {}
  const tr = p?.translations?.[locale.value] || {}
  return {
    name: tr.authorName || author.name || 'CandyPro',
    title: tr.authorTitle || author.title || '',
    bio: tr.authorBio || author.bio || '',
    avatar: author.avatar || ''
  }
})

const relatedPosts = computed(() => (relatedData.value || []) as any[])

const categories = computed(() => {
  const unique = Array.from(
    new Set([post.value.category, ...relatedPosts.value.map((p: any) => p.category)].filter(Boolean))
  )
  return unique.map((id) => ({ id, name: getCategoryName(id as string) }))
})

const breadcrumbItems = computed(() => [
  { label: t('blog.title'), to: '/blog' },
  { label: getCategoryName(post.value.category), to: `/blog?category=${post.value.category}` },
  { label: tField(post.value, 'title') }
])

const getCategoryName = (categoryId: string) => {
  const mapping: Record<string, string> = {
    compliance: t('blog.categories.compliance'),
    product_knowledge: t('blog.categories.product_knowledge'),
    packaging: t('blog.categories.packaging'),
    market_insights: t('blog.categories.market_insights')
  }
  return mapping[categoryId] || categoryId
}

const { formatDate } = useDisplay()
const fmtDate = (dateString: string) => formatDate(dateString, { month: 'long', day: 'numeric', year: 'numeric' })

const fullUrl = computed(() => `${config.public.siteUrl}/blog/${slug.value}`)

const shareOnSocial = (platform: string, title: string, url: string) => {
  const encodedTitle = encodeURIComponent(title)
  const encodedUrl = encodeURIComponent(url)

  const urls: Record<string, string> = {
    twitter: `https://twitter.com/intent/tweet?text=${encodedTitle}&url=${encodedUrl}`,
    linkedin: `https://www.linkedin.com/sharing/share-offsite/?url=${encodedUrl}`,
    facebook: `https://www.facebook.com/sharer/sharer.php?u=${encodedUrl}`
  }

  window.open(urls[platform], '_blank', 'noopener,noreferrer,width=600,height=400')
}

const copyLink = () => {
  navigator.clipboard.writeText(fullUrl.value)
  copied.value = true
  setTimeout(() => {
    copied.value = false
  }, 2000)
}

const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  if (!number) return ''
  const message = encodeURIComponent(`Hi, I have a question about your blog post: "${tField(post.value, 'title')}"`)
  return `https://wa.me/${number}?text=${message}`
})

usePageOgImage({
  title: tField(post.value, 'title'),
  description: tField(post.value, 'excerpt'),
  ogImage: (post.value as { ogImage?: string }).ogImage,
  thumbnail: post.value.thumbnail,
  ogType: 'article',
  schema: {
    '@context': 'https://schema.org',
    '@type': 'Article',
    headline: tField(post.value, 'title'),
    description: tField(post.value, 'excerpt'),
    image: post.value.thumbnail,
    author: {
      '@type': 'Person',
      name: translatedAuthor.value.name
    },
    publisher: {
      '@type': 'Organization',
      name: 'CandyPro OEM'
    },
    datePublished: post.value.publishedAt
  }
})
</script>

<style scoped>
.article__inner {
  max-width: 800px;
  margin: 0 auto;
}

.article__header {
  margin-bottom: var(--spacing-2xl);
}

.article__category {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-md);
  font-size: var(--text-sm);
  font-weight: 600;
  background-color: var(--color-accent);
  color: white;
  border-radius: var(--radius-full);
  margin-bottom: var(--spacing-md);
}

.article__title {
  font-size: clamp(1.75rem, 3vw, 2.5rem);
  margin-bottom: var(--spacing-md);
}

.article__excerpt {
  font-size: var(--text-xl);
  color: var(--color-text-light);
  line-height: 1.6;
  margin-bottom: var(--spacing-xl);
}

.article__meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--color-border-light);
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.article__author {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.article__author-avatar {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-full);
  overflow: hidden;
  background-color: var(--color-bg-alt);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-light);
}

.article__author-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.article__author-name {
  display: block;
  font-weight: 600;
  color: var(--color-primary);
}

.article__author-title {
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.article__dates {
  display: flex;
  gap: var(--spacing-lg);
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.article__dates span {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.article__cover {
  margin-bottom: var(--spacing-2xl);
  border-radius: var(--radius-xl);
  overflow: hidden;
}

.article__cover img {
  width: 100%;
  height: auto;
}

.article__content {
  font-size: var(--text-lg);
  line-height: 1.8;
  color: var(--color-text);
}

.article__content :deep(h2) {
  margin-top: var(--spacing-3xl);
  margin-bottom: var(--spacing-lg);
}

.article__content :deep(h3) {
  margin-top: var(--spacing-2xl);
  margin-bottom: var(--spacing-md);
}

.article__content :deep(p) {
  margin-bottom: var(--spacing-lg);
}

.article__content :deep(ul),
.article__content :deep(ol) {
  margin-bottom: var(--spacing-lg);
  padding-left: var(--spacing-lg);
}

.article__content :deep(li) {
  margin-bottom: var(--spacing-sm);
}

.article__tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-3xl);
  padding-top: var(--spacing-xl);
  border-top: 1px solid var(--color-border-light);
}

.article__tag {
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-sm);
  color: var(--color-accent);
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-sm);
}

/* Share */
.share__inner {
  text-align: center;
}

.share__inner h3 {
  margin-bottom: var(--spacing-md);
}

.share__buttons {
  display: flex;
  justify-content: center;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.share__button {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-sm);
  font-weight: 500;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.share__button:hover {
  border-color: var(--color-accent);
}

.share__button--twitter {
  color: #1DA1F2;
}

.share__button--linkedin {
  color: #0A66C2;
}

.share__button--facebook {
  color: #1877F2;
}

/* Author Bio */
.author-bio__inner {
  display: flex;
  gap: var(--spacing-lg);
  padding: var(--spacing-xl);
  background-color: white;
  border-radius: var(--radius-lg);
}

.author-bio__avatar {
  flex-shrink: 0;
  width: 80px;
  height: 80px;
  border-radius: var(--radius-full);
  overflow: hidden;
  background-color: var(--color-bg-alt);
  display: flex;
  align-items: center;
  justify-content: center;
}

.author-bio__avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.author-bio__content h4 {
  font-size: var(--text-lg);
  margin-bottom: var(--spacing-xs);
}

.author-bio__title {
  font-size: var(--text-sm);
  color: var(--color-accent);
  margin-bottom: var(--spacing-sm);
}

.author-bio__bio {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  line-height: 1.6;
  margin: 0;
}

/* Related Posts */
.related__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: var(--spacing-xl);
}

.related-post {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  text-decoration: none;
  color: inherit;
}

.related-post img {
  width: 100%;
  aspect-ratio: 16/10;
  object-fit: cover;
  border-radius: var(--radius-lg);
}

.related-post__category {
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-accent);
}

.related-post h4 {
  font-size: var(--text-base);
  margin-top: var(--spacing-xs);
}

/* CTA */
.cta__inner {
  text-align: center;
  max-width: 700px;
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

