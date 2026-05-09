<template>
  <div class="home-page">
    <!-- Hero Section -->
    <section class="hero">
      <div class="hero__bg">
        <div class="hero__pattern"></div>
        <div class="hero__gradient"></div>
      </div>

      <div class="container hero__container">
        <div class="hero__content">
          <div class="hero__badge">
            <span class="hero__badge-dot"></span>
            {{ $t('home.hero.subtitle') }}
          </div>

          <h1 class="hero__title">
            {{ $t('home.hero.title') }}
          </h1>

          <p class="hero__desc">
            {{ $t('home.hero.subtitle') }}
          </p>

          <div class="hero__actions">
            <NuxtLink :to="orderingCtaPath" class="btn btn-highlight btn-lg hero__btn">
              <Icon name="lucide:send" size="18" />
              {{ orderingCtaLabel }}
            </NuxtLink>
            <NuxtLink :to="localePath('/products')" class="btn btn-outline btn-lg hero__btn">
              <Icon name="lucide:package" size="18" />
              {{ $t('home.hero.cta_secondary') }}
            </NuxtLink>
          </div>

          <div class="hero__stats">
            <div class="hero__stat" v-for="stat in heroStats" :key="stat.label">
              <span class="hero__stat-num" :data-target="stat.value">{{ stat.display }}</span>
              <span class="hero__stat-label">{{ stat.label }}</span>
            </div>
          </div>

          <div class="hero__badges">
            <div class="hero__cert" v-for="cert in ['HACCP', 'ISO22000', 'BRC', 'HALAL']" :key="cert">
              <Icon name="lucide:shield-check" size="14" />
              {{ cert }}
            </div>
          </div>
        </div>

        <div class="hero__visual">
          <div class="hero__image-wrap">
            <img src="/images/hero-candy.jpg" :alt="$t('home.hero.title')" class="hero__img" loading="eager" />
            <div class="hero__image-accent"></div>
          </div>
          <div class="hero__float hero__float--1">
            <Icon name="lucide:award" size="20" />
            <span>15+ {{ $t('factory.years') }}</span>
          </div>
          <div class="hero__float hero__float--2">
            <Icon name="lucide:globe" size="20" />
            <span>50+ {{ $t('factory.countries') }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Categories Section -->
    <section class="categories section section-lg">
      <div class="container">
        <div class="section-header">
          <span class="section-tag">{{ $t('home.categories.title') }}</span>
          <h2>{{ $t('home.categories.title') }}</h2>
          <p>{{ $t('home.categories.subtitle') }}</p>
        </div>

        <div v-if="categoriesStatus === 'pending'" class="categories__grid">
          <SkeletonLoader v-for="i in 4" :key="i" width="100%" height="280px" borderRadius="16px" />
        </div>
        <div v-else-if="categoriesStatus === 'error' || !categories.length">
          <ErrorState :title="t('errors.default')" :message="t('offline_message')" :actionText="t('errors.tryAgain')" @action="refreshCategories" />
        </div>
        <div v-else class="categories__grid">
          <NuxtLink v-for="category in categories" :key="category.slug" :to="localePath(category.to)" class="cat-card">
            <div class="cat-card__img">
              <img :src="category.image" :alt="category.name" loading="lazy" />
              <div class="cat-card__overlay">
                <span class="cat-card__arrow"><Icon name="lucide:arrow-right" size="20" /></span>
              </div>
            </div>
            <div class="cat-card__body">
              <h3 class="cat-card__name">{{ category.name }}</h3>
              <span class="cat-card__count">{{ category.count }} {{ $t('product.products') }}</span>
            </div>
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- Why Choose Us + Certifications -->
    <section class="why section section-lg bg-alt">
      <div class="container">
        <div class="section-header">
          <span class="section-tag">{{ $t('home.trust.title') }}</span>
          <h2>{{ $t('home.trust.title') }}</h2>
          <p>{{ $t('home.trust.subtitle') }}</p>
        </div>

        <div class="why__grid">
          <div v-for="item in trustItems" :key="item.title" class="why-card">
            <div class="why-card__icon" :style="{ backgroundColor: item.color }">
              <Icon :name="item.icon" size="28" />
            </div>
            <h3 class="why-card__title">{{ item.title }}</h3>
            <p class="why-card__desc">{{ item.description }}</p>
          </div>
        </div>

        <div class="why__certs">
          <div v-for="cert in certifications" :key="cert.name" class="cert-badge">
            <div class="cert-badge__icon">
              <Icon :name="cert.icon" size="28" />
            </div>
            <div class="cert-badge__info">
              <h4>{{ cert.name }}</h4>
              <p>{{ cert.description }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Featured Products -->
    <section class="featured section section-lg">
      <div class="container">
        <div class="section-header">
          <span class="section-tag">{{ $t('home.featured.title') }}</span>
          <h2>{{ $t('home.featured.title') }}</h2>
          <p>{{ $t('home.featured.subtitle') }}</p>
        </div>

        <div v-if="featuredStatus === 'pending'" class="featured__grid">
          <SkeletonLoader v-for="i in 4" :key="i" width="100%" height="320px" borderRadius="16px" />
        </div>
        <div v-else-if="featuredStatus === 'error' || !featuredProducts.length">
          <ErrorState :title="t('errors.default')" :message="t('offline_message')" :actionText="t('errors.tryAgain')" @action="refreshFeatured" />
        </div>
        <div v-else class="featured__grid">
          <ProductCard v-for="product in featuredProducts" :key="product.id" :product="product" />
        </div>

        <div class="featured__cta">
          <NuxtLink :to="localePath('/products')" class="btn btn-outline btn-lg">
            {{ $t('product.view_products') }}
            <Icon name="lucide:arrow-right" size="16" />
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- How to Order -->
    <section class="steps section section-lg bg-alt">
      <div class="container">
        <div class="section-header">
          <span class="section-tag">{{ $t('home.how_to_order.title') }}</span>
          <h2>{{ $t('home.how_to_order.title') }}</h2>
          <p>{{ $t('home.how_to_order.subtitle') }}</p>
        </div>

        <div class="steps__track">
          <div class="steps__line"></div>
          <div class="steps__item" v-for="(step, idx) in orderSteps" :key="idx">
            <div class="steps__num">{{ idx + 1 }}</div>
            <div class="steps__icon"><Icon :name="step.icon" size="28" /></div>
            <h3 class="steps__title">{{ step.title }}</h3>
            <p class="steps__desc">{{ step.desc }}</p>
          </div>
        </div>

        <div class="steps__cta">
          <NuxtLink :to="orderingCtaPath" class="btn btn-highlight btn-lg">
            {{ orderingCtaLabel }}
            <Icon name="lucide:arrow-right" size="16" />
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- OEM Services -->
    <section class="oem section section-lg">
      <div class="container">
        <div class="oem__inner">
          <div class="oem__content">
            <span class="section-tag">{{ $t('oem.title') }}</span>
            <h2>{{ $t('home.oem.title') }}</h2>
            <p class="oem__subtitle">{{ $t('home.oem.subtitle') }}</p>

            <ul class="oem__features">
              <li v-for="feature in oemFeatures" :key="feature">
                <Icon name="lucide:check-circle" size="18" />
                <span>{{ feature }}</span>
              </li>
            </ul>

            <NuxtLink :to="localePath('/oem-solutions')" class="btn btn-highlight btn-lg">
              {{ $t('home.oem.cta') }}
              <Icon name="lucide:arrow-right" size="16" />
            </NuxtLink>
          </div>

          <div class="oem__visual">
            <img src="/images/oem-process.jpg" :alt="$t('oem.title')" loading="lazy" />
            <div class="oem__visual-accent"></div>
          </div>
        </div>
      </div>
    </section>

    <!-- Global Reach -->
    <section class="global section section-lg bg-alt">
      <div class="container">
        <div class="section-header">
          <span class="section-tag">{{ $t('home.global.title') }}</span>
          <h2>{{ $t('home.global.title') }}</h2>
          <p>{{ $t('home.global.subtitle') }}</p>
        </div>

        <div class="global__grid">
          <div v-for="region in markets" :key="region.name" class="region-card">
            <div class="region-card__flag">{{ region.flag }}</div>
            <h4 class="region-card__name">{{ region.name }}</h4>
            <p class="region-card__countries">{{ region.countries }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- FAQ -->
    <section class="faq section section-lg">
      <div class="container container-narrow">
        <div class="section-header">
          <span class="section-tag">{{ $t('home.faq.title') }}</span>
          <h2>{{ $t('home.faq.title') }}</h2>
          <p>{{ $t('home.faq.subtitle') }}</p>
        </div>

        <div class="faq__list">
          <div v-for="(faq, index) in faqs" :key="index" class="faq-item" :class="{ 'faq-item--open': openFaq === index }">
            <button class="faq-item__q" @click="toggleFaq(index)" :aria-expanded="openFaq === index">
              <span>{{ faq.question }}</span>
              <Icon name="lucide:chevron-down" size="18" class="faq-item__chevron" />
            </button>
            <Transition name="faq-slide">
              <div v-if="openFaq === index" class="faq-item__a">
                <p>{{ faq.answer }}</p>
              </div>
            </Transition>
          </div>
        </div>

        <div class="faq__more">
          <NuxtLink :to="localePath('/faq')" class="btn btn-ghost">
            {{ $t('home.faq.view_all') }}
            <Icon name="lucide:arrow-right" size="16" />
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- Final CTA Banner -->
    <section class="cta-banner section section-lg">
      <div class="cta-banner__bg"></div>
      <div class="container">
        <div class="cta-banner__inner">
          <h2>{{ $t('form.title') }}</h2>
          <p>{{ $t('form.subtitle') }}</p>

          <div class="cta-banner__actions">
            <NuxtLink :to="orderingCtaPath" class="btn btn-highlight btn-lg">
              <Icon name="lucide:send" size="18" />
              {{ orderingCtaLabel }}
            </NuxtLink>
            <a :href="whatsappUrl" target="_blank" rel="noopener noreferrer" class="btn btn-outline btn-lg cta-banner__wa">
              <WhatsAppIcon size="20" />
              {{ $t('whatsapp.us') }}
            </a>
            <NuxtLink :to="localePath('/contact')" class="btn btn-ghost btn-lg">
              {{ $t('contact.send_inquiry') }}
              <Icon name="lucide:arrow-right" size="16" />
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
const config = useRuntimeConfig()
const { isAuthenticated, isAdmin, isPending } = useAuth()
const { getCategories, getFeaturedProducts } = useApi()

const orderingCtaPath = computed(() => {
  if (isAuthenticated.value) {
    return localePath('/contact')
  }
  return localePath('/auth/register')
})

const orderingCtaLabel = computed(() => {
  if (isPending.value) {
    return t('product.pending_approval')
  }
  if (isAuthenticated.value) {
    return t('product.inquire_now')
  }
  return t('product.register_to_order')
})

const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  const message = encodeURIComponent(t('whatsapp.message'))
  return `https://wa.me/${number}?text=${message}`
})

// Hero stats
const heroStats = computed(() => [
  { display: '15+', value: 15, label: t('factory.years') },
  { display: '50+', value: 50, label: t('factory.countries') },
  { display: '100T', value: 100, label: t('factory.daily_output') },
])

// Categories
const { data: categoriesData, status: categoriesStatus, refresh: refreshCategories } = await useAsyncData(
  () => `home-categories-${locale.value}`,
  async () => {
    return await getCategories()
  }
)

const categories = computed(() => {
  return (categoriesData.value || []).map(category => ({
    slug: category.slug,
    to: `/products/${category.slug}`,
    name: category.name || t(`product.categories.${category.slug.replace(/-/g, '_')}`),
    count: category.productCount || 0,
    image: category.thumbnail || `/images/categories/${category.slug}.jpg`
  }))
})

// Trust items
const trustItems = computed(() => [
  { icon: 'lucide:award', title: t('factory.certifications'), description: t('factory.haccp') + ', ' + t('factory.iso'), color: 'rgba(234,88,12,0.08)' },
  { icon: 'lucide:users', title: t('home.trust.expert_team'), description: t('home.trust.expert_team_desc'), color: 'rgba(34,197,94,0.08)' },
  { icon: 'lucide:package', title: t('factory.capacity'), description: t('home.trust.capacity_desc'), color: 'rgba(59,130,246,0.08)' },
  { icon: 'lucide:globe', title: t('factory.export'), description: t('home.trust.export_desc'), color: 'rgba(245,158,11,0.08)' },
])

// Certifications
const certifications = computed(() => [
  { name: t('factory.haccp'), description: t('home.certifications.haccp_desc'), icon: 'lucide:shield-check' },
  { name: t('factory.iso'), description: t('home.certifications.iso_desc'), icon: 'lucide:check-circle-2' },
  { name: t('factory.brc'), description: t('home.certifications.brc_desc'), icon: 'lucide:award' },
  { name: t('factory.halal'), description: t('home.certifications.halal_desc'), icon: 'lucide:star' },
])

// Featured products
const { data: featuredProductsData, status: featuredStatus, refresh: refreshFeatured } = await useAsyncData(
  () => `home-featured-products-${locale.value}`,
  async () => {
    return await getFeaturedProducts(8)
  }
)

const featuredProducts = computed(() => featuredProductsData.value || [])

// Order steps
const orderSteps = computed(() => [
  { icon: 'lucide:search', title: t('home.how_to_order.step1_title'), desc: t('home.how_to_order.step1_desc') },
  { icon: 'lucide:user-plus', title: t('home.how_to_order.step2_title'), desc: t('home.how_to_order.step2_desc') },
  { icon: 'lucide:truck', title: t('home.how_to_order.step3_title'), desc: t('home.how_to_order.step3_desc') },
])

// OEM Features
const oemFeatures = computed(() => [
  t('home.oem.feature1'),
  t('home.oem.feature2'),
  t('home.oem.feature3'),
  t('home.oem.feature4'),
  t('home.oem.feature5'),
])

// Global Markets
const markets = computed(() => [
  { name: t('home.global.north_america'), flag: '🇺🇸', countries: t('home.global.north_america_countries') },
  { name: t('home.global.europe'), flag: '🇪🇺', countries: t('home.global.europe_countries') },
  { name: t('home.global.southeast_asia'), flag: '🌏', countries: t('home.global.southeast_asia_countries') },
  { name: t('home.global.middle_east'), flag: '🕌', countries: t('home.global.middle_east_countries') },
])

// FAQ
const openFaq = ref(0)

const faqs = computed(() => [
  { question: t('home.faq.q1'), answer: t('home.faq.a1') },
  { question: t('home.faq.q2'), answer: t('home.faq.a2') },
  { question: t('home.faq.q3'), answer: t('home.faq.a3') },
  { question: t('home.faq.q4'), answer: t('home.faq.a4') },
])

const toggleFaq = (index: number) => {
  openFaq.value = openFaq.value === index ? -1 : index
}

// SEO
useSeo({
  title: t('seo.home_title'),
  description: t('seo.home_description'),
  ogType: 'website',
})
</script>

<style scoped>
/* ===== Section Tag ===== */
.section-tag {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-md);
  background: rgba(var(--color-highlight-rgb), 0.08);
  color: var(--color-highlight);
  font-size: var(--text-sm);
  font-weight: 600;
  border-radius: var(--radius-full);
  letter-spacing: 0.02em;
  text-transform: uppercase;
  margin-bottom: var(--spacing-md);
}

/* ===== HERO ===== */
.hero {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  padding-top: var(--header-height);
  overflow: hidden;
}

.hero__bg {
  position: absolute;
  inset: 0;
  background: linear-gradient(160deg, #fffbf5 0%, #fff7ed 40%, #ffedd5 100%);
}

.hero__pattern {
  position: absolute;
  inset: 0;
  background-image: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none'%3E%3Cg fill='%23ea580c' fill-opacity='0.03'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
}

.hero__gradient {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 70% 50% at 75% 50%, rgba(var(--color-highlight-rgb), 0.06) 0%, transparent 70%),
    radial-gradient(ellipse 50% 60% at 20% 80%, rgba(var(--color-highlight-rgb), 0.04) 0%, transparent 60%);
}

.hero__container {
  position: relative;
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-3xl);
  padding-top: var(--spacing-3xl);
  padding-bottom: var(--spacing-3xl);
}

@media (min-width: 1024px) {
  .hero__container {
    grid-template-columns: 1.1fr 0.9fr;
    align-items: center;
    gap: var(--spacing-4xl);
    padding-top: var(--spacing-4xl);
    padding-bottom: var(--spacing-4xl);
  }
}

.hero__content {
  animation: heroFadeIn 0.7s ease-out;
}

.hero__badge {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: 6px var(--spacing-md);
  background: rgba(var(--color-highlight-rgb), 0.1);
  border: 1px solid rgba(var(--color-highlight-rgb), 0.15);
  border-radius: var(--radius-full);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-highlight);
  margin-bottom: var(--spacing-xl);
}

.hero__badge-dot {
  width: 8px;
  height: 8px;
  background: var(--color-highlight);
  border-radius: var(--radius-full);
  animation: pulse 2s infinite;
}

.hero__title {
  font-size: clamp(2.25rem, 5vw, 3.75rem);
  line-height: 1.08;
  font-weight: 800;
  color: var(--color-primary);
  margin-bottom: var(--spacing-lg);
  letter-spacing: -0.02em;
}

.hero__desc {
  font-size: var(--text-lg);
  color: var(--color-text-light);
  max-width: 520px;
  margin-bottom: var(--spacing-xl);
  line-height: 1.7;
}

.hero__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-2xl);
}

.hero__btn {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.hero__stats {
  display: flex;
  gap: var(--spacing-2xl);
  margin-bottom: var(--spacing-xl);
  padding-top: var(--spacing-xl);
  border-top: 1px solid var(--color-border);
}

.hero__stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.hero__stat-num {
  font-size: var(--text-3xl);
  font-weight: 800;
  color: var(--color-highlight);
  font-family: var(--font-display);
}

.hero__stat-label {
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.hero__badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

.hero__cert {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-text-light);
}

.hero__cert svg {
  color: var(--color-success);
}

/* Hero Visual */
.hero__visual {
  position: relative;
  animation: heroFadeIn 0.7s ease-out 0.15s backwards;
}

.hero__image-wrap {
  position: relative;
  border-radius: var(--radius-2xl);
  overflow: hidden;
  box-shadow: var(--shadow-2xl);
}

.hero__img {
  width: 100%;
  height: auto;
  display: block;
}

.hero__image-accent {
  position: absolute;
  inset: 0;
  border-radius: var(--radius-2xl);
  box-shadow: inset 0 0 0 1px rgba(255,255,255,0.2);
  pointer-events: none;
}

.hero__float {
  position: absolute;
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  background: white;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  font-size: var(--text-sm);
  font-weight: 600;
  color: var(--color-primary);
  white-space: nowrap;
}

.hero__float svg {
  color: var(--color-highlight);
}

.hero__float--1 {
  bottom: 20%;
  left: -12px;
  animation: floatBounce 3s ease-in-out infinite;
}

.hero__float--2 {
  top: 15%;
  right: -12px;
  animation: floatBounce 3s ease-in-out 1.5s infinite;
}

@media (max-width: 1023px) {
  .hero__float { display: none; }
  .hero__stats { gap: var(--spacing-xl); }
}

/* ===== CATEGORIES ===== */
.categories__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: var(--spacing-lg);
}

.cat-card {
  display: block;
  background: white;
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--color-border-light);
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.cat-card:hover {
  transform: translateY(-6px);
  box-shadow: var(--shadow-xl);
  border-color: rgba(var(--color-highlight-rgb), 0.2);
}

.cat-card__img {
  position: relative;
  aspect-ratio: 4/3;
  overflow: hidden;
  background: var(--color-bg-alt);
}

.cat-card__img img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-slow);
}

.cat-card:hover .cat-card__img img {
  transform: scale(1.08);
}

.cat-card__overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(28,25,23,0.55) 0%, transparent 60%);
  opacity: 0;
  transition: opacity var(--transition-base);
  display: flex;
  align-items: flex-end;
  justify-content: flex-end;
  padding: var(--spacing-md);
}

.cat-card:hover .cat-card__overlay {
  opacity: 1;
}

.cat-card__arrow {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: white;
  border-radius: var(--radius-full);
  color: var(--color-highlight);
  transform: translateY(8px);
  transition: transform var(--transition-base);
}

.cat-card:hover .cat-card__arrow {
  transform: translateY(0);
}

.cat-card__body {
  padding: var(--spacing-md) var(--spacing-lg);
}

.cat-card__name {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: 2px;
}

.cat-card__count {
  font-size: var(--text-sm);
  color: var(--color-text-lighter);
  font-weight: 500;
}

/* ===== WHY CHOOSE US ===== */
.why__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-3xl);
}

.why-card {
  background: white;
  padding: var(--spacing-xl) var(--spacing-lg);
  border-radius: var(--radius-xl);
  border: 1px solid var(--color-border-light);
  text-align: center;
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.why-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
  border-color: rgba(var(--color-highlight-rgb), 0.15);
}

.why-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: var(--radius-xl);
  margin-bottom: var(--spacing-md);
  color: var(--color-highlight);
}

.why-card__title {
  font-size: var(--text-lg);
  font-weight: 700;
  margin-bottom: var(--spacing-sm);
}

.why-card__desc {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  line-height: 1.6;
  margin: 0;
}

/* Certifications row */
.why__certs {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: var(--spacing-md);
}

.cert-badge {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  background: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.cert-badge:hover {
  border-color: var(--color-highlight);
  box-shadow: var(--shadow-md);
}

.cert-badge__icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background: rgba(var(--color-highlight-rgb), 0.08);
  border-radius: var(--radius-lg);
  color: var(--color-highlight);
}

.cert-badge__info h4 {
  font-size: var(--text-base);
  font-weight: 700;
  margin-bottom: 2px;
}

.cert-badge__info p {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin: 0;
}

/* ===== FEATURED PRODUCTS ===== */
.featured__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--spacing-xl);
  margin-bottom: var(--spacing-2xl);
}

.featured__cta {
  text-align: center;
}

.featured__cta .btn {
  gap: var(--spacing-sm);
}

/* ===== HOW TO ORDER STEPS ===== */
.steps__track {
  position: relative;
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-xl);
  margin-bottom: var(--spacing-2xl);
}

@media (min-width: 768px) {
  .steps__track {
    grid-template-columns: repeat(3, 1fr);
    gap: var(--spacing-2xl);
  }
}

.steps__line {
  display: none;
}

@media (min-width: 768px) {
  .steps__line {
    display: block;
    position: absolute;
    top: 52px;
    left: calc(16.67% + 24px);
    right: calc(16.67% + 24px);
    height: 2px;
    background: linear-gradient(90deg, var(--color-highlight), var(--color-accent), var(--color-highlight));
    opacity: 0.3;
  }
}

.steps__item {
  text-align: center;
  position: relative;
}

.steps__num {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: var(--color-highlight);
  color: white;
  font-size: var(--text-sm);
  font-weight: 800;
  border-radius: var(--radius-full);
  margin-bottom: var(--spacing-md);
  position: relative;
  z-index: 1;
  box-shadow: 0 0 0 4px rgba(var(--color-highlight-rgb), 0.15);
}

.steps__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  background: rgba(var(--color-highlight-rgb), 0.08);
  border-radius: var(--radius-xl);
  color: var(--color-highlight);
  margin-bottom: var(--spacing-md);
}

.steps__title {
  font-size: var(--text-lg);
  font-weight: 700;
  margin-bottom: var(--spacing-sm);
}

.steps__desc {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  line-height: 1.7;
  max-width: 280px;
  margin: 0 auto;
}

.steps__cta {
  text-align: center;
}

.steps__cta .btn {
  gap: var(--spacing-sm);
}

/* ===== OEM SERVICES ===== */
.oem__inner {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-3xl);
  align-items: center;
}

@media (min-width: 1024px) {
  .oem__inner {
    grid-template-columns: 1fr 1fr;
  }
}

.oem__subtitle {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-xl);
  max-width: 480px;
  line-height: 1.7;
}

.oem__features {
  list-style: none;
  margin-bottom: var(--spacing-xl);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.oem__features li {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: var(--text-base);
  color: var(--color-text);
  padding: var(--spacing-sm) 0;
}

.oem__features li svg {
  color: var(--color-success);
  flex-shrink: 0;
}

.oem__visual {
  position: relative;
}

.oem__visual img {
  width: 100%;
  border-radius: var(--radius-2xl);
  box-shadow: var(--shadow-xl);
}

.oem__visual-accent {
  position: absolute;
  top: var(--spacing-lg);
  left: var(--spacing-lg);
  right: calc(-1 * var(--spacing-md));
  bottom: calc(-1 * var(--spacing-md));
  border: 2px solid rgba(var(--color-highlight-rgb), 0.15);
  border-radius: var(--radius-2xl);
  z-index: -1;
}

/* ===== GLOBAL REACH ===== */
.global__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: var(--spacing-lg);
}

.region-card {
  text-align: center;
  padding: var(--spacing-xl) var(--spacing-lg);
  background: white;
  border-radius: var(--radius-xl);
  border: 1px solid var(--color-border-light);
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
}

.region-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
  border-color: rgba(var(--color-highlight-rgb), 0.2);
}

.region-card__flag {
  font-size: 3rem;
  margin-bottom: var(--spacing-md);
  line-height: 1;
}

.region-card__name {
  font-size: var(--text-lg);
  font-weight: 700;
  margin-bottom: var(--spacing-xs);
}

.region-card__countries {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin: 0;
}

/* ===== FAQ ===== */
.faq__list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-xl);
}

.faq-item {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: border-color var(--transition-base);
}

.faq-item--open {
  border-color: rgba(var(--color-highlight-rgb), 0.3);
  box-shadow: var(--shadow-sm);
}

.faq-item__q {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: var(--spacing-lg);
  font-size: var(--text-base);
  font-weight: 600;
  text-align: left;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-primary);
  gap: var(--spacing-md);
  transition: background-color var(--transition-fast);
}

.faq-item__q:hover {
  background: var(--color-bg-alt);
}

.faq-item__chevron {
  flex-shrink: 0;
  color: var(--color-text-lighter);
  transition: transform var(--transition-base);
}

.faq-item--open .faq-item__chevron {
  transform: rotate(180deg);
  color: var(--color-highlight);
}

.faq-item__a {
  padding: 0 var(--spacing-lg) var(--spacing-lg);
}

.faq-item__a p {
  color: var(--color-text-light);
  line-height: 1.8;
  margin: 0;
}

.faq__more {
  text-align: center;
}

/* FAQ transition */
.faq-slide-enter-active,
.faq-slide-leave-active {
  transition: background-color var(--transition-base), color var(--transition-base), transform var(--transition-base), box-shadow var(--transition-base);
  overflow: hidden;
}

.faq-slide-enter-from,
.faq-slide-leave-to {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  padding-bottom: 0;
}

/* ===== CTA BANNER ===== */
.cta-banner {
  position: relative;
  overflow: hidden;
}

.cta-banner__bg {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, #1c1917 0%, #292524 50%, #1c1917 100%);
}

.cta-banner__bg::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse 60% 50% at 30% 50%, rgba(var(--color-highlight-rgb), 0.12) 0%, transparent 70%),
    radial-gradient(ellipse 40% 40% at 80% 30%, rgba(var(--color-highlight-rgb), 0.08) 0%, transparent 60%);
}

.cta-banner__inner {
  position: relative;
  text-align: center;
  max-width: 700px;
  margin: 0 auto;
}

.cta-banner__inner h2 {
  color: white;
  margin-bottom: var(--spacing-md);
}

.cta-banner__inner p {
  color: rgba(255,255,255,0.7);
  margin-bottom: var(--spacing-2xl);
  font-size: var(--text-lg);
}

.cta-banner__actions {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.cta-banner__wa {
  border-color: rgba(255,255,255,0.3);
  color: white;
}

.cta-banner__wa:hover {
  border-color: white;
  background: rgba(255,255,255,0.1);
}

.cta-banner__actions .btn-ghost {
  color: rgba(255,255,255,0.8);
}

.cta-banner__actions .btn-ghost:hover {
  color: white;
  background: rgba(255,255,255,0.08);
}

/* ===== ANIMATIONS ===== */
@keyframes heroFadeIn {
  from {
    opacity: 0;
    transform: translateY(24px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

@keyframes floatBounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}
</style>
