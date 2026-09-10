<template>
  <div class="cases-page">
    <!-- Breadcrumb -->
    <div class="container">
      <Breadcrumb :items="[{ label: t('nav.cases') }]" />
    </div>

    <!-- Hero -->
    <section class="cases-hero section">
      <div class="container">
        <div class="cases-hero__inner">
          <h1 class="cases-hero__title">{{ t('cases.title') }}</h1>
          <p class="cases-hero__subtitle">{{ t('cases.subtitle') }}</p>
        </div>
      </div>
    </section>

    <!-- Filters -->
    <section v-if="industries.length" class="filters section-sm">
      <div class="container">
        <div class="filters__inner">
          <button
            class="filter-tab"
            :class="{ 'filter-tab--active': !activeIndustry }"
            @click="activeIndustry = ''"
          >
            {{ t('cases.all_industries') }}
          </button>
          <button
            v-for="industry in industries"
            :key="industry.id"
            class="filter-tab"
            :class="{ 'filter-tab--active': activeIndustry === industry.id }"
            @click="activeIndustry = industry.id"
          >
            {{ industry.name }}
          </button>
        </div>
      </div>
    </section>

    <!-- Cases Grid -->
    <section class="cases section">
      <div class="container">
        <div v-if="filteredCases.length === 0" class="cases__empty">
          <p>{{ t('cases.no_results') }}</p>
        </div>

        <div v-else class="cases__grid">
          <NuxtLink
            v-for="item in filteredCases"
            :key="item.id"
            :to="localePath(`/cases-clients/${item.slug}`)"
            class="case-card"
          >
            <div class="case-card__image">
              <img :src="item.thumbnail" :alt="tField(item, 'title')" />
              <div class="case-card__overlay">
                <span class="case-card__view">{{ t('cases.view_case') }}</span>
              </div>
            </div>
            <div class="case-card__content">
              <span class="case-card__industry">{{ tField(item, 'industry') }}</span>
              <h3 class="case-card__title">{{ tField(item, 'client') }}</h3>
              <p class="case-card__project">{{ tField(item, 'title') }}</p>
              <div class="case-card__result">
                <Icon name="lucide:trending-up" size="16" />
                <span>{{ tField(item, 'result') }}</span>
              </div>
            </div>
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="cta section bg-alt">
      <div class="container">
        <div class="cta__inner">
          <h2>{{ t('cases_detail.cta_title') }}</h2>
          <p>{{ t('cases_detail.cta_subtitle') }}</p>
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

const { t } = useI18n()
const localePath = useLocalePath()
const { tField } = useTranslation()
const config = useRuntimeConfig()
const { getCases } = useApi()

const activeIndustry = ref('')

// The backend clamps the page size to 50 (pagination.ParsePagination max), so
// requesting limit: 200 silently truncated the catalog. Page through the full
// set at the real page size so the grid and the industry filter see every case.
const CASES_PAGE_SIZE = 50

const { data: casesResponse } = await useAsyncData('cases-all', async () => {
  const first = await getCases({ page: 1, limit: CASES_PAGE_SIZE })
  const totalPages = first.pagination?.totalPages ?? 1
  if (totalPages <= 1) return first

  const rest = await Promise.all(
    Array.from({ length: totalPages - 1 }, (_, index) =>
      getCases({ page: index + 2, limit: CASES_PAGE_SIZE })
    )
  )
  return {
    data: first.data.concat(...rest.map((response) => response.data)),
    pagination: first.pagination
  }
})

const caseStudies = computed(() => {
  const items = (casesResponse.value?.data || []) as any[]
  return items.map((item) => ({
    ...item,
    project: item.title || item.project || item.slug
  }))
})

const industries = computed(() => {
  const names = Array.from(new Set(caseStudies.value.map((item) => item.industry).filter(Boolean)))
  return names.map((name) => ({
    id: String(name),
    name: String(name)
  }))
})

const filteredCases = computed(() => {
  if (!activeIndustry.value) return caseStudies.value
  return caseStudies.value.filter((c) => c.industry === activeIndustry.value)
})

const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  if (!number) return ''
  const message = encodeURIComponent('Hi, I\'m interested in learning about your case studies and success stories.')
  return `https://wa.me/${number}?text=${message}`
})

usePageOgImage({
  title: `${t('nav.cases')} | ${t('seo.default_title')}`,
  description: t('cases.subtitle'),
  ogType: 'website'
})
</script>

<style scoped>
.cases-hero__inner {
  text-align: center;
  max-width: 700px;
  margin: 0 auto;
}

.cases-hero__title {
  font-size: clamp(2rem, 4vw, 3rem);
  margin-bottom: var(--spacing-md);
}

.cases-hero__subtitle {
  font-size: var(--text-xl);
  color: var(--color-text-light);
}

/* Filters */
.filters__inner {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

.filter-tab {
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--text-sm);
  font-weight: 500;
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  cursor: pointer;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.filter-tab:hover {
  border-color: var(--color-accent);
}

.filter-tab--active {
  background-color: var(--color-primary);
  border-color: var(--color-primary);
  color: var(--color-text-on-primary);
}

/* Cases Grid */
.cases__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: var(--spacing-xl);
}

.case-card {
  display: block;
  background-color: var(--color-bg);
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: var(--shadow-md);
  text-decoration: none;
  color: inherit;
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.case-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-xl);
}

.case-card__image {
  position: relative;
  aspect-ratio: 16/10;
  overflow: hidden;
}

.case-card__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-slow);
}

.case-card:hover .case-card__image img {
  transform: scale(1.05);
}

.case-card__overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(0,0,0,0.7) 0%, transparent 100%);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding: var(--spacing-lg);
  opacity: 0;
  transition: opacity var(--transition-base);
}

.case-card:hover .case-card__overlay {
  opacity: 1;
}

.case-card__view {
  color: white;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  transform: translateY(20px);
  transition: transform var(--transition-base);
}

.case-card:hover .case-card__view {
  transform: translateY(0);
}

.case-card__content {
  padding: var(--spacing-lg);
}

.case-card__industry {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-xs);
  font-weight: 600;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-sm);
  color: var(--color-accent);
  margin-bottom: var(--spacing-sm);
}

.case-card__title {
  font-size: var(--text-lg);
  font-weight: 600;
  margin-bottom: var(--spacing-xs);
}

.case-card__project {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin-bottom: var(--spacing-md);
}

.case-card__result {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs) var(--spacing-sm);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-success);
  background-color: rgba(var(--color-success-rgb), 0.15);
  border-radius: var(--radius-sm);
}

.cases__empty {
  text-align: center;
  padding: var(--spacing-5xl) var(--spacing-lg);
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

