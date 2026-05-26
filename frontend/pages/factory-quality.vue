<template>
  <div class="factory-page">
    <!-- Breadcrumb -->
    <div class="container">
      <Breadcrumb :items="[{ label: t('nav.factory') }]" />
    </div>

    <!-- Hero -->
    <section class="factory-hero section">
      <div class="container">
        <div class="factory-hero__inner">
          <div class="factory-hero__content">
            <h1 class="factory-hero__title">{{ t('nav.factory') }} & {{ t('nav.about') }}</h1>
            <p class="factory-hero__subtitle">
              {{ t('factory.hero_subtitle') }}
            </p>

            <div class="factory-hero__stats">
              <div class="stat-item">
                <span class="stat-item__number">15+</span>
                <span class="stat-item__label">{{ t('factory.years') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-item__number">100T</span>
                <span class="stat-item__label">{{ t('factory.daily_output') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-item__number">50+</span>
                <span class="stat-item__label">{{ t('factory.countries') }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-item__number">150+</span>
                <span class="stat-item__label">{{ t('factory.staff') }}</span>
              </div>
            </div>
          </div>

          <div class="factory-hero__image">
            <img src="/images/factory-hero.jpg" :alt="$t('factory.alt_hero')" width="600" height="400" loading="lazy" />
          </div>

          <!-- Audio Tour Buttons -->
          <div class="factory-hero__audio mt-4 flex gap-2 flex-wrap">
            <button @click="tts.playFactoryTour()" class="px-3 py-1.5 text-sm bg-orange-100 text-orange-700 rounded-full hover:bg-orange-200 transition-colors flex items-center gap-1">
              <Icon name="heroicons:play" class="h-4 w-4" />
              {{ t('factory.audio_tour') }}
            </button>
            <button @click="tts.playQualityAssurance()" class="px-3 py-1.5 text-sm bg-orange-100 text-orange-700 rounded-full hover:bg-orange-200 transition-colors flex items-center gap-1">
              <Icon name="heroicons:play" class="h-4 w-4" />
              {{ t('factory.quality_audio') }}
            </button>
            <button @click="tts.playProductionCapacity()" class="px-3 py-1.5 text-sm bg-green-100 text-green-700 rounded-full hover:bg-green-200 transition-colors flex items-center gap-1">
              <Icon name="heroicons:play" class="h-4 w-4" />
              {{ t('factory.capacity_audio') }}
            </button>
            <button @click="tts.playCertifications()" class="px-3 py-1.5 text-sm bg-orange-100 text-orange-700 rounded-full hover:bg-orange-200 transition-colors flex items-center gap-1">
              <Icon name="heroicons:play" class="h-4 w-4" />
              {{ t('factory.certifications_audio') }}
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- About -->
    <section class="about section section-lg bg-alt">
      <div class="container">
        <div class="section-header">
          <h2>{{ t('about.story') }}</h2>
        </div>

        <div class="about__content">
          <p class="about__text">
            {{ t('factory.about_p1') }}
          </p>
          <p class="about__text">
            {{ t('factory.about_p2') }}
          </p>

          <div class="about__gallery">
            <img src="/images/factory-1.jpg" :alt="$t('factory.alt_production')" width="400" height="300" loading="lazy" />
            <img src="/images/factory-2.jpg" :alt="$t('factory.alt_quality')" width="400" height="300" loading="lazy" />
            <img src="/images/factory-3.jpg" :alt="$t('factory.alt_packaging')" width="400" height="300" loading="lazy" />
            <img src="/images/factory-4.jpg" :alt="$t('factory.alt_warehouse')" width="400" height="300" loading="lazy" />
          </div>
        </div>
      </div>
    </section>

    <!-- Certifications -->
    <section class="certifications section section-lg">
      <div class="container">
        <div class="section-header">
          <h2>{{ t('factory.certifications') }}</h2>
          <p>{{ t('home.certifications.subtitle') }}</p>
        </div>

        <div class="certifications__grid">
          <div v-for="cert in certifications" :key="cert.id" class="cert-card">
            <div class="cert-card__icon">
              <Icon :name="cert.icon" size="40" />
            </div>
            <h3 class="cert-card__name">{{ cert.name }}</h3>
            <p class="cert-card__description">{{ cert.description }}</p>
            <div class="cert-card__details">
              <span class="cert-card__issuer">{{ cert.issuer }}</span>
              <span class="cert-card__valid">{{ $t('factory.cert_valid_until', { date: cert.validUntil }) }}</span>
            </div>
            <button v-if="cert.downloadUrl" class="cert-card__download">
              <Icon name="lucide:download" size="16" />
              {{ t('factory.download_cert') }}
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- Quality Control -->
    <section class="quality section section-lg bg-alt">
      <div class="container">
        <div class="section-header">
          <h2>{{ t('factory.quality_control') }}</h2>
          <p>{{ t('factory.quality_subtitle') }}</p>
        </div>

        <div class="quality__stages">
          <div v-for="(stage, index) in qualityStages" :key="stage.id" class="quality-stage">
            <div class="quality-stage__header">
              <div class="quality-stage__number">{{ index + 1 }}</div>
              <h3>{{ stage.title }}</h3>
            </div>
            <p class="quality-stage__description">{{ stage.description }}</p>
            <ul class="quality-stage__checks">
              <li v-for="check in stage.checks" :key="check">
                <Icon name="lucide:check" size="16" />
                {{ check }}
              </li>
            </ul>
            <div v-if="stage.image" class="quality-stage__image">
              <img :src="stage.image" :alt="stage.title" />
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Traceability -->
    <section class="traceability section">
      <div class="container">
        <div class="section-header">
          <h2>{{ t('factory.traceability') }}</h2>
          <p>{{ t('factory.traceability_subtitle') }}</p>
        </div>

        <div class="traceability__flow">
          <div v-for="(step, index) in traceabilitySteps" :key="step.id" class="trace-step">
            <div class="trace-step__icon">
              <Icon :name="step.icon" size="24" />
            </div>
            <h4>{{ step.title }}</h4>
            <p>{{ step.description }}</p>
            <div v-if="index < traceabilitySteps.length - 1" class="trace-step__arrow">
              <Icon name="lucide:arrow-right" size="20" />
            </div>
          </div>
        </div>

        <div class="traceability__features">
          <div class="traceability-feature">
            <Icon name="lucide:barcode" size="32" />
            <h4>{{ t('factory.features.batch_title') }}</h4>
            <p>{{ t('factory.features.batch_desc') }}</p>
          </div>
          <div class="traceability-feature">
            <Icon name="lucide:archive" size="32" />
            <h4>{{ t('factory.features.sample_title') }}</h4>
            <p>{{ t('factory.features.sample_desc') }}</p>
          </div>
          <div class="traceability-feature">
            <Icon name="lucide:clipboard-list" size="32" />
            <h4>{{ t('factory.features.doc_title') }}</h4>
            <p>{{ t('factory.features.doc_desc') }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Laboratory -->
    <section class="laboratory section section-lg bg-alt">
      <div class="container">
        <div class="laboratory__inner">
          <div class="laboratory__content">
            <h2>{{ t('factory.testing') }}</h2>
            <p>{{ t('factory.testing_subtitle') }}</p>

            <ul class="laboratory__capabilities">
              <li>
                <Icon name="lucide:microscope" size="20" />
                <div>
                  <strong>{{ t('factory.labs.micro_title') }}</strong>
                  <p>{{ t('factory.labs.micro_desc') }}</p>
                </div>
              </li>
              <li>
                <Icon name="lucide:flask-conical" size="20" />
                <div>
                  <strong>{{ t('factory.labs.chem_title') }}</strong>
                  <p>{{ t('factory.labs.chem_desc') }}</p>
                </div>
              </li>
              <li>
                <Icon name="lucide:palette" size="20" />
                <div>
                  <strong>{{ t('factory.labs.phys_title') }}</strong>
                  <p>{{ t('factory.labs.phys_desc') }}</p>
                </div>
              </li>
              <li>
                <Icon name="lucide:globe" size="20" />
                <div>
                  <strong>{{ t('factory.labs.third_title') }}</strong>
                  <p>{{ t('factory.labs.third_desc') }}</p>
                </div>
              </li>
            </ul>
          </div>

          <div class="laboratory__image">
            <img src="/images/laboratory.jpg" :alt="$t('factory.alt_laboratory')" width="600" height="400" loading="lazy" />
          </div>
        </div>
      </div>
    </section>

    <!-- Timeline -->
    <section class="timeline section">
      <div class="container">
        <div class="section-header">
          <h2>{{ t('factory.milestones') }}</h2>
          <p>{{ t('factory.milestones_subtitle') }}</p>
        </div>

        <div class="timeline__items">
          <div v-for="item in timeline" :key="item.year" class="timeline-item">
            <div class="timeline-item__year">{{ item.year }}</div>
            <div class="timeline-item__content">
              <h4>{{ item.milestone }}</h4>
              <p>{{ item.description }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Downloads -->
    <section class="downloads section bg-alt">
      <div class="container">
        <div class="section-header">
          <h2>{{ t('factory.documents') }}</h2>
          <p>{{ t('factory.documents_subtitle') }}</p>
        </div>

        <div class="downloads__list">
          <a
            v-for="doc in documents"
            :key="doc.id"
            :href="doc.url"
            :download="doc.filename"
            class="download-item"
          >
            <div class="download-item__icon">
              <Icon name="lucide:file-text" size="24" />
            </div>
            <div class="download-item__info">
              <h4>{{ doc.title }}</h4>
              <p>{{ doc.description }}</p>
            </div>
            <Icon name="lucide:download" size="20" class="download-item__arrow" />
          </a>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="cta section">
      <div class="container">
        <div class="cta__inner">
          <h2>{{ t('factory.visit') }}</h2>
          <p>{{ t('factory.visit_subtitle') }}</p>
          <div class="cta__actions">
            <NuxtLink :to="localePath('/contact')" class="btn btn-highlight">
              {{ t('factory.schedule_visit') }}
            </NuxtLink>
            <a :href="whatsappUrl" target="_blank" rel="noopener noreferrer" class="btn btn-outline">
              <WhatsAppIcon size="18" />
              {{ t('whatsapp.us') }}
            </a>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n, useLocalePath } from '#i18n'

const { t } = useI18n()
const localePath = useLocalePath()
const config = useRuntimeConfig()
const tts = useTTS()

// Certifications
const certifications = [
  {
    id: 1,
    name: t('factory.haccp'),
    description: t('factory.cert_haccp_desc'),
    issuer: 'SGS',
    validUntil: '2027',
    icon: 'lucide:shield-check',
    downloadUrl: '/certificates/haccp.pdf'
  },
  {
    id: 2,
    name: t('factory.iso'),
    description: t('factory.cert_iso_desc'),
    issuer: 'BSI',
    validUntil: '2027',
    icon: 'lucide:check-circle-2',
    downloadUrl: '/certificates/iso22000.pdf'
  },
  {
    id: 3,
    name: t('factory.brc'),
    description: t('factory.cert_brc_desc'),
    issuer: 'BRC',
    validUntil: '2026',
    icon: 'lucide:award',
    downloadUrl: '/certificates/brc.pdf'
  },
  {
    id: 4,
    name: t('factory.halal'),
    description: t('factory.cert_halal_desc'),
    issuer: 'IFRC',
    validUntil: '2027',
    icon: 'lucide:star',
    downloadUrl: '/certificates/halal.pdf'
  }
]

// Quality stages
const qualityStages = computed(() => [
  {
    id: 'raw',
    title: t('factory.quality_stages.raw_title'),
    description: t('factory.quality_stages.raw_desc'),
    checks: [
      t('factory.quality_stages.raw_c1'),
      t('factory.quality_stages.raw_c2'),
      t('factory.quality_stages.raw_c3'),
      t('factory.quality_stages.raw_c4')
    ],
    image: '/images/quality-raw.jpg'
  },
  {
    id: 'production',
    title: t('factory.quality_stages.prod_title'),
    description: t('factory.quality_stages.prod_desc'),
    checks: [
      t('factory.quality_stages.prod_c1'),
      t('factory.quality_stages.prod_c2'),
      t('factory.quality_stages.prod_c3'),
      t('factory.quality_stages.prod_c4')
    ],
    image: '/images/quality-production.jpg'
  },
  {
    id: 'finished',
    title: t('factory.quality_stages.fin_title'),
    description: t('factory.quality_stages.fin_desc'),
    checks: [
      t('factory.quality_stages.fin_c1'),
      t('factory.quality_stages.fin_c2'),
      t('factory.quality_stages.fin_c3'),
      t('factory.quality_stages.fin_c4')
    ],
    image: '/images/quality-finished.jpg'
  }
])

// Traceability steps
const traceabilitySteps = computed(() => [
  {
    id: 1,
    icon: 'lucide:truck',
    title: t('factory.trace.raw_title'),
    description: t('factory.trace.raw_desc')
  },
  {
    id: 2,
    icon: 'lucide:settings',
    title: t('factory.trace.prod_title'),
    description: t('factory.trace.prod_desc')
  },
  {
    id: 3,
    icon: 'lucide:package',
    title: t('factory.trace.pack_title'),
    description: t('factory.trace.pack_desc')
  },
  {
    id: 4,
    icon: 'lucide:truck',
    title: t('factory.trace.ship_title'),
    description: t('factory.trace.ship_desc')
  }
])

// Timeline
const timeline = computed(() => [
  {
    year: '2008',
    milestone: t('factory.time.founded'),
    description: t('factory.time.founded_desc')
  },
  {
    year: '2012',
    milestone: t('factory.time.export'),
    description: t('factory.time.export_desc')
  },
  {
    year: '2015',
    milestone: t('factory.time.haccp'),
    description: t('factory.time.haccp_desc')
  },
  {
    year: '2018',
    milestone: t('factory.time.expand'),
    description: t('factory.time.expand_desc')
  },
  {
    year: '2020',
    milestone: t('factory.time.iso'),
    description: t('factory.time.iso_desc')
  },
  {
    year: '2023',
    milestone: t('factory.time.lines'),
    description: t('factory.time.lines_desc')
  }
])

// Documents
const documents = computed(() => [
  {
    id: 1,
    title: t('factory.docs.profile'),
    description: t('factory.docs.profile_desc'),
    filename: 'company-profile.pdf',
    url: '/documents/company-profile.pdf'
  },
  {
    id: 2,
    title: t('factory.docs.catalog'),
    description: t('factory.docs.catalog_desc'),
    filename: 'product-catalog.pdf',
    url: '/documents/product-catalog.pdf'
  },
  {
    id: 3,
    title: t('factory.docs.haccp'),
    description: t('factory.docs.haccp_desc'),
    filename: 'haccp-certificate.pdf',
    url: '/certificates/haccp.pdf'
  },
  {
    id: 4,
    title: t('factory.docs.iso'),
    description: t('factory.docs.iso_desc'),
    filename: 'iso22000-certificate.pdf',
    url: '/certificates/iso22000.pdf'
  },
  {
    id: 5,
    title: t('factory.docs.audit'),
    description: t('factory.docs.audit_desc'),
    filename: 'audit-report.pdf',
    url: '/documents/audit-report.pdf'
  }
])

// WhatsApp URL
const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  const message = encodeURIComponent(t('whatsapp.message'))
  return `https://wa.me/${number}?text=${message}`
})

// SEO
usePageOgImage({
  title: `${t('nav.factory')} | ${t('seo.default_title')}`,
  description: 'Learn about our factory, quality certifications, and production capabilities.',
  ogType: 'website'
})
</script>

<style scoped>
.factory-hero__inner {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-3xl);
}

@media (min-width: 1024px) {
  .factory-hero__inner {
    grid-template-columns: 1fr 1fr;
    align-items: center;
  }
}

.factory-hero__title {
  font-size: clamp(2rem, 4vw, 3rem);
  margin-bottom: var(--spacing-md);
}

.factory-hero__subtitle {
  font-size: var(--text-lg);
  color: var(--color-text-light);
  margin-bottom: var(--spacing-xl);
}

.factory-hero__stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-lg);
}

.stat-item {
  text-align: center;
}

.stat-item__number {
  display: block;
  font-size: var(--text-3xl);
  font-weight: 700;
  color: var(--color-highlight);
}

.stat-item__label {
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.factory-hero__image img {
  width: 100%;
  border-radius: var(--radius-2xl);
  box-shadow: var(--shadow-xl);
}

/* About */
.about__text {
  font-size: var(--text-lg);
  line-height: 1.8;
  color: var(--color-text);
  margin-bottom: var(--spacing-lg);
}

.about__gallery {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-md);
  margin-top: var(--spacing-xl);
}

.about__gallery img {
  width: 100%;
  border-radius: var(--radius-lg);
}

/* Certifications */
.certifications__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--spacing-xl);
}

.cert-card {
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  padding: var(--spacing-xl);
  text-align: center;
}

.cert-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-full);
  color: var(--color-accent);
  margin-bottom: var(--spacing-md);
}

.cert-card__name {
  font-size: var(--text-xl);
  margin-bottom: var(--spacing-sm);
}

.cert-card__description {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin-bottom: var(--spacing-md);
}

.cert-card__details {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  margin-bottom: var(--spacing-md);
  font-size: var(--text-sm);
}

.cert-card__issuer {
  font-weight: 600;
  color: var(--color-primary);
}

.cert-card__valid {
  color: var(--color-text-light);
}

.cert-card__download {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: var(--color-primary);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  font-weight: 500;
  cursor: pointer;
}

/* Quality Stages */
.quality__stages {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-xl);
}

.quality-stage {
  background-color: white;
  border-radius: var(--radius-xl);
  padding: var(--spacing-xl);
}

.quality-stage__header {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-md);
}

.quality-stage__number {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background-color: var(--color-highlight);
  color: white;
  border-radius: var(--radius-full);
  font-weight: 600;
}

.quality-stage__description {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-md);
}

.quality-stage__checks {
  list-style: none;
  padding: 0;
  margin: 0 0 var(--spacing-md);
}

.quality-stage__checks li {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: var(--text-sm);
  margin-bottom: var(--spacing-xs);
}

.quality-stage__checks svg {
  color: var(--color-success);
}

.quality-stage__image img {
  width: 100%;
  border-radius: var(--radius-lg);
}

/* Traceability */
.traceability__flow {
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--spacing-3xl);
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.trace-step {
  flex: 1;
  min-width: 150px;
  text-align: center;
  position: relative;
}

.trace-step__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  background-color: var(--color-accent);
  border-radius: var(--radius-full);
  color: white;
  margin-bottom: var(--spacing-sm);
}

.trace-step h4 {
  font-size: var(--text-base);
  margin-bottom: var(--spacing-xs);
}

.trace-step p {
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.trace-step__arrow {
  position: absolute;
  top: 28px;
  right: -20px;
  color: var(--color-text-lighter);
}

.traceability__features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--spacing-xl);
}

.traceability-feature {
  text-align: center;
  padding: var(--spacing-xl);
}

.traceability-feature svg {
  color: var(--color-primary);
  margin-bottom: var(--spacing-md);
}

.traceability-feature h4 {
  margin-bottom: var(--spacing-sm);
}

.traceability-feature p {
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

/* Laboratory */
.laboratory__inner {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-3xl);
}

@media (min-width: 1024px) {
  .laboratory__inner {
    grid-template-columns: 1fr 1fr;
    align-items: center;
  }
}

.laboratory__capabilities {
  list-style: none;
  padding: 0;
  margin: var(--spacing-xl) 0;
}

.laboratory__capabilities li {
  display: flex;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
}

.laboratory__capabilities svg {
  flex-shrink: 0;
  color: var(--color-accent);
}

.laboratory__capabilities strong {
  display: block;
  margin-bottom: var(--spacing-xs);
}

.laboratory__capabilities p {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin: 0;
}

.laboratory__image img {
  width: 100%;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
}

/* Timeline */
.timeline__items {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-xl);
}

.timeline-item {
  position: relative;
  padding: var(--spacing-lg);
  background-color: white;
  border-radius: var(--radius-lg);
}

.timeline-item__year {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-sm);
  background-color: var(--color-highlight);
  color: white;
  font-weight: 600;
  border-radius: var(--radius-sm);
  margin-bottom: var(--spacing-sm);
}

.timeline-item h4 {
  margin-bottom: var(--spacing-xs);
}

.timeline-item p {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin: 0;
}

/* Downloads */
.downloads__list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-md);
}

.download-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md);
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  text-decoration: none;
  color: inherit;
  transition: background-color var(--transition-fast), color var(--transition-fast), transform var(--transition-fast), box-shadow var(--transition-fast);
}

.download-item:hover {
  border-color: var(--color-highlight);
  box-shadow: var(--shadow-md);
}

.download-item__icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-md);
  color: var(--color-primary);
}

.download-item__info h4 {
  font-size: var(--text-base);
  margin-bottom: var(--spacing-xs);
}

.download-item__info p {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin: 0;
}

.download-item__arrow {
  flex-shrink: 0;
  color: var(--color-text-light);
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
