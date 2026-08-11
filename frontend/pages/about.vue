<template>
  <div class="about-page">
    <!-- Breadcrumb -->
    <div class="container">
      <Breadcrumb :items="[{ label: t('nav.about') }]" />
    </div>

    <!-- Hero -->
    <section class="about-hero section">
      <div class="container">
        <div class="about-hero__inner">
          <h1 class="about-hero__title">{{ t('nav.about') }}</h1>
          <p class="about-hero__subtitle">{{ t('about.story') }}</p>
        </div>
      </div>
    </section>

    <!-- Story -->
    <section class="story section section-lg">
      <div class="container">
        <div class="story__inner">
          <div class="story__content">
            <h2>{{ t('about.story') }}</h2>
            <p>{{ t('about.story_p1') }}</p>
            <p>{{ t('about.story_p2') }}</p>
            <p>{{ t('about.story_p3') }}</p>
          </div>
          <div class="story__image">
            <img src="/images/about-story.jpg" :alt="$t('about.alt_our_factory')" />
          </div>
        </div>
      </div>
    </section>

    <!-- Mission, Vision, Values -->
    <section class="mvv section bg-alt">
      <div class="container">
        <div class="mvv__grid">
          <div class="mvv-card">
            <Icon name="lucide:target" size="32" />
            <h3>{{ t('about.mission') }}</h3>
            <p>{{ t('about.mission_desc') }}</p>
          </div>
          <div class="mvv-card">
            <Icon name="lucide:eye" size="32" />
            <h3>{{ t('about.vision') }}</h3>
            <p>{{ t('about.vision_desc') }}</p>
          </div>
          <div class="mvv-card">
            <Icon name="lucide:heart" size="32" />
            <h3>{{ t('about.values') }}</h3>
            <p>{{ t('about.values_desc') }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Stats -->
    <section class="stats section">
      <div class="container">
        <FactoryStats :stats="factoryStats" variant="default" />
      </div>
    </section>

    <!-- Team -->
    <section class="team section section-lg">
      <div class="container">
        <div class="section-header">
          <h2>{{ t('about.team') }}</h2>
          <p>{{ t('about.team_subtitle') }}</p>
        </div>

        <div class="team__grid">
          <div v-for="member in team" :key="member.id" class="team-member">
            <div class="team-member__image">
              <img :src="member.avatar" :alt="member.name" />
            </div>
            <h4 class="team-member__name">{{ member.name }}</h4>
            <p class="team-member__title">{{ member.title }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Certifications -->
    <section class="certifications section bg-alt">
      <div class="container">
        <div class="section-header">
          <h2>{{ t('factory.certifications') }}</h2>
        </div>

        <CertificationList :certifications="certifications" variant="grid" />
      </div>
    </section>

    <!-- CTA -->
    <section class="cta section">
      <div class="container">
        <div class="cta__inner">
          <h2>{{ t('about.cta_title') }}</h2>
          <p>{{ t('about.cta_subtitle') }}</p>
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
import { useI18n, useLocalePath } from '#i18n'

const { t } = useI18n()
const localePath = useLocalePath()
const config = useRuntimeConfig()

// Factory stats
const factoryStats = [
  { id: 1, value: '15+', label: t('factory.years'), icon: 'lucide:calendar' },
  { id: 2, value: '100T', label: t('factory.daily_output'), icon: 'lucide:package' },
  { id: 3, value: '50+', label: t('factory.countries'), icon: 'lucide:globe' },
  { id: 4, value: '150+', label: t('about.team_members'), icon: 'lucide:users' }
]

// Team
const team = [
  { id: 1, name: 'James Wilson', title: t('about.titles.ceo'), avatar: '/images/team/james.jpg' },
  { id: 2, name: 'Sarah Chen', title: t('about.titles.qa_director'), avatar: '/images/team/sarah.jpg' },
  { id: 3, name: 'Michael Zhang', title: t('about.titles.rd_manager'), avatar: '/images/team/michael.jpg' },
  { id: 4, name: 'Emily Davis', title: t('about.titles.sales_director'), avatar: '/images/team/emily.jpg' },
  { id: 5, name: 'David Park', title: t('about.titles.production_manager'), avatar: '/images/team/david.jpg' },
  { id: 6, name: 'Anna Wang', title: t('about.titles.export_manager'), avatar: '/images/team/anna.jpg' }
]

// Certifications
const certifications = [
  {
    id: '1',
    name: t('factory.haccp'),
    abbreviation: 'HACCP',
    description: t('about.cert_desc.haccp'),
    issuer: 'SGS',
    validUntil: '2027',
    badgeUrl: '/certificates/haccp-badge.svg',
    certificateUrl: '/certificates/haccp.pdf'
  },
  {
    id: '2',
    name: t('factory.iso'),
    abbreviation: 'ISO',
    description: t('about.cert_desc.iso'),
    issuer: 'BSI',
    validUntil: '2027',
    badgeUrl: '/certificates/iso-badge.svg',
    certificateUrl: '/certificates/iso22000.pdf'
  },
  {
    id: '3',
    name: t('factory.brc'),
    abbreviation: 'BRC',
    description: t('about.cert_desc.brc'),
    issuer: 'BRC',
    validUntil: '2026',
    badgeUrl: '/certificates/brc-badge.svg',
    certificateUrl: '/certificates/brc.pdf'
  },
  {
    id: '4',
    name: t('factory.halal'),
    abbreviation: 'HALAL',
    description: t('about.cert_desc.halal'),
    issuer: 'IFRC',
    validUntil: '2027',
    badgeUrl: '/certificates/halal-badge.svg',
    certificateUrl: '/certificates/halal.pdf'
  }
]

// WhatsApp URL
const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  if (!number) return ''
  const message = encodeURIComponent(t('whatsapp.message'))
  return `https://wa.me/${number}?text=${message}`
})

// SEO
usePageOgImage({
  title: `${t('nav.about')} | ${t('seo.default_title')}`,
  description: t('seo.about_description'),
  ogType: 'website'
})
</script>

<style scoped>
.about-hero__inner {
  text-align: center;
  max-width: 700px;
  margin: 0 auto;
}

.about-hero__title {
  font-size: clamp(2rem, 4vw, 3rem);
  margin-bottom: var(--spacing-md);
}

.about-hero__subtitle {
  font-size: var(--text-xl);
  color: var(--color-text-light);
}

/* Story */
.story__inner {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-3xl);
}

@media (min-width: 1024px) {
  .story__inner {
    grid-template-columns: 1fr 1fr;
    align-items: center;
  }
}

.story__content h2 {
  font-size: var(--text-3xl);
  margin-bottom: var(--spacing-xl);
}

.story__content p {
  font-size: var(--text-lg);
  line-height: 1.8;
  color: var(--color-text);
  margin-bottom: var(--spacing-lg);
}

.story__image img {
  width: 100%;
  border-radius: var(--radius-2xl);
  box-shadow: var(--shadow-xl);
}

/* Mission, Vision, Values */
.mvv__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-xl);
}

.mvv-card {
  text-align: center;
  padding: var(--spacing-2xl);
  background-color: white;
  border-radius: var(--radius-xl);
}

.mvv-card svg {
  color: var(--color-accent);
  margin-bottom: var(--spacing-md);
}

.mvv-card h3 {
  font-size: var(--text-xl);
  margin-bottom: var(--spacing-md);
}

.mvv-card p {
  color: var(--color-text-light);
  line-height: 1.6;
}

/* Team */
.team__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-xl);
}

.team-member {
  text-align: center;
}

.team-member__image {
  width: 150px;
  height: 150px;
  margin: 0 auto var(--spacing-md);
  border-radius: var(--radius-full);
  overflow: hidden;
  background-color: var(--color-bg-alt);
}

.team-member__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.team-member__name {
  font-size: var(--text-lg);
  font-weight: 600;
  margin-bottom: var(--spacing-xs);
}

.team-member__title {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin: 0;
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
