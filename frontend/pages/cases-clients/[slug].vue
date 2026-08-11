<template>
  <div class="case-page">
    <div class="container">
      <Breadcrumb :items="breadcrumbItems" />
    </div>

    <section v-if="pending" class="section">
      <div class="container container-narrow">
        <p class="case-status">{{ t('cases_detail.loading') }}</p>
      </div>
    </section>

    <section v-else-if="!caseStudy" class="section">
      <div class="container container-narrow">
        <div class="case-status case-status--error">
          <h2>{{ t('cases_detail.not_found') }}</h2>
          <p>{{ loadError || t('cases_detail.not_found_desc') }}</p>
          <NuxtLink :to="localePath('/cases-clients')" class="btn btn-primary">
            {{ t('cases_detail.back') }}
          </NuxtLink>
        </div>
      </div>
    </section>

    <template v-else>
      <section class="case-hero section">
        <div class="container">
          <div class="case-hero__inner">
            <div class="case-hero__content">
              <span class="case-hero__industry">{{ tField(caseStudy, 'industry') }}</span>
              <h1 class="case-hero__title">{{ tField(caseStudy, 'client') }}</h1>
              <p class="case-hero__project">{{ tField(caseStudy, 'title') }}</p>
              <div class="case-hero__meta">
                <div v-if="caseStudy.location" class="case-hero__meta-item">
                  <Icon name="lucide:map-pin" size="18" />
                  <span>{{ tField(caseStudy, 'location') }}</span>
                </div>
                <div v-if="caseStudy.year" class="case-hero__meta-item">
                  <Icon name="lucide:calendar" size="18" />
                  <span>{{ caseStudy.year }}</span>
                </div>
              </div>
            </div>
            <div v-if="caseStudy.thumbnail" class="case-hero__image">
              <img :src="caseStudy.thumbnail" :alt="tField(caseStudy, 'client')" />
            </div>
          </div>
        </div>
      </section>

      <section v-if="caseStudy.result" class="result-banner section-sm bg-alt">
        <div class="container">
          <div class="result-banner__inner">
            <h3>{{ t('cases_detail.result') }}</h3>
            <p class="result-banner__text">{{ tField(caseStudy, 'result') }}</p>
          </div>
        </div>
      </section>

      <section v-if="caseStudy.challenge" class="challenge section">
        <div class="container container-narrow">
          <div class="case-section">
            <div class="case-section__header">
              <Icon name="lucide:alert-triangle" size="32" />
              <h2>{{ t('cases_detail.challenge') }}</h2>
            </div>
            <div class="case-section__content">{{ tField(caseStudy, 'challenge') }}</div>
          </div>
        </div>
      </section>

      <section v-if="caseStudy.solution" class="solution section bg-alt">
        <div class="container container-narrow">
          <div class="case-section">
            <div class="case-section__header">
              <Icon name="lucide:lightbulb" size="32" />
              <h2>{{ t('cases_detail.solution') }}</h2>
            </div>
            <div class="case-section__content">{{ tField(caseStudy, 'solution') }}</div>

            <div v-if="caseStudy.services.length" class="case-services">
              <h4>{{ t('cases_detail.services') }}</h4>
              <div class="case-services__list">
                <span
                  v-for="service in caseStudy.services"
                  :key="service"
                  class="case-services__item"
                >
                  <Icon name="lucide:check" size="16" />
                  {{ service }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section v-if="caseStudy.images.length" class="gallery section">
        <div class="container">
          <div class="section-header">
            <h2>{{ t('cases_detail.gallery') }}</h2>
          </div>
          <div class="gallery__grid">
            <img
              v-for="(image, index) in caseStudy.images"
              :key="`${caseStudy.id}-${index}`"
              :src="image"
              :alt="`${tField(caseStudy, 'client')} image ${index + 1}`"
              class="gallery__image"
            />
          </div>
        </div>
      </section>

      <section v-if="timelineSteps.length || caseStudy.timelineText" class="timeline section bg-alt">
        <div class="container">
          <div class="section-header">
            <h2>{{ t('cases_detail.timeline') }}</h2>
          </div>
          <ProcessFlow v-if="timelineSteps.length" :steps="timelineSteps" variant="default" />
          <p v-else class="timeline__text">{{ tField(caseStudy, 'timeline') }}</p>
        </div>
      </section>

      <section v-if="relatedCases.length" class="related section">
        <div class="container">
          <div class="section-header">
            <h2>{{ t('cases_detail.similar') }}</h2>
          </div>

          <div class="related__grid">
            <NuxtLink
              v-for="related in relatedCases"
              :key="related.id"
              :to="localePath(`/cases-clients/${related.slug}`)"
              class="related-card"
            >
              <img v-if="related.thumbnail" :src="related.thumbnail" :alt="tField(related, 'client')" />
              <div class="related-card__content">
                <span class="related-card__industry">{{ tField(related, 'industry') }}</span>
                <h4>{{ tField(related, 'client') }}</h4>
                <p>{{ tField(related, 'title') }}</p>
              </div>
            </NuxtLink>
          </div>
        </div>
      </section>

      <section class="cta section bg-alt">
        <div class="container">
          <div class="cta__inner">
            <h2>{{ t('cases_detail.cta_title') }}</h2>
            <p>{{ t('cases_detail.cta_subtitle') }}</p>
            <div class="cta__actions">
              <a v-if="whatsappUrl" :href="whatsappUrl" target="_blank" rel="noopener noreferrer" class="btn btn-highlight btn-lg">
                <WhatsAppIcon size="20" />
                {{ t('whatsapp.us') }}
              </a>
              <NuxtLink :to="localePath('/contact')" class="btn btn-outline btn-lg">
                {{ t('form.submit') }}
              </NuxtLink>
            </div>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n, useLocalePath } from '#i18n'
import { useTranslation } from '~/composables/useTranslation'

interface RawCaseStudy {
  id: string
  slug: string
  title?: string
  client?: string
  industry?: string
  location?: string
  thumbnail?: string
  images?: string[]
  challenge?: string
  solution?: string
  result?: string
  timeline?: string
  services?: string[]
  createdAt?: string
}

interface TimelineStep {
  id: string
  title: string
  description: string
}

const { t } = useI18n()
const localePath = useLocalePath()
const { tField } = useTranslation()
const route = useRoute()
const config = useRuntimeConfig()
const { getCase, getRelatedCases } = useApi()

const slug = computed(() => String(route.params.slug || ''))

const { data: caseData, pending, error } = await useAsyncData(
  () => `case-${slug.value}`,
  async () => await getCase(slug.value),
  { watch: [slug] }
)

const normalizeCase = (raw: RawCaseStudy) => {
  const createdDate = raw.createdAt ? new Date(raw.createdAt) : null
  return {
    ...raw,
    client: raw.client || t('cases_detail.fallback_client'),
    project: raw.title || raw.slug,
    industry: raw.industry || t('cases_detail.fallback_industry'),
    location: raw.location || '',
    year: createdDate ? String(createdDate.getFullYear()) : '',
    thumbnail: raw.thumbnail || '',
    images: raw.images || [],
    challenge: raw.challenge || '',
    solution: raw.solution || '',
    result: raw.result || '',
    timelineText: raw.timeline || '',
    services: raw.services || []
  }
}

const caseStudy = computed(() => {
  if (!caseData.value) {
    return null
  }
  return normalizeCase(caseData.value as RawCaseStudy)
})

const parseTimeline = (timelineText: string): TimelineStep[] => {
  const value = timelineText.trim()
  if (!value) {
    return []
  }

  const stepTitle = (index: number) => t('cases_detail.timeline_step', { n: index + 1 })

  try {
    const parsed = JSON.parse(value) as Array<{ title?: string; description?: string } | string>
    if (Array.isArray(parsed)) {
      return parsed
        .map((item, index) => {
          if (typeof item === 'string') {
            return { id: `step-${index + 1}`, title: stepTitle(index), description: item }
          }
          return {
            id: `step-${index + 1}`,
            title: item.title || stepTitle(index),
            description: item.description || ''
          }
        })
        .filter((item) => item.description.trim() !== '')
    }
  } catch {
    // timeline can be plain text instead of JSON
  }

  const parts = value
    .split(/\r?\n|;/)
    .map((part) => part.trim())
    .filter(Boolean)

  return parts.map((part, index) => ({
    id: `step-${index + 1}`,
    title: stepTitle(index),
    description: part
  }))
}

const timelineSteps = computed(() => {
  if (!caseStudy.value) {
    return []
  }
  return parseTimeline(caseStudy.value.timelineText)
})

const { data: relatedData } = await useAsyncData(
  () => `case-related-${slug.value}`,
  async () => {
    return await getRelatedCases(slug.value, 3)
  },
  { watch: [slug] }
)

const relatedCases = computed(() => {
  const rows = (relatedData.value || []) as RawCaseStudy[]
  return rows.map(normalizeCase)
})

const breadcrumbItems = computed(() => {
  const label = tField(caseStudy.value, 'client') || t('cases_detail.fallback_client')
  const items: { label: string; to?: string }[] = [{ label: t('nav.cases'), to: '/cases-clients' }]
  const industry = tField(caseStudy.value, 'industry')
  if (industry) {
    items.push({ label: industry })
  }
  items.push({ label })
  return items
})

const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  if (!number) return ''
  const client = tField(caseStudy.value, 'client') || t('cases_detail.fallback_client')
  const message = encodeURIComponent(t('cases_detail.whatsapp_similar', { client }))
  return `https://wa.me/${number}?text=${message}`
})

const loadError = computed(() => {
  const value = error.value as { message?: string } | null
  return value?.message || ''
})

usePageOgImage({
  title: computed(() => `${tField(caseStudy.value, 'client') || t('cases_detail.fallback_client')} | ${t('seo.default_title')}`),
  description: computed(() => tField(caseStudy.value, 'result') || t('cases.subtitle')),
  ogImage: computed(() => (caseStudy.value as { ogImage?: string } | null)?.ogImage),
  thumbnail: computed(() => caseStudy.value?.thumbnail),
  images: computed(() => caseStudy.value?.images),
  ogType: 'article',
})
</script>

<style scoped>
.case-status {
  padding: var(--spacing-xl);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background-color: white;
  text-align: center;
}

.case-status--error {
  display: grid;
  gap: var(--spacing-md);
  justify-items: center;
}

.case-hero__inner {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-2xl);
  align-items: center;
}

.case-hero__industry {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-sm);
  background-color: var(--color-accent-light);
  color: var(--color-accent-dark);
  border-radius: var(--radius-full);
  font-size: var(--text-sm);
  font-weight: 600;
  margin-bottom: var(--spacing-md);
}

.case-hero__title {
  font-size: clamp(2rem, 4vw, 3rem);
  margin-bottom: var(--spacing-sm);
}

.case-hero__project {
  font-size: var(--text-xl);
  color: var(--color-text-light);
  margin-bottom: var(--spacing-lg);
}

.case-hero__meta {
  display: flex;
  gap: var(--spacing-lg);
  flex-wrap: wrap;
}

.case-hero__meta-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--color-text-light);
  font-size: var(--text-sm);
}

.case-hero__image {
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: var(--shadow-xl);
}

.case-hero__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.result-banner__inner {
  text-align: center;
  max-width: 800px;
  margin: 0 auto;
}

.result-banner__text {
  font-size: var(--text-xl);
  font-weight: 500;
  color: var(--color-primary);
  margin: var(--spacing-md) 0 0;
}

.case-section__header {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
  color: var(--color-primary);
}

.case-section__header h2 {
  margin: 0;
}

.case-section__content {
  font-size: var(--text-lg);
  line-height: 1.8;
}

.case-section__content :deep(ul) {
  margin: var(--spacing-md) 0;
  padding-left: var(--spacing-xl);
}

.case-services {
  margin-top: var(--spacing-xl);
  padding-top: var(--spacing-xl);
  border-top: 1px solid var(--color-border);
}

.case-services h4 {
  margin-bottom: var(--spacing-md);
}

.case-services__list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

.case-services__item {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  font-size: var(--text-sm);
}

.gallery__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: var(--spacing-md);
}

.gallery__image {
  width: 100%;
  aspect-ratio: 4/3;
  object-fit: cover;
  border-radius: var(--radius-lg);
}

.timeline__text {
  max-width: 760px;
  margin: 0 auto;
  color: var(--color-text-light);
  text-align: center;
}

.related__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--spacing-lg);
}

.related-card {
  display: block;
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-md);
  text-decoration: none;
  color: inherit;
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.related-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-xl);
}

.related-card img {
  width: 100%;
  aspect-ratio: 16/10;
  object-fit: cover;
}

.related-card__content {
  padding: var(--spacing-md);
}

.related-card__industry {
  font-size: var(--text-xs);
  color: var(--color-accent);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.related-card h4 {
  margin: var(--spacing-xs) 0;
}

.related-card p {
  color: var(--color-text-light);
  font-size: var(--text-sm);
  margin: 0;
}

.cta__inner {
  text-align: center;
  max-width: 700px;
  margin: 0 auto;
}

.cta__actions {
  display: flex;
  justify-content: center;
  gap: var(--spacing-md);
}

.btn-lg {
  padding: var(--spacing-md) var(--spacing-xl);
}

@media (max-width: 1024px) {
  .case-hero__inner {
    grid-template-columns: 1fr;
  }

  .related__grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .related__grid {
    grid-template-columns: 1fr;
  }

  .cta__actions {
    flex-direction: column;
  }
}
</style>
