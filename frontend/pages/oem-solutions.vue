<template>
  <div class="oem-page">
    <!-- Breadcrumb -->
    <div class="container">
      <Breadcrumb :items="[{ label: t('nav.oem') }]" />
    </div>

    <!-- Hero -->
    <section class="oem-hero section">
      <div class="container">
        <div class="oem-hero__inner">
          <div class="oem-hero__content">
            <h1 class="oem-hero__title">{{ t('oem.title') }}</h1>
            <p class="oem-hero__subtitle">{{ t('oem.subtitle') }}</p>

            <div class="oem-hero__actions">
              <a :href="whatsappUrl" target="_blank" rel="noopener noreferrer" class="btn btn-highlight btn-lg">
                <WhatsAppIcon size="20" />
                {{ t('whatsapp.us') }}
              </a>
              <NuxtLink :to="localePath('/contact')" class="btn btn-outline btn-lg">
                {{ t('form.submit') }}
              </NuxtLink>
            </div>
          </div>

          <div class="oem-hero__visual">
            <div class="oem-hero__process">
              <div v-for="(step, index) in processSteps" :key="index" class="process-step">
                <div class="process-step__number">{{ index + 1 }}</div>
                <div class="process-step__content">
                  <h4>{{ step.title }}</h4>
                  <p>{{ step.description }}</p>
                </div>
                <div v-if="index < processSteps.length - 1" class="process-step__arrow">
                  <Icon name="lucide:arrow-down" size="20" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- OEM Paths -->
    <section class="oem-paths section section-lg bg-alt">
      <div class="container">
        <div class="section-header">
          <h2>{{ t('oem.process') }}</h2>
          <p>{{ t('oem.process_subtitle') }}</p>
        </div>

        <div class="oem-paths__grid">
          <!-- Quick ODM -->
          <div class="oem-path oem-path--featured">
            <div class="oem-path__badge">{{ t('oem.quick_odm') }}</div>
            <h3 class="oem-path__title">{{ t('oem.odm_subtitle') }}</h3>
            <p class="oem-path__description">
              {{ t('oem.odm_desc') }}
            </p>

            <ul class="oem-path__features">
              <li>
                <Icon name="lucide:check-circle" size="18" />
                <span>{{ t('oem.odm_f1') }}</span>
              </li>
              <li>
                <Icon name="lucide:check-circle" size="18" />
                <span>{{ t('oem.odm_f2') }}</span>
              </li>
              <li>
                <Icon name="lucide:check-circle" size="18" />
                <span>{{ t('oem.odm_f3') }}</span>
              </li>
              <li>
                <Icon name="lucide:check-circle" size="18" />
                <span>{{ t('oem.odm_f4') }}</span>
              </li>
            </ul>

            <div class="oem-path__timeline">
              <div class="timeline-item">
                <span class="timeline-item__label">{{ t('oem.timeline_design') }}</span>
                <span class="timeline-item__time">3-5 {{ t('oem.days') }}</span>
              </div>
              <div class="timeline-item">
                <span class="timeline-item__label">{{ t('oem.timeline_sample') }}</span>
                <span class="timeline-item__time">5-7 {{ t('oem.days') }}</span>
              </div>
              <div class="timeline-item">
                <span class="timeline-item__label">{{ t('oem.timeline_production') }}</span>
                <span class="timeline-item__time">10-14 {{ t('oem.days') }}</span>
              </div>
            </div>

            <NuxtLink :to="localePath('/contact')" class="btn btn-primary">
              {{ t('oem.start_odm') }}
            </NuxtLink>
          </div>

          <!-- Full OEM -->
          <div class="oem-path">
            <div class="oem-path__badge oem-path__badge--secondary">{{ t('oem.full_oem') }}</div>
            <h3 class="oem-path__title">{{ t('oem.oem_subtitle') }}</h3>
            <p class="oem-path__description">
              {{ t('oem.oem_desc') }}
            </p>

            <ul class="oem-path__features">
              <li>
                <Icon name="lucide:check-circle" size="18" />
                <span>{{ t('oem.oem_f1') }}</span>
              </li>
              <li>
                <Icon name="lucide:check-circle" size="18" />
                <span>{{ t('oem.oem_f2') }}</span>
              </li>
              <li>
                <Icon name="lucide:check-circle" size="18" />
                <span>{{ t('oem.oem_f3') }}</span>
              </li>
              <li>
                <Icon name="lucide:check-circle" size="18" />
                <span>{{ t('oem.oem_f4') }}</span>
              </li>
              <li>
                <Icon name="lucide:check-circle" size="18" />
                <span>{{ t('oem.oem_f5') }}</span>
              </li>
            </ul>

            <div class="oem-path__timeline">
              <div class="timeline-item">
                <span class="timeline-item__label">{{ t('oem.timeline_rd') }}</span>
                <span class="timeline-item__time">1-2 {{ t('oem.weeks') }}</span>
              </div>
              <div class="timeline-item">
                <span class="timeline-item__label">{{ t('oem.timeline_sample') }}</span>
                <span class="timeline-item__time">1-2 {{ t('oem.weeks') }}</span>
              </div>
              <div class="timeline-item">
                <span class="timeline-item__label">{{ t('oem.timeline_production') }}</span>
                <span class="timeline-item__time">2-4 {{ t('oem.weeks') }}</span>
              </div>
            </div>

            <NuxtLink :to="localePath('/contact')" class="btn btn-outline">
              {{ t('oem.start_oem') }}
            </NuxtLink>
          </div>
        </div>
      </div>
    </section>

    <!-- Solutions -->
    <section class="solutions section section-lg">
      <div class="container">
        <div class="section-header">
          <h2>{{ t('cases.title') }}</h2>
          <p>{{ t('oem.solutions_subtitle') }}</p>
        </div>

        <div class="solutions__grid">
          <div v-for="solution in solutions" :key="solution.id" class="solution-card">
            <div class="solution-card__image">
              <img :src="solution.image" :alt="solution.title" />
              <div class="solution-card__overlay">
                <span class="solution-card__moq">MOQ: {{ solution.moq }}</span>
              </div>
            </div>
            <div class="solution-card__content">
              <h3>{{ solution.title }}</h3>
              <p>{{ solution.description }}</p>
              <ul class="solution-card__applications">
                <li v-for="app in solution.applications" :key="app">
                  {{ app }}
                </li>
              </ul>
              <button class="solution-card__cta" @click="inquireAbout(solution)">
                {{ t('product.inquire_now') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- MOQ & Lead Time Table -->
    <section class="moq-table section bg-alt">
      <div class="container container-narrow">
        <div class="section-header">
          <h2>{{ t('oem.moq') }} & {{ t('oem.lead_time') }}</h2>
          <p>{{ t('oem.moq_subtitle') }}</p>
        </div>

        <table class="moq-table__table">
          <thead>
            <tr>
              <th>{{ t('oem.table_order_type') }}</th>
              <th>{{ t('oem.moq') }}</th>
              <th>{{ t('oem.lead_time') }}</th>
              <th>{{ t('oem.table_setup') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>{{ t('oem.table_quick_odm') }}</td>
              <td>{{ t('oem.moq_odm') }}</td>
              <td>{{ t('oem.lead_odm') }}</td>
              <td>{{ t('oem.setup_odm') }}</td>
            </tr>
            <tr>
              <td>{{ t('oem.table_full_oem') }}</td>
              <td>{{ t('oem.moq_oem') }}</td>
              <td>{{ t('oem.lead_oem') }}</td>
              <td>{{ t('oem.setup_oem') }}</td>
            </tr>
            <tr>
              <td>{{ t('oem.table_gift_set') }}</td>
              <td>{{ t('oem.moq_gift') }}</td>
              <td>{{ t('oem.lead_gift') }}</td>
              <td>{{ t('oem.setup_gift') }}</td>
            </tr>
            <tr>
              <td>{{ t('oem.table_contract') }}</td>
              <td>{{ t('oem.moq_contract') }}</td>
              <td>{{ t('oem.lead_contract') }}</td>
              <td>{{ t('oem.setup_contract') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- Sample Process -->
    <section class="sample-process section">
      <div class="container">
        <div class="section-header">
          <h2>{{ t('oem.sample') }}</h2>
          <p>{{ t('oem.sample_subtitle') }}</p>
        </div>

        <div class="sample-process__steps">
          <div v-for="(step, index) in sampleSteps" :key="index" class="sample-step">
            <div class="sample-step__icon">
              <Icon :name="step.icon" size="24" />
            </div>
            <div class="sample-step__content">
              <h4>{{ step.title }}</h4>
              <p>{{ step.description }}</p>
              <span class="sample-step__duration">{{ step.duration }}</span>
            </div>
            <div v-if="index < sampleSteps.length - 1" class="sample-step__line"></div>
          </div>
        </div>

        <div class="sample-process__cta">
          <NuxtLink :to="localePath('/contact')" class="btn btn-highlight">
            {{ t('product.request_sample') }}
          </NuxtLink>
          <p class="sample-process__note">
            {{ t('oem.sample_note') }}
          </p>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="cta section">
      <div class="container">
        <div class="cta__inner">
          <h2>{{ t('oem.cta_title') }}</h2>
          <p>{{ t('oem.cta_subtitle') }}</p>
          <div class="cta__actions">
            <a :href="whatsappUrl" target="_blank" rel="noopener noreferrer" class="btn btn-highlight btn-lg">
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
  </div>
</template>

<script setup lang="ts">
import { useI18n, useLocalePath } from '#i18n'

const { t } = useI18n()
const localePath = useLocalePath()
const config = useRuntimeConfig()
const route = useRoute()

// Process steps
const processSteps = computed(() => [
  { title: t('oem.steps.consultation'), description: t('oem.steps.consultation_desc') },
  { title: t('oem.steps.proposal'), description: t('oem.steps.proposal_desc') },
  { title: t('oem.steps.sampling'), description: t('oem.steps.sampling_desc') },
  { title: t('oem.steps.production'), description: t('oem.steps.production_desc') },
  { title: t('oem.steps.delivery'), description: t('oem.steps.delivery_desc') }
])

// Solutions
const solutions = computed(() => [
  {
    id: 1,
    title: t('oem.sols.gift_sets'),
    description: t('oem.sols.gift_sets_desc'),
    image: '/images/oem/gift-sets.jpg',
    moq: '500 sets',
    applications: [t('oem.apps.holiday_gifts'), t('oem.apps.corporate_gifts'), t('oem.apps.wedding_favors')]
  },
  {
    id: 2,
    title: t('oem.sols.toy_candies'),
    description: t('oem.sols.toy_candies_desc'),
    image: '/images/oem/toy-candies.jpg',
    moq: '3,000 pcs',
    applications: [t('oem.apps.supermarkets'), t('oem.apps.toy_stores'), t('oem.apps.amusement_parks')]
  },
  {
    id: 3,
    title: t('oem.sols.mixed_buckets'),
    description: t('oem.sols.mixed_buckets_desc'),
    image: '/images/oem/mixed-buckets.jpg',
    moq: '1,000 pcs',
    applications: [t('oem.apps.cinema'), t('oem.apps.convenience_stores'), t('oem.apps.vending')]
  },
  {
    id: 4,
    title: t('oem.sols.functional'),
    description: t('oem.sols.functional_desc'),
    image: '/images/oem/functional.jpg',
    moq: '5,000 pcs',
    applications: [t('oem.apps.health_stores'), t('oem.apps.pharmacies'), t('oem.apps.online_retail')]
  }
])

// Sample process steps
const sampleSteps = computed(() => [
  {
    icon: 'lucide:message-circle',
    title: t('oem.sample_process.request'),
    description: t('oem.sample_process.request_desc'),
    duration: t('oem.durations.1_day')
  },
  {
    icon: 'lucide:flask-conical',
    title: t('oem.sample_process.production'),
    description: t('oem.sample_process.production_desc'),
    duration: t('oem.durations.3_7_days')
  },
  {
    icon: 'lucide:package',
    title: t('oem.sample_process.delivery'),
    description: t('oem.sample_process.delivery_desc'),
    duration: t('oem.durations.3_5_days')
  },
  {
    icon: 'lucide:check-circle',
    title: t('oem.sample_process.approval'),
    description: t('oem.sample_process.approval_desc'),
    duration: t('oem.durations.1_3_days')
  },
  {
    icon: 'lucide:rocket',
    title: t('oem.sample_process.bulk'),
    description: t('oem.sample_process.bulk_desc'),
    duration: t('oem.durations.based_on_order')
  }
])

const inquireAbout = (solution: typeof solutions[0]) => {
  // Navigate to contact page with pre-filled info
  navigateTo({
    path: localePath('/contact'),
    query: {
      product: solution.title,
      category: 'OEM Solutions'
    }
  })
}

// WhatsApp URL
const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  const message = encodeURIComponent(t('whatsapp.message'))
  return `https://wa.me/${number}?text=${message}`
})

// SEO
useSeo({
  title: `${t('nav.oem')} | ${t('seo.default_title')}`,
  description: t('oem.subtitle'),
  ogType: 'website'
})
</script>

<style scoped>
.oem-hero__inner {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-3xl);
}

@media (min-width: 1024px) {
  .oem-hero__inner {
    grid-template-columns: 1fr 1fr;
    align-items: center;
  }
}

.oem-hero__title {
  font-size: clamp(2rem, 4vw, 3rem);
  margin-bottom: var(--spacing-md);
}

.oem-hero__subtitle {
  font-size: var(--text-xl);
  color: var(--color-text-light);
  margin-bottom: var(--spacing-xl);
}

.oem-hero__actions {
  display: flex;
  gap: var(--spacing-md);
  flex-wrap: wrap;
}

/* Process visual */
.oem-hero__process {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.process-step {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.process-step__number {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background-color: var(--color-highlight);
  color: white;
  border-radius: var(--radius-full);
  font-weight: 600;
  flex-shrink: 0;
}

.process-step__content h4 {
  font-size: var(--text-base);
  margin-bottom: 0;
}

.process-step__content p {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin: 0;
}

.process-step__arrow {
  color: var(--color-text-lighter);
  margin-left: var(--spacing-md);
}

/* OEM Paths */
.oem-paths__grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-2xl);
}

@media (min-width: 1024px) {
  .oem-paths__grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.oem-path {
  padding: var(--spacing-2xl);
  background-color: white;
  border-radius: var(--radius-xl);
  border: 2px solid var(--color-border);
}

.oem-path--featured {
  border-color: var(--color-highlight);
  background: linear-gradient(to bottom, #fff5f2 0%, white 20%);
}

.oem-path__badge {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-md);
  background-color: var(--color-highlight);
  color: white;
  font-size: var(--text-sm);
  font-weight: 600;
  border-radius: var(--radius-full);
  margin-bottom: var(--spacing-md);
}

.oem-path__badge--secondary {
  background-color: var(--color-primary);
}

.oem-path__title {
  font-size: var(--text-xl);
  margin-bottom: var(--spacing-md);
}

.oem-path__description {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-lg);
}

.oem-path__features {
  list-style: none;
  padding: 0;
  margin: 0 0 var(--spacing-xl);
}

.oem-path__features li {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
  font-size: var(--text-sm);
}

.oem-path__features svg {
  color: var(--color-success);
}

.oem-path__timeline {
  display: flex;
  justify-content: space-between;
  padding: var(--spacing-lg);
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-lg);
  margin-bottom: var(--spacing-xl);
}

.timeline-item {
  text-align: center;
}

.timeline-item__label {
  display: block;
  font-size: var(--text-xs);
  color: var(--color-text-light);
  margin-bottom: var(--spacing-xs);
}

.timeline-item__time {
  display: block;
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-primary);
}

/* Solutions */
.solutions__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: var(--spacing-xl);
}

.solution-card {
  background-color: white;
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-md);
}

.solution-card__image {
  position: relative;
  aspect-ratio: 4/3;
}

.solution-card__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.solution-card__overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: var(--spacing-md);
  background: linear-gradient(to top, rgba(0,0,0,0.7) 0%, transparent 100%);
}

.solution-card__moq {
  color: white;
  font-size: var(--text-sm);
  font-weight: 500;
}

.solution-card__content {
  padding: var(--spacing-lg);
}

.solution-card__content h3 {
  font-size: var(--text-lg);
  margin-bottom: var(--spacing-sm);
}

.solution-card__content p {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin-bottom: var(--spacing-md);
}

.solution-card__applications {
  list-style: none;
  padding: 0;
  margin: 0 0 var(--spacing-md);
}

.solution-card__applications li {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-sm);
  margin-right: var(--spacing-xs);
  margin-bottom: var(--spacing-xs);
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-sm);
  font-size: var(--text-xs);
}

.solution-card__cta {
  width: 100%;
  padding: var(--spacing-sm);
  background-color: var(--color-primary);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-weight: 500;
  cursor: pointer;
  transition: background-color var(--transition-fast);
}

.solution-card__cta:hover {
  background-color: var(--color-primary-light);
}

/* MOQ Table */
.moq-table__table {
  width: 100%;
  border-collapse: collapse;
  background-color: white;
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.moq-table__table th,
.moq-table__table td {
  padding: var(--spacing-md);
  text-align: left;
  border-bottom: 1px solid var(--color-border-light);
}

.moq-table__table th {
  background-color: var(--color-primary);
  color: white;
  font-weight: 600;
}

.moq-table__table tr:last-child td {
  border-bottom: none;
}

/* Sample Process */
.sample-process__steps {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
  margin-bottom: var(--spacing-3xl);
}

.sample-step {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-lg);
}

.sample-step__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background-color: var(--color-accent);
  border-radius: var(--radius-full);
  color: white;
  flex-shrink: 0;
}

.sample-step__content {
  flex: 1;
}

.sample-step__content h4 {
  font-size: var(--text-lg);
  margin-bottom: var(--spacing-xs);
}

.sample-step__content p {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-xs);
}

.sample-step__duration {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-sm);
  background-color: var(--color-highlight);
  color: white;
  font-size: var(--text-xs);
  font-weight: 600;
  border-radius: var(--radius-sm);
}

.sample-step__line {
  position: absolute;
  left: 24px;
  top: 48px;
  width: 2px;
  height: calc(100% + var(--spacing-xl));
  background: linear-gradient(to bottom, var(--color-accent) 0%, var(--color-highlight) 100%);
}

.sample-process__cta {
  text-align: center;
}

.sample-process__note {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin-top: var(--spacing-md);
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
