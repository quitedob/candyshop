<template>
  <div class="home-page">
    <!-- Hero Section -->
    <section class="hero section">
      <div class="hero__bg">
        <div class="hero__overlay"></div>
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

          <p class="hero__description">
            {{ $t('home.hero.subtitle') }}
          </p>

          <div class="hero__actions">
            <NuxtLink :to="orderingCtaPath" class="btn btn-highlight btn-lg">
              {{ orderingCtaLabel }}
            </NuxtLink>
            <NuxtLink :to="localePath('/products')" class="btn btn-outline btn-lg">
              {{ $t('home.hero.cta_secondary') }}
            </NuxtLink>
          </div>

          <!-- Trust Indicators -->
          <div class="hero__trust">
            <div class="hero__trust-item">
              <span class="hero__trust-number">15+</span>
              <span class="hero__trust-label">{{ $t('factory.years') }}</span>
            </div>
            <div class="hero__trust-item">
              <span class="hero__trust-number">50+</span>
              <span class="hero__trust-label">{{ $t('factory.countries') }}</span>
            </div>
            <div class="hero__trust-item">
              <span class="hero__trust-number">100T</span>
              <span class="hero__trust-label">{{ $t('factory.daily_output') }}</span>
            </div>
          </div>
        </div>

        <!-- Hero Image -->
        <div class="hero__image">
          <div class="hero__image-wrapper">
            <img
              src="/images/hero-candy.jpg"
              :alt="$t('home.hero.title')"
              class="hero__img"
              loading="eager"
            />
          </div>
        </div>
      </div>
    </section>

    <!-- Categories Section -->
    <section class="categories section section-lg">
      <div class="container">
        <div class="section-header">
          <h2>{{ $t('home.categories.title') }}</h2>
          <p>{{ $t('home.categories.subtitle') }}</p>
        </div>

        <div v-if="categoriesStatus === 'pending'" class="categories__grid">
          <SkeletonLoader v-for="i in 4" :key="i" width="100%" height="250px" borderRadius="16px" />
        </div>
        <div v-else-if="categoriesStatus === 'error' || !categories.length">
          <ErrorState
            :title="t('errors.default')"
            :message="t('offline_message')"
            :actionText="t('errors.tryAgain')"
            @action="refreshCategories"
          />
        </div>
        <div v-else class="categories__grid">
          <NuxtLink
            v-for="category in categories"
            :key="category.slug"
            :to="localePath(category.to)"
            class="category-card"
          >
            <div class="category-card__image">
              <img :src="category.image" :alt="category.name" loading="lazy" />
              <div class="category-card__overlay">
                <span class="category-card__arrow">
                  <Icon name="lucide:arrow-right" size="20" />
                </span>
              </div>
            </div>
            <div class="category-card__content">
              <h3 class="category-card__title">{{ category.name }}</h3>
              <p class="category-card__count">{{ category.count }} {{ $t('product.products') }}</p>
            </div>
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- Trust Section -->
    <section class="trust section bg-alt">
      <div class="container">
        <div class="section-header">
          <h2>{{ $t('home.trust.title') }}</h2>
          <p>{{ $t('home.trust.subtitle') }}</p>
        </div>

        <div class="trust__grid">
          <div v-for="item in trustItems" :key="item.title" class="trust-card">
            <div class="trust-card__icon" :style="{ backgroundColor: item.color }">
              <Icon :name="item.icon" size="28" />
            </div>
            <h3 class="trust-card__title">{{ item.title }}</h3>
            <p class="trust-card__description">{{ item.description }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Certifications Section -->
    <section class="certifications section">
      <div class="container">
        <div class="section-header">
          <h2>{{ $t('home.certifications.title') }}</h2>
          <p>{{ $t('home.certifications.subtitle') }}</p>
        </div>

        <div class="certifications__list">
          <div v-for="cert in certifications" :key="cert.name" class="certification-badge">
            <div class="certification-badge__icon">
              <Icon :name="cert.icon" size="32" />
            </div>
            <div class="certification-badge__info">
              <h4 class="certification-badge__name">{{ cert.name }}</h4>
              <p class="certification-badge__description">{{ cert.description }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Featured Products -->
    <section class="featured section section-lg bg-alt">
      <div class="container">
        <div class="section-header">
          <h2>{{ $t('home.featured.title') }}</h2>
          <p>{{ $t('home.featured.subtitle') }}</p>
        </div>

        <div v-if="featuredStatus === 'pending'" class="featured__grid">
          <SkeletonLoader v-for="i in 4" :key="i" width="100%" height="320px" borderRadius="16px" />
        </div>
        <div v-else-if="featuredStatus === 'error' || !featuredProducts.length">
          <ErrorState
            :title="t('errors.default')"
            :message="t('offline_message')"
            :actionText="t('errors.tryAgain')"
            @action="refreshFeatured"
          />
        </div>
        <div v-else class="featured__grid">
          <ProductCard
            v-for="product in featuredProducts"
            :key="product.id"
            :product="product"
          />
        </div>

        <div class="featured__actions">
          <NuxtLink :to="localePath('/products')" class="btn btn-outline btn-lg">
            {{ $t('product.view_products') }}
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- How to Order -->
    <section class="how-to-order section">
      <div class="container">
        <div class="section-header">
          <h2>{{ $t('home.how_to_order.title') }}</h2>
          <p>{{ $t('home.how_to_order.subtitle') }}</p>
        </div>

        <div class="how-to-order__steps">
          <div class="how-to-order__step">
            <div class="how-to-order__step-number">1</div>
            <div class="how-to-order__step-icon">
              <Icon name="lucide:search" size="28" />
            </div>
            <h3 class="how-to-order__step-title">{{ $t('home.how_to_order.step1_title') }}</h3>
            <p class="how-to-order__step-desc">{{ $t('home.how_to_order.step1_desc') }}</p>
          </div>
          <div class="how-to-order__connector">
            <Icon name="lucide:chevron-right" size="24" />
          </div>
          <div class="how-to-order__step">
            <div class="how-to-order__step-number">2</div>
            <div class="how-to-order__step-icon">
              <Icon name="lucide:user-plus" size="28" />
            </div>
            <h3 class="how-to-order__step-title">{{ $t('home.how_to_order.step2_title') }}</h3>
            <p class="how-to-order__step-desc">{{ $t('home.how_to_order.step2_desc') }}</p>
          </div>
          <div class="how-to-order__connector">
            <Icon name="lucide:chevron-right" size="24" />
          </div>
          <div class="how-to-order__step">
            <div class="how-to-order__step-number">3</div>
            <div class="how-to-order__step-icon">
              <Icon name="lucide:truck" size="28" />
            </div>
            <h3 class="how-to-order__step-title">{{ $t('home.how_to_order.step3_title') }}</h3>
            <p class="how-to-order__step-desc">{{ $t('home.how_to_order.step3_desc') }}</p>
          </div>
        </div>

        <div class="how-to-order__cta">
          <NuxtLink :to="orderingCtaPath" class="btn btn-highlight btn-lg">
            {{ orderingCtaLabel }}
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- OEM Teaser -->
    <section class="oem-teaser section">
      <div class="container">
        <div class="oem-teaser__inner">
          <div class="oem-teaser__content">
            <span class="oem-teaser__badge">{{ $t('oem.title') }}</span>
            <h2>{{ $t('home.oem.title') }}</h2>
            <p>{{ $t('home.oem.subtitle') }}</p>

            <ul class="oem-teaser__features">
              <li v-for="feature in oemFeatures" :key="feature">
                <Icon name="lucide:check-circle" size="18" />
                <span>{{ feature }}</span>
              </li>
            </ul>

            <NuxtLink :to="localePath('/oem-solutions')" class="btn btn-highlight btn-lg">
              {{ $t('home.oem.cta') }}
            </NuxtLink>
          </div>

          <div class="oem-teaser__image">
            <img src="/images/oem-process.jpg" :alt="$t('oem.title')" loading="lazy" />
          </div>
        </div>
      </div>
    </section>

    <!-- Global Reach -->
    <section class="global section bg-alt">
      <div class="container">
        <div class="section-header">
          <h2>{{ $t('home.global.title') }}</h2>
          <p>{{ $t('home.global.subtitle') }}</p>
        </div>

        <div class="global__markets">
          <div v-for="region in markets" :key="region.name" class="market-card">
            <div class="market-card__flag">{{ region.flag }}</div>
            <h4 class="market-card__name">{{ region.name }}</h4>
            <p class="market-card__countries">{{ region.countries }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- FAQ Preview -->
    <section class="faq section">
      <div class="container container-narrow">
        <div class="section-header">
          <h2>{{ $t('home.faq.title') }}</h2>
          <p>{{ $t('home.faq.subtitle') }}</p>
        </div>

        <div class="faq__list">
          <div
            v-for="(faq, index) in faqs"
            :key="index"
            class="faq-item"
            :class="{ 'faq-item--open': openFaq === index }"
          >
            <button class="faq-item__question" @click="toggleFaq(index)">
              <span>{{ faq.question }}</span>
              <Icon
                name="lucide:chevron-down"
                size="18"
                class="faq-item__icon"
                :class="{ 'faq-item__icon--open': openFaq === index }"
              />
            </button>
            <div v-if="openFaq === index" class="faq-item__answer">
              {{ faq.answer }}
            </div>
          </div>
        </div>

        <div class="faq__actions">
          <NuxtLink :to="localePath('/faq')" class="btn btn-ghost">
            {{ $t('home.faq.view_all') }}
            <Icon name="lucide:arrow-right" size="16" />
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="cta section">
      <div class="container">
        <div class="cta__inner">
          <h2>{{ $t('form.title') }}</h2>
          <p>{{ $t('form.subtitle') }}</p>

          <div class="cta__actions">
            <NuxtLink :to="orderingCtaPath" class="btn btn-highlight btn-lg">
              {{ orderingCtaLabel }}
            </NuxtLink>
            <a :href="whatsappUrl" target="_blank" rel="noopener noreferrer" class="btn btn-outline btn-lg">
              <WhatsAppIcon size="20" />
              {{ $t('whatsapp.us') }}
            </a>
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
import { ref, computed } from 'vue'
import { useI18n, useLocalePath } from '#i18n'

const { t, locale } = useI18n()
const localePath = useLocalePath()
const config = useRuntimeConfig()
const { isAuthenticated, isAdmin, isPending } = useAuth()
const { getCategories, getFeaturedProducts } = useApi()

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
    return t('customer.products.add_to_cart')
  }
  return t('product.register_to_order')
})

// WhatsApp URL
const whatsappUrl = computed(() => {
  const number = config.public.whatsappNumber
  const message = encodeURIComponent(t('whatsapp.message'))
  return `https://wa.me/${number}?text=${message}`
})

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

// Trust items - reactive computed for proper i18n updates
const trustItems = computed(() => [
  {
    icon: 'lucide:award',
    title: t('factory.certifications'),
    description: t('factory.haccp') + ', ' + t('factory.iso'),
    color: '#dbeafe'
  },
  {
    icon: 'lucide:users',
    title: t('home.trust.expert_team'),
    description: t('home.trust.expert_team_desc'),
    color: '#fce7f3'
  },
  {
    icon: 'lucide:package',
    title: t('factory.capacity'),
    description: t('home.trust.capacity_desc'),
    color: '#dcfce7'
  },
  {
    icon: 'lucide:globe',
    title: t('factory.export'),
    description: t('home.trust.export_desc'),
    color: '#fef3c7'
  }
])

// Certifications - reactive computed for proper i18n updates
const certifications = computed(() => [
  { name: t('factory.haccp'), description: t('home.certifications.haccp_desc'), icon: 'lucide:shield-check' },
  { name: t('factory.iso'), description: t('home.certifications.iso_desc'), icon: 'lucide:check-circle-2' },
  { name: t('factory.brc'), description: t('home.certifications.brc_desc'), icon: 'lucide:award' },
  { name: t('factory.halal'), description: t('home.certifications.halal_desc'), icon: 'lucide:star' }
])

const { data: featuredProductsData, status: featuredStatus, refresh: refreshFeatured } = await useAsyncData(
  () => `home-featured-products-${locale.value}`,
  async () => {
    return await getFeaturedProducts(8)
  }
)

const featuredProducts = computed(() => featuredProductsData.value || [])

// OEM Features
const oemFeatures = computed(() => [
  t('home.oem.feature1'),
  t('home.oem.feature2'),
  t('home.oem.feature3'),
  t('home.oem.feature4'),
  t('home.oem.feature5')
])

// Global Markets
const markets = computed(() => [
  { name: t('home.global.north_america'), flag: '🇺🇸', countries: t('home.global.north_america_countries') },
  { name: t('home.global.europe'), flag: '🇪🇺', countries: t('home.global.europe_countries') },
  { name: t('home.global.southeast_asia'), flag: '🌏', countries: t('home.global.southeast_asia_countries') },
  { name: t('home.global.middle_east'), flag: '🕌', countries: t('home.global.middle_east_countries') }
])

// FAQ
const openFaq = ref(0)

const faqs = computed(() => [
  {
    question: t('home.faq.q1'),
    answer: t('home.faq.a1')
  },
  {
    question: t('home.faq.q2'),
    answer: t('home.faq.a2')
  },
  {
    question: t('home.faq.q3'),
    answer: t('home.faq.a3')
  },
  {
    question: t('home.faq.q4'),
    answer: t('home.faq.a4')
  }
])

const toggleFaq = (index: number) => {
  openFaq.value = openFaq.value === index ? -1 : index
}

// SEO
useSeo({
  title: t('seo.home_title'),
  description: t('seo.home_description'),
  ogType: 'website'
})
</script>

<style scoped>
/* Hero */
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
  background: linear-gradient(135deg, #fff5f0 0%, #fef3e2 50%, #fde8d8 100%);
}

.hero__overlay {
  position: absolute;
  inset: 0;
  background-image:
    radial-gradient(ellipse 80% 60% at 70% 40%, rgba(212, 165, 116, 0.12) 0%, transparent 70%),
    radial-gradient(ellipse 60% 50% at 20% 80%, rgba(255, 107, 74, 0.08) 0%, transparent 60%),
    url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23d4a574' fill-opacity='0.05'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
}

.hero__container {
  position: relative;
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-3xl);
}

@media (min-width: 1024px) {
  .hero__container {
    grid-template-columns: 1fr 1fr;
    align-items: center;
  }
}

.hero__content {
  animation: fadeInUp 0.8s ease;
}

.hero__badge {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: rgba(255, 107, 74, 0.1);
  border-radius: var(--radius-full);
  font-size: var(--text-sm);
  font-weight: 500;
  color: var(--color-highlight);
  margin-bottom: var(--spacing-lg);
}

.hero__badge-dot {
  width: 8px;
  height: 8px;
  background-color: var(--color-highlight);
  border-radius: var(--radius-full);
  animation: pulse 2s infinite;
}

.hero__title {
  font-size: clamp(2.5rem, 5vw, 4rem);
  line-height: 1.1;
  margin-bottom: var(--spacing-lg);
  color: var(--color-primary);
}

.hero__description {
  font-size: var(--text-lg);
  color: var(--color-text-light);
  max-width: 500px;
  margin-bottom: var(--spacing-xl);
}

.hero__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-3xl);
}

.hero__trust {
  display: flex;
  gap: var(--spacing-xl);
}

.hero__trust-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.hero__trust-number {
  font-size: var(--text-3xl);
  font-weight: 700;
  color: var(--color-highlight);
}

.hero__trust-label {
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

.hero__image-wrapper {
  position: relative;
  animation: fadeInUp 0.8s ease 0.2s backwards;
}

.hero__img {
  width: 100%;
  height: auto;
  border-radius: var(--radius-2xl);
  box-shadow: var(--shadow-2xl);
}

/* Categories */
.categories__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-lg);
}

.category-card {
  display: block;
  background-color: white;
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: var(--shadow-md);
  transition: all var(--transition-base);
}

.category-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-xl);
}

.category-card__image {
  position: relative;
  aspect-ratio: 4/3;
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
  background: linear-gradient(to top, rgba(0,0,0,0.5) 0%, transparent 100%);
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
  width: 48px;
  height: 48px;
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
  padding: var(--spacing-md);
  text-align: center;
}

.category-card__title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: var(--spacing-xs);
}

.category-card__count {
  font-size: var(--text-sm);
  color: var(--color-text-light);
}

/* Trust */
.trust__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--spacing-xl);
}

.trust-card {
  text-align: center;
  padding: var(--spacing-xl);
  background-color: white;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-base);
}

.trust-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.trust-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  margin-bottom: var(--spacing-md);
  border-radius: var(--radius-xl);
  color: var(--color-primary);
}

.trust-card__title {
  font-size: var(--text-xl);
  font-weight: 600;
  margin-bottom: var(--spacing-sm);
}

.trust-card__description {
  color: var(--color-text-light);
}

/* Certifications */
.certifications__list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-lg);
}

.certification-badge {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  background-color: white;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.certification-badge__icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-md);
  color: var(--color-accent);
}

.certification-badge__name {
  font-size: var(--text-base);
  font-weight: 600;
  margin-bottom: var(--spacing-xs);
}

.certification-badge__description {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin: 0;
}

/* Featured Products */
.featured__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--spacing-xl);
  margin-bottom: var(--spacing-xl);
}

.featured__actions {
  text-align: center;
}

/* How to Order */
.how-to-order__steps {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  gap: 0;
  margin-bottom: var(--spacing-2xl);
}

.how-to-order__step {
  flex: 1;
  max-width: 300px;
  text-align: center;
  padding: var(--spacing-xl);
  position: relative;
}

.how-to-order__step-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: var(--color-highlight);
  color: white;
  font-size: var(--text-sm);
  font-weight: 700;
  border-radius: var(--radius-full);
  margin-bottom: var(--spacing-md);
}

.how-to-order__step-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  background-color: var(--color-bg-alt);
  border-radius: var(--radius-xl);
  color: var(--color-highlight);
  margin-bottom: var(--spacing-md);
}

.how-to-order__step-title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: var(--spacing-sm);
}

.how-to-order__step-desc {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  line-height: 1.6;
}

.how-to-order__connector {
  display: flex;
  align-items: center;
  padding-top: 100px;
  color: var(--color-border);
}

.how-to-order__cta {
  text-align: center;
}

@media (max-width: 768px) {
  .how-to-order__steps {
    flex-direction: column;
    align-items: center;
  }

  .how-to-order__connector {
    transform: rotate(90deg);
    padding-top: 0;
    margin: calc(-1 * var(--spacing-sm)) 0;
  }
}

/* OEM Teaser */
.oem-teaser__inner {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--spacing-3xl);
  align-items: center;
}

@media (min-width: 1024px) {
  .oem-teaser__inner {
    grid-template-columns: 1fr 1fr;
  }
}

.oem-teaser__badge {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-md);
  background-color: var(--color-accent);
  color: white;
  font-size: var(--text-sm);
  font-weight: 600;
  border-radius: var(--radius-full);
  margin-bottom: var(--spacing-md);
}

.oem-teaser__content > h2 {
  margin-bottom: var(--spacing-md);
}

.oem-teaser__content > p {
  color: var(--color-text-light);
  margin-bottom: var(--spacing-xl);
  max-width: 500px;
}

.oem-teaser__features {
  list-style: none;
  margin-bottom: var(--spacing-xl);
}

.oem-teaser__features li {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-sm);
  color: var(--color-text);
}

.oem-teaser__features li svg {
  color: var(--color-success);
  flex-shrink: 0;
}

.oem-teaser__image img {
  width: 100%;
  border-radius: var(--radius-2xl);
  box-shadow: var(--shadow-xl);
}

/* Global Markets */
.global__markets {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-lg);
}

.market-card {
  text-align: center;
  padding: var(--spacing-xl);
  background-color: white;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
}

.market-card__flag {
  font-size: var(--text-4xl);
  margin-bottom: var(--spacing-md);
}

.market-card__name {
  font-size: var(--text-lg);
  font-weight: 600;
  margin-bottom: var(--spacing-xs);
}

.market-card__countries {
  font-size: var(--text-sm);
  color: var(--color-text-light);
  margin: 0;
}

/* FAQ */
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
}

.faq-item__question {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: var(--spacing-lg);
  font-size: var(--text-base);
  font-weight: 500;
  text-align: left;
  background: none;
  border: none;
  cursor: pointer;
  transition: background-color var(--transition-fast);
}

.faq-item__question:hover {
  background-color: var(--color-bg-alt);
}

.faq-item__icon {
  flex-shrink: 0;
  color: var(--color-text-light);
  transition: transform var(--transition-fast);
}

.faq-item__icon--open {
  transform: rotate(180deg);
}

.faq-item__answer {
  padding: 0 var(--spacing-lg) var(--spacing-lg);
  color: var(--color-text-light);
  line-height: 1.7;
}

.faq__actions {
  text-align: center;
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
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

/* Animations */
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
